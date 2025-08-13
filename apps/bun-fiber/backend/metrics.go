package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "code", "route"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route"},
	)
	httpInprogress = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_inprogress_requests",
			Help: "Number of in-progress HTTP requests",
		},
	)
)

func init() {
	// Rejestracja metryk
	prometheus.MustRegister(httpRequestsTotal, httpRequestDuration, httpInprogress)
	collectors.NewGoCollector()
}

func metricsMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		httpInprogress.Inc()

		err := c.Next()

		duration := time.Since(start).Seconds()
		route := c.Route().Path
		code := strconv.Itoa(c.Response().StatusCode())

		httpRequestsTotal.WithLabelValues(c.Method(), code, route).Inc()
		httpRequestDuration.WithLabelValues(c.Method(), route).Observe(duration)

		httpInprogress.Dec()

		return err
	}
	
}

func runMetricsServer() {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":3001", nil)
}