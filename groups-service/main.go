package main

import (
	"context"
	"net/http"
	"os"

	"github.com/dydi/groups-service/internal/handler"
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
		port = "8082"
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

	u := handler.NewUserHandler(pool)
	r.Post("/users/sync", u.SyncUser)

	h := handler.NewGroupHandler(pool)
	r.Get("/groups", h.ListMyGroups)
	r.Post("/groups", h.CreateGroup)
	r.Get("/groups/{id}", h.GetGroup)
	r.Post("/groups/{id}/join", h.JoinGroup)
	r.Get("/groups/{id}/members", h.ListMembers)
	r.Delete("/groups/{id}/leave", h.LeaveGroup)

	p := handler.NewProposalHandler(pool, os.Getenv("HABITS_SERVICE_URL"))
	r.Post("/groups/{groupID}/proposals", p.CreateProposal)
	r.Get("/groups/{groupID}/proposals", p.ListProposals)
	r.Post("/proposals/{proposalID}/vote", p.Vote)

	return r
}
