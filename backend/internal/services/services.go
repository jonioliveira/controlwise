package services

import (
	"github.com/controlwise/backend/internal/config"
	"github.com/controlwise/backend/internal/database"
)

type Services struct {
	Auth         *AuthService
	Organization *OrganizationService
	User         *UserService
	Client       *ClientService
	Worksheet    *WorksheetService
	Budget       *BudgetService
	Project      *ProjectService
	Task         *TaskService
	Payment      *PaymentService
	Notification *NotificationService
	Report       *ReportService
	Storage      *StorageService
	Email        *EmailService
	Module       *ModuleService
	// Appointments module
	Patient        *PatientService
	Therapist      *TherapistService
	Session        *SessionService
	SessionPayment *SessionPaymentService
	// Notifications module
	WhatsApp *WhatsAppService
	// Workflow engine
	Workflow *WorkflowService
	// System Admin services
	SystemAdmin       *SystemAdminService
	AdminOrganization *AdminOrganizationService
	AdminUser         *AdminUserService
	AdminAudit        *AdminAuditService
	AdminStats        *AdminStatsService
	Impersonation     *ImpersonationService
}

func NewServices(db *database.DB, redis *database.Redis, cfg *config.Config) *Services {
	// Initialize storage service
	storageService := NewStorageService(cfg.Storage)

	// Initialize email service
	emailService := NewEmailService(cfg.Email)

	// Initialize notification service
	notificationService := NewNotificationService(db, emailService)

	// Initialize system admin service
	systemAdminService := NewSystemAdminService(db, cfg.JWT)

	// Initialize workflow service
	workflowService := NewWorkflowService(db)

	// Initialize session service with workflow integration
	sessionService := NewSessionService(db)
	sessionService.SetWorkflowService(workflowService)

	// Initialize budget service with workflow integration
	budgetService := NewBudgetService(db, storageService, notificationService)
	budgetService.SetWorkflowService(workflowService)

	return &Services{
		Auth:         NewAuthService(db, cfg.JWT),
		Organization: NewOrganizationService(db),
		User:         NewUserService(db),
		Client:       NewClientService(db),
		Worksheet:    NewWorksheetService(db, storageService, notificationService),
		Budget:       budgetService,
		Project:      NewProjectService(db, storageService, notificationService),
		Task:         NewTaskService(db, notificationService),
		Payment:      NewPaymentService(db, notificationService),
		Notification: notificationService,
		Report:       NewReportService(db),
		Storage:      storageService,
		Email:        emailService,
		Module:       NewModuleService(db),
		// Appointments module
		Patient:        NewPatientService(db),
		Therapist:      NewTherapistService(db),
		Session:        sessionService,
		SessionPayment: NewSessionPaymentService(db),
		// Notifications module
		WhatsApp: NewWhatsAppService(db, cfg.Encryption.Key),
		// Workflow engine
		Workflow: workflowService,
		// System Admin services
		SystemAdmin:       systemAdminService,
		AdminOrganization: NewAdminOrganizationService(db),
		AdminUser:         NewAdminUserService(db),
		AdminAudit:        NewAdminAuditService(db),
		AdminStats:        NewAdminStatsService(db),
		Impersonation:     NewImpersonationService(db, systemAdminService),
	}
}
