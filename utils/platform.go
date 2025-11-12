package utils

import (
	"net/http"
	"strings"

	"url-shortener/models"
)

type PlatformInfo struct {
	Platform string
	OS       string
	Browser  string
}

func DetectPlatform(r *http.Request) PlatformInfo {
	userAgent := strings.ToLower(r.Header.Get("User-Agent"))

	info := PlatformInfo{}

	// Detect Platform
	if strings.Contains(userAgent, "iphone") || strings.Contains(userAgent, "ipad") {
		info.Platform = "ios"
		info.OS = "iOS"
	} else if strings.Contains(userAgent, "android") {
		info.Platform = "android"
		info.OS = "Android"
	} else if strings.Contains(userAgent, "macintosh") || strings.Contains(userAgent, "mac os") {
		info.Platform = "mac"
		info.OS = "macOS"
	} else if strings.Contains(userAgent, "windows") {
		info.Platform = "desktop"
		info.OS = "Windows"
	} else if strings.Contains(userAgent, "linux") {
		info.Platform = "desktop"
		info.OS = "Linux"
	} else {
		info.Platform = "desktop"
		info.OS = "Unknown"
	}

	// Detect Browser
	if strings.Contains(userAgent, "chrome") {
		info.Browser = "Chrome"
	} else if strings.Contains(userAgent, "firefox") {
		info.Browser = "Firefox"
	} else if strings.Contains(userAgent, "safari") && !strings.Contains(userAgent, "chrome") {
		info.Browser = "Safari"
	} else if strings.Contains(userAgent, "edge") {
		info.Browser = "Edge"
	} else {
		info.Browser = "Unknown"
	}

	return info
}

func GetRedirectURL(url *models.URL, platform string) string {
	switch platform {
	case "ios":
		if url.IOSRedirectURL != "" {
			return url.IOSRedirectURL
		}
	case "android":
		if url.AndroidRedirectURL != "" {
			return url.AndroidRedirectURL
		}
	case "mac":
		if url.MacRedirectURL != "" {
			return url.MacRedirectURL
		}
	case "desktop":
		if url.DesktopRedirectURL != "" {
			return url.DesktopRedirectURL
		}
	}
	return url.OriginalURL
}
