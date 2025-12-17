package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/controlewise/backend/internal/database"
	"github.com/controlewise/backend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ModuleService struct {
	db *database.DB
}

func NewModuleService(db *database.DB) *ModuleService {
	return &ModuleService{db: db}
}

// ListAvailable returns all available modules in the system
func (s *ModuleService) ListAvailable(ctx context.Context) ([]*models.AvailableModule, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT name, display_name, description, icon, dependencies, is_active, created_at
		FROM available_modules
		WHERE is_active = TRUE
		ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query available modules: %w", err)
	}
	defer rows.Close()

	var modules []*models.AvailableModule
	for rows.Next() {
		var m models.AvailableModule
		err := rows.Scan(
			&m.Name,
			&m.DisplayName,
			&m.Description,
			&m.Icon,
			&m.Dependencies,
			&m.IsActive,
			&m.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan module: %w", err)
		}
		modules = append(modules, &m)
	}

	return modules, nil
}

// ListForOrganization returns all modules with their status for an organization
func (s *ModuleService) ListForOrganization(ctx context.Context, orgID uuid.UUID) ([]*models.OrganizationModuleWithDetails, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT
			am.name,
			am.display_name,
			am.description,
			am.icon,
			am.dependencies,
			COALESCE(om.id, '00000000-0000-0000-0000-000000000000'::uuid) as id,
			COALESCE(om.organization_id, $1) as organization_id,
			COALESCE(om.is_enabled, FALSE) as is_enabled,
			COALESCE(om.config, '{}') as config,
			om.enabled_at,
			om.enabled_by,
			COALESCE(om.created_at, CURRENT_TIMESTAMP) as created_at,
			COALESCE(om.updated_at, CURRENT_TIMESTAMP) as updated_at
		FROM available_modules am
		LEFT JOIN organization_modules om ON om.module_name = am.name AND om.organization_id = $1
		WHERE am.is_active = TRUE
		ORDER BY am.name
	`, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to query organization modules: %w", err)
	}
	defer rows.Close()

	var modules []*models.OrganizationModuleWithDetails
	for rows.Next() {
		var m models.OrganizationModuleWithDetails
		var dependencies json.RawMessage
		err := rows.Scan(
			&m.ModuleName,
			&m.DisplayName,
			&m.Description,
			&m.Icon,
			&dependencies,
			&m.ID,
			&m.OrganizationID,
			&m.IsEnabled,
			&m.Config,
			&m.EnabledAt,
			&m.EnabledBy,
			&m.CreatedAt,
			&m.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan module: %w", err)
		}

		// Parse dependencies
		if dependencies != nil {
			if err := json.Unmarshal(dependencies, &m.Dependencies); err != nil {
				m.Dependencies = []string{}
			}
		}

		modules = append(modules, &m)
	}

	return modules, nil
}

// IsEnabled checks if a module is enabled for an organization
func (s *ModuleService) IsEnabled(ctx context.Context, orgID uuid.UUID, moduleName models.ModuleName) (bool, error) {
	var isEnabled bool
	err := s.db.Pool.QueryRow(ctx, `
		SELECT is_enabled
		FROM organization_modules
		WHERE organization_id = $1 AND module_name = $2
	`, orgID, moduleName).Scan(&isEnabled)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check module status: %w", err)
	}
	return isEnabled, nil
}

// GetEnabledModules returns a list of enabled module names for an organization
func (s *ModuleService) GetEnabledModules(ctx context.Context, orgID uuid.UUID) ([]models.ModuleName, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT module_name
		FROM organization_modules
		WHERE organization_id = $1 AND is_enabled = TRUE
	`, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to query enabled modules: %w", err)
	}
	defer rows.Close()

	var modules []models.ModuleName
	for rows.Next() {
		var name models.ModuleName
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("failed to scan module name: %w", err)
		}
		modules = append(modules, name)
	}

	return modules, nil
}

// Enable enables a module for an organization (called by regular users)
func (s *ModuleService) Enable(ctx context.Context, orgID uuid.UUID, moduleName models.ModuleName, enabledBy uuid.UUID) error {
	return s.enableModule(ctx, orgID, moduleName, &enabledBy, nil)
}

// EnableByAdmin enables a module for an organization (called by system admins)
func (s *ModuleService) EnableByAdmin(ctx context.Context, orgID uuid.UUID, moduleName models.ModuleName, adminID uuid.UUID) error {
	return s.enableModule(ctx, orgID, moduleName, nil, &adminID)
}

