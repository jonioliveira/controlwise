package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/controlwise/backend/internal/database"
	"github.com/controlwise/backend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type SessionService struct {
	db       *database.DB
	workflow *WorkflowService
}

func NewSessionService(db *database.DB) *SessionService {
	return &SessionService{db: db}
}

// SetWorkflowService sets the workflow service for session workflow integration
func (s *SessionService) SetWorkflowService(ws *WorkflowService) {
	s.workflow = ws
}

// List returns sessions for an organization with filters
func (s *SessionService) List(ctx context.Context, orgID uuid.UUID, filters SessionFilters) ([]*models.SessionWithDetails, int, error) {
	args := []interface{}{orgID}
	argNum := 1

	whereClause := "WHERE s.organization_id = $1 AND s.deleted_at IS NULL"

	if filters.TherapistID != nil {
		argNum++
		whereClause += fmt.Sprintf(" AND s.therapist_id = $%d", argNum)
		args = append(args, *filters.TherapistID)
	}

	if filters.PatientID != nil {
		argNum++
		whereClause += fmt.Sprintf(" AND s.patient_id = $%d", argNum)
		args = append(args, *filters.PatientID)
	}

	if filters.Status != nil {
		argNum++
		whereClause += fmt.Sprintf(" AND s.status = $%d", argNum)
		args = append(args, *filters.Status)
	}

	if filters.StartDate != nil {
		argNum++
		whereClause += fmt.Sprintf(" AND s.scheduled_at >= $%d", argNum)
		args = append(args, *filters.StartDate)
	}

	if filters.EndDate != nil {
		argNum++
		whereClause += fmt.Sprintf(" AND s.scheduled_at <= $%d", argNum)
		args = append(args, *filters.EndDate)
	}

	// Get total count
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM sessions s %s", whereClause)
	err := s.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count sessions: %w", err)
	}

	// Get sessions with details
	query := fmt.Sprintf(`
		SELECT
			s.id, s.organization_id, s.therapist_id, s.patient_id,
			s.scheduled_at, s.duration_minutes, s.price_cents, s.status,
			s.session_type, s.notes, s.cancel_reason, s.cancelled_at,
			s.cancelled_by, s.completed_at, s.created_by, s.created_at, s.updated_at,
			t.name as therapist_name,
			p.name as patient_name, p.phone as patient_phone, p.email as patient_email
		FROM sessions s
		JOIN therapists t ON t.id = s.therapist_id
		JOIN patients p ON p.id = s.patient_id
		%s
		ORDER BY s.scheduled_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argNum+1, argNum+2)

	args = append(args, filters.Limit, filters.Offset)

	rows, err := s.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*models.SessionWithDetails
	for rows.Next() {
		var sd models.SessionWithDetails
		err := rows.Scan(
			&sd.ID,
			&sd.OrganizationID,
			&sd.TherapistID,
			&sd.PatientID,
			&sd.ScheduledAt,
			&sd.DurationMinutes,
			&sd.PriceCents,
			&sd.Status,
			&sd.SessionType,
			&sd.Notes,
			&sd.CancelReason,
			&sd.CancelledAt,
			&sd.CancelledBy,
			&sd.CompletedAt,
			&sd.CreatedBy,
			&sd.CreatedAt,
			&sd.UpdatedAt,
			&sd.TherapistName,
			&sd.PatientName,
			&sd.PatientPhone,
			&sd.PatientEmail,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, &sd)
	}

	return sessions, total, nil
}

// GetCalendarEvents returns sessions formatted for calendar display
func (s *SessionService) GetCalendarEvents(ctx context.Context, orgID uuid.UUID, start, end time.Time, therapistID *uuid.UUID) ([]models.CalendarEvent, error) {
	args := []interface{}{orgID, start, end}
	query := `
		SELECT
			s.id, s.organization_id, s.therapist_id, s.patient_id,
			s.scheduled_at, s.duration_minutes, s.price_cents, s.status,
			s.session_type, s.notes, s.cancel_reason, s.cancelled_at,
			s.cancelled_by, s.completed_at, s.created_by, s.created_at, s.updated_at,
			t.name as therapist_name,
			p.name as patient_name, p.phone as patient_phone, p.email as patient_email
		FROM sessions s
		JOIN therapists t ON t.id = s.therapist_id
		JOIN patients p ON p.id = s.patient_id
		WHERE s.organization_id = $1 AND s.deleted_at IS NULL
			AND s.scheduled_at >= $2 AND s.scheduled_at <= $3
	`

	if therapistID != nil {
		query += " AND s.therapist_id = $4"
		args = append(args, *therapistID)
	}

	query += " ORDER BY s.scheduled_at ASC"

	rows, err := s.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions: %w", err)
	}
	defer rows.Close()

	var events []models.CalendarEvent
	for rows.Next() {
		var sd models.SessionWithDetails
		err := rows.Scan(
			&sd.ID,
			&sd.OrganizationID,
			&sd.TherapistID,
			&sd.PatientID,
			&sd.ScheduledAt,
			&sd.DurationMinutes,
			&sd.PriceCents,
			&sd.Status,
			&sd.SessionType,
			&sd.Notes,
			&sd.CancelReason,
			&sd.CancelledAt,
			&sd.CancelledBy,
			&sd.CompletedAt,
			&sd.CreatedBy,
			&sd.CreatedAt,
			&sd.UpdatedAt,
			&sd.TherapistName,
			&sd.PatientName,
			&sd.PatientPhone,
			&sd.PatientEmail,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		events = append(events, sd.ToCalendarEvent())
	}

	return events, nil
}

// GetByID returns a single session by ID with details
func (s *SessionService) GetByID(ctx context.Context, id, orgID uuid.UUID) (*models.SessionWithDetails, error) {
	var sd models.SessionWithDetails
	err := s.db.Pool.QueryRow(ctx, `
		SELECT
			s.id, s.organization_id, s.therapist_id, s.patient_id,
			s.scheduled_at, s.duration_minutes, s.price_cents, s.status,
			s.session_type, s.notes, s.cancel_reason, s.cancelled_at,
			s.cancelled_by, s.completed_at, s.created_by, s.created_at, s.updated_at,
			t.name as therapist_name,
			p.name as patient_name, p.phone as patient_phone, p.email as patient_email
		FROM sessions s
		JOIN therapists t ON t.id = s.therapist_id
		JOIN patients p ON p.id = s.patient_id
		WHERE s.id = $1 AND s.organization_id = $2 AND s.deleted_at IS NULL
	`, id, orgID).Scan(
		&sd.ID,
		&sd.OrganizationID,
		&sd.TherapistID,
		&sd.PatientID,
		&sd.ScheduledAt,
		&sd.DurationMinutes,
		&sd.PriceCents,
		&sd.Status,
		&sd.SessionType,
		&sd.Notes,
		&sd.CancelReason,
		&sd.CancelledAt,
		&sd.CancelledBy,
		&sd.CompletedAt,
		&sd.CreatedBy,
		&sd.CreatedAt,
		&sd.UpdatedAt,
		&sd.TherapistName,
		&sd.PatientName,
		&sd.PatientPhone,
		&sd.PatientEmail,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &sd, nil
}

// Create creates a new session with conflict detection
func (s *SessionService) Create(ctx context.Context, session *models.Session, createdBy uuid.UUID) error {
	// Validate required fields
	if session.TherapistID == uuid.Nil {
		return errors.New("therapist is required")
	}
	if session.PatientID == uuid.Nil {
		return errors.New("patient is required")
	}
	if session.ScheduledAt.IsZero() {
		return errors.New("scheduled time is required")
	}
	if session.DurationMinutes <= 0 {
		return errors.New("duration must be positive")
	}

	// Check for scheduling conflicts
	endTime := session.ScheduledAt.Add(time.Duration(session.DurationMinutes) * time.Minute)
	hasConflict, err := s.hasConflict(ctx, session.OrganizationID, session.TherapistID, session.ScheduledAt, endTime, nil)
	if err != nil {
		return fmt.Errorf("failed to check conflicts: %w", err)
	}
	if hasConflict {
		return errors.New("scheduling conflict: therapist already has a session at this time")
	}

	// Set defaults
	session.ID = uuid.New()
	session.Status = models.SessionStatusPending
	if session.SessionType == "" {
		session.SessionType = models.SessionTypeRegular
	}
	session.CreatedBy = &createdBy

	// Insert session
	_, err = s.db.Pool.Exec(ctx, `
		INSERT INTO sessions (
			id, organization_id, therapist_id, patient_id, scheduled_at,
			duration_minutes, price_cents, status, session_type, notes, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, session.ID, session.OrganizationID, session.TherapistID, session.PatientID,
		session.ScheduledAt, session.DurationMinutes, session.PriceCents,
		session.Status, session.SessionType, session.Notes, session.CreatedBy)

	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Record history
	s.recordHistory(ctx, session.ID, "created", nil, session, &createdBy)

	// Trigger workflow for session creation (entering pending state)
	if s.workflow != nil {
		if err := s.workflow.OnSessionStateChange(ctx, session.OrganizationID, session.ID, "", string(session.Status), session.ScheduledAt); err != nil {
			// Log but don't fail session creation
			fmt.Printf("Failed to trigger workflow: %v\n", err)
		}
	}

	return nil
}

