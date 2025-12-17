package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Patient represents a patient in the appointments module
// Patient is a healthcare extension of Client - every patient must be linked to a client
type Patient struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	OrganizationID   uuid.UUID  `json:"organization_id" db:"organization_id"`
	ClientID         uuid.UUID  `json:"client_id" db:"client_id"` // Required FK to clients table
	DateOfBirth      *time.Time `json:"date_of_birth" db:"date_of_birth"`
	Notes            *string    `json:"notes" db:"notes"` // Medical notes
	EmergencyContact *string    `json:"emergency_contact" db:"emergency_contact"`
	EmergencyPhone   *string    `json:"emergency_phone" db:"emergency_phone"`
	IsActive         bool       `json:"is_active" db:"is_active"`
	CreatedBy        *uuid.UUID `json:"created_by" db:"created_by"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
	// Joined fields from client
	Client *Client `json:"client,omitempty" db:"-"`
}

// PatientWithClient includes the linked client data
type PatientWithClient struct {
	Patient
	ClientName  string  `json:"client_name" db:"client_name"`
	ClientEmail string  `json:"client_email" db:"client_email"`
	ClientPhone string  `json:"client_phone" db:"client_phone"`
}

// Therapist represents a therapist/service provider in the appointments module
type Therapist struct {
	ID                     uuid.UUID       `json:"id" db:"id"`
	OrganizationID         uuid.UUID       `json:"organization_id" db:"organization_id"`
	UserID                 *uuid.UUID      `json:"user_id" db:"user_id"`
	Name                   string          `json:"name" db:"name"`
	Email                  *string         `json:"email" db:"email"`
	Phone                  *string         `json:"phone" db:"phone"`
	Specialty              *string         `json:"specialty" db:"specialty"`
	WorkingHours           json.RawMessage `json:"working_hours" db:"working_hours"`
	SessionDurationMinutes int             `json:"session_duration_minutes" db:"session_duration_minutes"`
	DefaultPriceCents      int             `json:"default_price_cents" db:"default_price_cents"`
	Timezone               string          `json:"timezone" db:"timezone"`
	IsActive               bool            `json:"is_active" db:"is_active"`
	CreatedAt              time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time       `json:"updated_at" db:"updated_at"`
	DeletedAt              *time.Time      `json:"deleted_at,omitempty" db:"deleted_at"`
}

// WorkingHoursDay represents working hours for a single day
type WorkingHoursDay struct {
	Start string `json:"start"` // "09:00"
	End   string `json:"end"`   // "18:00"
}

// WorkingHours represents the working hours configuration
type WorkingHours map[string]WorkingHoursDay // key: monday, tuesday, etc.

// GetWorkingHours parses the working hours JSON
func (t *Therapist) GetWorkingHours() (WorkingHours, error) {
	if t.WorkingHours == nil {
		return WorkingHours{}, nil
	}
	var hours WorkingHours
	if err := json.Unmarshal(t.WorkingHours, &hours); err != nil {
		return nil, err
	}
	return hours, nil
}

// SetWorkingHours serializes working hours to JSON
func (t *Therapist) SetWorkingHours(hours WorkingHours) error {
	data, err := json.Marshal(hours)
	if err != nil {
		return err
	}
	t.WorkingHours = data
	return nil
}

// SessionStatus represents the status of a session
type SessionStatus string

const (
	SessionStatusPending   SessionStatus = "pending"
	SessionStatusConfirmed SessionStatus = "confirmed"
	SessionStatusCancelled SessionStatus = "cancelled"
	SessionStatusCompleted SessionStatus = "completed"
	SessionStatusNoShow    SessionStatus = "no_show"
)

// SessionType represents the type of session
type SessionType string

const (
	SessionTypeRegular    SessionType = "regular"
	SessionTypeEvaluation SessionType = "evaluation"
	SessionTypeFollowUp   SessionType = "follow_up"
)

// Session represents an appointment/session
type Session struct {
	ID              uuid.UUID     `json:"id" db:"id"`
	OrganizationID  uuid.UUID     `json:"organization_id" db:"organization_id"`
	TherapistID     uuid.UUID     `json:"therapist_id" db:"therapist_id"`
	PatientID       uuid.UUID     `json:"patient_id" db:"patient_id"`
	ScheduledAt     time.Time     `json:"scheduled_at" db:"scheduled_at"`
	DurationMinutes int           `json:"duration_minutes" db:"duration_minutes"`
	PriceCents      int           `json:"price_cents" db:"price_cents"`
	Status          SessionStatus `json:"status" db:"status"`
	SessionType     SessionType   `json:"session_type" db:"session_type"`
	Notes           *string       `json:"notes" db:"notes"`
	CancelReason    *string       `json:"cancel_reason" db:"cancel_reason"`
	CancelledAt     *time.Time    `json:"cancelled_at" db:"cancelled_at"`
	CancelledBy     *uuid.UUID    `json:"cancelled_by" db:"cancelled_by"`
	CompletedAt     *time.Time    `json:"completed_at" db:"completed_at"`
	CreatedBy       *uuid.UUID    `json:"created_by" db:"created_by"`
	CreatedAt       time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at" db:"updated_at"`
	DeletedAt       *time.Time    `json:"deleted_at,omitempty" db:"deleted_at"`
}

