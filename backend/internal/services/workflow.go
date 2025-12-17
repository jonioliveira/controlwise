package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/controlewise/backend/internal/database"
	"github.com/controlewise/backend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type WorkflowService struct {
	db *database.DB
}

func NewWorkflowService(db *database.DB) *WorkflowService {
	return &WorkflowService{db: db}
}

// ============ Workflow CRUD ============

// ListWorkflows returns all workflows for an organization
func (s *WorkflowService) ListWorkflows(ctx context.Context, orgID uuid.UUID, module string) ([]*models.WorkflowWithStats, error) {
	query := `
		SELECT
			w.id, w.organization_id, w.name, w.description, w.module, w.entity_type,
			w.is_active, w.is_default, w.created_at, w.updated_at,
			(SELECT COUNT(*) FROM workflow_states WHERE workflow_id = w.id) as state_count,
			(SELECT COUNT(*) FROM workflow_triggers WHERE workflow_id = w.id) as trigger_count,
			(SELECT COUNT(*) FROM workflow_actions wa
			 JOIN workflow_triggers wt ON wt.id = wa.trigger_id
			 WHERE wt.workflow_id = w.id) as action_count
		FROM workflows w
		WHERE w.organization_id = $1`

	args := []interface{}{orgID}
	if module != "" {
		query += " AND w.module = $2"
		args = append(args, module)
	}
	query += " ORDER BY w.name ASC"

	rows, err := s.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list workflows: %w", err)
	}
	defer rows.Close()

	var workflows []*models.WorkflowWithStats
	for rows.Next() {
		var w models.WorkflowWithStats
		err := rows.Scan(
			&w.ID, &w.OrganizationID, &w.Name, &w.Description, &w.Module, &w.EntityType,
			&w.IsActive, &w.IsDefault, &w.CreatedAt, &w.UpdatedAt,
			&w.StateCount, &w.TriggerCount, &w.ActionCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workflow: %w", err)
		}
		workflows = append(workflows, &w)
	}

	return workflows, nil
}

// GetWorkflowByID returns a single workflow with all states, transitions, triggers, and actions
func (s *WorkflowService) GetWorkflowByID(ctx context.Context, id, orgID uuid.UUID) (*models.Workflow, error) {
	var w models.Workflow
	err := s.db.Pool.QueryRow(ctx, `
		SELECT id, organization_id, name, description, module, entity_type,
		       is_active, is_default, created_at, updated_at
		FROM workflows
		WHERE id = $1 AND organization_id = $2
	`, id, orgID).Scan(
		&w.ID, &w.OrganizationID, &w.Name, &w.Description, &w.Module, &w.EntityType,
		&w.IsActive, &w.IsDefault, &w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("workflow not found")
		}
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	// Load states
	states, err := s.ListStates(ctx, id)
	if err != nil {
		return nil, err
	}
	w.States = states

	// Load transitions
	transitions, err := s.ListTransitions(ctx, id)
	if err != nil {
		return nil, err
	}
	w.Transitions = transitions

	// Load triggers with actions
	triggers, err := s.ListTriggers(ctx, id)
	if err != nil {
		return nil, err
	}
	w.Triggers = triggers

	return &w, nil
}

// CreateWorkflow creates a new workflow
func (s *WorkflowService) CreateWorkflow(ctx context.Context, workflow *models.Workflow) error {
	workflow.ID = uuid.New()
	workflow.IsActive = true

	_, err := s.db.Pool.Exec(ctx, `
		INSERT INTO workflows (id, organization_id, name, description, module, entity_type, is_active, is_default)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, workflow.ID, workflow.OrganizationID, workflow.Name, workflow.Description,
		workflow.Module, workflow.EntityType, workflow.IsActive, workflow.IsDefault)

	if err != nil {
		return fmt.Errorf("failed to create workflow: %w", err)
	}
	return nil
}

// UpdateWorkflow updates an existing workflow
func (s *WorkflowService) UpdateWorkflow(ctx context.Context, id, orgID uuid.UUID, workflow *models.Workflow) error {
	result, err := s.db.Pool.Exec(ctx, `
		UPDATE workflows
		SET name = $1, description = $2, is_active = $3, is_default = $4, updated_at = NOW()
		WHERE id = $5 AND organization_id = $6
	`, workflow.Name, workflow.Description, workflow.IsActive, workflow.IsDefault, id, orgID)

	if err != nil {
		return fmt.Errorf("failed to update workflow: %w", err)
	}
	if result.RowsAffected() == 0 {
		return errors.New("workflow not found")
	}
	return nil
}

// DeleteWorkflow deletes a workflow
func (s *WorkflowService) DeleteWorkflow(ctx context.Context, id, orgID uuid.UUID) error {
	result, err := s.db.Pool.Exec(ctx, `
		DELETE FROM workflows WHERE id = $1 AND organization_id = $2
	`, id, orgID)

	if err != nil {
		return fmt.Errorf("failed to delete workflow: %w", err)
	}
	if result.RowsAffected() == 0 {
		return errors.New("workflow not found")
	}
	return nil
}

// DuplicateWorkflow creates a copy of an existing workflow
func (s *WorkflowService) DuplicateWorkflow(ctx context.Context, id, orgID uuid.UUID, newName string) (*models.Workflow, error) {
	// Get the original workflow
	original, err := s.GetWorkflowByID(ctx, id, orgID)
	if err != nil {
		return nil, err
	}

	// Create new workflow
	newWorkflow := &models.Workflow{
		OrganizationID: orgID,
		Name:           newName,
		Description:    original.Description,
		Module:         original.Module,
		EntityType:     original.EntityType,
		IsActive:       false, // Start as inactive
		IsDefault:      false,
	}
	if err := s.CreateWorkflow(ctx, newWorkflow); err != nil {
		return nil, err
	}

	// Map old state IDs to new state IDs
	stateMap := make(map[uuid.UUID]uuid.UUID)

	// Copy states
	for _, state := range original.States {
		newState := &models.WorkflowState{
			WorkflowID:  newWorkflow.ID,
			Name:        state.Name,
			DisplayName: state.DisplayName,
			Description: state.Description,
			StateType:   state.StateType,
			Color:       state.Color,
			Position:    state.Position,
		}
		if err := s.CreateState(ctx, newState); err != nil {
			return nil, err
		}
		stateMap[state.ID] = newState.ID
	}

	// Copy transitions
	for _, trans := range original.Transitions {
		newTrans := &models.WorkflowTransition{
			WorkflowID:           newWorkflow.ID,
			FromStateID:          stateMap[trans.FromStateID],
			ToStateID:            stateMap[trans.ToStateID],
			Name:                 trans.Name,
			RequiresConfirmation: trans.RequiresConfirmation,
		}
		if err := s.CreateTransition(ctx, newTrans); err != nil {
			return nil, err
		}
	}

	// Copy triggers and actions
	for _, trigger := range original.Triggers {
		newTrigger := &models.WorkflowTrigger{
			WorkflowID:        newWorkflow.ID,
			TriggerType:       trigger.TriggerType,
			TimeOffsetMinutes: trigger.TimeOffsetMinutes,
			TimeField:         trigger.TimeField,
			RecurringCron:     trigger.RecurringCron,
			Conditions:        trigger.Conditions,
			IsActive:          trigger.IsActive,
		}
		if trigger.StateID != nil {
			newStateID := stateMap[*trigger.StateID]
			newTrigger.StateID = &newStateID
		}
		if err := s.CreateTrigger(ctx, newTrigger); err != nil {
			return nil, err
		}

		// Copy actions
		for _, action := range trigger.Actions {
			newAction := &models.WorkflowAction{
				TriggerID:    newTrigger.ID,
				ActionType:   action.ActionType,
				ActionOrder:  action.ActionOrder,
				TemplateID:   action.TemplateID,
				ActionConfig: action.ActionConfig,
				IsActive:     action.IsActive,
			}
			if err := s.CreateAction(ctx, newAction); err != nil {
				return nil, err
			}
		}
	}

	return s.GetWorkflowByID(ctx, newWorkflow.ID, orgID)
}

// ============ State CRUD ============

// ListStates returns all states for a workflow
func (s *WorkflowService) ListStates(ctx context.Context, workflowID uuid.UUID) ([]models.WorkflowState, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT id, workflow_id, name, display_name, description, state_type, color, position, created_at
		FROM workflow_states
		WHERE workflow_id = $1
		ORDER BY position ASC
	`, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to list states: %w", err)
	}
	defer rows.Close()

	var states []models.WorkflowState
	for rows.Next() {
		var state models.WorkflowState
		err := rows.Scan(
			&state.ID, &state.WorkflowID, &state.Name, &state.DisplayName,
			&state.Description, &state.StateType, &state.Color, &state.Position, &state.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan state: %w", err)
		}
		states = append(states, state)
	}

	return states, nil
}

