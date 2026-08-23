package metrics

import (
	"encoding/json"
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// MetricsV2 holds enhanced server statistics with better timing precision
type MetricsV2 struct {
	StartTime          time.Time
	TotalRequests      atomic.Int64
	TotalResponses     atomic.Int64
	RequestsByMethod   map[string]*atomic.Int64
	ResponsesByStatus  map[int]*atomic.Int64
	TotalBytesServed   atomic.Int64
	ActiveConnections  atomic.Int32
	TotalConnections   atomic.Int64
	QueuedConnections  atomic.Int32
	ThreadPoolSize     int
	mu                 sync.RWMutex
	requestTimes       []float64 // Store response times in microseconds as float
}

var (
	globalMetricsV2 *MetricsV2
	onceV2          sync.Once
)

// InitV2 initializes the enhanced metrics instance
func InitV2(threadPoolSize int) {
	onceV2.Do(func() {
		globalMetricsV2 = &MetricsV2{
			StartTime:         time.Now(),
			RequestsByMethod:  make(map[string]*atomic.Int64),
			ResponsesByStatus: make(map[int]*atomic.Int64),
			ThreadPoolSize:    threadPoolSize,
			requestTimes:      make([]float64, 0, 10000), // Store more samples
		}

		// Initialize method counters
		methods := []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS", "PATCH"}
		for _, method := range methods {
			counter := &atomic.Int64{}
			globalMetricsV2.RequestsByMethod[method] = counter
		}

		// Initialize status code counters
		statusCodes := []int{200, 201, 204, 400, 401, 403, 404, 405, 415, 429, 500, 502, 503, 504}
		for _, code := range statusCodes {
			counter := &atomic.Int64{}
			globalMetricsV2.ResponsesByStatus[code] = counter
		}
	})
}

// GetV2 returns the enhanced metrics instance
func GetV2() *MetricsV2 {
	if globalMetricsV2 == nil {
		InitV2(10)
	}
	return globalMetricsV2
}

// RecordRequest records a new request
func (m *MetricsV2) RecordRequest(method string) {
	m.TotalRequests.Add(1)
	if counter, ok := m.RequestsByMethod[method]; ok {
		counter.Add(1)
	}
}

// RecordResponse records a response with precise timing in nanoseconds
func (m *MetricsV2) RecordResponse(statusCode int, bytesServed int, durationNs int64) {
	m.TotalResponses.Add(1)
	m.TotalBytesServed.Add(int64(bytesServed))

	if counter, ok := m.ResponsesByStatus[statusCode]; ok {
		counter.Add(1)
	}

	// Store response time in milliseconds with microsecond precision
	timeMs := float64(durationNs) / 1_000_000.0

	m.mu.Lock()
	// Keep last 10000 samples for better statistics
	if len(m.requestTimes) >= 10000 {
		m.requestTimes = m.requestTimes[1:]
	}
	m.requestTimes = append(m.requestTimes, timeMs)
	m.mu.Unlock()
}

// IncrementActiveConnections increments active connection count
func (m *MetricsV2) IncrementActiveConnections() {
	m.ActiveConnections.Add(1)
	m.TotalConnections.Add(1)
}

// DecrementActiveConnections decrements active connection count
func (m *MetricsV2) DecrementActiveConnections() {
	m.ActiveConnections.Add(-1)
}

// SetQueuedConnections sets the number of queued connections
func (m *MetricsV2) SetQueuedConnections(count int32) {
	m.QueuedConnections.Store(count)
}

// ResponseTimeStats holds detailed response time statistics
type ResponseTimeStats struct {
	Count   int     `json:"count"`
	Mean    float64 `json:"mean_ms"`
	Median  float64 `json:"median_ms"`
	Min     float64 `json:"min_ms"`
	Max     float64 `json:"max_ms"`
	P50     float64 `json:"p50_ms"`
	P75     float64 `json:"p75_ms"`
	P90     float64 `json:"p90_ms"`
	P95     float64 `json:"p95_ms"`
	P99     float64 `json:"p99_ms"`
	P999    float64 `json:"p999_ms"`
	StdDev  float64 `json:"stddev_ms"`
}

// GetResponseTimeStats calculates detailed response time statistics
func (m *MetricsV2) GetResponseTimeStats() ResponseTimeStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := ResponseTimeStats{}
	
	if len(m.requestTimes) == 0 {
		return stats
	}

	stats.Count = len(m.requestTimes)

	// Create a sorted copy for percentile calculations
	sorted := make([]float64, len(m.requestTimes))
	copy(sorted, m.requestTimes)
	sort.Float64s(sorted)

	// Min and Max
	stats.Min = sorted[0]
	stats.Max = sorted[len(sorted)-1]

	// Mean
	var sum float64
	for _, t := range sorted {
		sum += t
	}
	stats.Mean = sum / float64(len(sorted))

	// Median (P50)
	stats.Median = percentile(sorted, 50)
	stats.P50 = stats.Median

	// Percentiles
	stats.P75 = percentile(sorted, 75)
	stats.P90 = percentile(sorted, 90)
	stats.P95 = percentile(sorted, 95)
	stats.P99 = percentile(sorted, 99)
	stats.P999 = percentile(sorted, 99.9)

	// Standard Deviation
	var sumSquares float64
	for _, t := range sorted {
		diff := t - stats.Mean
		sumSquares += diff * diff
	}
	stats.StdDev = math.Sqrt(sumSquares / float64(len(sorted)))

	return stats
}

