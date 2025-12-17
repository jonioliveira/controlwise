package middleware

import (
	"context"
	"net/http"

	"github.com/controlewise/backend/internal/services"
	"github.com/controlewise/backend/internal/utils"
)

type OrganizationMiddleware struct {
	orgService *services.OrganizationService
}

func NewOrganizationMiddleware(orgService *services.OrganizationService) *OrganizationMiddleware {
	return &OrganizationMiddleware{
		orgService: orgService,
	}
}

func (m *OrganizationMiddleware) ExtractOrganization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orgID, ok := GetOrganizationID(r.Context())
		if !ok {
			utils.ErrorResponse(w, http.StatusUnauthorized, "Organization ID not found in context")
			return
		}

		// Verify organization exists and is active
		org, err := m.orgService.GetByID(r.Context(), orgID)
		if err != nil {
			utils.ErrorResponse(w, http.StatusNotFound, "Organization not found")
			return
		}

		if !org.IsActive {
			utils.ErrorResponse(w, http.StatusForbidden, "Organization is not active")
			return
		}

		// Add organization to context for easy access
		ctx := context.WithValue(r.Context(), "organization", org)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
