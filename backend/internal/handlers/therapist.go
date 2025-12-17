package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/controlwise/backend/internal/middleware"
	"github.com/controlwise/backend/internal/models"
	"github.com/controlwise/backend/internal/services"
	"github.com/controlwise/backend/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type TherapistHandler struct {
	service *services.TherapistService
}

func NewTherapistHandler(service *services.TherapistService) *TherapistHandler {
	return &TherapistHandler{service: service}
}

type CreateTherapistRequest struct {
	UserID                 *string             `json:"user_id"`
	Name                   string              `json:"name"`
	Email                  *string             `json:"email"`
	Phone                  *string             `json:"phone"`
	Specialty              *string             `json:"specialty"`
	WorkingHours           models.WorkingHours `json:"working_hours"`
	SessionDurationMinutes int                 `json:"session_duration_minutes"`
	DefaultPriceCents      int                 `json:"default_price_cents"`
	Timezone               string              `json:"timezone"`
}

type UpdateTherapistRequest struct {
	UserID                 *string             `json:"user_id"`
	Name                   string              `json:"name"`
	Email                  *string             `json:"email"`
	Phone                  *string             `json:"phone"`
	Specialty              *string             `json:"specialty"`
	WorkingHours           models.WorkingHours `json:"working_hours"`
	SessionDurationMinutes int                 `json:"session_duration_minutes"`
	DefaultPriceCents      int                 `json:"default_price_cents"`
	Timezone               string              `json:"timezone"`
	IsActive               *bool               `json:"is_active"`
}

func (h *TherapistHandler) List(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	// Check if we should filter by active only
	activeOnly := r.URL.Query().Get("active") == "true"

	therapists, err := h.service.List(r.Context(), orgID, activeOnly)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(w, http.StatusOK, map[string]interface{}{
		"therapists": therapists,
		"total":      len(therapists),
	})
}

func (h *TherapistHandler) Get(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid therapist ID")
		return
	}

	therapist, err := h.service.GetByID(r.Context(), id, orgID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	utils.SuccessResponse(w, http.StatusOK, therapist)
}

func (h *TherapistHandler) Create(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	var req CreateTherapistRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Parse user ID if provided
	var userID *uuid.UUID
	if req.UserID != nil && *req.UserID != "" {
		parsed, err := uuid.Parse(*req.UserID)
		if err != nil {
			utils.ErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
			return
		}
		userID = &parsed
	}

	// Convert working hours to JSON
	var workingHoursJSON json.RawMessage
	if req.WorkingHours != nil {
		data, err := json.Marshal(req.WorkingHours)
		if err != nil {
			utils.ErrorResponse(w, http.StatusBadRequest, "Invalid working hours format")
			return
		}
		workingHoursJSON = data
	}

	therapist := &models.Therapist{
		OrganizationID:         orgID,
		UserID:                 userID,
		Name:                   req.Name,
		Email:                  req.Email,
		Phone:                  req.Phone,
		Specialty:              req.Specialty,
		WorkingHours:           workingHoursJSON,
		SessionDurationMinutes: req.SessionDurationMinutes,
		DefaultPriceCents:      req.DefaultPriceCents,
		Timezone:               req.Timezone,
	}

	if err := h.service.Create(r.Context(), therapist); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessMessageResponse(w, http.StatusCreated, "Therapist created successfully", therapist)
}

func (h *TherapistHandler) Update(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid therapist ID")
		return
	}

	var req UpdateTherapistRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Parse user ID if provided
	var userID *uuid.UUID
	if req.UserID != nil && *req.UserID != "" {
		parsed, err := uuid.Parse(*req.UserID)
		if err != nil {
			utils.ErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
			return
		}
		userID = &parsed
	}

	// Convert working hours to JSON
	var workingHoursJSON json.RawMessage
	if req.WorkingHours != nil {
		data, err := json.Marshal(req.WorkingHours)
		if err != nil {
			utils.ErrorResponse(w, http.StatusBadRequest, "Invalid working hours format")
			return
		}
		workingHoursJSON = data
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	therapist := &models.Therapist{
		UserID:                 userID,
		Name:                   req.Name,
		Email:                  req.Email,
		Phone:                  req.Phone,
		Specialty:              req.Specialty,
		WorkingHours:           workingHoursJSON,
		SessionDurationMinutes: req.SessionDurationMinutes,
		DefaultPriceCents:      req.DefaultPriceCents,
		Timezone:               req.Timezone,
		IsActive:               isActive,
	}

	if err := h.service.Update(r.Context(), id, orgID, therapist); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get updated therapist
	updatedTherapist, err := h.service.GetByID(r.Context(), id, orgID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to get updated therapist")
		return
	}

	utils.SuccessMessageResponse(w, http.StatusOK, "Therapist updated successfully", updatedTherapist)
}

func (h *TherapistHandler) Delete(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid therapist ID")
		return
	}

	if err := h.service.Delete(r.Context(), id, orgID); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessMessageResponse(w, http.StatusOK, "Therapist deleted successfully", nil)
}

func (h *TherapistHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	stats, err := h.service.GetStats(r.Context(), orgID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(w, http.StatusOK, stats)
}
