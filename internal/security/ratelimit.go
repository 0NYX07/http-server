package security

import (
	"sync"
	"time"
)

// RateLimiter implements per-IP rate limiting
type RateLimiter struct {
	mu            sync.RWMutex
	clients       map[string]*clientRecord
	maxRequests   int
	windowSeconds int
	cleanupTicker *time.Ticker
}

// clientRecord tracks request count for a client IP
type clientRecord struct {
	count     int
	timestamp time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxRequestsPerMinute int) *RateLimiter {
	rl := &RateLimiter{
		clients:       make(map[string]*clientRecord),
		maxRequests:   maxRequestsPerMinute,
		windowSeconds: 60,
		cleanupTicker: time.NewTicker(1 * time.Minute),
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// AllowRequest checks if a request from the given IP should be allowed
func (rl *RateLimiter) AllowRequest(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Get or create client record
	record, exists := rl.clients[ip]
	if !exists {
		rl.clients[ip] = &clientRecord{
			count:     1,
			timestamp: now,
		}
		return true
	}

	// Check if window has expired
	if now.Sub(record.timestamp) > time.Duration(rl.windowSeconds)*time.Second {
		record.count = 1
		record.timestamp = now
		return true
	}

	// Check if limit exceeded
	if record.count >= rl.maxRequests {
		return false
	}

	// Increment count
	record.count++
	return true
}

// cleanup removes old client records periodically
func (rl *RateLimiter) cleanup() {
	for range rl.cleanupTicker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, record := range rl.clients {
			if now.Sub(record.timestamp) > 2*time.Duration(rl.windowSeconds)*time.Second {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Stop stops the rate limiter cleanup goroutine
func (rl *RateLimiter) Stop() {
	if rl.cleanupTicker != nil {
		rl.cleanupTicker.Stop()
	}
}

// GetStats returns current rate limiter statistics
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return map[string]interface{}{
		"tracked_ips":     len(rl.clients),
		"max_per_window":  rl.maxRequests,
		"window_seconds":  rl.windowSeconds,
	}
}

