# Convoy

A collaborative route-sharing platform that enables people to plan, share, and join routes together. Perfect for group rides, carpools, running clubs, cycling groups, or any coordinated travel.

## Overview

Convoy allows users to:
- Create and share planned routes on a map
- Allow others to join and follow the same route
- Coordinate meetup points and timing
- Track participants in real-time

## Tech Stack

### Backend
- **Go 1.25+** - API server
- **PostgreSQL 14+** - Database
- **Mapbox API** - Map provider (with support for Google Maps as alternative)
- **JWT** - Authentication

### Frontend
- **React 18** - UI framework
- **Vite** - Build tool
- **TailwindCSS** - Styling
- **React Router** - Client-side routing

## Project Structure

```
convoy/
├── cmd/
│   └── server/          # Application entrypoint
├── internal/
│   ├── api/            # HTTP handlers and routes
│   ├── auth/           # Authentication logic
│   ├── models/         # Data models
│   ├── repository/     # Database layer
│   ├── service/        # Business logic
│   └── maps/           # Map provider integrations
├── web/               # React frontend application
│   ├── src/
│   │   ├── components/ # UI components
│   │   ├── pages/      # Page components
│   │   ├── contexts/   # React contexts
│   │   └── lib/        # API client & utilities
│   └── public/         # Static assets
├── migrations/         # Database migrations
├── scripts/           # Setup and utility scripts
├── docs/              # Documentation
│   ├── USER_STORIES.md
│   ├── ARCHITECTURE.md
│   └── API.md
└── tests/             # Integration tests
```

## Prerequisites

- **Docker & Docker Compose** - All services run in containers
- **Mapbox API key** (optional) - Map provider (free tier available)

## AI Usage

This project was created with the assistance of AI tools. I primarily used Claude 4.5 Sonnet to generate the initial skeleton of the code and documentation. I created the user stories and architecture files to guide the development process. I ended up using AI a bit more than I intended but it was helpful in getting started, and could allow me to play around with mapbox more. With AI being such a powerful tool, and adoption increasing, I felt it was a good idea to use it to get the frontend up and running since I am not as familiar with the current state of react.

## Quick Start

The easiest way to run Convoy is using Docker Compose - it runs everything (database, API, and web interface):

```bash
# 1. Clone the repository
git clone https://github.com/N3phi1im/convoy.git
cd convoy

# 2. Create .env file and configure
cp .env.example .env

# Edit .env and set your Mapbox token:
# MAPBOX_API_KEY=your_backend_mapbox_key
# VITE_MAPBOX_TOKEN=your_frontend_mapbox_key
# (Get free tokens at https://account.mapbox.com/access-tokens/)

# 3. Start all services (Database + API + Web)
docker compose up -d
```

**That's it! Access the application:**
- **Web Interface**: http://localhost:3000
- **API**: http://localhost:8080
- **Database**: localhost:5432

**View logs:**
```bash
docker compose logs -f        # All services
docker compose logs -f web    # Web only
docker compose logs -f api    # API only
```

**Stop services:**
```bash
docker compose down           # Stop all services
docker compose down -v        # Stop and remove volumes
```

## Manual Setup

### Backend Setup

#### 1. Environment Configuration

Create a `.env` file in the project root:

```env
# Server
PORT=8080
ENV=development

# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=convoy
DB_PASSWORD=stuffnthings
DB_NAME=convoy_db
DB_SSLMODE=disable

# Map Provider
MAPBOX_API_KEY=your_mapbox_api_key

# JWT
JWT_SECRET=your_jwt_secret_change_in_production
JWT_EXPIRY=24h

# CORS
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173
```

#### 2. Run

```bash
# Starts database, runs migrations automatically, and starts API
docker compose up -d

# View logs
docker compose logs -f api
```

The API will be available at `http://localhost:8080`

**Reset Database (if needed):**
```bash
# WARNING: This deletes all data
docker compose down -v
docker compose up -d
```

### Frontend Setup

#### 1. Install Dependencies

```bash
cd web
npm install
```

#### 2. Configure Environment

Create `web/.env`:

```env
VITE_API_URL=http://localhost:8080/api/v1
VITE_MAPBOX_TOKEN=your_mapbox_api_key
```

#### 3. Start Development Server

```bash
npm run dev
```

The web interface will be available at `http://localhost:5173`

#### 4. Build for Production

```bash
npm run build
# Output will be in web/dist/
```

## API Endpoints

See [docs/API.md](docs/API.md) for detailed API documentation.

## Development

### Running Tests

**Unit Tests (mocked dependencies):**
```bash
# Run all unit tests
go test ./...

# Verbose output
go test ./... -v

# With coverage
go test ./... -cover
```

### Code Quality

```bash
# Format code
go fmt ./...
```

### Mapbox Features Used

- **Directions API**: Route planning and optimization
- **Geocoding API**: Address to coordinates conversion

## Deployment

### Docker

```bash
docker build -t convoy:latest .
docker compose up
```

### Production Considerations

- Use environment variables for all secrets
- Enable HTTPS/TLS
- Set up database connection pooling
- Implement rate limiting
- Configure CORS appropriately
- Set up monitoring and logging

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see LICENSE file for details

## Roadmap

- [ ] Phase 1: Core route creation and sharing
- [ ] Phase 2: Mobile app (React Native)
- [ ] Phase 3: Real-time participant tracking

## Support

For issues and questions, please open an issue on GitHub.
