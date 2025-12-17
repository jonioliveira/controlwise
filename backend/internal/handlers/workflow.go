package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/controlwise/backend/internal/middleware"
	"github.com/controlwise/backend/internal/models"
	"github.com/controlwise/backend/internal/services"
	"github.com/controlwise/backend/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type WorkflowHandler struct {
	service *services.WorkflowService
}

func NewWorkflowHandler(service *services.WorkflowService) *WorkflowHandler {
	return &WorkflowHandler{service: service}
}

// ============ Workflow Handlers ============

type CreateWorkflowRequest struct {
	Name        string  `json:"name" validate:"required,min=2,max=100"`
	Description *string `json:"description"`
	Module      string  `json:"module" validate:"required,oneof=appointments construction"`
	EntityType  string  `json:"entity_type" validate:"required,oneof=session budget project"`
	IsDefault   bool    `json:"is_default"`
}

type UpdateWorkflowRequest struct {
	Name        string  `json:"name" validate:"required,min=2,max=100"`
	Description *string `json:"description"`
	IsActive    bool    `json:"is_active"`
	IsDefault   bool    `json:"is_default"`
}

func (h *WorkflowHandler) ListWorkflows(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	module := r.URL.Query().Get("module")

	workflows, err := h.service.ListWorkflows(r.Context(), orgID, module)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(w, http.StatusOK, map[string]interface{}{
		"workflows": workflows,
	})
}

func (h *WorkflowHandler) GetWorkflow(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid workflow ID")
		return
	}

	workflow, err := h.service.GetWorkflowByID(r.Context(), id, orgID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	utils.SuccessResponse(w, http.StatusOK, workflow)
}

func (h *WorkflowHandler) CreateWorkflow(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	var req CreateWorkflowRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	workflow := &models.Workflow{
		OrganizationID: orgID,
		Name:           req.Name,
		Description:    req.Description,
		Module:         models.WorkflowModule(req.Module),
		EntityType:     models.WorkflowEntityType(req.EntityType),
		IsDefault:      req.IsDefault,
	}

	if err := h.service.CreateWorkflow(r.Context(), workflow); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessMessageResponse(w, http.StatusCreated, "Workflow created successfully", workflow)
}

func (h *WorkflowHandler) UpdateWorkflow(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid workflow ID")
		return
	}

	var req UpdateWorkflowRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	workflow := &models.Workflow{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    req.IsActive,
		IsDefault:   req.IsDefault,
	}

	if err := h.service.UpdateWorkflow(r.Context(), id, orgID, workflow); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	updated, _ := h.service.GetWorkflowByID(r.Context(), id, orgID)
	utils.SuccessMessageResponse(w, http.StatusOK, "Workflow updated successfully", updated)
}

func (h *WorkflowHandler) DeleteWorkflow(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid workflow ID")
		return
	}

	if err := h.service.DeleteWorkflow(r.Context(), id, orgID); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessMessageResponse(w, http.StatusOK, "Workflow deleted successfully", nil)
}

func (h *WorkflowHandler) DuplicateWorkflow(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid workflow ID")
		return
	}

	var req struct {
		Name string `json:"name" validate:"required"`
	}
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	duplicated, err := h.service.DuplicateWorkflow(r.Context(), id, orgID, req.Name)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessMessageResponse(w, http.StatusCreated, "Workflow duplicated successfully", duplicated)
}

// ============ State Handlers ============

type CreateStateRequest struct {
	Name        string  `json:"name" validate:"required,min=2,max=50"`
	DisplayName string  `json:"display_name" validate:"required,min=2,max=100"`
	Description *string `json:"description"`
	StateType   string  `json:"state_type" validate:"required,oneof=initial intermediate final"`
	Color       *string `json:"color"`
	Position    int     `json:"position"`
}

func (h *WorkflowHandler) CreateState(w http.ResponseWriter, r *http.Request) {
	workflowID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid workflow ID")
		return
	}

	var req CreateStateRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	state := &models.WorkflowState{
		WorkflowID:  workflowID,
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		StateType:   models.StateType(req.StateType),
		Color:       req.Color,
		Position:    req.Position,
	}

	if err := h.service.CreateState(r.Context(), state); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessMessageResponse(w, http.StatusCreated, "State created successfully", state)
}

