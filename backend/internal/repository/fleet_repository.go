package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"firewall-manager/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrLaptopNotFound = errors.New("laptop not found")

type FleetRepository struct {
	pool *pgxpool.Pool
}

func NewFleetRepository(pool *pgxpool.Pool) *FleetRepository {
	return &FleetRepository{pool: pool}
}

func (r *FleetRepository) ListDepartments(ctx context.Context) ([]models.Department, error) {
	const q = `SELECT id, name, description, created_at, updated_at FROM departments ORDER BY name ASC`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.Department, 0)
	for rows.Next() {
		var d models.Department
		if err := rows.Scan(&d.ID, &d.Name, &d.Description, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, d)
	}
	return items, rows.Err()
}

func (r *FleetRepository) CreateDepartment(ctx context.Context, name, description string) (models.Department, error) {
	const q = `
		INSERT INTO departments (name, description)
		VALUES ($1, $2)
		RETURNING id, name, description, created_at, updated_at`
	var d models.Department
	err := r.pool.QueryRow(ctx, q, strings.TrimSpace(name), strings.TrimSpace(description)).Scan(
		&d.ID, &d.Name, &d.Description, &d.CreatedAt, &d.UpdatedAt,
	)
	return d, err
}

func (r *FleetRepository) ListLaptops(ctx context.Context) ([]models.EmployeeLaptop, error) {
	const q = `
		SELECT id, hostname, employee_name, employee_email, os_type, department_id,
		       is_active, usb_storage_blocked, COALESCE(agent_token, ''), last_seen_at, created_at, updated_at
		FROM employee_laptops
		ORDER BY hostname ASC`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.EmployeeLaptop, 0)
	for rows.Next() {
		var l models.EmployeeLaptop
		if err := rows.Scan(
			&l.ID,
			&l.Hostname,
			&l.EmployeeName,
			&l.EmployeeEmail,
			&l.OSType,
			&l.DepartmentID,
			&l.IsActive,
			&l.USBStorageBlocked,
			&l.AgentToken,
			&l.LastSeenAt,
			&l.CreatedAt,
			&l.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, l)
	}
	return items, rows.Err()
}

func (r *FleetRepository) CreateLaptop(ctx context.Context, l models.EmployeeLaptop) (models.EmployeeLaptop, error) {
	if strings.TrimSpace(l.AgentToken) == "" {
		token, err := randomAgentToken(24)
		if err != nil {
			return models.EmployeeLaptop{}, err
		}
		l.AgentToken = token
	}

	const q = `
		INSERT INTO employee_laptops (hostname, employee_name, employee_email, os_type, department_id, is_active, usb_storage_blocked, agent_token)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, hostname, employee_name, employee_email, os_type, department_id,
		          is_active, usb_storage_blocked, COALESCE(agent_token, ''), last_seen_at, created_at, updated_at`
	var out models.EmployeeLaptop
	err := r.pool.QueryRow(
		ctx,
		q,
		strings.TrimSpace(l.Hostname),
		strings.TrimSpace(l.EmployeeName),
		strings.ToLower(strings.TrimSpace(l.EmployeeEmail)),
		l.OSType,
		l.DepartmentID,
		l.IsActive,
		l.USBStorageBlocked,
		strings.TrimSpace(l.AgentToken),
	).Scan(
		&out.ID,
		&out.Hostname,
		&out.EmployeeName,
		&out.EmployeeEmail,
		&out.OSType,
		&out.DepartmentID,
		&out.IsActive,
		&out.USBStorageBlocked,
		&out.AgentToken,
		&out.LastSeenAt,
		&out.CreatedAt,
		&out.UpdatedAt,
	)
	return out, err
}

