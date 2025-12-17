package handlers

import (
	"net/http"

	"github.com/controlwise/backend/internal/middleware"
	"github.com/controlwise/backend/internal/models"
	"github.com/controlwise/backend/internal/services"
	"github.com/controlwise/backend/internal/utils"
)

type NotificationConfigHandler struct {
	whatsappService *services.WhatsAppService
}

func NewNotificationConfigHandler(whatsappService *services.WhatsAppService) *NotificationConfigHandler {
	return &NotificationConfigHandler{whatsappService: whatsappService}
}

// GetConfig returns the notification configuration for the organization
func (h *NotificationConfigHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	config, err := h.whatsappService.GetConfig(r.Context(), orgID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if config == nil {
		// Return default config structure
		utils.SuccessResponse(w, http.StatusOK, map[string]interface{}{
			"config": map[string]interface{}{
				"whatsapp_enabled":     false,
				"twilio_configured":    false,
				"reminder_24h_enabled": true,
				"reminder_2h_enabled":  true,
			},
		})
		return
	}

	utils.SuccessResponse(w, http.StatusOK, map[string]interface{}{
		"config": config.ToPublic(),
	})
}

// UpdateConfig updates the notification configuration
func (h *NotificationConfigHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	// Check if user is admin
	role, ok := middleware.GetUserRole(r.Context())
	if !ok || role != string(models.RoleAdmin) {
		utils.ErrorResponse(w, http.StatusForbidden, "Only administrators can update notification settings")
		return
	}

	var req services.NotificationConfigInput
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.whatsappService.SaveConfig(r.Context(), orgID, &req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get updated config
	config, err := h.whatsappService.GetConfig(r.Context(), orgID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessMessageResponse(w, http.StatusOK, "Notification settings updated successfully", config.ToPublic())
}

// TestWhatsApp sends a test message to verify configuration
func (h *NotificationConfigHandler) TestWhatsApp(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Organization not found")
		return
	}

	var req struct {
		PhoneNumber string `json:"phone_number"`
	}
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.PhoneNumber == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Phone number is required")
		return
	}

	testMessage := "Esta e uma mensagem de teste do controlwise. Se recebeu esta mensagem, a configuracao do WhatsApp esta correta!"

	_, err := h.whatsappService.SendMessage(r.Context(), orgID, req.PhoneNumber, testMessage, nil)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessMessageResponse(w, http.StatusOK, "Test message sent successfully", nil)
}