func (h *WorkflowHandler) UpdateState(w http.ResponseWriter, r *http.Request) {
	stateID, err := uuid.Parse(chi.URLParam(r, "stateId"))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid state ID")
		return
	}

	var req CreateStateRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	state := &models.WorkflowState{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		StateType:   models.StateType(req.StateType),
		Color:       req.Color,
		Position:    req.Position,
	}

	if err := h.service.UpdateState(r.Context(), stateID, state); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessMessageResponse(w, http.StatusOK, "State updated successfully", state)
}

func (h *WorkflowHandler) DeleteState(w http.ResponseWriter, r *http.Request) {
	stateID, err := uuid.Parse(chi.URLParam(r, "stateId"))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid state ID")
		return
	}

	if err := h.service.DeleteState(r.Context(), stateID); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessMessageResponse(w, http.StatusOK, "State deleted successfully", nil)
}

func (h *WorkflowHandler) ReorderStates(w http.ResponseWriter, r *http.Request) {
	workflowID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid workflow ID")
		return
	}

	var req struct {
		StateIDs []string `json:"state_ids" validate:"required"`
	}
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	stateIDs := make([]uuid.UUID, len(req.StateIDs))
	for i, idStr := range req.StateIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			utils.ErrorResponse(w, http.StatusBadRequest, "Invalid state ID in list")
			return
		}
		stateIDs[i] = id
	}

	if err := h.service.ReorderStates(r.Context(), workflowID, stateIDs); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessMessageResponse(w, http.StatusOK, "States reordered successfully", nil)
}

// ============ Trigger Handlers ============

type CreateTriggerRequest struct {
	StateID           *string `json:"state_id"`
	TransitionID      *string `json:"transition_id"`
	TriggerType       string  `json:"trigger_type" validate:"required,oneof=on_enter on_exit time_before time_after recurring"`
	TimeOffsetMinutes *int    `json:"time_offset_minutes"`
	TimeField         *string `json:"time_field"`
	RecurringCron     *string `json:"recurring_cron"`
	Conditions        *json.RawMessage `json:"conditions"`
}

func (h *WorkflowHandler) CreateTrigger(w http.ResponseWriter, r *http.Request) {
	workflowID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid workflow ID")
		return
	}

	var req CreateTriggerRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	trigger := &models.WorkflowTrigger{
		WorkflowID:        workflowID,
		TriggerType:       models.TriggerType(req.TriggerType),
		TimeOffsetMinutes: req.TimeOffsetMinutes,
		TimeField:         req.TimeField,
		RecurringCron:     req.RecurringCron,
	}

	if req.StateID != nil {
		id, err := uuid.Parse(*req.StateID)
		if err != nil {
			utils.ErrorResponse(w, http.StatusBadRequest, "Invalid state ID")
			return
		}
		trigger.StateID = &id
	}

	if req.TransitionID != nil {
		id, err := uuid.Parse(*req.TransitionID)
		if err != nil {
			utils.ErrorResponse(w, http.StatusBadRequest, "Invalid transition ID")
			return
		}
		trigger.TransitionID = &id
	}

	if req.Conditions != nil {
		trigger.Conditions = *req.Conditions
	}

	if err := h.service.CreateTrigger(r.Context(), trigger); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessMessageResponse(w, http.StatusCreated, "Trigger created successfully", trigger)
}

func (h *WorkflowHandler) UpdateTrigger(w http.ResponseWriter, r *http.Request) {
	triggerID, err := uuid.Parse(chi.URLParam(r, "triggerId"))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid trigger ID")
		return
	}

	var req struct {
		StateID           *string          `json:"state_id"`
		TransitionID      *string          `json:"transition_id"`
		TriggerType       string           `json:"trigger_type"`
		TimeOffsetMinutes *int             `json:"time_offset_minutes"`
		TimeField         *string          `json:"time_field"`
		RecurringCron     *string          `json:"recurring_cron"`
		Conditions        *json.RawMessage `json:"conditions"`
		IsActive          bool             `json:"is_active"`
	}
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	trigger := &models.WorkflowTrigger{
		TriggerType:       models.TriggerType(req.TriggerType),
		TimeOffsetMinutes: req.TimeOffsetMinutes,
		TimeField:         req.TimeField,
		RecurringCron:     req.RecurringCron,
		IsActive:          req.IsActive,
	}

	if req.StateID != nil {
		id, _ := uuid.Parse(*req.StateID)
		trigger.StateID = &id
	}
	if req.TransitionID != nil {
		id, _ := uuid.Parse(*req.TransitionID)
		trigger.TransitionID = &id
	}
	if req.Conditions != nil {
		trigger.Conditions = *req.Conditions
	}

	if err := h.service.UpdateTrigger(r.Context(), triggerID, trigger); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessMessageResponse(w, http.StatusOK, "Trigger updated successfully", trigger)
}

