package main

import (
	"context"
	"crypto/subtle"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
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

// envFloat lee un float de env; con valor ausente, malformado o no positivo
// regresa el default (el limiter no acepta tasas ≤ 0).
func envFloat(key string, fallback float64) float64 {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	f, err := strconv.ParseFloat(raw, 64)
	if err != nil || f <= 0 {
		log.Printf("%s=%q inválido, usando default %v", key, raw, fallback)
		return fallback
	}
	return f
}

func setupRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(apimiddleware.CORS)

	// El limiter agrupa por X-User-ID, así que las corridas k6 (todos los VUs
	// con la misma cuenta loadtest) comparten UNA cubeta: para el experimento
	// se suben estos valores por env en Render. Producción usa los defaults.
	limiter := apimiddleware.NewRateLimiter(
		envFloat("RATE_LIMIT_RPS", 5.0),
		envFloat("RATE_LIMIT_BURST", 20.0),
	)

	// /metrics is unauthenticated and lives at the root (not under /api) so a
	// Prometheus scraper / Grafana Cloud agent can read P95 latency.
	r.Handle("/metrics", promhttp.Handler())

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	wakeMu := &sync.Mutex{}
	wakeInProgress := false
	wakeLimiter := apimiddleware.NewRateLimiter(1.0, 2.0) // 1 req/s, burst 2

	r.With(apimiddleware.RateLimit(wakeLimiter)).Post("/ops/wake", func(w http.ResponseWriter, r *http.Request) {
		expectedToken := os.Getenv("WAKE_TOKEN")
		if expectedToken == "" {
			http.Error(w, "wake-up disabled: WAKE_TOKEN not configured", http.StatusInternalServerError)
			return
		}

		providedToken := r.Header.Get("X-Wake-Token")
		if providedToken == "" || subtle.ConstantTimeCompare([]byte(providedToken), []byte(expectedToken)) != 1 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		wakeMu.Lock()
		if wakeInProgress {
			wakeMu.Unlock()
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte("already in progress"))
			return
		}
		wakeInProgress = true
		wakeMu.Unlock()

		services := map[string]string{
			"groups":   os.Getenv("GROUPS_SERVICE_URL"),
			"habits":   os.Getenv("HABITS_SERVICE_URL"),
			"realtime": os.Getenv("REALTIME_SERVICE_URL"),
		}

		go func() {
			defer func() {
				wakeMu.Lock()
				wakeInProgress = false
				wakeMu.Unlock()
			}()

			var wg sync.WaitGroup
			for name, u := range services {
				if u == "" {
					log.Printf("keep-alive: %s_SERVICE_URL is empty — skipped", name)
					continue
				}
				wg.Add(1)
				go func(serviceName, serviceURL string) {
					defer wg.Done()
					client := &http.Client{Timeout: 45 * time.Second}
					resp, err := client.Get(serviceURL + "/health")
					if err != nil {
						log.Printf("keep-alive: failed to wake %s (%s): %v", serviceName, serviceURL, err)
						return
					}
					if resp.StatusCode != http.StatusOK {
						log.Printf("keep-alive: %s (%s) returned %d", serviceName, serviceURL, resp.StatusCode)
					}
					_ = resp.Body.Close()
				}(name, u)
			}
			wg.Wait()
		}()

		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("accepted"))
	})

	r.Route("/api", func(r chi.Router) {
		r.Use(apimiddleware.Auth)
		r.Use(apimiddleware.RateLimit(limiter))
		r.Use(observability) // P95 latency per proxied route (kept off /ws to preserve hijacking)
		r.Mount("/users", proxy.To(os.Getenv("GROUPS_SERVICE_URL")))
		r.Mount("/groups", proxy.To(os.Getenv("GROUPS_SERVICE_URL")))
		r.Mount("/proposals", proxy.To(os.Getenv("GROUPS_SERVICE_URL")))
		r.Mount("/habits", proxy.To(os.Getenv("HABITS_SERVICE_URL")))
		r.Mount("/penalties", proxy.To(os.Getenv("HABITS_SERVICE_URL")))
	})

	r.With(apimiddleware.Auth, apimiddleware.RateLimit(limiter)).Mount("/ws", proxy.WebSocket(os.Getenv("REALTIME_SERVICE_URL")))

	return r
}
