package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/controlewise/backend/internal/middleware"
	"github.com/controlewise/backend/internal/models"
	"github.com/controlewise/backend/internal/services"
	"github.com/controlewise/backend/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type SessionPaymentHandler struct {
	service *services.SessionPaymentService
}

func NewSessionPaymentHandler(service *services.SessionPaymentService) *SessionPaymentHandler {
	return &SessionPaymentHandler{service: service}
}

// GetSessionPayment returns the payment record for a session
func (h *SessionPaymentHandler) GetSessionPayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID, ok := middleware.GetOrganizationID(ctx)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	sessionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	payment, err := h.service.GetBySessionID(ctx, sessionID, orgID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if payment == nil {
		// Return empty payment with session defaults
		utils.SuccessResponse(w, http.StatusOK, nil)
		return
	}

	utils.SuccessResponse(w, http.StatusOK, payment)
}

// UpdateSessionPaymentRequest is the request body for updating a session payment
type UpdateSessionPaymentRequest struct {
	AmountCents          int     `json:"amount_cents"`
	PaymentStatus        string  `json:"payment_status"`
	PaymentMethod        *string `json:"payment_method"`
	InsuranceProvider    *string `json:"insurance_provider"`
	InsuranceAmountCents *int    `json:"insurance_amount_cents"`
	DueDate              *string `json:"due_date"`
	Notes                *string `json:"notes"`
}

// UpdateSessionPayment creates or updates a session payment
func (h *SessionPaymentHandler) UpdateSessionPayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID, ok := middleware.GetOrganizationID(ctx)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	sessionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	var req UpdateSessionPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	payment := &models.SessionPayment{
		AmountCents:          req.AmountCents,
		PaymentStatus:        models.SessionPaymentStatus(req.PaymentStatus),
		InsuranceProvider:    req.InsuranceProvider,
		InsuranceAmountCents: req.InsuranceAmountCents,
		Notes:                req.Notes,
	}

	if req.PaymentMethod != nil {
		method := models.PaymentMethod(*req.PaymentMethod)
		payment.PaymentMethod = &method
	}

	if req.DueDate != nil && *req.DueDate != "" {
		dueDate, err := time.Parse("2006-01-02", *req.DueDate)
		if err == nil {
			payment.DueDate = &dueDate
		}
	}

	// If marking as paid, set paid_at
	if req.PaymentStatus == string(models.SessionPaymentStatusPaid) {
		now := time.Now()
		payment.PaidAt = &now
	}

	if err := h.service.CreateOrUpdate(ctx, sessionID, orgID, payment); err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Get updated payment
	updated, err := h.service.GetBySessionID(ctx, sessionID, orgID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(w, http.StatusOK, updated)
}

// MarkAsPaid marks a session payment as paid
func (h *SessionPaymentHandler) MarkAsPaid(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID, ok := middleware.GetOrganizationID(ctx)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	sessionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	var req struct {
		PaymentMethod *string `json:"payment_method"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Default to no method specified
		req.PaymentMethod = nil
	}

	var method *models.PaymentMethod
	if req.PaymentMethod != nil {
		m := models.PaymentMethod(*req.PaymentMethod)
		method = &m
	}

	if err := h.service.MarkAsPaid(ctx, sessionID, orgID, method); err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessMessageResponse(w, http.StatusOK, "Payment marked as paid", nil)
}

// ListUnpaid returns all unpaid session payments
func (h *SessionPaymentHandler) ListUnpaid(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID, ok := middleware.GetOrganizationID(ctx)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	// Parse query parameters
	filters := services.SessionPaymentFilters{
		Limit:  20,
		Offset: 0,
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 && l <= 100 {
			filters.Limit = l
		}
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			filters.Offset = o
		}
	}

	if therapistID := r.URL.Query().Get("therapist_id"); therapistID != "" {
		if id, err := uuid.Parse(therapistID); err == nil {
			filters.TherapistID = &id
		}
	}

	if patientID := r.URL.Query().Get("patient_id"); patientID != "" {
		if id, err := uuid.Parse(patientID); err == nil {
			filters.PatientID = &id
		}
	}

	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			filters.StartDate = &t
		}
	}

	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			filters.EndDate = &t
		}
	}

	payments, total, err := h.service.ListUnpaid(ctx, orgID, filters)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(w, http.StatusOK, map[string]interface{}{
		"payments": payments,
		"total":    total,
		"limit":    filters.Limit,
		"offset":   filters.Offset,
	})
}

// GetPaymentStats returns payment statistics
func (h *SessionPaymentHandler) GetPaymentStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID, ok := middleware.GetOrganizationID(ctx)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	var startDate, endDate *time.Time

	if start := r.URL.Query().Get("start_date"); start != "" {
		if t, err := time.Parse("2006-01-02", start); err == nil {
			startDate = &t
		}
	}

	if end := r.URL.Query().Get("end_date"); end != "" {
		if t, err := time.Parse("2006-01-02", end); err == nil {
			endDate = &t
		}
	}

	stats, err := h.service.GetPaymentStats(ctx, orgID, startDate, endDate)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(w, http.StatusOK, stats)
}

// ListByPatient returns all payment records for a patient
func (h *SessionPaymentHandler) ListByPatient(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID, ok := middleware.GetOrganizationID(ctx)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	patientID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid patient ID")
		return
	}

	payments, err := h.service.ListByPatient(ctx, patientID, orgID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(w, http.StatusOK, map[string]interface{}{
		"payments": payments,
	})
}
