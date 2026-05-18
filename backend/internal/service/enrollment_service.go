package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"firewall-manager/internal/models"
	"firewall-manager/internal/repository"
)

var (
	ErrPermissionRequired = errors.New("admin or manager approval required")
	ErrInvalidLink        = errors.New("invalid or expired enrollment link")
)

type EnrollmentService struct {
	enrollRepo *repository.EnrollmentRepository
	userRepo   *repository.UserRepository
	fleetRepo  *repository.FleetRepository
	appBaseURL string
}

func NewEnrollmentService(enrollRepo *repository.EnrollmentRepository, userRepo *repository.UserRepository, fleetRepo *repository.FleetRepository, appBaseURL string) *EnrollmentService {
	return &EnrollmentService{enrollRepo: enrollRepo, userRepo: userRepo, fleetRepo: fleetRepo, appBaseURL: strings.TrimRight(strings.TrimSpace(appBaseURL), "/")}
}

func (s *EnrollmentService) GenerateEnrollmentLink(ctx context.Context, actorUserID int64, expiresHours, maxUses int) (models.EnrollmentLink, string, error) {
	if err := s.ensureApprover(ctx, actorUserID); err != nil {
		return models.EnrollmentLink{}, "", err
	}
	if expiresHours <= 0 {
		expiresHours = 48
	}
	if maxUses <= 0 {
		maxUses = 1
	}
	token, err := randomToken(32)
	if err != nil {
		return models.EnrollmentLink{}, "", err
	}
	link, err := s.enrollRepo.CreateLink(ctx, models.EnrollmentLink{
		Token:           token,
		CreatedBy:       actorUserID,
		ExpiresAt:       time.Now().UTC().Add(time.Duration(expiresHours) * time.Hour),
		MaxUses:         maxUses,
		IsActive:        true,
		RequireApproval: true,
	})
	if err != nil {
		return models.EnrollmentLink{}, "", err
	}
	joinURL := fmt.Sprintf("%s/login?enroll_token=%s", s.appBaseURL, token)
	_, _ = s.enrollRepo.InsertNotification(ctx, models.SystemNotification{Type: "enrollment.link.created", Message: "New enrollment link generated", TargetRole: "admin"})
	return link, joinURL, nil
}

func (s *EnrollmentService) AcceptEnrollment(ctx context.Context, token string, payload models.DeviceEnrollment) (models.DeviceEnrollment, error) {
	link, err := s.enrollRepo.GetLinkByToken(ctx, strings.TrimSpace(token))
	if err != nil {
		return models.DeviceEnrollment{}, ErrInvalidLink
	}
	if !link.IsActive || link.UsedCount >= link.MaxUses || time.Now().UTC().After(link.ExpiresAt) {
		return models.DeviceEnrollment{}, ErrInvalidLink
	}

	payload.Status = "pending"
	payload.LinkID = link.ID
	payload.Hostname = strings.TrimSpace(payload.Hostname)
	payload.EmployeeName = strings.TrimSpace(payload.EmployeeName)
	payload.EmployeeEmail = strings.ToLower(strings.TrimSpace(payload.EmployeeEmail))
	if payload.Hostname == "" || payload.EmployeeName == "" || payload.EmployeeEmail == "" {
		return models.DeviceEnrollment{}, errors.New("hostname, employee_name, and employee_email are required")
	}

	enrollment, err := s.enrollRepo.CreateEnrollment(ctx, payload)
	if err != nil {
		return models.DeviceEnrollment{}, err
	}
	_ = s.enrollRepo.IncrementLinkUsage(ctx, link.ID)
	_, _ = s.enrollRepo.InsertNotification(ctx, models.SystemNotification{
		Type:       "enrollment.request.pending",
		Message:    fmt.Sprintf("Device %s requested enrollment approval", payload.Hostname),
		TargetRole: "admin",
	})
	return enrollment, nil
}

func (s *EnrollmentService) ListEnrollments(ctx context.Context, actorUserID int64, status string) ([]models.DeviceEnrollment, error) {
	if err := s.ensureApprover(ctx, actorUserID); err != nil {
		return nil, err
	}
	return s.enrollRepo.ListEnrollments(ctx, strings.TrimSpace(status))
}

func (s *EnrollmentService) ApproveEnrollment(ctx context.Context, actorUserID, enrollmentID int64, departmentID *int64) (models.DeviceEnrollment, error) {
	if err := s.ensureApprover(ctx, actorUserID); err != nil {
		return models.DeviceEnrollment{}, err
	}
	enrollment, err := s.enrollRepo.GetEnrollmentByID(ctx, enrollmentID)
	if err != nil {
		return models.DeviceEnrollment{}, err
	}
	if enrollment.Status != "pending" {
		return models.DeviceEnrollment{}, errors.New("only pending enrollments can be approved")
	}

	laptop, err := s.fleetRepo.CreateLaptop(ctx, models.EmployeeLaptop{
		Hostname:      enrollment.Hostname,
		EmployeeName:  enrollment.EmployeeName,
		EmployeeEmail: enrollment.EmployeeEmail,
		OSType:        enrollment.OSType,
		DepartmentID:  departmentID,
		IsActive:      true,
	})
	if err != nil {
		return models.DeviceEnrollment{}, err
	}

	if err := s.enrollRepo.ApproveEnrollment(ctx, enrollmentID, actorUserID, laptop.ID); err != nil {
		return models.DeviceEnrollment{}, err
	}
	updated, err := s.enrollRepo.GetEnrollmentByID(ctx, enrollmentID)
	if err != nil {
		return models.DeviceEnrollment{}, err
	}
	_, _ = s.enrollRepo.InsertNotification(ctx, models.SystemNotification{
		Type:       "enrollment.approved",
		Message:    fmt.Sprintf("Device %s approved and connected to policy control", updated.Hostname),
		TargetRole: "admin",
	})
	_ = s.enrollRepo.InsertAuditLog(ctx, actorUserID, "enrollment.approve", "device_enrollment", updated.ID, map[string]any{"hostname": updated.Hostname})
	return updated, nil
}

func (s *EnrollmentService) DisableEnrollment(ctx context.Context, actorUserID, enrollmentID int64) error {
	if err := s.ensureApprover(ctx, actorUserID); err != nil {
		return err
	}
	if err := s.enrollRepo.DisableEnrollment(ctx, enrollmentID, actorUserID); err != nil {
		return err
	}
	enrollment, _ := s.enrollRepo.GetEnrollmentByID(ctx, enrollmentID)
	_, _ = s.enrollRepo.InsertNotification(ctx, models.SystemNotification{
		Type:       "enrollment.disabled",
		Message:    fmt.Sprintf("Device %s access was disabled by approver", enrollment.Hostname),
		TargetRole: "admin",
	})
	_ = s.enrollRepo.InsertAuditLog(ctx, actorUserID, "enrollment.disable", "device_enrollment", enrollmentID, nil)
	return nil
}

func (s *EnrollmentService) ListNotifications(ctx context.Context, actorUserID int64) ([]models.SystemNotification, error) {
	if err := s.ensureApprover(ctx, actorUserID); err != nil {
		return nil, err
	}
	return s.enrollRepo.ListNotifications(ctx, 30)
}

func (s *EnrollmentService) ensureApprover(ctx context.Context, actorUserID int64) error {
	u, err := s.userRepo.GetByID(ctx, actorUserID)
	if err != nil {
		return ErrPermissionRequired
	}
	if !u.IsActive {
		return ErrPermissionRequired
	}
	if u.Role != "admin" && u.Role != "manager" {
		return ErrPermissionRequired
	}
	return nil
}

func randomToken(length int) (string, error) {
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
