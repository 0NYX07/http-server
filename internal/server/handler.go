package server

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/0NYX07/http-server/internal/config"
	"github.com/0NYX07/http-server/internal/logger"
	"github.com/0NYX07/http-server/internal/metrics"
	"github.com/0NYX07/http-server/internal/security"
)

// Handler handles HTTP requests
type Handler struct {
	config       *config.Config
	log          *logger.Logger
	rateLimiter  *security.RateLimiter
	streamConfig StreamConfig
}

// NewHandler creates a new request handler
func NewHandler(cfg *config.Config) *Handler {
	var rateLimiter *security.RateLimiter
	if cfg.Security.EnableRateLimiting {
		rateLimiter = security.NewRateLimiter(cfg.Security.RateLimitPerMinute)
	}

	return &Handler{
		config:       cfg,
		log:          logger.GetLogger(),
		rateLimiter:  rateLimiter,
		streamConfig: DefaultStreamConfig(),
	}
}

// HandleRequest routes and handles an HTTP request
func (h *Handler) HandleRequest(req *HTTPRequest, threadName string) *HTTPResponse {
	// Record request in metrics V2
	metrics.GetV2().RecordRequest(req.Method)

	// Check rate limit
	if h.rateLimiter != nil {
		// Extract IP from RemoteAddr (remove port)
		ip := strings.Split(req.RemoteAddr, ":")[0]
		if !h.rateLimiter.AllowRequest(ip) {
			h.log.WarnThread(threadName, "Rate limit exceeded for IP: %s", ip)
			return NewTooManyRequestsResponse()
		}
	}

	// Validate HTTP method
	if req.Method != "GET" && req.Method != "POST" {
		h.log.WarnThread(threadName, "Method not allowed: %s", req.Method)
		return NewMethodNotAllowedResponse()
	}

	// Validate Host header
	if h.config.Security.EnableHostValidation {
		err := security.ValidateHost(req.Host, h.config.Server.Host, h.config.Server.Port)
		if err != nil {
			h.log.WarnThread(threadName, "Host validation failed: %v", err)
			if req.Host == "" {
				return NewBadRequestResponse("Missing Host header")
			}
			return NewForbiddenResponse(fmt.Sprintf("Invalid Host header: %s", req.Host))
		}
		h.log.DebugThread(threadName, "Host validation: %s ✓", req.Host)
	}

	// Route request based on method
	switch req.Method {
	case "GET":
		return h.handleGET(req, threadName)
	case "POST":
		return h.handlePOST(req, threadName)
	default:
		return NewMethodNotAllowedResponse()
	}
}

// handleGET handles GET requests
func (h *Handler) handleGET(req *HTTPRequest, threadName string) *HTTPResponse {
	// Special endpoint for metrics
	if req.Path == "/__metrics__" {
		return h.handleMetrics(threadName)
	}

	// Resolve the file path
	filePath := req.Path

	// Default to index.html for root path
	if filePath == "/" {
		filePath = "/index.html"
	}

	// Validate path safety (check for malicious patterns)
	if h.config.Security.EnablePathValidation {
		if !security.IsPathSafe(filePath) {
			h.log.WarnThread(threadName, "Unsafe path detected: %s", filePath)
			return NewForbiddenResponse("Invalid path characters")
		}

		// Validate path doesn't escape base directory
		fullPath, err := security.ValidatePath(filePath, h.config.Paths.ResourcesDir)
		if err != nil {
			h.log.WarnThread(threadName, "Path validation failed for %s: %v", filePath, err)
			return NewForbiddenResponse("Path traversal attempt detected")
		}

		// Use the validated path
		return h.serveFile(fullPath, filePath, threadName)
	}

	// Fallback without validation (if disabled in config)
	fullPath := filepath.Join(h.config.Paths.ResourcesDir, strings.TrimPrefix(filePath, "/"))
	return h.serveFile(fullPath, filePath, threadName)
}

