package middleware

import (
	"net/http"

	"github.com/controlewise/backend/internal/utils"
)

// RequireSystemAdmin middleware blocks requests that don't come from system admins
func RequireSystemAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !IsSysAdmin(r.Context()) {
			utils.ErrorResponse(w, http.StatusForbidden, "System administrator access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireNonImpersonation middleware blocks requests during impersonation sessions
// Use this for sensitive operations that shouldn't be performed while impersonating
func RequireNonImpersonation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if IsImpersonation(r.Context()) {
			utils.ErrorResponse(w, http.StatusForbidden, "This action cannot be performed during impersonation")
			return
		}
		next.ServeHTTP(w, r)
	})
}
