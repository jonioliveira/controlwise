package validator

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"

	apperrors "github.com/controlewise/backend/internal/errors"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	// Register custom validations
	validate.RegisterValidation("password", validatePassword)
	validate.RegisterValidation("phone_pt", validatePhonePortugal)
}

// Validate validates a struct and returns validation errors
func Validate(s interface{}) error {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return err
	}

	var errors []apperrors.ValidationError
	for _, e := range validationErrors {
		errors = append(errors, apperrors.ValidationError{
			Field:   toSnakeCase(e.Field()),
			Message: getErrorMessage(e),
		})
	}

	return apperrors.NewValidationErrors(errors)
}

// validatePassword checks password complexity
func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if len(password) < 8 {
		return false
	}

	var hasUpper, hasLower, hasNumber bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		}
	}

	return hasUpper && hasLower && hasNumber
}

// validatePhonePortugal validates Portuguese phone numbers
func validatePhonePortugal(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	// Remove spaces and dashes
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")

	// Accept formats: +351XXXXXXXXX, 351XXXXXXXXX, 9XXXXXXXX, 2XXXXXXXX
	patterns := []string{
		`^\+351[0-9]{9}$`,
		`^351[0-9]{9}$`,
		`^[92][0-9]{8}$`,
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, phone); matched {
			return true
		}
	}

	return false
}

// getErrorMessage returns a user-friendly error message for a validation error
func getErrorMessage(e validator.FieldError) string {
	field := toSnakeCase(e.Field())

	switch e.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return "Invalid email format"
	case "min":
		return field + " must be at least " + e.Param() + " characters"
	case "max":
		return field + " must be at most " + e.Param() + " characters"
	case "password":
		return "Password must be at least 8 characters and contain uppercase, lowercase, and number"
	case "phone_pt":
		return "Invalid Portuguese phone number"
	case "uuid":
		return "Invalid ID format"
	case "oneof":
		return field + " must be one of: " + e.Param()
	case "gte":
		return field + " must be greater than or equal to " + e.Param()
	case "lte":
		return field + " must be less than or equal to " + e.Param()
	case "gt":
		return field + " must be greater than " + e.Param()
	case "lt":
		return field + " must be less than " + e.Param()
	default:
		return field + " is invalid"
	}
}

// toSnakeCase converts a PascalCase string to snake_case
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// Request validation structs with validation tags

type RegisterRequest struct {
	OrganizationName string `json:"organization_name" validate:"required,min=2,max=100"`
	Email            string `json:"email" validate:"required,email"`
	Password         string `json:"password" validate:"required,password"`
	FirstName        string `json:"first_name" validate:"required,min=2,max=50"`
	LastName         string `json:"last_name" validate:"required,min=2,max=50"`
	Phone            string `json:"phone" validate:"required,min=9,max=20"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=1"`
}

type CreateClientRequest struct {
	Name    string  `json:"name" validate:"required,min=2,max=100"`
	Email   string  `json:"email" validate:"required,email"`
	Phone   string  `json:"phone" validate:"required,min=9,max=20"`
	Address *string `json:"address" validate:"omitempty,max=500"`
	Notes   *string `json:"notes" validate:"omitempty,max=1000"`
}

type UpdateClientRequest struct {
	Name    *string `json:"name" validate:"omitempty,min=2,max=100"`
	Email   *string `json:"email" validate:"omitempty,email"`
	Phone   *string `json:"phone" validate:"omitempty,min=9,max=20"`
	Address *string `json:"address" validate:"omitempty,max=500"`
	Notes   *string `json:"notes" validate:"omitempty,max=1000"`
}

type CreateWorksheetRequest struct {
	ClientID    string `json:"client_id" validate:"required,uuid"`
	Title       string `json:"title" validate:"required,min=2,max=200"`
	Description string `json:"description" validate:"omitempty,max=5000"`
	Address     string `json:"address" validate:"omitempty,max=500"`
	Items       []WorksheetItemRequest `json:"items" validate:"dive"`
}

type WorksheetItemRequest struct {
	Description string  `json:"description" validate:"required,min=1,max=500"`
	Quantity    float64 `json:"quantity" validate:"required,gt=0"`
	Unit        string  `json:"unit" validate:"required,max=20"`
	Notes       string  `json:"notes" validate:"omitempty,max=500"`
}

type CreateBudgetRequest struct {
	WorksheetID string `json:"worksheet_id" validate:"required,uuid"`
	ValidUntil  string `json:"valid_until" validate:"required"`
	Notes       string `json:"notes" validate:"omitempty,max=2000"`
	Items       []BudgetItemRequest `json:"items" validate:"required,min=1,dive"`
}

type BudgetItemRequest struct {
	Description string  `json:"description" validate:"required,min=1,max=500"`
	Quantity    float64 `json:"quantity" validate:"required,gt=0"`
	Unit        string  `json:"unit" validate:"required,max=20"`
	UnitPrice   float64 `json:"unit_price" validate:"required,gte=0"`
}

type CreateTaskRequest struct {
	ProjectID   string  `json:"project_id" validate:"required,uuid"`
	Title       string  `json:"title" validate:"required,min=2,max=200"`
	Description string  `json:"description" validate:"omitempty,max=2000"`
	AssignedTo  *string `json:"assigned_to" validate:"omitempty,uuid"`
	DueDate     *string `json:"due_date"`
	Priority    string  `json:"priority" validate:"required,oneof=low medium high urgent"`
}

type CreatePaymentRequest struct {
	ProjectID     string  `json:"project_id" validate:"required,uuid"`
	Amount        float64 `json:"amount" validate:"required,gt=0"`
	DueDate       string  `json:"due_date" validate:"required"`
	Description   string  `json:"description" validate:"omitempty,max=500"`
	PaymentMethod string  `json:"payment_method" validate:"omitempty,max=50"`
}

// System Admin validation structs

type AdminLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=1"`
}

type AdminChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required,min=1"`
	NewPassword string `json:"new_password" validate:"required,password"`
}

type AdminCreateOrganizationRequest struct {
	Name      string  `json:"name" validate:"required,min=2,max=100"`
	Email     string  `json:"email" validate:"required,email"`
	Phone     string  `json:"phone" validate:"omitempty,min=9,max=20"`
	Address   string  `json:"address" validate:"omitempty,max=500"`
	TaxID     string  `json:"tax_id" validate:"omitempty,max=50"`
	// Admin user for the organization
	AdminEmail     string `json:"admin_email" validate:"required,email"`
	AdminPassword  string `json:"admin_password" validate:"required,password"`
	AdminFirstName string `json:"admin_first_name" validate:"required,min=2,max=50"`
	AdminLastName  string `json:"admin_last_name" validate:"required,min=2,max=50"`
	AdminPhone     string `json:"admin_phone" validate:"omitempty,min=9,max=20"`
}

type AdminUpdateOrganizationRequest struct {
	Name    *string `json:"name" validate:"omitempty,min=2,max=100"`
	Email   *string `json:"email" validate:"omitempty,email"`
	Phone   *string `json:"phone" validate:"omitempty,min=9,max=20"`
	Address *string `json:"address" validate:"omitempty,max=500"`
	TaxID   *string `json:"tax_id" validate:"omitempty,max=50"`
}

type AdminSuspendRequest struct {
	Reason string `json:"reason" validate:"required,min=5,max=500"`
}

type AdminStartImpersonationRequest struct {
	Reason string `json:"reason" validate:"required,min=5,max=500"`
}

type AdminResetUserPasswordRequest struct {
	NewPassword string `json:"new_password" validate:"required,password"`
}
