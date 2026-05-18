package service

import (
	"context"
	"errors"
	"strings"

	"firewall-manager/internal/models"
	"firewall-manager/internal/repository"
)

var ErrFleetForbidden = errors.New("admin access required")

type FleetService struct {
	fleetRepo *repository.FleetRepository
	userRepo  *repository.UserRepository
}

func NewFleetService(fleetRepo *repository.FleetRepository, userRepo *repository.UserRepository) *FleetService {
	return &FleetService{fleetRepo: fleetRepo, userRepo: userRepo}
}

func (s *FleetService) ListDepartments(ctx context.Context, actorUserID int64) ([]models.Department, error) {
	if err := s.ensureAdmin(ctx, actorUserID); err != nil {
		return nil, err
	}
	return s.fleetRepo.ListDepartments(ctx)
}

func (s *FleetService) CreateDepartment(ctx context.Context, actorUserID int64, name, description string) (models.Department, error) {
	if err := s.ensureAdmin(ctx, actorUserID); err != nil {
		return models.Department{}, err
	}
	d, err := s.fleetRepo.CreateDepartment(ctx, name, description)
	if err != nil {
		return models.Department{}, err
	}
	_ = s.fleetRepo.InsertAuditLog(ctx, actorUserID, "department.create", "department", d.ID, map[string]any{"name": d.Name})
	return d, nil
}

func (s *FleetService) ListLaptops(ctx context.Context, actorUserID int64) ([]models.EmployeeLaptop, error) {
	if err := s.ensureAdmin(ctx, actorUserID); err != nil {
		return nil, err
	}
	return s.fleetRepo.ListLaptops(ctx)
}

func (s *FleetService) CreateLaptop(ctx context.Context, actorUserID int64, laptop models.EmployeeLaptop) (models.EmployeeLaptop, error) {
	if err := s.ensureAdmin(ctx, actorUserID); err != nil {
		return models.EmployeeLaptop{}, err
	}
	laptop.Hostname = strings.TrimSpace(laptop.Hostname)
	laptop.EmployeeName = strings.TrimSpace(laptop.EmployeeName)
	laptop.EmployeeEmail = strings.TrimSpace(laptop.EmployeeEmail)
	if laptop.Hostname == "" || laptop.EmployeeName == "" || laptop.EmployeeEmail == "" {
		return models.EmployeeLaptop{}, errors.New("hostname, employee_name, and employee_email are required")
	}
	if laptop.OSType != "windows" && laptop.OSType != "macos" && laptop.OSType != "linux" {
		return models.EmployeeLaptop{}, errors.New("os_type must be windows, macos, or linux")
	}

	created, err := s.fleetRepo.CreateLaptop(ctx, laptop)
	if err != nil {
		return models.EmployeeLaptop{}, err
	}
	_ = s.fleetRepo.InsertAuditLog(ctx, actorUserID, "laptop.create", "employee_laptop", created.ID, map[string]any{
		"hostname": created.Hostname,
		"os_type":  created.OSType,
	})
	return created, nil
}

func (s *FleetService) CreatePolicyAssignment(ctx context.Context, actorUserID int64, assignment models.PolicyAssignment) (models.PolicyAssignment, error) {
	if err := s.ensureAdmin(ctx, actorUserID); err != nil {
		return models.PolicyAssignment{}, err
	}
	if assignment.PolicyID <= 0 {
		return models.PolicyAssignment{}, errors.New("policy_id is required")
	}
	if assignment.AssignmentType != "department" && assignment.AssignmentType != "laptop" {
		return models.PolicyAssignment{}, errors.New("assignment_type must be department or laptop")
	}
	if assignment.AssignmentType == "department" && assignment.DepartmentID == nil {
		return models.PolicyAssignment{}, errors.New("department_id is required for department assignment")
	}
	if assignment.AssignmentType == "laptop" && assignment.LaptopID == nil {
		return models.PolicyAssignment{}, errors.New("laptop_id is required for laptop assignment")
	}

	created, err := s.fleetRepo.CreatePolicyAssignment(ctx, assignment)
	if err != nil {
		return models.PolicyAssignment{}, err
	}
	_ = s.fleetRepo.InsertAuditLog(ctx, actorUserID, "policy.assign", "policy_assignment", created.ID, map[string]any{
		"policy_id":       created.PolicyID,
		"assignment_type": created.AssignmentType,
	})
	return created, nil
}

func (s *FleetService) ListPolicyAssignments(ctx context.Context, actorUserID, policyID int64) ([]models.PolicyAssignment, error) {
	if err := s.ensureAdmin(ctx, actorUserID); err != nil {
		return nil, err
	}
	return s.fleetRepo.ListPolicyAssignments(ctx, policyID)
}

func (s *FleetService) ensureAdmin(ctx context.Context, actorUserID int64) error {
	u, err := s.userRepo.GetByID(ctx, actorUserID)
	if err != nil {
		return ErrFleetForbidden
	}
	if !u.IsActive || u.Role != "admin" {
		return ErrFleetForbidden
	}
	return nil
}
