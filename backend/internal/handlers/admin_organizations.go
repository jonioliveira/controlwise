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

type AdminOrganizationsHandler struct {
	orgService    *services.AdminOrganizationService
	auditService  *services.AdminAuditService
	moduleService *services.ModuleService
}

func NewAdminOrganizationsHandler(orgService *services.AdminOrganizationService, auditService *services.AdminAuditService, moduleService *services.ModuleService) *AdminOrganizationsHandler {
	return &AdminOrganizationsHandler{
		orgService:    orgService,
		auditService:  auditService,
		moduleService: moduleService,
	}
}

func (h *AdminOrganizationsHandler) List(w http.ResponseWriter, r *http.Request) {
	// Parse query params
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	search := r.URL.Query().Get("search")

	var isActive *bool
	if activeStr := r.URL.Query().Get("is_active"); activeStr != "" {
		active := activeStr == "true"
		isActive = &active
	}

	params := services.ListOrganizationsParams{
		Search:   search,
		IsActive: isActive,
		Page:     page,
		Limit:    limit,
	}

	orgs, total, err := h.orgService.List(r.Context(), params)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to list organizations")
		return
	}

	utils.PaginatedResponse(w, http.StatusOK, orgs, page, limit, total)
}

func (h *AdminOrganizationsHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid organization ID")
		return
	}

	org, err := h.orgService.GetByID(r.Context(), id)
	if err != nil {
		utils.ErrorResponse(w, http.StatusNotFound, "Organization not found")
		return
	}

	utils.SuccessResponse(w, http.StatusOK, org)
}

func (h *AdminOrganizationsHandler) Create(w http.ResponseWriter, r *http.Request) {
	adminID, ok := middleware.GetSystemAdminID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Admin not found in context")
		return
	}

	var req validator.AdminCreateOrganizationRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.AppErrorResponse(w, err)
		return
	}

	if err := validator.Validate(req); err != nil {
		utils.AppErrorResponse(w, err)
		return
	}

	serviceReq := services.CreateOrganizationRequest{
		Name:           req.Name,
		Email:          req.Email,
		Phone:          req.Phone,
		Address:        req.Address,
		TaxID:          req.TaxID,
		AdminEmail:     req.AdminEmail,
		AdminPassword:  req.AdminPassword,
		AdminFirstName: req.AdminFirstName,
		AdminLastName:  req.AdminLastName,
		AdminPhone:     req.AdminPhone,
	}

	org, err := h.orgService.Create(r.Context(), adminID, serviceReq)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to create organization: "+err.Error())
		return
	}

	// Audit log
	h.auditService.Log(r.Context(), adminID, models.AuditActionCreate, models.AuditEntityOrganization, &org.ID,
		map[string]interface{}{"name": org.Name, "email": org.Email},
		r.RemoteAddr, r.UserAgent())

	utils.SuccessResponse(w, http.StatusCreated, org)
}

func (h *AdminOrganizationsHandler) Update(w http.ResponseWriter, r *http.Request) {
	adminID, ok := middleware.GetSystemAdminID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Admin not found in context")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid organization ID")
		return
	}

	var req validator.AdminUpdateOrganizationRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.AppErrorResponse(w, err)
		return
	}

	if err := validator.Validate(req); err != nil {
		utils.AppErrorResponse(w, err)
		return
	}

	serviceReq := services.UpdateOrganizationRequest{
		Name:    req.Name,
		Email:   req.Email,
		Phone:   req.Phone,
		Address: req.Address,
		TaxID:   req.TaxID,
	}

	org, err := h.orgService.Update(r.Context(), id, serviceReq)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to update organization: "+err.Error())
		return
	}

	// Audit log
	h.auditService.Log(r.Context(), adminID, models.AuditActionUpdate, models.AuditEntityOrganization, &id,
		map[string]interface{}{"changes": req},
		r.RemoteAddr, r.UserAgent())

	utils.SuccessResponse(w, http.StatusOK, org)
}

