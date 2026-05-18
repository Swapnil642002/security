package service

import (
	"context"
	"fmt"
	"strings"

	"firewall-manager/internal/firewall"
	"firewall-manager/internal/repository"
)

type FirewallSyncService struct {
	userRepo    *repository.UserRepository
	policyRepo  *repository.PolicyRepository
	provider    firewall.Provider
	providerTag string
}

func NewFirewallSyncService(userRepo *repository.UserRepository, policyRepo *repository.PolicyRepository, provider firewall.Provider, providerTag string) *FirewallSyncService {
	return &FirewallSyncService{
		userRepo:    userRepo,
		policyRepo:  policyRepo,
		provider:    provider,
		providerTag: providerTag,
	}
}

func (s *FirewallSyncService) SyncPolicies(ctx context.Context, actorUserID int64) (firewall.SyncResult, error) {
	u, err := s.userRepo.GetByID(ctx, actorUserID)
	if err != nil || !u.IsActive || u.Role != "admin" {
		return firewall.SyncResult{}, fmt.Errorf("admin access required")
	}

	policies, err := s.policyRepo.List(ctx)
	if err != nil {
		return firewall.SyncResult{}, err
	}

	res, err := s.provider.Sync(ctx, policies)
	if err != nil {
		return firewall.SyncResult{}, err
	}

	_ = s.policyRepo.InsertAuditLog(ctx, actorUserID, "firewall.sync", "firewall_provider", 0, map[string]any{
		"provider": s.providerTag,
		"applied":  res.Applied,
		"skipped":  res.Skipped,
		"dry_run":  res.DryRun,
	})
	return res, nil
}

func BuildProvider(providerName string, dryRun bool, nftBin, opnsenseURL, opnsenseKey, opnsenseSecret string) (firewall.Provider, string, error) {
	switch strings.ToLower(strings.TrimSpace(providerName)) {
	case "nftables", "":
		return firewall.NewNFTablesProvider(nftBin, dryRun), "nftables", nil
	case "opnsense":
		return firewall.NewOPNsenseProvider(opnsenseURL, opnsenseKey, opnsenseSecret, dryRun), "opnsense", nil
	default:
		return nil, "", fmt.Errorf("unsupported FIREWALL_PROVIDER: %s", providerName)
	}
}
