package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/controlwise/backend/internal/database"
	"github.com/controlwise/backend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type TherapistService struct {
	db *database.DB
}

func NewTherapistService(db *database.DB) *TherapistService {
	return &TherapistService{db: db}
}

// List returns all therapists for an organization
func (s *TherapistService) List(ctx context.Context, orgID uuid.UUID, activeOnly bool) ([]*models.Therapist, error) {
	query := `
		SELECT
			id, organization_id, user_id, name, email, phone, specialty,
			working_hours, session_duration_minutes, default_price_cents,
			timezone, is_active, created_at, updated_at
		FROM therapists
		WHERE organization_id = $1 AND deleted_at IS NULL
	`
	if activeOnly {
		query += " AND is_active = TRUE"
	}
	query += " ORDER BY name ASC"

	rows, err := s.db.Pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to query therapists: %w", err)
	}
	defer rows.Close()

	var therapists []*models.Therapist
	for rows.Next() {
		var t models.Therapist
		err := rows.Scan(
			&t.ID,
			&t.OrganizationID,
			&t.UserID,
			&t.Name,
			&t.Email,
			&t.Phone,
			&t.Specialty,
			&t.WorkingHours,
			&t.SessionDurationMinutes,
			&t.DefaultPriceCents,
			&t.Timezone,
			&t.IsActive,
			&t.CreatedAt,
			&t.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan therapist: %w", err)
		}
		therapists = append(therapists, &t)
	}

	return therapists, nil
}

// GetByID returns a single therapist by ID
func (s *TherapistService) GetByID(ctx context.Context, id, orgID uuid.UUID) (*models.Therapist, error) {
	var t models.Therapist
	err := s.db.Pool.QueryRow(ctx, `
		SELECT
			id, organization_id, user_id, name, email, phone, specialty,
			working_hours, session_duration_minutes, default_price_cents,
			timezone, is_active, created_at, updated_at
		FROM therapists
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL
	`, id, orgID).Scan(
		&t.ID,
		&t.OrganizationID,
		&t.UserID,
		&t.Name,
		&t.Email,
		&t.Phone,
		&t.Specialty,
		&t.WorkingHours,
		&t.SessionDurationMinutes,
		&t.DefaultPriceCents,
		&t.Timezone,
		&t.IsActive,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("therapist not found")
		}
		return nil, fmt.Errorf("failed to get therapist: %w", err)
	}

	return &t, nil
}

