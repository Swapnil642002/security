package repository

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"firewall-manager/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrPolicyNotFound = errors.New("policy not found")

type PolicyRepository struct {
	pool *pgxpool.Pool
}

func NewPolicyRepository(pool *pgxpool.Pool) *PolicyRepository {
	return &PolicyRepository{pool: pool}
}

func (r *PolicyRepository) List(ctx context.Context) ([]models.FirewallPolicy, error) {
	const q = `
		SELECT id, name, policy_type, action, target, COALESCE(department, ''), schedule_json::text,
		       is_enabled, created_by, created_at, updated_at
		FROM firewall_policies
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.FirewallPolicy, 0)
	for rows.Next() {
		var p models.FirewallPolicy
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.PolicyType,
			&p.Action,
			&p.Target,
			&p.Department,
			&p.ScheduleJSON,
			&p.IsEnabled,
			&p.CreatedBy,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *PolicyRepository) Create(ctx context.Context, p models.FirewallPolicy) (models.FirewallPolicy, error) {
	const q = `
		INSERT INTO firewall_policies (name, policy_type, action, target, department, schedule_json, is_enabled, created_by)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8)
		RETURNING id, name, policy_type, action, target, COALESCE(department, ''), schedule_json::text,
		          is_enabled, created_by, created_at, updated_at`

	var created models.FirewallPolicy
	err := r.pool.QueryRow(
		ctx,
		q,
		strings.TrimSpace(p.Name),
		p.PolicyType,
		p.Action,
		strings.TrimSpace(p.Target),
		nullIfEmpty(strings.TrimSpace(p.Department)),
		p.ScheduleJSON,
		p.IsEnabled,
		p.CreatedBy,
	).Scan(
		&created.ID,
		&created.Name,
		&created.PolicyType,
		&created.Action,
		&created.Target,
		&created.Department,
		&created.ScheduleJSON,
		&created.IsEnabled,
		&created.CreatedBy,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	return created, err
}

func (r *PolicyRepository) Update(ctx context.Context, id int64, p models.FirewallPolicy) (models.FirewallPolicy, error) {
	const q = `
		UPDATE firewall_policies
		SET name = $1,
		    policy_type = $2,
		    action = $3,
		    target = $4,
		    department = $5,
		    schedule_json = $6::jsonb,
		    is_enabled = $7,
		    updated_at = NOW()
		WHERE id = $8
		RETURNING id, name, policy_type, action, target, COALESCE(department, ''), schedule_json::text,
		          is_enabled, created_by, created_at, updated_at`

	var updated models.FirewallPolicy
	err := r.pool.QueryRow(
		ctx,
		q,
		strings.TrimSpace(p.Name),
		p.PolicyType,
		p.Action,
		strings.TrimSpace(p.Target),
		nullIfEmpty(strings.TrimSpace(p.Department)),
		p.ScheduleJSON,
		p.IsEnabled,
		id,
	).Scan(
		&updated.ID,
		&updated.Name,
		&updated.PolicyType,
		&updated.Action,
		&updated.Target,
		&updated.Department,
		&updated.ScheduleJSON,
		&updated.IsEnabled,
		&updated.CreatedBy,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.FirewallPolicy{}, ErrPolicyNotFound
	}
	return updated, err
}

func (r *PolicyRepository) Delete(ctx context.Context, id int64) error {
	const q = `DELETE FROM firewall_policies WHERE id = $1`
	ct, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrPolicyNotFound
	}
	return nil
}

func (r *PolicyRepository) InsertAuditLog(ctx context.Context, actorUserID int64, action, entityType string, entityID int64, details any) error {
	detailsJSON := "{}"
	if details != nil {
		b, err := json.Marshal(details)
		if err != nil {
			return err
		}
		detailsJSON = string(b)
	}

	const q = `
		INSERT INTO audit_logs (actor_user_id, action, entity_type, entity_id, details_json, created_at)
		VALUES ($1, $2, $3, $4, $5::jsonb, $6)`
	_, err := r.pool.Exec(ctx, q, actorUserID, action, entityType, entityID, detailsJSON, time.Now().UTC())
	return err
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}
