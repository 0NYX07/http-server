# Multi-threaded HTTP Server

A high-performance HTTP server built from scratch using low-level socket programming in Go. This project implements HTTP/1.1 and HTTP/1.0 protocols with multi-threading, persistent connections, and comprehensive security features.

## Overview


### Security Features
- **Path Traversal Protection**: Robust canonicalization and validation
- **Host Header Validation**: Prevent host header injection attacks
- **Rate Limiting**: Per-IP rate limiting to prevent abuse
- **Input Validation**: Comprehensive request validation and sanitization

## Core Features

**Protocol Support**
- Full HTTP/1.1 and HTTP/1.0 implementation
- GET and POST methods with proper status code handling
- Persistent connections with keep-alive support
- Configurable connection timeouts and request limits

**File Serving**
- Static HTML file serving with proper content-type headers
- Binary file transfer for images and documents
- Memory-efficient streaming for large files (>1MB)
- Support for .html, .txt, .png, .jpg, .jpeg file types

**Concurrency**
- Multi-threaded architecture with configurable worker pool (default: 10 threads)
- Connection queuing when thread pool is saturated
- Thread-safe request handling with proper synchronization
- Handles 700+ concurrent requests per second

**Security**
- Path traversal attack prevention with canonicalization
- Host header validation to prevent injection attacks
- Per-IP rate limiting (configurable, default: 100 req/min)
- Input validation and sanitization

**Monitoring & Logging**
- Real-time metrics endpoint with detailed statistics
- Response time percentiles (P50, P75, P90, P95, P99, P99.9)
- Structured logging with timestamps and thread identification
- Color-coded terminal output for easy debugging

## Requirements

- Go 1.21 or higher
- Unix-like OS (Linux, macOS) or Windows

## Installation

```bash
# Clone the repository
git clone https://github.com/0NYX07/http-server.git
cd http-server

# Build the server
go build -o server cmd/server/main.go

# Run the server
./server
```

## Usage

**Basic Usage:**
```bash
./server                    # Run on localhost:8080 with 10 threads
./server 8000              # Custom port
./server 8000 0.0.0.0      # Custom port and host
./server 8000 0.0.0.0 20   # Custom port, host, and thread pool size
```

**Configuration File:**

You can also configure the server using `config.yaml`:

```yaml
server:
  host: "127.0.0.1"
  port: 8080
  thread_pool_size: 10
  connection_timeout: 30
  max_requests_per_connection: 100

security:
  enable_rate_limiting: true
  rate_limit_per_minute: 100
  enable_host_validation: true
  enable_path_validation: true
```

## Performance

The server has been tested under various load conditions:

| Metric | Value |
|--------|-------|
| Sequential throughput | 116 requests/second |
| Concurrent throughput | 786 requests/second |
| Median response time | 0.144 ms |
| 95th percentile | 1.470 ms |
| 99th percentile | 5.859 ms |
| Success rate | 99.9% under load |

**Response Time Distribution:**
- 50% of requests: < 0.14 ms
- 75% of requests: < 0.27 ms
- 90% of requests: < 0.96 ms
- 95% of requests: < 1.47 ms
- 99% of requests: < 5.86 ms

Detailed performance analysis and benchmarks are available in [PERFORMANCE.md](PERFORMANCE.md).

## Project Structure

```
http-server/
├── cmd/server/main.go           # Application entry point
├── internal/
│   ├── config/                  # Configuration management
│   ├── logger/                  # Structured logging
│   ├── server/                  # HTTP server implementation
│   │   ├── server.go           # Core server and connection handling
│   │   ├── handler.go          # Request routing and handlers
│   │   ├── request.go          # HTTP request parsing
│   │   ├── response.go         # HTTP response building
│   │   └── stream.go           # File streaming for large files
│   ├── security/                # Security features
│   │   ├── path.go             # Path traversal protection
│   │   ├── host.go             # Host validation
│   │   └── ratelimit.go        # Rate limiting
│   └── metrics/                 # Performance metrics
│       ├── metrics.go          # Basic metrics
│       └── metrics_v2.go       # Enhanced metrics with percentiles
├── resources/                   # Static files
│   ├── *.html                  # HTML pages (4 files)
│   ├── *.png                   # PNG images (3 files, one 23MB)
│   ├── *.jpg                   # JPEG images (2 files)
│   ├── *.txt                   # Text files (2 files)
│   ├── test-data.json          # Sample JSON for POST testing
│   └── uploads/                # JSON upload directory
├── test-server.sh              # Integration test suite
├── benchmark.sh                # Performance benchmarks
└── load-test.sh                # Load testing (2000+ requests)
```

## Testing

**Run Integration Tests:**
```bash
./test-server.sh
```

Tests include:
- GET requests for HTML, images, and text files
- POST requests with JSON data
- Error handling (404, 405, 415)
- Security features (path traversal, host validation)
- Keep-alive connections

**Run Performance Benchmarks:**
```bash
./benchmark.sh        # Quick performance test
./load-test.sh        # Load test with 2000+ concurrent requests
```

**Manual Testing:**
```bash
# Test HTML serving
curl http://localhost:8080/

# Test binary file download
curl http://localhost:8080/logo.png -O

# Test JSON POST (inline data)
curl -X POST http://localhost:8080/upload \
  -H "Content-Type: application/json" \
  -d '{"name": "test", "value": 123}'

# Test JSON POST (from file)
curl -X POST http://localhost:8080/upload \
  -H "Content-Type: application/json" \
  -d @resources/test-data.json

# View server metrics
curl http://localhost:8080/__metrics__

# Test keep-alive
curl -v -H "Connection: keep-alive" http://localhost:8080/
```

## HTTP API

