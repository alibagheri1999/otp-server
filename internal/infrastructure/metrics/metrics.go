package metrics

import (
	"context"
	"net/http"
	"sync"
	"time"

	"otp-server/internal/infrastructure/logger"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metric struct {
	Name      string
	Value     float64
	Labels    map[string]string
	Timestamp time.Time
	Type      string
}

type MetricsService struct {
	logger    logger.Logger
	startTime time.Time
	metrics   map[string]*Metric
	mu        sync.RWMutex

	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	otpOperationsTotal   *prometheus.CounterVec
	userOperationsTotal  *prometheus.CounterVec
	rateLimitExceeded    *prometheus.CounterVec
	cacheOperationsTotal *prometheus.CounterVec
}

func NewMetricsService(logger logger.Logger) *MetricsService {
	httpRequestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status_code"},
	)

	httpRequestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	otpOperationsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "otp_operations_total",
			Help: "Total number of OTP operations",
		},
		[]string{"operation", "success"},
	)

	userOperationsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_operations_total",
			Help: "Total number of user operations",
		},
		[]string{"operation"},
	)

	rateLimitExceeded := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_exceeded_total",
			Help: "Total number of rate limit violations",
		},
		[]string{"endpoint_type"},
	)

	cacheOperationsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_operations_total",
			Help: "Total number of cache operations",
		},
		[]string{"cache_type", "result"},
	)

	prometheus.MustRegister(httpRequestsTotal, httpRequestDuration, otpOperationsTotal, userOperationsTotal, rateLimitExceeded, cacheOperationsTotal)

	return &MetricsService{
		logger:    logger,
		startTime: time.Now(),
		metrics:   make(map[string]*Metric),

		httpRequestsTotal:    httpRequestsTotal,
		httpRequestDuration:  httpRequestDuration,
		otpOperationsTotal:   otpOperationsTotal,
		userOperationsTotal:  userOperationsTotal,
		rateLimitExceeded:    rateLimitExceeded,
		cacheOperationsTotal: cacheOperationsTotal,
	}
}

func (m *MetricsService) recordMetric(name string, value float64, labels map[string]string, metricType string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	metric := &Metric{
		Name:      name,
		Value:     value,
		Labels:    labels,
		Timestamp: time.Now(),
		Type:      metricType,
	}

	m.metrics[name] = metric

	m.logger.Info(context.Background(), "Metric recorded",
		logger.F("name", name),
		logger.F("value", value),
		logger.F("labels", labels),
		logger.F("type", metricType),
	)
}

func (m *MetricsService) RecordRequest(method, path string, statusCode int, duration time.Duration) {
	labels := map[string]string{
		"method":      method,
		"path":        path,
		"status_code": string(rune(statusCode)),
	}

	m.recordMetric("http_requests_total", 1, labels, "counter")
	m.recordMetric("http_request_duration_ms", float64(duration.Milliseconds()), labels, "histogram")

	m.httpRequestsTotal.WithLabelValues(method, path, string(rune(statusCode))).Inc()
	m.httpRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}

func (m *MetricsService) RecordOTPGenerated(phoneNumber string) {
	labels := map[string]string{
		"operation": "generate",
	}
	m.recordMetric("otp_operations_total", 1, labels, "counter")

	m.otpOperationsTotal.WithLabelValues("generate", "true").Inc()
}

func (m *MetricsService) RecordOTPVerified(phoneNumber string, success bool) {
	labels := map[string]string{
		"operation": "verify",
		"success":   string(rune(map[bool]int{true: 1, false: 0}[success])),
	}
	m.recordMetric("otp_operations_total", 1, labels, "counter")

	successStr := "true"
	if !success {
		successStr = "false"
	}
	m.otpOperationsTotal.WithLabelValues("verify", successStr).Inc()
}

func (m *MetricsService) RecordUserRegistration(userID int, phoneNumber string) {
	labels := map[string]string{
		"operation": "register",
	}
	m.recordMetric("user_operations_total", 1, labels, "counter")

	m.userOperationsTotal.WithLabelValues("register").Inc()
}

func (m *MetricsService) RecordUserLogin(userID int, phoneNumber string) {
	labels := map[string]string{
		"operation": "login",
	}
	m.recordMetric("user_operations_total", 1, labels, "counter")

	m.userOperationsTotal.WithLabelValues("login").Inc()
}

func (m *MetricsService) RecordRateLimitExceeded(endpointType, identifier string) {
	labels := map[string]string{
		"endpoint_type": endpointType,
	}
	m.recordMetric("rate_limit_exceeded_total", 1, labels, "counter")

	m.rateLimitExceeded.WithLabelValues(endpointType).Inc()
}

func (m *MetricsService) RecordCacheHit(cacheType, key string) {
	labels := map[string]string{
		"cache_type": cacheType,
		"result":     "hit",
	}
	m.recordMetric("cache_operations_total", 1, labels, "counter")

	m.cacheOperationsTotal.WithLabelValues(cacheType, "hit").Inc()
}

func (m *MetricsService) RecordCacheMiss(cacheType, key string) {
	labels := map[string]string{
		"cache_type": cacheType,
		"result":     "miss",
	}
	m.recordMetric("cache_operations_total", 1, labels, "counter")

	
	m.cacheOperationsTotal.WithLabelValues(cacheType, "miss").Inc()
}

func (m *MetricsService) GetUptime() time.Duration {
	return time.Since(m.startTime)
}

func (m *MetricsService) GetHealthStatus() map[string]interface{} {
	return map[string]interface{}{
		"status":      "healthy",
		"uptime":      m.GetUptime().String(),
		"start_time":  m.startTime,
		"timestamp":   time.Now(),
		"service":     "otp-server",
		"environment": "development",
	}
}

func (m *MetricsService) GetMetrics() map[string]*Metric {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*Metric)
	for k, v := range m.metrics {
		result[k] = v
	}
	return result
}

func (m *MetricsService) GetMetric(name string) *Metric {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.metrics[name]
}

// GetPrometheusHandler returns HTTP handler for Prometheus metrics
func (m *MetricsService) GetPrometheusHandler() http.Handler {
	return promhttp.Handler()
}
