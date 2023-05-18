package uos

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func setupMonitoring() {
	if Config.Monitoring.PortPPROF > 0 {
		go func() {
			Log.Info("starting PPROF web interface")
			pprofMux := http.NewServeMux()

			pprofMux.HandleFunc("/debug/pprof/", pprof.Index)
			pprofMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
			pprofMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
			pprofMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
			pprofMux.HandleFunc("/debug/pprof/trace", pprof.Trace)

			err := http.ListenAndServe(fmt.Sprintf(":%d", Config.Monitoring.PortPPROF), pprofMux)
			if err != nil {
				Log.ErrorObj("profiling web interface stopped", err)
			}
		}()
	}

	if Config.Monitoring.PortMetrics > 0 {
		// initialize metrics registry
		Metrics = newMetricsRegistry()

		go func() {
			Log.Info("starting metrics server")
			metricsMux := http.NewServeMux()

			metricsMux.Handle("/metrics", promhttp.Handler())

			err := http.ListenAndServe(fmt.Sprintf(":%d", Config.Monitoring.PortMetrics), metricsMux)
			if err != nil {
				Log.ErrorObj("profiling web interface stopped", err)
			}
		}()
	}
}

type metricsRegistry struct {
	// used to insert NEW metrics - the prometheus.* objects are already thread-safe
	mutex sync.Mutex

	metrics map[string]int

	counter map[int]prometheus.Counter
	gauge   map[int]prometheus.Gauge
}

// metric IDs - will be assigned during registration
var (
	mStartupTime int

	mRequestCount    int
	mRequestDuration int
	mRequestActive   int
	mRequestFailed   int
	mRequestSlow     int

	mLogMessage        int
	mLogMessageWarning int
	mLogMessageError   int
	mLogMessagePanic   int
)

func newMetricsRegistry() *metricsRegistry {
	registry := metricsRegistry{
		metrics: map[string]int{},

		counter: map[int]prometheus.Counter{},
		gauge:   map[int]prometheus.Gauge{},
	}

	// register standard metrics
	mStartupTime = registry.RegisterGauge(
		"app_start_timestamp_seconds",
		"Timestamp of application start.",
	)
	registry.GaugeCurrentTime(mStartupTime)

	mRequestCount = registry.RegisterCounter(
		"app_http_requests_count",
		"Total number of received HTTP requests.",
	)
	mRequestDuration = registry.RegisterCounter(
		"app_http_requests_duration_ms",
		"Total time spent request processing (status code < 500).",
	)
	mRequestFailed = registry.RegisterCounter(
		"app_http_requests_failed_count",
		"Total number of failed HTTP requests (status code >= 500).",
	)
	mRequestSlow = registry.RegisterCounter(
		"app_http_requests_slow_count",
		"Total number of slow HTTP requests (duration >= 2s).",
	)
	mRequestActive = registry.RegisterGauge(
		"app_http_requests_active_count",
		"Current number of active HTTP request.",
	)

	mLogMessage = registry.RegisterCounter(
		"app_log_messages_count",
		"Number of recorded log messages (level INFO or higher).",
	)
	mLogMessageWarning = registry.RegisterCounter(
		"app_log_messages_warning_count",
		"Number of recorded WARNING log messages.",
	)
	mLogMessageError = registry.RegisterCounter(
		"app_log_messages_error_count",
		"Number of recorded ERROR log messages.",
	)
	mLogMessagePanic = registry.RegisterCounter(
		"app_log_messages_panic_count",
		"Number of recorded PANIC/FATAL log messages.",
	)

	return &registry
}

// Metrics allows to publish application metrics
var Metrics *metricsRegistry

func (m *metricsRegistry) assignMetricID(name string) int {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, ok := m.metrics[name]; !ok {
		// assign new ID
		m.metrics[name] = len(m.metrics) + 1
		return m.metrics[name]
	}

	// metric name already defined
	return 0
}

// RegisterCounter creates a new counter with the given name. Does nothing if the name is already registered.
// Prepends 'app_' and appends '_count' to the name. Returns metric ID - used for actual metric operation.
func (m *metricsRegistry) RegisterCounter(name, help string) int {
	metricID := m.assignMetricID(name)
	if metricID == 0 {
		// already registered
		return 0
	}

	// create
	m.counter[metricID] = promauto.NewCounter(prometheus.CounterOpts{
		Name: name,
		Help: help,
	})

	return metricID
}

// CounterInc increases the counter value.
func (m *metricsRegistry) CounterInc(id int) {
	if m == nil {
		return
	}
	if c, ok := m.counter[id]; ok {
		c.Inc()
	}
}

// CounterIncValue increases the counter value by the specified value. Does nothin if value <= 0.
func (m *metricsRegistry) CounterIncValue(id int, value int64) {
	if m == nil || value <= 0 {
		return
	}
	if c, ok := m.counter[id]; ok {
		c.Inc()
	}
}

// CounterIncCondition increases the counter value if the specified condition is 'true'.
func (m *metricsRegistry) CounterIncCondition(id int, condition bool) {
	if m == nil || !condition {
		return
	}
	if c, ok := m.counter[id]; ok {
		c.Inc()
	}
}

// CounterIncValueCondition increases the counter value by the specified value if the condition is 'true'.
// Does nothin if value <= 0.
func (m *metricsRegistry) CounterIncValueCondition(id int, value int64, condition bool) {
	if m == nil || value <= 0 || !condition {
		return
	}
	if c, ok := m.counter[id]; ok {
		c.Add(float64(value))
	}
}

// RegisterGauge creates a new gauge with the given name. Does nothing if the name is already registered.
// Prepends 'app_' to the name. The name should end with the unit of the value (e.g. '_seconds').
// Returns metric ID - used for actual metric operation.
func (m *metricsRegistry) RegisterGauge(name, help string) int {
	metricID := m.assignMetricID(name)
	if metricID == 0 {
		// already registered
		return 0
	}

	// create
	m.gauge[metricID] = promauto.NewGauge(prometheus.GaugeOpts{
		Name: name,
		Help: help,
	})

	return metricID
}

// GaugeCurrentTime sets the specified gauge to the current time.
func (m *metricsRegistry) GaugeCurrentTime(id int) {
	if m == nil {
		return
	}
	if g, ok := m.gauge[id]; ok {
		g.SetToCurrentTime()
	}
}

// GaugeInc increases the gauge value.
func (m *metricsRegistry) GaugeInc(id int) {
	if m == nil {
		return
	}
	if g, ok := m.gauge[id]; ok {
		g.Inc()
	}
}

// GaugeDec decreases the gauge value.
func (m *metricsRegistry) GaugeDec(id int) {
	if m == nil {
		return
	}
	if g, ok := m.gauge[id]; ok {
		g.Dec()
	}
}
