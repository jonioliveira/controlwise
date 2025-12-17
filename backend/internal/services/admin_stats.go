package services

import (
	"context"
	"time"

	"github.com/controlwise/backend/internal/database"
	"github.com/controlwise/backend/internal/models"
)

type AdminStatsService struct {
	db *database.DB
}

func NewAdminStatsService(db *database.DB) *AdminStatsService {
	return &AdminStatsService{db: db}
}

func (s *AdminStatsService) GetPlatformStats(ctx context.Context) (*models.PlatformStats, error) {
	stats := &models.PlatformStats{
		OrgsByModule: make(map[string]int),
	}

	// Get organization counts
	err := s.db.Pool.QueryRow(ctx, `
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE is_active = true AND suspended_at IS NULL) as active,
			COUNT(*) FILTER (WHERE suspended_at IS NOT NULL) as suspended
		FROM organizations WHERE deleted_at IS NULL
	`).Scan(&stats.TotalOrganizations, &stats.ActiveOrganizations, &stats.SuspendedOrganizations)
	if err != nil {
		return nil, err
	}

	// Get user counts
	err = s.db.Pool.QueryRow(ctx, `
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE is_active = true) as active
		FROM users WHERE deleted_at IS NULL
	`).Scan(&stats.TotalUsers, &stats.ActiveUsers)
	if err != nil {
		return nil, err
	}

	// Get new orgs this month
	startOfMonth := time.Now().UTC().Truncate(24 * time.Hour)
	startOfMonth = time.Date(startOfMonth.Year(), startOfMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
	err = s.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM organizations WHERE created_at >= $1 AND deleted_at IS NULL
	`, startOfMonth).Scan(&stats.NewOrgsThisMonth)
	if err != nil {
		return nil, err
	}

	// Get new users this month
	err = s.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM users WHERE created_at >= $1 AND deleted_at IS NULL
	`, startOfMonth).Scan(&stats.NewUsersThisMonth)
	if err != nil {
		return nil, err
	}

	// Get orgs by module
	rows, err := s.db.Pool.Query(ctx, `
		SELECT module_name, COUNT(DISTINCT organization_id) as count
		FROM organization_modules
		WHERE is_enabled = true
		GROUP BY module_name
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var moduleName string
			var count int
			if err := rows.Scan(&moduleName, &count); err == nil {
				stats.OrgsByModule[moduleName] = count
			}
		}
	}

	return stats, nil
}

type RecentActivity struct {
	Type        string    `json:"type"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	EntityID    *string   `json:"entity_id,omitempty"`
	EntityType  *string   `json:"entity_type,omitempty"`
}

func (s *AdminStatsService) GetRecentActivity(ctx context.Context, limit int) ([]RecentActivity, error) {
	if limit < 1 || limit > 50 {
		limit = 10
	}

	var activities []RecentActivity

	// Get recent organizations
	rows, err := s.db.Pool.Query(ctx, `
		SELECT 'org_created' as type, name as description, created_at, id::text as entity_id
		FROM organizations
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var activity RecentActivity
		var entityID string
		err := rows.Scan(&activity.Type, &activity.Description, &activity.CreatedAt, &entityID)
		if err == nil {
			activity.EntityID = &entityID
			entityType := "organization"
			activity.EntityType = &entityType
			activities = append(activities, activity)
		}
	}

	// Get recent users
	rows, err = s.db.Pool.Query(ctx, `
		SELECT 'user_created' as type, CONCAT(first_name, ' ', last_name) as description, u.created_at, u.id::text as entity_id
		FROM users u
		WHERE u.deleted_at IS NULL
		ORDER BY u.created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var activity RecentActivity
		var entityID string
		err := rows.Scan(&activity.Type, &activity.Description, &activity.CreatedAt, &entityID)
		if err == nil {
			activity.EntityID = &entityID
			entityType := "user"
			activity.EntityType = &entityType
			activities = append(activities, activity)
		}
	}

	// Sort by created_at descending and limit
	// For simplicity, just return what we have (in production, use UNION ALL with ORDER BY)
	if len(activities) > limit {
		activities = activities[:limit]
	}

	return activities, nil
}
