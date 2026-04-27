package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"convoy/internal/models"
)

// RouteRepository handles database operations for routes
type RouteRepository struct {
	db *sql.DB
}

// NewRouteRepository creates a new route repository
func NewRouteRepository(db *sql.DB) *RouteRepository {
	return &RouteRepository{db: db}
}

// Create inserts a new route into the database
func (r *RouteRepository) Create(ctx context.Context, route *models.Route) error {
	query := `
		INSERT INTO routes (
			id, creator_id, name, description, route_type, status,
			visibility, start_time, max_participants, difficulty,
			distance, duration, geometry, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		route.ID,
		route.CreatorID,
		route.Name,
		route.Description,
		route.RouteType,
		route.Status,
		route.Visibility,
		route.StartTime,
		route.MaxParticipants,
		route.Difficulty,
		route.Distance,
		route.Duration,
		route.Geometry,
		time.Now(),
		time.Now(),
	).Scan(&route.ID, &route.CreatedAt, &route.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create route: %w", err)
	}

	return nil
}

// GetByID retrieves a route by its ID
func (r *RouteRepository) GetByID(ctx context.Context, id string) (*models.Route, error) {
	query := `
		SELECT 
			id, creator_id, name, description, route_type, status,
			visibility, start_time, max_participants, difficulty,
			distance, duration, geometry, created_at, updated_at
		FROM routes
		WHERE id = $1 AND deleted_at IS NULL
	`

	route := &models.Route{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&route.ID,
		&route.CreatorID,
		&route.Name,
		&route.Description,
		&route.RouteType,
		&route.Status,
		&route.Visibility,
		&route.StartTime,
		&route.MaxParticipants,
		&route.Difficulty,
		&route.Distance,
		&route.Duration,
		&route.Geometry,
		&route.CreatedAt,
		&route.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("route not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get route: %w", err)
	}

	return route, nil
}

// List retrieves routes based on filters with pagination
func (r *RouteRepository) List(ctx context.Context, filters *models.RouteFilters) ([]*models.Route, int, error) {
	// Build WHERE clause
	whereClause := "WHERE deleted_at IS NULL"
	args := []interface{}{}
	argCount := 1

	if filters.CreatorID != "" {
		whereClause += fmt.Sprintf(" AND creator_id = $%d", argCount)
		args = append(args, filters.CreatorID)
		argCount++
	}

	if filters.Status != "" {
		whereClause += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, filters.Status)
		argCount++
	}

	if filters.RouteType != "" {
		whereClause += fmt.Sprintf(" AND route_type = $%d", argCount)
		args = append(args, filters.RouteType)
		argCount++
	}

	if filters.Visibility != "" {
		whereClause += fmt.Sprintf(" AND visibility = $%d", argCount)
		args = append(args, filters.Visibility)
		argCount++
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM routes %s", whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count routes: %w", err)
	}

	// Get routes with pagination
	query := fmt.Sprintf(`
		SELECT 
			id, creator_id, name, description, route_type, status,
			visibility, start_time, max_participants, difficulty,
			distance, duration, geometry, created_at, updated_at
		FROM routes
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argCount, argCount+1)

	// Set defaults for pagination
	limit := filters.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := filters.Offset
	if offset < 0 {
		offset = 0
	}

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list routes: %w", err)
	}
	defer rows.Close()

	routes := []*models.Route{}
	for rows.Next() {
		route := &models.Route{}
		err := rows.Scan(
			&route.ID,
			&route.CreatorID,
			&route.Name,
			&route.Description,
			&route.RouteType,
			&route.Status,
			&route.Visibility,
			&route.StartTime,
			&route.MaxParticipants,
			&route.Difficulty,
			&route.Distance,
			&route.Duration,
			&route.Geometry,
			&route.CreatedAt,
			&route.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan route: %w", err)
		}
		routes = append(routes, route)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating routes: %w", err)
	}

	return routes, total, nil
}

// Update updates an existing route
func (r *RouteRepository) Update(ctx context.Context, route *models.Route) error {
	query := `
		UPDATE routes
		SET 
			name = $2,
			description = $3,
			route_type = $4,
			status = $5,
			visibility = $6,
			start_time = $7,
			max_participants = $8,
			difficulty = $9,
			distance = $10,
			duration = $11,
			updated_at = $12
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		route.ID,
		route.Name,
		route.Description,
		route.RouteType,
		route.Status,
		route.Visibility,
		route.StartTime,
		route.MaxParticipants,
		route.Difficulty,
		route.Distance,
		route.Duration,
		time.Now(),
	).Scan(&route.UpdatedAt)

	if err == sql.ErrNoRows {
		return fmt.Errorf("route not found")
	}
	if err != nil {
		return fmt.Errorf("failed to update route: %w", err)
	}

	return nil
}

// Delete soft deletes a route
func (r *RouteRepository) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE routes
		SET deleted_at = $2, updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete route: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("route not found")
	}

	return nil
}

// GetWithDetails retrieves a route with its waypoints and participants
func (r *RouteRepository) GetWithDetails(ctx context.Context, id string) (*models.RouteWithDetails, error) {
	// Get the route
	route, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	details := &models.RouteWithDetails{
		Route: *route,
	}

	// Get waypoints
	waypointsQuery := `
		SELECT id, route_id, "order", latitude, longitude, name, address, created_at
		FROM waypoints
		WHERE route_id = $1
		ORDER BY "order" ASC
	`

	waypointRows, err := r.db.QueryContext(ctx, waypointsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get waypoints: %w", err)
	}
	defer waypointRows.Close()

	waypoints := []models.Waypoint{}
	for waypointRows.Next() {
		var wp models.Waypoint
		err := waypointRows.Scan(
			&wp.ID,
			&wp.RouteID,
			&wp.Order,
			&wp.Latitude,
			&wp.Longitude,
			&wp.Name,
			&wp.Address,
			&wp.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan waypoint: %w", err)
		}
		waypoints = append(waypoints, wp)
	}
	details.Waypoints = waypoints

	// Get participants with user details
	participantsQuery := `
		SELECT 
			p.id, p.route_id, p.user_id, p.status, p.joined_at,
			u.email, u.display_name
		FROM participants p
		JOIN users u ON p.user_id = u.id
		WHERE p.route_id = $1
		ORDER BY p.joined_at ASC
	`

	participantRows, err := r.db.QueryContext(ctx, participantsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get participants: %w", err)
	}
	defer participantRows.Close()

	participants := []models.ParticipantWithUser{}
	for participantRows.Next() {
		var p models.ParticipantWithUser
		err := participantRows.Scan(
			&p.ID,
			&p.RouteID,
			&p.UserID,
			&p.Status,
			&p.JoinedAt,
			&p.User.Email,
			&p.User.DisplayName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan participant: %w", err)
		}
		p.User.ID = p.UserID
		participants = append(participants, p)
	}
	details.Participants = participants

	return details, nil
}
