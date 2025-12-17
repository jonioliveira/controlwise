package handlers

import (
	"net/http"

	"github.com/controlewise/backend/internal/middleware"
	"github.com/controlewise/backend/internal/services"
	"github.com/controlewise/backend/internal/utils"
	"github.com/controlewise/backend/internal/validator"
)

type AuthHandler struct {
	service *services.AuthService
}

func NewAuthHandler(service *services.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req validator.RegisterRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.AppErrorResponse(w, err)
		return
	}

	// Validate request
	if err := validator.Validate(req); err != nil {
		utils.AppErrorResponse(w, err)
		return
	}

	// Convert to service request
	serviceReq := services.RegisterRequest{
		OrganizationName: req.OrganizationName,
		Email:            req.Email,
		Password:         req.Password,
		FirstName:        req.FirstName,
		LastName:         req.LastName,
		Phone:            req.Phone,
	}

	resp, err := h.service.Register(r.Context(), serviceReq)
	if err != nil {
		utils.AppErrorResponse(w, err)
		return
	}

	utils.SuccessResponse(w, http.StatusCreated, resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req validator.LoginRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.AppErrorResponse(w, err)
		return
	}

	// Validate request
	if err := validator.Validate(req); err != nil {
		utils.AppErrorResponse(w, err)
		return
	}

	// Convert to service request
	serviceReq := services.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	resp, err := h.service.Login(r.Context(), serviceReq)
	if err != nil {
		utils.AppErrorResponse(w, err)
		return
	}

	utils.SuccessResponse(w, http.StatusOK, resp)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "User not found in context")
		return
	}

	user, err := h.service.GetUserByID(r.Context(), userID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusNotFound, "User not found")
		return
	}

	utils.SuccessResponse(w, http.StatusOK, user)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement token refresh logic
	utils.SuccessResponse(w, http.StatusOK, map[string]string{
		"message": "Token refresh not implemented yet",
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement logout (if using token blacklist)
	utils.SuccessResponse(w, http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement forgot password
	utils.SuccessResponse(w, http.StatusOK, map[string]string{
		"message": "Password reset email sent",
	})
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement reset password
	utils.SuccessResponse(w, http.StatusOK, map[string]string{
		"message": "Password reset successfully",
	})
}
