package services

import (
	"time"

	"url-shortener/models"

	"gorm.io/gorm"
)

type AnalyticsService struct {
	db *gorm.DB
}

func NewAnalyticsService(db *gorm.DB) *AnalyticsService {
	return &AnalyticsService{db: db}
}

func (s *AnalyticsService) RecordClick(click models.Click) error {
	return s.db.Create(&click).Error
}

func (s *AnalyticsService) GetAnalytics(code string) (*models.AnalyticsResponse, error) {
	var url models.URL
	if err := s.db.Where("code = ?", code).First(&url).Error; err != nil {
		return nil, err
	}

	var clicks []models.Click
	s.db.Where("url_code = ?", code).Order("clicked_at desc").Limit(10).Find(&clicks)

	// Platform stats
	platformStats := make(map[string]int64)
	var platformResults []struct {
		Platform string
		Count    int64
	}
	s.db.Model(&models.Click{}).Select("platform, count(*) as count").
		Where("url_code = ?", code).Group("platform").Scan(&platformResults)
	for _, result := range platformResults {
		platformStats[result.Platform] = result.Count
	}

	// Browser stats
	browserStats := make(map[string]int64)
	var browserResults []struct {
		Browser string
		Count   int64
	}
	s.db.Model(&models.Click{}).Select("browser, count(*) as count").
		Where("url_code = ?", code).Group("browser").Scan(&browserResults)
	for _, result := range browserResults {
		browserStats[result.Browser] = result.Count
	}

	// OS stats
	osStats := make(map[string]int64)
	var osResults []struct {
		OS    string
		Count int64
	}
	s.db.Model(&models.Click{}).Select("os, count(*) as count").
		Where("url_code = ?", code).Group("os").Scan(&osResults)
	for _, result := range osResults {
		osStats[result.OS] = result.Count
	}

	return &models.AnalyticsResponse{
		Code:          url.Code,
		OriginalURL:   url.OriginalURL,
		ClickCount:    url.ClickCount,
		CreatedAt:     url.CreatedAt,
		PlatformStats: platformStats,
		BrowserStats:  browserStats,
		OSStats:       osStats,
		RecentClicks:  clicks,
	}, nil
}

