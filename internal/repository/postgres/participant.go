package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"convoy/internal/models"
)

// ParticipantRepository handles database operations for participants
type ParticipantRepository struct {
	db *sql.DB
}

// NewParticipantRepository creates a new participant repository
func NewParticipantRepository(db *sql.DB) *ParticipantRepository {
	return &ParticipantRepository{db: db}
}

// Create inserts a new participant
func (r *ParticipantRepository) Create(ctx context.Context, participant *models.Participant) error {
	query := `
		INSERT INTO participants (id, route_id, user_id, status, joined_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, joined_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		participant.ID,
		participant.RouteID,
		participant.UserID,
		participant.Status,
		time.Now(),
	).Scan(&participant.ID, &participant.JoinedAt)

	if err != nil {
		return fmt.Errorf("failed to create participant: %w", err)
	}

	return nil
}

// GetByRouteAndUser retrieves a participant by route and user ID
func (r *ParticipantRepository) GetByRouteAndUser(ctx context.Context, routeID, userID string) (*models.Participant, error) {
	query := `
		SELECT id, route_id, user_id, status, joined_at
		FROM participants
		WHERE route_id = $1 AND user_id = $2
	`

	participant := &models.Participant{}
	err := r.db.QueryRowContext(ctx, query, routeID, userID).Scan(
		&participant.ID,
		&participant.RouteID,
		&participant.UserID,
		&participant.Status,
		&participant.JoinedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Not found, but not an error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get participant: %w", err)
	}

	return participant, nil
}

// ListByRoute retrieves all participants for a route
func (r *ParticipantRepository) ListByRoute(ctx context.Context, routeID string) ([]models.ParticipantWithUser, error) {
	query := `
		SELECT 
			p.id, p.route_id, p.user_id, p.status, p.joined_at,
			u.email, u.display_name
		FROM participants p
		JOIN users u ON p.user_id = u.id
		WHERE p.route_id = $1
		ORDER BY p.joined_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, routeID)
	if err != nil {
		return nil, fmt.Errorf("failed to query participants: %w", err)
	}
	defer rows.Close()

	participants := []models.ParticipantWithUser{}
	for rows.Next() {
		var p models.ParticipantWithUser
		err := rows.Scan(
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

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating participants: %w", err)
	}

	return participants, nil
}

// CountByRoute counts participants for a route
func (r *ParticipantRepository) CountByRoute(ctx context.Context, routeID string) (int, error) {
	query := `SELECT COUNT(*) FROM participants WHERE route_id = $1`

	var count int
	err := r.db.QueryRowContext(ctx, query, routeID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count participants: %w", err)
	}

	return count, nil
}

// UpdateStatus updates a participant's status
func (r *ParticipantRepository) UpdateStatus(ctx context.Context, id, status string) error {
	query := `
		UPDATE participants
		SET status = $2
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("failed to update participant status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("participant not found")
	}

	return nil
}

// Delete removes a participant
func (r *ParticipantRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM participants WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete participant: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("participant not found")
	}

	return nil
}

// DeleteByRouteAndUser removes a participant by route and user
func (r *ParticipantRepository) DeleteByRouteAndUser(ctx context.Context, routeID, userID string) error {
	query := `DELETE FROM participants WHERE route_id = $1 AND user_id = $2`

	result, err := r.db.ExecContext(ctx, query, routeID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete participant: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("participant not found")
	}

	return nil
}