// CreateState creates a new workflow state
func (s *WorkflowService) CreateState(ctx context.Context, state *models.WorkflowState) error {
	state.ID = uuid.New()

	_, err := s.db.Pool.Exec(ctx, `
		INSERT INTO workflow_states (id, workflow_id, name, display_name, description, state_type, color, position)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, state.ID, state.WorkflowID, state.Name, state.DisplayName,
		state.Description, state.StateType, state.Color, state.Position)

	if err != nil {
		return fmt.Errorf("failed to create state: %w", err)
	}
	return nil
}

// UpdateState updates an existing state
func (s *WorkflowService) UpdateState(ctx context.Context, id uuid.UUID, state *models.WorkflowState) error {
	result, err := s.db.Pool.Exec(ctx, `
		UPDATE workflow_states
		SET name = $1, display_name = $2, description = $3, state_type = $4, color = $5, position = $6
		WHERE id = $7
	`, state.Name, state.DisplayName, state.Description, state.StateType, state.Color, state.Position, id)

	if err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}
	if result.RowsAffected() == 0 {
		return errors.New("state not found")
	}
	return nil
}

// DeleteState deletes a state
func (s *WorkflowService) DeleteState(ctx context.Context, id uuid.UUID) error {
	result, err := s.db.Pool.Exec(ctx, `DELETE FROM workflow_states WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete state: %w", err)
	}
	if result.RowsAffected() == 0 {
		return errors.New("state not found")
	}
	return nil
}

