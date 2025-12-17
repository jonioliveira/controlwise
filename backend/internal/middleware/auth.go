package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/controlwise/backend/internal/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const (
	UserIDKey           contextKey = "user_id"
	OrganizationIDKey   contextKey = "organization_id"
	UserRoleKey         contextKey = "user_role"
	// System admin context keys
	IsSystemAdminKey           contextKey = "is_system_admin"
	SystemAdminIDKey           contextKey = "system_admin_id"
	// Impersonation context keys
	IsImpersonationKey         contextKey = "is_impersonation"
	ImpersonatorIDKey          contextKey = "impersonator_id"
	ImpersonationSessionIDKey  contextKey = "impersonation_session_id"
)

type AuthMiddleware struct {
	jwtSecret string
}

func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret: jwtSecret,
	}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.ErrorResponse(w, http.StatusUnauthorized, "Missing authorization header")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.ErrorResponse(w, http.StatusUnauthorized, "Invalid authorization header format")
			return
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(m.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			utils.ErrorResponse(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			utils.ErrorResponse(w, http.StatusUnauthorized, "Invalid token claims")
			return
		}

		ctx := r.Context()

		// Check if this is a system admin token
		if isSysAdmin, ok := claims["is_sysadmin"].(bool); ok && isSysAdmin {
			adminIDStr, ok := claims["admin_id"].(string)
			if !ok {
				utils.ErrorResponse(w, http.StatusUnauthorized, "Invalid admin ID in token")
				return
			}
			adminID, err := uuid.Parse(adminIDStr)
			if err != nil {
				utils.ErrorResponse(w, http.StatusUnauthorized, "Invalid admin ID in token")
				return
			}

			ctx = context.WithValue(ctx, IsSystemAdminKey, true)
			ctx = context.WithValue(ctx, SystemAdminIDKey, adminID)

			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Regular user token (or impersonation token)
		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			utils.ErrorResponse(w, http.StatusUnauthorized, "Invalid user ID in token")
			return
		}
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			utils.ErrorResponse(w, http.StatusUnauthorized, "Invalid user ID in token")
			return
		}

		orgIDStr, ok := claims["organization_id"].(string)
		if !ok {
			utils.ErrorResponse(w, http.StatusUnauthorized, "Invalid organization ID in token")
			return
		}
		organizationID, err := uuid.Parse(orgIDStr)
		if err != nil {
			utils.ErrorResponse(w, http.StatusUnauthorized, "Invalid organization ID in token")
			return
		}

		role, _ := claims["role"].(string)

		ctx = context.WithValue(ctx, UserIDKey, userID)
		ctx = context.WithValue(ctx, OrganizationIDKey, organizationID)
		ctx = context.WithValue(ctx, UserRoleKey, role)
		ctx = context.WithValue(ctx, IsSystemAdminKey, false)

		// Check if this is an impersonation token
		if isImpersonation, ok := claims["is_impersonation"].(bool); ok && isImpersonation {
			ctx = context.WithValue(ctx, IsImpersonationKey, true)

			if impersonatorIDStr, ok := claims["impersonator_id"].(string); ok {
				if impersonatorID, err := uuid.Parse(impersonatorIDStr); err == nil {
					ctx = context.WithValue(ctx, ImpersonatorIDKey, impersonatorID)
				}
			}

			if sessionIDStr, ok := claims["impersonation_session_id"].(string); ok {
				if sessionID, err := uuid.Parse(sessionIDStr); err == nil {
					ctx = context.WithValue(ctx, ImpersonationSessionIDKey, sessionID)
				}
			}
		} else {
			ctx = context.WithValue(ctx, IsImpersonationKey, false)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID returns the user ID from context
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return userID, ok
}

// GetOrganizationID returns the organization ID from context
func GetOrganizationID(ctx context.Context) (uuid.UUID, bool) {
	orgID, ok := ctx.Value(OrganizationIDKey).(uuid.UUID)
	return orgID, ok
}

// GetUserRole returns the user role from context
func GetUserRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(UserRoleKey).(string)
	return role, ok
}

// IsSysAdmin returns true if the request is from a system admin
func IsSysAdmin(ctx context.Context) bool {
	isAdmin, ok := ctx.Value(IsSystemAdminKey).(bool)
	return ok && isAdmin
}

// GetSystemAdminID returns the system admin ID from context
func GetSystemAdminID(ctx context.Context) (uuid.UUID, bool) {
	adminID, ok := ctx.Value(SystemAdminIDKey).(uuid.UUID)
	return adminID, ok
}

// IsImpersonation returns true if the request is from an impersonation session
func IsImpersonation(ctx context.Context) bool {
	isImp, ok := ctx.Value(IsImpersonationKey).(bool)
	return ok && isImp
}

// GetImpersonatorID returns the impersonator admin ID from context
func GetImpersonatorID(ctx context.Context) (uuid.UUID, bool) {
	impID, ok := ctx.Value(ImpersonatorIDKey).(uuid.UUID)
	return impID, ok
}

// GetImpersonationSessionID returns the impersonation session ID from context
func GetImpersonationSessionID(ctx context.Context) (uuid.UUID, bool) {
	sessionID, ok := ctx.Value(ImpersonationSessionIDKey).(uuid.UUID)
	return sessionID, ok
}
