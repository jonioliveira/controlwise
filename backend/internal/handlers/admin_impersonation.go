package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/controlwise/backend/internal/middleware"
	"github.com/controlwise/backend/internal/models"
	"github.com/controlwise/backend/internal/services"
	"github.com/controlwise/backend/internal/utils"
	"github.com/controlwise/backend/internal/validator"
)

type AdminImpersonationHandler struct {
	impersonationService *services.ImpersonationService
	auditService         *services.AdminAuditService
}

func NewAdminImpersonationHandler(impersonationService *services.ImpersonationService, auditService *services.AdminAuditService) *AdminImpersonationHandler {
	return &AdminImpersonationHandler{
		impersonationService: impersonationService,
		auditService:         auditService,
	}
}

func (h *AdminImpersonationHandler) Start(w http.ResponseWriter, r *http.Request) {
	adminID, ok := middleware.GetSystemAdminID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Admin not found in context")
		return
	}

	userIDStr := chi.URLParam(r, "userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req validator.AdminStartImpersonationRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.AppErrorResponse(w, err)
		return
	}

	if err := validator.Validate(req); err != nil {
		utils.AppErrorResponse(w, err)
		return
	}

	// Get client IP
	ipAddress := r.RemoteAddr

	token, err := h.impersonationService.StartImpersonation(r.Context(), adminID, userID, req.Reason, ipAddress)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Audit log
	h.auditService.Log(r.Context(), adminID, models.AuditActionUserImpersonated, models.AuditEntityUser, &userID,
		map[string]interface{}{
			"reason":     req.Reason,
			"session_id": token.SessionID.String(),
		},
		r.RemoteAddr, r.UserAgent())

	utils.SuccessResponse(w, http.StatusOK, token)
}

func (h *AdminImpersonationHandler) End(w http.ResponseWriter, r *http.Request) {
	// Check if this is an impersonation session
	sessionID, ok := middleware.GetImpersonationSessionID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusBadRequest, "No active impersonation session")
		return
	}

	impersonatorID, _ := middleware.GetImpersonatorID(r.Context())

	err := h.impersonationService.EndImpersonation(r.Context(), sessionID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to end impersonation: "+err.Error())
		return
	}

	// Audit log
	h.auditService.Log(r.Context(), impersonatorID, models.AuditActionImpersonationEnded, models.AuditEntityAdmin, &sessionID,
		nil,
		r.RemoteAddr, r.UserAgent())

	utils.SuccessResponse(w, http.StatusOK, map[string]string{
		"message": "Impersonation session ended",
	})
}

func (h *AdminImpersonationHandler) GetActiveSession(w http.ResponseWriter, r *http.Request) {
	adminID, ok := middleware.GetSystemAdminID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Admin not found in context")
		return
	}

	session, err := h.impersonationService.GetActiveSession(r.Context(), adminID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusNotFound, "No active impersonation session")
		return
	}

	utils.SuccessResponse(w, http.StatusOK, session)
}

func (h *AdminImpersonationHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	var adminID *uuid.UUID
	if adminIDStr := r.URL.Query().Get("admin_id"); adminIDStr != "" {
		if id, err := uuid.Parse(adminIDStr); err == nil {
			adminID = &id
		}
	}

	sessions, total, err := h.impersonationService.ListSessions(r.Context(), adminID, page, limit)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to list sessions")
		return
	}

	utils.PaginatedResponse(w, http.StatusOK, sessions, page, limit, total)
}