// ReorderStates updates the position of all states in a workflow
func (s *WorkflowService) ReorderStates(ctx context.Context, workflowID uuid.UUID, stateIDs []uuid.UUID) error {
	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	for i, stateID := range stateIDs {
		_, err := tx.Exec(ctx, `
			UPDATE workflow_states SET position = $1 WHERE id = $2 AND workflow_id = $3
		`, i, stateID, workflowID)
		if err != nil {
			return fmt.Errorf("failed to update state position: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// ============ Transition CRUD ============

// ListTransitions returns all transitions for a workflow
func (s *WorkflowService) ListTransitions(ctx context.Context, workflowID uuid.UUID) ([]models.WorkflowTransition, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT id, workflow_id, from_state_id, to_state_id, name, requires_confirmation, created_at
		FROM workflow_transitions
		WHERE workflow_id = $1
	`, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to list transitions: %w", err)
	}
	defer rows.Close()

	var transitions []models.WorkflowTransition
	for rows.Next() {
		var t models.WorkflowTransition
		err := rows.Scan(
			&t.ID, &t.WorkflowID, &t.FromStateID, &t.ToStateID,
			&t.Name, &t.RequiresConfirmation, &t.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transition: %w", err)
		}
		transitions = append(transitions, t)
	}

	return transitions, nil
}

// CreateTransition creates a new transition
func (s *WorkflowService) CreateTransition(ctx context.Context, transition *models.WorkflowTransition) error {
	transition.ID = uuid.New()

	_, err := s.db.Pool.Exec(ctx, `
		INSERT INTO workflow_transitions (id, workflow_id, from_state_id, to_state_id, name, requires_confirmation)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, transition.ID, transition.WorkflowID, transition.FromStateID,
		transition.ToStateID, transition.Name, transition.RequiresConfirmation)

	if err != nil {
		return fmt.Errorf("failed to create transition: %w", err)
	}
	return nil
}

// DeleteTransition deletes a transition
func (s *WorkflowService) DeleteTransition(ctx context.Context, id uuid.UUID) error {
	result, err := s.db.Pool.Exec(ctx, `DELETE FROM workflow_transitions WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete transition: %w", err)
	}
	if result.RowsAffected() == 0 {
		return errors.New("transition not found")
	}
	return nil
}

// ============ Trigger CRUD ============

// ListTriggers returns all triggers for a workflow with their actions
func (s *WorkflowService) ListTriggers(ctx context.Context, workflowID uuid.UUID) ([]models.WorkflowTrigger, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT id, workflow_id, state_id, transition_id, trigger_type,
		       time_offset_minutes, time_field, recurring_cron, conditions, is_active, created_at
		FROM workflow_triggers
		WHERE workflow_id = $1
	`, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to list triggers: %w", err)
	}
	defer rows.Close()

	var triggers []models.WorkflowTrigger
	for rows.Next() {
		var t models.WorkflowTrigger
		err := rows.Scan(
			&t.ID, &t.WorkflowID, &t.StateID, &t.TransitionID, &t.TriggerType,
			&t.TimeOffsetMinutes, &t.TimeField, &t.RecurringCron, &t.Conditions, &t.IsActive, &t.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trigger: %w", err)
		}
		triggers = append(triggers, t)
	}

	// Load actions for each trigger
	for i := range triggers {
		actions, err := s.ListActions(ctx, triggers[i].ID)
		if err != nil {
			return nil, err
		}
		triggers[i].Actions = actions
	}

	return triggers, nil
}

// GetTriggerByID returns a single trigger with its actions
func (s *WorkflowService) GetTriggerByID(ctx context.Context, id uuid.UUID) (*models.WorkflowTrigger, error) {
	var t models.WorkflowTrigger
	err := s.db.Pool.QueryRow(ctx, `
		SELECT id, workflow_id, state_id, transition_id, trigger_type,
		       time_offset_minutes, time_field, recurring_cron, conditions, is_active, created_at
		FROM workflow_triggers
		WHERE id = $1
	`, id).Scan(
		&t.ID, &t.WorkflowID, &t.StateID, &t.TransitionID, &t.TriggerType,
		&t.TimeOffsetMinutes, &t.TimeField, &t.RecurringCron, &t.Conditions, &t.IsActive, &t.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("trigger not found")
		}
		return nil, fmt.Errorf("failed to get trigger: %w", err)
	}

	// Load actions for the trigger
	actions, err := s.ListActions(ctx, t.ID)
	if err != nil {
		return nil, err
	}
	t.Actions = actions

	return &t, nil
}

// CreateTrigger creates a new trigger
func (s *WorkflowService) CreateTrigger(ctx context.Context, trigger *models.WorkflowTrigger) error {
	trigger.ID = uuid.New()
	trigger.IsActive = true

	_, err := s.db.Pool.Exec(ctx, `
		INSERT INTO workflow_triggers (id, workflow_id, state_id, transition_id, trigger_type,
		                               time_offset_minutes, time_field, recurring_cron, conditions, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, trigger.ID, trigger.WorkflowID, trigger.StateID, trigger.TransitionID, trigger.TriggerType,
		trigger.TimeOffsetMinutes, trigger.TimeField, trigger.RecurringCron, trigger.Conditions, trigger.IsActive)

	if err != nil {
		return fmt.Errorf("failed to create trigger: %w", err)
	}
	return nil
}

// UpdateTrigger updates an existing trigger
func (s *WorkflowService) UpdateTrigger(ctx context.Context, id uuid.UUID, trigger *models.WorkflowTrigger) error {
	result, err := s.db.Pool.Exec(ctx, `
		UPDATE workflow_triggers
		SET state_id = $1, transition_id = $2, trigger_type = $3, time_offset_minutes = $4,
		    time_field = $5, recurring_cron = $6, conditions = $7, is_active = $8
		WHERE id = $9
	`, trigger.StateID, trigger.TransitionID, trigger.TriggerType, trigger.TimeOffsetMinutes,
		trigger.TimeField, trigger.RecurringCron, trigger.Conditions, trigger.IsActive, id)

	if err != nil {
		return fmt.Errorf("failed to update trigger: %w", err)
	}
	if result.RowsAffected() == 0 {
		return errors.New("trigger not found")
	}
	return nil
}

// DeleteTrigger deletes a trigger
func (s *WorkflowService) DeleteTrigger(ctx context.Context, id uuid.UUID) error {
	result, err := s.db.Pool.Exec(ctx, `DELETE FROM workflow_triggers WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete trigger: %w", err)
	}
	if result.RowsAffected() == 0 {
		return errors.New("trigger not found")
	}
	return nil
}

// ============ Action CRUD ============

// ListActions returns all actions for a trigger
func (s *WorkflowService) ListActions(ctx context.Context, triggerID uuid.UUID) ([]models.WorkflowAction, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT id, trigger_id, action_type, action_order, template_id, action_config, is_active, created_at
		FROM workflow_actions
		WHERE trigger_id = $1
		ORDER BY action_order ASC
	`, triggerID)
	if err != nil {
		return nil, fmt.Errorf("failed to list actions: %w", err)
	}
	defer rows.Close()

	var actions []models.WorkflowAction
	for rows.Next() {
		var a models.WorkflowAction
		err := rows.Scan(
			&a.ID, &a.TriggerID, &a.ActionType, &a.ActionOrder,
			&a.TemplateID, &a.ActionConfig, &a.IsActive, &a.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan action: %w", err)
		}
		actions = append(actions, a)
	}

	return actions, nil
}

// CreateAction creates a new action
func (s *WorkflowService) CreateAction(ctx context.Context, action *models.WorkflowAction) error {
	action.ID = uuid.New()
	action.IsActive = true

	if action.ActionConfig == nil {
		action.ActionConfig = json.RawMessage("{}")
	}

	_, err := s.db.Pool.Exec(ctx, `
		INSERT INTO workflow_actions (id, trigger_id, action_type, action_order, template_id, action_config, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, action.ID, action.TriggerID, action.ActionType, action.ActionOrder,
		action.TemplateID, action.ActionConfig, action.IsActive)

	if err != nil {
		return fmt.Errorf("failed to create action: %w", err)
	}
	return nil
}

// UpdateAction updates an existing action
func (s *WorkflowService) UpdateAction(ctx context.Context, id uuid.UUID, action *models.WorkflowAction) error {
	result, err := s.db.Pool.Exec(ctx, `
		UPDATE workflow_actions
		SET action_type = $1, action_order = $2, template_id = $3, action_config = $4, is_active = $5
		WHERE id = $6
	`, action.ActionType, action.ActionOrder, action.TemplateID, action.ActionConfig, action.IsActive, id)

	if err != nil {
		return fmt.Errorf("failed to update action: %w", err)
	}
	if result.RowsAffected() == 0 {
		return errors.New("action not found")
	}
	return nil
}

// DeleteAction deletes an action
func (s *WorkflowService) DeleteAction(ctx context.Context, id uuid.UUID) error {
	result, err := s.db.Pool.Exec(ctx, `DELETE FROM workflow_actions WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete action: %w", err)
	}
	if result.RowsAffected() == 0 {
		return errors.New("action not found")
	}
	return nil
}

// ============ Message Template CRUD ============

// ListTemplates returns all message templates for an organization
func (s *WorkflowService) ListTemplates(ctx context.Context, orgID uuid.UUID, channel string) ([]*models.MessageTemplate, error) {
	query := `
		SELECT id, organization_id, name, channel, subject, body, variables, is_active, created_at, updated_at
		FROM message_templates
		WHERE organization_id = $1`

	args := []interface{}{orgID}
	if channel != "" {
		query += " AND channel = $2"
		args = append(args, channel)
	}
	query += " ORDER BY name ASC"

	rows, err := s.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}
	defer rows.Close()

	var templates []*models.MessageTemplate
	for rows.Next() {
		var t models.MessageTemplate
		err := rows.Scan(
			&t.ID, &t.OrganizationID, &t.Name, &t.Channel, &t.Subject,
			&t.Body, &t.Variables, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan template: %w", err)
		}
		templates = append(templates, &t)
	}

	return templates, nil
}

// GetTemplateByID returns a single template
func (s *WorkflowService) GetTemplateByID(ctx context.Context, id, orgID uuid.UUID) (*models.MessageTemplate, error) {
	var t models.MessageTemplate
	err := s.db.Pool.QueryRow(ctx, `
		SELECT id, organization_id, name, channel, subject, body, variables, is_active, created_at, updated_at
		FROM message_templates
		WHERE id = $1 AND organization_id = $2
	`, id, orgID).Scan(
		&t.ID, &t.OrganizationID, &t.Name, &t.Channel, &t.Subject,
		&t.Body, &t.Variables, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("template not found")
		}
		return nil, fmt.Errorf("failed to get template: %w", err)
	}
	return &t, nil
}

// CreateTemplate creates a new message template
func (s *WorkflowService) CreateTemplate(ctx context.Context, template *models.MessageTemplate) error {
	template.ID = uuid.New()
	template.IsActive = true

	if template.Variables == nil {
		template.Variables = json.RawMessage("[]")
	}

	_, err := s.db.Pool.Exec(ctx, `
		INSERT INTO message_templates (id, organization_id, name, channel, subject, body, variables, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, template.ID, template.OrganizationID, template.Name, template.Channel,
		template.Subject, template.Body, template.Variables, template.IsActive)

	if err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}
	return nil
}

// UpdateTemplate updates an existing template
func (s *WorkflowService) UpdateTemplate(ctx context.Context, id, orgID uuid.UUID, template *models.MessageTemplate) error {
	result, err := s.db.Pool.Exec(ctx, `
		UPDATE message_templates
		SET name = $1, channel = $2, subject = $3, body = $4, variables = $5, is_active = $6, updated_at = NOW()
		WHERE id = $7 AND organization_id = $8
	`, template.Name, template.Channel, template.Subject, template.Body,
		template.Variables, template.IsActive, id, orgID)

	if err != nil {
		return fmt.Errorf("failed to update template: %w", err)
	}
	if result.RowsAffected() == 0 {
		return errors.New("template not found")
	}
	return nil
}

// DeleteTemplate deletes a template
func (s *WorkflowService) DeleteTemplate(ctx context.Context, id, orgID uuid.UUID) error {
	result, err := s.db.Pool.Exec(ctx, `
		DELETE FROM message_templates WHERE id = $1 AND organization_id = $2
	`, id, orgID)

	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}
	if result.RowsAffected() == 0 {
		return errors.New("template not found")
	}
	return nil
}

// ============ Workflow Defaults ============

// GetDefaultWorkflow returns the default workflow for a module and entity type
func (s *WorkflowService) GetDefaultWorkflow(ctx context.Context, orgID uuid.UUID, module models.WorkflowModule, entityType models.WorkflowEntityType) (*models.Workflow, error) {
	var w models.Workflow
	err := s.db.Pool.QueryRow(ctx, `
		SELECT id, organization_id, name, description, module, entity_type,
		       is_active, is_default, created_at, updated_at
		FROM workflows
		WHERE organization_id = $1 AND module = $2 AND entity_type = $3 AND is_default = true AND is_active = true
	`, orgID, module, entityType).Scan(
		&w.ID, &w.OrganizationID, &w.Name, &w.Description, &w.Module, &w.EntityType,
		&w.IsActive, &w.IsDefault, &w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // No default workflow
		}
		return nil, fmt.Errorf("failed to get default workflow: %w", err)
	}

	// Load full workflow data
	return s.GetWorkflowByID(ctx, w.ID, orgID)
}

