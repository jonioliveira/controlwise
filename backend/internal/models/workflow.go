package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// WorkflowModule represents the module a workflow belongs to
type WorkflowModule string

const (
	WorkflowModuleAppointments  WorkflowModule = "appointments"
	WorkflowModuleConstruction  WorkflowModule = "construction"
)

// WorkflowEntityType represents the entity type a workflow manages
type WorkflowEntityType string

const (
	WorkflowEntitySession WorkflowEntityType = "session"
	WorkflowEntityBudget  WorkflowEntityType = "budget"
	WorkflowEntityProject WorkflowEntityType = "project"
)

// Workflow represents a configurable workflow definition
type Workflow struct {
	ID             uuid.UUID          `json:"id" db:"id"`
	OrganizationID uuid.UUID          `json:"organization_id" db:"organization_id"`
	Name           string             `json:"name" db:"name"`
	Description    *string            `json:"description" db:"description"`
	Module         WorkflowModule     `json:"module" db:"module"`
	EntityType     WorkflowEntityType `json:"entity_type" db:"entity_type"`
	IsActive       bool               `json:"is_active" db:"is_active"`
	IsDefault      bool               `json:"is_default" db:"is_default"`
	CreatedAt      time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at" db:"updated_at"`
	// Nested data for full workflow retrieval
	States      []WorkflowState      `json:"states,omitempty" db:"-"`
	Transitions []WorkflowTransition `json:"transitions,omitempty" db:"-"`
	Triggers    []WorkflowTrigger    `json:"triggers,omitempty" db:"-"`
}

// StateType represents the type of workflow state
type StateType string

const (
	StateTypeInitial      StateType = "initial"
	StateTypeIntermediate StateType = "intermediate"
	StateTypeFinal        StateType = "final"
)

// WorkflowState represents a state in a workflow
type WorkflowState struct {
	ID          uuid.UUID `json:"id" db:"id"`
	WorkflowID  uuid.UUID `json:"workflow_id" db:"workflow_id"`
	Name        string    `json:"name" db:"name"`
	DisplayName string    `json:"display_name" db:"display_name"`
	Description *string   `json:"description" db:"description"`
	StateType   StateType `json:"state_type" db:"state_type"`
	Color       *string   `json:"color" db:"color"`
	Position    int       `json:"position" db:"position"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	// Nested data
	Triggers []WorkflowTrigger `json:"triggers,omitempty" db:"-"`
}

// WorkflowTransition represents a transition between states
type WorkflowTransition struct {
	ID                   uuid.UUID `json:"id" db:"id"`
	WorkflowID           uuid.UUID `json:"workflow_id" db:"workflow_id"`
	FromStateID          uuid.UUID `json:"from_state_id" db:"from_state_id"`
	ToStateID            uuid.UUID `json:"to_state_id" db:"to_state_id"`
	Name                 string    `json:"name" db:"name"`
	RequiresConfirmation bool      `json:"requires_confirmation" db:"requires_confirmation"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	// Joined data
	FromStateName string            `json:"from_state_name,omitempty" db:"-"`
	ToStateName   string            `json:"to_state_name,omitempty" db:"-"`
	Triggers      []WorkflowTrigger `json:"triggers,omitempty" db:"-"`
}

// TriggerType represents when a trigger should fire
type TriggerType string

const (
	TriggerTypeOnEnter    TriggerType = "on_enter"
	TriggerTypeOnExit     TriggerType = "on_exit"
	TriggerTypeTimeBefore TriggerType = "time_before"
	TriggerTypeTimeAfter  TriggerType = "time_after"
	TriggerTypeRecurring  TriggerType = "recurring"
)

// WorkflowTrigger represents a trigger that fires actions
type WorkflowTrigger struct {
	ID                uuid.UUID       `json:"id" db:"id"`
	WorkflowID        uuid.UUID       `json:"workflow_id" db:"workflow_id"`
	StateID           *uuid.UUID      `json:"state_id" db:"state_id"`
	TransitionID      *uuid.UUID      `json:"transition_id" db:"transition_id"`
	TriggerType       TriggerType     `json:"trigger_type" db:"trigger_type"`
	TimeOffsetMinutes *int            `json:"time_offset_minutes" db:"time_offset_minutes"`
	TimeField         *string         `json:"time_field" db:"time_field"`
	RecurringCron     *string         `json:"recurring_cron" db:"recurring_cron"`
	Conditions        json.RawMessage `json:"conditions" db:"conditions"`
	IsActive          bool            `json:"is_active" db:"is_active"`
	CreatedAt         time.Time       `json:"created_at" db:"created_at"`
	// Nested data
	Actions []WorkflowAction `json:"actions,omitempty" db:"-"`
}

// ActionType represents the type of action to execute
type ActionType string

const (
	ActionTypeSendWhatsApp ActionType = "send_whatsapp"
	ActionTypeSendEmail    ActionType = "send_email"
	ActionTypeUpdateField  ActionType = "update_field"
	ActionTypeCreateTask   ActionType = "create_task"
)

