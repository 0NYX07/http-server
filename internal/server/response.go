package server

import (
	"fmt"
	"time"
)

// HTTPResponse represents an HTTP response
type HTTPResponse struct {
	StatusCode int
	StatusText string
	Headers    map[string]string
	Body       []byte
}

// Status codes and their text
var statusTexts = map[int]string{
	200: "OK",
	201: "Created",
	400: "Bad Request",
	403: "Forbidden",
	404: "Not Found",
	405: "Method Not Allowed",
	415: "Unsupported Media Type",
	429: "Too Many Requests",
	500: "Internal Server Error",
	503: "Service Unavailable",
}

// NewResponse creates a new HTTP response
func NewResponse(statusCode int) *HTTPResponse {
	statusText, ok := statusTexts[statusCode]
	if !ok {
		statusText = "Unknown"
	}

	return &HTTPResponse{
		StatusCode: statusCode,
		StatusText: statusText,
		Headers: map[string]string{
			"Server": "Multi-threaded HTTP Server",
			"Date":   formatHTTPDate(time.Now()),
		},
		Body: nil,
	}
}

// SetHeader sets a response header
func (r *HTTPResponse) SetHeader(key, value string) {
	r.Headers[key] = value
}

// SetBody sets the response body and Content-Length header
func (r *HTTPResponse) SetBody(body []byte) {
	r.Body = body
	r.Headers["Content-Length"] = fmt.Sprintf("%d", len(body))
}

// SetJSON sets the response body as JSON with appropriate headers
func (r *HTTPResponse) SetJSON(body []byte) {
	r.SetBody(body)
	r.Headers["Content-Type"] = "application/json"
}

// SetHTML sets the response body as HTML with appropriate headers
func (r *HTTPResponse) SetHTML(body []byte) {
	r.SetBody(body)
	r.Headers["Content-Type"] = "text/html; charset=utf-8"
}

// SetBinary sets the response body as binary with appropriate headers
func (r *HTTPResponse) SetBinary(body []byte, filename string) {
	r.SetBody(body)
	r.Headers["Content-Type"] = "application/octet-stream"
	r.Headers["Content-Disposition"] = fmt.Sprintf("attachment; filename=\"%s\"", filename)
}

// SetConnectionHeaders sets connection-related headers
func (r *HTTPResponse) SetConnectionHeaders(keepAlive bool, timeout int, maxRequests int) {
	if keepAlive {
		r.Headers["Connection"] = "keep-alive"
		r.Headers["Keep-Alive"] = fmt.Sprintf("timeout=%d, max=%d", timeout, maxRequests)
	} else {
		r.Headers["Connection"] = "close"
	}
}

// Bytes returns the complete HTTP response as bytes
func (r *HTTPResponse) Bytes() []byte {
	// Build status line
	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", r.StatusCode, r.StatusText)

	// Build headers
	headers := ""
	for key, value := range r.Headers {
		headers += fmt.Sprintf("%s: %s\r\n", key, value)
	}

	// Combine status line, headers, and body
	response := statusLine + headers + "\r\n"
	responseBytes := []byte(response)

	if r.Body != nil {
		responseBytes = append(responseBytes, r.Body...)
	}

	return responseBytes
}

// formatHTTPDate formats a time according to RFC 7231
func formatHTTPDate(t time.Time) string {
	return t.UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
}

// Error response helpers

// NewErrorResponse creates a standard error response
func NewErrorResponse(statusCode int, message string) *HTTPResponse {
	resp := NewResponse(statusCode)
	body := fmt.Sprintf(`{"error": "%s", "status": %d}`, message, statusCode)
	resp.SetJSON([]byte(body))
	return resp
}

// NewBadRequestResponse creates a 400 Bad Request response
func NewBadRequestResponse(message string) *HTTPResponse {
	return NewErrorResponse(400, message)
}

// NewForbiddenResponse creates a 403 Forbidden response
func NewForbiddenResponse(message string) *HTTPResponse {
	return NewErrorResponse(403, message)
}

// NewNotFoundResponse creates a 404 Not Found response
func NewNotFoundResponse(message string) *HTTPResponse {
	return NewErrorResponse(404, message)
}

// NewMethodNotAllowedResponse creates a 405 Method Not Allowed response
func NewMethodNotAllowedResponse() *HTTPResponse {
	return NewErrorResponse(405, "Method not allowed. Only GET and POST are supported.")
}

// NewUnsupportedMediaTypeResponse creates a 415 Unsupported Media Type response
func NewUnsupportedMediaTypeResponse(message string) *HTTPResponse {
	return NewErrorResponse(415, message)
}

// NewTooManyRequestsResponse creates a 429 Too Many Requests response
func NewTooManyRequestsResponse() *HTTPResponse {
	resp := NewErrorResponse(429, "Too many requests. Please slow down.")
	resp.SetHeader("Retry-After", "60")
	return resp
}

// NewInternalServerErrorResponse creates a 500 Internal Server Error response
func NewInternalServerErrorResponse(message string) *HTTPResponse {
	return NewErrorResponse(500, message)
}

// NewServiceUnavailableResponse creates a 503 Service Unavailable response
func NewServiceUnavailableResponse() *HTTPResponse {
	resp := NewErrorResponse(503, "Service temporarily unavailable. Thread pool saturated.")
	resp.SetHeader("Retry-After", "10")
	return resp
}

