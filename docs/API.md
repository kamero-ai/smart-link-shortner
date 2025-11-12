# API Documentation

Complete API reference for Kamero Smart Link Shortener.

## Base URL

```
http://localhost:8080/api/v1
```

Replace `localhost:8080` with your server's hostname and port in production.

## Authentication

### API Key Authentication

API keys are optional for most endpoints but provide enhanced features. When authenticated, you can:
- Track which URLs were created by your API key
- Retrieve all URLs created with your API key
- Access additional analytics features

**Format:**
```
Authorization: Bearer KEY_ID:KEY_SECRET
```

**Example:**
```http
Authorization: Bearer abc123xyz:secret_key_here
```

### Basic Authentication (Admin Endpoints)

Admin endpoints require Basic HTTP Authentication using your admin credentials.

**Format:**
```
Authorization: Basic base64(username:password)
```

**Example:**
```http
Authorization: Basic YWRtaW46cGFzc3dvcmQ=
```

## Public API Endpoints

### Create Short URL

Create a new short URL. If a URL with the same destination already exists, the existing short URL is returned.

**Endpoint:** `POST /api/v1/shorten`

**Authentication:** Optional (API key recommended)

**Request Body:**
```json
{
  "url": "https://example.com",
  "ios_redirect_url": "https://apps.apple.com/app/123456",
  "android_redirect_url": "https://play.google.com/store/apps/details?id=com.example",
  "desktop_redirect_url": "https://example.com/desktop",
  "mac_redirect_url": "https://example.com/mac"
}
```

**Request Fields:**
- `url` (required): The original URL to shorten
- `ios_redirect_url` (optional): Custom redirect URL for iOS devices
- `android_redirect_url` (optional): Custom redirect URL for Android devices
- `desktop_redirect_url` (optional): Custom redirect URL for desktop browsers
- `mac_redirect_url` (optional): Custom redirect URL for macOS

**Response (201 Created - New URL):**
```json
{
  "code": "abc123",
  "short_url": "http://localhost:8080/abc123",
  "original_url": "https://example.com",
  "is_new": true
}
```

**Response (200 OK - Existing URL):**
```json
{
  "code": "abc123",
  "short_url": "http://localhost:8080/abc123",
  "original_url": "https://example.com",
  "is_new": false
}
```

**Error Responses:**
- `400 Bad Request`: Invalid URL format or missing required fields
- `500 Internal Server Error`: Server error creating short URL

**Example Request:**
```bash
curl -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer KEY_ID:KEY_SECRET" \
  -d '{
    "url": "https://example.com",
    "ios_redirect_url": "https://apps.apple.com/app/123456"
  }'
```

### Get Analytics

Retrieve analytics data for a specific short URL.

**Endpoint:** `GET /api/v1/analytics/:code`

**Authentication:** Optional

**URL Parameters:**
- `code` (required): The 6-character short URL code

**Response (200 OK):**
```json
{
  "code": "abc123",
  "original_url": "https://example.com",
  "click_count": 42,
  "created_at": "2024-01-01T00:00:00Z",
  "platform_stats": {
    "ios": 10,
    "android": 15,
    "desktop": 17
  },
  "browser_stats": {
    "Chrome": 25,
    "Safari": 17
  },
  "os_stats": {
    "iOS": 10,
    "Android": 15,
    "Windows": 12,
    "macOS": 5
  },
  "recent_clicks": [
    {
      "id": 1,
      "url_code": "abc123",
      "ip_address": "192.168.1.1",
      "user_agent": "Mozilla/5.0...",
      "platform": "ios",
      "browser": "Safari",
      "os": "iOS",
      "country": "United States",
      "city": "New York",
      "referrer": "https://google.com",
      "clicked_at": "2024-01-15T10:30:00Z"
    }
  ],
  "geo_stats": {
    "United States": 25,
    "United Kingdom": 10,
    "Canada": 7
  },
  "referrer_stats": {
    "https://google.com": 15,
    "https://twitter.com": 10,
    "direct": 17
  }
}
```

