package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"convoy/internal/auth"
	"convoy/internal/models"
	"convoy/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestRequireAuth_ValidToken(t *testing.T) {
	mockRepo := new(MockUserRepository)
	tokenManager := auth.NewTokenManager("test-secret", 24*time.Hour)
	authService := service.NewAuthService(mockRepo, tokenManager)

	userID := uuid.New()
	user := &models.User{
		ID:          userID,
		Email:       "test@example.com",
		DisplayName: "Test User",
	}

	token, err := tokenManager.GenerateToken(userID)
	require.NoError(t, err)

	mockRepo.On("GetByID", userID).Return(user, nil)

	handler := RequireAuth(authService)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		extractedUserID := r.Context().Value(UserIDKey)
		assert.NotNil(t, extractedUserID)
		assert.Equal(t, userID, extractedUserID.(uuid.UUID))
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockRepo.AssertExpectations(t)
}

func TestRequireAuth_MissingAuthHeader(t *testing.T) {
	mockRepo := new(MockUserRepository)
	tokenManager := auth.NewTokenManager("test-secret", 24*time.Hour)
	authService := service.NewAuthService(mockRepo, tokenManager)

	handler := RequireAuth(authService)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "missing authorization header")
}

func TestRequireAuth_InvalidAuthHeaderFormat(t *testing.T) {
	mockRepo := new(MockUserRepository)
	tokenManager := auth.NewTokenManager("test-secret", 24*time.Hour)
	authService := service.NewAuthService(mockRepo, tokenManager)

	tests := []struct {
		name   string
		header string
	}{
		{
			name:   "missing Bearer prefix",
			header: "token123",
		},
		{
			name:   "wrong prefix",
			header: "Basic token123",
		},
		{
			name:   "empty token",
			header: "Bearer ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := RequireAuth(authService)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Fatal("handler should not be called")
			}))

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", tt.header)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})
	}
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	mockRepo := new(MockUserRepository)
	tokenManager := auth.NewTokenManager("test-secret", 24*time.Hour)
	authService := service.NewAuthService(mockRepo, tokenManager)

	handler := RequireAuth(authService)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid or expired token")
}

func TestRequireAuth_ExpiredToken(t *testing.T) {
	mockRepo := new(MockUserRepository)
	tokenManager := auth.NewTokenManager("test-secret", 1*time.Millisecond)
	authService := service.NewAuthService(mockRepo, tokenManager)

	userID := uuid.New()
	token, err := tokenManager.GenerateToken(userID)
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	handler := RequireAuth(authService)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRequireAuth_UserNotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	tokenManager := auth.NewTokenManager("test-secret", 24*time.Hour)
	authService := service.NewAuthService(mockRepo, tokenManager)

	userID := uuid.New()
	token, err := tokenManager.GenerateToken(userID)
	require.NoError(t, err)

	mockRepo.On("GetByID", userID).Return(nil, assert.AnError)

	handler := RequireAuth(authService)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	mockRepo.AssertExpectations(t)
}
