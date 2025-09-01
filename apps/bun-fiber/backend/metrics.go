package main

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
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
			Help: "Total number of HTTP requests.",
		},
		[]string{"host", "method", "code", "route"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"host", "method", "route"},
	)
)

func initMetrics() {
	once.Do(func() {
		registry.MustRegister(
			collectors.NewGoCollector(),
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
			httpRequestsTotal,
			httpRequestDuration,
		)
	})
}

// labels: "<html>", "<static>", "<spa fallback>", "<404>", or raw API path
func routeLabel(raw string, status int, isAPI, servedHTML, servedStatic, staticTry, servedSPA bool) string {
	// reduce accidental variants
	for strings.Contains(raw, "//") {
		raw = strings.ReplaceAll(raw, "//", "/")
	}
	if isAPI {
		return raw
	}
	if servedHTML {
		return "<html>"
	}
	if servedStatic || staticTry {
		return "<static>"
	}
	if servedSPA {
		return "<spa fallback>"
	}
	if status == http.StatusNotFound {
		return "<404>"
	}
	return raw
}

func metricsMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Snapshot BEFORE Next(); copy to avoid Fiber zero-alloc reuse.
		host := utils.CopyString(strings.ToLower(c.Hostname()))
		if host == "" {
			host = "unknown"
		}
		method := utils.CopyString(strings.ToUpper(c.Method()))
		rawPath := utils.CopyString(c.Path())

		err := c.Next()

		code := c.Response().StatusCode()
		isAPI, _ := c.Locals(locIsAPI).(bool)
		servedStatic, _ := c.Locals(locServedStatic).(bool)
		servedHTML, _ := c.Locals(locServedHTML).(bool)
		staticTry, _ := c.Locals(locStaticTry).(bool)
		servedSPA, _ := c.Locals(locSPA).(bool)

		route := routeLabel(rawPath, code, isAPI, servedHTML, servedStatic, staticTry, servedSPA)

		httpRequestsTotal.WithLabelValues(host, method, strconv.Itoa(code), route).Inc()
		httpRequestDuration.WithLabelValues(host, method, route).Observe(time.Since(start).Seconds())

		return err
	}
}

func runMetricsServer() {
	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
		ErrorHandling:     promhttp.ContinueOnError,
	})
	mux := http.NewServeMux()
	mux.Handle("/metrics", handler)
	_ = http.ListenAndServe(":3001", mux)
}
