package jobs

import (
	"github.com/google/uuid"
)

// Job type constants
const (
	TypeSendNotification = "workflow:send_notification"
	TypeExecuteTrigger   = "workflow:execute_trigger"
	TypeCheckTimeTriggers = "workflow:check_time_triggers"
)

// SendNotificationPayload contains data for sending a notification
type SendNotificationPayload struct {
	OrganizationID uuid.UUID `json:"organization_id"`
	ActionID       uuid.UUID `json:"action_id"`
	EntityType     string    `json:"entity_type"`
	EntityID       uuid.UUID `json:"entity_id"`
	Channel        string    `json:"channel"` // "whatsapp" or "email"
	TemplateID     uuid.UUID `json:"template_id"`
}

// ExecuteTriggerPayload contains data for executing a workflow trigger
type ExecuteTriggerPayload struct {
	OrganizationID uuid.UUID `json:"organization_id"`
	TriggerID      uuid.UUID `json:"trigger_id"`
	EntityType     string    `json:"entity_type"`
	EntityID       uuid.UUID `json:"entity_id"`
}

// CheckTimeTriggersPayload is empty - used for periodic job
type CheckTimeTriggersPayload struct{}