func (h *WorkflowHandler) DeleteTrigger(w http.ResponseWriter, r *http.Request) {
	triggerID, err := uuid.Parse(chi.URLParam(r, "triggerId"))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid trigger ID")
		return
	}

	if err := h.service.DeleteTrigger(r.Context(), triggerID); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessMessageResponse(w, http.StatusOK, "Trigger deleted successfully", nil)
}

// ============ Action Handlers ============

type CreateActionRequest struct {
	ActionType   string           `json:"action_type" validate:"required,oneof=send_whatsapp send_email update_field create_task"`
	ActionOrder  int              `json:"action_order"`
	TemplateID   *string          `json:"template_id"`
	ActionConfig *json.RawMessage `json:"action_config"`
}

func (h *WorkflowHandler) CreateAction(w http.ResponseWriter, r *http.Request) {
	triggerID, err := uuid.Parse(chi.URLParam(r, "triggerId"))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid trigger ID")
		return
	}

	var req CreateActionRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	action := &models.WorkflowAction{
		TriggerID:   triggerID,
		ActionType:  models.ActionType(req.ActionType),
		ActionOrder: req.ActionOrder,
	}

	if req.TemplateID != nil {
		id, err := uuid.Parse(*req.TemplateID)
		if err != nil {
			utils.ErrorResponse(w, http.StatusBadRequest, "Invalid template ID")
			return
		}
		action.TemplateID = &id
	}

	if req.ActionConfig != nil {
		action.ActionConfig = *req.ActionConfig
	}

	if err := h.service.CreateAction(r.Context(), action); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessMessageResponse(w, http.StatusCreated, "Action created successfully", action)
}

func (h *WorkflowHandler) UpdateAction(w http.ResponseWriter, r *http.Request) {
	actionID, err := uuid.Parse(chi.URLParam(r, "actionId"))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid action ID")
		return
	}

	var req struct {
		ActionType   string           `json:"action_type"`
		ActionOrder  int              `json:"action_order"`
		TemplateID   *string          `json:"template_id"`
		ActionConfig *json.RawMessage `json:"action_config"`
		IsActive     bool             `json:"is_active"`
	}
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	action := &models.WorkflowAction{
		ActionType:  models.ActionType(req.ActionType),
		ActionOrder: req.ActionOrder,
		IsActive:    req.IsActive,
	}

	if req.TemplateID != nil {
		id, _ := uuid.Parse(*req.TemplateID)
		action.TemplateID = &id
	}
	if req.ActionConfig != nil {
		action.ActionConfig = *req.ActionConfig
	}

	if err := h.service.UpdateAction(r.Context(), actionID, action); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessMessageResponse(w, http.StatusOK, "Action updated successfully", action)
}

func (h *WorkflowHandler) DeleteAction(w http.ResponseWriter, r *http.Request) {
	actionID, err := uuid.Parse(chi.URLParam(r, "actionId"))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid action ID")
		return
	}

	if err := h.service.DeleteAction(r.Context(), actionID); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessMessageResponse(w, http.StatusOK, "Action deleted successfully", nil)
}

// ============ Template Handlers ============

type CreateTemplateRequest struct {
	Name      string           `json:"name" validate:"required,min=2,max=100"`
	Channel   string           `json:"channel" validate:"required,oneof=whatsapp email"`
	Subject   *string          `json:"subject"`
	Body      string           `json:"body" validate:"required"`
	Variables *json.RawMessage `json:"variables"`
}

func (h *WorkflowHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	channel := r.URL.Query().Get("channel")

	templates, err := h.service.ListTemplates(r.Context(), orgID, channel)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(w, http.StatusOK, map[string]interface{}{
		"templates": templates,
	})
}

func (h *WorkflowHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid template ID")
		return
	}

	template, err := h.service.GetTemplateByID(r.Context(), id, orgID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	utils.SuccessResponse(w, http.StatusOK, template)
}

