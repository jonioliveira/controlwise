package handlers

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/controlewise/backend/internal/services"
	"github.com/controlewise/backend/internal/utils"
)

type AdminAuditHandler struct {
	auditService *services.AdminAuditService
}

func NewAdminAuditHandler(auditService *services.AdminAuditService) *AdminAuditHandler {
	return &AdminAuditHandler{
		auditService: auditService,
	}
}

func (h *AdminAuditHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	action := r.URL.Query().Get("action")
	entityType := r.URL.Query().Get("entity_type")

	var adminID *uuid.UUID
	if adminIDStr := r.URL.Query().Get("admin_id"); adminIDStr != "" {
		if id, err := uuid.Parse(adminIDStr); err == nil {
			adminID = &id
		}
	}

	var entityID *uuid.UUID
	if entityIDStr := r.URL.Query().Get("entity_id"); entityIDStr != "" {
		if id, err := uuid.Parse(entityIDStr); err == nil {
			entityID = &id
		}
	}

	params := services.AuditLogParams{
		AdminID:    adminID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Page:       page,
		Limit:      limit,
	}

	logs, total, err := h.auditService.List(r.Context(), params)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to list audit logs")
		return
	}

	utils.PaginatedResponse(w, http.StatusOK, logs, page, limit, total)
}
