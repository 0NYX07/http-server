package security

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ValidatePath checks if a requested path is safe and doesn't contain path traversal attempts
func ValidatePath(requestedPath string, baseDir string) (string, error) {
	// Check for obviously malicious patterns
	if strings.Contains(requestedPath, "..") {
		return "", fmt.Errorf("path traversal attempt detected: '..' in path")
	}

	if strings.Contains(requestedPath, "//") {
		return "", fmt.Errorf("invalid path: double slashes detected")
	}

	// Remove leading slash if present
	cleanPath := strings.TrimPrefix(requestedPath, "/")

	// Build the full path
	fullPath := filepath.Join(baseDir, cleanPath)

	// Clean the path to resolve any . or .. components
	fullPath = filepath.Clean(fullPath)

	// Get absolute paths for comparison
	absBasePath, err := filepath.Abs(baseDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve base directory: %w", err)
	}

	absFullPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve requested path: %w", err)
	}

	// Ensure the full path is within the base directory
	if !strings.HasPrefix(absFullPath, absBasePath) {
		return "", fmt.Errorf("path traversal attempt: path escapes base directory")
	}

	return absFullPath, nil
}

// IsPathSafe performs additional safety checks on a path
func IsPathSafe(path string) bool {
	// Check for null bytes
	if strings.Contains(path, "\x00") {
		return false
	}

	// Check for control characters
	for _, ch := range path {
		if ch < 32 && ch != '\t' && ch != '\n' {
			return false
		}
	}

	return true
}

