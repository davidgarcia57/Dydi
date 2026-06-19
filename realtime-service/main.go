package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dydi/realtime-service/internal/delivery/websocket"
	"github.com/dydi/realtime-service/internal/usecase"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Fail closed: the gateway proves a WebSocket came through it (JWT-validated)
	// by stamping this secret, and habits/groups send it on /internal/broadcast.
	if os.Getenv("INTERNAL_TOKEN") == "" {
		log.Fatal("INTERNAL_TOKEN is required (shared gateway↔services secret)")
	}
	// Required so the WebSocket membership check (isMember) can't be silently
	// skipped in production — only tests, which don't run main, may omit it.
	if os.Getenv("GROUPS_SERVICE_URL") == "" {
		log.Fatal("GROUPS_SERVICE_URL is required (WebSocket membership check)")
	}

	h := usecase.NewHubUseCase()
	go h.Run()

	srv := &http.Server{
		Addr:    ":" + port(),
		Handler: setupRouter(h),
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		log.Printf("realtime-service listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-quit
	log.Println("shutdown signal received")

	h.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown error: %v", err)
	}
	log.Println("realtime-service stopped")
}

func port() string {
	p := os.Getenv("PORT")
	if p == "" {
		return "8084"
	}
	return p
}

// requireInternalToken rejects any request lacking the shared gateway↔services
// secret. A no-op when INTERNAL_TOKEN is unset (tests only — main refuses to
// boot without it).
func requireInternalToken(next http.Handler) http.Handler {
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

func setupRouter(h *usecase.HubUseCase) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]int{
			"active_connections": h.ConnectionCount(),
		})
	})

	r.Handle("/metrics", promhttp.Handler())

	// The /ws handshake and /internal/broadcast both arrive carrying the shared
	// secret (the gateway stamps it on the proxied WS upgrade; habits/groups on
	// broadcast). Reachable only through that trust boundary.
	r.Group(func(r chi.Router) {
		r.Use(requireInternalToken)
		r.Get("/ws/{groupID}", websocket.WebSocketHandler(h))
		r.Post("/internal/broadcast", websocket.BroadcastHandler(h))
	})

	return r
}
