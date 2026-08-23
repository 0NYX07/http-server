package server

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/0NYX07/http-server/internal/config"
	"github.com/0NYX07/http-server/internal/logger"
	"github.com/0NYX07/http-server/internal/metrics"
)

// Server represents the HTTP server
type Server struct {
	config     *config.Config
	listener   net.Listener
	handler    *Handler
	workerPool chan net.Conn
	wg         sync.WaitGroup
	shutdown   atomic.Bool
	log        *logger.Logger
}

// New creates a new HTTP server
func New(cfg *config.Config) (*Server, error) {
	// Ensure directories exist
	if err := ensureDirectories(cfg); err != nil {
		return nil, fmt.Errorf("failed to create directories: %w", err)
	}

	// Initialize enhanced metrics with better precision
	metrics.InitV2(cfg.Server.ThreadPoolSize)

	// Create handler
	handler := NewHandler(cfg)

	// Create server
	srv := &Server{
		config:     cfg,
		handler:    handler,
		workerPool: make(chan net.Conn, cfg.Server.ThreadPoolSize*2),
		log:        logger.GetLogger(),
	}

	return srv, nil
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Create TCP listener using low-level socket
	addr := s.config.Address()
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to bind to %s: %w", addr, err)
	}
	s.listener = listener

	// Log startup information
	s.log.Info("HTTP Server started on http://%s", addr)
	s.log.Info("Thread pool size: %d", s.config.Server.ThreadPoolSize)
	s.log.Info("Serving files from '%s' directory", s.config.Paths.ResourcesDir)
	s.log.Info("Press Ctrl+C to stop the server")

	// Start worker goroutines (thread pool)
	for i := 0; i < s.config.Server.ThreadPoolSize; i++ {
		s.wg.Add(1)
		go s.worker(i + 1)
	}

	// Accept connections
	connectionID := 0
	for {
		conn, err := listener.Accept()
		if err != nil {
			if s.shutdown.Load() {
				// Server is shutting down
				break
			}
			s.log.Error("Failed to accept connection: %v", err)
			continue
		}

		connectionID++
		s.log.Debug("Connection accepted from %s (ID: %d)", conn.RemoteAddr(), connectionID)

		// Try to send connection to worker pool
		select {
		case s.workerPool <- conn:
			// Connection queued successfully
		default:
			// Worker pool is full
			s.log.Warn("Thread pool saturated, queuing connection")
			// Block until a worker is available
			s.workerPool <- conn
			s.log.Info("Connection dequeued, assigning to worker")
		}
	}

	return nil
}

// Stop gracefully stops the server
func (s *Server) Stop() {
	s.shutdown.Store(true)

	// Close listener to stop accepting new connections
	if s.listener != nil {
		s.listener.Close()
	}

	// Close worker pool channel
	close(s.workerPool)

	// Wait for all workers to finish
	s.wg.Wait()

	s.log.Info("All connections closed")
}

// worker processes connections from the worker pool
func (s *Server) worker(workerID int) {
	defer s.wg.Done()

	threadName := fmt.Sprintf("Thread-%d", workerID)
	s.log.Debug("%s started", threadName)

	for conn := range s.workerPool {
		s.handleConnection(conn, threadName)
	}

	s.log.Debug("%s stopped", threadName)
}

// handleConnection handles a single client connection
func (s *Server) handleConnection(conn net.Conn, threadName string) {
	// Track connection metrics with V2
	metrics.GetV2().IncrementActiveConnections()
	defer metrics.GetV2().DecrementActiveConnections()

	defer func() {
		if r := recover(); r != nil {
			s.log.ErrorThread(threadName, "Panic recovered: %v", r)
		}
		conn.Close()
	}()

	remoteAddr := conn.RemoteAddr().String()
	s.log.InfoThread(threadName, "Connection from %s", remoteAddr)

	reader := bufio.NewReader(conn)
	requestCount := 0
	maxRequests := s.config.Server.MaxRequestsPerConnection

	for {
		// Set read timeout
		conn.SetReadDeadline(time.Now().Add(time.Duration(s.config.Server.ConnectionTimeout) * time.Second))

		// Parse request
		startTime := time.Now()
		req, err := ParseRequest(reader, remoteAddr)
		if err != nil {
			if err.Error() == "connection closed" {
				s.log.DebugThread(threadName, "Client closed connection")
			} else {
				s.log.ErrorThread(threadName, "Failed to parse request: %v", err)
				// Send 400 Bad Request
				resp := NewBadRequestResponse("Malformed HTTP request")
				conn.Write(resp.Bytes())
			}
			break
		}

		requestCount++
		s.log.InfoThread(threadName, "Request: %s", req.String())

		// Handle request
		resp := s.handler.HandleRequest(req, threadName)

		// Set connection headers
		keepAlive := req.IsKeepAlive() && requestCount < maxRequests
		resp.SetConnectionHeaders(keepAlive, s.config.Server.ConnectionTimeout, maxRequests-requestCount)

		// Send response
		responseBytes := resp.Bytes()
		_, err = conn.Write(responseBytes)
		if err != nil {
			s.log.ErrorThread(threadName, "Failed to send response: %v", err)
			break
		}

		// Log response and record metrics with nanosecond precision
		durationNs := time.Since(startTime).Nanoseconds()
		durationMs := float64(durationNs) / 1_000_000.0
		s.log.InfoThread(threadName, "Response: %d %s (%d bytes, %.3fms)",
			resp.StatusCode, resp.StatusText, len(responseBytes), durationMs)

		// Record response metrics with nanosecond precision
		metrics.GetV2().RecordResponse(resp.StatusCode, len(responseBytes), durationNs)

		// Check if we should keep the connection alive
		if !keepAlive {
			s.log.InfoThread(threadName, "Connection: close")
			break
		}

		s.log.DebugThread(threadName, "Connection: keep-alive (request %d/%d)", requestCount, maxRequests)

		// Check if maximum requests reached
		if requestCount >= maxRequests {
			s.log.InfoThread(threadName, "Maximum requests reached (%d), closing connection", maxRequests)
			break
		}
	}

	s.log.DebugThread(threadName, "Connection closed")
}

// ensureDirectories creates necessary directories if they don't exist
func ensureDirectories(cfg *config.Config) error {
	dirs := []string{
		cfg.Paths.ResourcesDir,
		cfg.Paths.UploadsDir,
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
		}
	}

	return nil
}
