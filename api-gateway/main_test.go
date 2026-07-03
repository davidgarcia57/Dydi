package main

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// eventually polls condition every 10 ms and fails the test if it does not
// become true within timeout.  It is intended only for waiting on goroutine
// side-effects that cannot be synchronised with channels or WaitGroups
// (e.g. the internal wakeInProgress flag that is not exported).
func eventually(t *testing.T, timeout time.Duration, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition was not met within timeout")
}

// -- Health & auth ------------------------------------------------------------

func TestHealthEndpoint(t *testing.T) {
	r := setupRouter()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "ok" {
		t.Fatalf("expected 'ok', got %q", w.Body.String())
	}
}

func TestProtectedRouteRequiresAuth(t *testing.T) {
	r := setupRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/groups/some-id", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d", w.Code)
	}
}

// -- /ops/wake - token validation ---------------------------------------------

func TestOpsWakeWithoutTokenConfigured(t *testing.T) {
	t.Setenv("WAKE_TOKEN", "")
	r := setupRouter()
	req := httptest.NewRequest(http.MethodPost, "/ops/wake", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 when token not configured, got %d", w.Code)
	}
}

func TestOpsWakeWithInvalidToken(t *testing.T) {
	t.Setenv("WAKE_TOKEN", "secret-token")
	r := setupRouter()
	req := httptest.NewRequest(http.MethodPost, "/ops/wake", nil)
	req.Header.Set("X-Wake-Token", "wrong-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 with invalid token, got %d", w.Code)
	}
}

func TestOpsWakeWithMissingToken(t *testing.T) {
	t.Setenv("WAKE_TOKEN", "secret-token")
	r := setupRouter()
	req := httptest.NewRequest(http.MethodPost, "/ops/wake", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 with missing token, got %d", w.Code)
	}
}

// -- /ops/wake - fan-out, deduplication & lifecycle ----------------------------

func TestOpsWakeFanOutAllServices(t *testing.T) {
	var calls int32
	var handlerWg sync.WaitGroup
	handlerWg.Add(3) // one call per configured service

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusOK)
		handlerWg.Done()
	}))
	defer ts.Close()

	t.Setenv("WAKE_TOKEN", "test-token")
	t.Setenv("GROUPS_SERVICE_URL", ts.URL)
	t.Setenv("HABITS_SERVICE_URL", ts.URL)
	t.Setenv("REALTIME_SERVICE_URL", ts.URL)

	r := setupRouter()
	req := httptest.NewRequest(http.MethodPost, "/ops/wake", nil)
	req.Header.Set("X-Wake-Token", "test-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", w.Code)
	}

	// Deterministic: wait until all three downstream handlers have returned.
	handlerWg.Wait()

	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("expected 3 downstream calls, got %d", got)
	}
}

func TestOpsWakeDeduplication(t *testing.T) {
	var calls int32
	// gate blocks the downstream handler so we can make a second wake request
	// while the first is still in-flight.
	gate := make(chan struct{})
	var handlerWg sync.WaitGroup
	handlerWg.Add(1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		<-gate // hold until test releases
		w.WriteHeader(http.StatusOK)
		handlerWg.Done()
	}))
	defer ts.Close()

	t.Setenv("WAKE_TOKEN", "test-token")
	t.Setenv("GROUPS_SERVICE_URL", ts.URL)
	t.Setenv("HABITS_SERVICE_URL", "")
	t.Setenv("REALTIME_SERVICE_URL", "")

	r := setupRouter()

	// First request triggers fan-out.
	req1 := httptest.NewRequest(http.MethodPost, "/ops/wake", nil)
	req1.Header.Set("X-Wake-Token", "test-token")
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	if w1.Code != http.StatusAccepted {
		t.Fatalf("expected 202 on first call, got %d", w1.Code)
	}
	if w1.Body.String() != "accepted" {
		t.Fatalf("expected body 'accepted', got %q", w1.Body.String())
	}

	// Second request while the first is still in-flight.
	req2 := httptest.NewRequest(http.MethodPost, "/ops/wake", nil)
	req2.Header.Set("X-Wake-Token", "test-token")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusAccepted {
		t.Fatalf("expected 202 on second call, got %d", w2.Code)
	}
	if w2.Body.String() != "already in progress" {
		t.Fatalf("expected body 'already in progress', got %q", w2.Body.String())
	}

	// Release downstream and wait for handler to finish.
	close(gate)
	handlerWg.Wait()

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected exactly 1 downstream call due to dedup, got %d", got)
	}
}

