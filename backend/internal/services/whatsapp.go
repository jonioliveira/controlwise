package services

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/controlwise/backend/internal/database"
	"github.com/controlwise/backend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type WhatsAppService struct {
	db            *database.DB
	encryptionKey []byte
}

func NewWhatsAppService(db *database.DB, encryptionKey string) *WhatsAppService {
	// Key should be 32 bytes for AES-256
	key := []byte(encryptionKey)
	if len(key) < 32 {
		// Pad or truncate to 32 bytes
		padded := make([]byte, 32)
		copy(padded, key)
		key = padded
	} else if len(key) > 32 {
		key = key[:32]
	}

	return &WhatsAppService{
		db:            db,
		encryptionKey: key,
	}
}

// GetConfig returns the notification config for an organization
func (s *WhatsAppService) GetConfig(ctx context.Context, orgID uuid.UUID) (*models.NotificationConfig, error) {
	var config models.NotificationConfig
	err := s.db.Pool.QueryRow(ctx, `
		SELECT
			id, organization_id, whatsapp_enabled, twilio_account_sid,
			twilio_auth_token_encrypted, twilio_whatsapp_number,
			reminder_24h_enabled, reminder_2h_enabled,
			reminder_24h_template, reminder_2h_template,
			confirmation_response_template, created_at, updated_at
		FROM notification_configs
		WHERE organization_id = $1
	`, orgID).Scan(
		&config.ID,
		&config.OrganizationID,
		&config.WhatsAppEnabled,
		&config.TwilioAccountSID,
		&config.TwilioAuthTokenEncrypted,
		&config.TwilioWhatsAppNumber,
		&config.Reminder24hEnabled,
		&config.Reminder2hEnabled,
		&config.Reminder24hTemplate,
		&config.Reminder2hTemplate,
		&config.ConfirmationResponseTmpl,
		&config.CreatedAt,
		&config.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get notification config: %w", err)
	}
	return &config, nil
}

// SaveConfig creates or updates notification config
func (s *WhatsAppService) SaveConfig(ctx context.Context, orgID uuid.UUID, config *NotificationConfigInput) error {
	// Encrypt auth token if provided
	var encryptedToken *string
	if config.TwilioAuthToken != nil && *config.TwilioAuthToken != "" {
		encrypted, err := s.encrypt(*config.TwilioAuthToken)
		if err != nil {
			return fmt.Errorf("failed to encrypt auth token: %w", err)
		}
		encryptedToken = &encrypted
	}

	_, err := s.db.Pool.Exec(ctx, `
		INSERT INTO notification_configs (
			organization_id, whatsapp_enabled, twilio_account_sid,
			twilio_auth_token_encrypted, twilio_whatsapp_number,
			reminder_24h_enabled, reminder_2h_enabled,
			reminder_24h_template, reminder_2h_template,
			confirmation_response_template
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (organization_id) DO UPDATE SET
			whatsapp_enabled = EXCLUDED.whatsapp_enabled,
			twilio_account_sid = COALESCE(EXCLUDED.twilio_account_sid, notification_configs.twilio_account_sid),
			twilio_auth_token_encrypted = COALESCE(EXCLUDED.twilio_auth_token_encrypted, notification_configs.twilio_auth_token_encrypted),
			twilio_whatsapp_number = COALESCE(EXCLUDED.twilio_whatsapp_number, notification_configs.twilio_whatsapp_number),
			reminder_24h_enabled = EXCLUDED.reminder_24h_enabled,
			reminder_2h_enabled = EXCLUDED.reminder_2h_enabled,
			reminder_24h_template = COALESCE(EXCLUDED.reminder_24h_template, notification_configs.reminder_24h_template),
			reminder_2h_template = COALESCE(EXCLUDED.reminder_2h_template, notification_configs.reminder_2h_template),
			confirmation_response_template = COALESCE(EXCLUDED.confirmation_response_template, notification_configs.confirmation_response_template),
			updated_at = CURRENT_TIMESTAMP
	`, orgID, config.WhatsAppEnabled, config.TwilioAccountSID, encryptedToken,
		config.TwilioWhatsAppNumber, config.Reminder24hEnabled, config.Reminder2hEnabled,
		config.Reminder24hTemplate, config.Reminder2hTemplate, config.ConfirmationResponseTmpl)

	if err != nil {
		return fmt.Errorf("failed to save notification config: %w", err)
	}
	return nil
}

