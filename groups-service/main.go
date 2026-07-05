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
	sharedDb "github.com/dydi/shared/db"
	sharedMiddleware "github.com/dydi/shared/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Reject requests that don't carry the shared gateway↔services secret
	// (see RequireInternalToken). Fail closed: a missing token would otherwise
	// leave every endpoint trusting the X-User-ID header from any caller.
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

const requestTimeout = 15 * time.Second

func setupRouter(pool *pgxpool.Pool) *chi.Mux {
	r := chi.NewRouter()
	r.Use(observability)
	r.Use(middleware.Recoverer)
	// Bound every request so a slow query under load fails fast (504) instead of
	// piling up goroutines and pushing the 512 MB instance toward an OOM kill.
	r.Use(middleware.Timeout(requestTimeout))

	// Public, unauthenticated: health probe + Prometheus scrape.
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

	// Everything else is reachable only through the gateway (or sibling
	// services), proven by the shared internal token.
	r.Group(func(r chi.Router) {
		r.Use(sharedMiddleware.RequireInternalToken)

		u := handler.NewUserHandler(pool)
		r.Post("/users/sync", u.SyncUser)
		r.Delete("/users/me", u.DeleteMe)

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

		// Internal: realtime-service checks membership before accepting a
		// WebSocket, so a logged-in user can't subscribe to a group they don't
		// belong to.
		r.Get("/internal/groups/{groupID}/members/{userID}", h.CheckMembership)
	})

	return r
}
