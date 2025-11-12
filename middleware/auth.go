package middleware

import (
	"net/http"
	"strings"

	"url-shortener/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func APIKeyAuth(db *gorm.DB) gin.HandlerFunc {
	apiKeyService := services.NewAPIKeyService(db)

	return gin.HandlerFunc(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Expected format: "Bearer ak_xxxxx:sk_xxxxx"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		// Parse key ID and secret
		keyParts := strings.Split(parts[1], ":")
		if len(keyParts) != 2 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key format"})
			c.Abort()
			return
		}

		keyID := keyParts[0]
		keySecret := keyParts[1]

		// Validate API key
		apiKey, err := apiKeyService.ValidateAPIKey(keyID, keySecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
			c.Abort()
			return
		}

		// Store API key info in context
		c.Set("api_key", apiKey)
		c.Set("api_key_id", apiKey.KeyID)
		c.Next()
	})
}

// Optional API key middleware - allows both authenticated and unauthenticated requests
func OptionalAPIKeyAuth(db *gorm.DB) gin.HandlerFunc {
	apiKeyService := services.NewAPIKeyService(db)

	return gin.HandlerFunc(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Set("api_key_id", "")
			c.Next()
			return
		}

		// Parse and validate API key
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			keyParts := strings.Split(parts[1], ":")
			if len(keyParts) == 2 {
				apiKey, err := apiKeyService.ValidateAPIKey(keyParts[0], keyParts[1])
				if err == nil {
					c.Set("api_key", apiKey)
					c.Set("api_key_id", apiKey.KeyID)
				}
			}
		}

		if c.GetString("api_key_id") == "" {
			c.Set("api_key_id", "")
		}

		c.Next()
	})
}