// Update updates an existing session
func (s *SessionService) Update(ctx context.Context, id, orgID uuid.UUID, session *models.Session, updatedBy uuid.UUID) error {
	// Get existing session for history
	existing, err := s.GetByID(ctx, id, orgID)
	if err != nil {
		return err
	}

	// Check if session can be modified
	if existing.Status == models.SessionStatusCompleted || existing.Status == models.SessionStatusCancelled {
		return errors.New("cannot modify completed or cancelled sessions")
	}

	// Check for scheduling conflicts if time changed
	if !session.ScheduledAt.Equal(existing.ScheduledAt) || session.DurationMinutes != existing.DurationMinutes {
		endTime := session.ScheduledAt.Add(time.Duration(session.DurationMinutes) * time.Minute)
		hasConflict, err := s.hasConflict(ctx, orgID, session.TherapistID, session.ScheduledAt, endTime, &id)
		if err != nil {
			return fmt.Errorf("failed to check conflicts: %w", err)
		}
		if hasConflict {
			return errors.New("scheduling conflict: therapist already has a session at this time")
		}
	}

	// Update session
	result, err := s.db.Pool.Exec(ctx, `
		UPDATE sessions
		SET therapist_id = $1, patient_id = $2, scheduled_at = $3,
		    duration_minutes = $4, price_cents = $5, session_type = $6, notes = $7
		WHERE id = $8 AND organization_id = $9 AND deleted_at IS NULL
	`, session.TherapistID, session.PatientID, session.ScheduledAt,
		session.DurationMinutes, session.PriceCents, session.SessionType,
		session.Notes, id, orgID)

	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("session not found or already deleted")
	}

	// Record history
	s.recordHistory(ctx, id, "updated", &existing.Session, session, &updatedBy)

	return nil
}

