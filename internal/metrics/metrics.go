package metrics

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Metrics holds server statistics
type Metrics struct {
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
	requestTimes       []int64 // Store response times in milliseconds
}

var (
	globalMetrics *Metrics
	once          sync.Once
)

// Init initializes the global metrics instance
func Init(threadPoolSize int) {
	once.Do(func() {
		globalMetrics = &Metrics{
			StartTime:         time.Now(),
			RequestsByMethod:  make(map[string]*atomic.Int64),
			ResponsesByStatus: make(map[int]*atomic.Int64),
			ThreadPoolSize:    threadPoolSize,
			requestTimes:      make([]int64, 0, 1000),
		}

		// Initialize method counters
		methods := []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"}
		for _, method := range methods {
			counter := &atomic.Int64{}
			globalMetrics.RequestsByMethod[method] = counter
		}

		// Initialize status code counters
		statusCodes := []int{200, 201, 400, 403, 404, 405, 415, 429, 500, 503}
		for _, code := range statusCodes {
			counter := &atomic.Int64{}
			globalMetrics.ResponsesByStatus[code] = counter
		}
	})
}

// Get returns the global metrics instance
func Get() *Metrics {
	if globalMetrics == nil {
		Init(10)
	}
	return globalMetrics
}

// RecordRequest records a new request
func (m *Metrics) RecordRequest(method string) {
	m.TotalRequests.Add(1)

	if counter, ok := m.RequestsByMethod[method]; ok {
		counter.Add(1)
	}
}

// RecordResponse records a response with status code and size
func (m *Metrics) RecordResponse(statusCode int, bytesServed int, responseTime int64) {
	m.TotalResponses.Add(1)
	m.TotalBytesServed.Add(int64(bytesServed))

	if counter, ok := m.ResponsesByStatus[statusCode]; ok {
		counter.Add(1)
	}

	// Store response time (keep last 1000)
	m.mu.Lock()
	if len(m.requestTimes) >= 1000 {
		m.requestTimes = m.requestTimes[1:]
	}
	m.requestTimes = append(m.requestTimes, responseTime)
	m.mu.Unlock()
}

// IncrementActiveConnections increments active connection count
func (m *Metrics) IncrementActiveConnections() {
	m.ActiveConnections.Add(1)
	m.TotalConnections.Add(1)
}

// DecrementActiveConnections decrements active connection count
func (m *Metrics) DecrementActiveConnections() {
	m.ActiveConnections.Add(-1)
}

// SetQueuedConnections sets the number of queued connections
func (m *Metrics) SetQueuedConnections(count int32) {
	m.QueuedConnections.Store(count)
}

// GetAverageResponseTime calculates average response time
func (m *Metrics) GetAverageResponseTime() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.requestTimes) == 0 {
		return 0
	}

	var sum int64
	for _, t := range m.requestTimes {
		sum += t
	}
	return sum / int64(len(m.requestTimes))
}

// GetUptime returns server uptime duration
func (m *Metrics) GetUptime() time.Duration {
	return time.Since(m.StartTime)
}

// ToJSON returns metrics as JSON
func (m *Metrics) ToJSON() ([]byte, error) {
	uptime := m.GetUptime()

	data := map[string]interface{}{
		"server": map[string]interface{}{
			"uptime_seconds":   int64(uptime.Seconds()),
			"uptime_human":     formatDuration(uptime),
			"started_at":       m.StartTime.Format(time.RFC3339),
			"thread_pool_size": m.ThreadPoolSize,
		},
		"connections": map[string]interface{}{
			"active":  m.ActiveConnections.Load(),
			"total":   m.TotalConnections.Load(),
			"queued":  m.QueuedConnections.Load(),
		},
		"requests": map[string]interface{}{
			"total": m.TotalRequests.Load(),
			"by_method": map[string]int64{
				"GET":    m.RequestsByMethod["GET"].Load(),
				"POST":   m.RequestsByMethod["POST"].Load(),
				"PUT":    m.RequestsByMethod["PUT"].Load(),
				"DELETE": m.RequestsByMethod["DELETE"].Load(),
			},
		},
		"responses": map[string]interface{}{
			"total":              m.TotalResponses.Load(),
			"total_bytes_served": m.TotalBytesServed.Load(),
			"by_status": map[string]int64{
				"200": m.ResponsesByStatus[200].Load(),
				"201": m.ResponsesByStatus[201].Load(),
				"400": m.ResponsesByStatus[400].Load(),
				"403": m.ResponsesByStatus[403].Load(),
				"404": m.ResponsesByStatus[404].Load(),
				"405": m.ResponsesByStatus[405].Load(),
				"415": m.ResponsesByStatus[415].Load(),
				"429": m.ResponsesByStatus[429].Load(),
				"500": m.ResponsesByStatus[500].Load(),
				"503": m.ResponsesByStatus[503].Load(),
			},
		},
		"performance": map[string]interface{}{
			"avg_response_time_ms": m.GetAverageResponseTime(),
			"requests_per_second":  float64(m.TotalRequests.Load()) / uptime.Seconds(),
		},
	}

	return json.MarshalIndent(data, "", "  ")
}

// formatDuration formats a duration in human-readable format
func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

