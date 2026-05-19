package firewall

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"firewall-manager/internal/models"
)

type OPNsenseProvider struct {
	baseURL string
	apiKey  string
	secret  string
	dryRun  bool
	client  *http.Client
}

func NewOPNsenseProvider(baseURL, apiKey, secret string, dryRun bool) *OPNsenseProvider {
	return &OPNsenseProvider{
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		apiKey:  strings.TrimSpace(apiKey),
		secret:  strings.TrimSpace(secret),
		dryRun:  dryRun,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *OPNsenseProvider) Sync(ctx context.Context, policies []models.FirewallPolicy) (SyncResult, error) {
	if p.baseURL == "" || p.apiKey == "" || p.secret == "" {
		return SyncResult{}, fmt.Errorf("opnsense base url, key, and secret are required")
	}

	if err := p.healthCheck(ctx); err != nil {
		return SyncResult{}, err
	}

	commands := make([]string, 0, len(policies))
	applied := 0
	skipped := 0
	warnings := make([]string, 0)

	for _, policy := range policies {
		if !policy.IsEnabled {
			skipped++
			continue
		}
		payload, _ := json.Marshal(map[string]any{
			"id":          policy.ID,
			"name":        policy.Name,
			"policy_type": policy.PolicyType,
			"action":      policy.Action,
			"target":      policy.Target,
		})
		commands = append(commands, string(payload))
		skipped++
	}

	warnings = append(warnings, "OPNsense sync currently runs in compatibility mode: policies are validated and prepared, but auto-push should be enabled carefully after endpoint-specific mapping is finalized.")

	return SyncResult{
		Provider: "opnsense",
		Applied:  applied,
		Skipped:  skipped,
		Commands: commands,
		Warnings: warnings,
		DryRun:   p.dryRun,
	}, nil
}

func (p *OPNsenseProvider) healthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"/api/core/system/status", nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(p.apiKey, p.secret)

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("opnsense health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("opnsense health check returned status %d", resp.StatusCode)
	}
	return nil
}
