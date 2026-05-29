package handler

import (
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

		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: false,
		})
		if err != nil {
			return
		}

		ctx := r.Context()
		client := hub.NewClient(groupID, userID, conn)
		h.Register(client)
		defer h.Unregister(client)

		h.Broadcast(hub.Event{
			Type:    "member_online",
			GroupID: groupID,
			UserID:  userID,
		})

		client.WritePump(ctx)

		h.Broadcast(hub.Event{
			Type:    "member_offline",
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
