package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Organization represents a company/tenant
type Organization struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	Name      string     `json:"name" db:"name"`
	Email     string     `json:"email" db:"email"`
	Phone     string     `json:"phone" db:"phone"`
	Address   string     `json:"address" db:"address"`
	TaxID     string     `json:"tax_id" db:"tax_id"`
	Logo      *string    `json:"logo" db:"logo"`
	IsActive  bool       `json:"is_active" db:"is_active"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// User represents a user in the system
type User struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	Email          string     `json:"email" db:"email"`
	PasswordHash   string     `json:"-" db:"password_hash"`
	FirstName      string     `json:"first_name" db:"first_name"`
	LastName       string     `json:"last_name" db:"last_name"`
	Phone          *string    `json:"phone" db:"phone"`
	Avatar         *string    `json:"avatar" db:"avatar"`
	Role           Role       `json:"role" db:"role"`
	IsActive       bool       `json:"is_active" db:"is_active"`
	LastLoginAt    *time.Time `json:"last_login_at" db:"last_login_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

type Role string

const (
	RoleAdmin      Role = "admin"
	RoleManager    Role = "manager"
	RoleEmployee   Role = "employee"
	RoleClient     Role = "client"
	RoleAccountant Role = "accountant"
)

// Client represents a customer
type Client struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	Name           string     `json:"name" db:"name"`
	Email          string     `json:"email" db:"email"`
	Phone          string     `json:"phone" db:"phone"`
	Address        *string    `json:"address" db:"address"`
	TaxID          *string    `json:"tax_id" db:"tax_id"`
	Notes          *string    `json:"notes" db:"notes"`
	UserID         *uuid.UUID `json:"user_id" db:"user_id"` // If client has portal access
	CreatedBy      uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// WorkSheet represents a folha de obra
type WorkSheet struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	OrganizationID uuid.UUID       `json:"organization_id" db:"organization_id"`
	ClientID       uuid.UUID       `json:"client_id" db:"client_id"`
	Title          string          `json:"title" db:"title"`
	Description    string          `json:"description" db:"description"`
	Status         WorkSheetStatus `json:"status" db:"status"`
	CreatedBy      uuid.UUID       `json:"created_by" db:"created_by"`
	ReviewedBy     *uuid.UUID      `json:"reviewed_by" db:"reviewed_by"`
	ReviewedAt     *time.Time      `json:"reviewed_at" db:"reviewed_at"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
	DeletedAt      *time.Time      `json:"deleted_at,omitempty" db:"deleted_at"`
}

type WorkSheetStatus string

const (
	WorkSheetStatusDraft      WorkSheetStatus = "draft"
	WorkSheetStatusUnderReview WorkSheetStatus = "under_review"
	WorkSheetStatusApproved   WorkSheetStatus = "approved"
)

// WorkSheetItem represents an item in the worksheet
type WorkSheetItem struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	WorkSheetID uuid.UUID  `json:"worksheet_id" db:"worksheet_id"`
	Description string     `json:"description" db:"description"`
	Quantity    float64    `json:"quantity" db:"quantity"`
	Unit        string     `json:"unit" db:"unit"`
	Notes       *string    `json:"notes" db:"notes"`
	Order       int        `json:"order" db:"order"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Budget represents an orçamento
type Budget struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	OrganizationID uuid.UUID       `json:"organization_id" db:"organization_id"`
	WorkSheetID    uuid.UUID       `json:"worksheet_id" db:"worksheet_id"`
	BudgetNumber   string          `json:"budget_number" db:"budget_number"`
	Status         BudgetStatus    `json:"status" db:"status"`
	Subtotal       decimal.Decimal `json:"subtotal" db:"subtotal"`
	Tax            decimal.Decimal `json:"tax" db:"tax"`
	Total          decimal.Decimal `json:"total" db:"total"`
	ValidUntil     time.Time       `json:"valid_until" db:"valid_until"`
	Notes          *string         `json:"notes" db:"notes"`
	CreatedBy      uuid.UUID       `json:"created_by" db:"created_by"`
	SentAt         *time.Time      `json:"sent_at" db:"sent_at"`
	ApprovedBy     *uuid.UUID      `json:"approved_by" db:"approved_by"`
	ApprovedAt     *time.Time      `json:"approved_at" db:"approved_at"`
	RejectedAt     *time.Time      `json:"rejected_at" db:"rejected_at"`
	RejectionNotes *string         `json:"rejection_notes" db:"rejection_notes"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
	DeletedAt      *time.Time      `json:"deleted_at,omitempty" db:"deleted_at"`
}

type BudgetStatus string

const (
	BudgetStatusDraft    BudgetStatus = "draft"
	BudgetStatusSent     BudgetStatus = "sent"
	BudgetStatusApproved BudgetStatus = "approved"
	BudgetStatusRejected BudgetStatus = "rejected"
	BudgetStatusExpired  BudgetStatus = "expired"
)

// BudgetItem represents an item in the budget
type BudgetItem struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	BudgetID        uuid.UUID       `json:"budget_id" db:"budget_id"`
	WorkSheetItemID *uuid.UUID      `json:"worksheet_item_id" db:"worksheet_item_id"`
	Description     string          `json:"description" db:"description"`
	Quantity        float64         `json:"quantity" db:"quantity"`
	Unit            string          `json:"unit" db:"unit"`
	UnitPrice       decimal.Decimal `json:"unit_price" db:"unit_price"`
	Tax             decimal.Decimal `json:"tax" db:"tax"`
	Total           decimal.Decimal `json:"total" db:"total"`
	Order           int             `json:"order" db:"order"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
	DeletedAt       *time.Time      `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Project represents a construction project
type Project struct {
	ID             uuid.UUID     `json:"id" db:"id"`
	OrganizationID uuid.UUID     `json:"organization_id" db:"organization_id"`
	BudgetID       uuid.UUID     `json:"budget_id" db:"budget_id"`
	ProjectNumber  string        `json:"project_number" db:"project_number"`
	Title          string        `json:"title" db:"title"`
	Description    *string       `json:"description" db:"description"`
	Status         ProjectStatus `json:"status" db:"status"`
	Progress       int           `json:"progress" db:"progress"` // 0-100
	StartDate      time.Time     `json:"start_date" db:"start_date"`
	ExpectedEndDate time.Time    `json:"expected_end_date" db:"expected_end_date"`
	ActualEndDate  *time.Time    `json:"actual_end_date" db:"actual_end_date"`
	CreatedBy      uuid.UUID     `json:"created_by" db:"created_by"`
	CreatedAt      time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at" db:"updated_at"`
	DeletedAt      *time.Time    `json:"deleted_at,omitempty" db:"deleted_at"`
}

