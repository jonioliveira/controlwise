package services

import (
	"context"
	"errors"
	"time"

	"github.com/controlewise/backend/internal/database"
	"github.com/controlewise/backend/internal/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AdminUserService struct {
	db *database.DB
}

func NewAdminUserService(db *database.DB) *AdminUserService {
	return &AdminUserService{db: db}
}

type ListUsersParams struct {
	Search         string
	OrganizationID *uuid.UUID
	IsActive       *bool
	Role           string
	Page           int
	Limit          int
}

func (s *AdminUserService) List(ctx context.Context, params ListUsersParams) ([]models.UserWithOrg, int, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 || params.Limit > 100 {
		params.Limit = 20
	}
	offset := (params.Page - 1) * params.Limit

	// Build query
	query := `
		SELECT
			u.id, u.organization_id, u.email, u.password_hash, u.first_name, u.last_name,
			u.phone, u.avatar, u.role, u.is_active, u.last_login_at,
			u.created_at, u.updated_at, u.deleted_at,
			o.name as org_name,
			u.suspended_at, u.suspended_by, u.suspend_reason
		FROM users u
		JOIN organizations o ON u.organization_id = o.id
		WHERE u.deleted_at IS NULL
	`
	countQuery := `SELECT COUNT(*) FROM users u WHERE u.deleted_at IS NULL`
	args := []interface{}{}
	argCount := 0

	if params.Search != "" {
		argCount++
		query += ` AND (u.email ILIKE $` + string(rune('0'+argCount)) + ` OR u.first_name ILIKE $` + string(rune('0'+argCount)) + ` OR u.last_name ILIKE $` + string(rune('0'+argCount)) + `)`
		countQuery += ` AND (u.email ILIKE $` + string(rune('0'+argCount)) + ` OR u.first_name ILIKE $` + string(rune('0'+argCount)) + ` OR u.last_name ILIKE $` + string(rune('0'+argCount)) + `)`
		args = append(args, "%"+params.Search+"%")
	}

	if params.OrganizationID != nil {
		argCount++
		query += ` AND u.organization_id = $` + string(rune('0'+argCount))
		countQuery += ` AND u.organization_id = $` + string(rune('0'+argCount))
		args = append(args, *params.OrganizationID)
	}

	if params.IsActive != nil {
		argCount++
		query += ` AND u.is_active = $` + string(rune('0'+argCount))
		countQuery += ` AND u.is_active = $` + string(rune('0'+argCount))
		args = append(args, *params.IsActive)
	}

	if params.Role != "" {
		argCount++
		query += ` AND u.role = $` + string(rune('0'+argCount))
		countQuery += ` AND u.role = $` + string(rune('0'+argCount))
		args = append(args, params.Role)
	}

	// Get total count
	var total int
	err := s.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query += ` ORDER BY u.created_at DESC`
	query += ` LIMIT $` + string(rune('0'+argCount+1)) + ` OFFSET $` + string(rune('0'+argCount+2))
	args = append(args, params.Limit, offset)

	rows, err := s.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []models.UserWithOrg
	for rows.Next() {
		var user models.UserWithOrg
		err := rows.Scan(
			&user.ID, &user.OrganizationID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName,
			&user.Phone, &user.Avatar, &user.Role, &user.IsActive, &user.LastLoginAt,
			&user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
			&user.OrgName,
			&user.SuspendedAt, &user.SuspendedBy, &user.SuspendReason,
		)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}

	return users, total, nil
}

func (s *AdminUserService) GetByID(ctx context.Context, id uuid.UUID) (*models.UserWithOrg, error) {
	var user models.UserWithOrg

	err := s.db.Pool.QueryRow(ctx, `
		SELECT
			u.id, u.organization_id, u.email, u.password_hash, u.first_name, u.last_name,
			u.phone, u.avatar, u.role, u.is_active, u.last_login_at,
			u.created_at, u.updated_at, u.deleted_at,
			o.name as org_name,
			u.suspended_at, u.suspended_by, u.suspend_reason
		FROM users u
		JOIN organizations o ON u.organization_id = o.id
		WHERE u.id = $1 AND u.deleted_at IS NULL
	`, id).Scan(
		&user.ID, &user.OrganizationID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName,
		&user.Phone, &user.Avatar, &user.Role, &user.IsActive, &user.LastLoginAt,
		&user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
		&user.OrgName,
		&user.SuspendedAt, &user.SuspendedBy, &user.SuspendReason,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *AdminUserService) ListByOrganization(ctx context.Context, orgID uuid.UUID, page, limit int) ([]models.UserWithOrg, int, error) {
	params := ListUsersParams{
		OrganizationID: &orgID,
		Page:           page,
		Limit:          limit,
	}
	return s.List(ctx, params)
}

func (s *AdminUserService) Suspend(ctx context.Context, id uuid.UUID, adminID uuid.UUID, reason string) error {
	result, err := s.db.Pool.Exec(ctx, `
		UPDATE users
		SET is_active = false, suspended_at = $1, suspended_by = $2, suspend_reason = $3, updated_at = $1
		WHERE id = $4 AND deleted_at IS NULL
	`, time.Now(), adminID, reason, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("user not found")
	}

	return nil
}

func (s *AdminUserService) Reactivate(ctx context.Context, id uuid.UUID) error {
	result, err := s.db.Pool.Exec(ctx, `
		UPDATE users
		SET is_active = true, suspended_at = NULL, suspended_by = NULL, suspend_reason = NULL, updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`, time.Now(), id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("user not found")
	}

	return nil
}

func (s *AdminUserService) ResetPassword(ctx context.Context, id uuid.UUID, newPassword string) error {
	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}

	result, err := s.db.Pool.Exec(ctx, `
		UPDATE users
		SET password_hash = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`, string(hashedPassword), time.Now(), id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("user not found")
	}

	return nil
}
