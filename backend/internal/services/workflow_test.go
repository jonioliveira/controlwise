package services

import (
	"testing"
)

func TestGetSampleDataForEntityType(t *testing.T) {
	tests := []struct {
		name         string
		entityType   string
		expectedKeys []string
	}{
		{
			name:       "session entity type",
			entityType: "session",
			expectedKeys: []string{
				"patient_name",
				"patient_phone",
				"patient_email",
				"therapist_name",
				"session_date",
				"session_time",
				"session_type",
				"amount",
				"organization_name",
				"organization_email",
			},
		},
		{
			name:       "budget entity type",
			entityType: "budget",
			expectedKeys: []string{
				"client_name",
				"client_email",
				"client_phone",
				"project_name",
				"budget_number",
				"budget_total",
				"budget_link",
				"approval_link",
				"organization_name",
				"organization_email",
			},
		},
		{
			name:       "project entity type",
			entityType: "project",
			expectedKeys: []string{
				"client_name",
				"client_email",
				"client_phone",
				"project_name",
				"project_number",
				"project_status",
				"organization_name",
				"organization_email",
			},
		},
		{
			name:       "unknown entity type",
			entityType: "unknown",
			expectedKeys: []string{
				"name",
				"email",
				"phone",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := GetSampleDataForEntityType(tt.entityType)
			if data == nil {
				t.Error("GetSampleDataForEntityType() returned nil")
				return
			}

			for _, key := range tt.expectedKeys {
				if _, ok := data[key]; !ok {
					t.Errorf("GetSampleDataForEntityType(%q) missing expected key %q", tt.entityType, key)
				}
			}
		})
	}
}

func TestRenderTemplateString(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "simple replacement",
			template: "Hello {{name}}",
			data:     map[string]interface{}{"name": "World"},
			expected: "Hello World",
		},
		{
			name:     "multiple replacements",
			template: "{{greeting}} {{name}}, welcome to {{place}}!",
			data: map[string]interface{}{
				"greeting": "Hello",
				"name":     "João",
				"place":    "ControleWise",
			},
			expected: "Hello João, welcome to ControleWise!",
		},
		{
			name:     "repeated variable",
			template: "{{name}} said: My name is {{name}}",
			data:     map[string]interface{}{"name": "Maria"},
			expected: "Maria said: My name is Maria",
		},
		{
			name:     "missing variable unchanged",
			template: "Hello {{name}}, {{missing}}",
			data:     map[string]interface{}{"name": "Test"},
			expected: "Hello Test, {{missing}}",
		},
		{
			name:     "empty data map",
			template: "Hello {{name}}",
			data:     map[string]interface{}{},
			expected: "Hello {{name}}",
		},
		{
			name:     "numeric values",
			template: "Total: {{amount}}€",
			data:     map[string]interface{}{"amount": 100},
			expected: "Total: 100€",
		},
		{
			name:     "float values",
			template: "Price: {{price}}",
			data:     map[string]interface{}{"price": 99.99},
			expected: "Price: 99.99",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderTemplateString(tt.template, tt.data)
			if result != tt.expected {
				t.Errorf("renderTemplateString() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestParseActionConfigJSON(t *testing.T) {
	tests := []struct {
		name     string
		config   []byte
		hasKey   string
		keyValue interface{}
	}{
		{
			name:     "valid config",
			config:   []byte(`{"subject": "Test Subject", "to_field": "client_email"}`),
			hasKey:   "subject",
			keyValue: "Test Subject",
		},
		{
			name:     "empty config",
			config:   []byte(`{}`),
			hasKey:   "",
			keyValue: nil,
		},
		{
			name:     "nil config",
			config:   nil,
			hasKey:   "",
			keyValue: nil,
		},
		{
			name:     "invalid JSON",
			config:   []byte(`{invalid`),
			hasKey:   "",
			keyValue: nil,
		},
		{
			name:     "nested config",
			config:   []byte(`{"field": "status", "value": "completed"}`),
			hasKey:   "field",
			keyValue: "status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseActionConfigJSON(tt.config)

			if tt.hasKey == "" {
				// Just check it doesn't panic
				return
			}

			if result == nil && tt.keyValue != nil {
				t.Error("parseActionConfigJSON() returned nil but expected a value")
				return
			}

			if result != nil && tt.hasKey != "" {
				if val, ok := result[tt.hasKey]; !ok {
					t.Errorf("parseActionConfigJSON() missing expected key %q", tt.hasKey)
				} else if val != tt.keyValue {
					t.Errorf("parseActionConfigJSON()[%q] = %v, want %v", tt.hasKey, val, tt.keyValue)
				}
			}
		})
	}
}

func TestStringPtr(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"simple string", "test"},
		{"unicode string", "Olá Mundo"},
		{"long string", "This is a much longer string with multiple words and special characters: @#$%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringPtr(tt.input)
			if result == nil {
				t.Error("stringPtr() returned nil")
				return
			}
			if *result != tt.input {
				t.Errorf("stringPtr() = %q, want %q", *result, tt.input)
			}
			// Verify it's actually a new pointer
			result2 := stringPtr(tt.input)
			if result == result2 {
				t.Error("stringPtr() returned same pointer for different calls")
			}
		})
	}
}