// SetDefaultWorkflow sets a workflow as the default for its module/entity type
func (s *WorkflowService) SetDefaultWorkflow(ctx context.Context, id, orgID uuid.UUID) error {
	// Get the workflow to find its module and entity type
	var module, entityType string
	err := s.db.Pool.QueryRow(ctx, `
		SELECT module, entity_type FROM workflows WHERE id = $1 AND organization_id = $2
	`, id, orgID).Scan(&module, &entityType)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("workflow not found")
		}
		return fmt.Errorf("failed to get workflow: %w", err)
	}

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Unset current default
	_, err = tx.Exec(ctx, `
		UPDATE workflows SET is_default = false
		WHERE organization_id = $1 AND module = $2 AND entity_type = $3
	`, orgID, module, entityType)
	if err != nil {
		return fmt.Errorf("failed to unset default: %w", err)
	}

	// Set new default
	_, err = tx.Exec(ctx, `
		UPDATE workflows SET is_default = true WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("failed to set default: %w", err)
	}

	return tx.Commit(ctx)
}

// ============ Workflow Execution ============

// OnSessionStateChange triggers workflow actions when a session changes state
func (s *WorkflowService) OnSessionStateChange(ctx context.Context, orgID uuid.UUID, sessionID uuid.UUID, fromStatus, toStatus string, scheduledAt time.Time) error {
	// Get the default workflow for sessions
	workflow, err := s.GetDefaultWorkflow(ctx, orgID, models.WorkflowModuleAppointments, models.WorkflowEntitySession)
	if err != nil {
		return fmt.Errorf("failed to get default workflow: %w", err)
	}
	if workflow == nil {
		// No default workflow configured, nothing to do
		return nil
	}

	// Find the state in the workflow that matches the new status
	var targetState *models.WorkflowState
	for i := range workflow.States {
		if workflow.States[i].Name == toStatus {
			targetState = &workflow.States[i]
			break
		}
	}
	if targetState == nil {
		// No matching state in workflow
		return nil
	}

	// If we're exiting a state, cancel pending jobs for this session
	if fromStatus != "" {
		if err := s.cancelPendingJobsForEntity(ctx, "session", sessionID); err != nil {
			// Log but don't fail
			fmt.Printf("Failed to cancel pending jobs: %v\n", err)
		}
	}

	// Process triggers for this state
	for _, trigger := range workflow.Triggers {
		if trigger.StateID == nil || *trigger.StateID != targetState.ID {
			continue
		}
		if !trigger.IsActive {
			continue
		}

		switch trigger.TriggerType {
		case models.TriggerTypeOnEnter:
			// Execute immediately by scheduling for now
			if err := s.scheduleJob(ctx, orgID, trigger.ID, "session", sessionID, time.Now()); err != nil {
				return fmt.Errorf("failed to schedule on_enter trigger: %w", err)
			}
		case models.TriggerTypeTimeBefore:
			// Schedule for time before scheduled_at
			if trigger.TimeOffsetMinutes != nil {
				offset := time.Duration(*trigger.TimeOffsetMinutes) * time.Minute
				executeAt := scheduledAt.Add(-offset)
				if executeAt.After(time.Now()) {
					if err := s.scheduleJob(ctx, orgID, trigger.ID, "session", sessionID, executeAt); err != nil {
						return fmt.Errorf("failed to schedule time_before trigger: %w", err)
					}
				}
			}
		case models.TriggerTypeTimeAfter:
			// Schedule for time after scheduled_at
			if trigger.TimeOffsetMinutes != nil {
				offset := time.Duration(*trigger.TimeOffsetMinutes) * time.Minute
				executeAt := scheduledAt.Add(offset)
				if err := s.scheduleJob(ctx, orgID, trigger.ID, "session", sessionID, executeAt); err != nil {
					return fmt.Errorf("failed to schedule time_after trigger: %w", err)
				}
			}
		}
	}

	return nil
}

// scheduleJob creates a scheduled job for later execution
func (s *WorkflowService) scheduleJob(ctx context.Context, orgID, triggerID uuid.UUID, entityType string, entityID uuid.UUID, scheduledFor time.Time) error {
	_, err := s.db.Pool.Exec(ctx, `
		INSERT INTO scheduled_jobs (id, organization_id, trigger_id, entity_type, entity_id, scheduled_for, status)
		VALUES ($1, $2, $3, $4, $5, $6, 'pending')
	`, uuid.New(), orgID, triggerID, entityType, entityID, scheduledFor)
	return err
}

// cancelPendingJobsForEntity cancels all pending jobs for an entity
func (s *WorkflowService) cancelPendingJobsForEntity(ctx context.Context, entityType string, entityID uuid.UUID) error {
	_, err := s.db.Pool.Exec(ctx, `
		UPDATE scheduled_jobs
		SET status = 'cancelled'
		WHERE entity_type = $1 AND entity_id = $2 AND status = 'pending'
	`, entityType, entityID)
	return err
}

// GetScheduledJobStats returns statistics about scheduled jobs
func (s *WorkflowService) GetScheduledJobStats(ctx context.Context, orgID uuid.UUID) (map[string]int, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT status, COUNT(*) as count
		FROM scheduled_jobs
		WHERE organization_id = $1
		GROUP BY status
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		stats[status] = count
	}
	return stats, nil
}

// OnBudgetStateChange triggers workflow actions when a budget changes state
func (s *WorkflowService) OnBudgetStateChange(ctx context.Context, orgID uuid.UUID, budgetID uuid.UUID, fromStatus, toStatus string) error {
	// Get the default workflow for budgets
	workflow, err := s.GetDefaultWorkflow(ctx, orgID, models.WorkflowModuleConstruction, models.WorkflowEntityBudget)
	if err != nil {
		return fmt.Errorf("failed to get default workflow: %w", err)
	}
	if workflow == nil {
		// No default workflow configured, nothing to do
		return nil
	}

	// Find the state in the workflow that matches the new status
	var targetState *models.WorkflowState
	for i := range workflow.States {
		if workflow.States[i].Name == toStatus {
			targetState = &workflow.States[i]
			break
		}
	}
	if targetState == nil {
		// No matching state in workflow
		return nil
	}

	// If we're exiting a state, cancel pending jobs for this budget
	if fromStatus != "" {
		if err := s.cancelPendingJobsForEntity(ctx, "budget", budgetID); err != nil {
			// Log but don't fail
			fmt.Printf("Failed to cancel pending jobs: %v\n", err)
		}
	}

	// Process triggers for this state
	for _, trigger := range workflow.Triggers {
		if trigger.StateID == nil || *trigger.StateID != targetState.ID {
			continue
		}
		if !trigger.IsActive {
			continue
		}

		switch trigger.TriggerType {
		case models.TriggerTypeOnEnter:
			// Execute immediately by scheduling for now
			if err := s.scheduleJob(ctx, orgID, trigger.ID, "budget", budgetID, time.Now()); err != nil {
				return fmt.Errorf("failed to schedule on_enter trigger: %w", err)
			}
		case models.TriggerTypeTimeAfter:
			// Schedule for time after the state change
			if trigger.TimeOffsetMinutes != nil {
				offset := time.Duration(*trigger.TimeOffsetMinutes) * time.Minute
				executeAt := time.Now().Add(offset)
				if err := s.scheduleJob(ctx, orgID, trigger.ID, "budget", budgetID, executeAt); err != nil {
					return fmt.Errorf("failed to schedule time_after trigger: %w", err)
				}
			}
		}
	}

	return nil
}

// OnProjectStateChange triggers workflow actions when a project changes state
func (s *WorkflowService) OnProjectStateChange(ctx context.Context, orgID uuid.UUID, projectID uuid.UUID, fromStatus, toStatus string) error {
	// Get the default workflow for projects
	workflow, err := s.GetDefaultWorkflow(ctx, orgID, models.WorkflowModuleConstruction, models.WorkflowEntityProject)
	if err != nil {
		return fmt.Errorf("failed to get default workflow: %w", err)
	}
	if workflow == nil {
		// No default workflow configured, nothing to do
		return nil
	}

	// Find the state in the workflow that matches the new status
	var targetState *models.WorkflowState
	for i := range workflow.States {
		if workflow.States[i].Name == toStatus {
			targetState = &workflow.States[i]
			break
		}
	}
	if targetState == nil {
		// No matching state in workflow
		return nil
	}

	// If we're exiting a state, cancel pending jobs for this project
	if fromStatus != "" {
		if err := s.cancelPendingJobsForEntity(ctx, "project", projectID); err != nil {
			// Log but don't fail
			fmt.Printf("Failed to cancel pending jobs: %v\n", err)
		}
	}

	// Process triggers for this state
	for _, trigger := range workflow.Triggers {
		if trigger.StateID == nil || *trigger.StateID != targetState.ID {
			continue
		}
		if !trigger.IsActive {
			continue
		}

		switch trigger.TriggerType {
		case models.TriggerTypeOnEnter:
			// Execute immediately by scheduling for now
			if err := s.scheduleJob(ctx, orgID, trigger.ID, "project", projectID, time.Now()); err != nil {
				return fmt.Errorf("failed to schedule on_enter trigger: %w", err)
			}
		case models.TriggerTypeTimeAfter:
			// Schedule for time after the state change
			if trigger.TimeOffsetMinutes != nil {
				offset := time.Duration(*trigger.TimeOffsetMinutes) * time.Minute
				executeAt := time.Now().Add(offset)
				if err := s.scheduleJob(ctx, orgID, trigger.ID, "project", projectID, executeAt); err != nil {
					return fmt.Errorf("failed to schedule time_after trigger: %w", err)
				}
			}
		}
	}

	return nil
}

// ============ Default Workflow Creation ============

// CreateDefaultBudgetWorkflow creates the default workflow for budget lifecycle
func (s *WorkflowService) CreateDefaultBudgetWorkflow(ctx context.Context, orgID uuid.UUID) (*models.Workflow, error) {
	// Check if default workflow already exists
	existing, err := s.GetDefaultWorkflow(ctx, orgID, models.WorkflowModuleConstruction, models.WorkflowEntityBudget)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	// Create the workflow
	workflow := &models.Workflow{
		OrganizationID: orgID,
		Name:           "Ciclo de Vida do Orçamento",
		Description:    stringPtr("Workflow padrão para gestão de orçamentos de construção"),
		Module:         models.WorkflowModuleConstruction,
		EntityType:     models.WorkflowEntityBudget,
		IsActive:       true,
		IsDefault:      true,
	}
	if err := s.CreateWorkflow(ctx, workflow); err != nil {
		return nil, fmt.Errorf("failed to create workflow: %w", err)
	}

	// Create states matching Budget status enum
	states := []struct {
		name        string
		displayName string
		description string
		stateType   models.StateType
		color       string
		position    int
	}{
		{"draft", "Rascunho", "Orçamento em preparação", models.StateTypeInitial, "#6B7280", 0},
		{"sent", "Enviado", "Orçamento enviado ao cliente", models.StateTypeIntermediate, "#3B82F6", 1},
		{"approved", "Aprovado", "Orçamento aprovado pelo cliente", models.StateTypeFinal, "#10B981", 2},
		{"rejected", "Rejeitado", "Orçamento rejeitado pelo cliente", models.StateTypeFinal, "#EF4444", 3},
		{"expired", "Expirado", "Orçamento expirou sem resposta", models.StateTypeFinal, "#F59E0B", 4},
	}

	stateMap := make(map[string]uuid.UUID)
	for _, st := range states {
		state := &models.WorkflowState{
			WorkflowID:  workflow.ID,
			Name:        st.name,
			DisplayName: st.displayName,
			Description: stringPtr(st.description),
			StateType:   st.stateType,
			Color:       stringPtr(st.color),
			Position:    st.position,
		}
		if err := s.CreateState(ctx, state); err != nil {
			return nil, fmt.Errorf("failed to create state %s: %w", st.name, err)
		}
		stateMap[st.name] = state.ID
	}

	// Create transitions
	transitions := []struct {
		from, to, name       string
		requiresConfirmation bool
	}{
		{"draft", "sent", "Enviar ao Cliente", false},
		{"sent", "approved", "Cliente Aprova", false},
		{"sent", "rejected", "Cliente Rejeita", false},
		{"sent", "expired", "Expirar", false},
		{"rejected", "draft", "Voltar a Rascunho", false},
	}

	for _, tr := range transitions {
		transition := &models.WorkflowTransition{
			WorkflowID:           workflow.ID,
			FromStateID:          stateMap[tr.from],
			ToStateID:            stateMap[tr.to],
			Name:                 tr.name,
			RequiresConfirmation: tr.requiresConfirmation,
		}
		if err := s.CreateTransition(ctx, transition); err != nil {
			return nil, fmt.Errorf("failed to create transition %s: %w", tr.name, err)
		}
	}

	// Create trigger for when budget is sent (on_enter "sent" state)
	sentStateID := stateMap["sent"]
	triggerSent := &models.WorkflowTrigger{
		WorkflowID:  workflow.ID,
		StateID:     &sentStateID,
		TriggerType: models.TriggerTypeOnEnter,
		IsActive:    true,
	}
	if err := s.CreateTrigger(ctx, triggerSent); err != nil {
		return nil, fmt.Errorf("failed to create sent trigger: %w", err)
	}

	// Create send_email action for the sent trigger
	actionSent := &models.WorkflowAction{
		TriggerID:   triggerSent.ID,
		ActionType:  models.ActionTypeSendEmail,
		ActionOrder: 0,
		IsActive:    true,
		ActionConfig: json.RawMessage(`{
			"subject": "Novo orçamento disponível - {{budget_number}}",
			"to_field": "client_email"
		}`),
	}
	if err := s.CreateAction(ctx, actionSent); err != nil {
		return nil, fmt.Errorf("failed to create sent action: %w", err)
	}

	// Create trigger for when budget is approved (on_enter "approved" state)
	approvedStateID := stateMap["approved"]
	triggerApproved := &models.WorkflowTrigger{
		WorkflowID:  workflow.ID,
		StateID:     &approvedStateID,
		TriggerType: models.TriggerTypeOnEnter,
		IsActive:    true,
	}
	if err := s.CreateTrigger(ctx, triggerApproved); err != nil {
		return nil, fmt.Errorf("failed to create approved trigger: %w", err)
	}

	// Create send_email action for the approved trigger (notify organization)
	actionApproved := &models.WorkflowAction{
		TriggerID:   triggerApproved.ID,
		ActionType:  models.ActionTypeSendEmail,
		ActionOrder: 0,
		IsActive:    true,
		ActionConfig: json.RawMessage(`{
			"subject": "Orçamento {{budget_number}} foi aprovado!",
			"to_field": "organization_email"
		}`),
	}
	if err := s.CreateAction(ctx, actionApproved); err != nil {
		return nil, fmt.Errorf("failed to create approved action: %w", err)
	}

	return s.GetWorkflowByID(ctx, workflow.ID, orgID)
}

// CreateDefaultProjectWorkflow creates the default workflow for project lifecycle
func (s *WorkflowService) CreateDefaultProjectWorkflow(ctx context.Context, orgID uuid.UUID) (*models.Workflow, error) {
	// Check if default workflow already exists
	existing, err := s.GetDefaultWorkflow(ctx, orgID, models.WorkflowModuleConstruction, models.WorkflowEntityProject)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	// Create the workflow
	workflow := &models.Workflow{
		OrganizationID: orgID,
		Name:           "Ciclo de Vida do Projeto",
		Description:    stringPtr("Workflow padrão para gestão de projetos de construção"),
		Module:         models.WorkflowModuleConstruction,
		EntityType:     models.WorkflowEntityProject,
		IsActive:       true,
		IsDefault:      true,
	}
	if err := s.CreateWorkflow(ctx, workflow); err != nil {
		return nil, fmt.Errorf("failed to create workflow: %w", err)
	}

	// Create states matching Project status enum
	states := []struct {
		name        string
		displayName string
		description string
		stateType   models.StateType
		color       string
		position    int
	}{
		{"in_progress", "Em Progresso", "Projeto em execução", models.StateTypeInitial, "#3B82F6", 0},
		{"on_hold", "Em Espera", "Projeto pausado", models.StateTypeIntermediate, "#F59E0B", 1},
		{"completed", "Concluído", "Projeto finalizado", models.StateTypeFinal, "#10B981", 2},
		{"cancelled", "Cancelado", "Projeto cancelado", models.StateTypeFinal, "#EF4444", 3},
	}

	stateMap := make(map[string]uuid.UUID)
	for _, st := range states {
		state := &models.WorkflowState{
			WorkflowID:  workflow.ID,
			Name:        st.name,
			DisplayName: st.displayName,
			Description: stringPtr(st.description),
			StateType:   st.stateType,
			Color:       stringPtr(st.color),
			Position:    st.position,
		}
		if err := s.CreateState(ctx, state); err != nil {
			return nil, fmt.Errorf("failed to create state %s: %w", st.name, err)
		}
		stateMap[st.name] = state.ID
	}

	// Create transitions
	transitions := []struct {
		from, to, name       string
		requiresConfirmation bool
	}{
		{"in_progress", "on_hold", "Pausar Projeto", false},
		{"in_progress", "completed", "Concluir Projeto", true},
		{"in_progress", "cancelled", "Cancelar Projeto", true},
		{"on_hold", "in_progress", "Retomar Projeto", false},
		{"on_hold", "cancelled", "Cancelar Projeto", true},
	}

	for _, tr := range transitions {
		transition := &models.WorkflowTransition{
			WorkflowID:           workflow.ID,
			FromStateID:          stateMap[tr.from],
			ToStateID:            stateMap[tr.to],
			Name:                 tr.name,
			RequiresConfirmation: tr.requiresConfirmation,
		}
		if err := s.CreateTransition(ctx, transition); err != nil {
			return nil, fmt.Errorf("failed to create transition %s: %w", tr.name, err)
		}
	}

	// Create trigger for when project is completed (on_enter "completed" state)
	completedStateID := stateMap["completed"]
	triggerCompleted := &models.WorkflowTrigger{
		WorkflowID:  workflow.ID,
		StateID:     &completedStateID,
		TriggerType: models.TriggerTypeOnEnter,
		IsActive:    true,
	}
	if err := s.CreateTrigger(ctx, triggerCompleted); err != nil {
		return nil, fmt.Errorf("failed to create completed trigger: %w", err)
	}

	// Create send_email action for the completed trigger
	actionCompleted := &models.WorkflowAction{
		TriggerID:   triggerCompleted.ID,
		ActionType:  models.ActionTypeSendEmail,
		ActionOrder: 0,
		IsActive:    true,
		ActionConfig: json.RawMessage(`{
			"subject": "Projeto {{project_name}} foi concluído!",
			"to_field": "client_email"
		}`),
	}
	if err := s.CreateAction(ctx, actionCompleted); err != nil {
		return nil, fmt.Errorf("failed to create completed action: %w", err)
	}

	return s.GetWorkflowByID(ctx, workflow.ID, orgID)
}

// CreateDefaultTemplates creates default message templates for a module
func (s *WorkflowService) CreateDefaultTemplates(ctx context.Context, orgID uuid.UUID, module string) error {
	var templates []struct {
		name    string
		channel models.MessageChannel
		subject string
		body    string
		vars    []models.TemplateVariable
	}

	switch module {
	case "construction":
		templates = []struct {
			name    string
			channel models.MessageChannel
			subject string
			body    string
			vars    []models.TemplateVariable
		}{
			{
				name:    "Orçamento Enviado",
				channel: models.MessageChannelEmail,
				subject: "Novo Orçamento - {{budget_number}}",
				body: `Caro(a) {{client_name}},

Enviamos em anexo o orçamento {{budget_number}} para o projeto "{{project_name}}".

Valor Total: {{budget_total}}€

Para visualizar ou aprovar o orçamento, aceda ao seguinte link:
{{budget_link}}

Ficamos ao dispor para qualquer esclarecimento.

Com os melhores cumprimentos,
{{organization_name}}`,
				vars: []models.TemplateVariable{
					{Name: "client_name", Description: "Nome do cliente"},
					{Name: "budget_number", Description: "Número do orçamento"},
					{Name: "project_name", Description: "Nome do projeto"},
					{Name: "budget_total", Description: "Valor total do orçamento"},
					{Name: "budget_link", Description: "Link para visualizar o orçamento"},
					{Name: "organization_name", Description: "Nome da organização"},
				},
			},
			{
				name:    "Orçamento Aprovado",
				channel: models.MessageChannelEmail,
				subject: "Orçamento {{budget_number}} Aprovado!",
				body: `O orçamento {{budget_number}} para o cliente {{client_name}} foi aprovado!

Projeto: {{project_name}}
Valor: {{budget_total}}€

O projeto pode agora ser iniciado.

{{organization_name}}`,
				vars: []models.TemplateVariable{
					{Name: "client_name", Description: "Nome do cliente"},
					{Name: "budget_number", Description: "Número do orçamento"},
					{Name: "project_name", Description: "Nome do projeto"},
					{Name: "budget_total", Description: "Valor total do orçamento"},
					{Name: "organization_name", Description: "Nome da organização"},
				},
			},
			{
				name:    "Orçamento Rejeitado",
				channel: models.MessageChannelEmail,
				subject: "Orçamento {{budget_number}} Rejeitado",
				body: `O orçamento {{budget_number}} para o cliente {{client_name}} foi rejeitado.

Projeto: {{project_name}}
Valor: {{budget_total}}€

Poderá ser necessário rever o orçamento e reenviar ao cliente.

{{organization_name}}`,
				vars: []models.TemplateVariable{
					{Name: "client_name", Description: "Nome do cliente"},
					{Name: "budget_number", Description: "Número do orçamento"},
					{Name: "project_name", Description: "Nome do projeto"},
					{Name: "budget_total", Description: "Valor total do orçamento"},
					{Name: "organization_name", Description: "Nome da organização"},
				},
			},
			{
				name:    "Projeto Concluído",
				channel: models.MessageChannelEmail,
				subject: "Projeto {{project_name}} Concluído",
				body: `Caro(a) {{client_name}},

Temos o prazer de informar que o projeto "{{project_name}}" foi concluído com sucesso!

Agradecemos a sua confiança e estamos ao dispor para futuros projetos.

Com os melhores cumprimentos,
{{organization_name}}`,
				vars: []models.TemplateVariable{
					{Name: "client_name", Description: "Nome do cliente"},
					{Name: "project_name", Description: "Nome do projeto"},
					{Name: "organization_name", Description: "Nome da organização"},
				},
			},
		}

	case "appointments":
		templates = []struct {
			name    string
			channel models.MessageChannel
			subject string
			body    string
			vars    []models.TemplateVariable
		}{
			{
				name:    "Lembrete 24h",
				channel: models.MessageChannelWhatsApp,
				subject: "",
				body: `Olá {{patient_name}}! 👋

Lembramos que tem uma consulta agendada para amanhã:

📅 Data: {{session_date}}
🕐 Hora: {{session_time}}
👤 Terapeuta: {{therapist_name}}

Por favor, confirme a sua presença respondendo a esta mensagem.

{{organization_name}}`,
				vars: []models.TemplateVariable{
					{Name: "patient_name", Description: "Nome do paciente"},
					{Name: "session_date", Description: "Data da sessão"},
					{Name: "session_time", Description: "Hora da sessão"},
					{Name: "therapist_name", Description: "Nome do terapeuta"},
					{Name: "organization_name", Description: "Nome da organização"},
				},
			},
			{
				name:    "Lembrete 2h",
				channel: models.MessageChannelWhatsApp,
				subject: "",
				body: `Olá {{patient_name}}! 👋

A sua consulta é daqui a 2 horas:

🕐 {{session_time}}
👤 {{therapist_name}}

Esperamos por si!

{{organization_name}}`,
				vars: []models.TemplateVariable{
					{Name: "patient_name", Description: "Nome do paciente"},
					{Name: "session_time", Description: "Hora da sessão"},
					{Name: "therapist_name", Description: "Nome do terapeuta"},
					{Name: "organization_name", Description: "Nome da organização"},
				},
			},
			{
				name:    "Sessão Confirmada",
				channel: models.MessageChannelWhatsApp,
				subject: "",
				body: `Olá {{patient_name}}! ✅

A sua consulta está confirmada:

📅 Data: {{session_date}}
🕐 Hora: {{session_time}}
👤 Terapeuta: {{therapist_name}}

Até breve!
{{organization_name}}`,
				vars: []models.TemplateVariable{
					{Name: "patient_name", Description: "Nome do paciente"},
					{Name: "session_date", Description: "Data da sessão"},
					{Name: "session_time", Description: "Hora da sessão"},
					{Name: "therapist_name", Description: "Nome do terapeuta"},
					{Name: "organization_name", Description: "Nome da organização"},
				},
			},
			{
				name:    "Lembrete Pagamento",
				channel: models.MessageChannelWhatsApp,
				subject: "",
				body: `Olá {{patient_name}}! 👋

Gostaríamos de lembrar que tem sessões pendentes de pagamento no valor de {{amount}}€.

Por favor, regularize o pagamento na próxima consulta ou contacte-nos para mais informações.

Obrigado,
{{organization_name}}`,
				vars: []models.TemplateVariable{
					{Name: "patient_name", Description: "Nome do paciente"},
					{Name: "amount", Description: "Valor em dívida"},
					{Name: "organization_name", Description: "Nome da organização"},
				},
			},
			{
				name:    "Sessão Cancelada",
				channel: models.MessageChannelWhatsApp,
				subject: "",
				body: `Olá {{patient_name}},

A sua consulta do dia {{session_date}} às {{session_time}} foi cancelada.

Para reagendar, por favor contacte-nos.

{{organization_name}}`,
				vars: []models.TemplateVariable{
					{Name: "patient_name", Description: "Nome do paciente"},
					{Name: "session_date", Description: "Data da sessão"},
					{Name: "session_time", Description: "Hora da sessão"},
					{Name: "organization_name", Description: "Nome da organização"},
				},
			},
		}

	default:
		return fmt.Errorf("unknown module: %s", module)
	}

	// Create templates (skip if already exists)
	for _, t := range templates {
		// Check if template already exists
		var exists bool
		err := s.db.Pool.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM message_templates
				WHERE organization_id = $1 AND name = $2 AND channel = $3
			)
		`, orgID, t.name, t.channel).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check template existence: %w", err)
		}
		if exists {
			continue
		}

		varsJSON, err := json.Marshal(t.vars)
		if err != nil {
			return fmt.Errorf("failed to marshal variables: %w", err)
		}

		var subject *string
		if t.subject != "" {
			subject = &t.subject
		}

		template := &models.MessageTemplate{
			OrganizationID: orgID,
			Name:           t.name,
			Channel:        t.channel,
			Subject:        subject,
			Body:           t.body,
			Variables:      varsJSON,
			IsActive:       true,
		}
		if err := s.CreateTemplate(ctx, template); err != nil {
			return fmt.Errorf("failed to create template %s: %w", t.name, err)
		}
	}

	return nil
}

