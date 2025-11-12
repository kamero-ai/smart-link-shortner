package models

import (
	"time"

	"gorm.io/gorm"
)

type URL struct {
	ID                 uint           `json:"id" gorm:"primaryKey"`
	Code               string         `json:"code" gorm:"uniqueIndex;size:6"`
	OriginalURL        string         `json:"original_url" gorm:"not null;index"`
	URLHash            string         `json:"url_hash" gorm:"uniqueIndex;size:64"` // SHA256 hash of original URL + platform URLs
	IOSRedirectURL     string         `json:"ios_redirect_url"`
	AndroidRedirectURL string         `json:"android_redirect_url"`
	DesktopRedirectURL string         `json:"desktop_redirect_url"`
	MacRedirectURL     string         `json:"mac_redirect_url"`
	ClickCount         int64          `json:"click_count" gorm:"default:0"`
	CreatedByAPIKey    string         `json:"created_by_api_key" gorm:"index"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

type Click struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	URLCode   string    `json:"url_code" gorm:"index"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	Platform  string    `json:"platform"`
	Browser   string    `json:"browser"`
	OS        string    `json:"os"`
	Country   string    `json:"country"`
	City      string    `json:"city"`
	Referrer  string    `json:"referrer"`
	ClickedAt time.Time `json:"clicked_at"`
}

type APIKey struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	KeyID       string         `json:"key_id" gorm:"uniqueIndex;size:20"`
	KeySecret   string         `json:"key_secret" gorm:"size:64"` // Hashed
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	LastUsedAt  *time.Time     `json:"last_used_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// Request/Response models
type ShortenRequest struct {
	URL                string `json:"url" binding:"required,url"`
	IOSRedirectURL     string `json:"ios_redirect_url"`
	AndroidRedirectURL string `json:"android_redirect_url"`
	DesktopRedirectURL string `json:"desktop_redirect_url"`
	MacRedirectURL     string `json:"mac_redirect_url"`
}

type ShortenResponse struct {
	Code        string `json:"code"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	IsNew       bool   `json:"is_new"` // Indicates if this is a new URL or existing one
}

type APIKeyRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type APIKeyResponse struct {
	KeyID       string    `json:"key_id"`
	KeySecret   string    `json:"key_secret"` // Only returned on creation
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

// Analytics models
type AnalyticsResponse struct {
	Code          string           `json:"code"`
	OriginalURL   string           `json:"original_url"`
	ClickCount    int64            `json:"click_count"`
	CreatedAt     time.Time        `json:"created_at"`
	PlatformStats map[string]int64 `json:"platform_stats"`
	BrowserStats  map[string]int64 `json:"browser_stats"`
	OSStats       map[string]int64 `json:"os_stats"`
	RecentClicks  []Click          `json:"recent_clicks"`
	GeoStats      map[string]int64 `json:"geo_stats,omitempty"`
	ReferrerStats map[string]int64 `json:"referrer_stats,omitempty"`
	HourlyTrends  []HourlyTrend    `json:"hourly_trends,omitempty"`
	DailyTrends   []DailyTrend     `json:"daily_trends,omitempty"`
}

// URLAnalyticsSummary represents analytics data for a single URL
type URLAnalyticsSummary struct {
	Code          string           `json:"code"`
	OriginalURL   string           `json:"original_url"`
	ClickCount    int64            `json:"click_count"`
	CreatedAt     time.Time        `json:"created_at"`
	CreatedByAPI  bool             `json:"created_by_api"`
	APIKeyID      string           `json:"api_key_id,omitempty"`
	PlatformStats map[string]int64 `json:"platform_stats"`
	RecentClicks  []Click          `json:"recent_clicks"`
	HourlyTrends  []HourlyTrend    `json:"hourly_trends,omitempty"`
	DailyTrends   []DailyTrend     `json:"daily_trends,omitempty"`
}

// SystemStats represents enhanced system-wide statistics
type SystemStats struct {
	TotalURLs         int64            `json:"total_urls"`
	TotalClicks       int64            `json:"total_clicks"`
	URLsToday         int64            `json:"urls_today"`
	ClicksToday       int64            `json:"clicks_today"`
	URLsThisWeek      int64            `json:"urls_this_week"`
	ClicksThisWeek    int64            `json:"clicks_this_week"`
	URLsThisMonth     int64            `json:"urls_this_month"`
	ClicksThisMonth   int64            `json:"clicks_this_month"`
	URLsWithAPIKey    int64            `json:"urls_with_api_key"`
	URLsWithoutAPIKey int64            `json:"urls_without_api_key"`
	TopPlatforms      map[string]int64 `json:"top_platforms"`
	TopBrowsers       map[string]int64 `json:"top_browsers"`
	AvgClicksPerURL   float64          `json:"avg_clicks_per_url"`
	TopURLToday       string           `json:"top_url_today"`
	DailyTrends       []DailyTrend     `json:"daily_trends"`
	HourlyTrends      []HourlyTrend    `json:"hourly_trends"`
}

// Trend models
type DailyTrend struct {
	Date   time.Time `json:"date"`
	URLs   int64     `json:"urls"`
	Clicks int64     `json:"clicks"`
}

type HourlyTrend struct {
	Hour  int   `json:"hour"`
	Count int64 `json:"count"`
}

// Activity models
type RecentActivity struct {
	RecentURLs   []URL   `json:"recent_urls"`
	RecentClicks []Click `json:"recent_clicks"`
}

// Performance metrics
type PerformanceMetrics struct {
	AvgResponseTime  float64 `json:"avg_response_time"` // in milliseconds
	SuccessRate      float64 `json:"success_rate"`      // percentage
	UptimePercentage float64 `json:"uptime_percentage"` // percentage
	RequestsToday    int64   `json:"requests_today"`    // total requests today
	ErrorRate        float64 `json:"error_rate"`        // percentage
}

// API Key usage tracking
type APIKeyUsage struct {
	KeyID      string     `json:"key_id"`
	KeyName    string     `json:"key_name"`
	URLCount   int64      `json:"url_count"`
	ClickCount int64      `json:"click_count"`
	LastUsed   *time.Time `json:"last_used"`
}

// Enhanced analytics request
type AnalyticsRequest struct {
	Code        string    `json:"code,omitempty"`
	StartDate   time.Time `json:"start_date,omitempty"`
	EndDate     time.Time `json:"end_date,omitempty"`
	Granularity string    `json:"granularity,omitempty"` // hour, day, week, month
	Platform    string    `json:"platform,omitempty"`
	Country     string    `json:"country,omitempty"`
}

// Bulk analytics response
type BulkAnalyticsResponse struct {
	URLs       []URLAnalyticsSummary `json:"urls"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	Limit      int                   `json:"limit"`
	TotalPages int                   `json:"total_pages"`
}

// Export data structure
type ExportData struct {
	URLs       []URL                  `json:"urls"`
	Clicks     []Click                `json:"clicks"`
	ExportedAt time.Time              `json:"exported_at"`
	Filters    map[string]interface{} `json:"filters"`
}

// Real-time statistics
type RealTimeStats struct {
	ActiveUsers    int64            `json:"active_users"` // Users in last 5 minutes
	ClicksLastHour int64            `json:"clicks_last_hour"`
	URLsLastHour   int64            `json:"urls_last_hour"`
	TopCountries   map[string]int64 `json:"top_countries"`
	TopReferrers   map[string]int64 `json:"top_referrers"`
	LiveClicks     []Click          `json:"live_clicks"` // Last 10 clicks
}

// Geographic analytics
type GeoAnalytics struct {
	CountryStats map[string]GeoStat `json:"country_stats"`
	CityStats    map[string]GeoStat `json:"city_stats"`
	TotalClicks  int64              `json:"total_clicks"`
}

type GeoStat struct {
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
	Growth     float64 `json:"growth"` // Compared to previous period
}

// Device analytics
type DeviceAnalytics struct {
	PlatformBreakdown map[string]DeviceStat `json:"platform_breakdown"`
	BrowserBreakdown  map[string]DeviceStat `json:"browser_breakdown"`
	OSBreakdown       map[string]DeviceStat `json:"os_breakdown"`
	MobileVsDesktop   MobileDesktopStat     `json:"mobile_vs_desktop"`
}

type DeviceStat struct {
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
	Growth     float64 `json:"growth"`
}

type MobileDesktopStat struct {
	Mobile  DeviceStat `json:"mobile"`
	Desktop DeviceStat `json:"desktop"`
}

// Advanced filtering
type AnalyticsFilter struct {
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	Platform  string     `json:"platform,omitempty"`
	Browser   string     `json:"browser,omitempty"`
	OS        string     `json:"os,omitempty"`
	Country   string     `json:"country,omitempty"`
	City      string     `json:"city,omitempty"`
	Referrer  string     `json:"referrer,omitempty"`
	APIKeyID  string     `json:"api_key_id,omitempty"`
	MinClicks int64      `json:"min_clicks,omitempty"`
	MaxClicks int64      `json:"max_clicks,omitempty"`
	SortBy    string     `json:"sort_by,omitempty"`    // clicks, created_at, code
	SortOrder string     `json:"sort_order,omitempty"` // asc, desc
	Page      int        `json:"page,omitempty"`
	Limit     int        `json:"limit,omitempty"`
}

// Dashboard configuration
type DashboardConfig struct {
	DefaultTimeRange string   `json:"default_time_range"` // 24h, 7d, 30d, 90d
	DefaultCharts    []string `json:"default_charts"`     // platform, browser, geo, trends
	RefreshInterval  int      `json:"refresh_interval"`   // seconds
	MaxDataPoints    int      `json:"max_data_points"`    // for charts
}

// URLStats represents detailed statistics for a single URL
type URLStats struct {
	Code            string     `json:"code"`
	OriginalURL     string     `json:"original_url"`
	CreatedAt       time.Time  `json:"created_at"`
	ClickCount      int64      `json:"click_count"`
	TotalClicks     int64      `json:"total_clicks"`
	ClicksToday     int64      `json:"clicks_today"`
	ClicksThisWeek  int64      `json:"clicks_this_week"`
	ClicksThisMonth int64      `json:"clicks_this_month"`
	UniqueVisitors  int64      `json:"unique_visitors"`
	LastClickAt     *time.Time `json:"last_click_at"`
	AvgClicksPerDay float64    `json:"avg_clicks_per_day"`
}

// URLUsageReport represents overall URL usage statistics
type URLUsageReport struct {
	TotalURLs            int64   `json:"total_urls"`
	ActiveURLs           int64   `json:"active_urls"`
	URLsCreatedToday     int64   `json:"urls_created_today"`
	URLsCreatedThisWeek  int64   `json:"urls_created_this_week"`
	URLsCreatedThisMonth int64   `json:"urls_created_this_month"`
	URLsCreatedByAPI     int64   `json:"urls_created_by_api"`
	URLsCreatedByWeb     int64   `json:"urls_created_by_web"`
	MostPopularURL       string  `json:"most_popular_url"`
	MostPopularURLClicks int64   `json:"most_popular_url_clicks"`
	AvgClicksPerURL      float64 `json:"avg_clicks_per_url"`
}