func TestSampleDataHasRealisticValues(t *testing.T) {
	// Ensure sample data has realistic Portuguese values
	sessionData := GetSampleDataForEntityType("session")

	// Check patient name is a realistic Portuguese name
	patientName, ok := sessionData["patient_name"].(string)
	if !ok {
		t.Error("patient_name should be a string")
		return
	}
	if len(patientName) < 5 {
		t.Error("patient_name should be a realistic name")
	}

	// Check phone number format
	phone, ok := sessionData["patient_phone"].(string)
	if !ok {
		t.Error("patient_phone should be a string")
		return
	}
	if phone == "" || phone[0] != '+' {
		t.Error("patient_phone should be in international format")
	}

	// Check email format
	email, ok := sessionData["patient_email"].(string)
	if !ok {
		t.Error("patient_email should be a string")
		return
	}
	if !containsChar(email, '@') {
		t.Error("patient_email should be a valid email")
	}
}

func containsChar(s string, c byte) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return true
		}
	}
	return false
}

func TestBudgetSampleDataCompleteness(t *testing.T) {
	budgetData := GetSampleDataForEntityType("budget")

	// Verify budget_total is a monetary value string
	total, ok := budgetData["budget_total"].(string)
	if !ok {
		t.Error("budget_total should be a string")
		return
	}
	if len(total) == 0 {
		t.Error("budget_total should not be empty")
	}

	// Verify budget_number format
	number, ok := budgetData["budget_number"].(string)
	if !ok {
		t.Error("budget_number should be a string")
		return
	}
	if len(number) < 5 {
		t.Error("budget_number should be a realistic budget number")
	}

	// Verify links are valid URLs
	link, ok := budgetData["budget_link"].(string)
	if !ok {
		t.Error("budget_link should be a string")
		return
	}
	if len(link) < 10 || link[:4] != "http" {
		t.Error("budget_link should be a valid URL")
	}
}

func TestProjectSampleDataCompleteness(t *testing.T) {
	projectData := GetSampleDataForEntityType("project")

	// Verify project_number format
	number, ok := projectData["project_number"].(string)
	if !ok {
		t.Error("project_number should be a string")
		return
	}
	if len(number) < 5 {
		t.Error("project_number should be a realistic project number")
	}

	// Verify project_status is a valid status
	status, ok := projectData["project_status"].(string)
	if !ok {
		t.Error("project_status should be a string")
		return
	}
	if len(status) == 0 {
		t.Error("project_status should not be empty")
	}
}

func BenchmarkRenderTemplateString(b *testing.B) {
	template := "Dear {{patient_name}}, your appointment with {{therapist_name}} is scheduled for {{session_date}} at {{session_time}}. Please contact us at {{organization_email}} for any questions."
	data := map[string]interface{}{
		"patient_name":       "João Silva",
		"therapist_name":     "Dr. Maria Santos",
		"session_date":       "15/01/2025",
		"session_time":       "14:30",
		"organization_email": "clinica@exemplo.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderTemplateString(template, data)
	}
}

func BenchmarkParseActionConfigJSON(b *testing.B) {
	config := []byte(`{"subject": "Test Subject", "to_field": "client_email", "body": "This is a test body with {{variables}}", "extra": true}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseActionConfigJSON(config)
	}
}
