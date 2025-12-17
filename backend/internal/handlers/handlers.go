package handlers

import (
	"net/http"

	"github.com/controlewise/backend/internal/middleware"
	"github.com/controlewise/backend/internal/models"
	"github.com/controlewise/backend/internal/services"
	"github.com/controlewise/backend/internal/utils"
)

// OrganizationHandler
type OrganizationHandler struct {
	service *services.OrganizationService
}

func NewOrganizationHandler(service *services.OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{service: service}
}

func (h *OrganizationHandler) GetCurrent(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found in token")
		return
	}

	org, err := h.service.GetByID(r.Context(), orgID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to get organization")
		return
	}

	utils.SuccessResponse(w, http.StatusOK, org)
}

type UpdateOrganizationRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
	TaxID   string `json:"tax_id"`
}

func (h *OrganizationHandler) Update(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found in token")
		return
	}

	// Check if user is admin or owner
	role, ok := middleware.GetUserRole(r.Context())
	if !ok || (role != string(models.RoleAdmin) && role != "owner") {
		utils.ErrorResponse(w, http.StatusForbidden, "Only administrators and owners can update organization settings")
		return
	}

	var req UpdateOrganizationRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	org := &models.Organization{
		Name:    req.Name,
		Email:   req.Email,
		Phone:   req.Phone,
		Address: req.Address,
		TaxID:   req.TaxID,
	}

	if err := h.service.Update(r.Context(), orgID, org); err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to update organization")
		return
	}

	// Fetch the updated organization
	updatedOrg, err := h.service.GetByID(r.Context(), orgID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to get updated organization")
		return
	}

	utils.SuccessResponse(w, http.StatusOK, updatedOrg)
}

func (h *OrganizationHandler) UploadLogo(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "Upload logo"})
}

// UserHandler
type UserHandler struct {
	service *services.UserService
}

func NewUserHandler(service *services.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, []interface{}{})
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusCreated, map[string]string{"message": "User created"})
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "Get user"})
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "User updated"})
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "User deleted"})
}

// Note: ClientHandler is defined in client.go

// WorksheetHandler
type WorksheetHandler struct {
	service *services.WorksheetService
}

func NewWorksheetHandler(service *services.WorksheetService) *WorksheetHandler {
	return &WorksheetHandler{service: service}
}

func (h *WorksheetHandler) List(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, []interface{}{})
}

func (h *WorksheetHandler) Create(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusCreated, map[string]string{"message": "Worksheet created"})
}

func (h *WorksheetHandler) Get(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "Get worksheet"})
}

func (h *WorksheetHandler) Update(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "Worksheet updated"})
}

func (h *WorksheetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "Worksheet deleted"})
}

func (h *WorksheetHandler) Review(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "Worksheet reviewed"})
}

func (h *WorksheetHandler) UploadPhoto(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "Photo uploaded"})
}

func (h *WorksheetHandler) ListPhotos(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, []interface{}{})
}

// BudgetHandler
type BudgetHandler struct {
	service *services.BudgetService
}

func NewBudgetHandler(service *services.BudgetService) *BudgetHandler {
	return &BudgetHandler{service: service}
}

func (h *BudgetHandler) List(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, []interface{}{})
}

func (h *BudgetHandler) Create(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusCreated, map[string]string{"message": "Budget created"})
}

func (h *BudgetHandler) Get(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "Get budget"})
}

func (h *BudgetHandler) Update(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "Budget updated"})
}

func (h *BudgetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "Budget deleted"})
}

func (h *BudgetHandler) Send(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "Budget sent"})
}

func (h *BudgetHandler) Approve(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "Budget approved"})
}

func (h *BudgetHandler) Reject(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "Budget rejected"})
}

func (h *BudgetHandler) UploadPhoto(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "Photo uploaded"})
}

func (h *BudgetHandler) ListPhotos(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, []interface{}{})
}

func (h *BudgetHandler) GeneratePDF(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "PDF generated"})
}

// ProjectHandler
type ProjectHandler struct {
	service *services.ProjectService
}

func NewProjectHandler(service *services.ProjectService) *ProjectHandler {
	return &ProjectHandler{service: service}
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, []interface{}{})
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusCreated, map[string]string{"message": "Project created"})
}

func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "Get project"})
}

func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "Project updated"})
}

func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "Project deleted"})
}

func (h *ProjectHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "Project status updated"})
}

func (h *ProjectHandler) UpdateProgress(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "Project progress updated"})
}

func (h *ProjectHandler) UploadPhoto(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "Photo uploaded"})
}

func (h *ProjectHandler) ListPhotos(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, []interface{}{})
}

// TaskHandler
type TaskHandler struct {
	service *services.TaskService
}

func NewTaskHandler(service *services.TaskService) *TaskHandler {
	return &TaskHandler{service: service}
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, []interface{}{})
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusCreated, map[string]string{"message": "Task created"})
}

func (h *TaskHandler) Get(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "Get task"})
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "Task updated"})
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "Task deleted"})
}

func (h *TaskHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "Task status updated"})
}

func (h *TaskHandler) Assign(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "Task assigned"})
}

// PaymentHandler
type PaymentHandler struct {
	service *services.PaymentService
}

func NewPaymentHandler(service *services.PaymentService) *PaymentHandler {
	return &PaymentHandler{service: service}
}

func (h *PaymentHandler) List(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, []interface{}{})
}

func (h *PaymentHandler) Create(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusCreated, map[string]string{"message": "Payment created"})
}

func (h *PaymentHandler) Get(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "Get payment"})
}

func (h *PaymentHandler) Update(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "Payment updated"})
}

func (h *PaymentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "Payment deleted"})
}

func (h *PaymentHandler) MarkAsPaid(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "Payment marked as paid"})
}

// NotificationHandler
type NotificationHandler struct {
	service *services.NotificationService
}

func NewNotificationHandler(service *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: service}
}

func (h *NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, []interface{}{})
}

func (h *NotificationHandler) UnreadCount(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]int{"count": 0})
}

func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "Notification marked as read"})
}

func (h *NotificationHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "All notifications marked as read"})
}

// ReportHandler
type ReportHandler struct {
	service *services.ReportService
}

func NewReportHandler(service *services.ReportService) *ReportHandler {
	return &ReportHandler{service: service}
}

func (h *ReportHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "Dashboard report"})
}

func (h *ReportHandler) Projects(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "Projects report"})
}

func (h *ReportHandler) Financials(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "Financials report"})
}

func (h *ReportHandler) Clients(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "Clients report"})
}

func (h *ReportHandler) Tasks(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, map[string]string{"message": "Tasks report"})
}
