package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"firewall-manager/internal/models"
	"firewall-manager/internal/repository"
)

var ErrFleetForbidden = errors.New("admin access required")
var ErrAgentUnauthorized = errors.New("invalid agent token")

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

func (s *FleetService) DeleteLaptop(ctx context.Context, actorUserID, laptopID int64) (models.EmployeeLaptop, error) {
	if err := s.ensureAdmin(ctx, actorUserID); err != nil {
		return models.EmployeeLaptop{}, err
	}
	if laptopID <= 0 {
		return models.EmployeeLaptop{}, errors.New("invalid laptop id")
	}

	deleted, err := s.fleetRepo.DeleteLaptop(ctx, laptopID)
	if err != nil {
		return models.EmployeeLaptop{}, err
	}

	_ = s.fleetRepo.InsertAuditLog(ctx, actorUserID, "laptop.delete", "employee_laptop", deleted.ID, map[string]any{
		"hostname": deleted.Hostname,
		"email":    deleted.EmployeeEmail,
	})

	return deleted, nil
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
	if err := s.queueWebsiteCategoryCommandsForAssignment(ctx, actorUserID, created); err != nil {
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

func (s *FleetService) QueueUSBCommand(ctx context.Context, actorUserID, laptopID int64, block bool) (models.DeviceCommand, error) {
	if err := s.ensureAdmin(ctx, actorUserID); err != nil {
		return models.DeviceCommand{}, err
	}
	laptop, err := s.fleetRepo.GetLaptopByID(ctx, laptopID)
	if err != nil {
		return models.DeviceCommand{}, err
	}

	commandType := "usb.unblock"
	action := "usb.unblock"
	if block {
		commandType = "usb.block"
		action = "usb.block"
	}

	payload, _ := json.Marshal(map[string]any{"hostname": laptop.Hostname})
	created, err := s.fleetRepo.CreateDeviceCommand(ctx, models.DeviceCommand{
		LaptopID:    laptop.ID,
		CommandType: commandType,
		PayloadJSON: string(payload),
		CreatedBy:   actorUserID,
	})
	if err != nil {
		return models.DeviceCommand{}, err
	}

	_ = s.fleetRepo.InsertAuditLog(ctx, actorUserID, action, "employee_laptop", laptop.ID, map[string]any{
		"hostname": laptop.Hostname,
		"command":  commandType,
	})
	return created, nil
}

func (s *FleetService) AgentClaimNextCommand(ctx context.Context, agentToken string) (*models.DeviceCommand, error) {
	agentToken = strings.TrimSpace(agentToken)
	if agentToken == "" {
		return nil, ErrAgentUnauthorized
	}
	cmd, err := s.fleetRepo.ClaimNextPendingCommand(ctx, agentToken)
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

func (s *FleetService) AgentReportCommandResult(ctx context.Context, agentToken string, commandID int64, success bool, resultText string) error {
	agentToken = strings.TrimSpace(agentToken)
	if agentToken == "" {
		return ErrAgentUnauthorized
	}
	if commandID <= 0 {
		return fmt.Errorf("invalid command id")
	}
	if err := s.fleetRepo.CompleteClaimedCommand(ctx, agentToken, commandID, success, resultText); err != nil {
		if errors.Is(err, repository.ErrLaptopNotFound) {
			return ErrAgentUnauthorized
		}
		return err
	}
	return nil
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

func (s *FleetService) queueWebsiteCategoryCommandsForAssignment(ctx context.Context, actorUserID int64, assignment models.PolicyAssignment) error {
	policy, err := s.fleetRepo.GetPolicyByID(ctx, assignment.PolicyID)
	if err != nil {
		return err
	}
	if policy.PolicyType != "website_category" {
		return nil
	}

	laptopIDs, err := s.fleetRepo.ListActiveLaptopIDsForAssignment(ctx, assignment)
	if err != nil {
		return err
	}
	if len(laptopIDs) == 0 {
		return nil
	}

	commandType := "website.unblock_category"
	if policy.Action == "block" && policy.IsEnabled && assignment.IsEnabled {
		commandType = "website.block_category"
	}

	payloadBytes, _ := json.Marshal(map[string]any{
		"category":    policy.Target,
		"policy_id":   policy.ID,
		"policy_name": policy.Name,
	})

	for _, laptopID := range laptopIDs {
		if _, err := s.fleetRepo.CreateDeviceCommand(ctx, models.DeviceCommand{
			LaptopID:    laptopID,
			CommandType: commandType,
			PayloadJSON: string(payloadBytes),
			CreatedBy:   actorUserID,
		}); err != nil {
			return err
		}
	}

	_ = s.fleetRepo.InsertAuditLog(ctx, actorUserID, "policy.assign.command_queue", "firewall_policy", policy.ID, map[string]any{
		"assignment_id": assignment.ID,
		"command_type":  commandType,
		"target_count":  len(laptopIDs),
		"category":      policy.Target,
	})

	return nil
}
