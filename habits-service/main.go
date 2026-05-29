package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/dydi/habits-service/internal/handler"
	"github.com/dydi/habits-service/internal/model"
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

	catalog := loadCatalog()

	r := setupRouter(pool, catalog)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}
	http.ListenAndServe(":"+port, r)
}

func setupRouter(pool *pgxpool.Pool, catalog []model.Punishment) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	habits := handler.NewHabitHandler(pool)
	r.Get("/habits", habits.ListHabits)
	r.Post("/habits/assign", habits.AssignHabit)
	r.Post("/checkins", habits.CreateCheckin)
	r.Get("/checkins/{groupID}/today", habits.GetTodayCheckins)
	r.Get("/streaks/{userID}", habits.GetStreaks)

	penalties := handler.NewPenaltyHandler(pool, catalog, os.Getenv("REALTIME_SERVICE_URL"))
	r.Get("/penalties/{groupID}/eligible", penalties.GetEligible)
	r.Post("/penalties/spin", penalties.Spin)
	r.Get("/penalties/{groupID}", penalties.GetPendingDebts)
	r.Patch("/penalties/{id}/resolve", penalties.ResolveDebt)

	return r
}

func loadCatalog() []model.Punishment {
	path := os.Getenv("PUNISHMENT_CATALOG_PATH")
	if path == "" {
		path = "./punishments.json"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		panic("could not load punishment catalog: " + err.Error())
	}
	var catalog []model.Punishment
	if err := json.Unmarshal(data, &catalog); err != nil {
		panic("invalid punishment catalog: " + err.Error())
	}
	if len(catalog) == 0 {
		panic("punishment catalog is empty")
	}
	return catalog
}