func (h *WorkflowHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	var req CreateTemplateRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	template := &models.MessageTemplate{
		OrganizationID: orgID,
		Name:           req.Name,
		Channel:        models.MessageChannel(req.Channel),
		Subject:        req.Subject,
		Body:           req.Body,
	}

	if req.Variables != nil {
		template.Variables = *req.Variables
	}

	if err := h.service.CreateTemplate(r.Context(), template); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessMessageResponse(w, http.StatusCreated, "Template created successfully", template)
}

func (h *WorkflowHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid template ID")
		return
	}

	var req struct {
		Name      string           `json:"name"`
		Channel   string           `json:"channel"`
		Subject   *string          `json:"subject"`
		Body      string           `json:"body"`
		Variables *json.RawMessage `json:"variables"`
		IsActive  bool             `json:"is_active"`
	}
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	template := &models.MessageTemplate{
		Name:     req.Name,
		Channel:  models.MessageChannel(req.Channel),
		Subject:  req.Subject,
		Body:     req.Body,
		IsActive: req.IsActive,
	}

	if req.Variables != nil {
		template.Variables = *req.Variables
	}

	if err := h.service.UpdateTemplate(r.Context(), id, orgID, template); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	updated, _ := h.service.GetTemplateByID(r.Context(), id, orgID)
	utils.SuccessMessageResponse(w, http.StatusOK, "Template updated successfully", updated)
}

func (h *WorkflowHandler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid template ID")
		return
	}

	if err := h.service.DeleteTemplate(r.Context(), id, orgID); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessMessageResponse(w, http.StatusOK, "Template deleted successfully", nil)
}

// ============ Execution Log Handlers ============

// GetExecutionLogs returns workflow execution logs with optional filters
func (h *WorkflowHandler) GetExecutionLogs(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	// Parse query parameters
	filters := services.ExecutionLogFilters{
		EntityType: r.URL.Query().Get("entity_type"),
		EventType:  r.URL.Query().Get("event_type"),
	}

	if workflowIDStr := r.URL.Query().Get("workflow_id"); workflowIDStr != "" {
		id, err := uuid.Parse(workflowIDStr)
		if err != nil {
			utils.ErrorResponse(w, http.StatusBadRequest, "Invalid workflow_id")
			return
		}
		filters.WorkflowID = &id
	}

	if entityIDStr := r.URL.Query().Get("entity_id"); entityIDStr != "" {
		id, err := uuid.Parse(entityIDStr)
		if err != nil {
			utils.ErrorResponse(w, http.StatusBadRequest, "Invalid entity_id")
			return
		}
		filters.EntityID = &id
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filters.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filters.Offset = offset
		}
	}

	logs, total, err := h.service.GetExecutionLogs(r.Context(), orgID, filters)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(w, http.StatusOK, map[string]interface{}{
		"logs":  logs,
		"total": total,
	})
}

// GetScheduledJobs returns pending scheduled jobs
func (h *WorkflowHandler) GetScheduledJobs(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	status := r.URL.Query().Get("status")
	if status == "" {
		status = "pending"
	}

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	jobs, err := h.service.GetScheduledJobs(r.Context(), orgID, status, limit)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(w, http.StatusOK, map[string]interface{}{
		"jobs": jobs,
	})
}

// ============ Default Workflow Handlers ============

