package firewall

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"firewall-manager/internal/models"
)

type NFTablesProvider struct {
	binPath string
	dryRun  bool
}

func NewNFTablesProvider(binPath string, dryRun bool) *NFTablesProvider {
	if strings.TrimSpace(binPath) == "" {
		binPath = "nft"
	}
	return &NFTablesProvider{binPath: binPath, dryRun: dryRun}
}

func (p *NFTablesProvider) Sync(ctx context.Context, policies []models.FirewallPolicy) (SyncResult, error) {
	commands := []string{
		"add table inet fwmanager",
		"add chain inet fwmanager forward { type filter hook forward priority 0; policy accept; }",
		"flush chain inet fwmanager forward",
	}

	applied := 0
	skipped := 0
	warnings := make([]string, 0)

	for _, policy := range policies {
		if !policy.IsEnabled {
			skipped++
			continue
		}

		switch policy.PolicyType {
		case "port":
			cmd, err := buildPortRule(policy)
			if err != nil {
				skipped++
				warnings = append(warnings, fmt.Sprintf("policy %d skipped: %v", policy.ID, err))
				continue
			}
			commands = append(commands, cmd)
			applied++
		case "website_category":
			cmds, warn, err := buildWebsiteCategoryRules(ctx, policy)
			if err != nil {
				skipped++
				warnings = append(warnings, fmt.Sprintf("policy %d skipped: %v", policy.ID, err))
				continue
			}
			if warn != "" {
				warnings = append(warnings, warn)
			}
			commands = append(commands, cmds...)
			applied++
		default:
			skipped++
			warnings = append(warnings, fmt.Sprintf("policy %d skipped: unsupported type %s", policy.ID, policy.PolicyType))
		}
	}

	if !p.dryRun {
		script := strings.Join(commands, "\n") + "\n"
		cmd := exec.CommandContext(ctx, p.binPath, "-f", "-")
		cmd.Stdin = strings.NewReader(script)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return SyncResult{}, fmt.Errorf("nftables apply failed: %w: %s", err, stderr.String())
		}
	}

	return SyncResult{
		Provider: "nftables",
		Applied:  applied,
		Skipped:  skipped,
		Commands: commands,
		Warnings: warnings,
		DryRun:   p.dryRun,
	}, nil
}

func buildPortRule(policy models.FirewallPolicy) (string, error) {
	parts := strings.Split(strings.TrimSpace(policy.Target), "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("port target must be in format <port>/<protocol>")
	}
	port := parts[0]
	proto := strings.ToLower(parts[1])
	if proto != "tcp" && proto != "udp" {
		return "", fmt.Errorf("protocol must be tcp or udp")
	}

	verdict := "drop"
	if policy.Action == "allow" {
		verdict = "accept"
	}

	comment := strings.ReplaceAll(policy.Name, "\"", "")
	return fmt.Sprintf("add rule inet fwmanager forward %s dport %s %s comment \"%s\"", proto, port, verdict, comment), nil
}

// buildWebsiteCategoryRules resolves the domains for the given category to IPv4 addresses
// and returns nftables commands that create a named IP set and a forward rule for it.
func buildWebsiteCategoryRules(ctx context.Context, policy models.FirewallPolicy) ([]string, string, error) {
	domains, ok := WebsiteCategories[policy.Target]
	if !ok {
		return nil, "", fmt.Errorf("unknown website category %q; valid: %s",
			policy.Target, strings.Join(KnownCategories(), ", "))
	}

	ips := resolveDomainsToIPs(ctx, domains)
	var warn string
	if len(ips) == 0 {
		return nil, "", fmt.Errorf("category %q: DNS resolution returned no IPv4 addresses (check connectivity)", policy.Target)
	}
	if len(ips) < len(domains)/2 {
		warn = fmt.Sprintf("policy %d: only %d IPs resolved for category %q; coverage may be incomplete",
			policy.ID, len(ips), policy.Target)
	}

	setName := nftSetName(policy.Target, policy.ID)
	verdict := "drop"
	if policy.Action == "allow" {
		verdict = "accept"
	}
	comment := strings.ReplaceAll(policy.Name, "\"", "")

	cmds := []string{
		fmt.Sprintf("add set inet fwmanager %s { type ipv4_addr; flags interval; comment \"%s\"; }", setName, comment),
		fmt.Sprintf("add element inet fwmanager %s { %s }", setName, strings.Join(ips, ", ")),
		fmt.Sprintf("add rule inet fwmanager forward ip daddr @%s %s comment \"%s\"", setName, verdict, comment),
	}
	return cmds, warn, nil
}
