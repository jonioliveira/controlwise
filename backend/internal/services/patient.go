package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/controlewise/backend/internal/database"
	"github.com/controlewise/backend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type PatientService struct {
	db *database.DB
}

func NewPatientService(db *database.DB) *PatientService {
	return &PatientService{db: db}
}

// List returns all patients for an organization with pagination
// Joins with clients table to get client info (name, email, phone)
func (s *PatientService) List(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.PatientWithClient, int, error) {
	// Get total count
	var total int
	err := s.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM patients
		WHERE organization_id = $1 AND deleted_at IS NULL
	`, orgID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count patients: %w", err)
	}

	// Get patients with client data
	rows, err := s.db.Pool.Query(ctx, `
		SELECT
			p.id, p.organization_id, p.client_id, p.date_of_birth,
			p.notes, p.emergency_contact, p.emergency_phone, p.is_active,
			p.created_by, p.created_at, p.updated_at,
			c.name as client_name, c.email as client_email, c.phone as client_phone
		FROM patients p
		INNER JOIN clients c ON c.id = p.client_id
		WHERE p.organization_id = $1 AND p.deleted_at IS NULL
		ORDER BY c.name ASC
		LIMIT $2 OFFSET $3
	`, orgID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query patients: %w", err)
	}
	defer rows.Close()

	var patients []*models.PatientWithClient
	for rows.Next() {
		var p models.PatientWithClient
		err := rows.Scan(
			&p.ID,
			&p.OrganizationID,
			&p.ClientID,
			&p.DateOfBirth,
			&p.Notes,
			&p.EmergencyContact,
			&p.EmergencyPhone,
			&p.IsActive,
			&p.CreatedBy,
			&p.CreatedAt,
			&p.UpdatedAt,
			&p.ClientName,
			&p.ClientEmail,
			&p.ClientPhone,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan patient: %w", err)
		}
		patients = append(patients, &p)
	}

	return patients, total, nil
}

// Search patients by client name, phone, or email
func (s *PatientService) Search(ctx context.Context, orgID uuid.UUID, query string, limit int) ([]*models.PatientWithClient, error) {
	searchPattern := "%" + query + "%"

	rows, err := s.db.Pool.Query(ctx, `
		SELECT
			p.id, p.organization_id, p.client_id, p.date_of_birth,
			p.notes, p.emergency_contact, p.emergency_phone, p.is_active,
			p.created_by, p.created_at, p.updated_at,
			c.name as client_name, c.email as client_email, c.phone as client_phone
		FROM patients p
		INNER JOIN clients c ON c.id = p.client_id
		WHERE p.organization_id = $1
			AND p.deleted_at IS NULL
			AND (c.name ILIKE $2 OR c.phone ILIKE $2 OR c.email ILIKE $2)
		ORDER BY c.name
		LIMIT $3
	`, orgID, searchPattern, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search patients: %w", err)
	}
	defer rows.Close()

	var patients []*models.PatientWithClient
	for rows.Next() {
		var p models.PatientWithClient
		err := rows.Scan(
			&p.ID,
			&p.OrganizationID,
			&p.ClientID,
			&p.DateOfBirth,
			&p.Notes,
			&p.EmergencyContact,
			&p.EmergencyPhone,
			&p.IsActive,
			&p.CreatedBy,
			&p.CreatedAt,
			&p.UpdatedAt,
			&p.ClientName,
			&p.ClientEmail,
			&p.ClientPhone,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan patient: %w", err)
		}
		patients = append(patients, &p)
	}

	return patients, nil
}

// GetByID returns a single patient by ID with client data
func (s *PatientService) GetByID(ctx context.Context, id, orgID uuid.UUID) (*models.PatientWithClient, error) {
	var p models.PatientWithClient
	err := s.db.Pool.QueryRow(ctx, `
		SELECT
			p.id, p.organization_id, p.client_id, p.date_of_birth,
			p.notes, p.emergency_contact, p.emergency_phone, p.is_active,
			p.created_by, p.created_at, p.updated_at,
			c.name as client_name, c.email as client_email, c.phone as client_phone
		FROM patients p
		INNER JOIN clients c ON c.id = p.client_id
		WHERE p.id = $1 AND p.organization_id = $2 AND p.deleted_at IS NULL
	`, id, orgID).Scan(
		&p.ID,
		&p.OrganizationID,
		&p.ClientID,
		&p.DateOfBirth,
		&p.Notes,
		&p.EmergencyContact,
		&p.EmergencyPhone,
		&p.IsActive,
		&p.CreatedBy,
		&p.CreatedAt,
		&p.UpdatedAt,
		&p.ClientName,
		&p.ClientEmail,
		&p.ClientPhone,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("patient not found")
		}
		return nil, fmt.Errorf("failed to get patient: %w", err)
	}

	return &p, nil
}

// GetByClientID returns a patient by client ID
func (s *PatientService) GetByClientID(ctx context.Context, orgID, clientID uuid.UUID) (*models.PatientWithClient, error) {
	var p models.PatientWithClient
	err := s.db.Pool.QueryRow(ctx, `
		SELECT
			p.id, p.organization_id, p.client_id, p.date_of_birth,
			p.notes, p.emergency_contact, p.emergency_phone, p.is_active,
			p.created_by, p.created_at, p.updated_at,
			c.name as client_name, c.email as client_email, c.phone as client_phone
		FROM patients p
		INNER JOIN clients c ON c.id = p.client_id
		WHERE p.organization_id = $1 AND p.client_id = $2 AND p.deleted_at IS NULL
	`, orgID, clientID).Scan(
		&p.ID,
		&p.OrganizationID,
		&p.ClientID,
		&p.DateOfBirth,
		&p.Notes,
		&p.EmergencyContact,
		&p.EmergencyPhone,
		&p.IsActive,
		&p.CreatedBy,
		&p.CreatedAt,
		&p.UpdatedAt,
		&p.ClientName,
		&p.ClientEmail,
		&p.ClientPhone,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get patient: %w", err)
	}

	return &p, nil
}

// Create creates a new patient linked to a client
func (s *PatientService) Create(ctx context.Context, patient *models.Patient) error {
	// Validate required client_id
	if patient.ClientID == uuid.Nil {
		return errors.New("client_id is required")
	}

	// Verify client exists and belongs to the same organization
	var clientExists bool
	err := s.db.Pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM clients
			WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL
		)
	`, patient.ClientID, patient.OrganizationID).Scan(&clientExists)
	if err != nil {
		return fmt.Errorf("failed to verify client: %w", err)
	}
	if !clientExists {
		return errors.New("client not found")
	}

	// Check if client is already linked to a patient
	existing, err := s.GetByClientID(ctx, patient.OrganizationID, patient.ClientID)
	if err != nil {
		return fmt.Errorf("failed to check existing patient: %w", err)
	}
	if existing != nil {
		return errors.New("client is already linked to a patient")
	}

	// Generate ID
	patient.ID = uuid.New()
	patient.IsActive = true

	// Insert patient
	_, err = s.db.Pool.Exec(ctx, `
		INSERT INTO patients (
			id, organization_id, client_id, date_of_birth,
			notes, emergency_contact, emergency_phone, is_active, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, patient.ID, patient.OrganizationID, patient.ClientID, patient.DateOfBirth,
		patient.Notes, patient.EmergencyContact, patient.EmergencyPhone,
		patient.IsActive, patient.CreatedBy)

	if err != nil {
		return fmt.Errorf("failed to create patient: %w", err)
	}

	return nil
}

// Update updates an existing patient (healthcare fields only, not client link)
func (s *PatientService) Update(ctx context.Context, id, orgID uuid.UUID, patient *models.Patient) error {
	// Update patient healthcare fields (not client_id)
	result, err := s.db.Pool.Exec(ctx, `
		UPDATE patients
		SET date_of_birth = $1, notes = $2, emergency_contact = $3,
		    emergency_phone = $4, is_active = $5, updated_at = CURRENT_TIMESTAMP
		WHERE id = $6 AND organization_id = $7 AND deleted_at IS NULL
	`, patient.DateOfBirth, patient.Notes, patient.EmergencyContact,
		patient.EmergencyPhone, patient.IsActive, id, orgID)

	if err != nil {
		return fmt.Errorf("failed to update patient: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("patient not found or already deleted")
	}

	return nil
}

// Delete soft deletes a patient
func (s *PatientService) Delete(ctx context.Context, id, orgID uuid.UUID) error {
	// Check if patient has any future sessions
	var hasSessions bool
	err := s.db.Pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM sessions
			WHERE patient_id = $1
				AND deleted_at IS NULL
				AND scheduled_at > NOW()
				AND status NOT IN ('cancelled', 'completed')
		)
	`, id).Scan(&hasSessions)
	if err != nil {
		return fmt.Errorf("failed to check sessions: %w", err)
	}
	if hasSessions {
		return errors.New("cannot delete patient with future sessions")
	}

	// Soft delete
	result, err := s.db.Pool.Exec(ctx, `
		UPDATE patients
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL
	`, id, orgID)

	if err != nil {
		return fmt.Errorf("failed to delete patient: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("patient not found or already deleted")
	}

	return nil
}

// GetStats returns patient statistics
func (s *PatientService) GetStats(ctx context.Context, orgID uuid.UUID) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total active patients
	var total int
	err := s.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM patients
		WHERE organization_id = $1 AND deleted_at IS NULL AND is_active = TRUE
	`, orgID).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count patients: %w", err)
	}
	stats["total"] = total

	// Inactive patients
	var inactive int
	err = s.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM patients
		WHERE organization_id = $1 AND deleted_at IS NULL AND is_active = FALSE
	`, orgID).Scan(&inactive)
	if err != nil {
		return nil, fmt.Errorf("failed to count inactive patients: %w", err)
	}
	stats["inactive"] = inactive

	// New patients this month
	var thisMonth int
	err = s.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM patients
		WHERE organization_id = $1
			AND deleted_at IS NULL
			AND created_at >= DATE_TRUNC('month', CURRENT_DATE)
	`, orgID).Scan(&thisMonth)
	if err != nil {
		return nil, fmt.Errorf("failed to count monthly patients: %w", err)
	}
	stats["this_month"] = thisMonth

	return stats, nil
}
