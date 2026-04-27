package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"convoy/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParticipantRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewParticipantRepository(db)
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		participant := &models.Participant{
			ID:      uuid.New(),
			RouteID: uuid.New(),
			UserID:  uuid.New(),
			Status:  models.ParticipantStatusActive,
		}

		now := time.Now()
		mock.ExpectQuery("INSERT INTO participants").
			WithArgs(
				participant.ID,
				participant.RouteID,
				participant.UserID,
				participant.Status,
				sqlmock.AnyArg(),
			).
			WillReturnRows(sqlmock.NewRows([]string{"id", "joined_at"}).
				AddRow(participant.ID, now))

		err := repo.Create(ctx, participant)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		participant := &models.Participant{
			ID:      uuid.New(),
			RouteID: uuid.New(),
			UserID:  uuid.New(),
			Status:  models.ParticipantStatusActive,
		}

		mock.ExpectQuery("INSERT INTO participants").
			WillReturnError(sql.ErrConnDone)

		err := repo.Create(ctx, participant)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestParticipantRepository_GetByRouteAndUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewParticipantRepository(db)
	ctx := context.Background()

	t.Run("successful retrieval", func(t *testing.T) {
		routeID := uuid.New().String()
		userID := uuid.New().String()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "route_id", "user_id", "status", "joined_at",
		}).AddRow(uuid.New(), routeID, userID, "active", now)

		mock.ExpectQuery("SELECT (.+) FROM participants WHERE route_id = \\$1 AND user_id = \\$2").
			WithArgs(routeID, userID).
			WillReturnRows(rows)

		participant, err := repo.GetByRouteAndUser(ctx, routeID, userID)
		assert.NoError(t, err)
		assert.NotNil(t, participant)
		assert.Equal(t, models.ParticipantStatusActive, participant.Status)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("participant not found", func(t *testing.T) {
		routeID := uuid.New().String()
		userID := uuid.New().String()

		mock.ExpectQuery("SELECT (.+) FROM participants WHERE route_id = \\$1 AND user_id = \\$2").
			WithArgs(routeID, userID).
			WillReturnError(sql.ErrNoRows)

		participant, err := repo.GetByRouteAndUser(ctx, routeID, userID)
		assert.NoError(t, err)
		assert.Nil(t, participant)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		routeID := uuid.New().String()
		userID := uuid.New().String()

		mock.ExpectQuery("SELECT (.+) FROM participants WHERE route_id = \\$1 AND user_id = \\$2").
			WithArgs(routeID, userID).
			WillReturnError(sql.ErrConnDone)

		participant, err := repo.GetByRouteAndUser(ctx, routeID, userID)
		assert.Error(t, err)
		assert.Nil(t, participant)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestParticipantRepository_ListByRoute(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewParticipantRepository(db)
	ctx := context.Background()

	t.Run("successful list", func(t *testing.T) {
		routeID := uuid.New().String()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "route_id", "user_id", "status", "joined_at", "email", "display_name",
		}).
			AddRow(uuid.New(), routeID, uuid.New(), "active", now, "user1@example.com", "User 1").
			AddRow(uuid.New(), routeID, uuid.New(), "active", now, "user2@example.com", "User 2")

		mock.ExpectQuery("SELECT (.+) FROM participants p JOIN users u").
			WithArgs(routeID).
			WillReturnRows(rows)

		participants, err := repo.ListByRoute(ctx, routeID)
		assert.NoError(t, err)
		assert.Len(t, participants, 2)
		assert.Equal(t, "user1@example.com", participants[0].User.Email)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty list", func(t *testing.T) {
		routeID := uuid.New().String()

		mock.ExpectQuery("SELECT (.+) FROM participants p JOIN users u").
			WithArgs(routeID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "route_id", "user_id", "status", "joined_at", "email", "display_name",
			}))

		participants, err := repo.ListByRoute(ctx, routeID)
		assert.NoError(t, err)
		assert.Len(t, participants, 0)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		routeID := uuid.New().String()

		mock.ExpectQuery("SELECT (.+) FROM participants p JOIN users u").
			WithArgs(routeID).
			WillReturnError(sql.ErrConnDone)

		participants, err := repo.ListByRoute(ctx, routeID)
		assert.Error(t, err)
		assert.Nil(t, participants)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestParticipantRepository_CountByRoute(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewParticipantRepository(db)
	ctx := context.Background()

	t.Run("successful count", func(t *testing.T) {
		routeID := uuid.New().String()

		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM participants WHERE route_id").
			WithArgs(routeID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

		count, err := repo.CountByRoute(ctx, routeID)
		assert.NoError(t, err)
		assert.Equal(t, 5, count)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("zero count", func(t *testing.T) {
		routeID := uuid.New().String()

		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM participants WHERE route_id").
			WithArgs(routeID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		count, err := repo.CountByRoute(ctx, routeID)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestParticipantRepository_UpdateStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewParticipantRepository(db)
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		id := uuid.New().String()
		status := string(models.ParticipantStatusLeft)

		mock.ExpectExec("UPDATE participants SET status").
			WithArgs(id, status).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdateStatus(ctx, id, status)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("participant not found", func(t *testing.T) {
		id := uuid.New().String()
		status := string(models.ParticipantStatusLeft)

		mock.ExpectExec("UPDATE participants SET status").
			WithArgs(id, status).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.UpdateStatus(ctx, id, status)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestParticipantRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewParticipantRepository(db)
	ctx := context.Background()

	t.Run("successful deletion", func(t *testing.T) {
		id := uuid.New().String()

		mock.ExpectExec("DELETE FROM participants WHERE id").
			WithArgs(id).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(ctx, id)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("participant not found", func(t *testing.T) {
		id := uuid.New().String()

		mock.ExpectExec("DELETE FROM participants WHERE id").
			WithArgs(id).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Delete(ctx, id)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestParticipantRepository_DeleteByRouteAndUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewParticipantRepository(db)
	ctx := context.Background()

	t.Run("successful deletion", func(t *testing.T) {
		routeID := uuid.New().String()
		userID := uuid.New().String()

		mock.ExpectExec("DELETE FROM participants WHERE route_id = \\$1 AND user_id = \\$2").
			WithArgs(routeID, userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.DeleteByRouteAndUser(ctx, routeID, userID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("participant not found", func(t *testing.T) {
		routeID := uuid.New().String()
		userID := uuid.New().String()

		mock.ExpectExec("DELETE FROM participants WHERE route_id = \\$1 AND user_id = \\$2").
			WithArgs(routeID, userID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.DeleteByRouteAndUser(ctx, routeID, userID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
