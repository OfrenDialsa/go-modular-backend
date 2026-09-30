package metrics

import "github.com/prometheus/client_golang/prometheus"

func Init() {
	prometheus.MustRegister(
		HTTPRequests,
		HTTPDuration,
		AuthRequests,
		AuthDuration,
	)
}

var HTTPRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total HTTP requests",
	},
	[]string{"method", "endpoint", "status"},
)

var HTTPDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request latency",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"method", "endpoint"},
)
