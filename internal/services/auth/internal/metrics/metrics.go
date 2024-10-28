package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	once     sync.Once
	instance *AuthMetrics
)

type AuthMetrics struct {
	RequestDuration  *prometheus.HistogramVec
	CacheHits        *prometheus.CounterVec
	ActiveRequests   prometheus.Gauge
	ErrorsTotal      *prometheus.CounterVec
	TokenValidations *prometheus.CounterVec
	DatabaseLatency  *prometheus.HistogramVec
}

func NewAuthMetrics(namespace string) *AuthMetrics {
	var metrics *AuthMetrics

	once.Do(func() {
		metrics = &AuthMetrics{
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

			TokenValidations: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: namespace,
					Name:      "token_validations_total",
					Help:      "Number of token validations",
				},
				[]string{"result"},
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

func GetInstance() *AuthMetrics {
	if instance == nil {
		panic("metrics not initialized")
	}
	return instance
}

func (m *AuthMetrics) RecordRequestDuration(method string, status string, duration float64) {
	m.RequestDuration.WithLabelValues(method, status).Observe(duration)
}

func (m *AuthMetrics) RecordCacheHit(cacheType string) {
	m.CacheHits.WithLabelValues(cacheType).Inc()
}

func (m *AuthMetrics) IncActiveRequests() {
	m.ActiveRequests.Inc()
}

func (m *AuthMetrics) DecActiveRequests() {
	m.ActiveRequests.Dec()
}

func (m *AuthMetrics) RecordError(errorType string) {
	m.ErrorsTotal.WithLabelValues(errorType).Inc()
}

func (m *AuthMetrics) RecordTokenValidation(result string) {
	m.TokenValidations.WithLabelValues(result).Inc()
}

func (m *AuthMetrics) RecordDatabaseLatency(operation string, duration float64) {
	m.DatabaseLatency.WithLabelValues(operation).Observe(duration)
}

func (m *AuthMetrics) Close() error {
	return nil
}
