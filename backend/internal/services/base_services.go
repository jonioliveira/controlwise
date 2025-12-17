package services

import (
	"context"

	"github.com/controlwise/backend/internal/database"
	"github.com/controlwise/backend/internal/models"
	"github.com/google/uuid"
)

// UserService handles user operations within an organization
type UserService struct {
	db *database.DB
}

func NewUserService(db *database.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) List(ctx context.Context, orgID uuid.UUID) ([]*models.User, error) {
	// TODO: Implement
	return nil, nil
}

func (s *UserService) Create(ctx context.Context, user *models.User) error {
	// TODO: Implement
	return nil
}

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	// TODO: Implement
	return nil, nil
}

func (s *UserService) Update(ctx context.Context, id uuid.UUID, user *models.User) error {
	// TODO: Implement
	return nil
}

func (s *UserService) Delete(ctx context.Context, id uuid.UUID) error {
	// TODO: Implement
	return nil
}

// BudgetService handles budget operations
type BudgetService struct {
	db           *database.DB
	storage      *StorageService
	notification *NotificationService
	workflow     *WorkflowService
}

func NewBudgetService(db *database.DB, storage *StorageService, notification *NotificationService) *BudgetService {
	return &BudgetService{
		db:           db,
		storage:      storage,
		notification: notification,
	}
}

// SetWorkflowService sets the workflow service for triggering workflow actions
func (s *BudgetService) SetWorkflowService(ws *WorkflowService) {
	s.workflow = ws
}

// ProjectService handles project operations
type ProjectService struct {
	db           *database.DB
	storage      *StorageService
	notification *NotificationService
}

func NewProjectService(db *database.DB, storage *StorageService, notification *NotificationService) *ProjectService {
	return &ProjectService{
		db:           db,
		storage:      storage,
		notification: notification,
	}
}

// TaskService handles task operations
type TaskService struct {
	db           *database.DB
	notification *NotificationService
}

func NewTaskService(db *database.DB, notification *NotificationService) *TaskService {
	return &TaskService{
		db:           db,
		notification: notification,
	}
}

// PaymentService handles payment operations
type PaymentService struct {
	db           *database.DB
	notification *NotificationService
}

func NewPaymentService(db *database.DB, notification *NotificationService) *PaymentService {
	return &PaymentService{
		db:           db,
		notification: notification,
	}
}

// ReportService handles report generation
type ReportService struct {
	db *database.DB
}

func NewReportService(db *database.DB) *ReportService {
	return &ReportService{db: db}
}