// GetAllURLsAnalytics returns paginated analytics for all URLs with enhanced filtering
func (s *AnalyticsService) GetAllURLsAnalytics(offset, limit int) ([]models.URLAnalyticsSummary, int64, error) {
	var urls []models.URL
	var total int64

	// Get total count
	if err := s.db.Model(&models.URL{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get URLs with pagination, ordered by click count and creation date
	if err := s.db.Offset(offset).Limit(limit).
		Order("click_count desc, created_at desc").Find(&urls).Error; err != nil {
		return nil, 0, err
	}

	var result []models.URLAnalyticsSummary
	for _, url := range urls {
		// Get recent clicks
		var recentClicks []models.Click
		s.db.Where("url_code = ?", url.Code).
			Order("clicked_at desc").Limit(5).Find(&recentClicks)

		// Get platform stats
		platformStats := make(map[string]int64)
		var platformResults []struct {
			Platform string
			Count    int64
		}
		s.db.Model(&models.Click{}).Select("platform, count(*) as count").
			Where("url_code = ?", url.Code).Group("platform").Scan(&platformResults)
		for _, result := range platformResults {
			platformStats[result.Platform] = result.Count
		}

		summary := models.URLAnalyticsSummary{
			Code:          url.Code,
			OriginalURL:   url.OriginalURL,
			ClickCount:    url.ClickCount,
			CreatedAt:     url.CreatedAt,
			CreatedByAPI:  url.CreatedByAPIKey != "",
			APIKeyID:      url.CreatedByAPIKey,
			PlatformStats: platformStats,
			RecentClicks:  recentClicks,
		}

		result = append(result, summary)
	}

	return result, total, nil
}

// GetSystemStats returns enhanced system-wide statistics (PostgreSQL compatible)
func (s *AnalyticsService) GetSystemStats() (*models.SystemStats, error) {
	var stats models.SystemStats

	// Total URLs
	s.db.Model(&models.URL{}).Count(&stats.TotalURLs)

	// Total Clicks
	s.db.Model(&models.Click{}).Count(&stats.TotalClicks)

	// URLs created today
	today := time.Now().Truncate(24 * time.Hour)
	s.db.Model(&models.URL{}).Where("created_at >= ?", today).Count(&stats.URLsToday)

	// Clicks today
	s.db.Model(&models.Click{}).Where("clicked_at >= ?", today).Count(&stats.ClicksToday)

	// URLs created this week
	weekStart := today.AddDate(0, 0, -int(today.Weekday()))
	s.db.Model(&models.URL{}).Where("created_at >= ?", weekStart).Count(&stats.URLsThisWeek)

	// Clicks this week
	s.db.Model(&models.Click{}).Where("clicked_at >= ?", weekStart).Count(&stats.ClicksThisWeek)

	// URLs created this month
	monthStart := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
	s.db.Model(&models.URL{}).Where("created_at >= ?", monthStart).Count(&stats.URLsThisMonth)

	// Clicks this month
	s.db.Model(&models.Click{}).Where("clicked_at >= ?", monthStart).Count(&stats.ClicksThisMonth)

	// Top platforms (enhanced)
	var platformResults []struct {
		Platform string
		Count    int64
	}
	s.db.Model(&models.Click{}).Select("platform, count(*) as count").
		Group("platform").Order("count desc").Limit(10).Scan(&platformResults)

	stats.TopPlatforms = make(map[string]int64)
	for _, result := range platformResults {
		stats.TopPlatforms[result.Platform] = result.Count
	}

	// Top browsers
	var browserResults []struct {
		Browser string
		Count   int64
	}
	s.db.Model(&models.Click{}).Select("browser, count(*) as count").
		Group("browser").Order("count desc").Limit(5).Scan(&browserResults)

	stats.TopBrowsers = make(map[string]int64)
	for _, result := range browserResults {
		stats.TopBrowsers[result.Browser] = result.Count
	}

	// URLs with API keys vs without
	s.db.Model(&models.URL{}).Where("created_by_api_key != ''").Count(&stats.URLsWithAPIKey)
	stats.URLsWithoutAPIKey = stats.TotalURLs - stats.URLsWithAPIKey

	// Average clicks per URL
	if stats.TotalURLs > 0 {
		stats.AvgClicksPerURL = float64(stats.TotalClicks) / float64(stats.TotalURLs)
	}

	// Most clicked URL today
	var topURLToday struct {
		Code       string
		ClickCount int64
	}
	s.db.Model(&models.Click{}).
		Select("url_code as code, count(*) as click_count").
		Where("clicked_at >= ?", today).
		Group("url_code").
		Order("click_count desc").
		Limit(1).
		Scan(&topURLToday)
	stats.TopURLToday = topURLToday.Code

	// Daily trends for last 7 days (PostgreSQL compatible)
	stats.DailyTrends = s.getDailyTrends(7)

	// Hourly trends for last 24 hours (PostgreSQL compatible)
	stats.HourlyTrends = s.getGlobalHourlyTrends(24)

	return &stats, nil
}

// getDailyTrends returns daily trends for the specified number of days (PostgreSQL compatible)
func (s *AnalyticsService) getDailyTrends(days int) []models.DailyTrend {
	var trends []models.DailyTrend

	// Get URL creation trends
	var urlResults []struct {
		Date  time.Time
		Count int64
	}

	s.db.Model(&models.URL{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("created_at >= NOW() - INTERVAL '? days'", days).
		Group("DATE(created_at)").
		Order("date desc").
		Scan(&urlResults)

	// Get click trends
	var clickResults []struct {
		Date  time.Time
		Count int64
	}

	s.db.Model(&models.Click{}).
		Select("DATE(clicked_at) as date, COUNT(*) as count").
		Where("clicked_at >= NOW() - INTERVAL '? days'", days).
		Group("DATE(clicked_at)").
		Order("date desc").
		Scan(&clickResults)

	// Combine results
	dateMap := make(map[string]models.DailyTrend)

	for _, result := range urlResults {
		dateStr := result.Date.Format("2006-01-02")
		trend := dateMap[dateStr]
		trend.Date = result.Date
		trend.URLs = result.Count
		dateMap[dateStr] = trend
	}

	for _, result := range clickResults {
		dateStr := result.Date.Format("2006-01-02")
		trend := dateMap[dateStr]
		trend.Date = result.Date
		trend.Clicks = result.Count
		dateMap[dateStr] = trend
	}

	// Convert to slice
	for _, trend := range dateMap {
		trends = append(trends, trend)
	}

	return trends
}

// getGlobalHourlyTrends returns global hourly trends (PostgreSQL compatible)
func (s *AnalyticsService) getGlobalHourlyTrends(hours int) []models.HourlyTrend {
	var trends []models.HourlyTrend

	var results []struct {
		Hour  int
		Count int64
	}

	s.db.Model(&models.Click{}).
		Select("EXTRACT(HOUR FROM clicked_at)::int as hour, COUNT(*) as count").
		Where("clicked_at >= NOW() - INTERVAL '? hours'", hours).
		Group("EXTRACT(HOUR FROM clicked_at)").
		Order("hour").
		Scan(&results)

	// Create 24-hour trend data
	hourMap := make(map[int]int64)
	for _, result := range results {
		hourMap[result.Hour] = result.Count
	}

	for hour := 0; hour < 24; hour++ {
		trends = append(trends, models.HourlyTrend{
			Hour:  hour,
			Count: hourMap[hour],
		})
	}

	return trends
}

// GetTopURLs returns the most clicked URLs
func (s *AnalyticsService) GetTopURLs(limit int) ([]models.URLAnalyticsSummary, error) {
	var urls []models.URL

	if err := s.db.Where("click_count > 0").
		Order("click_count desc").
		Limit(limit).
		Find(&urls).Error; err != nil {
		return nil, err
	}

	var result []models.URLAnalyticsSummary
	for _, url := range urls {
		// Get platform stats
		platformStats := make(map[string]int64)
		var platformResults []struct {
			Platform string
			Count    int64
		}
		s.db.Model(&models.Click{}).Select("platform, count(*) as count").
			Where("url_code = ?", url.Code).Group("platform").Scan(&platformResults)
		for _, pResult := range platformResults {
			platformStats[pResult.Platform] = pResult.Count
		}

		summary := models.URLAnalyticsSummary{
			Code:          url.Code,
			OriginalURL:   url.OriginalURL,
			ClickCount:    url.ClickCount,
			CreatedAt:     url.CreatedAt,
			CreatedByAPI:  url.CreatedByAPIKey != "",
			APIKeyID:      url.CreatedByAPIKey,
			PlatformStats: platformStats,
		}

		result = append(result, summary)
	}

	return result, nil
}

// GetRecentActivity returns recent URL creations and clicks
func (s *AnalyticsService) GetRecentActivity(limit int) (*models.RecentActivity, error) {
	var recentURLs []models.URL
	var recentClicks []models.Click

	// Get recent URLs
	if err := s.db.Order("created_at desc").Limit(limit).Find(&recentURLs).Error; err != nil {
		return nil, err
	}

	// Get recent clicks
	if err := s.db.Order("clicked_at desc").Limit(limit).Find(&recentClicks).Error; err != nil {
		return nil, err
	}

	return &models.RecentActivity{
		RecentURLs:   recentURLs,
		RecentClicks: recentClicks,
	}, nil
}

// GetClicksByTimeRange returns clicks within a specific time range
func (s *AnalyticsService) GetClicksByTimeRange(code string, startTime, endTime time.Time) ([]models.Click, error) {
	var clicks []models.Click

	query := s.db.Where("clicked_at >= ? AND clicked_at <= ?", startTime, endTime)
	if code != "" {
		query = query.Where("url_code = ?", code)
	}

	if err := query.Order("clicked_at desc").Find(&clicks).Error; err != nil {
		return nil, err
	}

	return clicks, nil
}

// GetGeoStats returns geographical statistics for clicks
func (s *AnalyticsService) GetGeoStats(code string) (map[string]int64, error) {
	var results []struct {
		Country string
		Count   int64
	}

	query := s.db.Model(&models.Click{}).Select("country, count(*) as count").Group("country")
	if code != "" {
		query = query.Where("url_code = ?", code)
	}

	if err := query.Scan(&results).Error; err != nil {
		return nil, err
	}

	geoStats := make(map[string]int64)
	for _, result := range results {
		if result.Country != "" {
			geoStats[result.Country] = result.Count
		}
	}

	return geoStats, nil
}

// GetReferrerStats returns referrer statistics
func (s *AnalyticsService) GetReferrerStats(code string) (map[string]int64, error) {
	var results []struct {
		Referrer string
		Count    int64
	}

	query := s.db.Model(&models.Click{}).Select("referrer, count(*) as count").Group("referrer")
	if code != "" {
		query = query.Where("url_code = ?", code)
	}

	if err := query.Order("count desc").Limit(10).Scan(&results).Error; err != nil {
		return nil, err
	}

	referrerStats := make(map[string]int64)
	for _, result := range results {
		referrer := result.Referrer
		if referrer == "" {
			referrer = "Direct"
		}
		referrerStats[referrer] = result.Count
	}

	return referrerStats, nil
}

// GetClickTrends returns click trends over time
func (s *AnalyticsService) GetClickTrends(code string, days int) ([]models.DailyTrend, error) {
	var results []struct {
		Date  time.Time
		Count int64
	}

	query := s.db.Model(&models.Click{}).
		Select("DATE(clicked_at) as date, count(*) as count").
		Where("clicked_at >= NOW() - INTERVAL '? days'", days).
		Group("DATE(clicked_at)").
		Order("date desc")

	if code != "" {
		query = query.Where("url_code = ?", code)
	}

	if err := query.Scan(&results).Error; err != nil {
		return nil, err
	}

	var trends []models.DailyTrend
	for _, result := range results {
		trends = append(trends, models.DailyTrend{
			Date:   result.Date,
			Clicks: result.Count,
		})
	}

	return trends, nil
}

// GetPerformanceMetrics returns performance metrics for URLs
func (s *AnalyticsService) GetPerformanceMetrics() (*models.PerformanceMetrics, error) {
	var metrics models.PerformanceMetrics

	// Average response time simulation (in real app, you'd track this)
	metrics.AvgResponseTime = 150.0 // milliseconds

	// Success rate (URLs that have been clicked vs total URLs)
	var totalURLs, clickedURLs int64
	s.db.Model(&models.URL{}).Count(&totalURLs)
	s.db.Model(&models.URL{}).Where("click_count > 0").Count(&clickedURLs)

	if totalURLs > 0 {
		metrics.SuccessRate = float64(clickedURLs) / float64(totalURLs) * 100
	}

	// Uptime percentage (simulation - in real app, track actual uptime)
	metrics.UptimePercentage = 99.9

	// Total requests today
	today := time.Now().Truncate(24 * time.Hour)
	s.db.Model(&models.Click{}).Where("clicked_at >= ?", today).Count(&metrics.RequestsToday)

	// Error rate (simulation - track actual errors)
	metrics.ErrorRate = 0.01

	return &metrics, nil
}

// GetAPIKeyUsage returns usage statistics for API keys
func (s *AnalyticsService) GetAPIKeyUsage() ([]models.APIKeyUsage, error) {
	var results []struct {
		APIKeyID   string
		URLCount   int64
		ClickCount int64
	}

	s.db.Model(&models.URL{}).
		Select("created_by_api_key as api_key_id, COUNT(*) as url_count, COALESCE(SUM(click_count), 0) as click_count").
		Where("created_by_api_key != ''").
		Group("created_by_api_key").
		Order("click_count desc").
		Scan(&results)

	var usage []models.APIKeyUsage
	for _, result := range results {
		// Get API key details
		var apiKey models.APIKey
		s.db.Where("key_id = ?", result.APIKeyID).First(&apiKey)

		usage = append(usage, models.APIKeyUsage{
			KeyID:      result.APIKeyID,
			KeyName:    apiKey.Name,
			URLCount:   result.URLCount,
			ClickCount: result.ClickCount,
			LastUsed:   apiKey.LastUsedAt,
		})
	}

	return usage, nil
}
