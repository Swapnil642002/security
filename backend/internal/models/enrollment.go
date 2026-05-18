package models

import "time"

type EnrollmentLink struct {
	ID              int64     `json:"id"`
	Token           string    `json:"token"`
	CreatedBy       int64     `json:"created_by"`
	ExpiresAt       time.Time `json:"expires_at"`
	MaxUses         int       `json:"max_uses"`
	UsedCount       int       `json:"used_count"`
	IsActive        bool      `json:"is_active"`
	RequireApproval bool      `json:"require_approval"`
	CreatedAt       time.Time `json:"created_at"`
}

type DeviceEnrollment struct {
	ID            int64      `json:"id"`
	LinkID        int64      `json:"link_id"`
	Status        string     `json:"status"`
	Hostname      string     `json:"hostname"`
	EmployeeName  string     `json:"employee_name"`
	EmployeeEmail string     `json:"employee_email"`
	OSType        string     `json:"os_type"`
	CurrentIP     string     `json:"current_ip"`
	Fingerprint   string     `json:"fingerprint"`
	LaptopID      *int64     `json:"laptop_id"`
	ApprovedBy    *int64     `json:"approved_by"`
	ApprovedAt    *time.Time `json:"approved_at"`
	DisabledBy    *int64     `json:"disabled_by"`
	DisabledAt    *time.Time `json:"disabled_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type SystemNotification struct {
	ID         int64     `json:"id"`
	Type       string    `json:"type"`
	Message    string    `json:"message"`
	TargetRole string    `json:"target_role"`
	IsRead     bool      `json:"is_read"`
	CreatedAt  time.Time `json:"created_at"`
}
