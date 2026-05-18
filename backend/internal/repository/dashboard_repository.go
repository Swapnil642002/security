package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditLogItem struct {
	ID        int64     `json:"id"`
	Action    string    `json:"action"`
	Entity    string    `json:"entity"`
	CreatedAt time.Time `json:"created_at"`
}

type DashboardSummaryRow struct {
	Provider             string
	LastSyncAt           *time.Time
	WebsitePolicyCount   int
	PortPolicyCount      int
	TotalPolicyCount     int
	DepartmentGroupCount int
	PendingChanges       int
}

type DashboardRepository struct {
	pool *pgxpool.Pool
}

func NewDashboardRepository(pool *pgxpool.Pool) *DashboardRepository {
	return &DashboardRepository{pool: pool}
}

func (r *DashboardRepository) GetSummary(ctx context.Context) (DashboardSummaryRow, error) {
	const q = `
		WITH active_provider AS (
			SELECT provider, last_sync_at
			FROM firewall_integrations
			WHERE enabled = TRUE
			ORDER BY updated_at DESC
			LIMIT 1
		),
		policy_counts AS (
			SELECT
				COALESCE(SUM(CASE WHEN policy_type = 'website_category' AND is_enabled = TRUE THEN 1 ELSE 0 END), 0) AS website_count,
				COALESCE(SUM(CASE WHEN policy_type = 'port' AND is_enabled = TRUE THEN 1 ELSE 0 END), 0) AS port_count,
				COALESCE(SUM(CASE WHEN is_enabled = TRUE THEN 1 ELSE 0 END), 0) AS total_count
			FROM firewall_policies
		),
		department_counts AS (
			SELECT COALESCE(COUNT(*), 0) AS department_count FROM departments
		)
		SELECT
			COALESCE((SELECT provider FROM active_provider), 'not_configured') AS provider,
			(SELECT last_sync_at FROM active_provider) AS last_sync_at,
			(SELECT website_count FROM policy_counts) AS website_count,
			(SELECT port_count FROM policy_counts) AS port_count,
			(SELECT total_count FROM policy_counts) AS total_count,
			(SELECT department_count FROM department_counts) AS department_count,
			0 AS pending_changes`

	var row DashboardSummaryRow
	err := r.pool.QueryRow(ctx, q).Scan(
		&row.Provider,
		&row.LastSyncAt,
		&row.WebsitePolicyCount,
		&row.PortPolicyCount,
		&row.TotalPolicyCount,
		&row.DepartmentGroupCount,
		&row.PendingChanges,
	)
	return row, err
}

func (r *DashboardRepository) ListRecentAuditLogs(ctx context.Context, limit int) ([]AuditLogItem, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	const q = `
		SELECT id, action, entity_type, created_at
		FROM audit_logs
		ORDER BY created_at DESC
		LIMIT $1`

	rows, err := r.pool.Query(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]AuditLogItem, 0, limit)
	for rows.Next() {
		var item AuditLogItem
		if err := rows.Scan(&item.ID, &item.Action, &item.Entity, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