**Endpoints:**

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | Serves index.html |
| GET | `/<file>.html` | Serves HTML file |
| GET | `/<file>.txt` | Downloads text file |
| GET | `/<file>.png` | Downloads PNG image |
| GET | `/<file>.jpg` | Downloads JPEG image |
| POST | `/upload` | Accepts JSON data |
| GET | `/__metrics__` | Server statistics |

**Response Status Codes:**

| Code | Description |
|------|-------------|
| 200 | Successful request |
| 201 | Resource created (POST) |
| 400 | Bad request or malformed HTTP |
| 403 | Forbidden (security violation) |
| 404 | Resource not found |
| 405 | Method not allowed (only GET/POST supported) |
| 415 | Unsupported media type |
| 429 | Too many requests (rate limited) |
| 500 | Internal server error |
| 503 | Service unavailable (thread pool full) |

## Metrics Endpoint

Access real-time server statistics at `http://localhost:8080/__metrics__`

**Example Response:**
```json
{
  "server": {
    "uptime_seconds": 42.5,
    "thread_pool_size": 10
  },
  "connections": {
    "active": 1,
    "total": 2052,
    "queued": 0
  },
  "requests": {
    "total": 2052,
    "by_method": {
      "GET": 2037,
      "POST": 15
    }
  },
  "performance": {
    "requests_per_second": 48.28,
    "response_times": {
      "mean_ms": 0.519759,
      "median_ms": 0.144083,
      "p95_ms": 1.470042,
      "p99_ms": 5.859854,
      "min_ms": 0.027750,
      "max_ms": 52.621583,
      "stddev_ms": 2.625645
    }
  }
}
```

## Security Features

**Path Traversal Protection:**
- Validates all file paths before serving
- Blocks requests containing `..`, `./`, or absolute paths
- Ensures all files are within the resources directory
- Returns 403 Forbidden for attempted directory escapes

**Host Header Validation:**
- Checks Host header matches server configuration
- Prevents host header injection attacks
- Returns 403 for mismatched hosts
- Returns 400 for missing Host header

**Rate Limiting:**
- Per-IP request limiting (100 requests/minute by default)
- Automatic cleanup of old rate limit data
- Returns 429 Too Many Requests when limit exceeded
- Configurable via `config.yaml`

## Configuration Options

**Server Settings:**
- `host`: Server bind address (default: 127.0.0.1)
- `port`: Server port (default: 8080)
- `thread_pool_size`: Number of worker threads (default: 10)
- `connection_timeout`: Idle connection timeout in seconds (default: 30)
- `max_requests_per_connection`: Maximum requests per persistent connection (default: 100)

**Security Settings:**
- `enable_rate_limiting`: Enable/disable rate limiting (default: true)
- `rate_limit_per_minute`: Requests allowed per IP per minute (default: 100)
- `enable_host_validation`: Enable/disable host header checks (default: true)
- `enable_path_validation`: Enable/disable path traversal protection (default: true)

**Logging Settings:**
- `level`: Log level - debug, info, warn, error (default: info)
- `colored`: Enable colored terminal output (default: true)

## Architecture

**Connection Flow:**
1. TCP listener accepts incoming connections
2. Connection added to worker queue (channel)
3. Available worker thread picks up connection
4. Request parsed and validated
5. Security checks performed (path, host validation)
6. Request routed to appropriate handler
7. Response generated and sent
8. Connection kept alive or closed based on headers

**Thread Pool Design:**
- Fixed pool of goroutines (workers)
- Lock-free queue using Go channels
- Workers block on channel when idle
- Automatic load distribution
- Graceful degradation when pool saturated

**Memory Management:**
- Small files (<1MB): Loaded directly into memory
- Large files (>1MB): Streamed in 8KB chunks
- Pre-allocated buffers for requests (8KB)
- Efficient connection reuse with keep-alive

## Implementation Details

**HTTP Request Parsing:**
- Manual parsing of request line and headers
- Support for Content-Length and Transfer-Encoding
- Maximum request size: 8192 bytes
- Proper handling of HTTP/1.0 and HTTP/1.1 differences

**Binary File Transfer:**
- Files sent with `application/octet-stream` content type
- `Content-Disposition: attachment` header triggers download
- Proper Content-Length calculation
- Handles various image and text file formats

**Persistent Connections:**
- HTTP/1.1: Keep-alive by default
- HTTP/1.0: Close by default unless requested
- Connection timeout after 30 seconds of inactivity
- Maximum 100 requests per connection

## Known Limitations

- Maximum request size limited to 8KB
- Large files (>1MB) are streamed but still loaded into memory
- Rate limiting data stored in memory (resets on restart)
- Single-process architecture (no clustering support)
- No HTTPS/TLS support
- No response compression

## Future Enhancements

Potential improvements for production deployment:
- HTTPS/TLS support with certificate management
- HTTP/2 support for multiplexing
- Response compression (gzip, brotli)
- Request/response caching with ETag support
- True zero-copy file transfer with sendfile()
- Distributed rate limiting with Redis
- Access log file rotation
- Prometheus metrics export

## Development

**Adding New Endpoints:**
1. Add route check in `internal/server/handler.go`
2. Implement handler function
3. Add tests in `test-server.sh`

**Modifying Configuration:**
1. Update struct in `internal/config/config.go`
2. Add fields to `config.yaml`
3. Update validation logic

**Performance Tuning:**
- Increase thread pool size for high concurrency
- Adjust connection timeout for slow clients
- Modify rate limits based on capacity
- Enable/disable security features as needed

## License

This project is created for educational purposes as part of a university assignment on socket programming and HTTP server implementation.

## Author

Krishna Sahni

## Acknowledgments

Built using Go's standard library for network programming (`net`, `io`, `os`). All HTTP protocol handling, request parsing, and response generation implemented from scratch for educational purposes.