// SendMessage sends a WhatsApp message using Twilio
func (s *WhatsAppService) SendMessage(ctx context.Context, orgID uuid.UUID, to, message string, sessionID *uuid.UUID) (*models.WhatsAppMessage, error) {
	config, err := s.GetConfig(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, errors.New("notification config not found")
	}
	if !config.WhatsAppEnabled {
		return nil, errors.New("WhatsApp is not enabled")
	}
	if config.TwilioAccountSID == nil || config.TwilioAuthTokenEncrypted == nil {
		return nil, errors.New("Twilio credentials not configured")
	}

	// Decrypt auth token
	authToken, err := s.decrypt(*config.TwilioAuthTokenEncrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt auth token: %w", err)
	}

	// Format phone number for WhatsApp
	whatsappTo := formatWhatsAppNumber(to)
	whatsappFrom := *config.TwilioWhatsAppNumber

	// Create message log entry
	msgLog := &models.WhatsAppMessage{
		ID:             uuid.New(),
		OrganizationID: orgID,
		SessionID:      sessionID,
		Direction:      models.MessageDirectionOutbound,
		PhoneNumber:    to,
		MessageContent: &message,
		Status:         models.MessageStatusQueued,
		CreatedAt:      time.Now(),
	}

	// Save initial message log
	_, err = s.db.Pool.Exec(ctx, `
		INSERT INTO whatsapp_messages (
			id, organization_id, session_id, direction, phone_number,
			message_content, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, msgLog.ID, msgLog.OrganizationID, msgLog.SessionID, msgLog.Direction,
		msgLog.PhoneNumber, msgLog.MessageContent, msgLog.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to log message: %w", err)
	}

	// Send via Twilio API
	messageSID, err := s.sendTwilioMessage(*config.TwilioAccountSID, authToken, whatsappFrom, whatsappTo, message)
	if err != nil {
		// Update log with error
		errMsg := err.Error()
		s.db.Pool.Exec(ctx, `
			UPDATE whatsapp_messages
			SET status = $1, error_message = $2
			WHERE id = $3
		`, models.MessageStatusFailed, errMsg, msgLog.ID)
		msgLog.Status = models.MessageStatusFailed
		msgLog.ErrorMessage = &errMsg
		return msgLog, err
	}

	// Update log with success
	s.db.Pool.Exec(ctx, `
		UPDATE whatsapp_messages
		SET status = $1, message_sid = $2
		WHERE id = $3
	`, models.MessageStatusSent, messageSID, msgLog.ID)
	msgLog.Status = models.MessageStatusSent
	msgLog.MessageSID = &messageSID

	return msgLog, nil
}

// sendTwilioMessage sends a message via Twilio REST API
func (s *WhatsAppService) sendTwilioMessage(accountSID, authToken, from, to, body string) (string, error) {
	twilioURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", accountSID)

	data := url.Values{}
	data.Set("To", to)
	data.Set("From", from)
	data.Set("Body", body)

	req, err := http.NewRequest("POST", twilioURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(accountSID, authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body_bytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		var errorResp struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
		}
		json.Unmarshal(body_bytes, &errorResp)
		return "", fmt.Errorf("twilio error: %s (code: %d)", errorResp.Message, errorResp.Code)
	}

	var successResp struct {
		SID string `json:"sid"`
	}
	if err := json.Unmarshal(body_bytes, &successResp); err != nil {
		return "", err
	}

	return successResp.SID, nil
}

// SendSessionReminder sends a reminder for a session
func (s *WhatsAppService) SendSessionReminder(ctx context.Context, reminder *models.ScheduledReminderWithDetails, orgID uuid.UUID) error {
	config, err := s.GetConfig(ctx, orgID)
	if err != nil {
		return err
	}
	if config == nil {
		return errors.New("notification config not found")
	}

	// Get appropriate template
	var template string
	if reminder.Type == models.ReminderType24h {
		if !config.Reminder24hEnabled {
			return s.skipReminder(ctx, reminder.ID, "24h reminders disabled")
		}
		if config.Reminder24hTemplate != nil {
			template = *config.Reminder24hTemplate
		}
	} else {
		if !config.Reminder2hEnabled {
			return s.skipReminder(ctx, reminder.ID, "2h reminders disabled")
		}
		if config.Reminder2hTemplate != nil {
			template = *config.Reminder2hTemplate
		}
	}

	if template == "" {
		return s.skipReminder(ctx, reminder.ID, "no template configured")
	}

	// Build message from template
	message := s.buildMessage(template, models.MessageTemplateVars{
		PatientName:   reminder.PatientName,
		TherapistName: reminder.TherapistName,
		Date:          reminder.ScheduledAt.Format("02/01/2006"),
		Time:          reminder.ScheduledAt.Format("15:04"),
	})

	// Send message
	msgLog, err := s.SendMessage(ctx, orgID, reminder.PatientPhone, message, &reminder.SessionID)

	// Update reminder status
	if err != nil {
		errMsg := err.Error()
		s.db.Pool.Exec(ctx, `
			UPDATE scheduled_reminders
			SET status = 'failed', processed_at = NOW(), error_message = $1
			WHERE id = $2
		`, errMsg, reminder.ID)
		return err
	}

	s.db.Pool.Exec(ctx, `
		UPDATE scheduled_reminders
		SET status = 'sent', processed_at = NOW(), whatsapp_message_id = $1
		WHERE id = $2
	`, msgLog.ID, reminder.ID)

	return nil
}

// GetPendingReminders returns reminders that need to be sent
func (s *WhatsAppService) GetPendingReminders(ctx context.Context) ([]*models.ScheduledReminderWithDetails, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT
			sr.id, sr.session_id, sr.type, sr.scheduled_for, sr.status,
			sr.processed_at, sr.error_message, sr.whatsapp_message_id, sr.created_at,
			p.name as patient_name, p.phone as patient_phone,
			t.name as therapist_name, sess.scheduled_at
		FROM scheduled_reminders sr
		JOIN sessions sess ON sess.id = sr.session_id
		JOIN patients p ON p.id = sess.patient_id
		JOIN therapists t ON t.id = sess.therapist_id
		WHERE sr.status = 'pending'
			AND sr.scheduled_for <= NOW()
			AND sess.status IN ('pending', 'confirmed')
			AND sess.deleted_at IS NULL
		ORDER BY sr.scheduled_for ASC
		LIMIT 100
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending reminders: %w", err)
	}
	defer rows.Close()

	var reminders []*models.ScheduledReminderWithDetails
	for rows.Next() {
		var r models.ScheduledReminderWithDetails
		err := rows.Scan(
			&r.ID,
			&r.SessionID,
			&r.Type,
			&r.ScheduledFor,
			&r.Status,
			&r.ProcessedAt,
			&r.ErrorMessage,
			&r.WhatsAppMessageID,
			&r.CreatedAt,
			&r.PatientName,
			&r.PatientPhone,
			&r.TherapistName,
			&r.ScheduledAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reminder: %w", err)
		}
		reminders = append(reminders, &r)
	}

	return reminders, nil
}

// ProcessIncomingMessage handles incoming WhatsApp messages (webhook)
func (s *WhatsAppService) ProcessIncomingMessage(ctx context.Context, orgID uuid.UUID, from, body, messageSID string) error {
	// Log the incoming message
	content := body
	_, err := s.db.Pool.Exec(ctx, `
		INSERT INTO whatsapp_messages (
			id, organization_id, direction, phone_number, message_content, message_sid, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, uuid.New(), orgID, models.MessageDirectionInbound, from, content, messageSID, models.MessageStatusDelivered)
	if err != nil {
		return fmt.Errorf("failed to log incoming message: %w", err)
	}

	// Parse response (SIM/NAO, YES/NO, etc.)
	response := parseConfirmationResponse(body)

	// Find pending session for this phone number
	var sessionID uuid.UUID
	var currentStatus models.SessionStatus
	err = s.db.Pool.QueryRow(ctx, `
		SELECT s.id, s.status
		FROM sessions s
		JOIN patients p ON p.id = s.patient_id
		WHERE s.organization_id = $1
			AND p.phone = $2
			AND s.status IN ('pending', 'confirmed')
			AND s.scheduled_at > NOW()
			AND s.deleted_at IS NULL
		ORDER BY s.scheduled_at ASC
		LIMIT 1
	`, orgID, normalizePhone(from)).Scan(&sessionID, &currentStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// No pending session found - that's okay
			return nil
		}
		return fmt.Errorf("failed to find session: %w", err)
	}

	// Update session based on response
	if response == "confirmed" && currentStatus == models.SessionStatusPending {
		s.db.Pool.Exec(ctx, `
			UPDATE sessions SET status = 'confirmed' WHERE id = $1
		`, sessionID)

		// Update session confirmation record
		s.db.Pool.Exec(ctx, `
			UPDATE session_confirmations
			SET responded_at = NOW(), response = 'confirmed'
			WHERE session_id = $1 AND response IS NULL
		`, sessionID)
	} else if response == "cancelled" {
		s.db.Pool.Exec(ctx, `
			UPDATE sessions SET status = 'cancelled', cancel_reason = 'Cancelled via WhatsApp', cancelled_at = NOW()
			WHERE id = $1
		`, sessionID)

		s.db.Pool.Exec(ctx, `
			UPDATE session_confirmations
			SET responded_at = NOW(), response = 'cancelled'
			WHERE session_id = $1 AND response IS NULL
		`, sessionID)
	}

	return nil
}

