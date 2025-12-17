package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/controlwise/backend/internal/database"
	"github.com/controlwise/backend/internal/workflow"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// Handlers contains all job handlers
type Handlers struct {
	db     *database.DB
	engine *workflow.Engine
}

// NewHandlers creates a new Handlers instance
func NewHandlers(db *database.DB, engine *workflow.Engine) *Handlers {
	return &Handlers{
		db:     db,
		engine: engine,
	}
}

// HandleSendNotification processes notification sending jobs
func (h *Handlers) HandleSendNotification(ctx context.Context, t *asynq.Task) error {
	var payload SendNotificationPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("[SendNotification] Processing notification for entity %s/%s via %s",
		payload.EntityType, payload.EntityID, payload.Channel)

	// Get the action to execute
	executor := h.engine.GetExecutor()

	// Get action from database
	var actionConfig []byte
	var templateID *string
	err := h.db.Pool.QueryRow(ctx, `
		SELECT template_id, action_config FROM workflow_actions WHERE id = $1
	`, payload.ActionID).Scan(&templateID, &actionConfig)
	if err != nil {
		return fmt.Errorf("failed to get action: %w", err)
	}

	// Get entity data for template rendering
	entityData, err := h.getEntityData(ctx, payload.OrganizationID.String(), payload.EntityType, payload.EntityID)
	if err != nil {
		log.Printf("[SendNotification] Failed to get entity data: %v", err)
		// Continue without entity data
	}

	// Execute based on channel
	switch payload.Channel {
	case "whatsapp":
		if templateID == nil {
			return fmt.Errorf("WhatsApp notification requires a template")
		}
		// Get template and render
		templates := workflow.NewTemplateRenderer(h.db)
		template, err := templates.GetTemplate(ctx, payload.TemplateID, payload.OrganizationID)
		if err != nil {
			return fmt.Errorf("failed to get template: %w", err)
		}

		message, err := templates.RenderTemplate(template.Body, entityData)
		if err != nil {
			return fmt.Errorf("failed to render template: %w", err)
		}

		// Get phone number
		phone, ok := entityData["patient_phone"].(string)
		if !ok || phone == "" {
			phone, ok = entityData["client_phone"].(string)
			if !ok || phone == "" {
				return fmt.Errorf("no phone number available")
			}
		}

		log.Printf("[SendNotification] Would send WhatsApp to %s: %s", phone, message[:min(len(message), 50)])
		// Actual WhatsApp sending would go here via executor.notifySender

	case "email":
		if templateID == nil {
			return fmt.Errorf("Email notification requires a template")
		}
		// Get template and render
		templates := workflow.NewTemplateRenderer(h.db)
		template, err := templates.GetTemplate(ctx, payload.TemplateID, payload.OrganizationID)
		if err != nil {
			return fmt.Errorf("failed to get template: %w", err)
		}

		body, err := templates.RenderTemplate(template.Body, entityData)
		if err != nil {
			return fmt.Errorf("failed to render template: %w", err)
		}

		subject := "Notification"
		if template.Subject != nil {
			subject, _ = templates.RenderTemplate(*template.Subject, entityData)
		}

		// Get email
		email, ok := entityData["patient_email"].(string)
		if !ok || email == "" {
			email, ok = entityData["client_email"].(string)
			if !ok || email == "" {
				return fmt.Errorf("no email address available")
			}
		}

		log.Printf("[SendNotification] Would send email to %s: subject=%s, body_length=%d", email, subject, len(body))
		// Actual email sending would go here via executor.notifySender
	}

	_ = executor // Suppress unused warning until notification sender is implemented

	log.Printf("[SendNotification] Completed notification for entity %s/%s",
		payload.EntityType, payload.EntityID)

	return nil
}

// HandleExecuteTrigger processes workflow trigger execution jobs
func (h *Handlers) HandleExecuteTrigger(ctx context.Context, t *asynq.Task) error {
	var payload ExecuteTriggerPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("[ExecuteTrigger] Processing trigger %s for entity %s/%s",
		payload.TriggerID, payload.EntityType, payload.EntityID)

	// Use the workflow engine to execute the trigger
	err := h.engine.ExecuteTriggerByID(ctx, payload.OrganizationID, payload.TriggerID, payload.EntityType, payload.EntityID)
	if err != nil {
		return fmt.Errorf("failed to execute trigger: %w", err)
	}

	log.Printf("[ExecuteTrigger] Completed trigger %s", payload.TriggerID)

	return nil
}

// HandleCheckTimeTriggers processes periodic time-based trigger checks
func (h *Handlers) HandleCheckTimeTriggers(ctx context.Context, t *asynq.Task) error {
	log.Println("[CheckTimeTriggers] Starting time-based trigger scan")

	// Use the scheduler to process pending jobs
	scheduler := h.engine.GetScheduler()
	if err := scheduler.ProcessPendingJobs(ctx); err != nil {
		log.Printf("[CheckTimeTriggers] Error processing pending jobs: %v", err)
		return err
	}

	// Optionally cleanup old jobs periodically
	if err := scheduler.CleanupOldJobs(ctx); err != nil {
		log.Printf("[CheckTimeTriggers] Error cleaning up old jobs: %v", err)
		// Don't fail the job for cleanup errors
	}

	log.Println("[CheckTimeTriggers] Completed time-based trigger scan")

	return nil
}

// getEntityData retrieves entity data for notifications
func (h *Handlers) getEntityData(ctx context.Context, orgID string, entityType string, entityID uuid.UUID) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	switch entityType {
	case "session":
		row := h.db.Pool.QueryRow(ctx, `
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
		`, entityID, orgID)

		var scheduledAt interface{}
		var sessionType, status, patientName, therapistName string
		var patientPhone, patientEmail *string

		err := row.Scan(&scheduledAt, &sessionType, &status, &patientName, &patientPhone, &patientEmail, &therapistName)
		if err != nil {
			return nil, err
		}

		data["session_id"] = entityID
		data["scheduled_at"] = scheduledAt
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

	case "budget":
		row := h.db.Pool.QueryRow(ctx, `
			SELECT
				b.status,
				b.total_price_cents,
				COALESCE(c.name, '') as client_name,
				c.email as client_email,
				c.phone as client_phone
			FROM budgets b
			LEFT JOIN clients c ON c.id = b.client_id
			WHERE b.id = $1 AND b.organization_id = $2
		`, entityID, orgID)

		var status, clientName string
		var totalCents int
		var clientEmail, clientPhone *string

		err := row.Scan(&status, &totalCents, &clientName, &clientEmail, &clientPhone)
		if err != nil {
			return nil, err
		}

		data["budget_id"] = entityID
		data["status"] = status
		data["budget_total"] = fmt.Sprintf("%.2f", float64(totalCents)/100)
		data["client_name"] = clientName

		if clientEmail != nil {
			data["client_email"] = *clientEmail
		}
		if clientPhone != nil {
			data["client_phone"] = *clientPhone
		}
	}

	return data, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
