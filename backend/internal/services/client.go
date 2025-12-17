package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/controlwise/backend/internal/database"
	"github.com/controlwise/backend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ClientService struct {
	db *database.DB
}

func NewClientService(db *database.DB) *ClientService {
	return &ClientService{db: db}
}

// List returns all clients for an organization with pagination
func (s *ClientService) List(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.Client, int, error) {
	// Get total count
	var total int
	err := s.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM clients 
		WHERE organization_id = $1 AND deleted_at IS NULL
	`, orgID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count clients: %w", err)
	}

	// Get clients
	rows, err := s.db.Pool.Query(ctx, `
		SELECT 
			id, organization_id, name, email, phone, address, tax_id, 
			notes, user_id, created_by, created_at, updated_at
		FROM clients
		WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, orgID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query clients: %w", err)
	}
	defer rows.Close()

	var clients []*models.Client
	for rows.Next() {
		var c models.Client
		err := rows.Scan(
			&c.ID,
			&c.OrganizationID,
			&c.Name,
			&c.Email,
			&c.Phone,
			&c.Address,
			&c.TaxID,
			&c.Notes,
			&c.UserID,
			&c.CreatedBy,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan client: %w", err)
		}
		clients = append(clients, &c)
	}

	return clients, total, nil
}

// Search clients by name or email
func (s *ClientService) Search(ctx context.Context, orgID uuid.UUID, query string, limit int) ([]*models.Client, error) {
	searchPattern := "%" + query + "%"
	
	rows, err := s.db.Pool.Query(ctx, `
		SELECT 
			id, organization_id, name, email, phone, address, tax_id, 
			notes, user_id, created_by, created_at, updated_at
		FROM clients
		WHERE organization_id = $1 
			AND deleted_at IS NULL
			AND (name ILIKE $2 OR email ILIKE $2)
		ORDER BY name
		LIMIT $3
	`, orgID, searchPattern, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search clients: %w", err)
	}
	defer rows.Close()

	var clients []*models.Client
	for rows.Next() {
		var c models.Client
		err := rows.Scan(
			&c.ID,
			&c.OrganizationID,
			&c.Name,
			&c.Email,
			&c.Phone,
			&c.Address,
			&c.TaxID,
			&c.Notes,
			&c.UserID,
			&c.CreatedBy,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan client: %w", err)
		}
		clients = append(clients, &c)
	}

	return clients, nil
}

// GetByID returns a single client by ID
func (s *ClientService) GetByID(ctx context.Context, id, orgID uuid.UUID) (*models.Client, error) {
	var c models.Client
	err := s.db.Pool.QueryRow(ctx, `
		SELECT 
			id, organization_id, name, email, phone, address, tax_id, 
			notes, user_id, created_by, created_at, updated_at
		FROM clients
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL
	`, id, orgID).Scan(
		&c.ID,
		&c.OrganizationID,
		&c.Name,
		&c.Email,
		&c.Phone,
		&c.Address,
		&c.TaxID,
		&c.Notes,
		&c.UserID,
		&c.CreatedBy,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("client not found")
		}
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	return &c, nil
}

// Create creates a new client
func (s *ClientService) Create(ctx context.Context, client *models.Client) error {
	// Validate required fields
	if client.Name == "" {
		return errors.New("client name is required")
	}
	if client.Email == "" {
		return errors.New("client email is required")
	}
	if client.Phone == "" {
		return errors.New("client phone is required")
	}

	// Check if email already exists for this organization
	var exists bool
	err := s.db.Pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM clients 
			WHERE organization_id = $1 AND email = $2 AND deleted_at IS NULL
		)
	`, client.OrganizationID, client.Email).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return errors.New("client with this email already exists")
	}

	// Generate ID
	client.ID = uuid.New()

	// Insert client
	_, err = s.db.Pool.Exec(ctx, `
		INSERT INTO clients (
			id, organization_id, name, email, phone, address, tax_id, 
			notes, user_id, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, client.ID, client.OrganizationID, client.Name, client.Email, client.Phone,
		client.Address, client.TaxID, client.Notes, client.UserID, client.CreatedBy)
	
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	return nil
}

