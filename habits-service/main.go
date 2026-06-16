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
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		panic("db connect failed: " + err.Error())
	}
	defer pool.Close()

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

func setupRouter(pool *pgxpool.Pool) *chi.Mux {
	r := chi.NewRouter()
	r.Use(observability)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	r.Handle("/metrics", promhttp.Handler())

	habits := handler.NewHabitHandler(pool, os.Getenv("REALTIME_SERVICE_URL"))
	r.Get("/habits", habits.ListHabits)
	r.Post("/habits/checkins", habits.CreateCheckin)
	r.Get("/habits/checkins/{groupID}/today", habits.GetTodayCheckins)
	r.Get("/habits/history/{groupID}", habits.GetHistory)
	r.Get("/habits/streaks/{userID}", habits.GetStreaks)

	penalties := handler.NewPenaltyHandler(pool, os.Getenv("REALTIME_SERVICE_URL"))
	r.Get("/penalties/{groupID}/eligible", penalties.GetEligible)
	r.Post("/penalties/roulette", penalties.OpenRoulette)
	r.Post("/penalties/roulette/{entryID}/suggestions", penalties.SubmitSuggestion)
	r.Get("/penalties/roulette/{entryID}/suggestions", penalties.GetSuggestions)
	r.Post("/penalties/roulette/{entryID}/spin", penalties.Spin)
	r.Get("/penalties/{groupID}/debts", penalties.GetActiveDebts)

	// Internal: called by groups-service when a proposal is approved.
	r.Post("/internal/proposals/apply", habits.ApplyProposal)

	return r
}
