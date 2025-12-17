package handlers

import (
	"net/http"

	"github.com/controlwise/backend/internal/services"
	"github.com/controlwise/backend/internal/utils"
	"github.com/google/uuid"
)

type WebhookHandler struct {
	whatsappService *services.WhatsAppService
}

func NewWebhookHandler(whatsappService *services.WhatsAppService) *WebhookHandler {
	return &WebhookHandler{whatsappService: whatsappService}
}

// TwilioIncoming handles incoming WhatsApp messages from Twilio
func (h *WebhookHandler) TwilioIncoming(w http.ResponseWriter, r *http.Request) {
	// Parse Twilio webhook data
	if err := r.ParseForm(); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid form data")
		return
	}

	from := r.FormValue("From")
	body := r.FormValue("Body")
	messageSID := r.FormValue("MessageSid")
	accountSID := r.FormValue("AccountSid")

	if from == "" || body == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	// Find organization by Twilio account SID
	// In a real implementation, you'd look up the org by the account SID
	// For now, we'll use a header or query param
	orgIDStr := r.URL.Query().Get("org_id")
	if orgIDStr == "" {
		// Try to find org by account SID
		// This would require a database lookup
		_ = accountSID // Suppress unused warning
		utils.ErrorResponse(w, http.StatusBadRequest, "Organization ID required")
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid organization ID")
		return
	}

	// Process the incoming message
	if err := h.whatsappService.ProcessIncomingMessage(r.Context(), orgID, from, body, messageSID); err != nil {
		// Log error but don't fail - Twilio expects 200 OK
		// In production, log this error properly
	}

	// Twilio expects 200 OK with optional TwiML response
	w.Header().Set("Content-Type", "text/xml")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("<Response></Response>"))
}

// TwilioStatus handles message status callbacks from Twilio
func (h *WebhookHandler) TwilioStatus(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid form data")
		return
	}

	messageSID := r.FormValue("MessageSid")
	messageStatus := r.FormValue("MessageStatus")

	if messageSID == "" || messageStatus == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	// Update message status in database
	if err := h.whatsappService.UpdateMessageStatus(r.Context(), messageSID, messageStatus); err != nil {
		// Log error but don't fail
	}

	w.WriteHeader(http.StatusOK)
}
