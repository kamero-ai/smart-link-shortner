package services

import (
	"errors"
	"time"

	"url-shortener/models"
	"url-shortener/utils"

	"gorm.io/gorm"
)

type URLService struct {
	db *gorm.DB
}

func NewURLService(db *gorm.DB) *URLService {
	return &URLService{db: db}
}

func (s *URLService) CreateShortURL(req models.ShortenRequest, apiKeyID string) (*models.URL, bool, error) {
	// Generate hash for the URL combination
	urlHash := utils.GenerateURLHash(
		req.URL,
		req.IOSRedirectURL,
		req.AndroidRedirectURL,
		req.DesktopRedirectURL,
		req.MacRedirectURL,
	)

	// Check if URL combination already exists
	var existingURL models.URL
	result := s.db.Where("url_hash = ?", urlHash).First(&existingURL)
	if result.Error == nil {
		// URL already exists, return the existing one
		return &existingURL, false, nil
	}

	// If error is not "record not found", return the error
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, false, result.Error
	}

	// Generate unique short code
	var code string
	var err error

	for {
		code, err = utils.GenerateShortCode(6)
		if err != nil {
			return nil, false, err
		}

		// Check if code already exists
		var existingCode models.URL
		result := s.db.Where("code = ?", code).First(&existingCode)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			break
		}
	}

	url := models.URL{
		Code:               code,
		OriginalURL:        req.URL,
		URLHash:            urlHash,
		IOSRedirectURL:     req.IOSRedirectURL,
		AndroidRedirectURL: req.AndroidRedirectURL,
		DesktopRedirectURL: req.DesktopRedirectURL,
		MacRedirectURL:     req.MacRedirectURL,
		CreatedByAPIKey:    apiKeyID,
	}

	result = s.db.Create(&url)
	if result.Error != nil {
		return nil, false, result.Error
	}

	return &url, true, nil
}

func (s *URLService) GetURLByCode(code string) (*models.URL, error) {
	var url models.URL
	result := s.db.Where("code = ?", code).First(&url)
	if result.Error != nil {
		return nil, result.Error
	}
	return &url, nil
}

func (s *URLService) IncrementClickCount(code string) error {
	return s.db.Model(&models.URL{}).Where("code = ?", code).
		UpdateColumn("click_count", gorm.Expr("click_count + ?", 1)).Error
}

func (s *URLService) GetURLsByAPIKey(apiKeyID string) ([]models.URL, error) {
	var urls []models.URL
	result := s.db.Where("created_by_api_key = ?", apiKeyID).
		Order("created_at desc").Find(&urls)
	return urls, result.Error
}

// Enhanced methods for admin functionality

// GetAllURLs returns all URLs with pagination and filtering
func (s *URLService) GetAllURLs(offset, limit int, filter models.AnalyticsFilter) ([]models.URL, int64, error) {
	var urls []models.URL
	var total int64

	query := s.db.Model(&models.URL{})

	// Apply filters
	if filter.APIKeyID != "" {
		query = query.Where("created_by_api_key = ?", filter.APIKeyID)
	}

	if filter.StartDate != nil {
		query = query.Where("created_at >= ?", *filter.StartDate)
	}

	if filter.EndDate != nil {
		query = query.Where("created_at <= ?", *filter.EndDate)
	}

	if filter.MinClicks > 0 {
		query = query.Where("click_count >= ?", filter.MinClicks)
	}

	if filter.MaxClicks > 0 {
		query = query.Where("click_count <= ?", filter.MaxClicks)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	orderBy := "created_at desc"
	if filter.SortBy != "" {
		order := "desc"
		if filter.SortOrder == "asc" {
			order = "asc"
		}
		orderBy = filter.SortBy + " " + order
	}

	// Get URLs with pagination
	if err := query.Offset(offset).Limit(limit).Order(orderBy).Find(&urls).Error; err != nil {
		return nil, 0, err
	}

	return urls, total, nil
}

// SoftDeleteURL marks a URL as deleted
func (s *URLService) SoftDeleteURL(code string) error {
	result := s.db.Where("code = ?", code).Delete(&models.URL{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("URL not found")
	}
	return nil
}

// RestoreURL restores a soft-deleted URL
func (s *URLService) RestoreURL(code string) error {
	result := s.db.Unscoped().Model(&models.URL{}).Where("code = ?", code).Update("deleted_at", nil)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("URL not found")
	}
	return nil
}

// GetDeletedURLs returns soft-deleted URLs
func (s *URLService) GetDeletedURLs(offset, limit int) ([]models.URL, int64, error) {
	var urls []models.URL
	var total int64

	// Count deleted URLs
	s.db.Unscoped().Model(&models.URL{}).Where("deleted_at IS NOT NULL").Count(&total)

	// Get deleted URLs
	err := s.db.Unscoped().Where("deleted_at IS NOT NULL").
		Offset(offset).Limit(limit).
		Order("deleted_at desc").Find(&urls).Error

	return urls, total, err
}

// BulkDeleteURLs deletes multiple URLs
func (s *URLService) BulkDeleteURLs(codes []string) (int64, error) {
	result := s.db.Where("code IN ?", codes).Delete(&models.URL{})
	return result.RowsAffected, result.Error
}

// BulkRestoreURLs restores multiple URLs
func (s *URLService) BulkRestoreURLs(codes []string) (int64, error) {
	result := s.db.Unscoped().Model(&models.URL{}).
		Where("code IN ? AND deleted_at IS NOT NULL", codes).
		Update("deleted_at", nil)
	return result.RowsAffected, result.Error
}

// UpdateURL updates URL metadata
func (s *URLService) UpdateURL(code string, updates models.URL) error {
	result := s.db.Model(&models.URL{}).Where("code = ?", code).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("URL not found")
	}
	return nil
}