// Update updates an existing client
func (s *ClientService) Update(ctx context.Context, id, orgID uuid.UUID, client *models.Client) error {
	// Validate required fields
	if client.Name == "" {
		return errors.New("client name is required")
	}
	if client.Email == "" {
		return errors.New("client email is required")
	}
	if client.Phone == "" {
		return errors.New("client phone is required")
	}

	// Check if client exists and belongs to organization
	var exists bool
	err := s.db.Pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM clients 
			WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL
		)
	`, id, orgID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check client existence: %w", err)
	}
	if !exists {
		return errors.New("client not found")
	}

	// Check if email is taken by another client
	var emailTaken bool
	err = s.db.Pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM clients 
			WHERE organization_id = $1 AND email = $2 AND id != $3 AND deleted_at IS NULL
		)
	`, orgID, client.Email, id).Scan(&emailTaken)
	if err != nil {
		return fmt.Errorf("failed to check email: %w", err)
	}
	if emailTaken {
		return errors.New("email already in use by another client")
	}

	// Update client
	result, err := s.db.Pool.Exec(ctx, `
		UPDATE clients
		SET name = $1, email = $2, phone = $3, address = $4, 
		    tax_id = $5, notes = $6
		WHERE id = $7 AND organization_id = $8 AND deleted_at IS NULL
	`, client.Name, client.Email, client.Phone, client.Address,
		client.TaxID, client.Notes, id, orgID)
	
	if err != nil {
		return fmt.Errorf("failed to update client: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("client not found or already deleted")
	}

	return nil
}

// Delete soft deletes a client
func (s *ClientService) Delete(ctx context.Context, id, orgID uuid.UUID) error {
	// Check if client has any worksheets
	var hasWorksheets bool
	err := s.db.Pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM worksheets 
			WHERE client_id = $1 AND deleted_at IS NULL
		)
	`, id).Scan(&hasWorksheets)
	if err != nil {
		return fmt.Errorf("failed to check worksheets: %w", err)
	}
	if hasWorksheets {
		return errors.New("cannot delete client with existing worksheets")
	}

	// Soft delete
	result, err := s.db.Pool.Exec(ctx, `
		UPDATE clients
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL
	`, id, orgID)
	
	if err != nil {
		return fmt.Errorf("failed to delete client: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("client not found or already deleted")
	}

	return nil
}

// GetStats returns client statistics for dashboard
func (s *ClientService) GetStats(ctx context.Context, orgID uuid.UUID) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total clients
	var total int
	err := s.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM clients 
		WHERE organization_id = $1 AND deleted_at IS NULL
	`, orgID).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count clients: %w", err)
	}
	stats["total"] = total

	// Clients added this month
	var thisMonth int
	err = s.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM clients 
		WHERE organization_id = $1 
			AND deleted_at IS NULL
			AND created_at >= DATE_TRUNC('month', CURRENT_DATE)
	`, orgID).Scan(&thisMonth)
	if err != nil {
		return nil, fmt.Errorf("failed to count monthly clients: %w", err)
	}
	stats["this_month"] = thisMonth

	// Clients with active projects
	var withProjects int
	err = s.db.Pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT c.id) 
		FROM clients c
		INNER JOIN worksheets w ON w.client_id = c.id
		INNER JOIN budgets b ON b.worksheet_id = w.id
		INNER JOIN projects p ON p.budget_id = b.id
		WHERE c.organization_id = $1 
			AND c.deleted_at IS NULL
			AND p.status IN ('in_progress', 'on_hold')
	`, orgID).Scan(&withProjects)
	if err != nil {
		return nil, fmt.Errorf("failed to count clients with projects: %w", err)
	}
	stats["with_active_projects"] = withProjects

	return stats, nil
}