**Error Responses:**
- `404 Not Found`: Short URL code not found

**Example Request:**
```bash
curl http://localhost:8080/api/v1/analytics/abc123
```

### Get Detailed Analytics

Get comprehensive analytics with hourly and daily trends.

**Endpoint:** `GET /api/v1/analytics/:code/detailed`

**Authentication:** Optional

**URL Parameters:**
- `code` (required): The 6-character short URL code

**Response:** Same as Get Analytics, with additional fields:
```json
{
  "code": "abc123",
  "original_url": "https://example.com",
  "click_count": 42,
  "created_at": "2024-01-01T00:00:00Z",
  "platform_stats": {...},
  "browser_stats": {...},
  "os_stats": {...},
  "recent_clicks": [...],
  "hourly_trends": [
    {
      "hour": 10,
      "count": 5
    },
    {
      "hour": 11,
      "count": 8
    }
  ],
  "daily_trends": [
    {
      "date": "2024-01-15T00:00:00Z",
      "urls": 0,
      "clicks": 10
    }
  ]
}
```

### Get My URLs

Retrieve all URLs created with your API key.

**Endpoint:** `GET /api/v1/my-urls`

**Authentication:** Required (API key)

**Response (200 OK):**
```json
[
  {
    "code": "abc123",
    "short_url": "http://localhost:8080/abc123",
    "original_url": "https://example.com",
    "is_new": false
  },
  {
    "code": "xyz789",
    "short_url": "http://localhost:8080/xyz789",
    "original_url": "https://another-example.com",
    "is_new": false
  }
]
```

**Error Responses:**
- `401 Unauthorized`: Missing or invalid API key
- `500 Internal Server Error`: Server error retrieving URLs

**Example Request:**
```bash
curl http://localhost:8080/api/v1/my-urls \
  -H "Authorization: Bearer KEY_ID:KEY_SECRET"
```

### Redirect to Original URL

Accessing a short URL directly redirects to the original URL based on the user's platform.

**Endpoint:** `GET /:code`

**Authentication:** Not required

**URL Parameters:**
- `code` (required): The 6-character short URL code

**Response:**
- `307 Temporary Redirect`: Redirects to the appropriate URL based on platform
- `404 Not Found`: Short URL code not found

**Platform Detection:**
The system automatically detects the user's platform and redirects to:
- iOS devices → `ios_redirect_url` (if set) or `original_url`
- Android devices → `android_redirect_url` (if set) or `original_url`
- macOS → `mac_redirect_url` (if set) or `original_url`
- Desktop → `desktop_redirect_url` (if set) or `original_url`

**Example:**
```
http://localhost:8080/abc123
→ Redirects to https://example.com (or platform-specific URL)
```

## Admin API Endpoints

All admin endpoints require Basic HTTP Authentication.

### Create API Key

Create a new API key for programmatic access.

**Endpoint:** `POST /admin/api/v1/api-keys`

**Authentication:** Required (Basic Auth)

**Request Body:**
```json
{
  "name": "My API Key",
  "description": "For mobile app integration"
}
```

**Request Fields:**
- `name` (required): A descriptive name for the API key
- `description` (optional): Additional description or notes

**Response (201 Created):**
```json
{
  "key_id": "abc123xyz",
  "key_secret": "secret_key_here_keep_this_safe",
  "name": "My API Key",
  "description": "For mobile app integration",
  "is_active": true,
  "created_at": "2024-01-15T10:30:00Z"
}
```

**⚠️ Important:** The `key_secret` is only returned once when the key is created. Store it securely.

**Error Responses:**
- `400 Bad Request`: Missing required fields
- `401 Unauthorized`: Invalid admin credentials
- `500 Internal Server Error`: Server error creating API key

