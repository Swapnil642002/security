package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"firewall-manager/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrEnrollmentNotFound = errors.New("enrollment not found")

type EnrollmentRepository struct {
	pool *pgxpool.Pool
}

func NewEnrollmentRepository(pool *pgxpool.Pool) *EnrollmentRepository {
	return &EnrollmentRepository{pool: pool}
}

func (r *EnrollmentRepository) CreateLink(ctx context.Context, link models.EnrollmentLink) (models.EnrollmentLink, error) {
	const q = `
		INSERT INTO enrollment_links (token, created_by, expires_at, max_uses, is_active, require_approval)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, token, created_by, expires_at, max_uses, used_count, is_active, require_approval, created_at`
	var out models.EnrollmentLink
	err := r.pool.QueryRow(ctx, q, link.Token, link.CreatedBy, link.ExpiresAt, link.MaxUses, link.IsActive, link.RequireApproval).Scan(
		&out.ID,
		&out.Token,
		&out.CreatedBy,
		&out.ExpiresAt,
		&out.MaxUses,
		&out.UsedCount,
		&out.IsActive,
		&out.RequireApproval,
		&out.CreatedAt,
	)
	return out, err
}

func (r *EnrollmentRepository) GetLinkByToken(ctx context.Context, token string) (models.EnrollmentLink, error) {
	const q = `
		SELECT id, token, created_by, expires_at, max_uses, used_count, is_active, require_approval, created_at
		FROM enrollment_links
		WHERE token = $1`
	var out models.EnrollmentLink
	err := r.pool.QueryRow(ctx, q, token).Scan(
		&out.ID,
		&out.Token,
		&out.CreatedBy,
		&out.ExpiresAt,
		&out.MaxUses,
		&out.UsedCount,
		&out.IsActive,
		&out.RequireApproval,
		&out.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.EnrollmentLink{}, ErrEnrollmentNotFound
	}
	return out, err
}

func (r *EnrollmentRepository) IncrementLinkUsage(ctx context.Context, linkID int64) error {
	const q = `UPDATE enrollment_links SET used_count = used_count + 1 WHERE id = $1`
	_, err := r.pool.Exec(ctx, q, linkID)
	return err
}

func (r *EnrollmentRepository) CreateEnrollment(ctx context.Context, e models.DeviceEnrollment) (models.DeviceEnrollment, error) {
	const q = `
		INSERT INTO device_enrollments (link_id, status, hostname, employee_name, employee_email, os_type, current_ip, fingerprint)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, link_id, status, hostname, employee_name, employee_email, os_type, current_ip, fingerprint,
		          laptop_id, approved_by, approved_at, disabled_by, disabled_at, created_at, updated_at`
	var out models.DeviceEnrollment
	err := r.pool.QueryRow(ctx, q, e.LinkID, e.Status, e.Hostname, e.EmployeeName, e.EmployeeEmail, e.OSType, e.CurrentIP, e.Fingerprint).Scan(
		&out.ID,
		&out.LinkID,
		&out.Status,
		&out.Hostname,
		&out.EmployeeName,
		&out.EmployeeEmail,
		&out.OSType,
		&out.CurrentIP,
		&out.Fingerprint,
		&out.LaptopID,
		&out.ApprovedBy,
		&out.ApprovedAt,
		&out.DisabledBy,
		&out.DisabledAt,
		&out.CreatedAt,
		&out.UpdatedAt,
	)
	return out, err
}

