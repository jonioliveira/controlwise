package handlers

import (
	"net/http"

	"github.com/controlwise/backend/internal/middleware"
	"github.com/controlwise/backend/internal/models"
	"github.com/controlwise/backend/internal/services"
	"github.com/controlwise/backend/internal/utils"
	"github.com/go-chi/chi/v5"
)

type ModuleHandler struct {
	service *services.ModuleService
}

func NewModuleHandler(service *services.ModuleService) *ModuleHandler {
	return &ModuleHandler{service: service}
}

// ListAvailable returns all available modules in the system
func (h *ModuleHandler) ListAvailable(w http.ResponseWriter, r *http.Request) {
	modules, err := h.service.ListAvailable(r.Context())
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(w, http.StatusOK, map[string]interface{}{
		"modules": modules,
	})
}

// ListOrganizationModules returns all modules with their status for the current organization
func (h *ModuleHandler) ListOrganizationModules(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	modules, err := h.service.ListForOrganization(r.Context(), orgID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(w, http.StatusOK, map[string]interface{}{
		"modules": modules,
	})
}

// GetEnabledModules returns only the enabled module names for the current organization
func (h *ModuleHandler) GetEnabledModules(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	modules, err := h.service.GetEnabledModules(r.Context(), orgID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(w, http.StatusOK, map[string]interface{}{
		"enabled_modules": modules,
	})
}

type EnableModuleRequest struct {
	ModuleName string `json:"module_name"`
}

// EnableModule enables a module for the current organization
func (h *ModuleHandler) EnableModule(w http.ResponseWriter, r *http.Request) {
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

	// Check if user is admin
	role, ok := middleware.GetUserRole(r.Context())
	if !ok || role != string(models.RoleAdmin) {
		utils.ErrorResponse(w, http.StatusForbidden, "Only administrators can enable modules")
		return
	}

	moduleName := chi.URLParam(r, "module")
	if moduleName == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Module name is required")
		return
	}

	err := h.service.Enable(r.Context(), orgID, models.ModuleName(moduleName), userID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessMessageResponse(w, http.StatusOK, "Module enabled successfully", nil)
}

// DisableModule disables a module for the current organization
func (h *ModuleHandler) DisableModule(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	// Check if user is admin
	role, ok := middleware.GetUserRole(r.Context())
	if !ok || role != string(models.RoleAdmin) {
		utils.ErrorResponse(w, http.StatusForbidden, "Only administrators can disable modules")
		return
	}

	moduleName := chi.URLParam(r, "module")
	if moduleName == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Module name is required")
		return
	}

	err := h.service.Disable(r.Context(), orgID, models.ModuleName(moduleName))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessMessageResponse(w, http.StatusOK, "Module disabled successfully", nil)
}

type UpdateModuleConfigRequest struct {
	Config models.ModuleConfig `json:"config"`
}

// UpdateModuleConfig updates the configuration for a module
func (h *ModuleHandler) UpdateModuleConfig(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	// Check if user is admin
	role, ok := middleware.GetUserRole(r.Context())
	if !ok || role != string(models.RoleAdmin) {
		utils.ErrorResponse(w, http.StatusForbidden, "Only administrators can update module configuration")
		return
	}

	moduleName := chi.URLParam(r, "module")
	if moduleName == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Module name is required")
		return
	}

	var req UpdateModuleConfigRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := h.service.UpdateConfig(r.Context(), orgID, models.ModuleName(moduleName), req.Config)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessMessageResponse(w, http.StatusOK, "Module configuration updated successfully", nil)
}

// GetModuleConfig returns the configuration for a module
func (h *ModuleHandler) GetModuleConfig(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	moduleName := chi.URLParam(r, "module")
	if moduleName == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Module name is required")
		return
	}

	config, err := h.service.GetConfig(r.Context(), orgID, models.ModuleName(moduleName))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(w, http.StatusOK, map[string]interface{}{
		"module_name": moduleName,
		"config":      config,
	})
}
