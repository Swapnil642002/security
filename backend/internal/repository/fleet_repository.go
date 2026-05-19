package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"firewall-manager/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

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
		       is_active, last_seen_at, created_at, updated_at
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
	const q = `
		INSERT INTO employee_laptops (hostname, employee_name, employee_email, os_type, department_id, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, hostname, employee_name, employee_email, os_type, department_id,
		          is_active, last_seen_at, created_at, updated_at`
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
	).Scan(
		&out.ID,
		&out.Hostname,
		&out.EmployeeName,
		&out.EmployeeEmail,
		&out.OSType,
		&out.DepartmentID,
		&out.IsActive,
		&out.LastSeenAt,
		&out.CreatedAt,
		&out.UpdatedAt,
	)
	return out, err
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

func (r *FleetRepository) SetLaptopActive(ctx context.Context, laptopID int64, isActive bool) error {
	const q = `
		UPDATE employee_laptops
		SET is_active = $1, updated_at = NOW()
		WHERE id = $2`
	_, err := r.pool.Exec(ctx, q, isActive, laptopID)
	return err
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