// UpdateMessageStatus updates message status from Twilio webhook
func (s *WhatsAppService) UpdateMessageStatus(ctx context.Context, messageSID, status string) error {
	twilioStatus := mapTwilioStatus(status)
	_, err := s.db.Pool.Exec(ctx, `
		UPDATE whatsapp_messages
		SET status = $1
		WHERE message_sid = $2
	`, twilioStatus, messageSID)
	return err
}

// skipReminder marks a reminder as skipped
func (s *WhatsAppService) skipReminder(ctx context.Context, reminderID uuid.UUID, reason string) error {
	_, err := s.db.Pool.Exec(ctx, `
		UPDATE scheduled_reminders
		SET status = 'skipped', processed_at = NOW(), error_message = $1
		WHERE id = $2
	`, reason, reminderID)
	return err
}

// buildMessage replaces template variables with values
func (s *WhatsAppService) buildMessage(template string, vars models.MessageTemplateVars) string {
	result := template
	result = strings.ReplaceAll(result, "{{patient_name}}", vars.PatientName)
	result = strings.ReplaceAll(result, "{{therapist}}", vars.TherapistName)
	result = strings.ReplaceAll(result, "{{date}}", vars.Date)
	result = strings.ReplaceAll(result, "{{time}}", vars.Time)
	result = strings.ReplaceAll(result, "{{status}}", vars.Status)
	return result
}

