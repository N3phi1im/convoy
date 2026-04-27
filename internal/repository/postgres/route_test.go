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

func TestRouteRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRouteRepository(db)
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		route := &models.Route{
			ID:              uuid.New(),
			CreatorID:       uuid.New(),
			Name:            "Test Route",
			RouteType:       models.RouteTypeCycling,
			Status:          models.RouteStatusPlanned,
			Visibility:      models.VisibilityPublic,
			MaxParticipants: intPtr(10),
			Distance:        5000.0,
			Duration:        3600,
		}

		now := time.Now()
		mock.ExpectQuery("INSERT INTO routes").
			WithArgs(
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
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
			).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
				AddRow(route.ID, now, now))

		err := repo.Create(ctx, route)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		route := &models.Route{
			ID:        uuid.New(),
			CreatorID: uuid.New(),
			Name:      "Test Route",
		}

		mock.ExpectQuery("INSERT INTO routes").
			WillReturnError(sql.ErrConnDone)

		err := repo.Create(ctx, route)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRouteRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRouteRepository(db)
	ctx := context.Background()

	t.Run("successful retrieval", func(t *testing.T) {
		id := uuid.New().String()
		creatorID := uuid.New()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "creator_id", "name", "description", "route_type", "status",
			"visibility", "start_time", "max_participants", "difficulty",
			"distance", "duration", "geometry", "created_at", "updated_at",
		}).AddRow(
			id, creatorID, "Test Route", nil, "cycling", "planned",
			"public", nil, 10, nil,
			5000.0, 3600, nil, now, now,
		)

		mock.ExpectQuery("SELECT (.+) FROM routes WHERE id = \\$1").
			WithArgs(id).
			WillReturnRows(rows)

		route, err := repo.GetByID(ctx, id)
		assert.NoError(t, err)
		assert.NotNil(t, route)
		assert.Equal(t, "Test Route", route.Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("route not found", func(t *testing.T) {
		id := uuid.New().String()

		mock.ExpectQuery("SELECT (.+) FROM routes WHERE id = \\$1").
			WithArgs(id).
			WillReturnError(sql.ErrNoRows)

		route, err := repo.GetByID(ctx, id)
		assert.Error(t, err)
		assert.Nil(t, route)
		assert.Contains(t, err.Error(), "not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRouteRepository_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRouteRepository(db)
	ctx := context.Background()

	t.Run("list with filters", func(t *testing.T) {
		filters := &models.RouteFilters{
			Status:     "planned",
			RouteType:  "cycling",
			Visibility: "public",
			Limit:      20,
			Offset:     0,
		}

		// Expect count query
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM routes").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		// Expect list query
		now := time.Now()
		rows := sqlmock.NewRows([]string{
			"id", "creator_id", "name", "description", "route_type", "status",
			"visibility", "start_time", "max_participants", "difficulty",
			"distance", "duration", "geometry", "created_at", "updated_at",
		}).
			AddRow(uuid.New().String(), uuid.New(), "Route 1", nil, "cycling", "planned", "public", nil, 10, nil, 5000.0, 3600, nil, now, now).
			AddRow(uuid.New().String(), uuid.New(), "Route 2", nil, "cycling", "planned", "public", nil, 10, nil, 6000.0, 4000, nil, now, now)

		mock.ExpectQuery("SELECT (.+) FROM routes").
			WillReturnRows(rows)

		routes, total, err := repo.List(ctx, filters)
		assert.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, routes, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty list", func(t *testing.T) {
		filters := &models.RouteFilters{
			Limit:  20,
			Offset: 0,
		}

		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM routes").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		mock.ExpectQuery("SELECT (.+) FROM routes").
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "creator_id", "name", "description", "route_type", "status",
				"visibility", "start_time", "max_participants", "difficulty",
				"distance", "duration", "created_at", "updated_at",
			}))

		routes, total, err := repo.List(ctx, filters)
		assert.NoError(t, err)
		assert.Equal(t, 0, total)
		assert.Len(t, routes, 0)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRouteRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRouteRepository(db)
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		route := &models.Route{
			ID:              uuid.New(),
			CreatorID:       uuid.New(),
			Name:            "Updated Route",
			RouteType:       models.RouteTypeCycling,
			Status:          models.RouteStatusPlanned,
			Visibility:      models.VisibilityPublic,
			MaxParticipants: intPtr(15),
			Distance:        6000.0,
			Duration:        4000,
		}

		now := time.Now()
		mock.ExpectQuery("UPDATE routes SET").
			WithArgs(
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
				sqlmock.AnyArg(),
			).
			WillReturnRows(sqlmock.NewRows([]string{"updated_at"}).AddRow(now))

		err := repo.Update(ctx, route)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("route not found", func(t *testing.T) {
		route := &models.Route{
			ID:   uuid.New(),
			Name: "Updated Route",
		}

		mock.ExpectQuery("UPDATE routes SET").
			WillReturnError(sql.ErrNoRows)

		err := repo.Update(ctx, route)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRouteRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRouteRepository(db)
	ctx := context.Background()

	t.Run("successful deletion", func(t *testing.T) {
		id := uuid.New().String()

		mock.ExpectExec("UPDATE routes SET deleted_at").
			WithArgs(id, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(ctx, id)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("route not found", func(t *testing.T) {
		id := uuid.New().String()

		mock.ExpectExec("UPDATE routes SET deleted_at").
			WithArgs(id, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Delete(ctx, id)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRouteRepository_GetWithDetails(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRouteRepository(db)
	ctx := context.Background()

	t.Run("successful retrieval with details", func(t *testing.T) {
		id := uuid.New().String()
		creatorID := uuid.New()
		now := time.Now()

		// Mock route query
		routeRows := sqlmock.NewRows([]string{
			"id", "creator_id", "name", "description", "route_type", "status",
			"visibility", "start_time", "max_participants", "difficulty",
			"distance", "duration", "geometry", "created_at", "updated_at",
		}).AddRow(
			id, creatorID, "Test Route", nil, "cycling", "planned",
			"public", nil, 10, nil,
			5000.0, 3600, nil, now, now,
		)

		mock.ExpectQuery("SELECT (.+) FROM routes WHERE id = \\$1").
			WithArgs(id).
			WillReturnRows(routeRows)

		// Mock waypoints query
		waypointRows := sqlmock.NewRows([]string{
			"id", "route_id", "order", "latitude", "longitude", "name", "address", "created_at",
		}).
			AddRow(uuid.New(), id, 0, 40.7128, -74.0060, nil, nil, now).
			AddRow(uuid.New(), id, 1, 34.0522, -118.2437, nil, nil, now)

		mock.ExpectQuery("SELECT (.+) FROM waypoints WHERE route_id").
			WithArgs(id).
			WillReturnRows(waypointRows)

		// Mock participants query
		participantRows := sqlmock.NewRows([]string{
			"id", "route_id", "user_id", "status", "joined_at", "email", "display_name",
		}).
			AddRow(uuid.New(), id, creatorID, "active", now, "user@example.com", "Test User")

		mock.ExpectQuery("SELECT (.+) FROM participants p JOIN users u").
			WithArgs(id).
			WillReturnRows(participantRows)

		details, err := repo.GetWithDetails(ctx, id)
		assert.NoError(t, err)
		assert.NotNil(t, details)
		assert.Equal(t, "Test Route", details.Name)
		assert.Len(t, details.Waypoints, 2)
		assert.Len(t, details.Participants, 1)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func intPtr(i int) *int {
	return &i
}
