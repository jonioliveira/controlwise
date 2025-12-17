package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/controlewise/backend/internal/database"
	"github.com/controlewise/backend/internal/models"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// Engine handles workflow execution
type Engine struct {
	db        *database.DB
	client    *asynq.Client
	scheduler *Scheduler
	executor  *Executor
}

// NewEngine creates a new workflow engine
func NewEngine(db *database.DB, client *asynq.Client) *Engine {
	e := &Engine{
		db:     db,
		client: client,
	}
	e.scheduler = NewScheduler(db, client)
	e.executor = NewExecutor(db)
	return e
}

// OnStateEnter is called when an entity enters a state
// It fires on_enter triggers and schedules time-based triggers
func (e *Engine) OnStateEnter(ctx context.Context, orgID uuid.UUID, workflow *models.Workflow, stateName string, entityType string, entityID uuid.UUID, entityData map[string]interface{}) error {
	log.Printf("[WorkflowEngine] OnStateEnter: workflow=%s, state=%s, entity=%s/%s", workflow.ID, stateName, entityType, entityID)

	// Find the state
	var state *models.WorkflowState
	for i := range workflow.States {
		if workflow.States[i].Name == stateName {
			state = &workflow.States[i]
			break
		}
	}
	if state == nil {
		return fmt.Errorf("state %s not found in workflow", stateName)
	}

	// Log state entry
	if err := e.logEvent(ctx, orgID, workflow.ID, entityType, entityID, models.EventTypeStateChange, nil, &stateName, nil); err != nil {
		log.Printf("[WorkflowEngine] Failed to log state entry: %v", err)
	}

	// Find and execute on_enter triggers for this state
	for _, trigger := range workflow.Triggers {
		if trigger.StateID == nil || *trigger.StateID != state.ID {
			continue
		}
		if !trigger.IsActive {
			continue
		}

		switch trigger.TriggerType {
		case models.TriggerTypeOnEnter:
			// Execute immediately
			if err := e.executeTrigger(ctx, orgID, workflow, &trigger, entityType, entityID, entityData); err != nil {
				log.Printf("[WorkflowEngine] Failed to execute on_enter trigger %s: %v", trigger.ID, err)
			}
		case models.TriggerTypeTimeBefore, models.TriggerTypeTimeAfter:
			// Schedule for later
			if err := e.scheduler.ScheduleTimeTrigger(ctx, orgID, &trigger, entityType, entityID, entityData); err != nil {
				log.Printf("[WorkflowEngine] Failed to schedule time trigger %s: %v", trigger.ID, err)
			}
		case models.TriggerTypeRecurring:
			// Schedule recurring job
			if err := e.scheduler.ScheduleRecurringTrigger(ctx, orgID, &trigger, entityType, entityID); err != nil {
				log.Printf("[WorkflowEngine] Failed to schedule recurring trigger %s: %v", trigger.ID, err)
			}
		}
	}

	return nil
}

// OnStateExit is called when an entity exits a state
// It fires on_exit triggers and cancels pending scheduled jobs
func (e *Engine) OnStateExit(ctx context.Context, orgID uuid.UUID, workflow *models.Workflow, stateName string, entityType string, entityID uuid.UUID) error {
	log.Printf("[WorkflowEngine] OnStateExit: workflow=%s, state=%s, entity=%s/%s", workflow.ID, stateName, entityType, entityID)

	// Find the state
	var state *models.WorkflowState
	for i := range workflow.States {
		if workflow.States[i].Name == stateName {
			state = &workflow.States[i]
			break
		}
	}
	if state == nil {
		return fmt.Errorf("state %s not found in workflow", stateName)
	}

	// Cancel any pending scheduled jobs for this entity
	if err := e.scheduler.CancelPendingJobs(ctx, entityType, entityID); err != nil {
		log.Printf("[WorkflowEngine] Failed to cancel pending jobs: %v", err)
	}

	// Find and execute on_exit triggers for this state
	for _, trigger := range workflow.Triggers {
		if trigger.StateID == nil || *trigger.StateID != state.ID {
			continue
		}
		if !trigger.IsActive {
			continue
		}

		if trigger.TriggerType == models.TriggerTypeOnExit {
			if err := e.executeTrigger(ctx, orgID, workflow, &trigger, entityType, entityID, nil); err != nil {
				log.Printf("[WorkflowEngine] Failed to execute on_exit trigger %s: %v", trigger.ID, err)
			}
		}
	}

	return nil
}