// encrypt encrypts a string using AES-256-GCM
func (s *WhatsAppService) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt decrypts a string using AES-256-GCM
func (s *WhatsAppService) decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// Helper functions

func formatWhatsAppNumber(phone string) string {
	// Remove any existing whatsapp: prefix
	phone = strings.TrimPrefix(phone, "whatsapp:")
	// Remove spaces and dashes
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	// Add + if not present
	if !strings.HasPrefix(phone, "+") {
		// Assume Portugal country code if no prefix
		if !strings.HasPrefix(phone, "351") {
			phone = "351" + phone
		}
		phone = "+" + phone
	}
	return "whatsapp:" + phone
}

func normalizePhone(phone string) string {
	phone = strings.TrimPrefix(phone, "whatsapp:")
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	return phone
}

func parseConfirmationResponse(message string) string {
	message = strings.ToLower(strings.TrimSpace(message))

	// Positive responses
	positiveResponses := []string{"sim", "yes", "s", "y", "1", "confirmo", "confirmado", "ok"}
	for _, resp := range positiveResponses {
		if message == resp || strings.HasPrefix(message, resp+" ") {
			return "confirmed"
		}
	}

	// Negative responses
	negativeResponses := []string{"nao", "não", "no", "n", "0", "cancelar", "cancelo", "cancelado"}
	for _, resp := range negativeResponses {
		if message == resp || strings.HasPrefix(message, resp+" ") {
			return "cancelled"
		}
	}

	return "unknown"
}

func mapTwilioStatus(status string) models.WhatsAppMessageStatus {
	switch strings.ToLower(status) {
	case "queued":
		return models.MessageStatusQueued
	case "sending":
		return models.MessageStatusSending
	case "sent":
		return models.MessageStatusSent
	case "delivered":
		return models.MessageStatusDelivered
	case "read":
		return models.MessageStatusRead
	case "failed":
		return models.MessageStatusFailed
	case "undelivered":
		return models.MessageStatusUndelivered
	default:
		return models.MessageStatusSent
	}
}

// NotificationConfigInput represents input for saving config
type NotificationConfigInput struct {
	WhatsAppEnabled          bool    `json:"whatsapp_enabled"`
	TwilioAccountSID         *string `json:"twilio_account_sid"`
	TwilioAuthToken          *string `json:"twilio_auth_token"`
	TwilioWhatsAppNumber     *string `json:"twilio_whatsapp_number"`
	Reminder24hEnabled       bool    `json:"reminder_24h_enabled"`
	Reminder2hEnabled        bool    `json:"reminder_2h_enabled"`
	Reminder24hTemplate      *string `json:"reminder_24h_template"`
	Reminder2hTemplate       *string `json:"reminder_2h_template"`
	ConfirmationResponseTmpl *string `json:"confirmation_response_template"`
}
