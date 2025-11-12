package handlers

import (
	"net/http"

	"url-shortener/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AnalyticsHandler struct {
	analyticsService *services.AnalyticsService
}

func NewAnalyticsHandler(db *gorm.DB) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: services.NewAnalyticsService(db),
	}
}

func (h *AnalyticsHandler) GetAnalytics(c *gin.Context) {
	code := c.Param("code")

	analytics, err := h.analyticsService.GetAnalytics(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Analytics not found"})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

func (h *AnalyticsHandler) GetDetailedAnalytics(c *gin.Context) {
	code := c.Param("code")

	analytics, err := h.analyticsService.GetAnalytics(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Analytics not found"})
		return
	}

	c.JSON(http.StatusOK, analytics)
}
