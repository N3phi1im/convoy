package service

import (
	"context"
	"fmt"

	"convoy/internal/auth"
	"convoy/internal/models"
	"convoy/internal/repository"
	"github.com/google/uuid"
)

type AuthService struct {
	userRepo     repository.UserRepository
	tokenManager *auth.TokenManager
}

func NewAuthService(userRepo repository.UserRepository, tokenManager *auth.TokenManager) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		tokenManager: tokenManager,
	}
}

func (s *AuthService) Register(req *models.RegisterRequest) (*models.User, string, error) {
	existingUser, err := s.userRepo.GetByEmail(req.Email)
	if err == nil && existingUser != nil {
		return nil, "", fmt.Errorf("email already registered")
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: hashedPassword,
		DisplayName:  req.DisplayName,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, "", fmt.Errorf("failed to create user: %w", err)
	}

	token, err := s.tokenManager.GenerateToken(user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return user, token, nil
}

func (s *AuthService) Login(req *models.LoginRequest) (*models.User, string, error) {
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, "", fmt.Errorf("invalid email or password")
	}

	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		return nil, "", fmt.Errorf("invalid email or password")
	}

	token, err := s.tokenManager.GenerateToken(user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return user, token, nil
}

func (s *AuthService) ValidateToken(token string) (uuid.UUID, error) {
	claims, err := s.tokenManager.ValidateToken(token)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid token: %w", err)
	}

	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("user not found")
	}

	return user.ID, nil
}

func (s *AuthService) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}
