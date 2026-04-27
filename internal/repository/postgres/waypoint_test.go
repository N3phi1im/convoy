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

func TestWaypointRepository_CreateBatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewWaypointRepository(db)
	ctx := context.Background()

	t.Run("successful batch creation", func(t *testing.T) {
		routeID := uuid.New()
		waypoints := []models.Waypoint{
			{
				ID:        uuid.New(),
				RouteID:   routeID,
				Order:     0,
				Latitude:  40.7128,
				Longitude: -74.0060,
			},
			{
				ID:        uuid.New(),
				RouteID:   routeID,
				Order:     1,
				Latitude:  34.0522,
				Longitude: -118.2437,
			},
		}

		mock.ExpectBegin()
		mock.ExpectPrepare("INSERT INTO waypoints")
		for _, wp := range waypoints {
			mock.ExpectExec("INSERT INTO waypoints").
				WithArgs(wp.ID, wp.RouteID, wp.Order, wp.Latitude, wp.Longitude, wp.Name, wp.Address, sqlmock.AnyArg()).
				WillReturnResult(sqlmock.NewResult(1, 1))
		}
		mock.ExpectCommit()

		err := repo.CreateBatch(ctx, waypoints)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty waypoints", func(t *testing.T) {
		err := repo.CreateBatch(ctx, []models.Waypoint{})
		assert.NoError(t, err)
	})

	t.Run("transaction error", func(t *testing.T) {
		waypoints := []models.Waypoint{
			{
				ID:        uuid.New(),
				RouteID:   uuid.New(),
				Order:     0,
				Latitude:  40.7128,
				Longitude: -74.0060,
			},
		}

		mock.ExpectBegin().WillReturnError(sql.ErrConnDone)

		err := repo.CreateBatch(ctx, waypoints)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestWaypointRepository_GetByRouteID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewWaypointRepository(db)
	ctx := context.Background()

	t.Run("successful retrieval", func(t *testing.T) {
		routeID := uuid.New().String()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "route_id", "order", "latitude", "longitude", "name", "address", "created_at",
		}).
			AddRow(uuid.New(), routeID, 0, 40.7128, -74.0060, nil, nil, now).
			AddRow(uuid.New(), routeID, 1, 34.0522, -118.2437, nil, nil, now)

		mock.ExpectQuery("SELECT (.+) FROM waypoints WHERE route_id").
			WithArgs(routeID).
			WillReturnRows(rows)

		waypoints, err := repo.GetByRouteID(ctx, routeID)
		assert.NoError(t, err)
		assert.Len(t, waypoints, 2)
		assert.Equal(t, 0, waypoints[0].Order)
		assert.Equal(t, 1, waypoints[1].Order)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no waypoints found", func(t *testing.T) {
		routeID := uuid.New().String()

		mock.ExpectQuery("SELECT (.+) FROM waypoints WHERE route_id").
			WithArgs(routeID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "route_id", "order", "latitude", "longitude", "name", "address", "created_at",
			}))

		waypoints, err := repo.GetByRouteID(ctx, routeID)
		assert.NoError(t, err)
		assert.Len(t, waypoints, 0)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		routeID := uuid.New().String()

		mock.ExpectQuery("SELECT (.+) FROM waypoints WHERE route_id").
			WithArgs(routeID).
			WillReturnError(sql.ErrConnDone)

		waypoints, err := repo.GetByRouteID(ctx, routeID)
		assert.Error(t, err)
		assert.Nil(t, waypoints)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestWaypointRepository_DeleteByRouteID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewWaypointRepository(db)
	ctx := context.Background()

	t.Run("successful deletion", func(t *testing.T) {
		routeID := uuid.New().String()

		mock.ExpectExec("DELETE FROM waypoints WHERE route_id").
			WithArgs(routeID).
			WillReturnResult(sqlmock.NewResult(0, 2))

		err := repo.DeleteByRouteID(ctx, routeID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no waypoints to delete", func(t *testing.T) {
		routeID := uuid.New().String()

		mock.ExpectExec("DELETE FROM waypoints WHERE route_id").
			WithArgs(routeID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.DeleteByRouteID(ctx, routeID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		routeID := uuid.New().String()

		mock.ExpectExec("DELETE FROM waypoints WHERE route_id").
			WithArgs(routeID).
			WillReturnError(sql.ErrConnDone)

		err := repo.DeleteByRouteID(ctx, routeID)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestWaypointRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewWaypointRepository(db)
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		routeID := uuid.New().String()
		waypoints := []models.Waypoint{
			{
				ID:        uuid.New(),
				RouteID:   uuid.MustParse(routeID),
				Order:     0,
				Latitude:  40.7128,
				Longitude: -74.0060,
			},
		}

		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM waypoints WHERE route_id").
			WithArgs(routeID).
			WillReturnResult(sqlmock.NewResult(0, 2))
		mock.ExpectPrepare("INSERT INTO waypoints")
		mock.ExpectExec("INSERT INTO waypoints").
			WithArgs(waypoints[0].ID, routeID, waypoints[0].Order, waypoints[0].Latitude, waypoints[0].Longitude, waypoints[0].Name, waypoints[0].Address, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Update(ctx, routeID, waypoints)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update with empty waypoints", func(t *testing.T) {
		routeID := uuid.New().String()

		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM waypoints WHERE route_id").
			WithArgs(routeID).
			WillReturnResult(sqlmock.NewResult(0, 2))
		mock.ExpectCommit()

		err := repo.Update(ctx, routeID, []models.Waypoint{})
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("transaction error", func(t *testing.T) {
		routeID := uuid.New().String()
		waypoints := []models.Waypoint{
			{
				ID:        uuid.New(),
				RouteID:   uuid.MustParse(routeID),
				Order:     0,
				Latitude:  40.7128,
				Longitude: -74.0060,
			},
		}

		mock.ExpectBegin().WillReturnError(sql.ErrConnDone)

		err := repo.Update(ctx, routeID, waypoints)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
