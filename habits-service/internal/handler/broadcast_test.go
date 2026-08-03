package handler

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// broadcast retries only on 5xx and transport errors, so every case here resolves
// on the first attempt and none of them hit the backoff sleeps. The request count
// is the real assertion: a retry on 4xx would burn 14s of sleeps per event.
func TestBroadcastResolvesOnFirstAttempt(t *testing.T) {
	cases := []struct {
		name      string
		status    int
		delivered bool
	}{
		{"delivered on 200", http.StatusOK, true},
		{"delivered on 204", http.StatusNoContent, true},
		{"gives up on 401 without retrying", http.StatusUnauthorized, false},
		{"gives up on 404 without retrying", http.StatusNotFound, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var calls atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				calls.Add(1)
				w.WriteHeader(tc.status)
			}))
			defer srv.Close()

			if got := broadcast(srv.URL, []byte(`{"type":"checkin"}`)); got != tc.delivered {
				t.Errorf("broadcast() = %v, want %v", got, tc.delivered)
			}
			if n := calls.Load(); n != 1 {
				t.Errorf("made %d requests, want 1", n)
			}
		})
	}
}