func (r *EnrollmentRepository) ListEnrollments(ctx context.Context, status string) ([]models.DeviceEnrollment, error) {
	base := `
		SELECT id, link_id, status, hostname, employee_name, employee_email, os_type, current_ip, fingerprint,
		       laptop_id, approved_by, approved_at, disabled_by, disabled_at, created_at, updated_at
		FROM device_enrollments`
	var rows pgx.Rows
	var err error
	if status != "" {
		rows, err = r.pool.Query(ctx, base+` WHERE status = $1 ORDER BY created_at DESC`, status)
	} else {
		rows, err = r.pool.Query(ctx, base+` ORDER BY created_at DESC`)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.DeviceEnrollment, 0)
	for rows.Next() {
		var out models.DeviceEnrollment
		if err := rows.Scan(
			&out.ID,
			&out.LinkID,
			&out.Status,
			&out.Hostname,
			&out.EmployeeName,
			&out.EmployeeEmail,
			&out.OSType,
			&out.CurrentIP,
			&out.Fingerprint,
			&out.LaptopID,
			&out.ApprovedBy,
			&out.ApprovedAt,
			&out.DisabledBy,
			&out.DisabledAt,
			&out.CreatedAt,
			&out.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, out)
	}
	return items, rows.Err()
}

func (r *EnrollmentRepository) GetEnrollmentByID(ctx context.Context, id int64) (models.DeviceEnrollment, error) {
	const q = `
		SELECT id, link_id, status, hostname, employee_name, employee_email, os_type, current_ip, fingerprint,
		       laptop_id, approved_by, approved_at, disabled_by, disabled_at, created_at, updated_at
		FROM device_enrollments
		WHERE id = $1`
	var out models.DeviceEnrollment
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&out.ID,
		&out.LinkID,
		&out.Status,
		&out.Hostname,
		&out.EmployeeName,
		&out.EmployeeEmail,
		&out.OSType,
		&out.CurrentIP,
		&out.Fingerprint,
		&out.LaptopID,
		&out.ApprovedBy,
		&out.ApprovedAt,
		&out.DisabledBy,
		&out.DisabledAt,
		&out.CreatedAt,
		&out.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.DeviceEnrollment{}, ErrEnrollmentNotFound
	}
	return out, err
}

func (r *EnrollmentRepository) ApproveEnrollment(ctx context.Context, enrollmentID, approvedBy, laptopID int64) error {
	const q = `
		UPDATE device_enrollments
		SET status = 'approved', approved_by = $1, approved_at = NOW(), laptop_id = $2, updated_at = NOW()
		WHERE id = $3`
	ct, err := r.pool.Exec(ctx, q, approvedBy, laptopID, enrollmentID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrEnrollmentNotFound
	}
	return nil
}

func (r *EnrollmentRepository) DisableEnrollment(ctx context.Context, enrollmentID, disabledBy int64) error {
	const q = `
		UPDATE device_enrollments
		SET status = 'disabled', disabled_by = $1, disabled_at = NOW(), updated_at = NOW()
		WHERE id = $2`
	ct, err := r.pool.Exec(ctx, q, disabledBy, enrollmentID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrEnrollmentNotFound
	}
	return nil
}

func (r *EnrollmentRepository) InsertNotification(ctx context.Context, n models.SystemNotification) (models.SystemNotification, error) {
	const q = `
		INSERT INTO system_notifications (type, message, target_role, is_read)
		VALUES ($1, $2, $3, $4)
		RETURNING id, type, message, target_role, is_read, created_at`
	var out models.SystemNotification
	err := r.pool.QueryRow(ctx, q, n.Type, n.Message, n.TargetRole, n.IsRead).Scan(
		&out.ID,
		&out.Type,
		&out.Message,
		&out.TargetRole,
		&out.IsRead,
		&out.CreatedAt,
	)
	return out, err
}

func (r *EnrollmentRepository) ListNotifications(ctx context.Context, limit int) ([]models.SystemNotification, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	const q = `
		SELECT id, type, message, target_role, is_read, created_at
		FROM system_notifications
		ORDER BY created_at DESC
		LIMIT $1`
	rows, err := r.pool.Query(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]models.SystemNotification, 0, limit)
	for rows.Next() {
		var out models.SystemNotification
		if err := rows.Scan(&out.ID, &out.Type, &out.Message, &out.TargetRole, &out.IsRead, &out.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, out)
	}
	return items, rows.Err()
}

func (r *EnrollmentRepository) InsertAuditLog(ctx context.Context, actorUserID int64, action, entityType string, entityID int64, details any) error {
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