// GetURLStats returns basic statistics for a URL
func (s *URLService) GetURLStats(code string) (*models.URLStats, error) {
	var url models.URL
	if err := s.db.Where("code = ?", code).First(&url).Error; err != nil {
		return nil, err
	}

	var stats models.URLStats
	stats.Code = url.Code
	stats.OriginalURL = url.OriginalURL
	stats.CreatedAt = url.CreatedAt
	stats.ClickCount = url.ClickCount

	// Get click statistics
	s.db.Model(&models.Click{}).Where("url_code = ?", code).Count(&stats.TotalClicks)

	// Get clicks today
	today := time.Now().Truncate(24 * time.Hour)
	s.db.Model(&models.Click{}).Where("url_code = ? AND clicked_at >= ?", code, today).Count(&stats.ClicksToday)

	// Get clicks this week
	weekStart := today.AddDate(0, 0, -int(today.Weekday()))
	s.db.Model(&models.Click{}).Where("url_code = ? AND clicked_at >= ?", code, weekStart).Count(&stats.ClicksThisWeek)

	// Get clicks this month
	monthStart := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
	s.db.Model(&models.Click{}).Where("url_code = ? AND clicked_at >= ?", code, monthStart).Count(&stats.ClicksThisMonth)

	// Get unique visitors (simplified - by IP)
	s.db.Model(&models.Click{}).Select("DISTINCT ip_address").Where("url_code = ?", code).Count(&stats.UniqueVisitors)

	// Get last click time
	var lastClick models.Click
	if err := s.db.Where("url_code = ?", code).Order("clicked_at desc").First(&lastClick).Error; err == nil {
		stats.LastClickAt = &lastClick.ClickedAt
	}

	// Calculate average clicks per day
	daysSinceCreation := time.Since(url.CreatedAt).Hours() / 24
	if daysSinceCreation > 0 {
		stats.AvgClicksPerDay = float64(stats.TotalClicks) / daysSinceCreation
	}

	return &stats, nil
}

// SearchURLs searches URLs by various criteria
func (s *URLService) SearchURLs(query string, offset, limit int) ([]models.URL, int64, error) {
	var urls []models.URL
	var total int64

	searchQuery := s.db.Model(&models.URL{}).Where(
		"code ILIKE ? OR original_url ILIKE ? OR ios_redirect_url ILIKE ? OR android_redirect_url ILIKE ?",
		"%"+query+"%", "%"+query+"%", "%"+query+"%", "%"+query+"%",
	)

	// Get total count
	if err := searchQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get URLs with pagination
	if err := searchQuery.Offset(offset).Limit(limit).
		Order("click_count desc, created_at desc").Find(&urls).Error; err != nil {
		return nil, 0, err
	}

	return urls, total, nil
}

// GetPopularURLs returns URLs sorted by click count
func (s *URLService) GetPopularURLs(limit int, timeRange string) ([]models.URL, error) {
	var urls []models.URL
	query := s.db.Model(&models.URL{}).Where("click_count > 0")

	// Apply time range filter
	switch timeRange {
	case "today":
		today := time.Now().Truncate(24 * time.Hour)
		query = query.Where("created_at >= ?", today)
	case "week":
		weekStart := time.Now().AddDate(0, 0, -7)
		query = query.Where("created_at >= ?", weekStart)
	case "month":
		monthStart := time.Now().AddDate(0, -1, 0)
		query = query.Where("created_at >= ?", monthStart)
	}

	if err := query.Order("click_count desc").Limit(limit).Find(&urls).Error; err != nil {
		return nil, err
	}

	return urls, nil
}

