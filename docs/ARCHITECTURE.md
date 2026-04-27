# Convoy - Technical Architecture

## System Overview

Convoy is a RESTful API service built in Go that enables collaborative route planning and sharing. The system follows clean architecture principles with clear separation of concerns.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                         Clients                             │
│              (Web App, Mobile App, API Consumers)           │
└────────────────────────┬────────────────────────────────────┘
                         │
                         │ HTTPS/REST
                         │
┌────────────────────────▼────────────────────────────────────┐
│                    Router                                   │
└────────────────────────┬────────────────────────────────────┘
                         │
         ┌───────────────┼───────────────┐
         │               │               │
┌────────▼────────┐ ┌────▼───────┐ ┌─────▼─────────┐
│  Auth Middleware│ │Rate Limiter│ │CORS Middleware│
└────────┬────────┘ └────────────┘ └───────────────┘
         │
┌────────▼──────────────────────────────────────────────────┐
│                    Handler Layer                          │
│  (HTTP Request/Response, Validation, Serialization)       │
└────────┬──────────────────────────────────────────────────┘
         │
┌────────▼──────────────────────────────────────────────────┐
│                    Service Layer                          │
│                   (Business Logic)                        │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐                 │
│  │Auth Svc  │  │Route Svc │  │User Svc  │                 │
│  └──────────┘  └──────────┘  └──────────┘                 │
└────────┬──────────────────────────────────────────────────┘
         │
    ┌────┴────┐
    │         │
┌───▼────┐ ┌──▼─────────┐
│Repo    │ │Map Provider│
│Layer   │ │  Service   │
└───┬────┘ └─┬──────────┘
    │        │
┌───▼────┐ ┌─▼─────────┐
│Postgres│ │Mapbox API │
│Database│ │           │
└────────┘ └───────────┘
```

## Technology Stack

### Backend
- **Language**: Go 1.25+
- **Web Framework**: Chi Router (lightweight, idiomatic)
- **Database**: PostgreSQL 14+
- **ORM**: sqlx (lightweight) or GORM (feature-rich)
- **Authentication**: JWT (golang-jwt/jwt)
- **Validation**: go-playground/validator
- **Configuration**: viper
- **Logging**: zap or zerolog
- **Testing**: testify, httptest

### External Services
- **Map Provider**: Mapbox API
  - Directions API
  - Geocoding API
  - Static Maps API
- **Email**: SendGrid or AWS SES (future)
- **File Storage**: AWS S3 or local (profile pictures)

### Infrastructure
- **Containerization**: Docker
- **Orchestration**: Docker Compose (dev), Kubernetes (prod)
- **CI/CD**: GitHub Actions
- **Monitoring**: Prometheus + Grafana
- **Logging**: ELK Stack or Loki

## Project Structure

```
convoy/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
│
├── internal/                    # Private application code
│   ├── api/
│   │   ├── handlers/           # HTTP handlers
│   │   │   ├── auth.go
│   │   │   ├── routes.go
│   │   │   └── users.go
│   │   ├── middleware/         # HTTP middleware
│   │   │   ├── auth.go
│   │   │   ├── cors.go
│   │   │   ├── logging.go
│   │   │   └── ratelimit.go
│   │   └── router.go           # Route definitions
│   │
│   ├── models/                 # Domain models
│   │   ├── user.go
│   │   ├── route.go
│   │   ├── waypoint.go
│   │   └── participant.go
│   │
│   ├── repository/             # Data access layer
│   │   ├── postgres/
│   │   │   ├── user.go
│   │   │   ├── route.go
│   │   │   └── participant.go
│   │   └── repository.go       # Repository interfaces
│   │
│   ├── service/                # Business logic
│   │   ├── auth.go
│   │   ├── route.go
│   │   └── user.go
│   │
│   ├── maps/                   # Map provider integrations
│   │   ├── provider.go         # Provider interface
│   │   └── mapbox/
│   │       ├── client.go
│   │       ├── directions.go
│   │       └── geocoding.go
│   │
│   ├── auth/                   # Authentication utilities
│   │   ├── jwt.go
│   │   └── password.go
│   │
│   └── config/                 # Configuration
│       └── config.go
│
├── pkg/                        # Public, reusable packages
│   ├── errors/                 # Error types
│   │   └── errors.go
│   ├── logger/                 # Logging utilities
│   │   └── logger.go
│   └── validator/              # Validation utilities
│       └── validator.go
│
├── migrations/                 # Database migrations
│
├── config/                     # Configuration files
│   ├── config.yaml
│   └── config.example.yaml
│
├── scripts/                    # Utility scripts
│   ├── setup.sh
│   └── seed.sh
│
├── tests/                      # Integration tests
│   ├── integration/
│   │   ├── auth_test.go
│   │   └── routes_test.go
│   └── fixtures/
│       └── test_data.sql
│
├── docs/                       # Documentation
│   ├── API.md
│   ├── ARCHITECTURE.md
│   ├── USER_STORIES.md
│   └── MAP_INTEGRATION.md
│
├── .env.example
├── .gitignore
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
└── README.md
```

## Security Considerations

### Authentication
- JWT tokens with reasonable expiry (24h)
- Refresh token mechanism
- Password hashing with bcrypt (cost 12)
- Rate limiting on auth endpoints

### Authorization
- Route creators have full control
- Participants have read access
- Public routes visible to all
- Private routes require authentication

### Data Protection
- HTTPS only in production
- SQL injection prevention (parameterized queries)
- XSS prevention (input sanitization)
- CORS configuration
- Environment-based secrets

### Rate Limiting
- 100 requests/minute per IP for general endpoints
- 5 requests/minute for auth endpoints
- 1000 requests/hour per authenticated user

### Health Checks
- `/health` - Basic health check

## Testing Strategy

### Unit Tests
- Service layer logic
- Repository layer
- Utility functions
- 80%+ code coverage target

### Integration Tests
- API endpoint testing
- Database operations
- External API mocking

### E2E Tests
- Critical user flows
- Authentication flow
- Route creation and joining

## Future Enhancements

### Phase 2
- WebSocket support for real-time updates
- Push notifications
- Mobile app (React Native)

### Phase 3
- Real-time location tracking
- Route analytics and insights