// Confirm confirms a pending session
func (s *SessionService) Confirm(ctx context.Context, id, orgID uuid.UUID, confirmedBy uuid.UUID) error {
	existing, err := s.GetByID(ctx, id, orgID)
	if err != nil {
		return err
	}

	if existing.Status != models.SessionStatusPending {
		return errors.New("can only confirm pending sessions")
	}

	result, err := s.db.Pool.Exec(ctx, `
		UPDATE sessions
		SET status = $1
		WHERE id = $2 AND organization_id = $3 AND deleted_at IS NULL
	`, models.SessionStatusConfirmed, id, orgID)

	if err != nil {
		return fmt.Errorf("failed to confirm session: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("session not found")
	}

	// Record history
	s.recordHistory(ctx, id, "confirmed", &existing.Session, nil, &confirmedBy)

	// Trigger workflow for state change
	if s.workflow != nil {
		if err := s.workflow.OnSessionStateChange(ctx, orgID, id, string(existing.Status), string(models.SessionStatusConfirmed), existing.ScheduledAt); err != nil {
			fmt.Printf("Failed to trigger workflow: %v\n", err)
		}
	}

	return nil
}

// Cancel cancels a session
func (s *SessionService) Cancel(ctx context.Context, id, orgID uuid.UUID, reason string, cancelledBy uuid.UUID) error {
	existing, err := s.GetByID(ctx, id, orgID)
	if err != nil {
		return err
	}

	if existing.Status == models.SessionStatusCompleted || existing.Status == models.SessionStatusCancelled {
		return errors.New("cannot cancel completed or already cancelled sessions")
	}

	now := time.Now()
	result, err := s.db.Pool.Exec(ctx, `
		UPDATE sessions
		SET status = $1, cancel_reason = $2, cancelled_at = $3, cancelled_by = $4
		WHERE id = $5 AND organization_id = $6 AND deleted_at IS NULL
	`, models.SessionStatusCancelled, reason, now, cancelledBy, id, orgID)

	if err != nil {
		return fmt.Errorf("failed to cancel session: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("session not found")
	}

	// Record history
	s.recordHistory(ctx, id, "cancelled", &existing.Session, nil, &cancelledBy)

	// Trigger workflow for state change (cancelling pending jobs)
	if s.workflow != nil {
		if err := s.workflow.OnSessionStateChange(ctx, orgID, id, string(existing.Status), string(models.SessionStatusCancelled), existing.ScheduledAt); err != nil {
			fmt.Printf("Failed to trigger workflow: %v\n", err)
		}
	}

	return nil
}

// Complete marks a session as completed
func (s *SessionService) Complete(ctx context.Context, id, orgID uuid.UUID, completedBy uuid.UUID) error {
	existing, err := s.GetByID(ctx, id, orgID)
	if err != nil {
		return err
	}

	if existing.Status == models.SessionStatusCancelled {
		return errors.New("cannot complete cancelled sessions")
	}
	if existing.Status == models.SessionStatusCompleted {
		return errors.New("session is already completed")
	}

	now := time.Now()
	result, err := s.db.Pool.Exec(ctx, `
		UPDATE sessions
		SET status = $1, completed_at = $2
		WHERE id = $3 AND organization_id = $4 AND deleted_at IS NULL
	`, models.SessionStatusCompleted, now, id, orgID)

	if err != nil {
		return fmt.Errorf("failed to complete session: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("session not found")
	}

	// Record history
	s.recordHistory(ctx, id, "completed", &existing.Session, nil, &completedBy)

	// Trigger workflow for state change
	if s.workflow != nil {
		if err := s.workflow.OnSessionStateChange(ctx, orgID, id, string(existing.Status), string(models.SessionStatusCompleted), existing.ScheduledAt); err != nil {
			fmt.Printf("Failed to trigger workflow: %v\n", err)
		}
	}

	return nil
}

// MarkNoShow marks a session as no-show
func (s *SessionService) MarkNoShow(ctx context.Context, id, orgID uuid.UUID, markedBy uuid.UUID) error {
	existing, err := s.GetByID(ctx, id, orgID)
	if err != nil {
		return err
	}

	if existing.Status == models.SessionStatusCancelled || existing.Status == models.SessionStatusCompleted {
		return errors.New("cannot mark cancelled or completed sessions as no-show")
	}

	result, err := s.db.Pool.Exec(ctx, `
		UPDATE sessions
		SET status = $1
		WHERE id = $2 AND organization_id = $3 AND deleted_at IS NULL
	`, models.SessionStatusNoShow, id, orgID)

	if err != nil {
		return fmt.Errorf("failed to mark no-show: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("session not found")
	}

	// Record history
	s.recordHistory(ctx, id, "no_show", &existing.Session, nil, &markedBy)

	// Trigger workflow for state change
	if s.workflow != nil {
		if err := s.workflow.OnSessionStateChange(ctx, orgID, id, string(existing.Status), string(models.SessionStatusNoShow), existing.ScheduledAt); err != nil {
			fmt.Printf("Failed to trigger workflow: %v\n", err)
		}
	}

	return nil
}

// Delete soft deletes a session
func (s *SessionService) Delete(ctx context.Context, id, orgID uuid.UUID) error {
	result, err := s.db.Pool.Exec(ctx, `
		UPDATE sessions
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL
	`, id, orgID)

	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("session not found or already deleted")
	}

	return nil
}

// hasConflict checks if there's a scheduling conflict
func (s *SessionService) hasConflict(ctx context.Context, orgID, therapistID uuid.UUID, start, end time.Time, excludeID *uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM sessions
			WHERE organization_id = $1
				AND therapist_id = $2
				AND deleted_at IS NULL
				AND status NOT IN ('cancelled')
				AND (
					(scheduled_at < $4 AND scheduled_at + (duration_minutes * interval '1 minute') > $3)
				)
	`
	args := []interface{}{orgID, therapistID, start, end}

	if excludeID != nil {
		query += " AND id != $5"
		args = append(args, *excludeID)
	}

	query += ")"

	var hasConflict bool
	err := s.db.Pool.QueryRow(ctx, query, args...).Scan(&hasConflict)
	return hasConflict, err
}

// recordHistory records a change in session history
func (s *SessionService) recordHistory(ctx context.Context, sessionID uuid.UUID, action string, oldValues *models.Session, newValues *models.Session, changedBy *uuid.UUID) {
	var oldJSON, newJSON json.RawMessage
	if oldValues != nil {
		oldJSON, _ = json.Marshal(oldValues)
	}
	if newValues != nil {
		newJSON, _ = json.Marshal(newValues)
	}

	s.db.Pool.Exec(ctx, `
		INSERT INTO session_history (id, session_id, action, old_values, new_values, changed_by)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, uuid.New(), sessionID, action, oldJSON, newJSON, changedBy)
}

// GetStats returns session statistics
func (s *SessionService) GetStats(ctx context.Context, orgID uuid.UUID) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Today's sessions
	var todayCount int
	err := s.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM sessions
		WHERE organization_id = $1
			AND deleted_at IS NULL
			AND DATE(scheduled_at) = CURRENT_DATE
	`, orgID).Scan(&todayCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count today sessions: %w", err)
	}
	stats["today"] = todayCount

	// This week's sessions
	var weekCount int
	err = s.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM sessions
		WHERE organization_id = $1
			AND deleted_at IS NULL
			AND scheduled_at >= DATE_TRUNC('week', CURRENT_DATE)
			AND scheduled_at < DATE_TRUNC('week', CURRENT_DATE) + interval '7 days'
	`, orgID).Scan(&weekCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count week sessions: %w", err)
	}
	stats["this_week"] = weekCount

	// Pending confirmations
	var pendingCount int
	err = s.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM sessions
		WHERE organization_id = $1
			AND deleted_at IS NULL
			AND status = 'pending'
			AND scheduled_at > NOW()
	`, orgID).Scan(&pendingCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count pending sessions: %w", err)
	}
	stats["pending_confirmations"] = pendingCount

	return stats, nil
}

// SessionFilters represents filters for session queries
type SessionFilters struct {
	TherapistID *uuid.UUID
	PatientID   *uuid.UUID
	Status      *models.SessionStatus
	StartDate   *time.Time
	EndDate     *time.Time
	Limit       int
	Offset      int
}
