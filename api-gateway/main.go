package main

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	apimiddleware "github.com/dydi/api-gateway/internal/middleware"
	"github.com/dydi/api-gateway/internal/proxy"
)

func main() {
	r := setupRouter()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.ListenAndServe(":"+port, r)
}

func setupRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(apimiddleware.CORS)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	r.Route("/api", func(r chi.Router) {
		r.Use(apimiddleware.Auth)
		r.Mount("/groups", proxy.To(os.Getenv("GROUPS_SERVICE_URL")))
		r.Mount("/habits", proxy.To(os.Getenv("HABITS_SERVICE_URL")))
		r.Mount("/penalties", proxy.To(os.Getenv("HABITS_SERVICE_URL")))
	})

	r.With(apimiddleware.Auth).Mount("/ws", proxy.WebSocket(os.Getenv("REALTIME_SERVICE_URL")))

	return r
}