type ProjectStatus string

const (
	ProjectStatusInProgress ProjectStatus = "in_progress"
	ProjectStatusOnHold     ProjectStatus = "on_hold"
	ProjectStatusCompleted  ProjectStatus = "completed"
	ProjectStatusCancelled  ProjectStatus = "cancelled"
)

// Task represents a task in a project
type Task struct {
	ID          uuid.UUID   `json:"id" db:"id"`
	ProjectID   uuid.UUID   `json:"project_id" db:"project_id"`
	Title       string      `json:"title" db:"title"`
	Description *string     `json:"description" db:"description"`
	AssignedTo  *uuid.UUID  `json:"assigned_to" db:"assigned_to"`
	Status      TaskStatus  `json:"status" db:"status"`
	Priority    Priority    `json:"priority" db:"priority"`
	DueDate     *time.Time  `json:"due_date" db:"due_date"`
	CompletedAt *time.Time  `json:"completed_at" db:"completed_at"`
	CreatedBy   uuid.UUID   `json:"created_by" db:"created_by"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time  `json:"deleted_at,omitempty" db:"deleted_at"`
}

type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

// Payment represents a payment
type Payment struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	OrganizationID uuid.UUID       `json:"organization_id" db:"organization_id"`
	ProjectID      uuid.UUID       `json:"project_id" db:"project_id"`
	Amount         decimal.Decimal `json:"amount" db:"amount"`
	Status         PaymentStatus   `json:"status" db:"status"`
	DueDate        time.Time       `json:"due_date" db:"due_date"`
	PaidAt         *time.Time      `json:"paid_at" db:"paid_at"`
	Method         *string         `json:"method" db:"method"`
	Reference      *string         `json:"reference" db:"reference"`
	Notes          *string         `json:"notes" db:"notes"`
	CreatedBy      uuid.UUID       `json:"created_by" db:"created_by"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
	DeletedAt      *time.Time      `json:"deleted_at,omitempty" db:"deleted_at"`
}

type PaymentStatus string

const (
	PaymentStatusPending PaymentStatus = "pending"
	PaymentStatusPaid    PaymentStatus = "paid"
	PaymentStatusOverdue PaymentStatus = "overdue"
	PaymentStatusCancelled PaymentStatus = "cancelled"
)

// Photo represents an uploaded photo
type Photo struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	EntityType     string     `json:"entity_type" db:"entity_type"` // worksheet, budget, project, task
	EntityID       uuid.UUID  `json:"entity_id" db:"entity_id"`
	FileName       string     `json:"file_name" db:"file_name"`
	FileSize       int64      `json:"file_size" db:"file_size"`
	MimeType       string     `json:"mime_type" db:"mime_type"`
	URL            string     `json:"url" db:"url"`
	ThumbnailURL   *string    `json:"thumbnail_url" db:"thumbnail_url"`
	Caption        *string    `json:"caption" db:"caption"`
	UploadedBy     uuid.UUID  `json:"uploaded_by" db:"uploaded_by"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Notification represents a notification
type Notification struct {
	ID        uuid.UUID        `json:"id" db:"id"`
	UserID    uuid.UUID        `json:"user_id" db:"user_id"`
	Type      NotificationType `json:"type" db:"type"`
	Title     string           `json:"title" db:"title"`
	Message   string           `json:"message" db:"message"`
	EntityType *string         `json:"entity_type" db:"entity_type"`
	EntityID   *uuid.UUID      `json:"entity_id" db:"entity_id"`
	IsRead    bool             `json:"is_read" db:"is_read"`
	ReadAt    *time.Time       `json:"read_at" db:"read_at"`
	CreatedAt time.Time        `json:"created_at" db:"created_at"`
}

type NotificationType string

const (
	NotificationTypeWorkSheetReview NotificationType = "worksheet_review"
	NotificationTypeBudgetSent      NotificationType = "budget_sent"
	NotificationTypeBudgetApproved  NotificationType = "budget_approved"
	NotificationTypeTaskAssigned    NotificationType = "task_assigned"
	NotificationTypeTaskDue         NotificationType = "task_due"
	NotificationTypePaymentDue      NotificationType = "payment_due"
	NotificationTypeProjectUpdate   NotificationType = "project_update"
)
