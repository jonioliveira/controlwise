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

type WorksheetService struct {
	db           *database.DB
	storage      *StorageService
	notification *NotificationService
}

func NewWorksheetService(db *database.DB, storage *StorageService, notification *NotificationService) *WorksheetService {
	return &WorksheetService{
		db:           db,
		storage:      storage,
		notification: notification,
	}
}

type WorksheetWithItems struct {
	models.WorkSheet
	Items      []*models.WorkSheetItem `json:"items"`
	Photos     []*models.Photo         `json:"photos"`
	ClientName string                  `json:"client_name"`
}

// List returns worksheets with pagination
func (s *WorksheetService) List(ctx context.Context, orgID uuid.UUID, status *models.WorkSheetStatus, limit, offset int) ([]*WorksheetWithItems, int, error) {
	// Build query
	query := `
		SELECT 
			w.id, w.organization_id, w.client_id, w.title, w.description, w.status,
			w.created_by, w.reviewed_by, w.reviewed_at, w.created_at, w.updated_at,
			c.name as client_name
		FROM worksheets w
		INNER JOIN clients c ON c.id = w.client_id
		WHERE w.organization_id = $1 AND w.deleted_at IS NULL
	`
	args := []interface{}{orgID}
	argIndex := 2

	if status != nil {
		query += fmt.Sprintf(" AND w.status = $%d", argIndex)
		args = append(args, *status)
		argIndex++
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM (" + query + ") as count_query"
	var total int
	err := s.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count worksheets: %w", err)
	}

	// Add ordering and pagination
	query += fmt.Sprintf(" ORDER BY w.created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := s.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query worksheets: %w", err)
	}
	defer rows.Close()

	var worksheets []*WorksheetWithItems
	for rows.Next() {
		var w WorksheetWithItems
		err := rows.Scan(
			&w.ID, &w.OrganizationID, &w.ClientID, &w.Title, &w.Description, &w.Status,
			&w.CreatedBy, &w.ReviewedBy, &w.ReviewedAt, &w.CreatedAt, &w.UpdatedAt,
			&w.ClientName,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan worksheet: %w", err)
		}
		worksheets = append(worksheets, &w)
	}

	// Load items and photos for each worksheet
	for _, w := range worksheets {
		items, err := s.getItems(ctx, w.ID)
		if err != nil {
			return nil, 0, err
		}
		w.Items = items

		photos, err := s.getPhotos(ctx, w.ID)
		if err != nil {
			return nil, 0, err
		}
		w.Photos = photos
	}

	return worksheets, total, nil
}

// GetByID returns a worksheet with items and photos
func (s *WorksheetService) GetByID(ctx context.Context, id, orgID uuid.UUID) (*WorksheetWithItems, error) {
	var w WorksheetWithItems
	err := s.db.Pool.QueryRow(ctx, `
		SELECT 
			w.id, w.organization_id, w.client_id, w.title, w.description, w.status,
			w.created_by, w.reviewed_by, w.reviewed_at, w.created_at, w.updated_at,
			c.name as client_name
		FROM worksheets w
		INNER JOIN clients c ON c.id = w.client_id
		WHERE w.id = $1 AND w.organization_id = $2 AND w.deleted_at IS NULL
	`, id, orgID).Scan(
		&w.ID, &w.OrganizationID, &w.ClientID, &w.Title, &w.Description, &w.Status,
		&w.CreatedBy, &w.ReviewedBy, &w.ReviewedAt, &w.CreatedAt, &w.UpdatedAt,
		&w.ClientName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("worksheet not found")
		}
		return nil, fmt.Errorf("failed to get worksheet: %w", err)
	}

	// Load items
	items, err := s.getItems(ctx, w.ID)
	if err != nil {
		return nil, err
	}
	w.Items = items

	// Load photos
	photos, err := s.getPhotos(ctx, w.ID)
	if err != nil {
		return nil, err
	}
	w.Photos = photos

	return &w, nil
}

