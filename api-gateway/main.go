package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	apimiddleware "github.com/dydi/api-gateway/internal/middleware"
	"github.com/dydi/api-gateway/internal/proxy"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           setupRouter(),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		log.Printf("api-gateway listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-quit
	log.Println("shutdown signal received")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown error: %v", err)
	}
	log.Println("api-gateway stopped")
}
func setupRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(apimiddleware.CORS)

	// /metrics is unauthenticated and lives at the root (not under /api) so a
	// Prometheus scraper / Grafana Cloud agent can read P95 latency.
	r.Handle("/metrics", promhttp.Handler())

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		// Despertar todos los servicios en segundo plano (asíncrono)
		urls := []string{
			os.Getenv("GROUPS_SERVICE_URL"),
			os.Getenv("HABITS_SERVICE_URL"),
			os.Getenv("REALTIME_SERVICE_URL"),
		}

		for _, u := range urls {
			if u == "" {
				continue
			}
			go func(serviceUrl string) {
				// Timeout de seguridad para la petición interna
				client := &http.Client{Timeout: 45 * time.Second}
				resp, err := client.Get(serviceUrl + "/health")
				if err == nil {
					_ = resp.Body.Close()
				}
			}(u)
		}

		// Respondemos "ok" de inmediato para evitar el timeout de 30s de Render
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Route("/api", func(r chi.Router) {
		r.Use(apimiddleware.Auth)
		r.Use(observability) // P95 latency per proxied route (kept off /ws to preserve hijacking)
		r.Mount("/users", proxy.To(os.Getenv("GROUPS_SERVICE_URL")))
		r.Mount("/groups", proxy.To(os.Getenv("GROUPS_SERVICE_URL")))
		r.Mount("/proposals", proxy.To(os.Getenv("GROUPS_SERVICE_URL")))
		r.Mount("/habits", proxy.To(os.Getenv("HABITS_SERVICE_URL")))
		r.Mount("/penalties", proxy.To(os.Getenv("HABITS_SERVICE_URL")))
	})

	r.With(apimiddleware.Auth).Mount("/ws", proxy.WebSocket(os.Getenv("REALTIME_SERVICE_URL")))

	return r
}
