package util

import (
	"fmt"
	"os"
	"strings"
)

// GetEnvValueStrict fetches value for env variable "key". Returns error if not found or empty
func GetEnvValueStrict(key string) (string, error) {
	if val := os.Getenv(key); len(val) != 0 {
		return strings.TrimSpace(val), nil
	}

	return "", fmt.Errorf("env variable %s is not set", key)
}