// WorkflowAction represents an action to execute when a trigger fires
type WorkflowAction struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	TriggerID    uuid.UUID       `json:"trigger_id" db:"trigger_id"`
	ActionType   ActionType      `json:"action_type" db:"action_type"`
	ActionOrder  int             `json:"action_order" db:"action_order"`
	TemplateID   *uuid.UUID      `json:"template_id" db:"template_id"`
	ActionConfig json.RawMessage `json:"action_config" db:"action_config"`
	IsActive     bool            `json:"is_active" db:"is_active"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	// Joined data
	Template *MessageTemplate `json:"template,omitempty" db:"-"`
}

// MessageChannel represents the notification channel
type MessageChannel string

const (
	MessageChannelWhatsApp MessageChannel = "whatsapp"
	MessageChannelEmail    MessageChannel = "email"
)

// MessageTemplate represents a reusable message template
type MessageTemplate struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	OrganizationID uuid.UUID       `json:"organization_id" db:"organization_id"`
	Name           string          `json:"name" db:"name"`
	Channel        MessageChannel  `json:"channel" db:"channel"`
	Subject        *string         `json:"subject" db:"subject"`
	Body           string          `json:"body" db:"body"`
	Variables      json.RawMessage `json:"variables" db:"variables"`
	IsActive       bool            `json:"is_active" db:"is_active"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
}

// TemplateVariable represents a variable available in a template
type TemplateVariable struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// GetVariables parses the variables JSON
func (t *MessageTemplate) GetVariables() ([]TemplateVariable, error) {
	if t.Variables == nil {
		return []TemplateVariable{}, nil
	}
	var vars []TemplateVariable
	if err := json.Unmarshal(t.Variables, &vars); err != nil {
		return nil, err
	}
	return vars, nil
}

// SetVariables serializes variables to JSON
func (t *MessageTemplate) SetVariables(vars []TemplateVariable) error {
	data, err := json.Marshal(vars)
	if err != nil {
		return err
	}
	t.Variables = data
	return nil
}

// JobStatus represents the status of a scheduled job
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
	JobStatusCancelled  JobStatus = "cancelled"
)

// ScheduledJob represents a job scheduled for execution
type ScheduledJob struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	TriggerID      uuid.UUID  `json:"trigger_id" db:"trigger_id"`
	EntityType     string     `json:"entity_type" db:"entity_type"`
	EntityID       uuid.UUID  `json:"entity_id" db:"entity_id"`
	ScheduledFor   time.Time  `json:"scheduled_for" db:"scheduled_for"`
	Status         JobStatus  `json:"status" db:"status"`
	Attempts       int        `json:"attempts" db:"attempts"`
	LastError      *string    `json:"last_error" db:"last_error"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	ProcessedAt    *time.Time `json:"processed_at" db:"processed_at"`
}

// SessionPaymentStatus represents the payment status for a session
type SessionPaymentStatus string

const (
	SessionPaymentStatusUnpaid  SessionPaymentStatus = "unpaid"
	SessionPaymentStatusPartial SessionPaymentStatus = "partial"
	SessionPaymentStatusPaid    SessionPaymentStatus = "paid"
)

// PaymentMethod represents how a payment was made
type PaymentMethod string

const (
	PaymentMethodCash      PaymentMethod = "cash"
	PaymentMethodTransfer  PaymentMethod = "transfer"
	PaymentMethodInsurance PaymentMethod = "insurance"
	PaymentMethodCard      PaymentMethod = "card"
)

// SessionPayment represents payment information for a session
type SessionPayment struct {
	ID                   uuid.UUID             `json:"id" db:"id"`
	SessionID            uuid.UUID             `json:"session_id" db:"session_id"`
	AmountCents          int                   `json:"amount_cents" db:"amount_cents"`
	PaymentStatus        SessionPaymentStatus  `json:"payment_status" db:"payment_status"`
	PaymentMethod        *PaymentMethod `json:"payment_method" db:"payment_method"`
	InsuranceProvider    *string        `json:"insurance_provider" db:"insurance_provider"`
	InsuranceAmountCents *int           `json:"insurance_amount_cents" db:"insurance_amount_cents"`
	DueDate              *time.Time     `json:"due_date" db:"due_date"`
	PaidAt               *time.Time     `json:"paid_at" db:"paid_at"`
	Notes                *string        `json:"notes" db:"notes"`
	CreatedAt            time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at" db:"updated_at"`
}

// SessionPaymentWithDetails includes session information
type SessionPaymentWithDetails struct {
	SessionPayment
	PatientName   string    `json:"patient_name" db:"patient_name"`
	TherapistName string    `json:"therapist_name" db:"therapist_name"`
	ScheduledAt   time.Time `json:"scheduled_at" db:"scheduled_at"`
}

// EventType represents the type of workflow execution event
type EventType string

const (
	EventTypeStateChange    EventType = "state_change"
	EventTypeTriggerFired   EventType = "trigger_fired"
	EventTypeActionExecuted EventType = "action_executed"
	EventTypeActionFailed   EventType = "action_failed"
)

// WorkflowExecutionLog represents a log entry for workflow execution
type WorkflowExecutionLog struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	OrganizationID uuid.UUID       `json:"organization_id" db:"organization_id"`
	WorkflowID     uuid.UUID       `json:"workflow_id" db:"workflow_id"`
	EntityType     string          `json:"entity_type" db:"entity_type"`
	EntityID       uuid.UUID       `json:"entity_id" db:"entity_id"`
	TriggerID      *uuid.UUID      `json:"trigger_id" db:"trigger_id"`
	ActionID       *uuid.UUID      `json:"action_id" db:"action_id"`
	EventType      EventType       `json:"event_type" db:"event_type"`
	FromState      *string         `json:"from_state" db:"from_state"`
	ToState        *string         `json:"to_state" db:"to_state"`
	Details        json.RawMessage `json:"details" db:"details"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
}

// WorkflowWithStats includes workflow statistics
type WorkflowWithStats struct {
	Workflow
	StateCount   int `json:"state_count" db:"state_count"`
	TriggerCount int `json:"trigger_count" db:"trigger_count"`
	ActionCount  int `json:"action_count" db:"action_count"`
}
