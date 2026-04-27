package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Email        string    `json:"email" db:"email" validate:"required,email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	DisplayName  string    `json:"display_name" db:"display_name" validate:"required,min=2,max=50"`
	ProfilePic   *string   `json:"profile_pic,omitempty" db:"profile_pic"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type UserCreateRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8"`
	DisplayName string `json:"display_name" validate:"required,min=2,max=50"`
}

type UserUpdateRequest struct {
	DisplayName *string `json:"display_name,omitempty" validate:"omitempty,min=2,max=50"`
	ProfilePic  *string `json:"profile_pic,omitempty"`
}

type UserResponse struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	ProfilePic  *string   `json:"profile_pic,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		ProfilePic:  u.ProfilePic,
		CreatedAt:   u.CreatedAt,
	}
}
