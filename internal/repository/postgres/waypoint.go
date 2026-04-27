package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"convoy/internal/models"
)

// WaypointRepository handles database operations for waypoints
type WaypointRepository struct {
	db *sql.DB
}

// NewWaypointRepository creates a new waypoint repository
func NewWaypointRepository(db *sql.DB) *WaypointRepository {
	return &WaypointRepository{db: db}
}

// CreateBatch inserts multiple waypoints in a single transaction
func (r *WaypointRepository) CreateBatch(ctx context.Context, waypoints []models.Waypoint) error {
	if len(waypoints) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO waypoints (
			id, route_id, "order", latitude, longitude, name, address, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	for _, wp := range waypoints {
		_, err := stmt.ExecContext(
			ctx,
			wp.ID,
			wp.RouteID,
			wp.Order,
			wp.Latitude,
			wp.Longitude,
			wp.Name,
			wp.Address,
			now,
		)
		if err != nil {
			return fmt.Errorf("failed to insert waypoint: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetByRouteID retrieves all waypoints for a route, ordered by order
func (r *WaypointRepository) GetByRouteID(ctx context.Context, routeID string) ([]models.Waypoint, error) {
	query := `
		SELECT id, route_id, "order", latitude, longitude, name, address, created_at
		FROM waypoints
		WHERE route_id = $1
		ORDER BY "order" ASC
	`

	rows, err := r.db.QueryContext(ctx, query, routeID)
	if err != nil {
		return nil, fmt.Errorf("failed to query waypoints: %w", err)
	}
	defer rows.Close()

	waypoints := []models.Waypoint{}
	for rows.Next() {
		var wp models.Waypoint
		err := rows.Scan(
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

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating waypoints: %w", err)
	}

	return waypoints, nil
}

// DeleteByRouteID deletes all waypoints for a route
func (r *WaypointRepository) DeleteByRouteID(ctx context.Context, routeID string) error {
	query := `DELETE FROM waypoints WHERE route_id = $1`

	result, err := r.db.ExecContext(ctx, query, routeID)
	if err != nil {
		return fmt.Errorf("failed to delete waypoints: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	// It's okay if no rows were deleted (route had no waypoints)
	_ = rowsAffected

	return nil
}

// Update updates waypoints for a route (delete old, insert new)
func (r *WaypointRepository) Update(ctx context.Context, routeID string, waypoints []models.Waypoint) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing waypoints
	_, err = tx.ExecContext(ctx, `DELETE FROM waypoints WHERE route_id = $1`, routeID)
	if err != nil {
		return fmt.Errorf("failed to delete old waypoints: %w", err)
	}

	// Insert new waypoints
	if len(waypoints) > 0 {
		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO waypoints (
				id, route_id, "order", latitude, longitude, name, address, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`)
		if err != nil {
			return fmt.Errorf("failed to prepare statement: %w", err)
		}
		defer stmt.Close()

		now := time.Now()
		for _, wp := range waypoints {
			_, err := stmt.ExecContext(
				ctx,
				wp.ID,
				routeID,
				wp.Order,
				wp.Latitude,
				wp.Longitude,
				wp.Name,
				wp.Address,
				now,
			)
			if err != nil {
				return fmt.Errorf("failed to insert waypoint: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
