package services

import (
	"context"
	"errors"
	"time"

	"github.com/controlewise/backend/internal/database"
	"github.com/controlewise/backend/internal/models"
	"github.com/google/uuid"
)

type ImpersonationService struct {
	db              *database.DB
	sysAdminService *SystemAdminService
}

func NewImpersonationService(db *database.DB, sysAdminService *SystemAdminService) *ImpersonationService {
	return &ImpersonationService{
		db:              db,
		sysAdminService: sysAdminService,
	}
}

func (s *ImpersonationService) StartImpersonation(ctx context.Context, adminID, userID uuid.UUID, reason, ipAddress string) (*models.ImpersonationToken, error) {
	// Check if there's already an active impersonation session for this admin
	var existingID uuid.UUID
	err := s.db.Pool.QueryRow(ctx, `
		SELECT id FROM admin_impersonation_sessions
		WHERE admin_id = $1 AND ended_at IS NULL
	`, adminID).Scan(&existingID)
	if err == nil {
		return nil, errors.New("you already have an active impersonation session")
	}

	// Get user to impersonate
	var user models.User
	err = s.db.Pool.QueryRow(ctx, `
		SELECT id, organization_id, email, first_name, last_name, phone, avatar, role, is_active
		FROM users WHERE id = $1 AND deleted_at IS NULL
	`, userID).Scan(
		&user.ID, &user.OrganizationID, &user.Email, &user.FirstName, &user.LastName,
		&user.Phone, &user.Avatar, &user.Role, &user.IsActive,
	)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if !user.IsActive {
		return nil, errors.New("cannot impersonate inactive user")
	}

	// Create impersonation session
	sessionID := uuid.New()
	_, err = s.db.Pool.Exec(ctx, `
		INSERT INTO admin_impersonation_sessions (id, admin_id, impersonated_user_id, reason, ip_address)
		VALUES ($1, $2, $3, $4, $5)
	`, sessionID, adminID, userID, reason, ipAddress)
	if err != nil {
		return nil, err
	}

	// Generate impersonation token
	token, expiry, err := s.sysAdminService.GenerateImpersonationToken(adminID, &user, sessionID)
	if err != nil {
		return nil, err
	}

	return &models.ImpersonationToken{
		Token:     token,
		SessionID: sessionID,
		User:      &user,
		ExpiresAt: expiry,
	}, nil
}

func (s *ImpersonationService) EndImpersonation(ctx context.Context, sessionID uuid.UUID) error {
	result, err := s.db.Pool.Exec(ctx, `
		UPDATE admin_impersonation_sessions SET ended_at = $1 WHERE id = $2 AND ended_at IS NULL
	`, time.Now(), sessionID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("impersonation session not found or already ended")
	}

	return nil
}

func (s *ImpersonationService) GetActiveSession(ctx context.Context, adminID uuid.UUID) (*models.ImpersonationSession, error) {
	var session models.ImpersonationSession
	err := s.db.Pool.QueryRow(ctx, `
		SELECT id, admin_id, impersonated_user_id, started_at, ended_at, reason, ip_address
		FROM admin_impersonation_sessions
		WHERE admin_id = $1 AND ended_at IS NULL
	`, adminID).Scan(
		&session.ID, &session.AdminID, &session.ImpersonatedUserID,
		&session.StartedAt, &session.EndedAt, &session.Reason, &session.IPAddress,
	)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (s *ImpersonationService) ListSessions(ctx context.Context, adminID *uuid.UUID, page, limit int) ([]models.ImpersonationSessionWithDetails, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	query := `
		SELECT
			s.id, s.admin_id, s.impersonated_user_id, s.started_at, s.ended_at, s.reason, s.ip_address,
			u.email as user_email,
			CONCAT(u.first_name, ' ', u.last_name) as user_name,
			o.name as org_name,
			CONCAT(a.first_name, ' ', a.last_name) as admin_name
		FROM admin_impersonation_sessions s
		JOIN users u ON s.impersonated_user_id = u.id
		JOIN organizations o ON u.organization_id = o.id
		JOIN system_admins a ON s.admin_id = a.id
		WHERE 1=1
	`
	countQuery := `SELECT COUNT(*) FROM admin_impersonation_sessions WHERE 1=1`
	args := []interface{}{}
	argCount := 0

	if adminID != nil {
		argCount++
		query += ` AND s.admin_id = $` + string(rune('0'+argCount))
		countQuery += ` AND admin_id = $` + string(rune('0'+argCount))
		args = append(args, *adminID)
	}

	var total int
	err := s.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query += ` ORDER BY s.started_at DESC`
	query += ` LIMIT $` + string(rune('0'+argCount+1)) + ` OFFSET $` + string(rune('0'+argCount+2))
	args = append(args, limit, offset)

	rows, err := s.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var sessions []models.ImpersonationSessionWithDetails
	for rows.Next() {
		var session models.ImpersonationSessionWithDetails
		err := rows.Scan(
			&session.ID, &session.AdminID, &session.ImpersonatedUserID,
			&session.StartedAt, &session.EndedAt, &session.Reason, &session.IPAddress,
			&session.UserEmail, &session.UserName, &session.OrgName, &session.AdminName,
		)
		if err != nil {
			return nil, 0, err
		}
		sessions = append(sessions, session)
	}

	return sessions, total, nil
}
