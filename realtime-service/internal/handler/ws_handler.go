package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/dydi/realtime-service/internal/hub"
	"nhooyr.io/websocket"
)

func maxConnections() int {
	n, _ := strconv.Atoi(os.Getenv("MAX_CONNECTIONS_PER_GROUP"))
	if n <= 0 {
		n = 8
	}
	return n
}

func WebSocket(h *hub.Hub) http.HandlerFunc {
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
			InsecureSkipVerify: true, // connections arrive via api-gateway, not directly from browsers
		})
		if err != nil {
			return
		}

		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		client := hub.NewClient(groupID, userID, conn)
		h.Register(client)
		defer h.Unregister(client)

		h.Broadcast(hub.Event{
			Type:    hub.EventMemberOnline,
			GroupID: groupID,
			UserID:  userID,
		})

		go client.ReadPump(ctx, cancel)
		client.WritePump(ctx)

		h.Broadcast(hub.Event{
			Type:    hub.EventMemberOffline,
			GroupID: groupID,
			UserID:  userID,
		})
	}
}

func Broadcast(h *hub.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ev hub.Event
		if err := json.NewDecoder(r.Body).Decode(&ev); err != nil {
			http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
			return
		}
		h.Broadcast(ev)
		w.WriteHeader(http.StatusNoContent)
	}
}
