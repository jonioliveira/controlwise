package models

import (
	"time"

	"github.com/google/uuid"
)

// SystemAdmin represents a platform-level administrator (independent from organizations)
type SystemAdmin struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	Email        string     `json:"email" db:"email"`
	PasswordHash string     `json:"-" db:"password_hash"`
	FirstName    string     `json:"first_name" db:"first_name"`
	LastName     string     `json:"last_name" db:"last_name"`
	IsActive     bool       `json:"is_active" db:"is_active"`
	LastLoginAt  *time.Time `json:"last_login_at" db:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// FullName returns the admin's full name
func (a *SystemAdmin) FullName() string {
	return a.FirstName + " " + a.LastName
}

// SystemAdminAuditLog represents an audit log entry for admin actions
type SystemAdminAuditLog struct {
	ID         uuid.UUID              `json:"id" db:"id"`
	AdminID    uuid.UUID              `json:"admin_id" db:"admin_id"`
	Action     string                 `json:"action" db:"action"`
	EntityType string                 `json:"entity_type" db:"entity_type"`
	EntityID   *uuid.UUID             `json:"entity_id" db:"entity_id"`
	Details    map[string]interface{} `json:"details" db:"details"`
	IPAddress  string                 `json:"ip_address" db:"ip_address"`
	UserAgent  string                 `json:"user_agent" db:"user_agent"`
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
}

// AuditAction constants
type AuditAction string

const (
	// Generic CRUD actions
	AuditActionCreate     AuditAction = "create"
	AuditActionUpdate     AuditAction = "update"
	AuditActionDelete     AuditAction = "delete"
	AuditActionSuspend    AuditAction = "suspend"
	AuditActionReactivate AuditAction = "reactivate"
	AuditActionView       AuditAction = "view"

	// Specific actions
	AuditActionAdminLogin         AuditAction = "admin_login"
	AuditActionAdminLogout        AuditAction = "admin_logout"
	AuditActionUserImpersonated   AuditAction = "user_impersonated"
	AuditActionImpersonationEnded AuditAction = "impersonation_ended"
	AuditActionSettingUpdated     AuditAction = "setting_updated"
)

// AuditEntityType constants
type AuditEntityType string

const (
	AuditEntityOrganization AuditEntityType = "organization"
	AuditEntityUser         AuditEntityType = "user"
	AuditEntitySetting      AuditEntityType = "setting"
	AuditEntityAdmin        AuditEntityType = "admin"
)

// ImpersonationSession represents an admin impersonation session
type ImpersonationSession struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	AdminID            uuid.UUID  `json:"admin_id" db:"admin_id"`
	ImpersonatedUserID uuid.UUID  `json:"impersonated_user_id" db:"impersonated_user_id"`
	StartedAt          time.Time  `json:"started_at" db:"started_at"`
	EndedAt            *time.Time `json:"ended_at" db:"ended_at"`
	Reason             string     `json:"reason" db:"reason"`
	IPAddress          string     `json:"ip_address" db:"ip_address"`
}

// ImpersonationSessionWithDetails includes related user info
type ImpersonationSessionWithDetails struct {
	ImpersonationSession
	UserEmail    string `json:"user_email"`
	UserName     string `json:"user_name"`
	OrgName      string `json:"org_name"`
	AdminName    string `json:"admin_name"`
}

// SystemSetting represents a global system setting
type SystemSetting struct {
	Key         string                 `json:"key" db:"key"`
	Value       map[string]interface{} `json:"value" db:"value"`
	Description string                 `json:"description" db:"description"`
	UpdatedBy   *uuid.UUID             `json:"updated_by" db:"updated_by"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// Common system setting keys
const (
	SettingMaintenanceMode    = "maintenance_mode"
	SettingRegistrationEnabled = "registration_enabled"
	SettingMaxOrganizations    = "max_organizations"
	SettingDefaultModules      = "default_modules"
)

// OrganizationWithStats includes organization with admin-specific stats
type OrganizationWithStats struct {
	Organization
	UserCount       int       `json:"user_count" db:"user_count"`
	ActiveUserCount int       `json:"active_user_count" db:"active_user_count"`
	EnabledModules  []string  `json:"enabled_modules"`
	SuspendedAt     *time.Time `json:"suspended_at" db:"suspended_at"`
	SuspendedBy     *uuid.UUID `json:"suspended_by" db:"suspended_by"`
	SuspendReason   *string    `json:"suspend_reason" db:"suspend_reason"`
}

// UserWithOrg includes user with organization info for admin views
type UserWithOrg struct {
	User
	OrgName       string     `json:"org_name" db:"org_name"`
	SuspendedAt   *time.Time `json:"suspended_at" db:"suspended_at"`
	SuspendedBy   *uuid.UUID `json:"suspended_by" db:"suspended_by"`
	SuspendReason *string    `json:"suspend_reason" db:"suspend_reason"`
}

// PlatformStats represents platform-wide statistics
type PlatformStats struct {
	TotalOrganizations   int            `json:"total_organizations"`
	ActiveOrganizations  int            `json:"active_organizations"`
	SuspendedOrganizations int          `json:"suspended_organizations"`
	TotalUsers           int            `json:"total_users"`
	ActiveUsers          int            `json:"active_users"`
	NewOrgsThisMonth     int            `json:"new_orgs_this_month"`
	NewUsersThisMonth    int            `json:"new_users_this_month"`
	OrgsByModule         map[string]int `json:"orgs_by_module"`
}

// SystemAdminAuthResponse is the response after admin login
type SystemAdminAuthResponse struct {
	Token string       `json:"token"`
	Admin *SystemAdmin `json:"admin"`
}

// ImpersonationToken is the response after starting impersonation
type ImpersonationToken struct {
	Token     string    `json:"token"`
	SessionID uuid.UUID `json:"session_id"`
	User      *User     `json:"user"`
	ExpiresAt time.Time `json:"expires_at"`
}