// GetByUserID returns a therapist by user ID
func (s *TherapistService) GetByUserID(ctx context.Context, orgID, userID uuid.UUID) (*models.Therapist, error) {
	var t models.Therapist
	err := s.db.Pool.QueryRow(ctx, `
		SELECT
			id, organization_id, user_id, name, email, phone, specialty,
			working_hours, session_duration_minutes, default_price_cents,
			timezone, is_active, created_at, updated_at
		FROM therapists
		WHERE organization_id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, orgID, userID).Scan(
		&t.ID,
		&t.OrganizationID,
		&t.UserID,
		&t.Name,
		&t.Email,
		&t.Phone,
		&t.Specialty,
		&t.WorkingHours,
		&t.SessionDurationMinutes,
		&t.DefaultPriceCents,
		&t.Timezone,
		&t.IsActive,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get therapist: %w", err)
	}

	return &t, nil
}

// Create creates a new therapist
func (s *TherapistService) Create(ctx context.Context, therapist *models.Therapist) error {
	// Validate required fields
	if therapist.Name == "" {
		return errors.New("therapist name is required")
	}

	// Check if user_id is already linked to a therapist
	if therapist.UserID != nil {
		existing, err := s.GetByUserID(ctx, therapist.OrganizationID, *therapist.UserID)
		if err != nil {
			return fmt.Errorf("failed to check user linkage: %w", err)
		}
		if existing != nil {
			return errors.New("user is already linked to a therapist")
		}
	}

	// Set defaults
	therapist.ID = uuid.New()
	therapist.IsActive = true
	if therapist.SessionDurationMinutes == 0 {
		therapist.SessionDurationMinutes = 60
	}
	if therapist.Timezone == "" {
		therapist.Timezone = "Europe/Lisbon"
	}
	if therapist.WorkingHours == nil {
		// Default working hours
		defaultHours := models.WorkingHours{
			"monday":    {Start: "09:00", End: "18:00"},
			"tuesday":   {Start: "09:00", End: "18:00"},
			"wednesday": {Start: "09:00", End: "18:00"},
			"thursday":  {Start: "09:00", End: "18:00"},
			"friday":    {Start: "09:00", End: "18:00"},
		}
		data, _ := json.Marshal(defaultHours)
		therapist.WorkingHours = data
	}

	// Insert therapist
	_, err := s.db.Pool.Exec(ctx, `
		INSERT INTO therapists (
			id, organization_id, user_id, name, email, phone, specialty,
			working_hours, session_duration_minutes, default_price_cents,
			timezone, is_active
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, therapist.ID, therapist.OrganizationID, therapist.UserID, therapist.Name,
		therapist.Email, therapist.Phone, therapist.Specialty, therapist.WorkingHours,
		therapist.SessionDurationMinutes, therapist.DefaultPriceCents,
		therapist.Timezone, therapist.IsActive)

	if err != nil {
		return fmt.Errorf("failed to create therapist: %w", err)
	}

	return nil
}

// Update updates an existing therapist
func (s *TherapistService) Update(ctx context.Context, id, orgID uuid.UUID, therapist *models.Therapist) error {
	// Validate required fields
	if therapist.Name == "" {
		return errors.New("therapist name is required")
	}

	// Check if user_id is being changed and already linked
	if therapist.UserID != nil {
		existing, err := s.GetByUserID(ctx, orgID, *therapist.UserID)
		if err != nil {
			return fmt.Errorf("failed to check user linkage: %w", err)
		}
		if existing != nil && existing.ID != id {
			return errors.New("user is already linked to another therapist")
		}
	}

	// Update therapist
	result, err := s.db.Pool.Exec(ctx, `
		UPDATE therapists
		SET user_id = $1, name = $2, email = $3, phone = $4, specialty = $5,
		    working_hours = $6, session_duration_minutes = $7,
		    default_price_cents = $8, timezone = $9, is_active = $10
		WHERE id = $11 AND organization_id = $12 AND deleted_at IS NULL
	`, therapist.UserID, therapist.Name, therapist.Email, therapist.Phone,
		therapist.Specialty, therapist.WorkingHours, therapist.SessionDurationMinutes,
		therapist.DefaultPriceCents, therapist.Timezone, therapist.IsActive, id, orgID)

	if err != nil {
		return fmt.Errorf("failed to update therapist: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("therapist not found or already deleted")
	}

	return nil
}

// Delete soft deletes a therapist
func (s *TherapistService) Delete(ctx context.Context, id, orgID uuid.UUID) error {
	// Check if therapist has any future sessions
	var hasSessions bool
	err := s.db.Pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM sessions
			WHERE therapist_id = $1
				AND deleted_at IS NULL
				AND scheduled_at > NOW()
				AND status NOT IN ('cancelled', 'completed')
		)
	`, id).Scan(&hasSessions)
	if err != nil {
		return fmt.Errorf("failed to check sessions: %w", err)
	}
	if hasSessions {
		return errors.New("cannot delete therapist with future sessions")
	}

	// Soft delete
	result, err := s.db.Pool.Exec(ctx, `
		UPDATE therapists
		SET deleted_at = CURRENT_TIMESTAMP, is_active = FALSE
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL
	`, id, orgID)

	if err != nil {
		return fmt.Errorf("failed to delete therapist: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("therapist not found or already deleted")
	}

	return nil
}

// GetStats returns therapist statistics
func (s *TherapistService) GetStats(ctx context.Context, orgID uuid.UUID) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total active therapists
	var total int
	err := s.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM therapists
		WHERE organization_id = $1 AND deleted_at IS NULL AND is_active = TRUE
	`, orgID).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count therapists: %w", err)
	}
	stats["total"] = total

	return stats, nil
}
