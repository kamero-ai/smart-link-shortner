package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"url-shortener/models"
	"url-shortener/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AdminHandler struct {
	analyticsService *services.AnalyticsService
	urlService       *services.URLService
}

func NewAdminHandler(db *gorm.DB) *AdminHandler {
	return &AdminHandler{
		analyticsService: services.NewAnalyticsService(db),
		urlService:       services.NewURLService(db),
	}
}

// Get all URLs with their analytics (enhanced with filtering)
func (h *AdminHandler) GetAllURLsAnalytics(c *gin.Context) {
	// Get pagination parameters
	page := 1
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	limit := 25
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	// Get filter parameters
	source := c.Query("source")     // api, web
	platform := c.Query("platform") // ios, android, desktop, mac
	minClicks := c.Query("min_clicks")
	search := c.Query("search")

	offset := (page - 1) * limit

	analytics, total, err := h.analyticsService.GetAllURLsAnalytics(offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get analytics"})
		return
	}

	// Apply client-side filters (in production, move to database query)
	filteredAnalytics := h.applyFilters(analytics, source, platform, minClicks, search)

	// Calculate total pages
	totalPages := (total + int64(limit) - 1) / int64(limit)

	response := gin.H{
		"data": filteredAnalytics,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
		"filters": gin.H{
			"source":     source,
			"platform":   platform,
			"min_clicks": minClicks,
			"search":     search,
		},
		"base_url": getBaseURL(),
	}

	c.JSON(http.StatusOK, response)
}

// Get enhanced system-wide statistics
func (h *AdminHandler) GetSystemStats(c *gin.Context) {
	stats, err := h.analyticsService.GetSystemStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get system stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// Get top performing URLs
func (h *AdminHandler) GetTopURLs(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 10
	}

	topURLs, err := h.analyticsService.GetTopURLs(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get top URLs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"top_urls": topURLs,
		"base_url": getBaseURL(),
	})
}

// Get recent activity
func (h *AdminHandler) GetRecentActivity(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 50 {
		limit = 10
	}

	activity, err := h.analyticsService.GetRecentActivity(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recent activity"})
		return
	}

	c.JSON(http.StatusOK, activity)
}

// Get performance metrics
func (h *AdminHandler) GetPerformanceMetrics(c *gin.Context) {
	metrics, err := h.analyticsService.GetPerformanceMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get performance metrics"})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// Get API key usage statistics
func (h *AdminHandler) GetAPIKeyUsage(c *gin.Context) {
	usage, err := h.analyticsService.GetAPIKeyUsage()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get API key usage"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"api_key_usage": usage,
		"total_keys":    len(usage),
	})
}

// Get geographic analytics
func (h *AdminHandler) GetGeoAnalytics(c *gin.Context) {
	code := c.Query("code") // Optional: get geo stats for specific URL

	geoStats, err := h.analyticsService.GetGeoStats(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get geo analytics"})
		return
	}

	// Calculate total and percentages
	var total int64
	for _, count := range geoStats {
		total += count
	}

	geoAnalytics := make(map[string]interface{})
	for country, count := range geoStats {
		percentage := float64(0)
		if total > 0 {
			percentage = float64(count) / float64(total) * 100
		}
		geoAnalytics[country] = gin.H{
			"count":      count,
			"percentage": percentage,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"geo_analytics": geoAnalytics,
		"total_clicks":  total,
		"countries":     len(geoStats),
	})
}

// Get referrer analytics
func (h *AdminHandler) GetReferrerAnalytics(c *gin.Context) {
	code := c.Query("code") // Optional: get referrer stats for specific URL

	referrerStats, err := h.analyticsService.GetReferrerStats(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get referrer analytics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"referrer_stats": referrerStats,
	})
}

// Get click trends
func (h *AdminHandler) GetClickTrends(c *gin.Context) {
	code := c.Query("code")
	daysStr := c.DefaultQuery("days", "7")

	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 || days > 365 {
		days = 7
	}

	trends, err := h.analyticsService.GetClickTrends(code, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get click trends"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"trends": trends,
		"days":   days,
		"code":   code,
	})
}

// Export analytics data
func (h *AdminHandler) ExportAnalytics(c *gin.Context) {
	format := c.DefaultQuery("format", "json") // json, csv
	timeRange := c.DefaultQuery("range", "30") // days

	days, err := strconv.Atoi(timeRange)
	if err != nil || days <= 0 {
		days = 30
	}

	// Get data for export
	startTime := time.Now().AddDate(0, 0, -days)
	endTime := time.Now()

	clicks, err := h.analyticsService.GetClicksByTimeRange("", startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export data"})
		return
	}

	if format == "csv" {
		h.exportAsCSV(c, clicks)
	} else {
		exportData := models.ExportData{
			Clicks:     clicks,
			ExportedAt: time.Now(),
			Filters: map[string]interface{}{
				"time_range": timeRange,
				"format":     format,
			},
		}
		c.JSON(http.StatusOK, exportData)
	}
}

