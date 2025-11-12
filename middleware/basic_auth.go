package middleware

import (
	"crypto/subtle"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func getAdminCredentials() (string, string) {
	username := os.Getenv("ADMIN_USERNAME")
	if username == "" {
		username = "admin" // default fallback for development only
	}

	password := os.Getenv("ADMIN_PASSWORD")
	if password == "" {
		// In production, ADMIN_PASSWORD must be set
		// For development, you can set it in .env file
		password = "" // No default password for security
	}

	return username, password
}

func BasicAuth() gin.HandlerFunc {
	username, password := getAdminCredentials()
	return gin.BasicAuth(gin.Accounts{
		username: password,
	})
}

// Custom basic auth middleware for better control
func CustomBasicAuth() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		username, password := getAdminCredentials()

		user, pass, hasAuth := c.Request.BasicAuth()

		if hasAuth && secureCompare(user, username) && secureCompare(pass, password) {
			c.Next()
		} else {
			c.Header("WWW-Authenticate", "Basic realm=Restricted")
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	})
}

func secureCompare(given, actual string) bool {
	return subtle.ConstantTimeCompare([]byte(given), []byte(actual)) == 1
}

// Helper function to get admin username for logging
func GetAdminUsername() string {
	username, _ := getAdminCredentials()
	return username
}

// Helper function to get admin password for logging
func GetAdminPassword() string {
	_, password := getAdminCredentials()
	return password
}
