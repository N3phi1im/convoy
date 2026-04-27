package handlers

import (
	"encoding/json"
	"net/http"

	"convoy/internal/models"
	"convoy/internal/service"
	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	authService *service.AuthService
	validator   *validator.Validate
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   validator.New(),
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body", nil)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		validationErrors := formatValidationErrors(err)
		respondWithError(w, http.StatusBadRequest, "VALIDATION_ERROR", "validation failed", validationErrors)
		return
	}

	user, token, err := h.authService.Register(&req)
	if err != nil {
		if err.Error() == "email already registered" {
			respondWithError(w, http.StatusConflict, "EMAIL_EXISTS", err.Error(), nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to register user", nil)
		return
	}

	response := &models.LoginResponse{
		Token: token,
		User:  user.ToResponse(),
	}

	respondWithSuccess(w, http.StatusCreated, response)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body", nil)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		validationErrors := formatValidationErrors(err)
		respondWithError(w, http.StatusBadRequest, "VALIDATION_ERROR", "validation failed", validationErrors)
		return
	}

	user, token, err := h.authService.Login(&req)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid email or password", nil)
		return
	}

	response := &models.LoginResponse{
		Token: token,
		User:  user.ToResponse(),
	}

	respondWithSuccess(w, http.StatusOK, response)
}

func respondWithSuccess(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := models.NewSuccessResponse(data)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func respondWithError(w http.ResponseWriter, status int, code, message string, details interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := models.NewErrorResponse(code, message, details)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func formatValidationErrors(err error) map[string]string {
	errors := make(map[string]string)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := e.Field()
			switch e.Tag() {
			case "required":
				errors[field] = "this field is required"
			case "email":
				errors[field] = "must be a valid email address"
			case "min":
				errors[field] = "must be at least " + e.Param() + " characters"
			case "max":
				errors[field] = "must be at most " + e.Param() + " characters"
			default:
				errors[field] = "invalid value"
			}
		}
	}

	return errors
}
