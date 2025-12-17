package workflow

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/controlewise/backend/internal/database"
	"github.com/controlewise/backend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// TemplateRenderer handles message template rendering
type TemplateRenderer struct {
	db *database.DB
}

// NewTemplateRenderer creates a new template renderer
func NewTemplateRenderer(db *database.DB) *TemplateRenderer {
	return &TemplateRenderer{db: db}
}

// GetTemplate retrieves a message template by ID
func (r *TemplateRenderer) GetTemplate(ctx context.Context, id, orgID uuid.UUID) (*models.MessageTemplate, error) {
	var t models.MessageTemplate
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, organization_id, name, channel, subject, body, variables, is_active, created_at, updated_at
		FROM message_templates
		WHERE id = $1 AND organization_id = $2
	`, id, orgID).Scan(
		&t.ID, &t.OrganizationID, &t.Name, &t.Channel, &t.Subject,
		&t.Body, &t.Variables, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("template not found")
		}
		return nil, fmt.Errorf("failed to get template: %w", err)
	}
	return &t, nil
}

// RenderTemplate renders a template string with the given data
// Supports {{variable_name}} syntax
func (r *TemplateRenderer) RenderTemplate(template string, data map[string]interface{}) (string, error) {
	if data == nil {
		return template, nil
	}

	// Regular expression to match {{variable_name}}
	re := regexp.MustCompile(`\{\{(\w+)\}\}`)

	result := re.ReplaceAllStringFunc(template, func(match string) string {
		// Extract variable name (remove {{ and }})
		varName := strings.TrimPrefix(match, "{{")
		varName = strings.TrimSuffix(varName, "}}")
		varName = strings.TrimSpace(varName)

		// Look up value in data
		if value, ok := data[varName]; ok {
			return fmt.Sprintf("%v", value)
		}
		// Return original if not found
		return match
	})

	return result, nil
}

// PreviewTemplate renders a template with sample data for preview
func (r *TemplateRenderer) PreviewTemplate(ctx context.Context, template *models.MessageTemplate, entityType string) (string, string, error) {
	// Get sample data based on entity type
	sampleData := getSampleData(entityType)

	// Render body
	body, err := r.RenderTemplate(template.Body, sampleData)
	if err != nil {
		return "", "", err
	}

	// Render subject if present
	var subject string
	if template.Subject != nil {
		subject, err = r.RenderTemplate(*template.Subject, sampleData)
		if err != nil {
			return "", "", err
		}
	}

	return subject, body, nil
}

// getSampleData returns sample data for template preview
func getSampleData(entityType string) map[string]interface{} {
	switch entityType {
	case "session":
		return map[string]interface{}{
			"patient_name":    "João Silva",
			"patient_phone":   "+351912345678",
			"patient_email":   "joao.silva@email.com",
			"therapist_name":  "Dr. Maria Santos",
			"session_date":    "15/01/2025",
			"session_time":    "14:30",
			"session_type":    "Consulta Regular",
			"amount":          "50.00",
			"organization_name": "Clínica Exemplo",
		}
	case "budget":
		return map[string]interface{}{
			"client_name":   "Manuel Costa",
			"client_email":  "manuel.costa@email.com",
			"client_phone":  "+351923456789",
			"project_name":  "Remodelação Cozinha",
			"budget_total":  "15000.00",
			"budget_link":   "https://example.com/budgets/123",
			"approval_link": "https://example.com/budgets/123/approve",
			"organization_name": "Construções ABC",
		}
	case "project":
		return map[string]interface{}{
			"client_name":   "Ana Ferreira",
			"client_email":  "ana.ferreira@email.com",
			"client_phone":  "+351934567890",
			"project_name":  "Construção Moradia",
			"project_status": "Em Curso",
			"organization_name": "Construções ABC",
		}
	default:
		return map[string]interface{}{
			"name":  "Cliente Exemplo",
			"email": "cliente@email.com",
			"phone": "+351900000000",
		}
	}
}

// GetAvailableVariables returns the available variables for a given entity type
func GetAvailableVariables(entityType string) []models.TemplateVariable {
	switch entityType {
	case "session":
		return []models.TemplateVariable{
			{Name: "patient_name", Description: "Nome do paciente"},
			{Name: "patient_phone", Description: "Telefone do paciente"},
			{Name: "patient_email", Description: "Email do paciente"},
			{Name: "therapist_name", Description: "Nome do terapeuta"},
			{Name: "session_date", Description: "Data da sessão (DD/MM/AAAA)"},
			{Name: "session_time", Description: "Hora da sessão (HH:MM)"},
			{Name: "session_type", Description: "Tipo de sessão"},
			{Name: "amount", Description: "Valor da sessão"},
			{Name: "organization_name", Description: "Nome da organização"},
		}
	case "budget":
		return []models.TemplateVariable{
			{Name: "client_name", Description: "Nome do cliente"},
			{Name: "client_email", Description: "Email do cliente"},
			{Name: "client_phone", Description: "Telefone do cliente"},
			{Name: "project_name", Description: "Nome do projeto"},
			{Name: "budget_total", Description: "Valor total do orçamento"},
			{Name: "budget_link", Description: "Link para visualizar o orçamento"},
			{Name: "approval_link", Description: "Link para aprovar o orçamento"},
			{Name: "organization_name", Description: "Nome da organização"},
		}
	case "project":
		return []models.TemplateVariable{
			{Name: "client_name", Description: "Nome do cliente"},
			{Name: "client_email", Description: "Email do cliente"},
			{Name: "client_phone", Description: "Telefone do cliente"},
			{Name: "project_name", Description: "Nome do projeto"},
			{Name: "project_status", Description: "Estado do projeto"},
			{Name: "organization_name", Description: "Nome da organização"},
		}
	default:
		return []models.TemplateVariable{}
	}
}

// ValidateTemplate checks if a template uses valid variables
func (r *TemplateRenderer) ValidateTemplate(template string, entityType string) []string {
	availableVars := GetAvailableVariables(entityType)
	validVarNames := make(map[string]bool)
	for _, v := range availableVars {
		validVarNames[v.Name] = true
	}

	// Find all variables used in template
	re := regexp.MustCompile(`\{\{(\w+)\}\}`)
	matches := re.FindAllStringSubmatch(template, -1)

	var invalidVars []string
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 {
			varName := match[1]
			if !validVarNames[varName] && !seen[varName] {
				invalidVars = append(invalidVars, varName)
				seen[varName] = true
			}
		}
	}

	return invalidVars
}
