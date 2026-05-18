package service

import (
	"context"
	"time"

	"firewall-manager/internal/repository"
)

type DashboardSummary struct {
	FirewallProvider string                    `json:"firewall_provider"`
	LastSyncAt       *time.Time                `json:"last_sync_at"`
	PolicyCounts     DashboardPolicyCounts     `json:"policy_counts"`
	RecentAuditLogs  []repository.AuditLogItem `json:"recent_audit_logs"`
}

type DashboardPolicyCounts struct {
	WebsiteCategory int `json:"website_category"`
	Port            int `json:"port"`
	Total           int `json:"total"`
	DepartmentGroup int `json:"department_group"`
	PendingChanges  int `json:"pending_changes"`
}

type DashboardService struct {
	dashboardRepo *repository.DashboardRepository
}

func NewDashboardService(dashboardRepo *repository.DashboardRepository) *DashboardService {
	return &DashboardService{dashboardRepo: dashboardRepo}
}

func (s *DashboardService) GetSummary(ctx context.Context) (DashboardSummary, error) {
	row, err := s.dashboardRepo.GetSummary(ctx)
	if err != nil {
		return DashboardSummary{}, err
	}

	logs, err := s.dashboardRepo.ListRecentAuditLogs(ctx, 8)
	if err != nil {
		return DashboardSummary{}, err
	}

	return DashboardSummary{
		FirewallProvider: row.Provider,
		LastSyncAt:       row.LastSyncAt,
		PolicyCounts: DashboardPolicyCounts{
			WebsiteCategory: row.WebsitePolicyCount,
			Port:            row.PortPolicyCount,
			Total:           row.TotalPolicyCount,
			DepartmentGroup: row.DepartmentGroupCount,
			PendingChanges:  row.PendingChanges,
		},
		RecentAuditLogs: logs,
	}, nil
}
