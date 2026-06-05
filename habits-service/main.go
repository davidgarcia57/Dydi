package main

import (
	"context"
	"net/http"
	"os"

	"github.com/dydi/habits-service/internal/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		panic("db connect failed: " + err.Error())
	}
	defer pool.Close()

	r := setupRouter(pool)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}
	http.ListenAndServe(":"+port, r)
}

func setupRouter(pool *pgxpool.Pool) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	habits := handler.NewHabitHandler(pool)
	r.Get("/habits", habits.ListHabits)
	r.Post("/habits/checkins", habits.CreateCheckin)
	r.Get("/habits/checkins/{groupID}/today", habits.GetTodayCheckins)
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
