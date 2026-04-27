package service

import (
	"testing"
	"time"

	"convoy/internal/auth"
	"convoy/internal/models"
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

func TestAuthService_Register(t *testing.T) {
	mockRepo := new(MockUserRepository)
	tokenManager := auth.NewTokenManager("test-secret", 24*time.Hour)
	service := NewAuthService(mockRepo, tokenManager)

	req := &models.RegisterRequest{
		Email:       "test@example.com",
		Password:    "password123",
		DisplayName: "Test User",
	}

	mockRepo.On("GetByEmail", req.Email).Return(nil, assert.AnError)
	mockRepo.On("Create", mock.AnythingOfType("*models.User")).Return(nil)

	user, token, err := service.Register(req)

	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotEmpty(t, token)
	assert.Equal(t, req.Email, user.Email)
	assert.Equal(t, req.DisplayName, user.DisplayName)
	assert.NotEqual(t, req.Password, user.PasswordHash)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Register_EmailAlreadyExists(t *testing.T) {
	mockRepo := new(MockUserRepository)
	tokenManager := auth.NewTokenManager("test-secret", 24*time.Hour)
	service := NewAuthService(mockRepo, tokenManager)

	existingUser := &models.User{
		ID:          uuid.New(),
		Email:       "test@example.com",
		DisplayName: "Existing User",
	}

	req := &models.RegisterRequest{
		Email:       "test@example.com",
		Password:    "password123",
		DisplayName: "Test User",
	}

	mockRepo.On("GetByEmail", req.Email).Return(existingUser, nil)

	user, token, err := service.Register(req)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "already registered")
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Register_CreateUserFails(t *testing.T) {
	mockRepo := new(MockUserRepository)
	tokenManager := auth.NewTokenManager("test-secret", 24*time.Hour)
	service := NewAuthService(mockRepo, tokenManager)

	req := &models.RegisterRequest{
		Email:       "test@example.com",
		Password:    "password123",
		DisplayName: "Test User",
	}

	mockRepo.On("GetByEmail", req.Email).Return(nil, assert.AnError)
	mockRepo.On("Create", mock.AnythingOfType("*models.User")).Return(assert.AnError)

	user, token, err := service.Register(req)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Empty(t, token)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Login(t *testing.T) {
	mockRepo := new(MockUserRepository)
	tokenManager := auth.NewTokenManager("test-secret", 24*time.Hour)
	service := NewAuthService(mockRepo, tokenManager)

	password := "password123"
	hashedPassword, err := auth.HashPassword(password)
	require.NoError(t, err)

	existingUser := &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		DisplayName:  "Test User",
	}

	req := &models.LoginRequest{
		Email:    "test@example.com",
		Password: password,
	}

	mockRepo.On("GetByEmail", req.Email).Return(existingUser, nil)

	user, token, err := service.Login(req)

	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotEmpty(t, token)
	assert.Equal(t, existingUser.ID, user.ID)
	assert.Equal(t, existingUser.Email, user.Email)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	tokenManager := auth.NewTokenManager("test-secret", 24*time.Hour)
	service := NewAuthService(mockRepo, tokenManager)

	req := &models.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	mockRepo.On("GetByEmail", req.Email).Return(nil, assert.AnError)

	user, token, err := service.Login(req)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "invalid email or password")
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	tokenManager := auth.NewTokenManager("test-secret", 24*time.Hour)
	service := NewAuthService(mockRepo, tokenManager)

	hashedPassword, err := auth.HashPassword("correctpassword")
	require.NoError(t, err)

	existingUser := &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		DisplayName:  "Test User",
	}

	req := &models.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	mockRepo.On("GetByEmail", req.Email).Return(existingUser, nil)

	user, token, err := service.Login(req)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "invalid email or password")
	mockRepo.AssertExpectations(t)
}

func TestAuthService_ValidateToken(t *testing.T) {
	mockRepo := new(MockUserRepository)
	tokenManager := auth.NewTokenManager("test-secret", 24*time.Hour)
	service := NewAuthService(mockRepo, tokenManager)

	userID := uuid.New()
	existingUser := &models.User{
		ID:          userID,
		Email:       "test@example.com",
		DisplayName: "Test User",
	}

	token, err := tokenManager.GenerateToken(userID)
	require.NoError(t, err)

	mockRepo.On("GetByID", userID).Return(existingUser, nil)

	validatedUserID, err := service.ValidateToken(token)

	require.NoError(t, err)
	assert.Equal(t, userID, validatedUserID)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_ValidateToken_InvalidToken(t *testing.T) {
	mockRepo := new(MockUserRepository)
	tokenManager := auth.NewTokenManager("test-secret", 24*time.Hour)
	service := NewAuthService(mockRepo, tokenManager)

	validatedUserID, err := service.ValidateToken("invalid-token")

	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, validatedUserID)
	assert.Contains(t, err.Error(), "invalid token")
}

func TestAuthService_ValidateToken_UserNotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	tokenManager := auth.NewTokenManager("test-secret", 24*time.Hour)
	service := NewAuthService(mockRepo, tokenManager)

	userID := uuid.New()
	token, err := tokenManager.GenerateToken(userID)
	require.NoError(t, err)

	mockRepo.On("GetByID", userID).Return(nil, assert.AnError)

	validatedUserID, err := service.ValidateToken(token)

	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, validatedUserID)
	assert.Contains(t, err.Error(), "user not found")
	mockRepo.AssertExpectations(t)
}
