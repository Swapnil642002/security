package models

import "time"

type Department struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type EmployeeLaptop struct {
	ID                int64      `json:"id"`
	Hostname          string     `json:"hostname"`
	EmployeeName      string     `json:"employee_name"`
	EmployeeEmail     string     `json:"employee_email"`
	OSType            string     `json:"os_type"`
	DepartmentID      *int64     `json:"department_id"`
	IsActive          bool       `json:"is_active"`
	USBStorageBlocked bool       `json:"usb_storage_blocked"`
	AgentToken        string     `json:"agent_token,omitempty"`
	LastSeenAt        *time.Time `json:"last_seen_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type PolicyAssignment struct {
	ID             int64     `json:"id"`
	PolicyID       int64     `json:"policy_id"`
	AssignmentType string    `json:"assignment_type"`
	DepartmentID   *int64    `json:"department_id"`
	LaptopID       *int64    `json:"laptop_id"`
	IsEnabled      bool      `json:"is_enabled"`
	CreatedAt      time.Time `json:"created_at"`
}
