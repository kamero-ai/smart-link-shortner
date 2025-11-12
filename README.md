# Kamero Smart Link Shortener

A powerful, self-hosted URL shortener service built with Go and PostgreSQL. Create short links with platform-specific redirects, comprehensive analytics, and API key management.

**Repository:** [https://github.com/kamero-ai/smart-link-shortner](https://github.com/kamero-ai/smart-link-shortner)

## 🛠️ Tech Stack

### Backend
- **[Go](https://golang.org/)** 1.23+ - Programming language
- **[Gin](https://gin-gonic.com/)** - High-performance HTTP web framework
- **[GORM](https://gorm.io/)** - ORM library for database operations
- **[PostgreSQL](https://www.postgresql.org/)** - Relational database

### Libraries & Tools
- **[godotenv](https://github.com/joho/godotenv)** - Environment variable management
- **crypto/sha256** - API key hashing and URL hashing
- **crypto/rand** - Secure random number generation

### Frontend
- **HTML/CSS/JavaScript** - Admin dashboard and web interface
- **Vanilla JavaScript** - No framework dependencies

### DevOps & Deployment
- **[Docker](https://www.docker.com/)** - Containerization
- **Docker Compose** - Multi-container orchestration
- **Kubernetes** - Container orchestration (optional)

### Development Tools
- **[Air](https://github.com/cosmtrek/air)** - Live reload for Go applications

## ✨ Features

### Core Functionality
- **Smart URL Shortening**: Generate short 6-character codes for any URL
- **Platform-Specific Redirects**: Automatically redirect users based on their device (iOS, Android, Desktop, Mac)
- **Duplicate Detection**: Automatically reuses existing short URLs for the same destination
- **Click Analytics**: Track clicks with detailed information including:
  - Platform detection (iOS, Android, Desktop, Mac)
  - Browser and OS information
  - Geographic location (Country, City)
  - Referrer tracking
  - Timestamp tracking

### API & Authentication
- **RESTful API**: Complete API for programmatic access
- **API Key Management**: Create and manage API keys for authenticated requests
- **Optional Authentication**: Public endpoints work without API keys, with enhanced features for authenticated users
- **Admin Dashboard**: Web-based admin interface for managing URLs and API keys

### Analytics & Reporting
- **Real-time Analytics**: View click statistics in real-time
- **Platform Statistics**: Breakdown of clicks by platform, browser, and OS
- **Geographic Analytics**: Track clicks by country and city
- **Referrer Tracking**: See where your traffic is coming from
- **Time-based Trends**: Hourly and daily click trends
- **System-wide Statistics**: Overall system health and usage metrics

### Admin Features
- **Admin Dashboard**: Beautiful web interface for managing your shortener
- **API Key Management**: Create, view, and deactivate API keys
- **URL Management**: View all URLs, filter by source, platform, and more
- **Bulk Operations**: Delete multiple URLs at once
- **Export Functionality**: Export analytics data in JSON or CSV format

### Security & Performance
- **Secure Authentication**: Basic HTTP authentication for admin routes
- **API Key Security**: Hashed API key secrets stored securely
- **CORS Support**: Configurable CORS for cross-origin requests
- **Database Optimization**: Optimized connection pooling and indexes
- **Soft Deletes**: URLs and API keys are soft-deleted for data recovery

## 🚀 Quick Start

### Prerequisites

- **Go 1.23+**: [Install Go](https://golang.org/doc/install)
- **PostgreSQL 12+**: [Install PostgreSQL](https://www.postgresql.org/download/)
- **Git**: For cloning the repository

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/kamero-ai/smart-link-shortner.git
   cd smart-link-shortner
   ```

2. **Set up the database**
   ```bash
   # Create a PostgreSQL database
   createdb urlshortener
   
   # Run migrations (optional, auto-migration is enabled)
   psql -d urlshortener -f migrations/init.sql
   ```

3. **Configure environment variables**
   ```bash
   # Copy the example environment file
   cp .env.example .env
   
   # Edit .env with your configuration
   nano .env
   ```

4. **Install dependencies**
   ```bash
   go mod download
   ```

5. **Run the application**
   ```bash
   go run main.go
   ```

The server will start on `http://localhost:8080` (or your configured port).

## 🛠️ Development Mode

### Running in Development

1. **Set up your `.env` file** (copy from `.env.example` and configure):
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

2. **Start the development server**:
   ```bash
   # Using Go directly
   go run main.go
   
   # Or using Air for hot reload (if installed)
   air
   ```

3. **Access the application**:
   - **Web Interface**: http://localhost:8080
   - **Admin Dashboard**: http://localhost:8080/admin
   - **API Documentation**: See API section below

### Development Tools

- **Air (Hot Reload)**: Install [Air](https://github.com/cosmtrek/air) for automatic reloading during development
  ```bash
   go install github.com/cosmtrek/air@latest
   air
   ```

- **Database Migrations**: The application auto-migrates on startup, but you can also run migrations manually:
  ```bash
   psql -d urlshortener -f migrations/init.sql
   ```

## 🏠 Self-Hosting

### Docker Deployment

1. **Build the Docker image**:
   ```bash
   docker build -t kamero-url-shortener .
   ```

2. **Run with Docker Compose** (recommended):
   ```bash
   # Copy and configure the docker-compose template
   cp docker-compose.yml.example docker-compose.yml
   # Edit docker-compose.yml with your configuration
   docker-compose up -d
   ```

### Manual Deployment

1. **Build the binary**:
   ```bash
   go build -o kamero-url-shortener main.go
   ```

2. **Configure environment variables** (use systemd, supervisor, or your process manager):
   - Copy `.env.example` to `.env` and configure all variables
   - Or set environment variables directly in your process manager

3. **Run the application**:
   ```bash
   ./kamero-url-shortener
   ```

## 📡 API Documentation

Complete API documentation is available in the [docs folder](docs/API.md).

The API provides endpoints for:
- Creating and managing short URLs
- Retrieving analytics and statistics
- Managing API keys
- Admin operations

See [docs/API.md](docs/API.md) for detailed endpoint documentation, request/response examples, and code samples in multiple languages.

## 🌐 Web Interface

### Public Pages
- **Home**: `/` - Create short URLs via web interface

### Protected Pages (Require Admin Authentication)
- **Admin Dashboard**: `/admin` - API key management
- **Analytics Dashboard**: `/admin/analytics` - System-wide analytics
- **User Dashboard**: `/dashboard` - Individual URL analytics

## 🗄️ Database Schema

The application uses three main tables:

- **urls**: Stores shortened URLs with platform-specific redirects
- **clicks**: Tracks all click events with analytics data
- **api_keys**: Manages API keys for authenticated access

See `migrations/init.sql` for the complete schema.

## 🔧 Configuration

All configuration is done via environment variables. See `.env.example` for all available options and their descriptions.

## 🚧 Upcoming Features

We're actively working on these exciting features:

- **📊 Detailed Analytics**: Enhanced analytics with deeper insights, custom date ranges, and advanced filtering
- **👥 Demographics Tracking**: Track user demographics including age groups, interests, and behavioral patterns
- **📱 QR Code Generation**: Generate QR codes for your short links with custom logo/branding support

Contributions are welcome! If you'd like to help implement these features or suggest new ones, please see the Contributing section below.

## 🤝 Contributing

Contributions are welcome! We're open source and appreciate any help you can provide.

### How to Contribute

1. **Fork the repository**
2. **Create a feature branch**
   ```bash
   git checkout -b feature/amazing-feature
   ```
3. **Make your changes**
4. **Test your changes**
5. **Commit your changes**
   ```bash
   git commit -m "Add amazing feature"
   ```
6. **Push to your fork**
   ```bash
   git push origin feature/amazing-feature
   ```
7. **Open a Pull Request**

### Contribution Guidelines

- Follow Go code style guidelines
- Add tests for new features
- Update documentation as needed
- Ensure all tests pass
- Write clear commit messages

### Areas for Contribution

- **Upcoming Features**: Help implement detailed analytics, demographics tracking, and QR code generation
- **Performance Improvements**: Database query optimization
- **New Features**: Platform detection improvements, custom domains, etc.
- **Documentation**: Improve docs, add examples
- **Testing**: Add more test coverage
- **UI/UX**: Improve admin dashboard design

### Reporting Issues

If you find a bug or have a feature request, please open an issue on GitHub with:
- Clear description of the problem/feature
- Steps to reproduce (for bugs)
- Expected vs actual behavior
- Environment details (OS, Go version, etc.)

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

You are free to use, modify, distribute, and use this software for any purpose, including commercial use, without restrictions.

## 🙏 Acknowledgments

- Built with [Gin](https://gin-gonic.com/) web framework
- Database powered by [PostgreSQL](https://www.postgresql.org/)
- ORM by [GORM](https://gorm.io/)

## 📞 Support

For support, please open an issue on GitHub or contact the maintainers.

---

**Made with ❤️ by the Kamero team**

