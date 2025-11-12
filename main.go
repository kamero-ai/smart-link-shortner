package main

import (
	"log"
	"os"

	"url-shortener/config"
	"url-shortener/handlers"
	"url-shortener/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db := config.InitDB(cfg)

	// Set Gin mode
	if cfg.GinMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize router
	r := gin.Default()

	// Middleware
	r.Use(middleware.CORS())
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	// Serve static files
	r.Static("/static", "./static")
	r.LoadHTMLGlob("static/*.html")

	// Initialize handlers
	urlHandler := handlers.NewURLHandler(db)
	analyticsHandler := handlers.NewAnalyticsHandler(db)
	apiKeyHandler := handlers.NewAPIKeyHandler(db)
	adminHandler := handlers.NewAdminHandler(db)

	// Public API routes (with optional API key auth)
	api := r.Group("/api/v1")
	api.Use(middleware.OptionalAPIKeyAuth(db))
	{
		api.POST("/shorten", urlHandler.ShortenURL)
		api.GET("/analytics/:code", analyticsHandler.GetAnalytics)
		api.GET("/analytics/:code/detailed", analyticsHandler.GetDetailedAnalytics)
	}

	// Protected API routes (require API key)
	protectedAPI := r.Group("/api/v1")
	protectedAPI.Use(middleware.APIKeyAuth(db))
	{
		protectedAPI.GET("/my-urls", urlHandler.GetMyURLs)
	}

	// Protected Admin API routes (require basic auth)
	adminAPI := r.Group("/admin/api/v1")
	adminAPI.Use(middleware.CustomBasicAuth())
	{
		adminAPI.POST("/api-keys", apiKeyHandler.CreateAPIKey)
		adminAPI.GET("/api-keys", apiKeyHandler.GetAPIKeys)
		adminAPI.DELETE("/api-keys/:keyId", apiKeyHandler.DeactivateAPIKey)
		adminAPI.GET("/urls/analytics", adminHandler.GetAllURLsAnalytics)
		adminAPI.GET("/system/stats", adminHandler.GetSystemStats)
	}

	// Public web routes
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{
			"BaseURL": getBaseURL(),
		})
	})

	// Protected web routes (require basic auth)
	// Protected web routes (require basic auth)
	protected := r.Group("/")
	protected.Use(middleware.CustomBasicAuth())
	{
		// Individual URL analytics dashboard
		protected.GET("/dashboard", func(c *gin.Context) {
			c.HTML(200, "dashboard.html", gin.H{
				"BaseURL": getBaseURL(),
			})
		})

		// API Keys Management - MAIN admin page
		protected.GET("/admin", func(c *gin.Context) {
			c.HTML(200, "admin.html", gin.H{
				"BaseURL": getBaseURL(),
			})
		})

		// Alternative route for API keys management (same page)
		protected.GET("/admin/api-keys", func(c *gin.Context) {
			c.HTML(200, "admin.html", gin.H{
				"BaseURL": getBaseURL(),
			})
		})

		// Admin Analytics Dashboard
		protected.GET("/admin/analytics", func(c *gin.Context) {
			c.HTML(200, "admin-analytics.html", gin.H{
				"BaseURL": getBaseURL(),
			})
		})
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"service":   "kamero-url-shortener",
			"version":   "1.0.0",
			"base_url":  getBaseURL(),
			"timestamp": cfg.Port,
		})
	})

	// 404 handler for web routes
	r.NoRoute(func(c *gin.Context) {
		// If it's an API request, return JSON
		if gin.IsDebugging() || len(c.Request.URL.Path) > 6 || c.Request.Header.Get("Content-Type") == "application/json" {
			c.JSON(404, gin.H{
				"error":   "Not Found",
				"message": "The requested resource was not found",
				"service": "kamero-url-shortener",
			})
			return
		}

		// Otherwise, try to handle as short URL redirect
		code := c.Param("code")
		if code == "" && len(c.Request.URL.Path) > 1 {
			// Extract code from path
			code = c.Request.URL.Path[1:] // Remove leading slash
		}

		if len(code) == 6 {
			// Try to redirect
			urlHandler.RedirectURL(c)
		} else {
			c.JSON(404, gin.H{
				"error":   "Not Found",
				"message": "Invalid short URL code",
				"service": "kamero-url-shortener",
			})
		}
	})

	// Redirect route - this handles the short URLs (must be last)
	r.GET("/:code", urlHandler.RedirectURL)

	log.Printf("🚀 Server starting on port %s", cfg.Port)
	log.Printf("🌐 Base URL: %s", cfg.BaseURL)
	log.Printf("💾 Database: %s@%s:%s/%s", cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)
	log.Printf("🔐 Admin username: %s", middleware.GetAdminUsername())
	log.Printf("📊 Admin Dashboard: %s/admin", cfg.BaseURL)
	log.Printf("🔑 API Keys Management: %s/admin/api-keys", cfg.BaseURL)
	log.Printf("📈 Analytics Dashboard: %s/admin/analytics", cfg.BaseURL)

	log.Fatal(r.Run(":" + cfg.Port))
}

// getBaseURL returns the base URL from environment or default
func getBaseURL() string {
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	return baseURL
}
