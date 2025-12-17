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

type PatientHandler struct {
	service *services.PatientService
}

func NewPatientHandler(service *services.PatientService) *PatientHandler {
	return &PatientHandler{service: service}
}

// CreatePatientRequest - patient is linked to client, name/email/phone come from client
type CreatePatientRequest struct {
	ClientID         string  `json:"client_id"` // Required - link to client
	DateOfBirth      *string `json:"date_of_birth"` // Format: "2006-01-02"
	Notes            *string `json:"notes"` // Medical notes
	EmergencyContact *string `json:"emergency_contact"`
	EmergencyPhone   *string `json:"emergency_phone"`
}

// UpdatePatientRequest - only healthcare fields, client link cannot be changed
type UpdatePatientRequest struct {
	DateOfBirth      *string `json:"date_of_birth"`
	Notes            *string `json:"notes"`
	EmergencyContact *string `json:"emergency_contact"`
	EmergencyPhone   *string `json:"emergency_phone"`
	IsActive         *bool   `json:"is_active"`
}

func (h *PatientHandler) List(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	// Parse pagination parameters
	page := 1
	limit := 20
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	offset := (page - 1) * limit

	// Check if search query
	if search := r.URL.Query().Get("search"); search != "" {
		patients, err := h.service.Search(r.Context(), orgID, search, limit)
		if err != nil {
			utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		utils.SuccessResponse(w, http.StatusOK, map[string]interface{}{
			"patients": patients,
			"total":    len(patients),
		})
		return
	}

	// Normal list
	patients, total, err := h.service.List(r.Context(), orgID, limit, offset)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(w, http.StatusOK, map[string]interface{}{
		"patients": patients,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

func (h *PatientHandler) Get(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid patient ID")
		return
	}

	patient, err := h.service.GetByID(r.Context(), id, orgID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	utils.SuccessResponse(w, http.StatusOK, patient)
}

func (h *PatientHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	var req CreatePatientRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required client_id
	if req.ClientID == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "client_id is required")
		return
	}

	clientID, err := uuid.Parse(req.ClientID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid client_id format")
		return
	}

	// Parse date of birth if provided
	var dob *time.Time
	if req.DateOfBirth != nil && *req.DateOfBirth != "" {
		parsed, err := time.Parse("2006-01-02", *req.DateOfBirth)
		if err != nil {
			utils.ErrorResponse(w, http.StatusBadRequest, "Invalid date of birth format")
			return
		}
		dob = &parsed
	}

	patient := &models.Patient{
		OrganizationID:   orgID,
		ClientID:         clientID,
		DateOfBirth:      dob,
		Notes:            req.Notes,
		EmergencyContact: req.EmergencyContact,
		EmergencyPhone:   req.EmergencyPhone,
		CreatedBy:        &userID,
	}

	if err := h.service.Create(r.Context(), patient); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Fetch the created patient with client data
	createdPatient, err := h.service.GetByID(r.Context(), patient.ID, orgID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Patient created but failed to fetch details")
		return
	}

	utils.SuccessMessageResponse(w, http.StatusCreated, "Patient created successfully", createdPatient)
}

func (h *PatientHandler) Update(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid patient ID")
		return
	}

	var req UpdatePatientRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Parse date of birth if provided
	var dob *time.Time
	if req.DateOfBirth != nil && *req.DateOfBirth != "" {
		parsed, err := time.Parse("2006-01-02", *req.DateOfBirth)
		if err != nil {
			utils.ErrorResponse(w, http.StatusBadRequest, "Invalid date of birth format")
			return
		}
		dob = &parsed
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	// Only update healthcare fields - client link cannot be changed
	patient := &models.Patient{
		DateOfBirth:      dob,
		Notes:            req.Notes,
		EmergencyContact: req.EmergencyContact,
		EmergencyPhone:   req.EmergencyPhone,
		IsActive:         isActive,
	}

	if err := h.service.Update(r.Context(), id, orgID, patient); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get updated patient
	updatedPatient, err := h.service.GetByID(r.Context(), id, orgID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to get updated patient")
		return
	}

	utils.SuccessMessageResponse(w, http.StatusOK, "Patient updated successfully", updatedPatient)
}

func (h *PatientHandler) Delete(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid patient ID")
		return
	}

	if err := h.service.Delete(r.Context(), id, orgID); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessMessageResponse(w, http.StatusOK, "Patient deleted successfully", nil)
}

func (h *PatientHandler) GetStats(w http.ResponseWriter, r *http.Request) {
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
