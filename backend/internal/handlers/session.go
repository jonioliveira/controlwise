package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/controlewise/backend/internal/middleware"
	"github.com/controlewise/backend/internal/models"
	"github.com/controlewise/backend/internal/services"
	"github.com/controlewise/backend/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type SessionHandler struct {
	service *services.SessionService
}

func NewSessionHandler(service *services.SessionService) *SessionHandler {
	return &SessionHandler{service: service}
}

type CreateSessionRequest struct {
	TherapistID     string  `json:"therapist_id"`
	PatientID       string  `json:"patient_id"`
	ScheduledAt     string  `json:"scheduled_at"` // RFC3339 format
	DurationMinutes int     `json:"duration_minutes"`
	PriceCents      int     `json:"price_cents"`
	SessionType     string  `json:"session_type"`
	Notes           *string `json:"notes"`
}

type UpdateSessionRequest struct {
	TherapistID     string  `json:"therapist_id"`
	PatientID       string  `json:"patient_id"`
	ScheduledAt     string  `json:"scheduled_at"`
	DurationMinutes int     `json:"duration_minutes"`
	PriceCents      int     `json:"price_cents"`
	SessionType     string  `json:"session_type"`
	Notes           *string `json:"notes"`
}

type CancelSessionRequest struct {
	Reason string `json:"reason"`
}

func (h *SessionHandler) List(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	// Parse filters
	filters := services.SessionFilters{
		Limit:  20,
		Offset: 0,
	}

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			filters.Offset = (parsed - 1) * filters.Limit
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			filters.Limit = parsed
		}
	}

	if therapistID := r.URL.Query().Get("therapist_id"); therapistID != "" {
		if parsed, err := uuid.Parse(therapistID); err == nil {
			filters.TherapistID = &parsed
		}
	}

	if patientID := r.URL.Query().Get("patient_id"); patientID != "" {
		if parsed, err := uuid.Parse(patientID); err == nil {
			filters.PatientID = &parsed
		}
	}

	if status := r.URL.Query().Get("status"); status != "" {
		s := models.SessionStatus(status)
		filters.Status = &s
	}

	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		if parsed, err := time.Parse("2006-01-02", startDate); err == nil {
			filters.StartDate = &parsed
		}
	}

	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		if parsed, err := time.Parse("2006-01-02", endDate); err == nil {
			// Set to end of day
			endOfDay := parsed.Add(24*time.Hour - time.Second)
			filters.EndDate = &endOfDay
		}
	}

	sessions, total, err := h.service.List(r.Context(), orgID, filters)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	page := (filters.Offset / filters.Limit) + 1
	utils.SuccessResponse(w, http.StatusOK, map[string]interface{}{
		"sessions": sessions,
		"total":    total,
		"page":     page,
		"limit":    filters.Limit,
	})
}

func (h *SessionHandler) GetCalendar(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	// Parse date range (default to current month)
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 1, 0).Add(-time.Second)

	if startStr := r.URL.Query().Get("start"); startStr != "" {
		if parsed, err := time.Parse("2006-01-02", startStr); err == nil {
			start = parsed
		}
	}

	if endStr := r.URL.Query().Get("end"); endStr != "" {
		if parsed, err := time.Parse("2006-01-02", endStr); err == nil {
			end = parsed.Add(24*time.Hour - time.Second)
		}
	}

	// Optional therapist filter
	var therapistID *uuid.UUID
	if therapistStr := r.URL.Query().Get("therapist_id"); therapistStr != "" {
		if parsed, err := uuid.Parse(therapistStr); err == nil {
			therapistID = &parsed
		}
	}

	events, err := h.service.GetCalendarEvents(r.Context(), orgID, start, end, therapistID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(w, http.StatusOK, map[string]interface{}{
		"events": events,
		"start":  start.Format("2006-01-02"),
		"end":    end.Format("2006-01-02"),
	})
}

func (h *SessionHandler) Get(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	session, err := h.service.GetByID(r.Context(), id, orgID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	utils.SuccessResponse(w, http.StatusOK, session)
}

func (h *SessionHandler) Create(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "User not found")
		return
	}

	var req CreateSessionRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	therapistID, err := uuid.Parse(req.TherapistID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid therapist ID")
		return
	}

	patientID, err := uuid.Parse(req.PatientID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid patient ID")
		return
	}

	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid scheduled time format")
		return
	}

	session := &models.Session{
		OrganizationID:  orgID,
		TherapistID:     therapistID,
		PatientID:       patientID,
		ScheduledAt:     scheduledAt,
		DurationMinutes: req.DurationMinutes,
		PriceCents:      req.PriceCents,
		SessionType:     models.SessionType(req.SessionType),
		Notes:           req.Notes,
	}

	if err := h.service.Create(r.Context(), session, userID); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get created session with details
	createdSession, err := h.service.GetByID(r.Context(), session.ID, orgID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to get created session")
		return
	}

	utils.SuccessMessageResponse(w, http.StatusCreated, "Session created successfully", createdSession)
}

func (h *SessionHandler) Update(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "User not found")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	var req UpdateSessionRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	therapistID, err := uuid.Parse(req.TherapistID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid therapist ID")
		return
	}

	patientID, err := uuid.Parse(req.PatientID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid patient ID")
		return
	}

	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid scheduled time format")
		return
	}

	session := &models.Session{
		TherapistID:     therapistID,
		PatientID:       patientID,
		ScheduledAt:     scheduledAt,
		DurationMinutes: req.DurationMinutes,
		PriceCents:      req.PriceCents,
		SessionType:     models.SessionType(req.SessionType),
		Notes:           req.Notes,
	}

	if err := h.service.Update(r.Context(), id, orgID, session, userID); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get updated session
	updatedSession, err := h.service.GetByID(r.Context(), id, orgID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to get updated session")
		return
	}

	utils.SuccessMessageResponse(w, http.StatusOK, "Session updated successfully", updatedSession)
}

func (h *SessionHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "User not found")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	if err := h.service.Confirm(r.Context(), id, orgID, userID); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessMessageResponse(w, http.StatusOK, "Session confirmed successfully", nil)
}

func (h *SessionHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "User not found")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	var req CancelSessionRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.Cancel(r.Context(), id, orgID, req.Reason, userID); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessMessageResponse(w, http.StatusOK, "Session cancelled successfully", nil)
}

func (h *SessionHandler) Complete(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "User not found")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	if err := h.service.Complete(r.Context(), id, orgID, userID); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessMessageResponse(w, http.StatusOK, "Session completed successfully", nil)
}

func (h *SessionHandler) MarkNoShow(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "User not found")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	if err := h.service.MarkNoShow(r.Context(), id, orgID, userID); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessMessageResponse(w, http.StatusOK, "Session marked as no-show successfully", nil)
}

func (h *SessionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	if err := h.service.Delete(r.Context(), id, orgID); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessMessageResponse(w, http.StatusOK, "Session deleted successfully", nil)
}

func (h *SessionHandler) GetStats(w http.ResponseWriter, r *http.Request) {
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
