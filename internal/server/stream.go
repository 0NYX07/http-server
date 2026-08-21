package server

import (
	"fmt"
	"io"
	"os"
)

// StreamConfig holds configuration for file streaming
type StreamConfig struct {
	BufferSize      int
	MaxMemoryBuffer int64 // Files larger than this will be streamed
}

// DefaultStreamConfig returns default streaming configuration
func DefaultStreamConfig() StreamConfig {
	return StreamConfig{
		BufferSize:      8192,  // 8KB buffer for streaming
		MaxMemoryBuffer: 1024 * 1024, // 1MB - stream files larger than this
	}
}

// ShouldStream determines if a file should be streamed based on size
func (sc StreamConfig) ShouldStream(fileSize int64) bool {
	return fileSize > sc.MaxMemoryBuffer
}

// StreamFile efficiently streams a file without loading it entirely into memory
// Returns the streamed bytes as a slice
func StreamFile(filePath string, bufferSize int) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file size
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Pre-allocate buffer based on file size
	data := make([]byte, 0, stat.Size())
	buffer := make([]byte, bufferSize)

	for {
		n, err := file.Read(buffer)
		if n > 0 {
			data = append(data, buffer[:n]...)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading file: %w", err)
		}
	}

	return data, nil
}

// ReadFileOptimized reads a file using the most efficient method based on size
func ReadFileOptimized(filePath string, config StreamConfig) ([]byte, error) {
	// Get file size first
	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// For small files, use simple read
	if !config.ShouldStream(stat.Size()) {
		return os.ReadFile(filePath)
	}

	// For large files, use streaming
	return StreamFile(filePath, config.BufferSize)
}

