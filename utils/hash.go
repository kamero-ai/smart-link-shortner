package utils

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

// GenerateURLHash creates a unique hash for URL combination
func GenerateURLHash(originalURL, iosURL, androidURL, desktopURL, macURL string) string {
	// Combine all URLs to create a unique identifier
	combined := strings.Join([]string{
		originalURL,
		iosURL,
		androidURL,
		desktopURL,
		macURL,
	}, "|")

	hash := sha256.Sum256([]byte(combined))
	return fmt.Sprintf("%x", hash)
}

// HashAPIKey hashes the API key secret
func HashAPIKey(secret string) string {
	hash := sha256.Sum256([]byte(secret))
	return fmt.Sprintf("%x", hash)
}
