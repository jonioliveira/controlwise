package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError represents an application-specific error with HTTP status code
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
	Err        error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError
func New(code string, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// Wrap wraps an existing error with an AppError
func Wrap(err error, code string, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}

// Common errors
var (
	// Authentication errors
	ErrInvalidCredentials = New("INVALID_CREDENTIALS", "Invalid email or password", http.StatusUnauthorized)
	ErrUnauthorized       = New("UNAUTHORIZED", "Authentication required", http.StatusUnauthorized)
	ErrForbidden          = New("FORBIDDEN", "You don't have permission to perform this action", http.StatusForbidden)
	ErrTokenExpired       = New("TOKEN_EXPIRED", "Your session has expired, please login again", http.StatusUnauthorized)
	ErrTokenInvalid       = New("TOKEN_INVALID", "Invalid authentication token", http.StatusUnauthorized)
	ErrAccountInactive    = New("ACCOUNT_INACTIVE", "Your account is not active", http.StatusForbidden)

	// Validation errors
	ErrValidation     = New("VALIDATION_ERROR", "Invalid input data", http.StatusBadRequest)
	ErrInvalidRequest = New("INVALID_REQUEST", "Invalid request body", http.StatusBadRequest)
	ErrRequestTooLarge = New("REQUEST_TOO_LARGE", "Request body is too large", http.StatusRequestEntityTooLarge)

	// Resource errors
	ErrNotFound         = New("NOT_FOUND", "Resource not found", http.StatusNotFound)
	ErrClientNotFound   = New("CLIENT_NOT_FOUND", "Client not found", http.StatusNotFound)
	ErrWorksheetNotFound = New("WORKSHEET_NOT_FOUND", "Worksheet not found", http.StatusNotFound)
	ErrBudgetNotFound   = New("BUDGET_NOT_FOUND", "Budget not found", http.StatusNotFound)
	ErrProjectNotFound  = New("PROJECT_NOT_FOUND", "Project not found", http.StatusNotFound)
	ErrUserNotFound     = New("USER_NOT_FOUND", "User not found", http.StatusNotFound)

	// Conflict errors
	ErrEmailExists        = New("EMAIL_EXISTS", "Email already registered", http.StatusConflict)
	ErrClientHasWorksheets = New("CLIENT_HAS_WORKSHEETS", "Cannot delete client with existing worksheets", http.StatusConflict)
	ErrWorksheetHasBudgets = New("WORKSHEET_HAS_BUDGETS", "Cannot delete worksheet with existing budgets", http.StatusConflict)
	ErrBudgetApproved     = New("BUDGET_APPROVED", "Cannot modify approved budget", http.StatusConflict)
	ErrWorksheetApproved  = New("WORKSHEET_APPROVED", "Cannot modify approved worksheet", http.StatusConflict)

	// Internal errors
	ErrInternal = New("INTERNAL_ERROR", "An internal error occurred", http.StatusInternalServerError)
	ErrDatabase = New("DATABASE_ERROR", "A database error occurred", http.StatusInternalServerError)
)

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors contains multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

func (e *ValidationErrors) Error() string {
	return fmt.Sprintf("validation failed: %d error(s)", len(e.Errors))
}

// NewValidationErrors creates a new ValidationErrors
func NewValidationErrors(errors []ValidationError) *ValidationErrors {
	return &ValidationErrors{Errors: errors}
}

// Is checks if an error is of a specific type
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As attempts to convert an error to a specific type
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// GetStatusCode returns the HTTP status code for an error
func GetStatusCode(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.StatusCode
	}
	return http.StatusInternalServerError
}

// GetCode returns the error code for an error
func GetCode(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return "INTERNAL_ERROR"
}

// GetMessage returns a safe message for an error (doesn't expose internal details)
func GetMessage(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Message
	}
	return "An internal error occurred"
}