// TransitionEntity moves an entity from one state to another
func (e *Engine) TransitionEntity(ctx context.Context, orgID uuid.UUID, workflow *models.Workflow, fromState, toState string, entityType string, entityID uuid.UUID, entityData map[string]interface{}) error {
	log.Printf("[WorkflowEngine] TransitionEntity: workflow=%s, %s -> %s, entity=%s/%s", workflow.ID, fromState, toState, entityType, entityID)

	// Exit the current state
	if fromState != "" {
		if err := e.OnStateExit(ctx, orgID, workflow, fromState, entityType, entityID); err != nil {
			return fmt.Errorf("failed to exit state %s: %w", fromState, err)
		}
	}

	// Log state change
	if err := e.logEvent(ctx, orgID, workflow.ID, entityType, entityID, models.EventTypeStateChange, &fromState, &toState, nil); err != nil {
		log.Printf("[WorkflowEngine] Failed to log state change: %v", err)
	}

	// Enter the new state
	if err := e.OnStateEnter(ctx, orgID, workflow, toState, entityType, entityID, entityData); err != nil {
		return fmt.Errorf("failed to enter state %s: %w", toState, err)
	}

	return nil
}

// executeTrigger executes a trigger and all its actions
func (e *Engine) executeTrigger(ctx context.Context, orgID uuid.UUID, workflow *models.Workflow, trigger *models.WorkflowTrigger, entityType string, entityID uuid.UUID, entityData map[string]interface{}) error {
	log.Printf("[WorkflowEngine] Executing trigger %s (type=%s)", trigger.ID, trigger.TriggerType)

	// Log trigger fired
	if err := e.logEvent(ctx, orgID, workflow.ID, entityType, entityID, models.EventTypeTriggerFired, nil, nil, map[string]interface{}{
		"trigger_id":   trigger.ID,
		"trigger_type": trigger.TriggerType,
	}); err != nil {
		log.Printf("[WorkflowEngine] Failed to log trigger fired: %v", err)
	}

	// Execute each action in order
	for _, action := range trigger.Actions {
		if !action.IsActive {
			continue
		}

		if err := e.executor.ExecuteAction(ctx, orgID, &action, entityType, entityID, entityData); err != nil {
			// Log failure
			e.logEvent(ctx, orgID, workflow.ID, entityType, entityID, models.EventTypeActionFailed, nil, nil, map[string]interface{}{
				"action_id":   action.ID,
				"action_type": action.ActionType,
				"error":       err.Error(),
			})
			log.Printf("[WorkflowEngine] Action %s failed: %v", action.ID, err)
			// Continue with other actions
			continue
		}

		// Log success
		e.logEvent(ctx, orgID, workflow.ID, entityType, entityID, models.EventTypeActionExecuted, nil, nil, map[string]interface{}{
			"action_id":   action.ID,
			"action_type": action.ActionType,
		})
	}

	return nil
}

// ExecuteTriggerByID executes a trigger by its ID (used by job handlers)
func (e *Engine) ExecuteTriggerByID(ctx context.Context, orgID, triggerID uuid.UUID, entityType string, entityID uuid.UUID) error {
	// Get trigger with workflow
	trigger, workflow, err := e.getTriggerWithWorkflow(ctx, triggerID, orgID)
	if err != nil {
		return fmt.Errorf("failed to get trigger: %w", err)
	}

	// Get entity data for template rendering
	entityData, err := e.getEntityData(ctx, orgID, entityType, entityID)
	if err != nil {
		log.Printf("[WorkflowEngine] Failed to get entity data: %v", err)
		// Continue without entity data
	}

	return e.executeTrigger(ctx, orgID, workflow, trigger, entityType, entityID, entityData)
}