// InitDefaultWorkflows creates the default workflows for the organization
func (h *WorkflowHandler) InitDefaultWorkflows(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	module := r.URL.Query().Get("module")

	var workflows []*models.Workflow

	// Create default workflows based on module
	switch module {
	case "construction":
		// Create budget workflow
		budgetWorkflow, err := h.service.CreateDefaultBudgetWorkflow(r.Context(), orgID)
		if err != nil {
			utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to create budget workflow: "+err.Error())
			return
		}
		if budgetWorkflow != nil {
			workflows = append(workflows, budgetWorkflow)
		}

		// Create project workflow
		projectWorkflow, err := h.service.CreateDefaultProjectWorkflow(r.Context(), orgID)
		if err != nil {
			utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to create project workflow: "+err.Error())
			return
		}
		if projectWorkflow != nil {
			workflows = append(workflows, projectWorkflow)
		}

		// Create default templates for construction
		if err := h.service.CreateDefaultTemplates(r.Context(), orgID, "construction"); err != nil {
			utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to create default templates: "+err.Error())
			return
		}
	case "appointments":
		// Create default templates for appointments
		if err := h.service.CreateDefaultTemplates(r.Context(), orgID, "appointments"); err != nil {
			utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to create default templates: "+err.Error())
			return
		}
	case "":
		// Create all default workflows
		budgetWorkflow, err := h.service.CreateDefaultBudgetWorkflow(r.Context(), orgID)
		if err != nil {
			utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to create budget workflow: "+err.Error())
			return
		}
		if budgetWorkflow != nil {
			workflows = append(workflows, budgetWorkflow)
		}

		projectWorkflow, err := h.service.CreateDefaultProjectWorkflow(r.Context(), orgID)
		if err != nil {
			utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to create project workflow: "+err.Error())
			return
		}
		if projectWorkflow != nil {
			workflows = append(workflows, projectWorkflow)
		}

		// Create default templates for all modules
		if err := h.service.CreateDefaultTemplates(r.Context(), orgID, "construction"); err != nil {
			utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to create default templates: "+err.Error())
			return
		}
		if err := h.service.CreateDefaultTemplates(r.Context(), orgID, "appointments"); err != nil {
			utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to create default templates: "+err.Error())
			return
		}
	default:
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid module. Use 'construction', 'appointments', or leave empty for all")
		return
	}

	utils.SuccessMessageResponse(w, http.StatusCreated, "Default workflows initialized", map[string]interface{}{
		"workflows": workflows,
	})
}

// TestTrigger tests a workflow trigger with sample data (preview mode)
func (h *WorkflowHandler) TestTrigger(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	workflowIDStr := chi.URLParam(r, "id")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid workflow ID")
		return
	}

	triggerIDStr := chi.URLParam(r, "triggerId")
	triggerID, err := uuid.Parse(triggerIDStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid trigger ID")
		return
	}

	result, err := h.service.TestWorkflowTrigger(r.Context(), orgID, workflowID, triggerID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to test trigger: "+err.Error())
		return
	}

	utils.SuccessMessageResponse(w, http.StatusOK, "Trigger test executed", map[string]interface{}{
		"result": result,
	})
}

// GetAvailableVariables returns available template variables for an entity type
func (h *WorkflowHandler) GetAvailableVariables(w http.ResponseWriter, r *http.Request) {
	entityType := r.URL.Query().Get("entity_type")
	if entityType == "" {
		entityType = "session"
	}

	// Return sample data and variable list for the entity type
	sampleData := services.GetSampleDataForEntityType(entityType)

	type VariableInfo struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		SampleValue string `json:"sample_value"`
	}

	variables := make([]VariableInfo, 0)
	for key, value := range sampleData {
		variables = append(variables, VariableInfo{
			Name:        key,
			Description: getVariableDescription(key),
			SampleValue: fmt.Sprintf("%v", value),
		})
	}

	utils.SuccessResponse(w, http.StatusOK, map[string]interface{}{
		"entity_type": entityType,
		"variables":   variables,
	})
}

// getVariableDescription returns a human-readable description for a variable
func getVariableDescription(varName string) string {
	descriptions := map[string]string{
		"patient_name":       "Nome do paciente",
		"patient_phone":      "Telefone do paciente",
		"patient_email":      "Email do paciente",
		"therapist_name":     "Nome do terapeuta",
		"session_date":       "Data da sessão (DD/MM/AAAA)",
		"session_time":       "Hora da sessão (HH:MM)",
		"session_type":       "Tipo de sessão",
		"amount":             "Valor da sessão/pagamento",
		"client_name":        "Nome do cliente",
		"client_email":       "Email do cliente",
		"client_phone":       "Telefone do cliente",
		"project_name":       "Nome do projeto",
		"project_number":     "Número do projeto",
		"project_status":     "Estado do projeto",
		"budget_number":      "Número do orçamento",
		"budget_total":       "Valor total do orçamento",
		"budget_link":        "Link para visualizar orçamento",
		"approval_link":      "Link para aprovar orçamento",
		"organization_name":  "Nome da organização",
		"organization_email": "Email da organização",
	}
	if desc, ok := descriptions[varName]; ok {
		return desc
	}
	return varName
}
