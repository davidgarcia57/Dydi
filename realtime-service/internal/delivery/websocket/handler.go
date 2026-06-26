package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/dydi/realtime-service/internal/domain"
	"github.com/dydi/realtime-service/internal/usecase"
	"github.com/go-chi/chi/v5"
)

// allowedOriginPatterns returns the list of origin patterns for WebSocket
// handshake validation. When ALLOWED_ORIGINS is empty (local dev), it returns
// ["*"] to accept any origin — matching the previous InsecureSkipVerify behavior.
func allowedOriginPatterns() []string {
	raw := os.Getenv("ALLOWED_ORIGINS")
	if raw == "" {
		return []string{"*"}
	}
	parts := strings.Split(raw, ",")
	patterns := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			patterns = append(patterns, s)
		}
	}
	if len(patterns) == 0 {
		return []string{"*"}
	}
	return patterns
}

// isMember asks groups-service whether userID is an active member of groupID,
// before we let them subscribe to that group's live events. Fail closed: if the
// check can't be completed (network/cold start), the connection is refused.
// When GROUPS_SERVICE_URL is unset (local/tests) the check is skipped.
func isMember(ctx context.Context, groupID, userID string) bool {
	base := os.Getenv("GROUPS_SERVICE_URL")
	if base == "" {
		return true
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	u := base + "/internal/groups/" + url.PathEscape(groupID) + "/members/" + url.PathEscape(userID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return false
	}
	if tok := os.Getenv("INTERNAL_TOKEN"); tok != "" {
		req.Header.Set("X-Internal-Token", tok)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode == http.StatusNoContent
}

func maxConnections() int {
	n, _ := strconv.Atoi(os.Getenv("MAX_CONNECTIONS_PER_GROUP"))
	if n <= 0 {
		n = 8
	}
	return n
}

func pingInterval() time.Duration {
	s, _ := strconv.Atoi(os.Getenv("PING_INTERVAL_SECONDS"))
	if s <= 0 {
		s = 30
	}
	return time.Duration(s) * time.Second
}

func writeWait() time.Duration {
	s, _ := strconv.Atoi(os.Getenv("WRITE_WAIT_SECONDS"))
	if s <= 0 {
		s = 10
	}
	return time.Duration(s) * time.Second
}

func WebSocketHandler(h *usecase.HubUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupID := chi.URLParam(r, "groupID")
		userID := r.Header.Get("X-User-ID")

		if userID == "" {
			http.Error(w, `{"error":"missing X-User-ID"}`, http.StatusUnauthorized)
			return
		}

		// A valid JWT alone isn't enough: only members of this group may listen
		// to its live events (debts, check-ins, presence).
		if !isMember(r.Context(), groupID, userID) {
			http.Error(w, `{"error":"not a member of this group"}`, http.StatusForbidden)
			return
		}

		if h.RoomConnectionCount(groupID) >= maxConnections() {
			http.Error(w, `{"error":"group connection limit reached"}`, http.StatusConflict)
			return
		}

		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			// Validate the Origin header against the configured allowed origins.
			// In local dev ALLOWED_ORIGINS is empty → accept all (like before).
			OriginPatterns: allowedOriginPatterns(),
		})
		if err != nil {
			return
		}

		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		client := domain.NewClient(groupID, userID)
		h.Register(client)
		defer h.Unregister(client)

		h.Broadcast(domain.Event{
			Type:    domain.EventMemberOnline,
			GroupID: groupID,
			UserID:  userID,
		})

		go readPump(ctx, cancel, conn)
		writePump(ctx, client, conn)

		h.Broadcast(domain.Event{
			Type:    domain.EventMemberOffline,
			GroupID: groupID,
			UserID:  userID,
		})
	}
}

func BroadcastHandler(h *usecase.HubUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Service-to-service auth is enforced by the requireInternalToken
		// middleware on this route (the endpoint is internet-reachable on Render).
		var ev domain.Event
		if err := json.NewDecoder(r.Body).Decode(&ev); err != nil {
			http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
			return
		}
		h.Broadcast(ev)
		w.WriteHeader(http.StatusNoContent)
	}
}

func writePump(ctx context.Context, c *domain.Client, conn *websocket.Conn) {
	ticker := time.NewTicker(pingInterval())
	defer ticker.Stop()

	for {
		select {
		case ev, ok := <-c.Send:
			if !ok {
				_ = conn.Close(websocket.StatusNormalClosure, "")
				return
			}
			writeCtx, cancel := context.WithTimeout(ctx, writeWait())
			err := wsjson.Write(writeCtx, conn, ev)
			cancel()
			if err != nil {
				return
			}

		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, writeWait())
			err := conn.Ping(pingCtx)
			cancel()
			if err != nil {
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

func readPump(ctx context.Context, cancel context.CancelFunc, conn *websocket.Conn) {
	defer cancel()
	for {
		if _, _, err := conn.Read(ctx); err != nil {
			return
		}
	}
}
