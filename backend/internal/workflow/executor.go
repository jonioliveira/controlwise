package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/controlewise/backend/internal/database"
	"github.com/controlewise/backend/internal/models"
	"github.com/google/uuid"
)

// NotificationSender interface for sending notifications
type NotificationSender interface {
	SendWhatsApp(ctx context.Context, phone, message string) error
	SendEmail(ctx context.Context, to, subject, body string) error
}

// Executor handles workflow action execution
type Executor struct {
	db             *database.DB
	templates      *TemplateRenderer
	notifySender   NotificationSender
}

// NewExecutor creates a new action executor
func NewExecutor(db *database.DB) *Executor {
	return &Executor{
		db:        db,
		templates: NewTemplateRenderer(db),
	}
}

// SetNotificationSender sets the notification sender implementation
func (e *Executor) SetNotificationSender(sender NotificationSender) {
	e.notifySender = sender
}

// ExecuteAction executes a single workflow action
func (e *Executor) ExecuteAction(ctx context.Context, orgID uuid.UUID, action *models.WorkflowAction, entityType string, entityID uuid.UUID, entityData map[string]interface{}) error {
	log.Printf("[Executor] Executing action %s (type=%s)", action.ID, action.ActionType)

	switch action.ActionType {
	case models.ActionTypeSendWhatsApp:
		return e.executeSendWhatsApp(ctx, orgID, action, entityType, entityID, entityData)
	case models.ActionTypeSendEmail:
		return e.executeSendEmail(ctx, orgID, action, entityType, entityID, entityData)
	case models.ActionTypeUpdateField:
		return e.executeUpdateField(ctx, orgID, action, entityType, entityID, entityData)
	case models.ActionTypeCreateTask:
		return e.executeCreateTask(ctx, orgID, action, entityType, entityID, entityData)
	default:
		return fmt.Errorf("unknown action type: %s", action.ActionType)
	}
}

// executeSendWhatsApp sends a WhatsApp message using a template
func (e *Executor) executeSendWhatsApp(ctx context.Context, orgID uuid.UUID, action *models.WorkflowAction, entityType string, entityID uuid.UUID, entityData map[string]interface{}) error {
	if action.TemplateID == nil {
		return fmt.Errorf("WhatsApp action requires a template")
	}

	// Get template
	template, err := e.templates.GetTemplate(ctx, *action.TemplateID, orgID)
	if err != nil {
		return fmt.Errorf("failed to get template: %w", err)
	}

	if template.Channel != models.MessageChannelWhatsApp {
		return fmt.Errorf("template is not a WhatsApp template")
	}

	// Get entity data if not provided
	if entityData == nil {
		entityData, err = e.getEntityData(ctx, orgID, entityType, entityID)
		if err != nil {
			return fmt.Errorf("failed to get entity data: %w", err)
		}
	}

	// Render template
	message, err := e.templates.RenderTemplate(template.Body, entityData)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	// Get recipient phone number
	phone, ok := entityData["patient_phone"].(string)
	if !ok || phone == "" {
		phone, ok = entityData["client_phone"].(string)
		if !ok || phone == "" {
			return fmt.Errorf("no phone number available for WhatsApp")
		}
	}

	log.Printf("[Executor] Sending WhatsApp to %s: %s", phone, truncateString(message, 50))

	// Send notification
	if e.notifySender != nil {
		if err := e.notifySender.SendWhatsApp(ctx, phone, message); err != nil {
			return fmt.Errorf("failed to send WhatsApp: %w", err)
		}
	} else {
		log.Printf("[Executor] WhatsApp sender not configured, skipping send")
	}

	return nil
}

// executeSendEmail sends an email using a template or inline config
func (e *Executor) executeSendEmail(ctx context.Context, orgID uuid.UUID, action *models.WorkflowAction, entityType string, entityID uuid.UUID, entityData map[string]interface{}) error {
	var subject, body string
	var err error

	// Get entity data if not provided
	if entityData == nil {
		entityData, err = e.getEntityData(ctx, orgID, entityType, entityID)
		if err != nil {
			return fmt.Errorf("failed to get entity data: %w", err)
		}
	}

	if action.TemplateID != nil {
		// Get template
		template, err := e.templates.GetTemplate(ctx, *action.TemplateID, orgID)
		if err != nil {
			return fmt.Errorf("failed to get template: %w", err)
		}

		if template.Channel != models.MessageChannelEmail {
			return fmt.Errorf("template is not an email template")
		}

		// Render template body
		body, err = e.templates.RenderTemplate(template.Body, entityData)
		if err != nil {
			return fmt.Errorf("failed to render template body: %w", err)
		}

		// Render subject if present
		subject = "Notificação"
		if template.Subject != nil {
			subject, err = e.templates.RenderTemplate(*template.Subject, entityData)
			if err != nil {
				return fmt.Errorf("failed to render template subject: %w", err)
			}
		}
	} else {
		// Use inline config
		config, err := parseActionConfig(action.ActionConfig)
		if err != nil {
			return fmt.Errorf("failed to parse action config: %w", err)
		}

		subjectTemplate, _ := config["subject"].(string)
		bodyTemplate, _ := config["body"].(string)

		if subjectTemplate == "" {
			subjectTemplate = "Notificação - {{client_name}}"
		}

		if bodyTemplate == "" {
			bodyTemplate = "Olá {{client_name}},\n\nTem uma nova notificação.\n\nCumprimentos"
		}

		subject, err = e.templates.RenderTemplate(subjectTemplate, entityData)
		if err != nil {
			return fmt.Errorf("failed to render subject: %w", err)
		}

		body, err = e.templates.RenderTemplate(bodyTemplate, entityData)
		if err != nil {
			return fmt.Errorf("failed to render body: %w", err)
		}
	}

	// Determine recipient email based on config or entity data
	var email string
	if action.ActionConfig != nil {
		config, _ := parseActionConfig(action.ActionConfig)
		toField, _ := config["to_field"].(string)
		if toField != "" && entityData[toField] != nil {
			email, _ = entityData[toField].(string)
		}
	}

	// Fall back to entity email fields
	if email == "" {
		email, _ = entityData["patient_email"].(string)
		if email == "" {
			email, _ = entityData["client_email"].(string)
		}
	}

	if email == "" {
		return fmt.Errorf("no email address available for entity %s/%s", entityType, entityID)
	}

	log.Printf("[Executor] Sending email to %s: subject=%s", email, subject)

	// Send notification
	if e.notifySender != nil {
		if err := e.notifySender.SendEmail(ctx, email, subject, body); err != nil {
			return fmt.Errorf("failed to send email: %w", err)
		}
	} else {
		log.Printf("[Executor] Email sender not configured, skipping send")
	}

	return nil
}

