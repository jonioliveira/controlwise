package services

import (
	"context"
	"encoding/json"

	"github.com/controlewise/backend/internal/database"
	"github.com/controlewise/backend/internal/models"
	"github.com/google/uuid"
)

type AdminAuditService struct {
	db *database.DB
}

func NewAdminAuditService(db *database.DB) *AdminAuditService {
	return &AdminAuditService{db: db}
}

type AuditLogParams struct {
	AdminID    *uuid.UUID
	Action     string
	EntityType string
	EntityID   *uuid.UUID
	Page       int
	Limit      int
}

func (s *AdminAuditService) Log(ctx context.Context, adminID uuid.UUID, action models.AuditAction, entityType models.AuditEntityType, entityID *uuid.UUID, details map[string]interface{}, ipAddress, userAgent string) error {
	detailsJSON, err := json.Marshal(details)
	if err != nil {
		detailsJSON = []byte("{}")
	}

	_, err = s.db.Pool.Exec(ctx, `
		INSERT INTO system_admin_audit_logs (id, admin_id, action, entity_type, entity_id, details, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, uuid.New(), adminID, string(action), string(entityType), entityID, detailsJSON, ipAddress, userAgent)

	return err
}

func (s *AdminAuditService) List(ctx context.Context, params AuditLogParams) ([]models.SystemAdminAuditLog, int, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 || params.Limit > 100 {
		params.Limit = 50
	}
	offset := (params.Page - 1) * params.Limit

	// Build query
	query := `
		SELECT id, admin_id, action, entity_type, entity_id, details, ip_address, user_agent, created_at
		FROM system_admin_audit_logs
		WHERE 1=1
	`
	countQuery := `SELECT COUNT(*) FROM system_admin_audit_logs WHERE 1=1`
	args := []interface{}{}
	argCount := 0

	if params.AdminID != nil {
		argCount++
		query += ` AND admin_id = $` + string(rune('0'+argCount))
		countQuery += ` AND admin_id = $` + string(rune('0'+argCount))
		args = append(args, *params.AdminID)
	}

	if params.Action != "" {
		argCount++
		query += ` AND action = $` + string(rune('0'+argCount))
		countQuery += ` AND action = $` + string(rune('0'+argCount))
		args = append(args, params.Action)
	}

	if params.EntityType != "" {
		argCount++
		query += ` AND entity_type = $` + string(rune('0'+argCount))
		countQuery += ` AND entity_type = $` + string(rune('0'+argCount))
		args = append(args, params.EntityType)
	}

	if params.EntityID != nil {
		argCount++
		query += ` AND entity_id = $` + string(rune('0'+argCount))
		countQuery += ` AND entity_id = $` + string(rune('0'+argCount))
		args = append(args, *params.EntityID)
	}

	// Get total count
	var total int
	err := s.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query += ` ORDER BY created_at DESC`
	query += ` LIMIT $` + string(rune('0'+argCount+1)) + ` OFFSET $` + string(rune('0'+argCount+2))
	args = append(args, params.Limit, offset)

	rows, err := s.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []models.SystemAdminAuditLog
	for rows.Next() {
		var log models.SystemAdminAuditLog
		var detailsJSON []byte
		err := rows.Scan(
			&log.ID, &log.AdminID, &log.Action, &log.EntityType, &log.EntityID,
			&detailsJSON, &log.IPAddress, &log.UserAgent, &log.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		if err := json.Unmarshal(detailsJSON, &log.Details); err != nil {
			log.Details = make(map[string]interface{})
		}
		logs = append(logs, log)
	}

	return logs, total, nil
}
