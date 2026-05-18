package firewall

import (
	"context"

	"firewall-manager/internal/models"
)

type SyncResult struct {
	Provider string   `json:"provider"`
	Applied  int      `json:"applied"`
	Skipped  int      `json:"skipped"`
	Commands []string `json:"commands"`
	Warnings []string `json:"warnings"`
	DryRun   bool     `json:"dry_run"`
}

type Provider interface {
	Sync(ctx context.Context, policies []models.FirewallPolicy) (SyncResult, error)
}
