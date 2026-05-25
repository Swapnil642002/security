package service

import (
	"context"
	"errors"
	"strings"

	"firewall-manager/internal/models"
	"firewall-manager/internal/repository"
)

var (
	ErrPolicyForbidden = errors.New("admin access required")
	ErrInvalidPolicy   = errors.New("invalid policy payload")
)

type PolicyService struct {
	policies *repository.PolicyRepository
	users    *repository.UserRepository
}

func NewPolicyService(policies *repository.PolicyRepository, users *repository.UserRepository) *PolicyService {
	return &PolicyService{policies: policies, users: users}
}

func (s *PolicyService) List(ctx context.Context, actorUserID int64) ([]models.FirewallPolicy, error) {
	if err := s.ensureAdmin(ctx, actorUserID); err != nil {
		return nil, err
	}
	return s.policies.List(ctx)
}

func (s *PolicyService) Create(ctx context.Context, actorUserID int64, p models.FirewallPolicy) (models.FirewallPolicy, error) {
	if err := s.ensureAdmin(ctx, actorUserID); err != nil {
		return models.FirewallPolicy{}, err
	}
	normalizePolicyInput(&p)
	if !p.Validate() {
		return models.FirewallPolicy{}, ErrInvalidPolicy
	}
	p.CreatedBy = actorUserID

	created, err := s.policies.Create(ctx, p)
	if err != nil {
		return models.FirewallPolicy{}, err
	}
	_ = s.policies.InsertAuditLog(ctx, actorUserID, "policy.create", "firewall_policy", created.ID, map[string]any{
		"policy_type": created.PolicyType,
		"action":      created.Action,
		"target":      created.Target,
	})
	return created, nil
}

func (s *PolicyService) Update(ctx context.Context, actorUserID, policyID int64, p models.FirewallPolicy) (models.FirewallPolicy, error) {
	if err := s.ensureAdmin(ctx, actorUserID); err != nil {
		return models.FirewallPolicy{}, err
	}
	normalizePolicyInput(&p)
	if !p.Validate() {
		return models.FirewallPolicy{}, ErrInvalidPolicy
	}

	updated, err := s.policies.Update(ctx, policyID, p)
	if err != nil {
		return models.FirewallPolicy{}, err
	}
	_ = s.policies.InsertAuditLog(ctx, actorUserID, "policy.update", "firewall_policy", updated.ID, map[string]any{
		"policy_type": updated.PolicyType,
		"action":      updated.Action,
		"target":      updated.Target,
	})
	return updated, nil
}

func (s *PolicyService) Delete(ctx context.Context, actorUserID, policyID int64) error {
	if err := s.ensureAdmin(ctx, actorUserID); err != nil {
		return err
	}
	if err := s.policies.Delete(ctx, policyID); err != nil {
		return err
	}
	_ = s.policies.InsertAuditLog(ctx, actorUserID, "policy.delete", "firewall_policy", policyID, nil)
	return nil
}

func (s *PolicyService) ensureAdmin(ctx context.Context, actorUserID int64) error {
	u, err := s.users.GetByID(ctx, actorUserID)
	if err != nil {
		return ErrPolicyForbidden
	}
	if !u.IsActive || u.Role != "admin" {
		return ErrPolicyForbidden
	}
	return nil
}

func normalizePolicyInput(p *models.FirewallPolicy) {
	p.Name = strings.TrimSpace(p.Name)
	p.PolicyType = strings.ToLower(strings.TrimSpace(p.PolicyType))
	p.Action = strings.ToLower(strings.TrimSpace(p.Action))
	p.Target = strings.TrimSpace(p.Target)

	if p.PolicyType == "website_category" {
		target := strings.ToLower(p.Target)
		target = strings.ReplaceAll(target, "-", "_")
		target = strings.Join(strings.Fields(target), "_")
		p.Target = target
	}
}
