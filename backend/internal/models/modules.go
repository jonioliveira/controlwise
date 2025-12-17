package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ModuleName represents the available module identifiers
type ModuleName string

const (
	ModuleConstruction  ModuleName = "construction"
	ModuleAppointments  ModuleName = "appointments"
	ModuleNotifications ModuleName = "notifications"
)

// AvailableModule represents a system-level module definition
type AvailableModule struct {
	Name         ModuleName      `json:"name" db:"name"`
	DisplayName  string          `json:"display_name" db:"display_name"`
	Description  *string         `json:"description" db:"description"`
	Icon         *string         `json:"icon" db:"icon"`
	Dependencies json.RawMessage `json:"dependencies" db:"dependencies"`
	IsActive     bool            `json:"is_active" db:"is_active"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
}

// GetDependencies parses the dependencies JSON into a string slice
func (m *AvailableModule) GetDependencies() ([]string, error) {
	if m.Dependencies == nil {
		return []string{}, nil
	}
	var deps []string
	if err := json.Unmarshal(m.Dependencies, &deps); err != nil {
		return nil, err
	}
	return deps, nil
}

// OrganizationModule represents a module enabled for an organization
type OrganizationModule struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	OrganizationID uuid.UUID       `json:"organization_id" db:"organization_id"`
	ModuleName     ModuleName      `json:"module_name" db:"module_name"`
	IsEnabled      bool            `json:"is_enabled" db:"is_enabled"`
	Config         json.RawMessage `json:"config" db:"config"`
	EnabledAt      *time.Time      `json:"enabled_at" db:"enabled_at"`
	EnabledBy      *uuid.UUID      `json:"enabled_by" db:"enabled_by"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
}

// OrganizationModuleWithDetails includes module details for API responses
type OrganizationModuleWithDetails struct {
	OrganizationModule
	DisplayName  string   `json:"display_name"`
	Description  *string  `json:"description"`
	Icon         *string  `json:"icon"`
	Dependencies []string `json:"dependencies"`
}

// ModuleConfig represents module-specific configuration
type ModuleConfig map[string]interface{}

// GetConfig parses the config JSON into a map
func (m *OrganizationModule) GetConfig() (ModuleConfig, error) {
	if m.Config == nil {
		return ModuleConfig{}, nil
	}
	var config ModuleConfig
	if err := json.Unmarshal(m.Config, &config); err != nil {
		return nil, err
	}
	return config, nil
}

// SetConfig serializes a config map to JSON
func (m *OrganizationModule) SetConfig(config ModuleConfig) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	m.Config = data
	return nil
}
