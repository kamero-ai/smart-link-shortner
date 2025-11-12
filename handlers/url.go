package handlers

import (
	"net/http"
	"os"
	"time"

	"url-shortener/models"
	"url-shortener/services"
	"url-shortener/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type URLHandler struct {
	urlService       *services.URLService
	analyticsService *services.AnalyticsService
}

func NewURLHandler(db *gorm.DB) *URLHandler {
	return &URLHandler{
		urlService:       services.NewURLService(db),
		analyticsService: services.NewAnalyticsService(db),
	}
}

func (h *URLHandler) ShortenURL(c *gin.Context) {
	var req models.ShortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get API key ID from context (empty string if not authenticated)
	apiKeyID := c.GetString("api_key_id")

	url, isNew, err := h.urlService.CreateShortURL(req, apiKeyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create short URL"})
		return
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	response := models.ShortenResponse{
		Code:        url.Code,
		ShortURL:    baseURL + "/" + url.Code,
		OriginalURL: url.OriginalURL,
		IsNew:       isNew,
	}

	statusCode := http.StatusCreated
	if !isNew {
		statusCode = http.StatusOK
	}

	c.JSON(statusCode, response)
}

func (h *URLHandler) RedirectURL(c *gin.Context) {
	code := c.Param("code")

	url, err := h.urlService.GetURLByCode(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	// Detect platform
	platformInfo := utils.DetectPlatform(c.Request)

	// Record click analytics
	click := models.Click{
		URLCode:   code,
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Platform:  platformInfo.Platform,
		Browser:   platformInfo.Browser,
		OS:        platformInfo.OS,
		Referrer:  c.Request.Referer(),
		ClickedAt: time.Now(),
	}

	go func() {
		h.analyticsService.RecordClick(click)
		h.urlService.IncrementClickCount(code)
	}()

	// Get platform-specific redirect URL
	redirectURL := utils.GetRedirectURL(url, platformInfo.Platform)

	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

func (h *URLHandler) GetMyURLs(c *gin.Context) {
	apiKeyID := c.GetString("api_key_id")
	if apiKeyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "API key required"})
		return
	}

	urls, err := h.urlService.GetURLsByAPIKey(apiKeyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get URLs"})
		return
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	var response []models.ShortenResponse
	for _, url := range urls {
		response = append(response, models.ShortenResponse{
			Code:        url.Code,
			ShortURL:    baseURL + "/" + url.Code,
			OriginalURL: url.OriginalURL,
			IsNew:       false,
		})
	}

	c.JSON(http.StatusOK, response)
}
