# Convoy API Documentation

## Base URL
```
http://localhost:8080/api/v1
```

## Authentication

Most endpoints require authentication via JWT token in the Authorization header:

```
Authorization: Bearer <your_jwt_token>
```

## Response Format

### Success Response
```json
{
  "success": true,
  "data": { ... }
}
```

### Error Response
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message",
    "details": { ... }
  }
}
```

### Paginated Response
```json
{
  "success": true,
  "data": {
    "data": [...],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 100,
      "total_pages": 5
    }
  }
}
```

## Error Codes

| Code | Description |
|------|-------------|
| `VALIDATION_ERROR` | Request validation failed |
| `UNAUTHORIZED` | Authentication required or failed |
| `FORBIDDEN` | Insufficient permissions |
| `NOT_FOUND` | Resource not found |
| `CONFLICT` | Resource already exists |
| `INTERNAL_ERROR` | Server error |
| `RATE_LIMIT_EXCEEDED` | Too many requests |

---

## Authentication Endpoints

### Register User

Create a new user account.

**Endpoint:** `POST /auth/register`

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securepassword123",
  "display_name": "John Doe"
}
```

**Response:** `201 Created`
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "user@example.com",
      "display_name": "John Doe",
      "created_at": "2024-01-15T10:30:00Z"
    }
  }
}
```

**Errors:**
- `400` - Validation error (invalid email, weak password, etc.)
- `409` - Email already registered

---

### Login

Authenticate and receive a JWT token.

**Endpoint:** `POST /auth/login`

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "user@example.com",
      "display_name": "John Doe",
      "created_at": "2024-01-15T10:30:00Z"
    }
  }
}
```

**Errors:**
- `401` - Invalid credentials

---

## User Endpoints

### Get Current User

Get the authenticated user's profile.

**Endpoint:** `GET /users/me`

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "display_name": "John Doe",
    "profile_pic": "https://example.com/pic.jpg",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

---

### Update User Profile

Update the authenticated user's profile.

**Endpoint:** `PUT /users/me`

**Headers:** `Authorization: Bearer <token>`

**Request Body:**
```json
{
  "display_name": "Jane Doe",
  "profile_pic": "https://example.com/newpic.jpg"
}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "display_name": "Jane Doe",
    "profile_pic": "https://example.com/newpic.jpg",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

---

## Route Endpoints

### Create Route

Create a new route.

**Endpoint:** `POST /routes`

**Headers:** `Authorization: Bearer <token>`

**Request Body:**
```json
{
  "name": "Morning Bike Ride",
  "description": "Scenic route through the park",
  "route_type": "cycling",
  "visibility": "public",
  "start_time": "2024-01-20T08:00:00Z",
  "max_participants": 10,
  "difficulty": "moderate",
  "waypoints": [
    {
      "latitude": 37.7749,
      "longitude": -122.4194,
      "order": 0,
      "name": "Start Point",
      "address": "123 Main St, San Francisco, CA"
    },
    {
      "latitude": 37.7849,
      "longitude": -122.4094,
      "order": 1,
      "name": "Park Entrance"
    },
    {
      "latitude": 37.7949,
      "longitude": -122.3994,
      "order": 2,
      "name": "End Point"
    }
  ]
}
```

**Response:** `201 Created`
```json
{
  "success": true,
  "data": {
    "id": "660e8400-e29b-41d4-a716-446655440000",
    "creator_id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Morning Bike Ride",
    "description": "Scenic route through the park",
    "route_type": "cycling",
    "visibility": "public",
    "start_time": "2024-01-20T08:00:00Z",
    "max_participants": 10,
    "distance": 5432.1,
    "duration": 1234,
    "difficulty": "moderate",
    "status": "planned",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z",
    "creator": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "user@example.com",
      "display_name": "John Doe"
    },
    "waypoints": [...],
    "participant_count": 1
  }
}
```

**Errors:**
- `400` - Validation error (need at least 2 waypoints)
- `401` - Unauthorized

---

### Get Route

Get details of a specific route.

**Endpoint:** `GET /routes/:id`

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "id": "660e8400-e29b-41d4-a716-446655440000",
    "name": "Morning Bike Ride",
    "route_type": "cycling",
    "distance": 5432.1,
    "duration": 1234,
    "waypoints": [...],
    "participants": [...]
  }
}
```

**Errors:**
- `404` - Route not found
- `403` - Private route, access denied

---

### List Routes

List all public routes with filtering and pagination.

**Endpoint:** `GET /routes`

