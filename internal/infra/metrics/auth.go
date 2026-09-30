package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var AuthRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "auth_requests_total",
		Help: "Total auth requests",
	},
	[]string{"operation", "status"},
)

var AuthDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "auth_duration_seconds",
		Help:    "Auth operation duration",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"operation"},
)

func TrackAuth(operation string, status string, duration time.Duration) {
	AuthRequests.WithLabelValues(operation, status).Inc()
	AuthDuration.WithLabelValues(operation).Observe(duration.Seconds())
}