// Delete URL (soft delete)
func (h *AdminHandler) DeleteURL(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Code is required"})
		return
	}

	err := h.urlService.SoftDeleteURL(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "URL deleted successfully",
		"code":    code,
	})
}

// Restore URL
func (h *AdminHandler) RestoreURL(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Code is required"})
		return
	}

	err := h.urlService.RestoreURL(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restore URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "URL restored successfully",
		"code":    code,
	})
}

// Get real-time statistics
func (h *AdminHandler) GetRealTimeStats(c *gin.Context) {
	// This would be implemented with WebSocket or Server-Sent Events in production
	// For now, return recent data

	// Get clicks in last 5 minutes for active users simulation
	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)
	oneHourAgo := time.Now().Add(-1 * time.Hour)

	recentClicks, _ := h.analyticsService.GetClicksByTimeRange("", fiveMinutesAgo, time.Now())
	hourlyClicks, _ := h.analyticsService.GetClicksByTimeRange("", oneHourAgo, time.Now())

	// Get recent activity
	activity, _ := h.analyticsService.GetRecentActivity(10)

	realTimeStats := models.RealTimeStats{
		ActiveUsers:    int64(len(recentClicks)), // Simplified - unique IPs in last 5 min
		ClicksLastHour: int64(len(hourlyClicks)),
		LiveClicks:     recentClicks[:min(10, len(recentClicks))],
	}

	if activity != nil {
		realTimeStats.URLsLastHour = int64(len(activity.RecentURLs))
	}

	c.JSON(http.StatusOK, realTimeStats)
}

// Bulk operations
func (h *AdminHandler) BulkDeleteURLs(c *gin.Context) {
	var request struct {
		Codes []string `json:"codes" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var results []gin.H
	successCount := 0

	for _, code := range request.Codes {
		err := h.urlService.SoftDeleteURL(code)
		if err != nil {
			results = append(results, gin.H{
				"code":    code,
				"success": false,
				"error":   err.Error(),
			})
		} else {
			results = append(results, gin.H{
				"code":    code,
				"success": true,
			})
			successCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Bulk delete completed",
		"total":         len(request.Codes),
		"success_count": successCount,
		"failed_count":  len(request.Codes) - successCount,
		"results":       results,
	})
}

// Helper functions

func (h *AdminHandler) applyFilters(analytics []models.URLAnalyticsSummary, source, platform, minClicksStr, search string) []models.URLAnalyticsSummary {
	filtered := make([]models.URLAnalyticsSummary, 0)

	minClicks := int64(0)
	if minClicksStr != "" {
		if parsed, err := strconv.ParseInt(minClicksStr, 10, 64); err == nil {
			minClicks = parsed
		}
	}

	for _, url := range analytics {
		// Source filter
		if source != "" {
			if source == "api" && !url.CreatedByAPI {
				continue
			}
			if source == "web" && url.CreatedByAPI {
				continue
			}
		}

		// Platform filter
		if platform != "" {
			if _, exists := url.PlatformStats[platform]; !exists {
				continue
			}
		}

		// Min clicks filter
		if url.ClickCount < minClicks {
			continue
		}

		// Search filter
		if search != "" {
			searchLower := strings.ToLower(search)
			if !strings.Contains(strings.ToLower(url.Code), searchLower) &&
				!strings.Contains(strings.ToLower(url.OriginalURL), searchLower) {
				continue
			}
		}

		filtered = append(filtered, url)
	}

	return filtered
}

func (h *AdminHandler) exportAsCSV(c *gin.Context, clicks []models.Click) {
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=analytics_export.csv")

	// CSV header
	csvData := "URL Code,Original URL,IP Address,User Agent,Platform,Browser,OS,Country,City,Referrer,Clicked At\n"

	// CSV data
	for _, click := range clicks {
		// Get URL info
		url, _ := h.urlService.GetURLByCode(click.URLCode)
		originalURL := ""
		if url != nil {
			originalURL = url.OriginalURL
		}

		csvData += fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
			click.URLCode,
			originalURL,
			click.IPAddress,
			click.UserAgent,
			click.Platform,
			click.Browser,
			click.OS,
			click.Country,
			click.City,
			click.Referrer,
			click.ClickedAt.Format("2006-01-02 15:04:05"),
		)
	}

	c.String(http.StatusOK, csvData)
}

// Get base URL helper
func getBaseURL() string {
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	return baseURL
}

// Utility function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
