package services

import (
	"context"
	"errors"
	"time"

	"github.com/controlwise/backend/internal/config"
	"github.com/controlwise/backend/internal/database"
	"github.com/controlwise/backend/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type SystemAdminService struct {
	db     *database.DB
	jwtCfg config.JWTConfig
}

func NewSystemAdminService(db *database.DB, jwtCfg config.JWTConfig) *SystemAdminService {
	return &SystemAdminService{
		db:     db,
		jwtCfg: jwtCfg,
	}
}

type AdminLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AdminChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func (s *SystemAdminService) Login(ctx context.Context, req AdminLoginRequest) (*models.SystemAdminAuthResponse, error) {
	var admin models.SystemAdmin
	err := s.db.Pool.QueryRow(ctx, `
		SELECT id, email, password_hash, first_name, last_name, is_active, last_login_at, created_at, updated_at
		FROM system_admins
		WHERE email = $1 AND deleted_at IS NULL
	`, req.Email).Scan(
		&admin.ID,
		&admin.Email,
		&admin.PasswordHash,
		&admin.FirstName,
		&admin.LastName,
		&admin.IsActive,
		&admin.LastLoginAt,
		&admin.CreatedAt,
		&admin.UpdatedAt,
	)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if !admin.IsActive {
		return nil, errors.New("admin account is not active")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Update last login
	_, err = s.db.Pool.Exec(ctx, `
		UPDATE system_admins SET last_login_at = $1, updated_at = $2 WHERE id = $3
	`, time.Now(), time.Now(), admin.ID)
	if err != nil {
		// Log error but don't fail login
	}

	// Generate token
	token, err := s.generateToken(&admin)
	if err != nil {
		return nil, err
	}

	return &models.SystemAdminAuthResponse{
		Token: token,
		Admin: &admin,
	}, nil
}

func (s *SystemAdminService) GetByID(ctx context.Context, adminID uuid.UUID) (*models.SystemAdmin, error) {
	var admin models.SystemAdmin
	err := s.db.Pool.QueryRow(ctx, `
		SELECT id, email, first_name, last_name, is_active, last_login_at, created_at, updated_at
		FROM system_admins
		WHERE id = $1 AND deleted_at IS NULL
	`, adminID).Scan(
		&admin.ID,
		&admin.Email,
		&admin.FirstName,
		&admin.LastName,
		&admin.IsActive,
		&admin.LastLoginAt,
		&admin.CreatedAt,
		&admin.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &admin, nil
}

func (s *SystemAdminService) GetByEmail(ctx context.Context, email string) (*models.SystemAdmin, error) {
	var admin models.SystemAdmin
	err := s.db.Pool.QueryRow(ctx, `
		SELECT id, email, first_name, last_name, is_active, last_login_at, created_at, updated_at
		FROM system_admins
		WHERE email = $1 AND deleted_at IS NULL
	`, email).Scan(
		&admin.ID,
		&admin.Email,
		&admin.FirstName,
		&admin.LastName,
		&admin.IsActive,
		&admin.LastLoginAt,
		&admin.CreatedAt,
		&admin.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &admin, nil
}

func (s *SystemAdminService) ChangePassword(ctx context.Context, adminID uuid.UUID, req AdminChangePasswordRequest) error {
	// Get current password hash
	var currentHash string
	err := s.db.Pool.QueryRow(ctx, `
		SELECT password_hash FROM system_admins WHERE id = $1 AND deleted_at IS NULL
	`, adminID).Scan(&currentHash)
	if err != nil {
		return errors.New("admin not found")
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(req.OldPassword)); err != nil {
		return errors.New("invalid current password")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update password
	_, err = s.db.Pool.Exec(ctx, `
		UPDATE system_admins SET password_hash = $1, updated_at = $2 WHERE id = $3
	`, string(hashedPassword), time.Now(), adminID)
	if err != nil {
		return err
	}

	return nil
}

func (s *SystemAdminService) generateToken(admin *models.SystemAdmin) (string, error) {
	claims := jwt.MapClaims{
		"admin_id":    admin.ID.String(),
		"is_sysadmin": true,
		"exp":         time.Now().Add(s.jwtCfg.Expiry).Unix(),
		"iat":         time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtCfg.Secret))
}

// GenerateImpersonationToken creates a token for impersonating a user
func (s *SystemAdminService) GenerateImpersonationToken(adminID uuid.UUID, user *models.User, sessionID uuid.UUID) (string, time.Time, error) {
	// Impersonation tokens have shorter expiry (1 hour)
	expiry := time.Now().Add(1 * time.Hour)

	claims := jwt.MapClaims{
		"user_id":                  user.ID.String(),
		"organization_id":          user.OrganizationID.String(),
		"role":                     user.Role,
		"is_impersonation":         true,
		"impersonator_id":          adminID.String(),
		"impersonation_session_id": sessionID.String(),
		"exp":                      expiry.Unix(),
		"iat":                      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtCfg.Secret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiry, nil
}
