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

type AdminOrganizationService struct {
	db *database.DB
}

func NewAdminOrganizationService(db *database.DB) *AdminOrganizationService {
	return &AdminOrganizationService{db: db}
}

type ListOrganizationsParams struct {
	Search   string
	IsActive *bool
	Page     int
	Limit    int
}

type CreateOrganizationRequest struct {
	Name           string
	Email          string
	Phone          string
	Address        string
	TaxID          string
	AdminEmail     string
	AdminPassword  string
	AdminFirstName string
	AdminLastName  string
	AdminPhone     string
}

type UpdateOrganizationRequest struct {
	Name    *string
	Email   *string
	Phone   *string
	Address *string
	TaxID   *string
}

func (s *AdminOrganizationService) List(ctx context.Context, params ListOrganizationsParams) ([]models.OrganizationWithStats, int, error) {
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
			o.id, o.name, o.email, o.phone, o.address, o.tax_id, o.logo, o.is_active,
			o.created_at, o.updated_at, o.deleted_at,
			o.suspended_at, o.suspended_by, o.suspend_reason,
			COALESCE((SELECT COUNT(*) FROM users u WHERE u.organization_id = o.id AND u.deleted_at IS NULL), 0) as user_count,
			COALESCE((SELECT COUNT(*) FROM users u WHERE u.organization_id = o.id AND u.is_active = true AND u.deleted_at IS NULL), 0) as active_user_count
		FROM organizations o
		WHERE o.deleted_at IS NULL
	`
	countQuery := `SELECT COUNT(*) FROM organizations o WHERE o.deleted_at IS NULL`

	args := []interface{}{}
	argCount := 0

	if params.Search != "" {
		argCount++
		query += ` AND (o.name ILIKE $` + string(rune('0'+argCount)) + ` OR o.email ILIKE $` + string(rune('0'+argCount)) + `)`
		countQuery += ` AND (o.name ILIKE $` + string(rune('0'+argCount)) + ` OR o.email ILIKE $` + string(rune('0'+argCount)) + `)`
		args = append(args, "%"+params.Search+"%")
	}

	if params.IsActive != nil {
		argCount++
		query += ` AND o.is_active = $` + string(rune('0'+argCount))
		countQuery += ` AND o.is_active = $` + string(rune('0'+argCount))
		args = append(args, *params.IsActive)
	}

	query += ` ORDER BY o.created_at DESC`
	query += ` LIMIT $` + string(rune('0'+argCount+1)) + ` OFFSET $` + string(rune('0'+argCount+2))

	// Get total count
	var total int
	err := s.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Add pagination args
	args = append(args, params.Limit, offset)

	// Get organizations
	rows, err := s.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orgs []models.OrganizationWithStats
	for rows.Next() {
		var org models.OrganizationWithStats
		err := rows.Scan(
			&org.ID, &org.Name, &org.Email, &org.Phone, &org.Address, &org.TaxID, &org.Logo, &org.IsActive,
			&org.CreatedAt, &org.UpdatedAt, &org.DeletedAt,
			&org.SuspendedAt, &org.SuspendedBy, &org.SuspendReason,
			&org.UserCount, &org.ActiveUserCount,
		)
		if err != nil {
			return nil, 0, err
		}
		orgs = append(orgs, org)
	}

	return orgs, total, nil
}

func (s *AdminOrganizationService) GetByID(ctx context.Context, id uuid.UUID) (*models.OrganizationWithStats, error) {
	var org models.OrganizationWithStats

	err := s.db.Pool.QueryRow(ctx, `
		SELECT
			o.id, o.name, o.email, o.phone, o.address, o.tax_id, o.logo, o.is_active,
			o.created_at, o.updated_at, o.deleted_at,
			o.suspended_at, o.suspended_by, o.suspend_reason,
			COALESCE((SELECT COUNT(*) FROM users u WHERE u.organization_id = o.id AND u.deleted_at IS NULL), 0) as user_count,
			COALESCE((SELECT COUNT(*) FROM users u WHERE u.organization_id = o.id AND u.is_active = true AND u.deleted_at IS NULL), 0) as active_user_count
		FROM organizations o
		WHERE o.id = $1 AND o.deleted_at IS NULL
	`, id).Scan(
		&org.ID, &org.Name, &org.Email, &org.Phone, &org.Address, &org.TaxID, &org.Logo, &org.IsActive,
		&org.CreatedAt, &org.UpdatedAt, &org.DeletedAt,
		&org.SuspendedAt, &org.SuspendedBy, &org.SuspendReason,
		&org.UserCount, &org.ActiveUserCount,
	)
	if err != nil {
		return nil, err
	}

	// Get enabled modules
	rows, err := s.db.Pool.Query(ctx, `
		SELECT module_name FROM organization_modules
		WHERE organization_id = $1 AND is_enabled = true
	`, id)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var moduleName string
			if err := rows.Scan(&moduleName); err == nil {
				org.EnabledModules = append(org.EnabledModules, moduleName)
			}
		}
	}

	return &org, nil
}

func (s *AdminOrganizationService) Create(ctx context.Context, adminID uuid.UUID, req CreateOrganizationRequest) (*models.Organization, error) {
	// Hash admin password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Start transaction
	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Create organization
	orgID := uuid.New()
	var org models.Organization
	err = tx.QueryRow(ctx, `
		INSERT INTO organizations (id, name, email, phone, address, tax_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, true, $7, $7)
		RETURNING id, name, email, phone, address, tax_id, logo, is_active, created_at, updated_at
	`, orgID, req.Name, req.Email, req.Phone, req.Address, req.TaxID, time.Now()).Scan(
		&org.ID, &org.Name, &org.Email, &org.Phone, &org.Address, &org.TaxID, &org.Logo, &org.IsActive, &org.CreatedAt, &org.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Create admin user for the organization
	userID := uuid.New()
	_, err = tx.Exec(ctx, `
		INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, phone, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, true, $9, $9)
	`, userID, orgID, req.AdminEmail, string(hashedPassword), req.AdminFirstName, req.AdminLastName, req.AdminPhone, models.RoleAdmin, time.Now())
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &org, nil
}

func (s *AdminOrganizationService) Update(ctx context.Context, id uuid.UUID, req UpdateOrganizationRequest) (*models.Organization, error) {
	var org models.Organization

	// Build update query dynamically
	query := "UPDATE organizations SET updated_at = $1"
	args := []interface{}{time.Now()}
	argIdx := 2

	if req.Name != nil {
		query += ", name = $" + string(rune('0'+argIdx))
		args = append(args, *req.Name)
		argIdx++
	}
	if req.Email != nil {
		query += ", email = $" + string(rune('0'+argIdx))
		args = append(args, *req.Email)
		argIdx++
	}
	if req.Phone != nil {
		query += ", phone = $" + string(rune('0'+argIdx))
		args = append(args, *req.Phone)
		argIdx++
	}
	if req.Address != nil {
		query += ", address = $" + string(rune('0'+argIdx))
		args = append(args, *req.Address)
		argIdx++
	}
	if req.TaxID != nil {
		query += ", tax_id = $" + string(rune('0'+argIdx))
		args = append(args, *req.TaxID)
		argIdx++
	}

	query += " WHERE id = $" + string(rune('0'+argIdx)) + " AND deleted_at IS NULL"
	query += " RETURNING id, name, email, phone, address, tax_id, logo, is_active, created_at, updated_at"
	args = append(args, id)

	err := s.db.Pool.QueryRow(ctx, query, args...).Scan(
		&org.ID, &org.Name, &org.Email, &org.Phone, &org.Address, &org.TaxID, &org.Logo, &org.IsActive, &org.CreatedAt, &org.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &org, nil
}

func (s *AdminOrganizationService) Suspend(ctx context.Context, id uuid.UUID, adminID uuid.UUID, reason string) error {
	result, err := s.db.Pool.Exec(ctx, `
		UPDATE organizations
		SET is_active = false, suspended_at = $1, suspended_by = $2, suspend_reason = $3, updated_at = $1
		WHERE id = $4 AND deleted_at IS NULL
	`, time.Now(), adminID, reason, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("organization not found")
	}

	return nil
}

func (s *AdminOrganizationService) Reactivate(ctx context.Context, id uuid.UUID) error {
	result, err := s.db.Pool.Exec(ctx, `
		UPDATE organizations
		SET is_active = true, suspended_at = NULL, suspended_by = NULL, suspend_reason = NULL, updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`, time.Now(), id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("organization not found")
	}

	return nil
}

func (s *AdminOrganizationService) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := s.db.Pool.Exec(ctx, `
		UPDATE organizations SET deleted_at = $1, updated_at = $1 WHERE id = $2 AND deleted_at IS NULL
	`, time.Now(), id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("organization not found")
	}

	return nil
}
