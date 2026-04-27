package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"convoy/internal/auth"
	"convoy/internal/config"
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
	if args.Get(0) != nil {
		return args.Error(0)
	}
	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	return nil
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

func TestRouter_HealthCheck(t *testing.T) {
	cfg := &config.Config{
		CORS: config.CORSConfig{
			AllowedOrigins: []string{"http://localhost:3000"},
		},
	}

	mockRepo := new(MockUserRepository)
	tokenManager := auth.NewTokenManager("test-secret", 24*time.Hour)
	authService := service.NewAuthService(mockRepo, tokenManager)

	router := NewRouter(cfg, authService, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "ok")
}

func TestRouter_Register(t *testing.T) {
	cfg := &config.Config{
		CORS: config.CORSConfig{
			AllowedOrigins: []string{"http://localhost:3000"},
		},
	}

	mockRepo := new(MockUserRepository)
	tokenManager := auth.NewTokenManager("test-secret", 24*time.Hour)
	authService := service.NewAuthService(mockRepo, tokenManager)

	router := NewRouter(cfg, authService, nil, nil)

	reqBody := models.RegisterRequest{
		Email:       "test@example.com",
		Password:    "password123",
		DisplayName: "Test User",
	}

	mockRepo.On("GetByEmail", reqBody.Email).Return(nil, assert.AnError)
	mockRepo.On("Create", mock.AnythingOfType("*models.User")).Return(nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	mockRepo.AssertExpectations(t)
}

func TestRouter_Login(t *testing.T) {
	cfg := &config.Config{
		CORS: config.CORSConfig{
			AllowedOrigins: []string{"http://localhost:3000"},
		},
	}

	mockRepo := new(MockUserRepository)
	tokenManager := auth.NewTokenManager("test-secret", 24*time.Hour)
	authService := service.NewAuthService(mockRepo, tokenManager)

	router := NewRouter(cfg, authService, nil, nil)

	password := "password123"
	hashedPassword, err := auth.HashPassword(password)
	require.NoError(t, err)

	existingUser := &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		DisplayName:  "Test User",
	}

	reqBody := models.LoginRequest{
		Email:    "test@example.com",
		Password: password,
	}

	mockRepo.On("GetByEmail", reqBody.Email).Return(existingUser, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockRepo.AssertExpectations(t)
}

func TestRouter_ProtectedRoute_WithValidToken(t *testing.T) {
	cfg := &config.Config{
		CORS: config.CORSConfig{
			AllowedOrigins: []string{"http://localhost:3000"},
		},
	}

	mockRepo := new(MockUserRepository)
	tokenManager := auth.NewTokenManager("test-secret", 24*time.Hour)
	authService := service.NewAuthService(mockRepo, tokenManager)

	router := NewRouter(cfg, authService, nil, nil)

	userID := uuid.New()
	user := &models.User{
		ID:          userID,
		Email:       "test@example.com",
		DisplayName: "Test User",
	}

	token, err := tokenManager.GenerateToken(userID)
	require.NoError(t, err)

	mockRepo.On("GetByID", userID).Return(user, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockRepo.AssertExpectations(t)
}

func TestRouter_ProtectedRoute_WithoutToken(t *testing.T) {
	cfg := &config.Config{
		CORS: config.CORSConfig{
			AllowedOrigins: []string{"http://localhost:3000"},
		},
	}

	mockRepo := new(MockUserRepository)
	tokenManager := auth.NewTokenManager("test-secret", 24*time.Hour)
	authService := service.NewAuthService(mockRepo, tokenManager)

	router := NewRouter(cfg, authService, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRouter_CORS(t *testing.T) {
	cfg := &config.Config{
		CORS: config.CORSConfig{
			AllowedOrigins: []string{"http://localhost:3000"},
		},
	}

	mockRepo := new(MockUserRepository)
	tokenManager := auth.NewTokenManager("test-secret", 24*time.Hour)
	authService := service.NewAuthService(mockRepo, tokenManager)

	router := NewRouter(cfg, authService, nil, nil)

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotEmpty(t, rec.Header().Get("Access-Control-Allow-Origin"))
}
