package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/dydi/realtime-service/internal/domain"
	"github.com/dydi/realtime-service/internal/usecase"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

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

		if h.RoomConnectionCount(groupID) >= maxConnections() {
			http.Error(w, `{"error":"group connection limit reached"}`, http.StatusConflict)
			return
		}

		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true,
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
				conn.Close(websocket.StatusNormalClosure, "")
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
