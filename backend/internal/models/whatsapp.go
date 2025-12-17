package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// NotificationConfig represents WhatsApp notification settings for an organization
type NotificationConfig struct {
	ID                         uuid.UUID `json:"id" db:"id"`
	OrganizationID             uuid.UUID `json:"organization_id" db:"organization_id"`
	WhatsAppEnabled            bool      `json:"whatsapp_enabled" db:"whatsapp_enabled"`
	TwilioAccountSID           *string   `json:"-" db:"twilio_account_sid"`
	TwilioAuthTokenEncrypted   *string   `json:"-" db:"twilio_auth_token_encrypted"`
	TwilioWhatsAppNumber       *string   `json:"twilio_whatsapp_number" db:"twilio_whatsapp_number"`
	Reminder24hEnabled         bool      `json:"reminder_24h_enabled" db:"reminder_24h_enabled"`
	Reminder2hEnabled          bool      `json:"reminder_2h_enabled" db:"reminder_2h_enabled"`
	Reminder24hTemplate        *string   `json:"reminder_24h_template" db:"reminder_24h_template"`
	Reminder2hTemplate         *string   `json:"reminder_2h_template" db:"reminder_2h_template"`
	ConfirmationResponseTmpl   *string   `json:"confirmation_response_template" db:"confirmation_response_template"`
	CreatedAt                  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt                  time.Time `json:"updated_at" db:"updated_at"`
}

// NotificationConfigPublic is the public-facing version without sensitive data
type NotificationConfigPublic struct {
	ID                       uuid.UUID `json:"id"`
	OrganizationID           uuid.UUID `json:"organization_id"`
	WhatsAppEnabled          bool      `json:"whatsapp_enabled"`
	TwilioConfigured         bool      `json:"twilio_configured"`
	TwilioWhatsAppNumber     *string   `json:"twilio_whatsapp_number"`
	Reminder24hEnabled       bool      `json:"reminder_24h_enabled"`
	Reminder2hEnabled        bool      `json:"reminder_2h_enabled"`
	Reminder24hTemplate      *string   `json:"reminder_24h_template"`
	Reminder2hTemplate       *string   `json:"reminder_2h_template"`
	ConfirmationResponseTmpl *string   `json:"confirmation_response_template"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}

// ToPublic converts NotificationConfig to public version
func (c *NotificationConfig) ToPublic() NotificationConfigPublic {
	return NotificationConfigPublic{
		ID:                       c.ID,
		OrganizationID:           c.OrganizationID,
		WhatsAppEnabled:          c.WhatsAppEnabled,
		TwilioConfigured:         c.TwilioAccountSID != nil && c.TwilioAuthTokenEncrypted != nil,
		TwilioWhatsAppNumber:     c.TwilioWhatsAppNumber,
		Reminder24hEnabled:       c.Reminder24hEnabled,
		Reminder2hEnabled:        c.Reminder2hEnabled,
		Reminder24hTemplate:      c.Reminder24hTemplate,
		Reminder2hTemplate:       c.Reminder2hTemplate,
		ConfirmationResponseTmpl: c.ConfirmationResponseTmpl,
		CreatedAt:                c.CreatedAt,
		UpdatedAt:                c.UpdatedAt,
	}
}

// WhatsAppMessageDirection represents message direction
type WhatsAppMessageDirection string

const (
	MessageDirectionOutbound WhatsAppMessageDirection = "outbound"
	MessageDirectionInbound  WhatsAppMessageDirection = "inbound"
)

// WhatsAppMessageStatus represents message delivery status
type WhatsAppMessageStatus string

const (
	MessageStatusQueued      WhatsAppMessageStatus = "queued"
	MessageStatusSending     WhatsAppMessageStatus = "sending"
	MessageStatusSent        WhatsAppMessageStatus = "sent"
	MessageStatusDelivered   WhatsAppMessageStatus = "delivered"
	MessageStatusRead        WhatsAppMessageStatus = "read"
	MessageStatusFailed      WhatsAppMessageStatus = "failed"
	MessageStatusUndelivered WhatsAppMessageStatus = "undelivered"
)

// WhatsAppMessage represents a WhatsApp message log entry
type WhatsAppMessage struct {
	ID             uuid.UUID               `json:"id" db:"id"`
	OrganizationID uuid.UUID               `json:"organization_id" db:"organization_id"`
	SessionID      *uuid.UUID              `json:"session_id" db:"session_id"`
	Direction      WhatsAppMessageDirection `json:"direction" db:"direction"`
	PhoneNumber    string                  `json:"phone_number" db:"phone_number"`
	MessageContent *string                 `json:"message_content" db:"message_content"`
	MessageSID     *string                 `json:"message_sid" db:"message_sid"`
	Status         WhatsAppMessageStatus   `json:"status" db:"status"`
	ErrorCode      *string                 `json:"error_code" db:"error_code"`
	ErrorMessage   *string                 `json:"error_message" db:"error_message"`
	RawPayload     json.RawMessage         `json:"raw_payload" db:"raw_payload"`
	CreatedAt      time.Time               `json:"created_at" db:"created_at"`
}

// ReminderType represents the type of scheduled reminder
type ReminderType string

const (
	ReminderType24h ReminderType = "reminder_24h"
	ReminderType2h  ReminderType = "reminder_2h"
)

// ReminderStatus represents the status of a scheduled reminder
type ReminderStatus string

const (
	ReminderStatusPending ReminderStatus = "pending"
	ReminderStatusSent    ReminderStatus = "sent"
	ReminderStatusFailed  ReminderStatus = "failed"
	ReminderStatusSkipped ReminderStatus = "skipped"
)

// ScheduledReminder represents a scheduled notification reminder
type ScheduledReminder struct {
	ID                uuid.UUID      `json:"id" db:"id"`
	SessionID         uuid.UUID      `json:"session_id" db:"session_id"`
	Type              ReminderType   `json:"type" db:"type"`
	ScheduledFor      time.Time      `json:"scheduled_for" db:"scheduled_for"`
	Status            ReminderStatus `json:"status" db:"status"`
	ProcessedAt       *time.Time     `json:"processed_at" db:"processed_at"`
	ErrorMessage      *string        `json:"error_message" db:"error_message"`
	WhatsAppMessageID *uuid.UUID     `json:"whatsapp_message_id" db:"whatsapp_message_id"`
	CreatedAt         time.Time      `json:"created_at" db:"created_at"`
}

// ScheduledReminderWithDetails includes session details for processing
type ScheduledReminderWithDetails struct {
	ScheduledReminder
	PatientName   string    `json:"patient_name"`
	PatientPhone  string    `json:"patient_phone"`
	TherapistName string    `json:"therapist_name"`
	ScheduledAt   time.Time `json:"scheduled_at"`
}

// MessageTemplateVars represents variables for message templates
type MessageTemplateVars struct {
	PatientName   string
	TherapistName string
	Date          string // Formatted date
	Time          string // Formatted time
	Status        string // For confirmation responses
}