// serveFile serves a file from the filesystem
func (h *Handler) serveFile(fullPath string, requestedPath string, threadName string) *HTTPResponse {
	// Check if file exists
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			h.log.WarnThread(threadName, "File not found: %s", fullPath)
			return NewNotFoundResponse(fmt.Sprintf("Resource not found: %s", requestedPath))
		}
		h.log.ErrorThread(threadName, "Error accessing file %s: %v", fullPath, err)
		return NewInternalServerErrorResponse("Failed to access file")
	}

	// Don't serve directories
	if fileInfo.IsDir() {
		h.log.WarnThread(threadName, "Attempted to access directory: %s", fullPath)
		return NewForbiddenResponse("Directory listing not allowed")
	}

	// Read file optimized with streaming for large files
	fileData, err := ReadFileOptimized(fullPath, h.streamConfig)
	if err != nil {
		h.log.ErrorThread(threadName, "Error reading file %s: %v", fullPath, err)
		return NewInternalServerErrorResponse("Failed to read file")
	}

	// Log if file was streamed
	if h.streamConfig.ShouldStream(fileInfo.Size()) {
		h.log.DebugThread(threadName, "Streamed large file: %s (%d bytes)", requestedPath, fileInfo.Size())
	}

	// Determine content type based on file extension
	ext := strings.ToLower(filepath.Ext(requestedPath))
	resp := NewResponse(200)

	switch ext {
	case ".html", ".htm":
		// Serve HTML files for rendering in browser
		resp.SetHTML(fileData)
		h.log.InfoThread(threadName, "Serving HTML file: %s (%d bytes)", requestedPath, len(fileData))

	case ".json":
		// Serve JSON files
		resp.SetHeader("Content-Type", "application/json")
		resp.SetBody(fileData)
		h.log.InfoThread(threadName, "Serving JSON file: %s (%d bytes)", requestedPath, len(fileData))

	case ".txt", ".png", ".jpg", ".jpeg":
		// Serve as binary download
		filename := filepath.Base(requestedPath)
		resp.SetBinary(fileData, filename)
		h.log.InfoThread(threadName, "Sending binary file: %s (%d bytes)", filename, len(fileData))

	default:
		// Unsupported file type
		h.log.WarnThread(threadName, "Unsupported file type: %s", ext)
		return NewUnsupportedMediaTypeResponse(fmt.Sprintf("File type %s is not supported", ext))
	}

	return resp
}

// handlePOST handles POST requests
func (h *Handler) handlePOST(req *HTTPRequest, threadName string) *HTTPResponse {
	// Validate Content-Type
	if req.ContentType == "" {
		h.log.WarnThread(threadName, "POST request missing Content-Type header")
		return NewBadRequestResponse("Missing Content-Type header")
	}

	// Only accept application/json
	if !strings.Contains(strings.ToLower(req.ContentType), "application/json") {
		h.log.WarnThread(threadName, "POST request with unsupported Content-Type: %s", req.ContentType)
		return NewUnsupportedMediaTypeResponse("Only application/json is supported for POST requests")
	}

	// Validate that body is present
	if len(req.Body) == 0 {
		h.log.WarnThread(threadName, "POST request with empty body")
		return NewBadRequestResponse("Request body is empty")
	}

	// Validate JSON
	var jsonData interface{}
	if err := json.Unmarshal(req.Body, &jsonData); err != nil {
		h.log.WarnThread(threadName, "Invalid JSON in POST request: %v", err)
		return NewBadRequestResponse(fmt.Sprintf("Invalid JSON: %v", err))
	}

	// Generate filename with timestamp and random ID
	timestamp := time.Now().Format("20060102_150405")
	randomID := generateRandomID(4)
	filename := fmt.Sprintf("upload_%s_%s.json", timestamp, randomID)
	filePath := filepath.Join(h.config.Paths.UploadsDir, filename)

	// Write JSON to file
	if err := os.WriteFile(filePath, req.Body, 0644); err != nil {
		h.log.ErrorThread(threadName, "Failed to write upload file %s: %v", filePath, err)
		return NewInternalServerErrorResponse("Failed to save uploaded data")
	}

	h.log.InfoThread(threadName, "Created upload file: %s (%d bytes)", filename, len(req.Body))

	// Create success response
	response := map[string]string{
		"status":   "success",
		"message":  "File created successfully",
		"filepath": "/uploads/" + filename,
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		h.log.ErrorThread(threadName, "Failed to marshal response: %v", err)
		return NewInternalServerErrorResponse("Failed to create response")
	}

	resp := NewResponse(201)
	resp.SetJSON(responseJSON)
	return resp
}

// generateRandomID generates a random alphanumeric ID of specified length
func generateRandomID(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

// handleMetrics returns server metrics as JSON
func (h *Handler) handleMetrics(threadName string) *HTTPResponse {
	h.log.InfoThread(threadName, "Serving enhanced metrics endpoint")

	metricsJSON, err := metrics.GetV2().ToJSON()
	if err != nil {
		h.log.ErrorThread(threadName, "Failed to generate metrics: %v", err)
		return NewInternalServerErrorResponse("Failed to generate metrics")
	}

	resp := NewResponse(200)
	resp.SetJSON(metricsJSON)
	return resp
}
