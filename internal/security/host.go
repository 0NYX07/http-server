package security

import (
	"fmt"
	"strings"
)

// ValidateHost checks if the Host header matches the expected server address
func ValidateHost(hostHeader string, serverHost string, serverPort int) error {
	if hostHeader == "" {
		return fmt.Errorf("missing Host header")
	}

	// Normalize the host header (remove any trailing whitespace)
	hostHeader = strings.TrimSpace(strings.ToLower(hostHeader))

	// Build list of valid host values
	validHosts := []string{
		fmt.Sprintf("%s:%d", serverHost, serverPort),
		fmt.Sprintf("localhost:%d", serverPort),
		fmt.Sprintf("127.0.0.1:%d", serverPort),
	}

	// Also accept without port if it's default (80) or common dev ports
	if serverPort == 80 || serverPort == 8080 {
		validHosts = append(validHosts, serverHost, "localhost", "127.0.0.1")
	}

	// Check if the host header matches any valid host
	for _, validHost := range validHosts {
		if hostHeader == strings.ToLower(validHost) {
			return nil
		}
	}

	return fmt.Errorf("host header mismatch: got '%s', expected one of %v", hostHeader, validHosts)
}

// IsValidHostFormat checks if a host string has a valid format
func IsValidHostFormat(host string) bool {
	if host == "" {
		return false
	}

	// Check for invalid characters
	invalidChars := []string{"\r", "\n", "\x00", " "}
	for _, char := range invalidChars {
		if strings.Contains(host, char) {
			return false
		}
	}

	return true
}