**Example Request:**
```bash
curl -X POST http://localhost:8080/admin/api/v1/api-keys \
  -H "Content-Type: application/json" \
  -H "Authorization: Basic YWRtaW46cGFzc3dvcmQ=" \
  -d '{
    "name": "My API Key",
    "description": "For mobile app"
  }'
```

### List API Keys

Retrieve all API keys (without secrets).

**Endpoint:** `GET /admin/api/v1/api-keys`

**Authentication:** Required (Basic Auth)

**Response (200 OK):**
```json
[
  {
    "key_id": "abc123xyz",
    "name": "My API Key",
    "description": "For mobile app integration",
    "is_active": true,
    "created_at": "2024-01-15T10:30:00Z"
  },
  {
    "key_id": "def456uvw",
    "name": "Another Key",
    "description": "For web integration",
    "is_active": true,
    "created_at": "2024-01-14T09:20:00Z"
  }
]
```

**Example Request:**
```bash
curl http://localhost:8080/admin/api/v1/api-keys \
  -H "Authorization: Basic YWRtaW46cGFzc3dvcmQ="
```

### Deactivate API Key

Deactivate an API key (soft delete).

**Endpoint:** `DELETE /admin/api/v1/api-keys/:keyId`

**Authentication:** Required (Basic Auth)

**URL Parameters:**
- `keyId` (required): The API key ID to deactivate

**Response (200 OK):**
```json
{
  "message": "API key deactivated successfully"
}
```

**Error Responses:**
- `401 Unauthorized`: Invalid admin credentials
- `404 Not Found`: API key not found
- `500 Internal Server Error`: Server error deactivating key

**Example Request:**
```bash
curl -X DELETE http://localhost:8080/admin/api/v1/api-keys/abc123xyz \
  -H "Authorization: Basic YWRtaW46cGFzc3dvcmQ="
```

### Get All URLs Analytics

Retrieve analytics for all URLs with pagination and filtering.

**Endpoint:** `GET /admin/api/v1/urls/analytics`

**Authentication:** Required (Basic Auth)

**Query Parameters:**
- `page` (optional, default: 1): Page number for pagination
- `limit` (optional, default: 25, max: 100): Number of results per page
- `source` (optional): Filter by source - `api` or `web`
- `platform` (optional): Filter by platform - `ios`, `android`, `desktop`, `mac`
- `min_clicks` (optional): Minimum click count filter
- `search` (optional): Search in URL code or original URL

**Response (200 OK):**
```json
{
  "data": [
    {
      "code": "abc123",
      "original_url": "https://example.com",
      "click_count": 42,
      "created_at": "2024-01-01T00:00:00Z",
      "created_by_api": true,
      "api_key_id": "abc123xyz",
      "platform_stats": {
        "ios": 10,
        "android": 15,
        "desktop": 17
      },
      "recent_clicks": [...]
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 25,
    "total": 100,
    "total_pages": 4
  },
  "filters": {
    "source": "api",
    "platform": null,
    "min_clicks": null,
    "search": null
  },
  "base_url": "http://localhost:8080"
}
```

**Example Request:**
```bash
curl "http://localhost:8080/admin/api/v1/urls/analytics?page=1&limit=25&source=api" \
  -H "Authorization: Basic YWRtaW46cGFzc3dvcmQ="
```

### Get System Stats

Get system-wide statistics and metrics.

**Endpoint:** `GET /admin/api/v1/system/stats`

**Authentication:** Required (Basic Auth)