// GetRecentURLs returns recently created URLs
func (s *URLService) GetRecentURLs(limit int) ([]models.URL, error) {
	var urls []models.URL
	if err := s.db.Order("created_at desc").Limit(limit).Find(&urls).Error; err != nil {
		return nil, err
	}
	return urls, nil
}

// GetURLsByDateRange returns URLs created within a date range
func (s *URLService) GetURLsByDateRange(startDate, endDate time.Time, offset, limit int) ([]models.URL, int64, error) {
	var urls []models.URL
	var total int64

	query := s.db.Model(&models.URL{}).Where("created_at >= ? AND created_at <= ?", startDate, endDate)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get URLs with pagination
	if err := query.Offset(offset).Limit(limit).Order("created_at desc").Find(&urls).Error; err != nil {
		return nil, 0, err
	}

	return urls, total, nil
}

// ValidateURLAccess checks if a URL is accessible and not deleted
func (s *URLService) ValidateURLAccess(code string) (*models.URL, error) {
	var url models.URL
	result := s.db.Where("code = ? AND deleted_at IS NULL", code).First(&url)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("URL not found or has been deleted")
		}
		return nil, result.Error
	}
	return &url, nil
}

// GetDuplicateURLs finds URLs with the same original URL
func (s *URLService) GetDuplicateURLs() (map[string][]models.URL, error) {
	var urls []models.URL
	if err := s.db.Find(&urls).Error; err != nil {
		return nil, err
	}

	duplicates := make(map[string][]models.URL)
	urlMap := make(map[string][]models.URL)

	// Group URLs by original URL
	for _, url := range urls {
		urlMap[url.OriginalURL] = append(urlMap[url.OriginalURL], url)
	}

	// Find duplicates
	for originalURL, urlList := range urlMap {
		if len(urlList) > 1 {
			duplicates[originalURL] = urlList
		}
	}

	return duplicates, nil
}

// CleanupExpiredURLs removes URLs that haven't been clicked in a specified time
func (s *URLService) CleanupExpiredURLs(daysInactive int) (int64, error) {
	cutoffDate := time.Now().AddDate(0, 0, -daysInactive)

	// Find URLs that haven't been clicked since cutoff date and have no recent clicks
	var expiredURLs []string
	s.db.Model(&models.URL{}).
		Select("code").
		Where("click_count = 0 AND created_at < ?", cutoffDate).
		Or("code NOT IN (SELECT DISTINCT url_code FROM clicks WHERE clicked_at > ?)", cutoffDate).
		Pluck("code", &expiredURLs)

	if len(expiredURLs) == 0 {
		return 0, nil
	}

	// Soft delete expired URLs
	result := s.db.Where("code IN ?", expiredURLs).Delete(&models.URL{})
	return result.RowsAffected, result.Error
}

// GetURLUsageReport generates a usage report for URLs
func (s *URLService) GetURLUsageReport() (*models.URLUsageReport, error) {
	var report models.URLUsageReport

	// Total URLs
	s.db.Model(&models.URL{}).Count(&report.TotalURLs)

	// Active URLs (clicked at least once)
	s.db.Model(&models.URL{}).Where("click_count > 0").Count(&report.ActiveURLs)

	// URLs created today
	today := time.Now().Truncate(24 * time.Hour)
	s.db.Model(&models.URL{}).Where("created_at >= ?", today).Count(&report.URLsCreatedToday)

	// URLs created this week
	weekStart := today.AddDate(0, 0, -int(today.Weekday()))
	s.db.Model(&models.URL{}).Where("created_at >= ?", weekStart).Count(&report.URLsCreatedThisWeek)

	// URLs created this month
	monthStart := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
	s.db.Model(&models.URL{}).Where("created_at >= ?", monthStart).Count(&report.URLsCreatedThisMonth)

	// Most popular URL
	var topURL models.URL
	if err := s.db.Order("click_count desc").First(&topURL).Error; err == nil {
		report.MostPopularURL = topURL.Code
		report.MostPopularURLClicks = topURL.ClickCount
	}

	// Average clicks per URL
	var avgClicks float64
	s.db.Model(&models.URL{}).Select("AVG(click_count)").Scan(&avgClicks)
	report.AvgClicksPerURL = avgClicks

	// URLs by creation source
	s.db.Model(&models.URL{}).Where("created_by_api_key != ''").Count(&report.URLsCreatedByAPI)
	report.URLsCreatedByWeb = report.TotalURLs - report.URLsCreatedByAPI

	return &report, nil
}