// percentile calculates the nth percentile from sorted data
func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if len(sorted) == 1 {
		return sorted[0]
	}
	
	index := (p / 100.0) * float64(len(sorted)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))
	
	if lower == upper {
		return sorted[lower]
	}
	
	// Linear interpolation
	weight := index - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

// GetUptime returns server uptime duration
func (m *MetricsV2) GetUptime() time.Duration {
	return time.Since(m.StartTime)
}

// ToJSON returns enhanced metrics as JSON
func (m *MetricsV2) ToJSON() ([]byte, error) {
	uptime := m.GetUptime()
	responseStats := m.GetResponseTimeStats()

	data := map[string]interface{}{
		"server": map[string]interface{}{
			"uptime_seconds":   uptime.Seconds(),
			"uptime_human":     formatDuration(uptime),
			"started_at":       m.StartTime.Format(time.RFC3339),
			"thread_pool_size": m.ThreadPoolSize,
		},
		"connections": map[string]interface{}{
			"active": m.ActiveConnections.Load(),
			"total":  m.TotalConnections.Load(),
			"queued": m.QueuedConnections.Load(),
		},
		"requests": map[string]interface{}{
			"total": m.TotalRequests.Load(),
			"by_method": map[string]int64{
				"GET":     m.RequestsByMethod["GET"].Load(),
				"POST":    m.RequestsByMethod["POST"].Load(),
				"PUT":     m.RequestsByMethod["PUT"].Load(),
				"DELETE":  m.RequestsByMethod["DELETE"].Load(),
				"HEAD":    m.RequestsByMethod["HEAD"].Load(),
				"OPTIONS": m.RequestsByMethod["OPTIONS"].Load(),
				"PATCH":   m.RequestsByMethod["PATCH"].Load(),
			},
		},
		"responses": map[string]interface{}{
			"total":              m.TotalResponses.Load(),
			"total_bytes_served": m.TotalBytesServed.Load(),
			"by_status": map[string]int64{
				"200": m.ResponsesByStatus[200].Load(),
				"201": m.ResponsesByStatus[201].Load(),
				"204": m.ResponsesByStatus[204].Load(),
				"400": m.ResponsesByStatus[400].Load(),
				"401": m.ResponsesByStatus[401].Load(),
				"403": m.ResponsesByStatus[403].Load(),
				"404": m.ResponsesByStatus[404].Load(),
				"405": m.ResponsesByStatus[405].Load(),
				"415": m.ResponsesByStatus[415].Load(),
				"429": m.ResponsesByStatus[429].Load(),
				"500": m.ResponsesByStatus[500].Load(),
				"502": m.ResponsesByStatus[502].Load(),
				"503": m.ResponsesByStatus[503].Load(),
				"504": m.ResponsesByStatus[504].Load(),
			},
		},
		"performance": map[string]interface{}{
			"requests_per_second": func() float64 {
				if uptime.Seconds() > 0 {
					return float64(m.TotalRequests.Load()) / uptime.Seconds()
				}
				return 0
			}(),
			"throughput_bytes_per_second": func() float64 {
				if uptime.Seconds() > 0 {
					return float64(m.TotalBytesServed.Load()) / uptime.Seconds()
				}
				return 0
			}(),
			"response_times": responseStats,
		},
	}

	return json.MarshalIndent(data, "", "  ")
}

