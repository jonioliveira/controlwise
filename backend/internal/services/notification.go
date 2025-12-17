package services

import (
	"context"
	"time"

	"github.com/controlwise/backend/internal/database"
	"github.com/controlwise/backend/internal/models"
	"github.com/google/uuid"
)

type NotificationService struct {
	db    *database.DB
	email *EmailService
}

func NewNotificationService(db *database.DB, email *EmailService) *NotificationService {
	return &NotificationService{
		db:    db,
		email: email,
	}
}

func (s *NotificationService) Create(ctx context.Context, notification *models.Notification) error {
	_, err := s.db.Pool.Exec(ctx, `
		INSERT INTO notifications (id, user_id, type, title, message, entity_type, entity_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, uuid.New(), notification.UserID, notification.Type, notification.Title, notification.Message,
		notification.EntityType, notification.EntityID, time.Now())

	return err
}

func (s *NotificationService) CreateAndEmail(ctx context.Context, notification *models.Notification, userEmail string) error {
	// Create in-app notification
	if err := s.Create(ctx, notification); err != nil {
		return err
	}

	// Send email notification
	go s.email.SendNotification(userEmail, notification.Title, notification.Message)

	return nil
}

func (s *NotificationService) List(ctx context.Context, userID uuid.UUID, limit int) ([]*models.Notification, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT id, user_id, type, title, message, entity_type, entity_id, is_read, read_at, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*models.Notification
	for rows.Next() {
		var n models.Notification
		if err := rows.Scan(
			&n.ID, &n.UserID, &n.Type, &n.Title, &n.Message,
			&n.EntityType, &n.EntityID, &n.IsRead, &n.ReadAt, &n.CreatedAt,
		); err != nil {
			return nil, err
		}
		notifications = append(notifications, &n)
	}

	return notifications, nil
}

func (s *NotificationService) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.Pool.Exec(ctx, `
		UPDATE notifications
		SET is_read = true, read_at = $1
		WHERE id = $2
	`, time.Now(), id)
	return err
}

func (s *NotificationService) UnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := s.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = false
	`, userID).Scan(&count)
	return count, err
}