// getTriggerWithWorkflow gets a trigger and its parent workflow
func (e *Engine) getTriggerWithWorkflow(ctx context.Context, triggerID, orgID uuid.UUID) (*models.WorkflowTrigger, *models.Workflow, error) {
	var trigger models.WorkflowTrigger
	var workflowID uuid.UUID

	err := e.db.Pool.QueryRow(ctx, `
		SELECT t.id, t.workflow_id, t.state_id, t.transition_id, t.trigger_type,
		       t.time_offset_minutes, t.time_field, t.recurring_cron, t.conditions, t.is_active, t.created_at
		FROM workflow_triggers t
		JOIN workflows w ON w.id = t.workflow_id
		WHERE t.id = $1 AND w.organization_id = $2
	`, triggerID, orgID).Scan(
		&trigger.ID, &workflowID, &trigger.StateID, &trigger.TransitionID, &trigger.TriggerType,
		&trigger.TimeOffsetMinutes, &trigger.TimeField, &trigger.RecurringCron, &trigger.Conditions,
		&trigger.IsActive, &trigger.CreatedAt,
	)
	if err != nil {
		return nil, nil, err
	}

	// Load actions
	rows, err := e.db.Pool.Query(ctx, `
		SELECT id, trigger_id, action_type, action_order, template_id, action_config, is_active, created_at
		FROM workflow_actions
		WHERE trigger_id = $1
		ORDER BY action_order ASC
	`, triggerID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var action models.WorkflowAction
		if err := rows.Scan(
			&action.ID, &action.TriggerID, &action.ActionType, &action.ActionOrder,
			&action.TemplateID, &action.ActionConfig, &action.IsActive, &action.CreatedAt,
		); err != nil {
			return nil, nil, err
		}
		trigger.Actions = append(trigger.Actions, action)
	}

	// Get workflow
	var w models.Workflow
	err = e.db.Pool.QueryRow(ctx, `
		SELECT id, organization_id, name, description, module, entity_type,
		       is_active, is_default, created_at, updated_at
		FROM workflows
		WHERE id = $1
	`, workflowID).Scan(
		&w.ID, &w.OrganizationID, &w.Name, &w.Description, &w.Module, &w.EntityType,
		&w.IsActive, &w.IsDefault, &w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		return nil, nil, err
	}

	return &trigger, &w, nil
}

// getEntityData retrieves entity data for template rendering
func (e *Engine) getEntityData(ctx context.Context, orgID uuid.UUID, entityType string, entityID uuid.UUID) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	switch entityType {
	case "session":
		return e.getSessionData(ctx, orgID, entityID)
	case "budget":
		return e.getBudgetData(ctx, orgID, entityID)
	case "project":
		return e.getProjectData(ctx, orgID, entityID)
	}

	return data, nil
}

// getSessionData retrieves session data with patient and therapist info
func (e *Engine) getSessionData(ctx context.Context, orgID uuid.UUID, sessionID uuid.UUID) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	var patientName, therapistName, sessionType, status string
	var scheduledAt time.Time
	var patientPhone, patientEmail *string

	err := e.db.Pool.QueryRow(ctx, `
		SELECT
			s.scheduled_at,
			s.session_type,
			s.status,
			COALESCE(c.name, '') as patient_name,
			c.phone as patient_phone,
			c.email as patient_email,
			COALESCE(u.name, '') as therapist_name
		FROM sessions s
		LEFT JOIN patients p ON p.id = s.patient_id
		LEFT JOIN clients c ON c.id = p.client_id
		LEFT JOIN users u ON u.id = s.therapist_id
		WHERE s.id = $1 AND s.organization_id = $2
	`, sessionID, orgID).Scan(
		&scheduledAt, &sessionType, &status,
		&patientName, &patientPhone, &patientEmail, &therapistName,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get session data: %w", err)
	}

	data["session_id"] = sessionID.String()
	data["scheduled_at"] = scheduledAt
	data["session_date"] = scheduledAt.Format("02/01/2006")
	data["session_time"] = scheduledAt.Format("15:04")
	data["session_type"] = sessionType
	data["status"] = status
	data["patient_name"] = patientName
	data["therapist_name"] = therapistName

	if patientPhone != nil {
		data["patient_phone"] = *patientPhone
	}
	if patientEmail != nil {
		data["patient_email"] = *patientEmail
	}

	return data, nil
}