// executeUpdateField updates a field on the entity
func (e *Executor) executeUpdateField(ctx context.Context, orgID uuid.UUID, action *models.WorkflowAction, entityType string, entityID uuid.UUID, entityData map[string]interface{}) error {
	// Parse action config
	config, err := parseActionConfig(action.ActionConfig)
	if err != nil {
		return fmt.Errorf("failed to parse action config: %w", err)
	}

	fieldName, ok := config["field"].(string)
	if !ok {
		return fmt.Errorf("update_field action requires 'field' in config")
	}
	fieldValue := config["value"]

	log.Printf("[Executor] Updating %s.%s = %v for entity %s", entityType, fieldName, fieldValue, entityID)

	// Build update query based on entity type
	var query string
	switch entityType {
	case "session":
		query = fmt.Sprintf(`UPDATE sessions SET %s = $1, updated_at = NOW() WHERE id = $2 AND organization_id = $3`, fieldName)
	case "budget":
		query = fmt.Sprintf(`UPDATE budgets SET %s = $1, updated_at = NOW() WHERE id = $2 AND organization_id = $3`, fieldName)
	case "project":
		query = fmt.Sprintf(`UPDATE projects SET %s = $1, updated_at = NOW() WHERE id = $2 AND organization_id = $3`, fieldName)
	default:
		return fmt.Errorf("unsupported entity type for update_field: %s", entityType)
	}

	_, err = e.db.Pool.Exec(ctx, query, fieldValue, entityID, orgID)
	if err != nil {
		return fmt.Errorf("failed to update field: %w", err)
	}

	return nil
}

// executeCreateTask creates a task/reminder
func (e *Executor) executeCreateTask(ctx context.Context, orgID uuid.UUID, action *models.WorkflowAction, entityType string, entityID uuid.UUID, entityData map[string]interface{}) error {
	// Parse action config
	config, err := parseActionConfig(action.ActionConfig)
	if err != nil {
		return fmt.Errorf("failed to parse action config: %w", err)
	}

	title, _ := config["title"].(string)
	description, _ := config["description"].(string)
	assigneeID, _ := config["assignee_id"].(string)

	if title == "" {
		title = fmt.Sprintf("Task for %s %s", entityType, entityID)
	}

	log.Printf("[Executor] Creating task: %s for entity %s/%s", title, entityType, entityID)

	// For now, just log the task creation
	// In a full implementation, this would create a task in a tasks table
	log.Printf("[Executor] Task details: title=%s, description=%s, assignee=%s", title, description, assigneeID)

	return nil
}

// getEntityData retrieves entity data for template rendering
func (e *Executor) getEntityData(ctx context.Context, orgID uuid.UUID, entityType string, entityID uuid.UUID) (map[string]interface{}, error) {
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
func (e *Executor) getSessionData(ctx context.Context, orgID uuid.UUID, sessionID uuid.UUID) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	row := e.db.Pool.QueryRow(ctx, `
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
	`, sessionID, orgID)

	var scheduledAt interface{}
	var sessionType, status, patientName, therapistName string
	var patientPhone, patientEmail *string

	err := row.Scan(&scheduledAt, &sessionType, &status, &patientName, &patientPhone, &patientEmail, &therapistName)
	if err != nil {
		return nil, err
	}

	data["session_id"] = sessionID.String()
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

	return data, nil
}

// getBudgetData retrieves budget data with client info
func (e *Executor) getBudgetData(ctx context.Context, orgID uuid.UUID, budgetID uuid.UUID) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	row := e.db.Pool.QueryRow(ctx, `
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
	`, budgetID, orgID)

	var status, budgetNumber, worksheetTitle, clientName string
	var total float64
	var clientEmail, clientPhone *string

	err := row.Scan(&status, &budgetNumber, &total, &worksheetTitle, &clientName, &clientEmail, &clientPhone)
	if err != nil {
		return nil, err
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
func (e *Executor) getProjectData(ctx context.Context, orgID uuid.UUID, projectID uuid.UUID) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	row := e.db.Pool.QueryRow(ctx, `
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
	`, projectID, orgID)

	var projectTitle, projectNumber, status, clientName string
	var clientEmail, clientPhone *string

	err := row.Scan(&projectTitle, &projectNumber, &status, &clientName, &clientEmail, &clientPhone)
	if err != nil {
		return nil, err
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

// parseActionConfig parses the action_config JSON
func parseActionConfig(config []byte) (map[string]interface{}, error) {
	if config == nil {
		return make(map[string]interface{}), nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(config, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// truncateString truncates a string to max length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
