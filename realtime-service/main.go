package main

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/dydi/realtime-service/internal/hub"
	"github.com/dydi/realtime-service/internal/handler"
)

func main() {
	h := hub.New()
	go h.Run()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]int{
			"active_connections": h.ConnectionCount(),
		})
	})

	r.Get("/ws/{groupID}", handler.WebSocket(h))

	// Internal HTTP endpoint for other services to broadcast events
	r.Post("/internal/broadcast", handler.Broadcast(h))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}
	http.ListenAndServe(":"+port, r)
}