// getBudgetData retrieves budget data with client info
func (e *Engine) getBudgetData(ctx context.Context, orgID uuid.UUID, budgetID uuid.UUID) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	var clientName, status, budgetNumber, worksheetTitle string
	var total float64
	var clientEmail, clientPhone *string

	err := e.db.Pool.QueryRow(ctx, `
		SELECT
			b.status,
			b.budget_number,
			b.total,
			w.title as worksheet_title,
			COALESCE(c.name, '') as client_name,
			c.email as client_email,
			c.phone as client_phone
		FROM budgets b
		LEFT JOIN worksheets w ON w.id = b.worksheet_id
		LEFT JOIN clients c ON c.id = w.client_id
		WHERE b.id = $1 AND b.organization_id = $2
	`, budgetID, orgID).Scan(
		&status, &budgetNumber, &total, &worksheetTitle, &clientName, &clientEmail, &clientPhone,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get budget data: %w", err)
	}

	data["budget_id"] = budgetID.String()
	data["budget_number"] = budgetNumber
	data["status"] = status
	data["budget_total"] = fmt.Sprintf("%.2f", total)
	data["project_name"] = worksheetTitle
	data["client_name"] = clientName

	if clientEmail != nil {
		data["client_email"] = *clientEmail
	}
	if clientPhone != nil {
		data["client_phone"] = *clientPhone
	}

	return data, nil
}

// getProjectData retrieves project data
func (e *Engine) getProjectData(ctx context.Context, orgID uuid.UUID, projectID uuid.UUID) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	var clientName, status, projectTitle, projectNumber string
	var clientEmail, clientPhone *string

	err := e.db.Pool.QueryRow(ctx, `
		SELECT
			p.title,
			p.project_number,
			p.status,
			COALESCE(c.name, '') as client_name,
			c.email as client_email,
			c.phone as client_phone
		FROM projects p
		LEFT JOIN budgets b ON b.id = p.budget_id
		LEFT JOIN worksheets w ON w.id = b.worksheet_id
		LEFT JOIN clients c ON c.id = w.client_id
		WHERE p.id = $1 AND p.organization_id = $2
	`, projectID, orgID).Scan(
		&projectTitle, &projectNumber, &status, &clientName, &clientEmail, &clientPhone,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get project data: %w", err)
	}

	data["project_id"] = projectID.String()
	data["project_name"] = projectTitle
	data["project_number"] = projectNumber
	data["status"] = status
	data["client_name"] = clientName

	if clientEmail != nil {
		data["client_email"] = *clientEmail
	}
	if clientPhone != nil {
		data["client_phone"] = *clientPhone
	}

	return data, nil
}

// logEvent logs a workflow execution event
func (e *Engine) logEvent(ctx context.Context, orgID, workflowID uuid.UUID, entityType string, entityID uuid.UUID, eventType models.EventType, fromState, toState *string, details map[string]interface{}) error {
	var detailsJSON []byte
	var err error
	if details != nil {
		detailsJSON, err = json.Marshal(details)
		if err != nil {
			return err
		}
	}

	_, err = e.db.Pool.Exec(ctx, `
		INSERT INTO workflow_execution_log
		(id, organization_id, workflow_id, entity_type, entity_id, event_type, from_state, to_state, details)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, uuid.New(), orgID, workflowID, entityType, entityID, eventType, fromState, toState, detailsJSON)

	return err
}

// GetScheduler returns the scheduler instance
func (e *Engine) GetScheduler() *Scheduler {
	return e.scheduler
}

// GetExecutor returns the executor instance
func (e *Engine) GetExecutor() *Executor {
	return e.executor
}
