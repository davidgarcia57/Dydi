package middleware

import (
	"net/http"
	"net/url"
	"os"
	"strings"
)

// WSOrigin enforces the browser Origin allowlist on the WebSocket handshake.
//
// The check lives here, at the edge, and deliberately NOT in realtime-service:
// the WS proxy rewrites Host to the upstream service (see proxy.WebSocket), so
// once the request is past this point "is Origin the same as Host" can never
// succeed, and the allowlist behind the proxy is forced to enumerate the
// synthetic Origin of every client platform. React Native's Android WebSocket
// always stamps an Origin derived from the URL it dialed
// (WebSocketModule.getDefaultOrigin: wss://host -> https://host) — that is, this
// gateway's own public origin, which is not a web origin anyone would think to
// put in ALLOWED_ORIGINS. That mismatch is exactly why the APK's handshake kept
// dying with a 403 while the web app worked fine through the same proxy.
//
// Here Host is still the real public host, so a native client's synthetic Origin
// is simply same-origin and passes with no configuration at all, while a browser
// served from another domain still has to be listed in ALLOWED_ORIGINS.
func WSOrigin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !wsOriginAllowed(r) {
			// A plain 403 before the upgrade: the client sees a failed handshake
			// instead of a socket that opens and dies.
			http.Error(w, `{"error":"origin not allowed"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// wsOriginAllowed reports whether this handshake's Origin may open a socket.
func wsOriginAllowed(r *http.Request) bool {
	origin := r.Header.Get("Origin")

	// Non-browser clients (curl, Go, several native stacks) send no Origin at
	// all. There is nothing to compare and nothing to defend: this WebSocket is
	// authorized by the bearer token in the query string, never by an ambient
	// cookie, so a cross-site page cannot forge an authenticated handshake the
	// way it could against cookie-authenticated endpoints.
	if origin == "" {
		return true
	}

	u, err := url.Parse(origin)
	if err != nil || u.Host == "" {
		return false
	}

	// Same-origin. Covers native clients, whose Origin mirrors the dialed URL.
	if strings.EqualFold(u.Host, r.Host) {
		return true
	}

	for _, allowed := range strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",") {
		if a := strings.TrimSpace(allowed); a != "" && strings.EqualFold(a, origin) {
			return true
		}
	}
	return false
}
