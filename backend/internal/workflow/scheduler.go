package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/controlwise/backend/internal/database"
	"github.com/controlwise/backend/internal/models"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// Job type constants (matching jobs package)
const (
	TypeExecuteTrigger = "workflow:execute_trigger"
)

// ExecuteTriggerPayload matches jobs.ExecuteTriggerPayload
type ExecuteTriggerPayload struct {
	OrganizationID uuid.UUID `json:"organization_id"`
	TriggerID      uuid.UUID `json:"trigger_id"`
	EntityType     string    `json:"entity_type"`
	EntityID       uuid.UUID `json:"entity_id"`
}

// Scheduler handles scheduling of workflow jobs
type Scheduler struct {
	db     *database.DB
	client *asynq.Client
}

// NewScheduler creates a new scheduler
func NewScheduler(db *database.DB, client *asynq.Client) *Scheduler {
	return &Scheduler{
		db:     db,
		client: client,
	}
}

// ScheduleTimeTrigger schedules a time-based trigger for later execution
func (s *Scheduler) ScheduleTimeTrigger(ctx context.Context, orgID uuid.UUID, trigger *models.WorkflowTrigger, entityType string, entityID uuid.UUID, entityData map[string]interface{}) error {
	if trigger.TimeOffsetMinutes == nil {
		return fmt.Errorf("time trigger requires time_offset_minutes")
	}

	// Calculate scheduled time based on time_field
	var baseTime time.Time
	timeField := "created_at"
	if trigger.TimeField != nil {
		timeField = *trigger.TimeField
	}

	// Get base time from entity data
	if t, ok := entityData[timeField]; ok {
		switch v := t.(type) {
		case time.Time:
			baseTime = v
		case string:
			parsed, err := time.Parse(time.RFC3339, v)
			if err == nil {
				baseTime = parsed
			}
		}
	}

	if baseTime.IsZero() {
		// Try to get from database based on entity type
		baseTime = time.Now()
		if entityType == "session" && timeField == "scheduled_at" {
			var scheduledAt time.Time
			err := s.db.Pool.QueryRow(ctx, `SELECT scheduled_at FROM sessions WHERE id = $1`, entityID).Scan(&scheduledAt)
			if err == nil {
				baseTime = scheduledAt
			}
		}
	}

	// Calculate scheduled time
	offset := time.Duration(*trigger.TimeOffsetMinutes) * time.Minute
	var scheduledFor time.Time
	if trigger.TriggerType == models.TriggerTypeTimeBefore {
		scheduledFor = baseTime.Add(-offset)
	} else {
		scheduledFor = baseTime.Add(offset)
	}

	// Don't schedule in the past
	if scheduledFor.Before(time.Now()) {
		log.Printf("[Scheduler] Skipping trigger %s - scheduled time %v is in the past", trigger.ID, scheduledFor)
		return nil
	}

	// Create scheduled job record
	jobID := uuid.New()
	_, err := s.db.Pool.Exec(ctx, `
		INSERT INTO scheduled_jobs (id, organization_id, trigger_id, entity_type, entity_id, scheduled_for, status)
		VALUES ($1, $2, $3, $4, $5, $6, 'pending')
	`, jobID, orgID, trigger.ID, entityType, entityID, scheduledFor)
	if err != nil {
		return fmt.Errorf("failed to create scheduled job: %w", err)
	}

	log.Printf("[Scheduler] Scheduled trigger %s for %v (entity %s/%s)", trigger.ID, scheduledFor, entityType, entityID)
	return nil
}

// ScheduleRecurringTrigger sets up a recurring trigger (placeholder - actual cron handling is done by Asynq scheduler)
func (s *Scheduler) ScheduleRecurringTrigger(ctx context.Context, orgID uuid.UUID, trigger *models.WorkflowTrigger, entityType string, entityID uuid.UUID) error {
	if trigger.RecurringCron == nil {
		return fmt.Errorf("recurring trigger requires cron expression")
	}

	log.Printf("[Scheduler] Recurring trigger %s registered with cron: %s", trigger.ID, *trigger.RecurringCron)
	// Recurring triggers are handled by the CheckTimeTriggers job which runs every minute
	// and checks for any due recurring triggers
	return nil
}