func (r *FleetRepository) UpsertLaptopByHostname(ctx context.Context, l models.EmployeeLaptop) (models.EmployeeLaptop, error) {
	if strings.TrimSpace(l.AgentToken) == "" {
		token, err := randomAgentToken(24)
		if err != nil {
			return models.EmployeeLaptop{}, err
		}
		l.AgentToken = token
	}

	const q = `
		INSERT INTO employee_laptops (hostname, employee_name, employee_email, os_type, department_id, is_active, usb_storage_blocked, agent_token)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (hostname)
		DO UPDATE
		SET employee_name = EXCLUDED.employee_name,
		    employee_email = EXCLUDED.employee_email,
		    os_type = EXCLUDED.os_type,
		    department_id = EXCLUDED.department_id,
		    is_active = EXCLUDED.is_active,
		    agent_token = COALESCE(employee_laptops.agent_token, EXCLUDED.agent_token),
		    updated_at = NOW()
		RETURNING id, hostname, employee_name, employee_email, os_type, department_id,
		          is_active, usb_storage_blocked, COALESCE(agent_token, ''), last_seen_at, created_at, updated_at`

	var out models.EmployeeLaptop
	err := r.pool.QueryRow(
		ctx,
		q,
		strings.TrimSpace(l.Hostname),
		strings.TrimSpace(l.EmployeeName),
		strings.ToLower(strings.TrimSpace(l.EmployeeEmail)),
		l.OSType,
		l.DepartmentID,
		l.IsActive,
		l.USBStorageBlocked,
		strings.TrimSpace(l.AgentToken),
	).Scan(
		&out.ID,
		&out.Hostname,
		&out.EmployeeName,
		&out.EmployeeEmail,
		&out.OSType,
		&out.DepartmentID,
		&out.IsActive,
		&out.USBStorageBlocked,
		&out.AgentToken,
		&out.LastSeenAt,
		&out.CreatedAt,
		&out.UpdatedAt,
	)
	return out, err
}

func (r *FleetRepository) GetLaptopByID(ctx context.Context, laptopID int64) (models.EmployeeLaptop, error) {
	const q = `
		SELECT id, hostname, employee_name, employee_email, os_type, department_id,
		       is_active, usb_storage_blocked, COALESCE(agent_token, ''), last_seen_at, created_at, updated_at
		FROM employee_laptops
		WHERE id = $1`

	var out models.EmployeeLaptop
	err := r.pool.QueryRow(ctx, q, laptopID).Scan(
		&out.ID,
		&out.Hostname,
		&out.EmployeeName,
		&out.EmployeeEmail,
		&out.OSType,
		&out.DepartmentID,
		&out.IsActive,
		&out.USBStorageBlocked,
		&out.AgentToken,
		&out.LastSeenAt,
		&out.CreatedAt,
		&out.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.EmployeeLaptop{}, ErrLaptopNotFound
	}
	return out, err
}

func (r *FleetRepository) DeleteLaptop(ctx context.Context, laptopID int64) (models.EmployeeLaptop, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return models.EmployeeLaptop{}, err
	}
	defer tx.Rollback(ctx)

	const getQ = `
		SELECT id, hostname, employee_name, employee_email, os_type, department_id,
		       is_active, usb_storage_blocked, COALESCE(agent_token, ''), last_seen_at, created_at, updated_at
		FROM employee_laptops
		WHERE id = $1
		FOR UPDATE`

	var out models.EmployeeLaptop
	if err := tx.QueryRow(ctx, getQ, laptopID).Scan(
		&out.ID,
		&out.Hostname,
		&out.EmployeeName,
		&out.EmployeeEmail,
		&out.OSType,
		&out.DepartmentID,
		&out.IsActive,
		&out.USBStorageBlocked,
		&out.AgentToken,
		&out.LastSeenAt,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.EmployeeLaptop{}, ErrLaptopNotFound
		}
		return models.EmployeeLaptop{}, err
	}

	if _, err := tx.Exec(ctx, `UPDATE device_enrollments SET laptop_id = NULL, updated_at = NOW() WHERE laptop_id = $1`, laptopID); err != nil {
		return models.EmployeeLaptop{}, err
	}

	deleteResult, err := tx.Exec(ctx, `DELETE FROM employee_laptops WHERE id = $1`, laptopID)
	if err != nil {
		return models.EmployeeLaptop{}, err
	}
	if deleteResult.RowsAffected() == 0 {
		return models.EmployeeLaptop{}, ErrLaptopNotFound
	}

	if err := tx.Commit(ctx); err != nil {
		return models.EmployeeLaptop{}, err
	}

	return out, nil
}

