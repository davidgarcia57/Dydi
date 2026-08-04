package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestWSOriginAllowlist pins the handshake rule that the mobile APK depends on.
//
// The regression this guards: React Native stamps Origin from the URL it dialed,
// so the app's Origin is the gateway's own public origin — never the web app's
// domain. Enforcing the browser allowlist against that value rejected every
// native handshake with a 403 while the browser kept working, which read as "the
// WebSocket is broken on mobile" for weeks.
func TestWSOriginAllowlist(t *testing.T) {
	const host = "api-gateway-j3yi.onrender.com"

	cases := []struct {
		name   string
		origin string
		want   int
	}{
		{"navegador en un origen permitido", "https://dydi-xi.vercel.app", http.StatusOK},
		{"movil nativo: Origin espejo del host marcado", "https://" + host, http.StatusOK},
		{"cliente sin Origin (curl, stacks nativos)", "", http.StatusOK},
		{"navegador en un origen ajeno", "https://evil.example.com", http.StatusForbidden},
		{"Origin sin host", "null", http.StatusForbidden},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("ALLOWED_ORIGINS", "https://dydi-xi.vercel.app")

			h := WSOrigin(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/ws/group-1", nil)
			req.Host = host
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)

			if w.Code != tc.want {
				t.Fatalf("Origin %q: se esperaba %d, se obtuvo %d", tc.origin, tc.want, w.Code)
			}
		})
	}
}

// The same-origin shortcut must not depend on ALLOWED_ORIGINS being set: that is
// what makes the native client work with no per-deploy configuration, so a new
// Render URL can never silently break the APK again.
func TestWSOriginSameOriginWithoutAllowlist(t *testing.T) {
	t.Setenv("ALLOWED_ORIGINS", "")

	h := WSOrigin(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/ws/group-1", nil)
	req.Host = "localhost:8080"
	req.Header.Set("Origin", "http://localhost:8080")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("same-origin sin allowlist: se esperaba 200, se obtuvo %d", w.Code)
	}
}
