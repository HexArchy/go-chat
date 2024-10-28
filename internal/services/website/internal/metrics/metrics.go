package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	once     sync.Once
	instance *WebsiteMetrics
)

type WebsiteMetrics struct {
	RequestDuration *prometheus.HistogramVec
	CacheHits       *prometheus.CounterVec
	ActiveRequests  prometheus.Gauge
	ErrorsTotal     *prometheus.CounterVec
	RoomsCreated    prometheus.Counter
	RoomsDeleted    prometheus.Counter
	SearchQueries   prometheus.Counter
	DatabaseLatency *prometheus.HistogramVec
}

func NewWebsiteMetrics(namespace string) *WebsiteMetrics {
	var metrics *WebsiteMetrics

	once.Do(func() {
		metrics = &WebsiteMetrics{
			RequestDuration: promauto.NewHistogramVec(
				prometheus.HistogramOpts{
					Namespace: namespace,
					Name:      "request_duration_seconds",
					Help:      "Time spent processing requests",
					Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
				},
				[]string{"method", "status"},
			),

			CacheHits: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: namespace,
					Name:      "cache_hits_total",
					Help:      "Number of cache hits",
				},
				[]string{"cache_type"},
			),

			ActiveRequests: promauto.NewGauge(
				prometheus.GaugeOpts{
					Namespace: namespace,
					Name:      "active_requests",
					Help:      "Number of requests currently being processed",
				},
			),

			ErrorsTotal: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: namespace,
					Name:      "errors_total",
					Help:      "Total number of errors by type",
				},
				[]string{"type"},
			),

			RoomsCreated: promauto.NewCounter(
				prometheus.CounterOpts{
					Namespace: namespace,
					Name:      "rooms_created_total",
					Help:      "Total number of rooms created",
				},
			),

			RoomsDeleted: promauto.NewCounter(
				prometheus.CounterOpts{
					Namespace: namespace,
					Name:      "rooms_deleted_total",
					Help:      "Total number of rooms deleted",
				},
			),

			SearchQueries: promauto.NewCounter(
				prometheus.CounterOpts{
					Namespace: namespace,
					Name:      "search_queries_total",
					Help:      "Total number of room search queries",
				},
			),

			DatabaseLatency: promauto.NewHistogramVec(
				prometheus.HistogramOpts{
					Namespace: namespace,
					Name:      "database_operation_duration_seconds",
					Help:      "Time spent executing database operations",
					Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
				},
				[]string{"operation"},
			),
		}

		instance = metrics
	})

	return instance
}

func GetInstance() *WebsiteMetrics {
	if instance == nil {
		panic("metrics not initialized")
	}
	return instance
}

func (m *WebsiteMetrics) RecordRequestDuration(method string, status string, duration float64) {
	m.RequestDuration.WithLabelValues(method, status).Observe(duration)
}

func (m *WebsiteMetrics) RecordCacheHit(cacheType string) {
	m.CacheHits.WithLabelValues(cacheType).Inc()
}

func (m *WebsiteMetrics) IncActiveRequests() {
	m.ActiveRequests.Inc()
}

func (m *WebsiteMetrics) DecActiveRequests() {
	m.ActiveRequests.Dec()
}

func (m *WebsiteMetrics) RecordError(errorType string) {
	m.ErrorsTotal.WithLabelValues(errorType).Inc()
}

func (m *WebsiteMetrics) RecordDatabaseLatency(operation string, duration float64) {
	m.DatabaseLatency.WithLabelValues(operation).Observe(duration)
}

func (m *WebsiteMetrics) RecordRoomCreation() {
	m.RoomsCreated.Inc()
}

func (m *WebsiteMetrics) RecordRoomDeletion() {
	m.RoomsDeleted.Inc()
}

func (m *WebsiteMetrics) RecordSearchQuery() {
	m.SearchQueries.Inc()
}
