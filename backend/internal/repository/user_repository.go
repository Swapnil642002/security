package repository

import (
	"context"
	"errors"
	"strings"

	"firewall-manager/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrSingleAdminExists = errors.New("single active admin already exists")
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, fullName, email, passwordHash, role string) (models.User, error) {
	const q = `
		INSERT INTO users (full_name, email, password_hash, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, full_name, email, password_hash, role, is_active, created_at, updated_at`

	var user models.User
	err := r.pool.QueryRow(ctx, q, fullName, strings.ToLower(strings.TrimSpace(email)), passwordHash, role).Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "uniq_single_active_admin" {
			return models.User{}, ErrSingleAdminExists
		}
	}
	return user, err
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (models.User, error) {
	const q = `
		SELECT id, full_name, email, password_hash, role, is_active, created_at, updated_at
		FROM users
		WHERE email = $1`

	var user models.User
	err := r.pool.QueryRow(ctx, q, strings.ToLower(strings.TrimSpace(email))).Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, ErrUserNotFound
	}
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (models.User, error) {
	const q = `
		SELECT id, full_name, email, password_hash, role, is_active, created_at, updated_at
		FROM users
		WHERE id = $1`

	var user models.User
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, ErrUserNotFound
	}
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (r *UserRepository) CountAdmins(ctx context.Context) (int, error) {
	const q = `SELECT COUNT(*) FROM users WHERE role = 'admin' AND is_active = TRUE`
	var count int
	err := r.pool.QueryRow(ctx, q).Scan(&count)
	return count, err
}
