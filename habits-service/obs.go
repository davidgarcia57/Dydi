package main

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// httpDuration is the request-latency histogram scraped at /metrics.
// Labeling by the chi route pattern (not the raw path) keeps cardinality bounded
// so P95/P99 can be computed per endpoint for the research paper.
var httpDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request latency by route and status.",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"method", "route", "status"},
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

// observability records a structured JSON log line with the request duration and
// feeds the Prometheus latency histogram. Replaces chi's unstructured Logger so
// P95 and per-endpoint error rates are computable from logs even without a
// metrics backend (Render captures stdout).
func observability(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		dur := time.Since(start)
		route := chi.RouteContext(r.Context()).RoutePattern()
		if route == "" {
			route = "unmatched"
		}
		status := ww.Status()
		if status == 0 {
			status = http.StatusOK
		}

		httpDuration.WithLabelValues(r.Method, route, strconv.Itoa(status)).Observe(dur.Seconds())
		logger.Info("request",
			"method", r.Method,
			"route", route,
			"status", status,
			"dur_ms", dur.Milliseconds(),
		)
	})
}
