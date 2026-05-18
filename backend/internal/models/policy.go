package models

import "time"

type FirewallPolicy struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	PolicyType   string    `json:"policy_type"`
	Action       string    `json:"action"`
	Target       string    `json:"target"`
	Department   string    `json:"department"`
	ScheduleJSON string    `json:"schedule_json"`
	IsEnabled    bool      `json:"is_enabled"`
	CreatedBy    int64     `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (p FirewallPolicy) Validate() bool {
	if p.Name == "" || p.Target == "" {
		return false
	}
	if p.PolicyType != "website_category" && p.PolicyType != "port" {
		return false
	}
	if p.Action != "allow" && p.Action != "block" {
		return false
	}
	if p.ScheduleJSON == "" {
		return false
	}
	return true
}
