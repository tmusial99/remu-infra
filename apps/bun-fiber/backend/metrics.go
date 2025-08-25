package main

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	registry = prometheus.NewRegistry()
	once     sync.Once

	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"host", "method", "code", "route"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"host", "method", "route"},
	)
	httpInprogress = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_inprogress_requests",
			Help: "Number of in-progress HTTP requests",
		},
	)
)

func initMetrics() {
	once.Do(func() {
		registry.MustRegister(
			collectors.NewGoCollector(),
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		)
		registry.MustRegister(httpRequestsTotal, httpRequestDuration, httpInprogress)
	})
}

func sanitizeMethod(m string) string {
	m = strings.ToUpper(m)
	switch m {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS":
		return m
	default:
		return "OTHER"
	}
}

func metricsMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		httpInprogress.Inc()

		err := c.Next()

		host := c.Hostname()
		path := c.Path()
		method := sanitizeMethod(c.Method())
		code := strconv.Itoa(c.Response().StatusCode())
		httpRequestsTotal.WithLabelValues(host, method, code, path).Inc()
		httpRequestDuration.WithLabelValues(host, method, path).Observe(time.Since(start).Seconds())

		httpInprogress.Dec()
		return err
	}
}

func runMetricsServer() {
	mh := promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
	mux := http.NewServeMux()
	mux.Handle("/metrics", mh)
	_ = http.ListenAndServe(":3001", mux)
}