// Helper function for string pointers
func stringPtr(s string) *string {
	return &s
}

// ============ Execution Log Queries ============

// ExecutionLogFilters contains filters for querying execution logs
type ExecutionLogFilters struct {
	WorkflowID *uuid.UUID
	EntityType string
	EntityID   *uuid.UUID
	EventType  string
	Limit      int
	Offset     int
}

// GetExecutionLogs returns execution logs for an organization with optional filters
func (s *WorkflowService) GetExecutionLogs(ctx context.Context, orgID uuid.UUID, filters ExecutionLogFilters) ([]*models.WorkflowExecutionLog, int, error) {
	// Build query
	baseQuery := `
		FROM workflow_execution_log
		WHERE organization_id = $1`
	args := []interface{}{orgID}
	argNum := 2

	if filters.WorkflowID != nil {
		baseQuery += fmt.Sprintf(" AND workflow_id = $%d", argNum)
		args = append(args, *filters.WorkflowID)
		argNum++
	}
	if filters.EntityType != "" {
		baseQuery += fmt.Sprintf(" AND entity_type = $%d", argNum)
		args = append(args, filters.EntityType)
		argNum++
	}
	if filters.EntityID != nil {
		baseQuery += fmt.Sprintf(" AND entity_id = $%d", argNum)
		args = append(args, *filters.EntityID)
		argNum++
	}
	if filters.EventType != "" {
		baseQuery += fmt.Sprintf(" AND event_type = $%d", argNum)
		args = append(args, filters.EventType)
		argNum++
	}

	// Get total count
	var total int
	countQuery := "SELECT COUNT(*) " + baseQuery
	if err := s.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count execution logs: %w", err)
	}

	// Get logs with pagination
	limit := filters.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filters.Offset
	if offset < 0 {
		offset = 0
	}

	selectQuery := `
		SELECT id, organization_id, workflow_id, entity_type, entity_id,
		       trigger_id, action_id, event_type, from_state, to_state, details, created_at
	` + baseQuery + fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argNum, argNum+1)
	args = append(args, limit, offset)

	rows, err := s.db.Pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query execution logs: %w", err)
	}
	defer rows.Close()

	var logs []*models.WorkflowExecutionLog
	for rows.Next() {
		var log models.WorkflowExecutionLog
		err := rows.Scan(
			&log.ID, &log.OrganizationID, &log.WorkflowID, &log.EntityType, &log.EntityID,
			&log.TriggerID, &log.ActionID, &log.EventType, &log.FromState, &log.ToState,
			&log.Details, &log.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan execution log: %w", err)
		}
		logs = append(logs, &log)
	}

	return logs, total, nil
}

