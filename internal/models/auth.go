package models

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token string        `json:"token"`
	User  *UserResponse `json:"user"`
}

type RegisterRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8"`
	DisplayName string `json:"display_name" validate:"required,min=2,max=50"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}
