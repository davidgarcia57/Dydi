package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dydi/groups-service/internal/handler"
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
		port = "8082"
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
		log.Printf("groups-service listening on %s", srv.Addr)
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
	log.Println("groups-service stopped")
}

func setupRouter(pool *pgxpool.Pool) *chi.Mux {
	r := chi.NewRouter()
	r.Use(observability)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Handle("/metrics", promhttp.Handler())

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