// enableModule is the internal function that handles module enablement
func (s *ModuleService) enableModule(ctx context.Context, orgID uuid.UUID, moduleName models.ModuleName, enabledBy *uuid.UUID, enabledByAdmin *uuid.UUID) error {
	// Check if module exists and is active
	var isActive bool
	err := s.db.Pool.QueryRow(ctx, `
		SELECT is_active FROM available_modules WHERE name = $1
	`, moduleName).Scan(&isActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("module not found")
		}
		return fmt.Errorf("failed to check module: %w", err)
	}
	if !isActive {
		return errors.New("module is not available")
	}

	// Check dependencies
	if err := s.checkDependencies(ctx, orgID, moduleName); err != nil {
		return err
	}

	// Enable or insert
	now := time.Now()
	_, err = s.db.Pool.Exec(ctx, `
		INSERT INTO organization_modules (organization_id, module_name, is_enabled, enabled_at, enabled_by, enabled_by_admin)
		VALUES ($1, $2, TRUE, $3, $4, $5)
		ON CONFLICT (organization_id, module_name)
		DO UPDATE SET is_enabled = TRUE, enabled_at = $3, enabled_by = $4, enabled_by_admin = $5, updated_at = CURRENT_TIMESTAMP
	`, orgID, moduleName, now, enabledBy, enabledByAdmin)
	if err != nil {
		return fmt.Errorf("failed to enable module: %w", err)
	}

	return nil
}

// Disable disables a module for an organization
func (s *ModuleService) Disable(ctx context.Context, orgID uuid.UUID, moduleName models.ModuleName) error {
	// Check if other enabled modules depend on this one
	if err := s.checkDependents(ctx, orgID, moduleName); err != nil {
		return err
	}

	// Disable
	result, err := s.db.Pool.Exec(ctx, `
		UPDATE organization_modules
		SET is_enabled = FALSE, updated_at = CURRENT_TIMESTAMP
		WHERE organization_id = $1 AND module_name = $2
	`, orgID, moduleName)
	if err != nil {
		return fmt.Errorf("failed to disable module: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("module not found or already disabled")
	}

	return nil
}

// UpdateConfig updates the configuration for a module
func (s *ModuleService) UpdateConfig(ctx context.Context, orgID uuid.UUID, moduleName models.ModuleName, config models.ModuleConfig) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	result, err := s.db.Pool.Exec(ctx, `
		UPDATE organization_modules
		SET config = $1, updated_at = CURRENT_TIMESTAMP
		WHERE organization_id = $2 AND module_name = $3 AND is_enabled = TRUE
	`, configJSON, orgID, moduleName)
	if err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("module not found or not enabled")
	}

	return nil
}

// GetConfig returns the configuration for a module
func (s *ModuleService) GetConfig(ctx context.Context, orgID uuid.UUID, moduleName models.ModuleName) (models.ModuleConfig, error) {
	var configJSON json.RawMessage
	err := s.db.Pool.QueryRow(ctx, `
		SELECT config
		FROM organization_modules
		WHERE organization_id = $1 AND module_name = $2
	`, orgID, moduleName).Scan(&configJSON)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.ModuleConfig{}, nil
		}
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	var config models.ModuleConfig
	if configJSON != nil {
		if err := json.Unmarshal(configJSON, &config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}

	return config, nil
}

// checkDependencies verifies that all dependencies are enabled
func (s *ModuleService) checkDependencies(ctx context.Context, orgID uuid.UUID, moduleName models.ModuleName) error {
	// Get dependencies
	var depsJSON json.RawMessage
	err := s.db.Pool.QueryRow(ctx, `
		SELECT dependencies FROM available_modules WHERE name = $1
	`, moduleName).Scan(&depsJSON)
	if err != nil {
		return fmt.Errorf("failed to get dependencies: %w", err)
	}

	var deps []string
	if depsJSON != nil {
		if err := json.Unmarshal(depsJSON, &deps); err != nil {
			return fmt.Errorf("failed to parse dependencies: %w", err)
		}
	}

	// Check each dependency
	for _, dep := range deps {
		enabled, err := s.IsEnabled(ctx, orgID, models.ModuleName(dep))
		if err != nil {
			return err
		}
		if !enabled {
			return fmt.Errorf("required module '%s' is not enabled", dep)
		}
	}

	return nil
}

// checkDependents verifies that no other enabled modules depend on this one
func (s *ModuleService) checkDependents(ctx context.Context, orgID uuid.UUID, moduleName models.ModuleName) error {
	// Get all enabled modules
	rows, err := s.db.Pool.Query(ctx, `
		SELECT am.name, am.dependencies
		FROM organization_modules om
		JOIN available_modules am ON am.name = om.module_name
		WHERE om.organization_id = $1 AND om.is_enabled = TRUE AND om.module_name != $2
	`, orgID, moduleName)
	if err != nil {
		return fmt.Errorf("failed to query dependents: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var depsJSON json.RawMessage
		if err := rows.Scan(&name, &depsJSON); err != nil {
			return fmt.Errorf("failed to scan dependent: %w", err)
		}

		if depsJSON != nil {
			var deps []string
			if err := json.Unmarshal(depsJSON, &deps); err != nil {
				continue
			}
			for _, dep := range deps {
				if dep == string(moduleName) {
					return fmt.Errorf("cannot disable: module '%s' depends on this module", name)
				}
			}
		}
	}

	return nil
}
