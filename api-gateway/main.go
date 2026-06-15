package main

import (
	"net/http"
	"os"
	"time"

	apimiddleware "github.com/dydi/api-gateway/internal/middleware"
	"github.com/dydi/api-gateway/internal/proxy"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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
		// Despertar todos los servicios en paralelo
		urls := []string{
			os.Getenv("GROUPS_SERVICE_URL"),
			os.Getenv("HABITS_SERVICE_URL"),
			os.Getenv("REALTIME_SERVICE_URL"),
		}

		errChan := make(chan error, len(urls))
		for _, u := range urls {
			go func(serviceUrl string) {
				if serviceUrl == "" {
					errChan <- nil
					return
				}
				// Usamos un timeout alto porque el cold start de Render tarda ~40s
				client := &http.Client{Timeout: 60 * time.Second}
				resp, err := client.Get(serviceUrl + "/health")
				if err == nil {
					resp.Body.Close()
				}
				errChan <- err
			}(u)
		}

		// Esperamos a que todos terminen (con éxito o error)
		for i := 0; i < len(urls); i++ {
			<-errChan
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	r.Route("/api", func(r chi.Router) {
		r.Use(apimiddleware.Auth)
		r.Mount("/users", proxy.To(os.Getenv("GROUPS_SERVICE_URL")))
		r.Mount("/groups", proxy.To(os.Getenv("GROUPS_SERVICE_URL")))
		r.Mount("/proposals", proxy.To(os.Getenv("GROUPS_SERVICE_URL")))
		r.Mount("/habits", proxy.To(os.Getenv("HABITS_SERVICE_URL")))
		r.Mount("/penalties", proxy.To(os.Getenv("HABITS_SERVICE_URL")))
	})

	r.With(apimiddleware.Auth).Mount("/ws", proxy.WebSocket(os.Getenv("REALTIME_SERVICE_URL")))

	return r
}