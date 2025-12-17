package router

import (
	"net/http"
	"time"

	"github.com/controlewise/backend/internal/config"
	"github.com/controlewise/backend/internal/handlers"
	"github.com/controlewise/backend/internal/middleware"
	"github.com/controlewise/backend/internal/models"
	"github.com/controlewise/backend/internal/services"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
)

func Setup(services *services.Services, cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	// Basic middleware
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.Timeout(60 * time.Second))

	// Security headers
	r.Use(securityHeaders)

	// Rate limiting - 100 requests per minute per IP
	r.Use(httprate.LimitByIP(100, time.Minute))

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.App.FrontendURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Custom middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWT.Secret)
	orgMiddleware := middleware.NewOrganizationMiddleware(services.Organization)

	// Initialize module middleware
	moduleMiddleware := middleware.NewModuleMiddleware(services.Module)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(services.Auth)
	organizationHandler := handlers.NewOrganizationHandler(services.Organization)
	userHandler := handlers.NewUserHandler(services.User)
	clientHandler := handlers.NewClientHandler(services.Client)
	worksheetHandler := handlers.NewWorksheetHandler(services.Worksheet)
	budgetHandler := handlers.NewBudgetHandler(services.Budget)
	projectHandler := handlers.NewProjectHandler(services.Project)
	taskHandler := handlers.NewTaskHandler(services.Task)
	paymentHandler := handlers.NewPaymentHandler(services.Payment)
	notificationHandler := handlers.NewNotificationHandler(services.Notification)
	reportHandler := handlers.NewReportHandler(services.Report)
	moduleHandler := handlers.NewModuleHandler(services.Module)
	// Appointments module handlers
	patientHandler := handlers.NewPatientHandler(services.Patient)
	therapistHandler := handlers.NewTherapistHandler(services.Therapist)
	sessionHandler := handlers.NewSessionHandler(services.Session)
	sessionPaymentHandler := handlers.NewSessionPaymentHandler(services.SessionPayment)
	// Notifications module handlers
	notificationConfigHandler := handlers.NewNotificationConfigHandler(services.WhatsApp)
	webhookHandler := handlers.NewWebhookHandler(services.WhatsApp)
	// Workflow engine handler
	workflowHandler := handlers.NewWorkflowHandler(services.Workflow)
	// System Admin handlers
	adminAuthHandler := handlers.NewAdminAuthHandler(services.SystemAdmin)
	adminOrgsHandler := handlers.NewAdminOrganizationsHandler(services.AdminOrganization, services.AdminAudit, services.Module)
	adminUsersHandler := handlers.NewAdminUsersHandler(services.AdminUser, services.AdminAudit)
	adminImpersonationHandler := handlers.NewAdminImpersonationHandler(services.Impersonation, services.AdminAudit)
	adminDashboardHandler := handlers.NewAdminDashboardHandler(services.AdminStats)
	adminAuditHandler := handlers.NewAdminAuditHandler(services.AdminAudit)

	// Public routes
	r.Group(func(r chi.Router) {
		r.Get("/health", healthCheck)
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)
		r.Post("/auth/forgot-password", authHandler.ForgotPassword)
		r.Post("/auth/reset-password", authHandler.ResetPassword)

		// Twilio webhooks (public endpoints)
		r.Route("/webhooks", func(r chi.Router) {
			r.Post("/whatsapp", webhookHandler.TwilioIncoming)
			r.Post("/whatsapp/status", webhookHandler.TwilioStatus)
		})

		// System Admin public routes (login only)
		r.Post("/admin/auth/login", adminAuthHandler.Login)
	})

	// System Admin protected routes
	r.Route("/admin", func(r chi.Router) {
		r.Use(authMiddleware.Authenticate)
		r.Use(middleware.RequireSystemAdmin)

		// Auth
		r.Get("/auth/me", adminAuthHandler.Me)
		r.Post("/auth/change-password", adminAuthHandler.ChangePassword)
		r.Post("/auth/logout", adminAuthHandler.Logout)

		// Dashboard
		r.Get("/dashboard/stats", adminDashboardHandler.GetStats)
		r.Get("/dashboard/recent-activity", adminDashboardHandler.GetRecentActivity)

		// Organizations
		r.Route("/organizations", func(r chi.Router) {
			r.Get("/", adminOrgsHandler.List)
			r.Post("/", adminOrgsHandler.Create)
			r.Get("/{id}", adminOrgsHandler.GetByID)
			r.Put("/{id}", adminOrgsHandler.Update)
			r.Post("/{id}/suspend", adminOrgsHandler.Suspend)
			r.Post("/{id}/reactivate", adminOrgsHandler.Reactivate)
			r.Delete("/{id}", adminOrgsHandler.Delete)
			r.Get("/{id}/users", adminUsersHandler.ListByOrganization)
			// Module management for organization
			r.Get("/{id}/modules", adminOrgsHandler.ListModules)
			r.Post("/{id}/modules/{module}/enable", adminOrgsHandler.EnableModule)
			r.Post("/{id}/modules/{module}/disable", adminOrgsHandler.DisableModule)
		})

		// Users
		r.Route("/users", func(r chi.Router) {
			r.Get("/", adminUsersHandler.List)
			r.Get("/{id}", adminUsersHandler.GetByID)
			r.Post("/{id}/suspend", adminUsersHandler.Suspend)
			r.Post("/{id}/reactivate", adminUsersHandler.Reactivate)
			r.Post("/{id}/reset-password", adminUsersHandler.ResetPassword)
		})

		// Impersonation
		r.Post("/impersonate/{userId}", adminImpersonationHandler.Start)
		r.Get("/impersonate/active", adminImpersonationHandler.GetActiveSession)
		r.Get("/impersonate/sessions", adminImpersonationHandler.ListSessions)

		// Audit Logs
		r.Get("/audit-logs", adminAuditHandler.List)
	})

	// End impersonation route (available during impersonation with regular user token)
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.Authenticate)
		r.Post("/admin/impersonate/end", adminImpersonationHandler.End)
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.Authenticate)
		r.Use(orgMiddleware.ExtractOrganization)

		// Auth
		r.Get("/auth/me", authHandler.Me)
		r.Post("/auth/refresh", authHandler.RefreshToken)
		r.Post("/auth/logout", authHandler.Logout)

		// Organizations
		r.Route("/organizations", func(r chi.Router) {
			r.Get("/", organizationHandler.GetCurrent)
			r.Put("/", organizationHandler.Update)
			r.Post("/logo", organizationHandler.UploadLogo)
		})

		// Modules
		r.Route("/modules", func(r chi.Router) {
			r.Get("/available", moduleHandler.ListAvailable)
			r.Get("/", moduleHandler.ListOrganizationModules)
			r.Get("/enabled", moduleHandler.GetEnabledModules)
			r.Post("/{module}/enable", moduleHandler.EnableModule)
			r.Post("/{module}/disable", moduleHandler.DisableModule)
			r.Get("/{module}/config", moduleHandler.GetModuleConfig)
			r.Put("/{module}/config", moduleHandler.UpdateModuleConfig)
		})

		// Users
		r.Route("/users", func(r chi.Router) {
			r.Get("/", userHandler.List)
			r.Post("/", userHandler.Create)
			r.Get("/{id}", userHandler.Get)
			r.Put("/{id}", userHandler.Update)
			r.Delete("/{id}", userHandler.Delete)
		})

		// Clients (Core feature - available to all organizations)
		r.Route("/clients", func(r chi.Router) {
			r.Get("/", clientHandler.List)
			r.Post("/", clientHandler.Create)
			r.Get("/{id}", clientHandler.Get)
			r.Put("/{id}", clientHandler.Update)
			r.Delete("/{id}", clientHandler.Delete)
		})

		// Worksheets (Construction module)
		r.Route("/worksheets", func(r chi.Router) {
			r.Use(moduleMiddleware.RequireModule(models.ModuleConstruction))
			r.Get("/", worksheetHandler.List)
			r.Post("/", worksheetHandler.Create)
			r.Get("/{id}", worksheetHandler.Get)
			r.Put("/{id}", worksheetHandler.Update)
			r.Delete("/{id}", worksheetHandler.Delete)
			r.Post("/{id}/review", worksheetHandler.Review)
			r.Post("/{id}/photos", worksheetHandler.UploadPhoto)
			r.Get("/{id}/photos", worksheetHandler.ListPhotos)
		})

		// Budgets (Construction module)
		r.Route("/budgets", func(r chi.Router) {
			r.Use(moduleMiddleware.RequireModule(models.ModuleConstruction))
			r.Get("/", budgetHandler.List)
			r.Post("/", budgetHandler.Create)
			r.Get("/{id}", budgetHandler.Get)
			r.Put("/{id}", budgetHandler.Update)
			r.Delete("/{id}", budgetHandler.Delete)
			r.Post("/{id}/send", budgetHandler.Send)
			r.Post("/{id}/approve", budgetHandler.Approve)
			r.Post("/{id}/reject", budgetHandler.Reject)
			r.Post("/{id}/photos", budgetHandler.UploadPhoto)
			r.Get("/{id}/photos", budgetHandler.ListPhotos)
			r.Get("/{id}/pdf", budgetHandler.GeneratePDF)
		})

		// Projects (Construction module)
		r.Route("/projects", func(r chi.Router) {
			r.Use(moduleMiddleware.RequireModule(models.ModuleConstruction))
			r.Get("/", projectHandler.List)
			r.Post("/", projectHandler.Create)
			r.Get("/{id}", projectHandler.Get)
			r.Put("/{id}", projectHandler.Update)
			r.Delete("/{id}", projectHandler.Delete)
			r.Patch("/{id}/status", projectHandler.UpdateStatus)
			r.Patch("/{id}/progress", projectHandler.UpdateProgress)
			r.Post("/{id}/photos", projectHandler.UploadPhoto)
			r.Get("/{id}/photos", projectHandler.ListPhotos)
		})

		// Tasks (Construction module)
		r.Route("/tasks", func(r chi.Router) {
			r.Use(moduleMiddleware.RequireModule(models.ModuleConstruction))
			r.Get("/", taskHandler.List)
			r.Post("/", taskHandler.Create)
			r.Get("/{id}", taskHandler.Get)
			r.Put("/{id}", taskHandler.Update)
			r.Delete("/{id}", taskHandler.Delete)
			r.Patch("/{id}/status", taskHandler.UpdateStatus)
			r.Patch("/{id}/assign", taskHandler.Assign)
		})

		// Payments (Construction module)
		r.Route("/payments", func(r chi.Router) {
			r.Use(moduleMiddleware.RequireModule(models.ModuleConstruction))
			r.Get("/", paymentHandler.List)
			r.Post("/", paymentHandler.Create)
			r.Get("/{id}", paymentHandler.Get)
			r.Put("/{id}", paymentHandler.Update)
			r.Delete("/{id}", paymentHandler.Delete)
			r.Post("/{id}/mark-paid", paymentHandler.MarkAsPaid)
		})

		// ============ Appointments Module ============

		// Patients (Appointments module)
		r.Route("/patients", func(r chi.Router) {
			r.Use(moduleMiddleware.RequireModule(models.ModuleAppointments))
			r.Get("/", patientHandler.List)
			r.Post("/", patientHandler.Create)
			r.Get("/stats", patientHandler.GetStats)
			r.Get("/{id}", patientHandler.Get)
			r.Put("/{id}", patientHandler.Update)
			r.Delete("/{id}", patientHandler.Delete)
			r.Get("/{id}/payments", sessionPaymentHandler.ListByPatient)
		})

		// Therapists (Appointments module)
		r.Route("/therapists", func(r chi.Router) {
			r.Use(moduleMiddleware.RequireModule(models.ModuleAppointments))
			r.Get("/", therapistHandler.List)
			r.Post("/", therapistHandler.Create)
			r.Get("/stats", therapistHandler.GetStats)
			r.Get("/{id}", therapistHandler.Get)
			r.Put("/{id}", therapistHandler.Update)
			r.Delete("/{id}", therapistHandler.Delete)
		})

		// Sessions (Appointments module)
		r.Route("/sessions", func(r chi.Router) {
			r.Use(moduleMiddleware.RequireModule(models.ModuleAppointments))
			r.Get("/", sessionHandler.List)
			r.Get("/calendar", sessionHandler.GetCalendar)
			r.Get("/stats", sessionHandler.GetStats)
			r.Post("/", sessionHandler.Create)
			r.Get("/{id}", sessionHandler.Get)
			r.Put("/{id}", sessionHandler.Update)
			r.Delete("/{id}", sessionHandler.Delete)
			r.Post("/{id}/confirm", sessionHandler.Confirm)
			r.Post("/{id}/cancel", sessionHandler.Cancel)
			r.Post("/{id}/complete", sessionHandler.Complete)
			r.Post("/{id}/no-show", sessionHandler.MarkNoShow)
			// Session payments
			r.Get("/{id}/payment", sessionPaymentHandler.GetSessionPayment)
			r.Put("/{id}/payment", sessionPaymentHandler.UpdateSessionPayment)
			r.Post("/{id}/payment/mark-paid", sessionPaymentHandler.MarkAsPaid)
		})

		// Session Payments (Appointments module)
		r.Route("/session-payments", func(r chi.Router) {
			r.Use(moduleMiddleware.RequireModule(models.ModuleAppointments))
			r.Get("/unpaid", sessionPaymentHandler.ListUnpaid)
			r.Get("/stats", sessionPaymentHandler.GetPaymentStats)
		})

		// ============ Notifications Module ============

		// Notification Configuration (Notifications module)
		r.Route("/notification-config", func(r chi.Router) {
			r.Use(moduleMiddleware.RequireModule(models.ModuleNotifications))
			r.Get("/", notificationConfigHandler.GetConfig)
			r.Put("/", notificationConfigHandler.UpdateConfig)
			r.Post("/test", notificationConfigHandler.TestWhatsApp)
		})

		// Notifications
		r.Route("/notifications", func(r chi.Router) {
			r.Get("/", notificationHandler.List)
			r.Get("/unread-count", notificationHandler.UnreadCount)
			r.Post("/{id}/read", notificationHandler.MarkAsRead)
			r.Post("/read-all", notificationHandler.MarkAllAsRead)
		})

		// Reports
		r.Route("/reports", func(r chi.Router) {
			r.Get("/dashboard", reportHandler.Dashboard)
			r.Get("/projects", reportHandler.Projects)
			r.Get("/financials", reportHandler.Financials)
			r.Get("/clients", reportHandler.Clients)
			r.Get("/tasks", reportHandler.Tasks)
		})

		// ============ Workflow Engine ============

		// Workflows
		r.Route("/workflows", func(r chi.Router) {
			r.Get("/", workflowHandler.ListWorkflows)
			r.Post("/", workflowHandler.CreateWorkflow)
			r.Post("/init-defaults", workflowHandler.InitDefaultWorkflows)
			r.Get("/{id}", workflowHandler.GetWorkflow)
			r.Put("/{id}", workflowHandler.UpdateWorkflow)
			r.Delete("/{id}", workflowHandler.DeleteWorkflow)
			r.Post("/{id}/duplicate", workflowHandler.DuplicateWorkflow)
			// States
			r.Post("/{id}/states", workflowHandler.CreateState)
			r.Put("/{id}/states/{stateId}", workflowHandler.UpdateState)
			r.Delete("/{id}/states/{stateId}", workflowHandler.DeleteState)
			r.Put("/{id}/states/reorder", workflowHandler.ReorderStates)
			// Triggers
			r.Post("/{id}/triggers", workflowHandler.CreateTrigger)
		})

		// Triggers (standalone routes for update/delete)
		r.Route("/triggers", func(r chi.Router) {
			r.Put("/{triggerId}", workflowHandler.UpdateTrigger)
			r.Delete("/{triggerId}", workflowHandler.DeleteTrigger)
			r.Post("/{triggerId}/actions", workflowHandler.CreateAction)
		})

		// Actions (standalone routes for update/delete)
		r.Route("/actions", func(r chi.Router) {
			r.Put("/{actionId}", workflowHandler.UpdateAction)
			r.Delete("/{actionId}", workflowHandler.DeleteAction)
		})

		// Message Templates
		r.Route("/templates", func(r chi.Router) {
			r.Get("/", workflowHandler.ListTemplates)
			r.Post("/", workflowHandler.CreateTemplate)
			r.Get("/{id}", workflowHandler.GetTemplate)
			r.Put("/{id}", workflowHandler.UpdateTemplate)
			r.Delete("/{id}", workflowHandler.DeleteTemplate)
		})

		// Execution Logs & Scheduled Jobs
		r.Get("/execution-logs", workflowHandler.GetExecutionLogs)
		r.Get("/scheduled-jobs", workflowHandler.GetScheduledJobs)

		// Testing & Variables
		r.Get("/variables", workflowHandler.GetAvailableVariables)
		r.Post("/{id}/triggers/{triggerId}/test", workflowHandler.TestTrigger)
	})

	return r
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
}

// securityHeaders adds security-related headers to all responses
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "DENY")
		// Enable XSS filter in browsers
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		// Enforce HTTPS
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		// Prevent information leakage
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		// Content Security Policy
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		next.ServeHTTP(w, r)
	})
}