**Response (200 OK):**
```json
{
  "total_urls": 1000,
  "total_clicks": 50000,
  "urls_today": 25,
  "clicks_today": 500,
  "urls_this_week": 150,
  "clicks_this_week": 3500,
  "urls_this_month": 600,
  "clicks_this_month": 15000,
  "urls_with_api_key": 800,
  "urls_without_api_key": 200,
  "top_platforms": {
    "desktop": 20000,
    "android": 15000,
    "ios": 10000,
    "mac": 5000
  },
  "top_browsers": {
    "Chrome": 25000,
    "Safari": 15000,
    "Firefox": 5000,
    "Edge": 5000
  },
  "avg_clicks_per_url": 50.0,
  "top_url_today": "abc123",
  "daily_trends": [
    {
      "date": "2024-01-15T00:00:00Z",
      "urls": 25,
      "clicks": 500
    }
  ],
  "hourly_trends": [
    {
      "hour": 10,
      "count": 50
    }
  ]
}
```

**Example Request:**
```bash
curl http://localhost:8080/admin/api/v1/system/stats \
  -H "Authorization: Basic YWRtaW46cGFzc3dvcmQ="
```

## Health Check

Check if the service is running and healthy.

**Endpoint:** `GET /health`

**Authentication:** Not required

**Response (200 OK):**
```json
{
  "status": "ok",
  "service": "kamero-url-shortener",
  "version": "1.0.0",
  "base_url": "http://localhost:8080",
  "timestamp": "8080"
}
```

**Example Request:**
```bash
curl http://localhost:8080/health
```

## Error Handling

All API endpoints return standard HTTP status codes:

- `200 OK`: Request successful
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid request parameters or body
- `401 Unauthorized`: Authentication required or invalid credentials
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

Error responses follow this format:
```json
{
  "error": "Error message describing what went wrong"
}
```

## Rate Limiting

Rate limiting may be applied depending on your deployment configuration. Check with your administrator for specific rate limits.

## Best Practices

1. **Store API Keys Securely**: Never commit API keys to version control
2. **Use HTTPS**: Always use HTTPS in production
3. **Handle Errors**: Implement proper error handling for all API calls
4. **Cache Responses**: Cache analytics data when appropriate to reduce API calls
5. **Respect Rate Limits**: Implement exponential backoff for rate-limited requests
6. **Validate URLs**: Ensure URLs are properly formatted before sending requests

## Examples

### Complete Workflow Example

```bash
# 1. Create a short URL
curl -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer KEY_ID:KEY_SECRET" \
  -d '{
    "url": "https://example.com",
    "ios_redirect_url": "https://apps.apple.com/app/123456"
  }'

# Response: {"code": "abc123", "short_url": "http://localhost:8080/abc123", ...}

# 2. Get analytics
curl http://localhost:8080/api/v1/analytics/abc123

# 3. Get all your URLs
curl http://localhost:8080/api/v1/my-urls \
  -H "Authorization: Bearer KEY_ID:KEY_SECRET"
```

### Using with cURL

```bash
# Create short URL
curl -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}'

# Get analytics
curl http://localhost:8080/api/v1/analytics/abc123
```

### Using with JavaScript (Fetch API)

```javascript
// Create short URL
const response = await fetch('http://localhost:8080/api/v1/shorten', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer KEY_ID:KEY_SECRET'
  },
  body: JSON.stringify({
    url: 'https://example.com',
    ios_redirect_url: 'https://apps.apple.com/app/123456'
  })
});

const data = await response.json();
console.log(data.short_url);

// Get analytics
const analytics = await fetch('http://localhost:8080/api/v1/analytics/abc123');
const analyticsData = await analytics.json();
console.log(analyticsData);
```

### Using with Python (requests)

```python
import requests

# Create short URL
response = requests.post(
    'http://localhost:8080/api/v1/shorten',
    json={
        'url': 'https://example.com',
        'ios_redirect_url': 'https://apps.apple.com/app/123456'
    },
    headers={
        'Authorization': 'Bearer KEY_ID:KEY_SECRET'
    }
)

data = response.json()
print(data['short_url'])

# Get analytics
analytics = requests.get('http://localhost:8080/api/v1/analytics/abc123')
analytics_data = analytics.json()
print(analytics_data)
```

---

For more information, visit the [main README](../README.md) or open an issue on GitHub.

