package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/controlwise/backend/internal/database"
	"github.com/controlwise/backend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// SessionPaymentService handles session payment operations
type SessionPaymentService struct {
	db *database.DB
}

// NewSessionPaymentService creates a new SessionPaymentService
func NewSessionPaymentService(db *database.DB) *SessionPaymentService {
	return &SessionPaymentService{db: db}
}

// SessionPaymentFilters contains filters for listing session payments
type SessionPaymentFilters struct {
	Status      *string
	TherapistID *uuid.UUID
	PatientID   *uuid.UUID
	StartDate   *time.Time
	EndDate     *time.Time
	Limit       int
	Offset      int
}

// GetBySessionID returns the payment record for a session
func (s *SessionPaymentService) GetBySessionID(ctx context.Context, sessionID, orgID uuid.UUID) (*models.SessionPayment, error) {
	var p models.SessionPayment
	err := s.db.Pool.QueryRow(ctx, `
		SELECT sp.id, sp.session_id, sp.amount_cents, sp.payment_status, sp.payment_method,
		       sp.insurance_provider, sp.insurance_amount_cents, sp.due_date, sp.paid_at,
		       sp.notes, sp.created_at, sp.updated_at
		FROM session_payments sp
		JOIN sessions s ON s.id = sp.session_id
		WHERE sp.session_id = $1 AND s.organization_id = $2
	`, sessionID, orgID).Scan(
		&p.ID, &p.SessionID, &p.AmountCents, &p.PaymentStatus, &p.PaymentMethod,
		&p.InsuranceProvider, &p.InsuranceAmountCents, &p.DueDate, &p.PaidAt,
		&p.Notes, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // No payment record yet
		}
		return nil, fmt.Errorf("failed to get session payment: %w", err)
	}
	return &p, nil
}