// CancelPendingJobs cancels all pending scheduled jobs for an entity
func (s *Scheduler) CancelPendingJobs(ctx context.Context, entityType string, entityID uuid.UUID) error {
	result, err := s.db.Pool.Exec(ctx, `
		UPDATE scheduled_jobs
		SET status = 'cancelled'
		WHERE entity_type = $1 AND entity_id = $2 AND status = 'pending'
	`, entityType, entityID)
	if err != nil {
		return fmt.Errorf("failed to cancel pending jobs: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("[Scheduler] Cancelled %d pending jobs for entity %s/%s", rowsAffected, entityType, entityID)
	}
	return nil
}

// ProcessPendingJobs processes all pending scheduled jobs that are due
// This is called by the CheckTimeTriggers periodic job
func (s *Scheduler) ProcessPendingJobs(ctx context.Context) error {
	// Find all pending jobs that are due
	rows, err := s.db.Pool.Query(ctx, `
		SELECT id, organization_id, trigger_id, entity_type, entity_id
		FROM scheduled_jobs
		WHERE status = 'pending' AND scheduled_for <= NOW()
		ORDER BY scheduled_for ASC
		LIMIT 100
	`)
	if err != nil {
		return fmt.Errorf("failed to query pending jobs: %w", err)
	}
	defer rows.Close()

	var jobs []struct {
		ID             uuid.UUID
		OrganizationID uuid.UUID
		TriggerID      uuid.UUID
		EntityType     string
		EntityID       uuid.UUID
	}

	for rows.Next() {
		var job struct {
			ID             uuid.UUID
			OrganizationID uuid.UUID
			TriggerID      uuid.UUID
			EntityType     string
			EntityID       uuid.UUID
		}
		if err := rows.Scan(&job.ID, &job.OrganizationID, &job.TriggerID, &job.EntityType, &job.EntityID); err != nil {
			return fmt.Errorf("failed to scan job: %w", err)
		}
		jobs = append(jobs, job)
	}

	if len(jobs) == 0 {
		return nil
	}

	log.Printf("[Scheduler] Processing %d pending jobs", len(jobs))

	for _, job := range jobs {
		// Mark as processing
		_, err := s.db.Pool.Exec(ctx, `
			UPDATE scheduled_jobs SET status = 'processing', attempts = attempts + 1
			WHERE id = $1
		`, job.ID)
		if err != nil {
			log.Printf("[Scheduler] Failed to mark job %s as processing: %v", job.ID, err)
			continue
		}

		// Enqueue the trigger execution task
		payload := ExecuteTriggerPayload{
			OrganizationID: job.OrganizationID,
			TriggerID:      job.TriggerID,
			EntityType:     job.EntityType,
			EntityID:       job.EntityID,
		}
		data, _ := json.Marshal(payload)
		task := asynq.NewTask(TypeExecuteTrigger, data)

		if s.client != nil {
			_, err = s.client.Enqueue(task, asynq.Queue("default"))
			if err != nil {
				log.Printf("[Scheduler] Failed to enqueue job %s: %v", job.ID, err)
				// Mark as failed
				s.db.Pool.Exec(ctx, `
					UPDATE scheduled_jobs SET status = 'failed', last_error = $1
					WHERE id = $2
				`, err.Error(), job.ID)
				continue
			}
		}

		// Mark as completed
		_, err = s.db.Pool.Exec(ctx, `
			UPDATE scheduled_jobs SET status = 'completed', processed_at = NOW()
			WHERE id = $1
		`, job.ID)
		if err != nil {
			log.Printf("[Scheduler] Failed to mark job %s as completed: %v", job.ID, err)
		}
	}

	return nil
}

// GetPendingJobCount returns the count of pending jobs
func (s *Scheduler) GetPendingJobCount(ctx context.Context) (int, error) {
	var count int
	err := s.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM scheduled_jobs WHERE status = 'pending'
	`).Scan(&count)
	return count, err
}

// CleanupOldJobs removes completed/cancelled jobs older than 30 days
func (s *Scheduler) CleanupOldJobs(ctx context.Context) error {
	result, err := s.db.Pool.Exec(ctx, `
		DELETE FROM scheduled_jobs
		WHERE status IN ('completed', 'cancelled', 'failed')
		AND created_at < NOW() - INTERVAL '30 days'
	`)
	if err != nil {
		return fmt.Errorf("failed to cleanup old jobs: %w", err)
	}

	if result.RowsAffected() > 0 {
		log.Printf("[Scheduler] Cleaned up %d old jobs", result.RowsAffected())
	}
	return nil
}