// Create creates a new worksheet with items
func (s *WorksheetService) Create(ctx context.Context, worksheet *models.WorkSheet, items []*models.WorkSheetItem) error {
	// Validate
	if worksheet.Title == "" {
		return errors.New("title is required")
	}
	if worksheet.Description == "" {
		return errors.New("description is required")
	}
	if worksheet.ClientID == uuid.Nil {
		return errors.New("client is required")
	}

	// Verify client exists and belongs to org
	var clientExists bool
	err := s.db.Pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM clients 
			WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL
		)
	`, worksheet.ClientID, worksheet.OrganizationID).Scan(&clientExists)
	if err != nil {
		return fmt.Errorf("failed to verify client: %w", err)
	}
	if !clientExists {
		return errors.New("client not found")
	}

	// Start transaction
	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Create worksheet
	worksheet.ID = uuid.New()
	worksheet.Status = models.WorkSheetStatusDraft

	_, err = tx.Exec(ctx, `
		INSERT INTO worksheets (
			id, organization_id, client_id, title, description, status, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, worksheet.ID, worksheet.OrganizationID, worksheet.ClientID, worksheet.Title,
		worksheet.Description, worksheet.Status, worksheet.CreatedBy)
	if err != nil {
		return fmt.Errorf("failed to create worksheet: %w", err)
	}

	// Create items
	for i, item := range items {
		item.ID = uuid.New()
		item.WorkSheetID = worksheet.ID
		item.Order = i

		_, err = tx.Exec(ctx, `
			INSERT INTO worksheet_items (
				id, worksheet_id, description, quantity, unit, notes, "order"
			) VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, item.ID, item.WorkSheetID, item.Description, item.Quantity, item.Unit, item.Notes, item.Order)
		if err != nil {
			return fmt.Errorf("failed to create worksheet item: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Update updates a worksheet and its items
func (s *WorksheetService) Update(ctx context.Context, id, orgID uuid.UUID, worksheet *models.WorkSheet, items []*models.WorkSheetItem) error {
	// Validate
	if worksheet.Title == "" {
		return errors.New("title is required")
	}

	// Check if worksheet can be edited
	var currentStatus models.WorkSheetStatus
	err := s.db.Pool.QueryRow(ctx, `
		SELECT status FROM worksheets 
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL
	`, id, orgID).Scan(&currentStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("worksheet not found")
		}
		return fmt.Errorf("failed to get worksheet: %w", err)
	}

	if currentStatus == models.WorkSheetStatusApproved {
		return errors.New("cannot edit approved worksheet")
	}

	// Start transaction
	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Update worksheet
	_, err = tx.Exec(ctx, `
		UPDATE worksheets
		SET title = $1, description = $2
		WHERE id = $3 AND organization_id = $4 AND deleted_at IS NULL
	`, worksheet.Title, worksheet.Description, id, orgID)
	if err != nil {
		return fmt.Errorf("failed to update worksheet: %w", err)
	}

	// Delete existing items
	_, err = tx.Exec(ctx, `
		UPDATE worksheet_items SET deleted_at = CURRENT_TIMESTAMP
		WHERE worksheet_id = $1 AND deleted_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("failed to delete old items: %w", err)
	}

	// Create new items
	for i, item := range items {
		item.ID = uuid.New()
		item.WorkSheetID = id
		item.Order = i

		_, err = tx.Exec(ctx, `
			INSERT INTO worksheet_items (
				id, worksheet_id, description, quantity, unit, notes, "order"
			) VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, item.ID, item.WorkSheetID, item.Description, item.Quantity, item.Unit, item.Notes, item.Order)
		if err != nil {
			return fmt.Errorf("failed to create worksheet item: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Review changes status to under_review or approved
func (s *WorksheetService) Review(ctx context.Context, id, orgID, reviewerID uuid.UUID, approve bool) error {
	var currentStatus models.WorkSheetStatus
	err := s.db.Pool.QueryRow(ctx, `
		SELECT status FROM worksheets 
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL
	`, id, orgID).Scan(&currentStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("worksheet not found")
		}
		return fmt.Errorf("failed to get worksheet: %w", err)
	}

	if currentStatus == models.WorkSheetStatusApproved {
		return errors.New("worksheet already approved")
	}

	newStatus := models.WorkSheetStatusUnderReview
	if approve {
		newStatus = models.WorkSheetStatusApproved
	}

	_, err = s.db.Pool.Exec(ctx, `
		UPDATE worksheets
		SET status = $1, reviewed_by = $2, reviewed_at = CURRENT_TIMESTAMP
		WHERE id = $3 AND organization_id = $4 AND deleted_at IS NULL
	`, newStatus, reviewerID, id, orgID)
	if err != nil {
		return fmt.Errorf("failed to review worksheet: %w", err)
	}

	// TODO: Send notification
	return nil
}

// Delete soft deletes a worksheet
func (s *WorksheetService) Delete(ctx context.Context, id, orgID uuid.UUID) error {
	// Check if worksheet has budgets
	var hasBudgets bool
	err := s.db.Pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM budgets 
			WHERE worksheet_id = $1 AND deleted_at IS NULL
		)
	`, id).Scan(&hasBudgets)
	if err != nil {
		return fmt.Errorf("failed to check budgets: %w", err)
	}
	if hasBudgets {
		return errors.New("cannot delete worksheet with existing budgets")
	}

	result, err := s.db.Pool.Exec(ctx, `
		UPDATE worksheets
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL
	`, id, orgID)
	if err != nil {
		return fmt.Errorf("failed to delete worksheet: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("worksheet not found or already deleted")
	}

	return nil
}

// Helper functions

func (s *WorksheetService) getItems(ctx context.Context, worksheetID uuid.UUID) ([]*models.WorkSheetItem, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT id, worksheet_id, description, quantity, unit, notes, "order", created_at, updated_at
		FROM worksheet_items
		WHERE worksheet_id = $1 AND deleted_at IS NULL
		ORDER BY "order"
	`, worksheetID)
	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}
	defer rows.Close()

	var items []*models.WorkSheetItem
	for rows.Next() {
		var item models.WorkSheetItem
		err := rows.Scan(
			&item.ID, &item.WorkSheetID, &item.Description, &item.Quantity,
			&item.Unit, &item.Notes, &item.Order, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		items = append(items, &item)
	}

	return items, nil
}

func (s *WorksheetService) getPhotos(ctx context.Context, worksheetID uuid.UUID) ([]*models.Photo, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT id, organization_id, entity_type, entity_id, file_name, file_size,
		       mime_type, url, thumbnail_url, caption, uploaded_by, created_at
		FROM photos
		WHERE entity_type = 'worksheet' AND entity_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, worksheetID)
	if err != nil {
		return nil, fmt.Errorf("failed to query photos: %w", err)
	}
	defer rows.Close()

	var photos []*models.Photo
	for rows.Next() {
		var photo models.Photo
		err := rows.Scan(
			&photo.ID, &photo.OrganizationID, &photo.EntityType, &photo.EntityID,
			&photo.FileName, &photo.FileSize, &photo.MimeType, &photo.URL,
			&photo.ThumbnailURL, &photo.Caption, &photo.UploadedBy, &photo.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan photo: %w", err)
		}
		photos = append(photos, &photo)
	}

	return photos, nil
}