func (r *FleetRepository) CreatePolicyAssignment(ctx context.Context, a models.PolicyAssignment) (models.PolicyAssignment, error) {
	const q = `
		INSERT INTO policy_assignments (policy_id, assignment_type, department_id, laptop_id, is_enabled)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, policy_id, assignment_type, department_id, laptop_id, is_enabled, created_at`
	var out models.PolicyAssignment
	err := r.pool.QueryRow(ctx, q, a.PolicyID, a.AssignmentType, a.DepartmentID, a.LaptopID, a.IsEnabled).Scan(
		&out.ID,
		&out.PolicyID,
		&out.AssignmentType,
		&out.DepartmentID,
		&out.LaptopID,
		&out.IsEnabled,
		&out.CreatedAt,
	)
	return out, err
}

func (r *FleetRepository) GetPolicyByID(ctx context.Context, policyID int64) (models.FirewallPolicy, error) {
	const q = `
		SELECT id, name, policy_type, action, target, COALESCE(department, ''), schedule_json::text,
		       is_enabled, created_by, created_at, updated_at
		FROM firewall_policies
		WHERE id = $1`

	var p models.FirewallPolicy
	err := r.pool.QueryRow(ctx, q, policyID).Scan(
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
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.FirewallPolicy{}, ErrPolicyNotFound
	}
	return p, err
}

