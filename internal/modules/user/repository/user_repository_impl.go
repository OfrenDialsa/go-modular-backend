package repository

import (
	"context"
	"errors"
	"fmt"
	"github/OfrenDialsa/go-gin-starter/database"
	"strings"

	"github/OfrenDialsa/go-gin-starter/internal/modules/user/dto"
	"github/OfrenDialsa/go-gin-starter/internal/modules/user/model"

	"github.com/jackc/pgx/v5"
)

type userRepositoryImpl struct {
	db *database.WrapDB
}

func NewUserRepository(db *database.WrapDB) UserRepository {
	return &userRepositoryImpl{db: db}
}

func (r *userRepositoryImpl) GetUser(ctx context.Context, filter dto.GetUserFilter) (*model.User, error) {
	query := `
		SELECT id, user_id, email, username, password_hash, name, avatar_url, status,
		       role, email_verified_at, last_login_at, created_at, updated_at, deleted_at
		FROM users
		WHERE deleted_at IS NULL
	`

	args := []any{}
	argIdx := 1
	var conditions []string

	if filter.UserId != "" {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIdx))
		args = append(args, filter.UserId)
		argIdx++
	}

	if filter.Username != "" {
		conditions = append(conditions, fmt.Sprintf("username = $%d", argIdx))
		args = append(args, filter.Username)
		argIdx++
	}

	if filter.Email != "" {
		conditions = append(conditions, fmt.Sprintf("email = $%d", argIdx))
		args = append(args, filter.Email)
		argIdx++
	}

	if len(conditions) > 0 {
		operator := " AND "
		if filter.OrMatch {
			operator = " OR "
		}

		query += " AND (" + strings.Join(conditions, operator) + ")"
	}

	query += " LIMIT 1"

	executor := database.GetExecutor(ctx, r.db.Database.Conn)
	var u model.User
	err := executor.QueryRow(ctx, query, args...).Scan(
		&u.Id, &u.UserId, &u.Email, &u.Username, &u.PasswordHash,
		&u.Name, &u.AvatarURL, &u.Status, &u.Role, &u.EmailVerifiedAt,
		&u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &u, nil
}

func (r *userRepositoryImpl) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (user_id, email, username, password_hash, name, avatar_url, role, status, email_verified_at, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id
	`
	executor := database.GetExecutor(ctx, r.db.Database.Conn)

	err := executor.QueryRow(ctx, query,
		user.UserId, user.Email, user.Username,
		user.PasswordHash, user.Name, user.AvatarURL,
		user.Role, user.Status, user.EmailVerifiedAt,
		user.CreatedAt, user.UpdatedAt,
	).Scan(&user.Id)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *userRepositoryImpl) Update(ctx context.Context, userId string, fields dto.UpdateUserFilter) error {
	query := "UPDATE users SET updated_at = NOW()"
	args := []any{}
	argIdx := 1

	if fields.Username != nil {
		query += fmt.Sprintf(", username = $%d", argIdx)
		args = append(args, *fields.Username)
		argIdx++
	}
	if fields.Name != nil {
		query += fmt.Sprintf(", name = $%d", argIdx)
		args = append(args, *fields.Name)
		argIdx++
	}
	if fields.PasswordHash != nil {
		query += fmt.Sprintf(", password_hash = $%d", argIdx)
		args = append(args, *fields.PasswordHash)
		argIdx++
	}
	if fields.AvatarURL != nil {
		query += fmt.Sprintf(", avatar_url = $%d", argIdx)
		args = append(args, fields.AvatarURL)
		argIdx++
	}
	if fields.EmailVerifiedAt != nil {
		query += fmt.Sprintf(", email_verified_at = $%d", argIdx)
		args = append(args, *fields.EmailVerifiedAt)
		argIdx++
	}
	if fields.LastLoginAt != nil {
		query += fmt.Sprintf(", last_login_at = $%d", argIdx)
		args = append(args, *fields.LastLoginAt)
		argIdx++
	}

	query += fmt.Sprintf(" WHERE user_id = $%d AND deleted_at IS NULL", argIdx)
	args = append(args, userId)

	executor := database.GetExecutor(ctx, r.db.Database.Conn)
	_, err := executor.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (r *userRepositoryImpl) Delete(ctx context.Context, userId string) error {
	query := `
        UPDATE users 
        SET deleted_at = NOW(), updated_at = NOW() 
        WHERE user_id = $1 AND deleted_at IS NULL
    `
	var err error
	executor := database.GetExecutor(ctx, r.db.Database.Conn)
	_, err = executor.Exec(ctx, query, userId)

	if err != nil {
		return fmt.Errorf("failed to soft delete user: %w", err)
	}

	return nil
}

func (r *userRepositoryImpl) IsExist(ctx context.Context, filter dto.GetUserFilter) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE deleted_at IS NULL"
	args := []any{}
	argIdx := 1

	if filter.UserId != "" {
		query += fmt.Sprintf(" AND user_id = $%d", argIdx)
		args = append(args, filter.UserId)
		argIdx++
	}
	if filter.Email != "" {
		query += fmt.Sprintf(" AND email = $%d", argIdx)
		args = append(args, filter.Email)
		argIdx++
	}
	if filter.Username != "" {
		query += fmt.Sprintf(" AND username = $%d", argIdx)
		args = append(args, filter.Username)
		argIdx++
	}

	query += ")"

	executor := database.GetExecutor(ctx, r.db.Database.Conn)
	var exists bool
	err := executor.QueryRow(ctx, query, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return exists, nil
}
