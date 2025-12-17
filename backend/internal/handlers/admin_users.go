package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/controlewise/backend/internal/middleware"
	"github.com/controlewise/backend/internal/models"
	"github.com/controlewise/backend/internal/services"
	"github.com/controlewise/backend/internal/utils"
	"github.com/controlewise/backend/internal/validator"
)

type AdminUsersHandler struct {
	userService  *services.AdminUserService
	auditService *services.AdminAuditService
}

func NewAdminUsersHandler(userService *services.AdminUserService, auditService *services.AdminAuditService) *AdminUsersHandler {
	return &AdminUsersHandler{
		userService:  userService,
		auditService: auditService,
	}
}

func (h *AdminUsersHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	search := r.URL.Query().Get("search")
	role := r.URL.Query().Get("role")

	var orgID *uuid.UUID
	if orgIDStr := r.URL.Query().Get("organization_id"); orgIDStr != "" {
		if id, err := uuid.Parse(orgIDStr); err == nil {
			orgID = &id
		}
	}

	var isActive *bool
	if activeStr := r.URL.Query().Get("is_active"); activeStr != "" {
		active := activeStr == "true"
		isActive = &active
	}

	params := services.ListUsersParams{
		Search:         search,
		OrganizationID: orgID,
		IsActive:       isActive,
		Role:           role,
		Page:           page,
		Limit:          limit,
	}

	users, total, err := h.userService.List(r.Context(), params)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to list users")
		return
	}

	utils.PaginatedResponse(w, http.StatusOK, users, page, limit, total)
}

func (h *AdminUsersHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := h.userService.GetByID(r.Context(), id)
	if err != nil {
		utils.ErrorResponse(w, http.StatusNotFound, "User not found")
		return
	}

	utils.SuccessResponse(w, http.StatusOK, user)
}

func (h *AdminUsersHandler) ListByOrganization(w http.ResponseWriter, r *http.Request) {
	orgIDStr := chi.URLParam(r, "id")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid organization ID")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	users, total, err := h.userService.ListByOrganization(r.Context(), orgID, page, limit)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to list users")
		return
	}

	utils.PaginatedResponse(w, http.StatusOK, users, page, limit, total)
}

func (h *AdminUsersHandler) Suspend(w http.ResponseWriter, r *http.Request) {
	adminID, ok := middleware.GetSystemAdminID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Admin not found in context")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req validator.AdminSuspendRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.AppErrorResponse(w, err)
		return
	}

	if err := validator.Validate(req); err != nil {
		utils.AppErrorResponse(w, err)
		return
	}

	err = h.userService.Suspend(r.Context(), id, adminID, req.Reason)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to suspend user: "+err.Error())
		return
	}

	// Audit log
	h.auditService.Log(r.Context(), adminID, models.AuditActionSuspend, models.AuditEntityUser, &id,
		map[string]interface{}{"reason": req.Reason},
		r.RemoteAddr, r.UserAgent())

	utils.SuccessResponse(w, http.StatusOK, map[string]string{
		"message": "User suspended successfully",
	})
}

func (h *AdminUsersHandler) Reactivate(w http.ResponseWriter, r *http.Request) {
	adminID, ok := middleware.GetSystemAdminID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Admin not found in context")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	err = h.userService.Reactivate(r.Context(), id)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to reactivate user: "+err.Error())
		return
	}

	// Audit log
	h.auditService.Log(r.Context(), adminID, models.AuditActionReactivate, models.AuditEntityUser, &id,
		nil,
		r.RemoteAddr, r.UserAgent())

	utils.SuccessResponse(w, http.StatusOK, map[string]string{
		"message": "User reactivated successfully",
	})
}

func (h *AdminUsersHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	adminID, ok := middleware.GetSystemAdminID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Admin not found in context")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req validator.AdminResetUserPasswordRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.AppErrorResponse(w, err)
		return
	}

	if err := validator.Validate(req); err != nil {
		utils.AppErrorResponse(w, err)
		return
	}

	err = h.userService.ResetPassword(r.Context(), id, req.NewPassword)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to reset password: "+err.Error())
		return
	}

	// Audit log
	h.auditService.Log(r.Context(), adminID, models.AuditActionUpdate, models.AuditEntityUser, &id,
		map[string]interface{}{"action": "reset_password"},
		r.RemoteAddr, r.UserAgent())

	utils.SuccessResponse(w, http.StatusOK, map[string]string{
		"message": "Password reset successfully",
	})
}
