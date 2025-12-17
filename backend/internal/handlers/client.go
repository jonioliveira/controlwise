package handlers

import (
	"net/http"
	"strconv"

	"github.com/controlwise/backend/internal/middleware"
	"github.com/controlwise/backend/internal/models"
	"github.com/controlwise/backend/internal/services"
	"github.com/controlwise/backend/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ClientHandler struct {
	service *services.ClientService
}

func NewClientHandler(service *services.ClientService) *ClientHandler {
	return &ClientHandler{service: service}
}

type CreateClientRequest struct {
	Name    string  `json:"name"`
	Email   string  `json:"email"`
	Phone   string  `json:"phone"`
	Address *string `json:"address"`
	TaxID   *string `json:"tax_id"`
	Notes   *string `json:"notes"`
}

type UpdateClientRequest struct {
	Name    string  `json:"name"`
	Email   string  `json:"email"`
	Phone   string  `json:"phone"`
	Address *string `json:"address"`
	TaxID   *string `json:"tax_id"`
	Notes   *string `json:"notes"`
}

func (h *ClientHandler) List(w http.ResponseWriter, r *http.Request) {
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
		clients, err := h.service.Search(r.Context(), orgID, search, limit)
		if err != nil {
			utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		utils.SuccessResponse(w, http.StatusOK, map[string]interface{}{
			"clients": clients,
			"total":   len(clients),
		})
		return
	}

	// Normal list
	clients, total, err := h.service.List(r.Context(), orgID, limit, offset)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(w, http.StatusOK, map[string]interface{}{
		"clients": clients,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

func (h *ClientHandler) Get(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid client ID")
		return
	}

	client, err := h.service.GetByID(r.Context(), id, orgID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	utils.SuccessResponse(w, http.StatusOK, client)
}

func (h *ClientHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	var req CreateClientRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	client := &models.Client{
		OrganizationID: orgID,
		Name:           req.Name,
		Email:          req.Email,
		Phone:          req.Phone,
		Address:        req.Address,
		TaxID:          req.TaxID,
		Notes:          req.Notes,
		CreatedBy:      userID,
	}

	if err := h.service.Create(r.Context(), client); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessMessageResponse(w, http.StatusCreated, "Client created successfully", client)
}

func (h *ClientHandler) Update(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid client ID")
		return
	}

	var req UpdateClientRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	client := &models.Client{
		Name:    req.Name,
		Email:   req.Email,
		Phone:   req.Phone,
		Address: req.Address,
		TaxID:   req.TaxID,
		Notes:   req.Notes,
	}

	if err := h.service.Update(r.Context(), id, orgID, client); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get updated client
	updatedClient, err := h.service.GetByID(r.Context(), id, orgID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to get updated client")
		return
	}

	utils.SuccessMessageResponse(w, http.StatusOK, "Client updated successfully", updatedClient)
}

func (h *ClientHandler) Delete(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid client ID")
		return
	}

	if err := h.service.Delete(r.Context(), id, orgID); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessMessageResponse(w, http.StatusOK, "Client deleted successfully", nil)
}

func (h *ClientHandler) GetStats(w http.ResponseWriter, r *http.Request) {
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
