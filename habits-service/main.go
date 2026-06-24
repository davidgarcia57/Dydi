package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/dydi/habits-service/internal/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Fail closed: without the shared gateway↔services secret every endpoint
	// would trust the X-User-ID header from any internet caller.
	if os.Getenv("INTERNAL_TOKEN") == "" {
		log.Fatal("INTERNAL_TOKEN is required (shared gateway↔services secret)")
	}

	pool, err := newDBPool(context.Background())
	if err != nil {
		panic("db connect failed: " + err.Error())
	}
	defer pool.Close()
	registerPoolMetrics(pool)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           setupRouter(pool),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		log.Printf("habits-service listening on %s", srv.Addr)
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
	log.Println("habits-service stopped")
}

// requireInternalToken rejects any request lacking the shared gateway↔services
// secret, so these endpoints are reachable only through the gateway (which
// validated the JWT) or sibling services. A no-op when INTERNAL_TOKEN is unset
// (tests only — main refuses to boot without it).
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

// requestTimeout bounds every HTTP request. Tuned for the load experiment: a
// slow query under stress fails fast instead of accumulating goroutines.
const requestTimeout = 15 * time.Second

// newDBPool builds the pgx pool. Tuned for Supabase's Supavisor transaction
// pooler (port 6543): MaxConns is bounded (and tunable per k6 ramp via
// DB_MAX_CONNS) so the four services don't exhaust the shared connection budget,
// and QueryExecModeExec avoids implicit prepared statements, which transaction
// pooling can't keep alive across multiplexed backend connections.
func newDBPool(ctx context.Context) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, err
	}
	cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec
	cfg.MaxConns = envInt32("DB_MAX_CONNS", 10)
	cfg.MinConns = 0
	cfg.MaxConnIdleTime = time.Minute
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.HealthCheckPeriod = 30 * time.Second
	return pgxpool.NewWithConfig(ctx, cfg)
}

func envInt32(key string, def int32) int32 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return int32(n)
		}
	}
	return def
}

func setupRouter(pool *pgxpool.Pool) *chi.Mux {
	r := chi.NewRouter()
	r.Use(observability)
	r.Use(middleware.Recoverer)
	// Bound every request so a slow query under load fails fast (504) instead of
	// piling up goroutines and pushing the 512 MB instance toward an OOM kill.
	r.Use(middleware.Timeout(requestTimeout))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Handle("/metrics", promhttp.Handler())

	r.Group(func(r chi.Router) {
		r.Use(requireInternalToken)

		habits := handler.NewHabitHandler(pool, os.Getenv("REALTIME_SERVICE_URL"))
		r.Get("/habits", habits.ListHabits)
		r.Post("/habits/checkins", habits.CreateCheckin)
		r.Get("/habits/checkins/{groupID}/today", habits.GetTodayCheckins)
		r.Get("/habits/history/{groupID}", habits.GetHistory)
		r.Get("/habits/streaks/{userID}", habits.GetStreaks)

		catalog := handler.LoadPunishmentCatalog(os.Getenv("PUNISHMENT_CATALOG_PATH"))
		penalties := handler.NewPenaltyHandler(pool, os.Getenv("REALTIME_SERVICE_URL"), catalog)
		r.Get("/penalties/{groupID}/eligible", penalties.GetEligible)
		r.Post("/penalties/roulette", penalties.OpenRoulette)
		r.Post("/penalties/roulette/{entryID}/suggestions", penalties.SubmitSuggestion)
		r.Get("/penalties/roulette/{entryID}/suggestions", penalties.GetSuggestions)
		r.Post("/penalties/roulette/{entryID}/spin", penalties.Spin)
		r.Get("/penalties/{groupID}/debts", penalties.GetActiveDebts)

		// Internal: called by groups-service when a proposal is approved.
		r.Post("/internal/proposals/apply", habits.ApplyProposal)
	})

	return r
}
