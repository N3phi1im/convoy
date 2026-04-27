package models

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type PaginatedResponse struct {
	Data       interface{}     `json:"data"`
	Pagination *PaginationMeta `json:"pagination"`
}

func NewSuccessResponse(data interface{}) *APIResponse {
	return &APIResponse{
		Success: true,
		Data:    data,
	}
}

func NewErrorResponse(code, message string, details interface{}) *APIResponse {
	return &APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

func NewPaginatedResponse(data interface{}, page, limit, total int) *APIResponse {
	totalPages := (total + limit - 1) / limit
	return &APIResponse{
		Success: true,
		Data: &PaginatedResponse{
			Data: data,
			Pagination: &PaginationMeta{
				Page:       page,
				Limit:      limit,
				Total:      total,
				TotalPages: totalPages,
			},
		},
	}
}