func (r *FleetRepository) ListActiveLaptopIDsForAssignment(ctx context.Context, a models.PolicyAssignment) ([]int64, error) {
	if a.AssignmentType == "laptop" && a.LaptopID != nil {
		const q = `SELECT id FROM employee_laptops WHERE id = $1 AND is_active = TRUE`
		rows, err := r.pool.Query(ctx, q, *a.LaptopID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		items := make([]int64, 0, 1)
		for rows.Next() {
			var id int64
			if err := rows.Scan(&id); err != nil {
				return nil, err
			}
			items = append(items, id)
		}
		return items, rows.Err()
	}

	if a.AssignmentType == "department" && a.DepartmentID != nil {
		const q = `SELECT id FROM employee_laptops WHERE department_id = $1 AND is_active = TRUE ORDER BY id ASC`
		rows, err := r.pool.Query(ctx, q, *a.DepartmentID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		items := make([]int64, 0)
		for rows.Next() {
			var id int64
			if err := rows.Scan(&id); err != nil {
				return nil, err
			}
			items = append(items, id)
		}
		return items, rows.Err()
	}

	return nil, nil
}

func (r *FleetRepository) SetLaptopActive(ctx context.Context, laptopID int64, isActive bool) error {
	const q = `
		UPDATE employee_laptops
		SET is_active = $1, updated_at = NOW()
		WHERE id = $2`
	_, err := r.pool.Exec(ctx, q, isActive, laptopID)
	return err
}

func (r *FleetRepository) CreateDeviceCommand(ctx context.Context, cmd models.DeviceCommand) (models.DeviceCommand, error) {
	const q = `
		INSERT INTO device_commands (laptop_id, command_type, payload_json, status, created_by)
		VALUES ($1, $2, $3::jsonb, 'pending', $4)
		RETURNING id, laptop_id, command_type, payload_json::text, status, result_text,
		          created_by, started_at, completed_at, created_at, updated_at`

	var out models.DeviceCommand
	err := r.pool.QueryRow(ctx, q, cmd.LaptopID, cmd.CommandType, nonEmptyJSON(cmd.PayloadJSON), cmd.CreatedBy).Scan(
		&out.ID,
		&out.LaptopID,
		&out.CommandType,
		&out.PayloadJSON,
		&out.Status,
		&out.ResultText,
		&out.CreatedBy,
		&out.StartedAt,
		&out.CompletedAt,
		&out.CreatedAt,
		&out.UpdatedAt,
	)
	return out, err
}

func (r *FleetRepository) ClaimNextPendingCommand(ctx context.Context, agentToken string) (*models.DeviceCommand, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	const pick = `
		SELECT dc.id
		FROM device_commands dc
		JOIN employee_laptops el ON el.id = dc.laptop_id
		WHERE el.agent_token = $1
		  AND dc.status = 'pending'
		ORDER BY dc.created_at ASC
		LIMIT 1
		FOR UPDATE SKIP LOCKED`

	var commandID int64
	err = tx.QueryRow(ctx, pick, strings.TrimSpace(agentToken)).Scan(&commandID)
	if errors.Is(err, pgx.ErrNoRows) {
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	const claim = `
		UPDATE device_commands
		SET status = 'in_progress', started_at = NOW(), updated_at = NOW()
		WHERE id = $1
		RETURNING id, laptop_id, command_type, payload_json::text, status, result_text,
		          created_by, started_at, completed_at, created_at, updated_at`

	var out models.DeviceCommand
	if err := tx.QueryRow(ctx, claim, commandID).Scan(
		&out.ID,
		&out.LaptopID,
		&out.CommandType,
		&out.PayloadJSON,
		&out.Status,
		&out.ResultText,
		&out.CreatedBy,
		&out.StartedAt,
		&out.CompletedAt,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &out, nil
}

func (r *FleetRepository) CompleteClaimedCommand(ctx context.Context, agentToken string, commandID int64, success bool, resultText string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	status := "failed"
	if success {
		status = "success"
	}

	const complete = `
		UPDATE device_commands dc
		SET status = $1,
		    result_text = $2,
		    completed_at = NOW(),
		    updated_at = NOW()
		FROM employee_laptops el
		WHERE dc.id = $3
		  AND dc.laptop_id = el.id
		  AND el.agent_token = $4
		  AND dc.status IN ('pending', 'in_progress')
		RETURNING dc.command_type, dc.laptop_id`

	var commandType string
	var laptopID int64
	err = tx.QueryRow(ctx, complete, status, strings.TrimSpace(resultText), commandID, strings.TrimSpace(agentToken)).Scan(&commandType, &laptopID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrLaptopNotFound
	}
	if err != nil {
		return err
	}

	if success {
		blocked := false
		if commandType == "usb.block" {
			blocked = true
		}
		if _, err := tx.Exec(ctx, `UPDATE employee_laptops SET usb_storage_blocked=$1, updated_at=NOW() WHERE id=$2`, blocked, laptopID); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *FleetRepository) ListPolicyAssignments(ctx context.Context, policyID int64) ([]models.PolicyAssignment, error) {
	const q = `
		SELECT id, policy_id, assignment_type, department_id, laptop_id, is_enabled, created_at
		FROM policy_assignments
		WHERE policy_id = $1
		ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, q, policyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.PolicyAssignment, 0)
	for rows.Next() {
		var a models.PolicyAssignment
		if err := rows.Scan(&a.ID, &a.PolicyID, &a.AssignmentType, &a.DepartmentID, &a.LaptopID, &a.IsEnabled, &a.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, a)
	}
	return items, rows.Err()
}

func (r *FleetRepository) InsertAuditLog(ctx context.Context, actorUserID int64, action, entityType string, entityID int64, details any) error {
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
	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}
	return nil
}

func nonEmptyJSON(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return "{}"
	}
	return v
}

func randomAgentToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