// SessionWithDetails includes therapist and patient information
type SessionWithDetails struct {
	Session
	TherapistName string  `json:"therapist_name"`
	PatientName   string  `json:"patient_name"`
	PatientPhone  string  `json:"patient_phone"`
	PatientEmail  *string `json:"patient_email"`
}

// EndTime calculates the end time of a session
func (s *Session) EndTime() time.Time {
	return s.ScheduledAt.Add(time.Duration(s.DurationMinutes) * time.Minute)
}

// ConfirmationType represents the type of confirmation message
type ConfirmationType string

const (
	ConfirmationTypeReminder24h ConfirmationType = "reminder_24h"
	ConfirmationTypeReminder2h  ConfirmationType = "reminder_2h"
	ConfirmationTypeManual      ConfirmationType = "manual"
	ConfirmationTypeFollowup    ConfirmationType = "followup"
)

// ConfirmationChannel represents the channel used for confirmation
type ConfirmationChannel string

const (
	ConfirmationChannelWhatsApp ConfirmationChannel = "whatsapp"
	ConfirmationChannelSMS      ConfirmationChannel = "sms"
	ConfirmationChannelEmail    ConfirmationChannel = "email"
)

// ConfirmationResponse represents the patient's response
type ConfirmationResponse string

const (
	ConfirmationResponseConfirmed   ConfirmationResponse = "confirmed"
	ConfirmationResponseCancelled   ConfirmationResponse = "cancelled"
	ConfirmationResponseRescheduled ConfirmationResponse = "rescheduled"
	ConfirmationResponseNoResponse  ConfirmationResponse = "no_response"
)

// SessionConfirmation tracks confirmation messages for a session
type SessionConfirmation struct {
	ID           uuid.UUID             `json:"id" db:"id"`
	SessionID    uuid.UUID             `json:"session_id" db:"session_id"`
	Type         ConfirmationType      `json:"type" db:"type"`
	Channel      ConfirmationChannel   `json:"channel" db:"channel"`
	SentAt       *time.Time            `json:"sent_at" db:"sent_at"`
	DeliveredAt  *time.Time            `json:"delivered_at" db:"delivered_at"`
	ReadAt       *time.Time            `json:"read_at" db:"read_at"`
	RespondedAt  *time.Time            `json:"responded_at" db:"responded_at"`
	Response     *ConfirmationResponse `json:"response" db:"response"`
	MessageID    *string               `json:"message_id" db:"message_id"`
	ErrorMessage *string               `json:"error_message" db:"error_message"`
	RawResponse  json.RawMessage       `json:"raw_response" db:"raw_response"`
	CreatedAt    time.Time             `json:"created_at" db:"created_at"`
}

// SessionHistory tracks changes to sessions
type SessionHistory struct {
	ID        uuid.UUID       `json:"id" db:"id"`
	SessionID uuid.UUID       `json:"session_id" db:"session_id"`
	Action    string          `json:"action" db:"action"`
	OldValues json.RawMessage `json:"old_values" db:"old_values"`
	NewValues json.RawMessage `json:"new_values" db:"new_values"`
	ChangedBy *uuid.UUID      `json:"changed_by" db:"changed_by"`
	ChangedAt time.Time       `json:"changed_at" db:"changed_at"`
}

// CalendarEvent represents a session formatted for calendar display
type CalendarEvent struct {
	ID            uuid.UUID     `json:"id"`
	Title         string        `json:"title"`
	Start         time.Time     `json:"start"`
	End           time.Time     `json:"end"`
	Status        SessionStatus `json:"status"`
	TherapistID   uuid.UUID     `json:"therapist_id"`
	TherapistName string        `json:"therapist_name"`
	PatientID     uuid.UUID     `json:"patient_id"`
	PatientName   string        `json:"patient_name"`
	Color         string        `json:"color,omitempty"` // For calendar display
}

// ToCalendarEvent converts a SessionWithDetails to a CalendarEvent
func (s *SessionWithDetails) ToCalendarEvent() CalendarEvent {
	color := "#3B82F6" // Default blue
	switch s.Status {
	case SessionStatusConfirmed:
		color = "#22C55E" // Green
	case SessionStatusCancelled:
		color = "#EF4444" // Red
	case SessionStatusCompleted:
		color = "#6B7280" // Gray
	case SessionStatusNoShow:
		color = "#F59E0B" // Orange
	}

	return CalendarEvent{
		ID:            s.ID,
		Title:         s.PatientName,
		Start:         s.ScheduledAt,
		End:           s.EndTime(),
		Status:        s.Status,
		TherapistID:   s.TherapistID,
		TherapistName: s.TherapistName,
		PatientID:     s.PatientID,
		PatientName:   s.PatientName,
		Color:         color,
	}
}
