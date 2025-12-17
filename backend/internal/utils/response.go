package utils

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	apperrors "github.com/controlewise/backend/internal/errors"
)

// Default max request body size (1MB)
const DefaultMaxBodySize int64 = 1 * 1024 * 1024

// MaxJSONBodySize is the maximum size for JSON request bodies (1MB)
const MaxJSONBodySize int64 = 1 * 1024 * 1024

type ErrorResponseBody struct {
	Error   string                      `json:"error"`
	Code    string                      `json:"code"`
	Message string                      `json:"message,omitempty"`
	Details []apperrors.ValidationError `json:"details,omitempty"`
}

type SuccessResponseBody struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// ErrorResponse sends an error response with a message
func ErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponseBody{
		Error:   http.StatusText(statusCode),
		Code:    http.StatusText(statusCode),
		Message: message,
	})
}

// AppErrorResponse sends an error response from an AppError
func AppErrorResponse(w http.ResponseWriter, err error) {
	var appErr *apperrors.AppError
	var validationErrs *apperrors.ValidationErrors

	if errors.As(err, &validationErrs) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponseBody{
			Error:   "Validation Error",
			Code:    "VALIDATION_ERROR",
			Message: "Invalid input data",
			Details: validationErrs.Errors,
		})
		return
	}

	if errors.As(err, &appErr) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(appErr.StatusCode)
		json.NewEncoder(w).Encode(ErrorResponseBody{
			Error:   http.StatusText(appErr.StatusCode),
			Code:    appErr.Code,
			Message: appErr.Message,
		})
		return
	}

	// Default to internal server error for unknown errors
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(ErrorResponseBody{
		Error:   "Internal Server Error",
		Code:    "INTERNAL_ERROR",
		Message: "An internal error occurred",
	})
}

// SuccessResponse sends a success response with data
func SuccessResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(SuccessResponseBody{
		Data: data,
	})
}

// SuccessMessageResponse sends a success response with a message and data
func SuccessMessageResponse(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(SuccessResponseBody{
		Data:    data,
		Message: message,
	})
}

// PaginatedResponseBody represents a paginated response
type PaginatedResponseBody struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// Pagination contains pagination metadata
type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// PaginatedResponse sends a paginated success response
func PaginatedResponse(w http.ResponseWriter, statusCode int, data interface{}, page, limit, total int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	totalPages := (total + limit - 1) / limit

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(PaginatedResponseBody{
		Data: data,
		Pagination: Pagination{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// ParseJSON decodes JSON from request body with size limits
func ParseJSON(r *http.Request, v interface{}) error {
	return ParseJSONWithLimit(r, v, MaxJSONBodySize)
}

// ParseJSONWithLimit decodes JSON from request body with a custom size limit
func ParseJSONWithLimit(r *http.Request, v interface{}, maxBytes int64) error {
	// Limit the size of the request body
	r.Body = http.MaxBytesReader(nil, r.Body, maxBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Strict parsing

	if err := decoder.Decode(v); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			return apperrors.ErrRequestTooLarge
		}

		var syntaxError *json.SyntaxError
		if errors.As(err, &syntaxError) {
			return apperrors.New("INVALID_JSON", "Invalid JSON syntax", http.StatusBadRequest)
		}

		var unmarshalTypeError *json.UnmarshalTypeError
		if errors.As(err, &unmarshalTypeError) {
			return apperrors.New("INVALID_TYPE", "Invalid type for field: "+unmarshalTypeError.Field, http.StatusBadRequest)
		}

		if errors.Is(err, io.EOF) {
			return apperrors.New("EMPTY_BODY", "Request body cannot be empty", http.StatusBadRequest)
		}

		return apperrors.ErrInvalidRequest
	}

	// Check for extra data after the JSON object
	if decoder.More() {
		return apperrors.New("EXTRA_DATA", "Request body contains extra data", http.StatusBadRequest)
	}

	return nil
}