// GetEntityExecutionLogs returns execution logs for a specific entity
func (s *WorkflowService) GetEntityExecutionLogs(ctx context.Context, orgID uuid.UUID, entityType string, entityID uuid.UUID, limit int) ([]*models.WorkflowExecutionLog, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := s.db.Pool.Query(ctx, `
		SELECT id, organization_id, workflow_id, entity_type, entity_id,
		       trigger_id, action_id, event_type, from_state, to_state, details, created_at
		FROM workflow_execution_log
		WHERE organization_id = $1 AND entity_type = $2 AND entity_id = $3
		ORDER BY created_at DESC
		LIMIT $4
	`, orgID, entityType, entityID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query entity execution logs: %w", err)
	}
	defer rows.Close()

	var logs []*models.WorkflowExecutionLog
	for rows.Next() {
		var log models.WorkflowExecutionLog
		err := rows.Scan(
			&log.ID, &log.OrganizationID, &log.WorkflowID, &log.EntityType, &log.EntityID,
			&log.TriggerID, &log.ActionID, &log.EventType, &log.FromState, &log.ToState,
			&log.Details, &log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan execution log: %w", err)
		}
		logs = append(logs, &log)
	}

	return logs, nil
}

// GetScheduledJobs returns pending scheduled jobs for an organization
func (s *WorkflowService) GetScheduledJobs(ctx context.Context, orgID uuid.UUID, status string, limit int) ([]*models.ScheduledJob, error) {
	if limit <= 0 {
		limit = 50
	}
	if status == "" {
		status = "pending"
	}

	rows, err := s.db.Pool.Query(ctx, `
		SELECT id, organization_id, trigger_id, entity_type, entity_id,
		       scheduled_for, status, attempts, last_error, created_at, processed_at
		FROM scheduled_jobs
		WHERE organization_id = $1 AND status = $2
		ORDER BY scheduled_for ASC
		LIMIT $3
	`, orgID, status, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query scheduled jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*models.ScheduledJob
	for rows.Next() {
		var job models.ScheduledJob
		err := rows.Scan(
			&job.ID, &job.OrganizationID, &job.TriggerID, &job.EntityType, &job.EntityID,
			&job.ScheduledFor, &job.Status, &job.Attempts, &job.LastError, &job.CreatedAt, &job.ProcessedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan scheduled job: %w", err)
		}
		jobs = append(jobs, &job)
	}

	return jobs, nil
}

// ============ Workflow Testing ============

// WorkflowTestResult represents the result of testing a workflow trigger
type WorkflowTestResult struct {
	Trigger     *models.WorkflowTrigger  `json:"trigger"`
	State       *models.WorkflowState    `json:"state,omitempty"`
	Transition  *models.WorkflowTransition `json:"transition,omitempty"`
	Actions     []*ActionTestResult      `json:"actions"`
	SampleData  map[string]interface{}   `json:"sample_data"`
}

// ActionTestResult represents the result of testing a single action
type ActionTestResult struct {
	Action          *models.WorkflowAction `json:"action"`
	ActionType      string                 `json:"action_type"`
	Template        *models.MessageTemplate `json:"template,omitempty"`
	RenderedSubject string                 `json:"rendered_subject,omitempty"`
	RenderedBody    string                 `json:"rendered_body"`
	Recipient       string                 `json:"recipient,omitempty"`
}

// TestWorkflowTrigger simulates a trigger execution with sample data
func (s *WorkflowService) TestWorkflowTrigger(ctx context.Context, orgID, workflowID, triggerID uuid.UUID) (*WorkflowTestResult, error) {
	// Get the workflow
	workflow, err := s.GetWorkflowByID(ctx, workflowID, orgID)
	if err != nil {
		return nil, err
	}

	// Find the trigger
	var trigger *models.WorkflowTrigger
	for i := range workflow.Triggers {
		if workflow.Triggers[i].ID == triggerID {
			trigger = &workflow.Triggers[i]
			break
		}
	}
	if trigger == nil {
		return nil, errors.New("trigger not found")
	}

	// Load full trigger with actions
	fullTrigger, err := s.GetTriggerByID(ctx, triggerID)
	if err != nil {
		return nil, err
	}

	// Get sample data based on entity type
	sampleData := GetSampleDataForEntityType(string(workflow.EntityType))

	result := &WorkflowTestResult{
		Trigger:    fullTrigger,
		SampleData: sampleData,
		Actions:    make([]*ActionTestResult, 0),
	}

	// Find associated state or transition
	if trigger.StateID != nil {
		for i := range workflow.States {
			if workflow.States[i].ID == *trigger.StateID {
				result.State = &workflow.States[i]
				break
			}
		}
	}
	if trigger.TransitionID != nil {
		for i := range workflow.Transitions {
			if workflow.Transitions[i].ID == *trigger.TransitionID {
				result.Transition = &workflow.Transitions[i]
				break
			}
		}
	}

	// Process each action
	for i := range fullTrigger.Actions {
		action := &fullTrigger.Actions[i]
		actionResult := &ActionTestResult{
			Action:     action,
			ActionType: string(action.ActionType),
		}

		switch action.ActionType {
		case models.ActionTypeSendWhatsApp, models.ActionTypeSendEmail:
			// Get template if specified
			if action.TemplateID != nil {
				template, err := s.GetTemplateByID(ctx, *action.TemplateID, orgID)
				if err == nil {
					actionResult.Template = template
					// Render with sample data
					actionResult.RenderedBody = renderTemplateString(template.Body, sampleData)
					if template.Subject != nil {
						actionResult.RenderedSubject = renderTemplateString(*template.Subject, sampleData)
					}
				}
			} else if action.ActionConfig != nil {
				// Use inline config
				config := parseActionConfigJSON(action.ActionConfig)
				if subject, ok := config["subject"].(string); ok {
					actionResult.RenderedSubject = renderTemplateString(subject, sampleData)
				}
				if body, ok := config["body"].(string); ok {
					actionResult.RenderedBody = renderTemplateString(body, sampleData)
				}
			}

			// Determine recipient
			if action.ActionType == models.ActionTypeSendEmail {
				if config := parseActionConfigJSON(action.ActionConfig); config != nil {
					if toField, ok := config["to_field"].(string); ok {
						if email, ok := sampleData[toField].(string); ok {
							actionResult.Recipient = email
						} else {
							actionResult.Recipient = toField + " (campo não encontrado)"
						}
					}
				}
			} else {
				// WhatsApp - use patient_phone or client_phone
				if phone, ok := sampleData["patient_phone"].(string); ok {
					actionResult.Recipient = phone
				} else if phone, ok := sampleData["client_phone"].(string); ok {
					actionResult.Recipient = phone
				}
			}

		case models.ActionTypeUpdateField:
			if action.ActionConfig != nil {
				config := parseActionConfigJSON(action.ActionConfig)
				if field, ok := config["field"].(string); ok {
					if value, ok := config["value"].(string); ok {
						actionResult.RenderedBody = fmt.Sprintf("Campo '%s' será atualizado para '%s'", field, value)
					}
				}
			}

		case models.ActionTypeCreateTask:
			if action.ActionConfig != nil {
				config := parseActionConfigJSON(action.ActionConfig)
				if title, ok := config["title"].(string); ok {
					actionResult.RenderedBody = renderTemplateString(title, sampleData)
				}
			}
		}

		result.Actions = append(result.Actions, actionResult)
	}

	return result, nil
}

// GetSampleDataForEntityType returns sample data for testing
func GetSampleDataForEntityType(entityType string) map[string]interface{} {
	switch entityType {
	case "session":
		return map[string]interface{}{
			"patient_name":      "João Silva",
			"patient_phone":     "+351912345678",
			"patient_email":     "joao.silva@email.com",
			"therapist_name":    "Dr. Maria Santos",
			"session_date":      "15/01/2025",
			"session_time":      "14:30",
			"session_type":      "Consulta Regular",
			"amount":            "50.00",
			"organization_name": "Clínica Exemplo",
			"organization_email": "clinica@exemplo.com",
		}
	case "budget":
		return map[string]interface{}{
			"client_name":        "Manuel Costa",
			"client_email":       "manuel.costa@email.com",
			"client_phone":       "+351923456789",
			"project_name":       "Remodelação Cozinha",
			"budget_number":      "ORC-2025-001",
			"budget_total":       "15000.00",
			"budget_link":        "https://app.controlewise.pt/budgets/123",
			"approval_link":      "https://app.controlewise.pt/budgets/123/approve",
			"organization_name":  "Construções ABC",
			"organization_email": "info@construcoes-abc.pt",
		}
	case "project":
		return map[string]interface{}{
			"client_name":        "Ana Ferreira",
			"client_email":       "ana.ferreira@email.com",
			"client_phone":       "+351934567890",
			"project_name":       "Construção Moradia",
			"project_number":     "PRJ-2025-001",
			"project_status":     "Em Curso",
			"organization_name":  "Construções ABC",
			"organization_email": "info@construcoes-abc.pt",
		}
	default:
		return map[string]interface{}{
			"name":  "Cliente Exemplo",
			"email": "cliente@email.com",
			"phone": "+351900000000",
		}
	}
}

// renderTemplateString renders a template string with data
func renderTemplateString(template string, data map[string]interface{}) string {
	result := template
	for key, value := range data {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	return result
}

// parseActionConfigJSON parses action config JSON
func parseActionConfigJSON(config json.RawMessage) map[string]interface{} {
	if config == nil {
		return nil
	}
	var result map[string]interface{}
	if err := json.Unmarshal(config, &result); err != nil {
		return nil
	}
	return result
}
