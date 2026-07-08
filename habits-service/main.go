package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dydi/habits-service/internal/handler"
	sharedDb "github.com/dydi/shared/db"
	sharedMiddleware "github.com/dydi/shared/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Fail closed: without the shared gateway↔services secret every endpoint
	// would trust the X-User-ID header from any internet caller.
	if os.Getenv("INTERNAL_TOKEN") == "" {
		log.Fatal("INTERNAL_TOKEN is required (shared gateway↔services secret)")
	}

	pool, err := sharedDb.NewDBPool(context.Background())
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

const requestTimeout = 15 * time.Second

func setupRouter(pool *pgxpool.Pool) *chi.Mux {
	r := chi.NewRouter()
	r.Use(observability)
	r.Use(middleware.Recoverer)
	// gzip de respuestas: recorta el egreso de este servicio hacia el gateway
	// (cuentas Render separadas → esa respuesta cuenta como egreso propio).
	r.Use(middleware.Compress(5))
	// Bound every request so a slow query under load fails fast (504) instead of
	// piling up goroutines and pushing the 512 MB instance toward an OOM kill.
	r.Use(middleware.Timeout(requestTimeout))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Readiness: /health proves the process is up; /ready also proves Postgres
	// is reachable. Lets the experiment tell "service up" from "DB up" when
	// something degrades under load.
	r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		w.Header().Set("Content-Type", "application/json")
		if err := pool.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"unavailable"}`))
			return
		}
		_, _ = w.Write([]byte(`{"status":"ready"}`))
	})

	r.Handle("/metrics", promhttp.Handler())

	r.Group(func(r chi.Router) {
		r.Use(sharedMiddleware.RequireInternalToken)

		habits := handler.NewHabitHandler(pool, os.Getenv("REALTIME_SERVICE_URL"))
		r.Get("/habits", habits.ListHabits)
		r.Post("/habits/checkins", habits.CreateCheckin)
		r.Get("/habits/checkins/{groupID}/today", habits.GetTodayCheckins)
		r.Get("/habits/history/{groupID}", habits.GetHistory)
		r.Get("/habits/streaks/{userID}", habits.GetStreaks)

		catalog := handler.LoadPunishmentCatalog(os.Getenv("PUNISHMENT_CATALOG_PATH"))
		penalties := handler.NewPenaltyHandler(pool, os.Getenv("REALTIME_SERVICE_URL"), catalog)
		r.Get("/penalties/{groupID}/eligible", penalties.GetEligible)
		r.Get("/penalties/{groupID}/roulette", penalties.GetOpenRoulettes)
		r.Post("/penalties/roulette", penalties.OpenRoulette)
		r.Post("/penalties/roulette/{entryID}/suggestions", penalties.SubmitSuggestion)
		r.Get("/penalties/roulette/{entryID}/suggestions", penalties.GetSuggestions)
		r.Post("/penalties/roulette/{entryID}/spin", penalties.Spin)
		r.Get("/penalties/{groupID}/debts", penalties.GetActiveDebts)
		r.Post("/penalties/debts/{debtID}/complete", penalties.CompleteDebt)
		r.Post("/penalties/debts/{debtID}/forgive", penalties.ForgiveDebt)

		// Internal: called by groups-service when a proposal is approved.
		r.Post("/internal/proposals/apply", habits.ApplyProposal)
	})

	return r
}
