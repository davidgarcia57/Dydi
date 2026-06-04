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

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/dydi/realtime-service/internal/handler"
	"github.com/dydi/realtime-service/internal/hub"
)

func main() {
	h := hub.New()
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

	// Broadcast member_offline for all active clients before closing
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

func setupRouter(h *hub.Hub) *chi.Mux {
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

	r.Handle("/metrics", promhttp.Handler())

	r.Get("/ws/{groupID}", handler.WebSocket(h))
	r.Post("/internal/broadcast", handler.Broadcast(h))

	return r
}
