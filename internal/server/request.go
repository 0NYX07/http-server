package server

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// HTTPRequest represents a parsed HTTP request
type HTTPRequest struct {
	Method      string
	Path        string
	Version     string
	Headers     map[string]string
	Body        []byte
	ContentType string
	Host        string
	Connection  string
	RemoteAddr  string
}

// ParseRequest parses an HTTP request from a reader
func ParseRequest(reader *bufio.Reader, remoteAddr string) (*HTTPRequest, error) {
	req := &HTTPRequest{
		Headers:    make(map[string]string),
		RemoteAddr: remoteAddr,
	}

	// Read request line (e.g., "GET /path HTTP/1.1")
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("connection closed")
		}
		return nil, fmt.Errorf("failed to read request line: %w", err)
	}

	requestLine = strings.TrimSpace(requestLine)
	parts := strings.Split(requestLine, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid request line: %s", requestLine)
	}

	req.Method = parts[0]
	req.Path = parts[1]
	req.Version = parts[2]

	// Validate HTTP version
	if req.Version != "HTTP/1.1" && req.Version != "HTTP/1.0" {
		return nil, fmt.Errorf("unsupported HTTP version: %s", req.Version)
	}

	// Read headers
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read header: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			// Empty line marks end of headers
			break
		}

		// Parse header (e.g., "Content-Type: application/json")
		colonIdx := strings.Index(line, ":")
		if colonIdx == -1 {
			return nil, fmt.Errorf("invalid header line: %s", line)
		}

		headerName := strings.TrimSpace(line[:colonIdx])
		headerValue := strings.TrimSpace(line[colonIdx+1:])

		// Store header (case-insensitive)
		headerNameLower := strings.ToLower(headerName)
		req.Headers[headerNameLower] = headerValue

		// Store commonly used headers directly
		switch headerNameLower {
		case "content-type":
			req.ContentType = headerValue
		case "host":
			req.Host = headerValue
		case "connection":
			req.Connection = strings.ToLower(headerValue)
		}
	}

	// Read body if Content-Length is present
	if contentLengthStr, ok := req.Headers["content-length"]; ok {
		contentLength, err := strconv.Atoi(contentLengthStr)
		if err != nil {
			return nil, fmt.Errorf("invalid content-length: %s", contentLengthStr)
		}

		if contentLength > 0 {
			if contentLength > 8192 {
				return nil, fmt.Errorf("request body too large: %d bytes (max 8192)", contentLength)
			}

			body := make([]byte, contentLength)
			_, err = io.ReadFull(reader, body)
			if err != nil {
				return nil, fmt.Errorf("failed to read request body: %w", err)
			}
			req.Body = body
		}
	}

	return req, nil
}

// IsKeepAlive returns true if the connection should be kept alive
func (r *HTTPRequest) IsKeepAlive() bool {
	// HTTP/1.1 defaults to keep-alive
	if r.Version == "HTTP/1.1" {
		return r.Connection != "close"
	}
	// HTTP/1.0 defaults to close
	return r.Connection == "keep-alive"
}

// String returns a string representation of the request
func (r *HTTPRequest) String() string {
	return fmt.Sprintf("%s %s %s", r.Method, r.Path, r.Version)
}

