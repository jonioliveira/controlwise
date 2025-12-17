package handlers

import (
	"net/http"

	"github.com/controlwise/backend/internal/middleware"
	"github.com/controlwise/backend/internal/services"
	"github.com/controlwise/backend/internal/utils"
	"github.com/controlwise/backend/internal/validator"
)

type AdminAuthHandler struct {
	service *services.SystemAdminService
}

func NewAdminAuthHandler(service *services.SystemAdminService) *AdminAuthHandler {
	return &AdminAuthHandler{service: service}
}

func (h *AdminAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req validator.AdminLoginRequest
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
	serviceReq := services.AdminLoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	resp, err := h.service.Login(r.Context(), serviceReq)
	if err != nil {
		utils.ErrorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}

	utils.SuccessResponse(w, http.StatusOK, resp)
}

func (h *AdminAuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	adminID, ok := middleware.GetSystemAdminID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Admin not found in context")
		return
	}

	admin, err := h.service.GetByID(r.Context(), adminID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusNotFound, "Admin not found")
		return
	}

	utils.SuccessResponse(w, http.StatusOK, admin)
}

func (h *AdminAuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	adminID, ok := middleware.GetSystemAdminID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Admin not found in context")
		return
	}

	var req validator.AdminChangePasswordRequest
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
	serviceReq := services.AdminChangePasswordRequest{
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	}

	err := h.service.ChangePassword(r.Context(), adminID, serviceReq)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(w, http.StatusOK, map[string]string{
		"message": "Password changed successfully",
	})
}

func (h *AdminAuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// JWT-based auth doesn't require server-side logout
	// Client should just delete the token
	utils.SuccessResponse(w, http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
}
