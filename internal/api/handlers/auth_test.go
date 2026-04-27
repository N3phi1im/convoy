package handlers

import (
	"bytes"
	"encoding/json"
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

func TestAuthHandler_Register_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	tokenManager := auth.NewTokenManager("test-secret", 24*time.Hour)
	authService := service.NewAuthService(mockRepo, tokenManager)
	handler := NewAuthHandler(authService)

	reqBody := models.RegisterRequest{
		Email:       "test@example.com",
		Password:    "password123",
		DisplayName: "Test User",
	}

	mockRepo.On("GetByEmail", reqBody.Email).Return(nil, assert.AnError)
	mockRepo.On("Create", mock.AnythingOfType("*models.User")).Return(nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var response models.APIResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)
	mockRepo.AssertExpectations(t)
}

func TestAuthHandler_Register_EmailExists(t *testing.T) {
	mockRepo := new(MockUserRepository)
	tokenManager := auth.NewTokenManager("test-secret", 24*time.Hour)
	authService := service.NewAuthService(mockRepo, tokenManager)
	handler := NewAuthHandler(authService)

	existingUser := &models.User{
		ID:          uuid.New(),
		Email:       "test@example.com",
		DisplayName: "Existing User",
	}

	reqBody := models.RegisterRequest{
		Email:       "test@example.com",
		Password:    "password123",
		DisplayName: "Test User",
	}

	mockRepo.On("GetByEmail", reqBody.Email).Return(existingUser, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)

	var response models.APIResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "EMAIL_EXISTS", response.Error.Code)
	mockRepo.AssertExpectations(t)
}

func TestAuthHandler_Register_ValidationError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	tokenManager := auth.NewTokenManager("test-secret", 24*time.Hour)
	authService := service.NewAuthService(mockRepo, tokenManager)
	handler := NewAuthHandler(authService)

	tests := []struct {
		name    string
		reqBody models.RegisterRequest
	}{
		{
			name: "missing email",
			reqBody: models.RegisterRequest{
				Password:    "password123",
				DisplayName: "Test User",
			},
		},
		{
			name: "invalid email",
			reqBody: models.RegisterRequest{
				Email:       "not-an-email",
				Password:    "password123",
				DisplayName: "Test User",
			},
		},
		{
			name: "short password",
			reqBody: models.RegisterRequest{
				Email:       "test@example.com",
				Password:    "short",
				DisplayName: "Test User",
			},
		},
		{
			name: "short display name",
			reqBody: models.RegisterRequest{
				Email:       "test@example.com",
				Password:    "password123",
				DisplayName: "A",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.Register(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var response models.APIResponse
			err := json.NewDecoder(rec.Body).Decode(&response)
			require.NoError(t, err)

			assert.False(t, response.Success)
			assert.NotNil(t, response.Error)
			assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
		})
	}
}

func TestAuthHandler_Register_InvalidJSON(t *testing.T) {
	mockRepo := new(MockUserRepository)
	tokenManager := auth.NewTokenManager("test-secret", 24*time.Hour)
	authService := service.NewAuthService(mockRepo, tokenManager)
	handler := NewAuthHandler(authService)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response models.APIResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, "INVALID_REQUEST", response.Error.Code)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	tokenManager := auth.NewTokenManager("test-secret", 24*time.Hour)
	authService := service.NewAuthService(mockRepo, tokenManager)
	handler := NewAuthHandler(authService)

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
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response models.APIResponse
	err = json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)
	mockRepo.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	mockRepo := new(MockUserRepository)
	tokenManager := auth.NewTokenManager("test-secret", 24*time.Hour)
	authService := service.NewAuthService(mockRepo, tokenManager)
	handler := NewAuthHandler(authService)

	reqBody := models.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	mockRepo.On("GetByEmail", reqBody.Email).Return(nil, assert.AnError)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var response models.APIResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, "INVALID_CREDENTIALS", response.Error.Code)
	mockRepo.AssertExpectations(t)
}

func TestAuthHandler_Login_WrongPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	tokenManager := auth.NewTokenManager("test-secret", 24*time.Hour)
	authService := service.NewAuthService(mockRepo, tokenManager)
	handler := NewAuthHandler(authService)

	hashedPassword, err := auth.HashPassword("correctpassword")
	require.NoError(t, err)

	existingUser := &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		DisplayName:  "Test User",
	}

	reqBody := models.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	mockRepo.On("GetByEmail", reqBody.Email).Return(existingUser, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var response models.APIResponse
	err = json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, "INVALID_CREDENTIALS", response.Error.Code)
	mockRepo.AssertExpectations(t)
}

func TestAuthHandler_Login_ValidationError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	tokenManager := auth.NewTokenManager("test-secret", 24*time.Hour)
	authService := service.NewAuthService(mockRepo, tokenManager)
	handler := NewAuthHandler(authService)

	tests := []struct {
		name    string
		reqBody models.LoginRequest
	}{
		{
			name: "missing email",
			reqBody: models.LoginRequest{
				Password: "password123",
			},
		},
		{
			name: "invalid email",
			reqBody: models.LoginRequest{
				Email:    "not-an-email",
				Password: "password123",
			},
		},
		{
			name: "missing password",
			reqBody: models.LoginRequest{
				Email: "test@example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.Login(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var response models.APIResponse
			err := json.NewDecoder(rec.Body).Decode(&response)
			require.NoError(t, err)

			assert.False(t, response.Success)
			assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
		})
	}
}
