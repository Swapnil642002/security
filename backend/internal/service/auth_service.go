package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"firewall-manager/internal/auth"
	"firewall-manager/internal/models"
	"firewall-manager/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInactiveUser       = errors.New("user is inactive")
	ErrAdminExists        = errors.New("admin already exists")
	ErrWeakPassword       = errors.New("password must be at least 10 characters")
)

type LoginResult struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      UserSafe  `json:"user"`
}

type UserSafe struct {
	ID       int64  `json:"id"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	IsActive bool   `json:"is_active"`
}

type AuthService struct {
	users *repository.UserRepository
	jwt   *auth.JWTManager
}

func NewAuthService(users *repository.UserRepository, jwt *auth.JWTManager) *AuthService {
	return &AuthService{users: users, jwt: jwt}
}

func (s *AuthService) IsBootstrapAllowed(ctx context.Context) (bool, error) {
	count, err := s.users.CountAdmins(ctx)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func (s *AuthService) BootstrapAdmin(ctx context.Context, fullName, email, password string) (UserSafe, error) {
	count, err := s.users.CountAdmins(ctx)
	if err != nil {
		return UserSafe{}, err
	}
	if count > 0 {
		return UserSafe{}, ErrAdminExists
	}

	if len(strings.TrimSpace(password)) < 10 {
		return UserSafe{}, ErrWeakPassword
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return UserSafe{}, err
	}

	u, err := s.users.Create(ctx, strings.TrimSpace(fullName), strings.TrimSpace(email), hash, "admin")
	if err != nil {
		if errors.Is(err, repository.ErrSingleAdminExists) {
			return UserSafe{}, ErrAdminExists
		}
		return UserSafe{}, err
	}
	return toSafeUser(u), nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (LoginResult, error) {
	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return LoginResult{}, ErrInvalidCredentials
		}
		return LoginResult{}, err
	}

	if !u.IsActive {
		return LoginResult{}, ErrInactiveUser
	}

	if !auth.VerifyPassword(password, u.PasswordHash) {
		return LoginResult{}, ErrInvalidCredentials
	}

	token, exp, err := s.jwt.Generate(u.ID, u.Role)
	if err != nil {
		return LoginResult{}, err
	}

	return LoginResult{
		Token:     token,
		ExpiresAt: exp,
		User:      toSafeUser(u),
	}, nil
}

func (s *AuthService) GetCurrentUser(ctx context.Context, userID int64) (UserSafe, error) {
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return UserSafe{}, err
	}
	return toSafeUser(u), nil
}

func toSafeUser(u models.User) UserSafe {
	return UserSafe{
		ID:       u.ID,
		FullName: u.FullName,
		Email:    u.Email,
		Role:     u.Role,
		IsActive: u.IsActive,
	}
}
