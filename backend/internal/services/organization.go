package services

import (
	"context"

	"github.com/controlewise/backend/internal/database"
	"github.com/controlewise/backend/internal/models"
	"github.com/google/uuid"
)

type OrganizationService struct {
	db *database.DB
}

func NewOrganizationService(db *database.DB) *OrganizationService {
	return &OrganizationService{db: db}
}

func (s *OrganizationService) GetByID(ctx context.Context, id uuid.UUID) (*models.Organization, error) {
	var org models.Organization
	err := s.db.Pool.QueryRow(ctx, `
		SELECT id, name, email, COALESCE(phone, ''), COALESCE(address, ''), COALESCE(tax_id, ''), logo, is_active, created_at, updated_at
		FROM organizations
		WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(
		&org.ID,
		&org.Name,
		&org.Email,
		&org.Phone,
		&org.Address,
		&org.TaxID,
		&org.Logo,
		&org.IsActive,
		&org.CreatedAt,
		&org.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &org, nil
}

func (s *OrganizationService) Update(ctx context.Context, id uuid.UUID, org *models.Organization) error {
	_, err := s.db.Pool.Exec(ctx, `
		UPDATE organizations
		SET name = $1, email = $2, phone = $3, address = $4, tax_id = $5
		WHERE id = $6 AND deleted_at IS NULL
	`, org.Name, org.Email, org.Phone, org.Address, org.TaxID, id)
	return err
}
