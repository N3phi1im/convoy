package postgres

import (
	"database/sql"
	"testing"
	"time"

	"convoy/internal/models"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(&DB{DB: sqlxDB})

	user := &models.User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		DisplayName:  "Test User",
	}

	mock.ExpectExec("INSERT INTO users").
		WithArgs(
			sqlmock.AnyArg(), // id
			user.Email,
			user.PasswordHash,
			user.DisplayName,
			user.ProfilePic,
			sqlmock.AnyArg(), // created_at
			sqlmock.AnyArg(), // updated_at
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Create(user)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, user.ID)
	assert.False(t, user.CreatedAt.IsZero())
	assert.False(t, user.UpdatedAt.IsZero())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(&DB{DB: sqlxDB})

	userID := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "email", "password_hash", "display_name", "profile_pic", "created_at", "updated_at",
	}).AddRow(
		userID, "test@example.com", "hashedpassword", "Test User", nil, now, now,
	)

	mock.ExpectQuery("SELECT (.+) FROM users WHERE id = \\$1").
		WithArgs(userID).
		WillReturnRows(rows)

	retrieved, err := repo.GetByID(userID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, userID, retrieved.ID)
	assert.Equal(t, "test@example.com", retrieved.Email)
	assert.Equal(t, "Test User", retrieved.DisplayName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(&DB{DB: sqlxDB})

	userID := uuid.New()

	mock.ExpectQuery("SELECT (.+) FROM users WHERE id = \\$1").
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	_, err = repo.GetByID(userID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(&DB{DB: sqlxDB})

	userID := uuid.New()
	now := time.Now()
	email := "test@example.com"

	rows := sqlmock.NewRows([]string{
		"id", "email", "password_hash", "display_name", "profile_pic", "created_at", "updated_at",
	}).AddRow(
		userID, email, "hashedpassword", "Test User", nil, now, now,
	)

	mock.ExpectQuery("SELECT (.+) FROM users WHERE email = \\$1").
		WithArgs(email).
		WillReturnRows(rows)

	retrieved, err := repo.GetByEmail(email)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, userID, retrieved.ID)
	assert.Equal(t, email, retrieved.Email)
	assert.Equal(t, "Test User", retrieved.DisplayName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByEmail_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(&DB{DB: sqlxDB})

	email := "nonexistent@example.com"

	mock.ExpectQuery("SELECT (.+) FROM users WHERE email = \\$1").
		WithArgs(email).
		WillReturnError(sql.ErrNoRows)

	_, err = repo.GetByEmail(email)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(&DB{DB: sqlxDB})

	user := &models.User{
		ID:           uuid.New(),
		Email:        "updated@example.com",
		PasswordHash: "hashedpassword",
		DisplayName:  "Updated Name",
	}

	mock.ExpectExec("UPDATE users").
		WithArgs(
			user.Email,
			user.PasswordHash,
			user.DisplayName,
			user.ProfilePic,
			sqlmock.AnyArg(), // updated_at
			user.ID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Update(user)
	assert.NoError(t, err)
	assert.False(t, user.UpdatedAt.IsZero())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Update_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(&DB{DB: sqlxDB})

	user := &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		DisplayName:  "Test User",
	}

	mock.ExpectExec("UPDATE users").
		WithArgs(
			user.Email,
			user.PasswordHash,
			user.DisplayName,
			user.ProfilePic,
			sqlmock.AnyArg(),
			user.ID,
		).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.Update(user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(&DB{DB: sqlxDB})

	userID := uuid.New()

	mock.ExpectExec("DELETE FROM users WHERE id = \\$1").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Delete(userID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Delete_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(&DB{DB: sqlxDB})

	userID := uuid.New()

	mock.ExpectExec("DELETE FROM users WHERE id = \\$1").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.Delete(userID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}
