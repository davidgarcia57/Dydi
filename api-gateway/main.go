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
	// The gateway is the only JWT-validating hop; it proves to the backend
	// services that a request came through it by stamping INTERNAL_TOKEN on
	// every proxied request. Without it, the services would have to trust the
	// X-User-ID header from any internet caller — refuse to boot instead.
	if os.Getenv("INTERNAL_TOKEN") == "" {
		log.Fatal("INTERNAL_TOKEN is required (shared gateway↔services secret)")
	}

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
		// Wake every backend service in the background. Each lives on a separate
		// Render account, so this 12-min ping is what keeps them off the
		// free-tier 15-min sleep. Failures are LOGGED (not silent) so a missing
		// or wrong *_SERVICE_URL surfaces instead of quietly dropping a service.
		services := map[string]string{
			"groups":   os.Getenv("GROUPS_SERVICE_URL"),
			"habits":   os.Getenv("HABITS_SERVICE_URL"),
			"realtime": os.Getenv("REALTIME_SERVICE_URL"),
		}

		for name, u := range services {
			if u == "" {
				log.Printf("keep-alive: %s_SERVICE_URL is empty — that service will NOT be kept awake", name)
				continue
			}
			go func(name, serviceURL string) {
				// 45s is enough to cover a Render cold start of the target.
				client := &http.Client{Timeout: 45 * time.Second}
				resp, err := client.Get(serviceURL + "/health")
				if err != nil {
					log.Printf("keep-alive: failed to wake %s (%s): %v", name, serviceURL, err)
					return
				}
				if resp.StatusCode != http.StatusOK {
					log.Printf("keep-alive: %s (%s) returned %d", name, serviceURL, resp.StatusCode)
				}
				_ = resp.Body.Close()
			}(name, u)
		}

		// Respond immediately so we never hit Render's 30s request timeout.
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