func (h *AdminOrganizationsHandler) Suspend(w http.ResponseWriter, r *http.Request) {
	adminID, ok := middleware.GetSystemAdminID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Admin not found in context")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid organization ID")
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

	err = h.orgService.Suspend(r.Context(), id, adminID, req.Reason)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to suspend organization: "+err.Error())
		return
	}

	// Audit log
	h.auditService.Log(r.Context(), adminID, models.AuditActionSuspend, models.AuditEntityOrganization, &id,
		map[string]interface{}{"reason": req.Reason},
		r.RemoteAddr, r.UserAgent())

	utils.SuccessResponse(w, http.StatusOK, map[string]string{
		"message": "Organization suspended successfully",
	})
}

func (h *AdminOrganizationsHandler) Reactivate(w http.ResponseWriter, r *http.Request) {
	adminID, ok := middleware.GetSystemAdminID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Admin not found in context")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid organization ID")
		return
	}

	err = h.orgService.Reactivate(r.Context(), id)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to reactivate organization: "+err.Error())
		return
	}

	// Audit log
	h.auditService.Log(r.Context(), adminID, models.AuditActionReactivate, models.AuditEntityOrganization, &id,
		nil,
		r.RemoteAddr, r.UserAgent())

	utils.SuccessResponse(w, http.StatusOK, map[string]string{
		"message": "Organization reactivated successfully",
	})
}

func (h *AdminOrganizationsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	adminID, ok := middleware.GetSystemAdminID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Admin not found in context")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid organization ID")
		return
	}

	err = h.orgService.Delete(r.Context(), id)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to delete organization: "+err.Error())
		return
	}

	// Audit log
	h.auditService.Log(r.Context(), adminID, models.AuditActionDelete, models.AuditEntityOrganization, &id,
		nil,
		r.RemoteAddr, r.UserAgent())

	utils.SuccessResponse(w, http.StatusOK, map[string]string{
		"message": "Organization deleted successfully",
	})
}

func (h *AdminOrganizationsHandler) ListModules(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid organization ID")
		return
	}

	modules, err := h.moduleService.ListForOrganization(r.Context(), id)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to list modules: "+err.Error())
		return
	}

	utils.SuccessResponse(w, http.StatusOK, modules)
}

func (h *AdminOrganizationsHandler) EnableModule(w http.ResponseWriter, r *http.Request) {
	adminID, ok := middleware.GetSystemAdminID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Admin not found in context")
		return
	}

	idStr := chi.URLParam(r, "id")
	orgID, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid organization ID")
		return
	}

	moduleName := chi.URLParam(r, "module")
	if moduleName == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Module name is required")
		return
	}

	err = h.moduleService.EnableByAdmin(r.Context(), orgID, models.ModuleName(moduleName), adminID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "failed to enable module: "+err.Error())
		return
	}

	// Audit log
	h.auditService.Log(r.Context(), adminID, models.AuditActionUpdate, models.AuditEntityOrganization, &orgID,
		map[string]interface{}{"action": "enable_module", "module": moduleName},
		r.RemoteAddr, r.UserAgent())

	utils.SuccessResponse(w, http.StatusOK, map[string]string{
		"message": "Module enabled successfully",
	})
}

func (h *AdminOrganizationsHandler) DisableModule(w http.ResponseWriter, r *http.Request) {
	adminID, ok := middleware.GetSystemAdminID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Admin not found in context")
		return
	}

	idStr := chi.URLParam(r, "id")
	orgID, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid organization ID")
		return
	}

	moduleName := chi.URLParam(r, "module")
	if moduleName == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Module name is required")
		return
	}

	err = h.moduleService.Disable(r.Context(), orgID, models.ModuleName(moduleName))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Audit log
	h.auditService.Log(r.Context(), adminID, models.AuditActionUpdate, models.AuditEntityOrganization, &orgID,
		map[string]interface{}{"action": "disable_module", "module": moduleName},
		r.RemoteAddr, r.UserAgent())

	utils.SuccessResponse(w, http.StatusOK, map[string]string{
		"message": "Module disabled successfully",
	})
}
