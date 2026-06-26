package middleware

import (
	"net/http"
	"os"
)

// RequireInternalToken rejects any request that does not carry the shared
// gateway↔services secret. The gateway stamps it on every proxied request and
// sibling services send it on /internal/* calls, so a backend endpoint can only
// be reached through the gateway (which validated the JWT). When INTERNAL_TOKEN
// is unset — only in tests; main refuses to boot without it — the check is a
// no-op so unit tests keep exercising the handlers directly.
func RequireInternalToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if expected := os.Getenv("INTERNAL_TOKEN"); expected != "" && r.Header.Get("X-Internal-Token") != expected {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}
