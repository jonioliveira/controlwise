package middleware

import (
	"net/http"

	"github.com/controlewise/backend/internal/models"
	"github.com/controlewise/backend/internal/services"
	"github.com/controlewise/backend/internal/utils"
)

// ModuleMiddleware checks if modules are enabled for organizations
type ModuleMiddleware struct {
	moduleService *services.ModuleService
}

// NewModuleMiddleware creates a new module middleware
func NewModuleMiddleware(moduleService *services.ModuleService) *ModuleMiddleware {
	return &ModuleMiddleware{
		moduleService: moduleService,
	}
}

// RequireModule returns middleware that checks if a module is enabled for the organization
func (m *ModuleMiddleware) RequireModule(moduleName models.ModuleName) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			orgID, ok := GetOrganizationID(r.Context())
			if !ok {
				utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found in context")
				return
			}

			enabled, err := m.moduleService.IsEnabled(r.Context(), orgID, moduleName)
			if err != nil {
				utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to check module status")
				return
			}

			if !enabled {
				utils.ErrorResponse(w, http.StatusForbidden, "Module not enabled for this organization")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyModule returns middleware that checks if any of the given modules is enabled
func (m *ModuleMiddleware) RequireAnyModule(moduleNames ...models.ModuleName) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			orgID, ok := GetOrganizationID(r.Context())
			if !ok {
				utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found in context")
				return
			}

			for _, moduleName := range moduleNames {
				enabled, err := m.moduleService.IsEnabled(r.Context(), orgID, moduleName)
				if err != nil {
					continue
				}
				if enabled {
					next.ServeHTTP(w, r)
					return
				}
			}

			utils.ErrorResponse(w, http.StatusForbidden, "None of the required modules are enabled")
		})
	}
}

// RequireAllModules returns middleware that checks if all given modules are enabled
func (m *ModuleMiddleware) RequireAllModules(moduleNames ...models.ModuleName) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			orgID, ok := GetOrganizationID(r.Context())
			if !ok {
				utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found in context")
				return
			}

			for _, moduleName := range moduleNames {
				enabled, err := m.moduleService.IsEnabled(r.Context(), orgID, moduleName)
				if err != nil {
					utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to check module status")
					return
				}
				if !enabled {
					utils.ErrorResponse(w, http.StatusForbidden, "Required module not enabled: "+string(moduleName))
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