// CreateOrUpdate creates or updates a session payment record
func (s *SessionPaymentService) CreateOrUpdate(ctx context.Context, sessionID, orgID uuid.UUID, payment *models.SessionPayment) error {
	// Verify session exists and belongs to organization
	var sessionOrgID uuid.UUID
	err := s.db.Pool.QueryRow(ctx, `
		SELECT organization_id FROM sessions WHERE id = $1 AND deleted_at IS NULL
	`, sessionID).Scan(&sessionOrgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("session not found")
		}
		return fmt.Errorf("failed to verify session: %w", err)
	}
	if sessionOrgID != orgID {
		return errors.New("session not found")
	}

	// Check if payment record exists
	existing, err := s.GetBySessionID(ctx, sessionID, orgID)
	if err != nil {
		return err
	}

	if existing != nil {
		// Update existing record
		_, err = s.db.Pool.Exec(ctx, `
			UPDATE session_payments
			SET amount_cents = $1, payment_status = $2, payment_method = $3,
			    insurance_provider = $4, insurance_amount_cents = $5, due_date = $6,
			    paid_at = $7, notes = $8, updated_at = NOW()
			WHERE session_id = $9
		`, payment.AmountCents, payment.PaymentStatus, payment.PaymentMethod,
			payment.InsuranceProvider, payment.InsuranceAmountCents, payment.DueDate,
			payment.PaidAt, payment.Notes, sessionID)
		if err != nil {
			return fmt.Errorf("failed to update session payment: %w", err)
		}
	} else {
		// Create new record
		payment.ID = uuid.New()
		payment.SessionID = sessionID
		_, err = s.db.Pool.Exec(ctx, `
			INSERT INTO session_payments
			(id, session_id, amount_cents, payment_status, payment_method,
			 insurance_provider, insurance_amount_cents, due_date, paid_at, notes)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, payment.ID, sessionID, payment.AmountCents, payment.PaymentStatus,
			payment.PaymentMethod, payment.InsuranceProvider, payment.InsuranceAmountCents,
			payment.DueDate, payment.PaidAt, payment.Notes)
		if err != nil {
			return fmt.Errorf("failed to create session payment: %w", err)
		}
	}

	return nil
}

// MarkAsPaid marks a session payment as paid
func (s *SessionPaymentService) MarkAsPaid(ctx context.Context, sessionID, orgID uuid.UUID, method *models.PaymentMethod) error {
	now := time.Now()
	result, err := s.db.Pool.Exec(ctx, `
		UPDATE session_payments sp
		SET payment_status = 'paid', payment_method = $1, paid_at = $2, updated_at = NOW()
		FROM sessions s
		WHERE sp.session_id = s.id AND sp.session_id = $3 AND s.organization_id = $4
	`, method, now, sessionID, orgID)
	if err != nil {
		return fmt.Errorf("failed to mark payment as paid: %w", err)
	}
	if result.RowsAffected() == 0 {
		return errors.New("payment record not found")
	}
	return nil
}

// ListUnpaid returns all unpaid session payments for an organization
func (s *SessionPaymentService) ListUnpaid(ctx context.Context, orgID uuid.UUID, filters SessionPaymentFilters) ([]*models.SessionPaymentWithDetails, int, error) {
	args := []interface{}{orgID}
	argNum := 1

	whereClause := `WHERE s.organization_id = $1 AND s.deleted_at IS NULL
		AND (sp.payment_status IS NULL OR sp.payment_status IN ('unpaid', 'partial'))`

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
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM sessions s
		LEFT JOIN session_payments sp ON sp.session_id = s.id
		%s
	`, whereClause)

	var total int
	err := s.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count unpaid sessions: %w", err)
	}

	// Get unpaid sessions with details
	query := fmt.Sprintf(`
		SELECT
			COALESCE(sp.id, '00000000-0000-0000-0000-000000000000'::uuid) as payment_id,
			s.id as session_id,
			COALESCE(sp.amount_cents, s.price_cents) as amount_cents,
			COALESCE(sp.payment_status, 'unpaid') as payment_status,
			sp.payment_method,
			sp.insurance_provider,
			sp.insurance_amount_cents,
			sp.due_date,
			sp.paid_at,
			sp.notes,
			COALESCE(sp.created_at, s.created_at) as created_at,
			COALESCE(sp.updated_at, s.updated_at) as updated_at,
			c.name as patient_name,
			t.name as therapist_name,
			s.scheduled_at
		FROM sessions s
		LEFT JOIN session_payments sp ON sp.session_id = s.id
		LEFT JOIN patients p ON p.id = s.patient_id
		LEFT JOIN clients c ON c.id = p.client_id
		LEFT JOIN therapists t ON t.id = s.therapist_id
		%s
		ORDER BY s.scheduled_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argNum+1, argNum+2)

	args = append(args, filters.Limit, filters.Offset)

	rows, err := s.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query unpaid sessions: %w", err)
	}
	defer rows.Close()

	var payments []*models.SessionPaymentWithDetails
	for rows.Next() {
		var p models.SessionPaymentWithDetails
		err := rows.Scan(
			&p.ID, &p.SessionID, &p.AmountCents, &p.PaymentStatus, &p.PaymentMethod,
			&p.InsuranceProvider, &p.InsuranceAmountCents, &p.DueDate, &p.PaidAt,
			&p.Notes, &p.CreatedAt, &p.UpdatedAt,
			&p.PatientName, &p.TherapistName, &p.ScheduledAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan payment: %w", err)
		}
		payments = append(payments, &p)
	}

	return payments, total, nil
}

// GetPaymentStats returns payment statistics for an organization
func (s *SessionPaymentService) GetPaymentStats(ctx context.Context, orgID uuid.UUID, startDate, endDate *time.Time) (*PaymentStats, error) {
	args := []interface{}{orgID}
	argNum := 1

	dateFilter := ""
	if startDate != nil {
		argNum++
		dateFilter += fmt.Sprintf(" AND s.scheduled_at >= $%d", argNum)
		args = append(args, *startDate)
	}
	if endDate != nil {
		argNum++
		dateFilter += fmt.Sprintf(" AND s.scheduled_at <= $%d", argNum)
		args = append(args, *endDate)
	}

	var stats PaymentStats

	// Get counts by status
	query := fmt.Sprintf(`
		SELECT
			COUNT(*) FILTER (WHERE sp.payment_status = 'paid' OR sp.payment_status IS NULL AND s.status = 'pending') as total_sessions,
			COUNT(*) FILTER (WHERE sp.payment_status = 'paid') as paid_count,
			COUNT(*) FILTER (WHERE sp.payment_status = 'unpaid' OR sp.payment_status IS NULL) as unpaid_count,
			COUNT(*) FILTER (WHERE sp.payment_status = 'partial') as partial_count,
			COALESCE(SUM(CASE WHEN sp.payment_status = 'paid' THEN sp.amount_cents ELSE 0 END), 0) as total_paid_cents,
			COALESCE(SUM(CASE WHEN sp.payment_status IN ('unpaid', 'partial') OR sp.payment_status IS NULL
				THEN COALESCE(sp.amount_cents, s.price_cents) ELSE 0 END), 0) as total_unpaid_cents
		FROM sessions s
		LEFT JOIN session_payments sp ON sp.session_id = s.id
		WHERE s.organization_id = $1 AND s.deleted_at IS NULL AND s.status IN ('completed', 'confirmed', 'pending')
		%s
	`, dateFilter)

	err := s.db.Pool.QueryRow(ctx, query, args...).Scan(
		&stats.TotalSessions, &stats.PaidCount, &stats.UnpaidCount, &stats.PartialCount,
		&stats.TotalPaidCents, &stats.TotalUnpaidCents,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment stats: %w", err)
	}

	return &stats, nil
}

// PaymentStats contains payment statistics
type PaymentStats struct {
	TotalSessions    int   `json:"total_sessions"`
	PaidCount        int   `json:"paid_count"`
	UnpaidCount      int   `json:"unpaid_count"`
	PartialCount     int   `json:"partial_count"`
	TotalPaidCents   int64 `json:"total_paid_cents"`
	TotalUnpaidCents int64 `json:"total_unpaid_cents"`
}

// ListByPatient returns all payment records for a patient
func (s *SessionPaymentService) ListByPatient(ctx context.Context, patientID, orgID uuid.UUID) ([]*models.SessionPaymentWithDetails, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT
			COALESCE(sp.id, '00000000-0000-0000-0000-000000000000'::uuid) as payment_id,
			s.id as session_id,
			COALESCE(sp.amount_cents, s.price_cents) as amount_cents,
			COALESCE(sp.payment_status, 'unpaid') as payment_status,
			sp.payment_method,
			sp.insurance_provider,
			sp.insurance_amount_cents,
			sp.due_date,
			sp.paid_at,
			sp.notes,
			COALESCE(sp.created_at, s.created_at) as created_at,
			COALESCE(sp.updated_at, s.updated_at) as updated_at,
			c.name as patient_name,
			t.name as therapist_name,
			s.scheduled_at
		FROM sessions s
		LEFT JOIN session_payments sp ON sp.session_id = s.id
		LEFT JOIN patients p ON p.id = s.patient_id
		LEFT JOIN clients c ON c.id = p.client_id
		LEFT JOIN therapists t ON t.id = s.therapist_id
		WHERE s.patient_id = $1 AND s.organization_id = $2 AND s.deleted_at IS NULL
		ORDER BY s.scheduled_at DESC
	`, patientID, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to query patient payments: %w", err)
	}
	defer rows.Close()

	var payments []*models.SessionPaymentWithDetails
	for rows.Next() {
		var p models.SessionPaymentWithDetails
		err := rows.Scan(
			&p.ID, &p.SessionID, &p.AmountCents, &p.PaymentStatus, &p.PaymentMethod,
			&p.InsuranceProvider, &p.InsuranceAmountCents, &p.DueDate, &p.PaidAt,
			&p.Notes, &p.CreatedAt, &p.UpdatedAt,
			&p.PatientName, &p.TherapistName, &p.ScheduledAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment: %w", err)
		}
		payments = append(payments, &p)
	}

	return payments, nil
}

// CreatePaymentForSession creates a payment record when a session is completed
func (s *SessionPaymentService) CreatePaymentForSession(ctx context.Context, sessionID, orgID uuid.UUID) error {
	// Get session price
	var priceCents int
	err := s.db.Pool.QueryRow(ctx, `
		SELECT price_cents FROM sessions WHERE id = $1 AND organization_id = $2
	`, sessionID, orgID).Scan(&priceCents)
	if err != nil {
		return fmt.Errorf("failed to get session price: %w", err)
	}

	// Check if payment already exists
	existing, err := s.GetBySessionID(ctx, sessionID, orgID)
	if err != nil {
		return err
	}
	if existing != nil {
		return nil // Payment already exists
	}

	// Create payment record
	_, err = s.db.Pool.Exec(ctx, `
		INSERT INTO session_payments (id, session_id, amount_cents, payment_status)
		VALUES ($1, $2, $3, 'unpaid')
	`, uuid.New(), sessionID, priceCents)
	if err != nil {
		return fmt.Errorf("failed to create session payment: %w", err)
	}

	return nil
}