func TestOpsWakeNewCycleAfterCompletion(t *testing.T) {
	var calls int32
	// Buffered so the handler never blocks.
	cycleDone := make(chan struct{}, 2)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusOK)
		cycleDone <- struct{}{}
	}))
	defer ts.Close()

	t.Setenv("WAKE_TOKEN", "test-token")
	t.Setenv("GROUPS_SERVICE_URL", ts.URL)
	t.Setenv("HABITS_SERVICE_URL", "")
	t.Setenv("REALTIME_SERVICE_URL", "")

	r := setupRouter()

	// First cycle
	req1 := httptest.NewRequest(http.MethodPost, "/ops/wake", nil)
	req1.Header.Set("X-Wake-Token", "test-token")
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusAccepted {
		t.Fatalf("first cycle: expected 202, got %d", w1.Code)
	}

	// Wait for the downstream handler to run.
	<-cycleDone

	// Wait for the internal wakeInProgress flag to be cleared so a new cycle
	// can start.  We detect this by observing that a fresh POST returns
	// "accepted" and is not rate-limited.
	eventually(t, 3*time.Second, func() bool {
		probe := httptest.NewRequest(http.MethodPost, "/ops/wake", nil)
		probe.Header.Set("X-Wake-Token", "test-token")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, probe)
		return rec.Code == http.StatusAccepted && rec.Body.String() == "accepted"
	})

	// Wait for the second cycle downstream call.
	<-cycleDone

	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("expected 2 total downstream calls across two cycles, got %d", got)
	}
}

func TestOpsWakeCleanupOnDownstreamFailure(t *testing.T) {
	// Downstream always returns 500.  The wake goroutine should still reset
	// wakeInProgress so subsequent calls are not permanently blocked.
	failDone := make(chan struct{}, 1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		select {
		case failDone <- struct{}{}:
		default:
		}
	}))
	defer ts.Close()

	t.Setenv("WAKE_TOKEN", "test-token")
	t.Setenv("GROUPS_SERVICE_URL", ts.URL)
	t.Setenv("HABITS_SERVICE_URL", "")
	t.Setenv("REALTIME_SERVICE_URL", "")

	r := setupRouter()

	req := httptest.NewRequest(http.MethodPost, "/ops/wake", nil)
	req.Header.Set("X-Wake-Token", "test-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", w.Code)
	}

	// Wait for the failing downstream call.
	<-failDone

	// Even after a downstream error, wakeInProgress must be reset.
	eventually(t, 3*time.Second, func() bool {
		probe := httptest.NewRequest(http.MethodPost, "/ops/wake", nil)
		probe.Header.Set("X-Wake-Token", "test-token")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, probe)
		return rec.Code == http.StatusAccepted && rec.Body.String() == "accepted"
	})
}

// -- envFloat (config del rate limiter) -----------------------------------------

func TestEnvFloat(t *testing.T) {
	cases := []struct {
		name string
		val  string
		want float64
	}{
		{"vacío regresa default", "", 5.0},
		{"entero válido", "250", 250},
		{"decimal válido", "0.5", 0.5},
		{"malformado regresa default", "abc", 5.0},
		{"cero regresa default", "0", 5.0},
		{"negativo regresa default", "-3", 5.0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("RATE_LIMIT_RPS", tc.val)
			if got := envFloat("RATE_LIMIT_RPS", 5.0); got != tc.want {
				t.Fatalf("envFloat(%q) = %v, want %v", tc.val, got, tc.want)
			}
		})
	}
}
