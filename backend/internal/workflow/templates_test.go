package workflow

import (
	"testing"

	"github.com/controlewise/backend/internal/models"
)

func TestRenderTemplate(t *testing.T) {
	renderer := &TemplateRenderer{}

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "simple variable replacement",
			template: "Hello {{name}}!",
			data:     map[string]interface{}{"name": "João"},
			expected: "Hello João!",
		},
		{
			name:     "multiple variables",
			template: "Dear {{patient_name}}, your appointment is on {{session_date}} at {{session_time}}.",
			data: map[string]interface{}{
				"patient_name": "Maria Silva",
				"session_date": "15/01/2025",
				"session_time": "14:30",
			},
			expected: "Dear Maria Silva, your appointment is on 15/01/2025 at 14:30.",
		},
		{
			name:     "missing variable keeps placeholder",
			template: "Hello {{name}}, your order {{order_id}} is ready.",
			data:     map[string]interface{}{"name": "Carlos"},
			expected: "Hello Carlos, your order {{order_id}} is ready.",
		},
		{
			name:     "no variables in template",
			template: "This is a plain message without variables.",
			data:     map[string]interface{}{"unused": "value"},
			expected: "This is a plain message without variables.",
		},
		{
			name:     "nil data",
			template: "Hello {{name}}!",
			data:     nil,
			expected: "Hello {{name}}!",
		},
		{
			name:     "empty data map",
			template: "Hello {{name}}!",
			data:     map[string]interface{}{},
			expected: "Hello {{name}}!",
		},
		{
			name:     "numeric value",
			template: "Total: {{amount}}€",
			data:     map[string]interface{}{"amount": 150.50},
			expected: "Total: 150.5€",
		},
		{
			name:     "integer value",
			template: "You have {{count}} items",
			data:     map[string]interface{}{"count": 5},
			expected: "You have 5 items",
		},
		{
			name:     "budget notification template",
			template: "Orçamento {{budget_number}} para {{client_name}} - Total: {{budget_total}}€",
			data: map[string]interface{}{
				"budget_number": "ORC-2025-001",
				"client_name":   "Manuel Costa",
				"budget_total":  "15000.00",
			},
			expected: "Orçamento ORC-2025-001 para Manuel Costa - Total: 15000.00€",
		},
		{
			name:     "whatsapp reminder template",
			template: "Olá {{patient_name}}! Lembramos que tem uma consulta agendada para amanhã às {{session_time}} com {{therapist_name}}.",
			data: map[string]interface{}{
				"patient_name":   "Ana Ferreira",
				"session_time":   "10:00",
				"therapist_name": "Dr. João Santos",
			},
			expected: "Olá Ana Ferreira! Lembramos que tem uma consulta agendada para amanhã às 10:00 com Dr. João Santos.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := renderer.RenderTemplate(tt.template, tt.data)
			if err != nil {
				t.Errorf("RenderTemplate() error = %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("RenderTemplate() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestValidateTemplate(t *testing.T) {
	renderer := &TemplateRenderer{}

	tests := []struct {
		name            string
		template        string
		entityType      string
		invalidVarsLen  int
		expectedInvalid []string
	}{
		{
			name:           "valid session variables",
			template:       "Dear {{patient_name}}, your session is at {{session_time}}.",
			entityType:     "session",
			invalidVarsLen: 0,
		},
		{
			name:           "invalid variable in session",
			template:       "Dear {{patient_name}}, your budget is {{budget_total}}.",
			entityType:     "session",
			invalidVarsLen: 1,
			expectedInvalid: []string{"budget_total"},
		},
		{
			name:           "valid budget variables",
			template:       "Client: {{client_name}}, Total: {{budget_total}}€",
			entityType:     "budget",
			invalidVarsLen: 0,
		},
		{
			name:           "multiple invalid variables",
			template:       "{{unknown1}} and {{unknown2}} and {{client_name}}",
			entityType:     "budget",
			invalidVarsLen: 2,
			expectedInvalid: []string{"unknown1", "unknown2"},
		},
		{
			name:           "no variables",
			template:       "This is a plain message.",
			entityType:     "session",
			invalidVarsLen: 0,
		},
		{
			name:           "valid project variables",
			template:       "Project {{project_name}} for {{client_name}}",
			entityType:     "project",
			invalidVarsLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invalid := renderer.ValidateTemplate(tt.template, tt.entityType)
			if len(invalid) != tt.invalidVarsLen {
				t.Errorf("ValidateTemplate() returned %d invalid vars, want %d: %v", len(invalid), tt.invalidVarsLen, invalid)
			}
			if tt.expectedInvalid != nil {
				for _, expected := range tt.expectedInvalid {
					found := false
					for _, inv := range invalid {
						if inv == expected {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("ValidateTemplate() expected invalid var %q not found in %v", expected, invalid)
					}
				}
			}
		})
	}
}

func TestGetAvailableVariables(t *testing.T) {
	tests := []struct {
		name       string
		entityType string
		minVars    int
		checkVar   string
	}{
		{
			name:       "session variables",
			entityType: "session",
			minVars:    5,
			checkVar:   "patient_name",
		},
		{
			name:       "budget variables",
			entityType: "budget",
			minVars:    5,
			checkVar:   "client_name",
		},
		{
			name:       "project variables",
			entityType: "project",
			minVars:    3,
			checkVar:   "project_name",
		},
		{
			name:       "unknown entity type",
			entityType: "unknown",
			minVars:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vars := GetAvailableVariables(tt.entityType)
			if len(vars) < tt.minVars {
				t.Errorf("GetAvailableVariables(%q) returned %d vars, want at least %d", tt.entityType, len(vars), tt.minVars)
			}
			if tt.checkVar != "" {
				found := false
				for _, v := range vars {
					if v.Name == tt.checkVar {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("GetAvailableVariables(%q) missing expected variable %q", tt.entityType, tt.checkVar)
				}
			}
		})
	}
}

func TestGetSampleData(t *testing.T) {
	tests := []struct {
		name       string
		entityType string
		checkKey   string
	}{
		{
			name:       "session sample data",
			entityType: "session",
			checkKey:   "patient_name",
		},
		{
			name:       "budget sample data",
			entityType: "budget",
			checkKey:   "budget_total",
		},
		{
			name:       "project sample data",
			entityType: "project",
			checkKey:   "project_name",
		},
		{
			name:       "default sample data",
			entityType: "unknown",
			checkKey:   "name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := getSampleData(tt.entityType)
			if data == nil {
				t.Error("getSampleData() returned nil")
				return
			}
			if _, ok := data[tt.checkKey]; !ok {
				t.Errorf("getSampleData(%q) missing expected key %q", tt.entityType, tt.checkKey)
			}
		})
	}
}

func TestTemplateVariableDescriptions(t *testing.T) {
	// Ensure all session variables have descriptions
	sessionVars := GetAvailableVariables("session")
	for _, v := range sessionVars {
		if v.Description == "" {
			t.Errorf("Session variable %q has empty description", v.Name)
		}
	}

	// Ensure all budget variables have descriptions
	budgetVars := GetAvailableVariables("budget")
	for _, v := range budgetVars {
		if v.Description == "" {
			t.Errorf("Budget variable %q has empty description", v.Name)
		}
	}

	// Ensure all project variables have descriptions
	projectVars := GetAvailableVariables("project")
	for _, v := range projectVars {
		if v.Description == "" {
			t.Errorf("Project variable %q has empty description", v.Name)
		}
	}
}

func TestSampleDataMatchesAvailableVariables(t *testing.T) {
	entityTypes := []string{"session", "budget", "project"}

	for _, entityType := range entityTypes {
		t.Run(entityType, func(t *testing.T) {
			vars := GetAvailableVariables(entityType)
			sampleData := getSampleData(entityType)

			// Check that all available variables have sample data
			for _, v := range vars {
				if _, ok := sampleData[v.Name]; !ok {
					t.Errorf("Variable %q is available but has no sample data for entity type %q", v.Name, entityType)
				}
			}
		})
	}
}

func TestPreviewTemplateRendering(t *testing.T) {
	// Test that a full template can be previewed with sample data
	template := &models.MessageTemplate{
		Body: "Olá {{patient_name}}! Sua consulta com {{therapist_name}} é em {{session_date}} às {{session_time}}.",
	}
	subject := "Lembrete: Consulta {{session_date}}"
	template.Subject = &subject

	renderer := &TemplateRenderer{}
	sampleData := getSampleData("session")

	// Render body
	body, err := renderer.RenderTemplate(template.Body, sampleData)
	if err != nil {
		t.Errorf("RenderTemplate() body error = %v", err)
		return
	}

	// Body should not contain any remaining placeholders from sample data
	if containsPlaceholder(body, "patient_name") ||
		containsPlaceholder(body, "therapist_name") ||
		containsPlaceholder(body, "session_date") ||
		containsPlaceholder(body, "session_time") {
		t.Errorf("Rendered body still contains placeholders: %s", body)
	}

	// Render subject
	renderedSubject, err := renderer.RenderTemplate(*template.Subject, sampleData)
	if err != nil {
		t.Errorf("RenderTemplate() subject error = %v", err)
		return
	}

	if containsPlaceholder(renderedSubject, "session_date") {
		t.Errorf("Rendered subject still contains placeholder: %s", renderedSubject)
	}
}

func containsPlaceholder(s, varName string) bool {
	return len(s) > 0 && len(varName) > 0 &&
		(len(s) >= len(varName)+4) &&
		(indexOf(s, "{{"+varName+"}}") >= 0)
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
