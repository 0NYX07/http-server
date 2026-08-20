package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration for the HTTP server
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Logging  LoggingConfig  `yaml:"logging"`
	Security SecurityConfig `yaml:"security"`
	Paths    PathsConfig    `yaml:"paths"`
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Host                     string `yaml:"host"`
	Port                     int    `yaml:"port"`
	ThreadPoolSize           int    `yaml:"thread_pool_size"`
	ConnectionTimeout        int    `yaml:"connection_timeout"`
	MaxRequestsPerConnection int    `yaml:"max_requests_per_connection"`
	ReadBufferSize           int    `yaml:"read_buffer_size"`
	ListenQueueSize          int    `yaml:"listen_queue_size"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level           string `yaml:"level"`
	Colored         bool   `yaml:"colored"`
	TimestampFormat string `yaml:"timestamp_format"`
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	EnableRateLimiting   bool `yaml:"enable_rate_limiting"`
	RateLimitPerMinute   int  `yaml:"rate_limit_per_minute"`
	EnableHostValidation bool `yaml:"enable_host_validation"`
	EnablePathValidation bool `yaml:"enable_path_validation"`
}

// PathsConfig holds path-related configuration
type PathsConfig struct {
	ResourcesDir string `yaml:"resources_dir"`
	UploadsDir   string `yaml:"uploads_dir"`
}

// Load reads configuration from file and command-line arguments
func Load(args []string) (*Config, error) {
	// Start with default configuration
	cfg := &Config{
		Server: ServerConfig{
			Host:                     "127.0.0.1",
			Port:                     8080,
			ThreadPoolSize:           10,
			ConnectionTimeout:        30,
			MaxRequestsPerConnection: 100,
			ReadBufferSize:           8192,
			ListenQueueSize:          50,
		},
		Logging: LoggingConfig{
			Level:           "info",
			Colored:         true,
			TimestampFormat: "2006-01-02 15:04:05",
		},
		Security: SecurityConfig{
			EnableRateLimiting:   true,
			RateLimitPerMinute:   100,
			EnableHostValidation: true,
			EnablePathValidation: true,
		},
		Paths: PathsConfig{
			ResourcesDir: "resources",
			UploadsDir:   "resources/uploads",
		},
	}

	// Try to load from config.yaml if it exists
	if _, err := os.Stat("config.yaml"); err == nil {
		data, err := os.ReadFile("config.yaml")
		if err != nil {
			return nil, fmt.Errorf("failed to read config.yaml: %w", err)
		}

		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config.yaml: %w", err)
		}
	}

	// Override with command-line arguments
	// Args format: [port] [host] [thread_pool_size]
	if len(args) > 1 {
		port, err := strconv.Atoi(args[1])
		if err != nil {
			return nil, fmt.Errorf("invalid port: %s", args[1])
		}
		cfg.Server.Port = port
	}

	if len(args) > 2 {
		cfg.Server.Host = args[2]
	}

	if len(args) > 3 {
		threadPoolSize, err := strconv.Atoi(args[3])
		if err != nil {
			return nil, fmt.Errorf("invalid thread pool size: %s", args[3])
		}
		cfg.Server.ThreadPoolSize = threadPoolSize
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid port: %d (must be 1-65535)", c.Server.Port)
	}

	if c.Server.ThreadPoolSize < 1 {
		return fmt.Errorf("invalid thread pool size: %d (must be at least 1)", c.Server.ThreadPoolSize)
	}

	if c.Server.ConnectionTimeout < 1 {
		return fmt.Errorf("invalid connection timeout: %d (must be at least 1)", c.Server.ConnectionTimeout)
	}

	if c.Server.ReadBufferSize < 1024 {
		return fmt.Errorf("invalid read buffer size: %d (must be at least 1024)", c.Server.ReadBufferSize)
	}

	return nil
}

// Address returns the full server address (host:port)
func (c *Config) Address() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

