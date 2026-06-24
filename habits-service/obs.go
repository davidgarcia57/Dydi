package main

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
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

// registerPoolMetrics exposes pgx pool saturation at /metrics so the k6 ramps
// can correlate DB connection pressure with latency and memory — the research
// explicitly tracks "Conexiones de PostgreSQL". The GaugeFuncs read pool.Stat()
// lazily on each scrape, so there is no background cost.
func registerPoolMetrics(pool *pgxpool.Pool) {
	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "db_pool_total_conns",
		Help: "Connections currently open in the pgx pool (acquired + idle).",
	}, func() float64 { return float64(pool.Stat().TotalConns()) })

	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "db_pool_acquired_conns",
		Help: "Connections currently checked out (in use).",
	}, func() float64 { return float64(pool.Stat().AcquiredConns()) })

	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "db_pool_idle_conns",
		Help: "Idle connections available in the pool.",
	}, func() float64 { return float64(pool.Stat().IdleConns()) })

	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "db_pool_max_conns",
		Help: "Configured maximum pool size (DB_MAX_CONNS).",
	}, func() float64 { return float64(pool.Stat().MaxConns()) })

	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "db_pool_empty_acquire_total",
		Help: "Cumulative acquires that had to wait because the pool was exhausted.",
	}, func() float64 { return float64(pool.Stat().EmptyAcquireCount()) })
}