**Query Parameters:**
- `page` (int, default: 1) - Page number
- `limit` (int, default: 20, max: 100) - Items per page
- `route_type` (string) - Filter by type: driving, cycling, walking, running
- `visibility` (string) - Filter by visibility (admin only)
- `status` (string) - Filter by status: planned, in_progress, completed
- `search` (string) - Search in name and description
- `sort_by` (string) - Sort field: created_at, start_time, distance
- `sort_order` (string) - Sort order: asc, desc

**Example:** `GET /routes?page=1&limit=10&route_type=cycling&sort_by=start_time&sort_order=asc`

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "data": [
      {
        "id": "660e8400-e29b-41d4-a716-446655440000",
        "name": "Morning Bike Ride",
        "route_type": "cycling",
        "start_time": "2024-01-20T08:00:00Z",
        "distance": 5432.1,
        "participant_count": 5
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 45,
      "total_pages": 5
    }
  }
}
```

---

### Update Route

Update a route (creator only).

**Endpoint:** `PUT /routes/:id`

**Headers:** `Authorization: Bearer <token>`

**Request Body:**
```json
{
  "name": "Updated Route Name",
  "description": "New description",
  "max_participants": 15
}
```

**Response:** `200 OK`

**Errors:**
- `401` - Unauthorized
- `403` - Not the route creator
- `404` - Route not found
- `409` - Cannot update route that has started

---

### Delete Route

Delete a route (creator only).

**Endpoint:** `DELETE /routes/:id`

**Headers:** `Authorization: Bearer <token>`

**Response:** `204 No Content`

**Errors:**
- `401` - Unauthorized
- `403` - Not the route creator
- `404` - Route not found
- `409` - Cannot delete route in progress

---

## Participant Endpoints

### Join Route

Join a route as a participant.

**Endpoint:** `POST /routes/:id/join`

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "id": "770e8400-e29b-41d4-a716-446655440000",
    "route_id": "660e8400-e29b-41d4-a716-446655440000",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "active",
    "joined_at": "2024-01-15T11:00:00Z"
  }
}
```

**Errors:**
- `401` - Unauthorized
- `404` - Route not found
- `409` - Already joined or route is full
- `403` - Route has already started

---

### Leave Route

Leave a route.

**Endpoint:** `POST /routes/:id/leave`

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

**Errors:**
- `401` - Unauthorized
- `404` - Route not found or not a participant
- `403` - Route has already started

---

### Get Route Participants

Get list of participants for a route.

**Endpoint:** `GET /routes/:id/participants`

**Response:** `200 OK`
```json
{
  "success": true,
  "data": [
    {
      "id": "770e8400-e29b-41d4-a716-446655440000",
      "user": {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "display_name": "John Doe",
        "profile_pic": "https://example.com/pic.jpg"
      },
      "status": "active",
      "joined_at": "2024-01-15T11:00:00Z"
    }
  ]
}
```

---

### Remove Participant

Remove a participant from a route (creator only).

**Endpoint:** `DELETE /routes/:id/participants/:userId`

**Headers:** `Authorization: Bearer <token>`

**Response:** `204 No Content`

**Errors:**
- `401` - Unauthorized
- `403` - Not the route creator
- `404` - Route or participant not found

---

## Share Endpoints

### Get Share Link

Get a shareable link for a route.

**Endpoint:** `GET /routes/:id/share`

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "url": "https://convoy.app/routes/660e8400-e29b-41d4-a716-446655440000",
    "short_url": "https://convoy.app/r/abc123"
  }
}
```

---

## Health Check

### Health Check

Check if the API is running.

**Endpoint:** `GET /health`

**Response:** `200 OK`
```json
{
  "status": "ok",
  "service": "convoy"
}
```

---

## Rate Limiting

The API implements rate limiting:

- **General endpoints**: 100 requests per minute per IP
- **Auth endpoints**: 5 requests per minute per IP
- **Authenticated users**: 1000 requests per hour

When rate limit is exceeded, you'll receive:

**Response:** `429 Too Many Requests`
```json
{
  "success": false,
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many requests. Please try again later.",
    "details": {
      "retry_after": 60
    }
  }
}
```

---

## Webhooks (Future)

Future versions will support webhooks for:
- Route participant joined
- Route participant left
- Route updated
- Route started
- Route completed

---

## Changelog

### v1.0.0 (Planned)
- Initial API release
- User authentication
- Route CRUD operations
- Participant management
- Basic sharing

### v1.1.0 (Future)
- Real-time location tracking
- WebSocket support
- Push notifications
- Advanced search filters
