package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github/OfrenDialsa/go-gin-starter/config"
	"github/OfrenDialsa/go-gin-starter/database"
	"github/OfrenDialsa/go-gin-starter/internal/infra/storage"
	auditDto "github/OfrenDialsa/go-gin-starter/internal/modules/audit_log/dto"
	"github/OfrenDialsa/go-gin-starter/internal/modules/audit_log/service"
	sessionSvc "github/OfrenDialsa/go-gin-starter/internal/modules/session/service"
	"github/OfrenDialsa/go-gin-starter/internal/modules/user/dto"
	"github/OfrenDialsa/go-gin-starter/internal/modules/user/model"
	"github/OfrenDialsa/go-gin-starter/internal/modules/user/repository"
	"github/OfrenDialsa/go-gin-starter/lib"
	"github/OfrenDialsa/go-gin-starter/utils"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

type userServiceImpl struct {
	env        *config.EnvironmentVariable
	userRepo   repository.UserRepository
	txManager  database.TxManager
	sessionSvc sessionSvc.SessionService
	auditSvc   service.AuditLogService
	storage    storage.StorageService
}

func NewUserService(env *config.EnvironmentVariable, txManager database.TxManager, userRepo repository.UserRepository, sessionSvc sessionSvc.SessionService,
	auditSvc service.AuditLogService, storage storage.StorageService) UserService {
	return &userServiceImpl{
		env:        env,
		txManager:  txManager,
		userRepo:   userRepo,
		sessionSvc: sessionSvc,
		auditSvc:   auditSvc,
		storage:    storage,
	}
}

func (u *userServiceImpl) Create(ctx context.Context, req dto.CreateUserRequest) error {
	username := strings.ToLower(req.Username)

	hashedPassword, err := lib.Hash(req.PasswordHash)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user := &model.User{
		UserId:          req.UserId,
		Email:           req.Email,
		Username:        username,
		PasswordHash:    &hashedPassword,
		Name:            req.Name,
		AvatarURL:       req.AvatarURL,
		Status:          "active",
		Role:            "user",
		EmailVerifiedAt: nil,
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (u *userServiceImpl) GetByEmail(ctx context.Context, email string) (*dto.GetUserResponse, error) {
	if email == "" {
		return nil, lib.ErrInvalidInput
	}
	sess, err := u.userRepo.GetUser(ctx, dto.GetUserFilter{Email: email})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, lib.ErrRecordNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if sess == nil {
		return nil, lib.ErrRecordNotFound
	}

	res := &dto.GetUserResponse{
		Id:           sess.Id,
		UserId:       sess.UserId,
		Email:        sess.Email,
		Username:     sess.Username,
		PasswordHash: sess.PasswordHash,
		Name:         sess.Name,
		AvatarURL:    sess.AvatarURL,
		Status:       sess.Status,
	}

	return res, nil
}

func (u *userServiceImpl) GetByEmailOrUsername(ctx context.Context, email string, username string) (*dto.GetUserResponse, error) {
	if email == "" || username == "" {
		return nil, lib.ErrInvalidInput
	}
	user, err := u.userRepo.GetUser(ctx, dto.GetUserFilter{Email: email, Username: username, OrMatch: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, lib.ErrRecordNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, lib.ErrRecordNotFound
	}

	res := &dto.GetUserResponse{
		Id:           user.Id,
		UserId:       user.UserId,
		Email:        user.Email,
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
		Name:         user.Name,
		AvatarURL:    user.AvatarURL,
		Status:       user.Status,
	}

	return res, nil
}

func (u *userServiceImpl) GetMe(ctx context.Context, userId string) (*dto.GetMeResponse, error) {
	filter := dto.GetUserFilter{
		UserId: userId,
	}

	user, err := u.userRepo.GetUser(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, lib.ErrUserNotFound
	}

	return &dto.GetMeResponse{
		User: dto.UserProfileData{
			UserId:          user.UserId,
			Email:           user.Email,
			Username:        user.Username,
			Name:            user.Name,
			AvatarURL:       user.AvatarURL,
			Status:          user.Status,
			EmailVerifiedAt: user.EmailVerifiedAt,
			LastLoginAt:     user.LastLoginAt,
			CreatedAt:       user.CreatedAt,
		},
	}, nil
}

func (u *userServiceImpl) UpdateProfile(ctx context.Context, req dto.UpdateProfileRequest) (*dto.UpdateProfileResponse, error) {
	user, err := u.userRepo.GetUser(ctx, dto.GetUserFilter{UserId: req.UserId})
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, lib.ErrUserNotFound
	}

	updateFields := dto.UpdateUserFilter{}
	if req.Username != "" && req.Username != user.Username {
		usernameExists, err := u.userRepo.IsExist(ctx, dto.GetUserFilter{Username: req.Username})
		if err != nil {
			return nil, fmt.Errorf("failed to check username availability: %w", err)
		}
		if usernameExists {
			return nil, lib.ErrUsernameNotAvailable
		}

		user.Username = req.Username
		updateFields.Username = &req.Username
	}

	if req.Name != "" && req.Name != user.Name {
		user.Name = req.Name
		updateFields.Name = &req.Name
	}

	var oldPath string
	var uploadedFile string

	if user.AvatarURL != nil {
		oldPath = *user.AvatarURL
	}

	err = u.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		if req.Avatar != nil {
			file, err := req.Avatar.Open()
			if err != nil {
				return fmt.Errorf("failed to open avatar file: %w", err)
			}
			defer file.Close()

			buf := new(bytes.Buffer)
			if _, err := io.Copy(buf, file); err != nil {
				return fmt.Errorf("error copying file to buffer: %w", err)
			}

			ext := filepath.Ext(req.Avatar.Filename)
			fileName := fmt.Sprintf("avatars/%s%s", utils.GenerateULID(), ext)

			if err := u.storage.UploadFile(txCtx, fileName, buf); err != nil {
				return fmt.Errorf("failed to upload avatar: %w", err)
			}

			uploadedFile = fileName

			publicURL, err := u.storage.GetPublicURL(fileName)
			if err != nil {
				return fmt.Errorf("failed to generate public url: %w", err)
			}

			user.AvatarURL = &publicURL
			updateFields.AvatarURL = &publicURL
		}

		if updateFields.Username != nil || updateFields.Name != nil || updateFields.AvatarURL != nil {
			if err := u.userRepo.Update(txCtx, req.UserId, updateFields); err != nil {
				return fmt.Errorf("failed to update user profile: %w", err)
			}
		}

		auditLog := auditDto.CreateRequest{
			UserID:       &req.UserId,
			Action:       "UPDATE_PROFILE",
			ResourceType: "users",
			ResourceID:   req.UserId,
			Details:      []byte(`{"message":"update profile"}`),
		}
		if err := u.auditSvc.Create(txCtx, auditLog); err != nil {
			return fmt.Errorf("failed to create audit log: %w", err)
		}

		return nil
	})

	if err != nil {
		if uploadedFile != "" {
			go func(path string) {
				if err := u.storage.DeleteFile(context.Background(), path); err != nil {
					log.Error().Err(err).Str("file_path", path).Msg("failed to delete uploaded file after rollback")
				}
			}(uploadedFile)
		}
		return nil, err
	}

	if uploadedFile != "" && oldPath != "" {
		go func(path string) {
			if err := u.storage.DeleteFile(context.Background(), path); err != nil {
				log.Error().Err(err).Str("file_path", path).Msg("failed to delete old avatar")
			}
		}(oldPath)
	}

	return &dto.UpdateProfileResponse{
		User: dto.UpdateProfileData{
			UserId:    user.UserId,
			Email:     user.Email,
			Username:  user.Username,
			Name:      user.Name,
			AvatarURL: user.AvatarURL,
		},
	}, nil
}

func (u *userServiceImpl) CheckUsernameExits(ctx context.Context, username string) (bool, error) {
	if username == "" {
		return false, lib.ErrInvalidInput
	}
	res, err := u.userRepo.IsExist(ctx, dto.GetUserFilter{Email: username})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, lib.ErrRecordNotFound
		}
		return false, fmt.Errorf("failed to check email exist: %w", err)
	}

	return res, nil
}

func (u *userServiceImpl) CheckEmailExits(ctx context.Context, email string) (bool, error) {
	if email == "" {
		return false, lib.ErrInvalidInput
	}
	res, err := u.userRepo.IsExist(ctx, dto.GetUserFilter{Email: email})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, lib.ErrRecordNotFound
		}
		return false, fmt.Errorf("failed to check email exist: %w", err)
	}

	return res, nil
}

func (u *userServiceImpl) UpdatePassword(ctx context.Context, userId string, password string) error {
	if userId == "" || password == "" {
		return lib.ErrInvalidInput
	}

	hashedPassword, err := lib.Hash(password)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	if err := u.userRepo.Update(ctx, userId, dto.UpdateUserFilter{PasswordHash: &hashedPassword}); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

func (u *userServiceImpl) GetByUserId(ctx context.Context, userId string) (*dto.GetUserResponse, error) {
	if userId == "" {
		return nil, lib.ErrInvalidInput
	}
	sess, err := u.userRepo.GetUser(ctx, dto.GetUserFilter{UserId: userId})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, lib.ErrRecordNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	res := &dto.GetUserResponse{
		Id:           sess.Id,
		UserId:       sess.UserId,
		Email:        sess.Email,
		Username:     sess.Username,
		PasswordHash: sess.PasswordHash,
		Name:         sess.Name,
		AvatarURL:    sess.AvatarURL,
		Status:       sess.Status,
	}

	return res, nil
}

func (u *userServiceImpl) MarkVerifiedEmail(ctx context.Context, userId string) error {
	if userId == "" {
		return lib.ErrInvalidInput
	}
	now := time.Now()

	if err := u.userRepo.Update(ctx, userId, dto.UpdateUserFilter{EmailVerifiedAt: &now}); err != nil {
		return fmt.Errorf("failed to mark verified email: %w", err)
	}

	return nil
}

func (u *userServiceImpl) MarkLastLogin(ctx context.Context, userId string) error {
	if userId == "" {
		return lib.ErrInvalidInput
	}
	now := time.Now()

	if err := u.userRepo.Update(ctx, userId, dto.UpdateUserFilter{LastLoginAt: &now}); err != nil {
		return fmt.Errorf("failed to mark last login: %w", err)
	}

	return nil
}

func (u *userServiceImpl) ChangePassword(ctx context.Context, req dto.ChangePasswordRequest) error {
	userIdFilter := dto.GetUserFilter{
		UserId: req.UserId,
	}
	user, err := u.userRepo.GetUser(ctx, userIdFilter)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return lib.ErrUserNotFound
	}

	if user.PasswordHash == nil {
		return lib.ErrUserHasNoPassword
	}

	if err := lib.Verify(*user.PasswordHash, req.Current); err != nil {
		return lib.ErrInvalidCurrentPassword
	}

	if !lib.IsValidPassword(req.New) {
		return lib.ErrWeakPassword
	}

	if err := lib.Verify(*user.PasswordHash, req.New); err == nil {
		return lib.ErrSamePassword
	}

	hashedPassword, err := lib.Hash(req.New)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	err = u.userRepo.Update(ctx, req.UserId, dto.UpdateUserFilter{PasswordHash: &hashedPassword})
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	if req.RevokeOtherSession != nil {
		if *req.RevokeOtherSession {
			err = u.sessionSvc.RevokeOtherSessions(ctx, req.UserId, req.SessionId)
			if err != nil {
				log.Warn().
					Err(err).
					Str("user_id", req.UserId).
					Msg("failed to revoke other sessions")
			}
		}
	}

	return nil
}

func (u *userServiceImpl) DeleteAccount(ctx context.Context, req dto.UserDeleteAccountRequest) error {
	userIdFilter := dto.GetUserFilter{
		UserId: req.UserId,
	}
	user, err := u.userRepo.GetUser(ctx, userIdFilter)

	if err != nil {
		return lib.ErrUserNotFound
	}

	if user.PasswordHash == nil || lib.Verify(*user.PasswordHash, req.Password) != nil {
		return lib.ErrInvalidPassword
	}

	var avatarPath string
	if user.AvatarURL != nil {
		avatarPath = *user.AvatarURL
	}

	err = u.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := u.userRepo.Delete(txCtx, req.UserId); err != nil {
			return fmt.Errorf("failed to delete user account: %w", err)
		}

		if err := u.sessionSvc.RevokeAllUserSessions(txCtx, req.UserId); err != nil {
			return fmt.Errorf("failed to revoke user sessions: %w", err)
		}
		auditLog := auditDto.CreateRequest{
			UserID:       &req.UserId,
			Action:       "DELETE_ACCOUNT",
			ResourceType: "users",
			ResourceID:   req.UserId,
			Details:      []byte(`{"message":"delete account"}`),
		}

		if err := u.auditSvc.Create(txCtx, auditLog); err != nil {
			return fmt.Errorf("failed to create audit log: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	if avatarPath != "" {
		go func(path string) {
			if err := u.storage.DeleteFile(context.Background(), path); err != nil {
				log.Error().
					Err(err).
					Str("file_path", path).
					Msg("failed to delete user avatar after account deletion")
			}
		}(avatarPath)
	}

	return nil
}

func (u *userServiceImpl) DeleteAvatar(ctx context.Context, userId string) error {
	userIdFilter := dto.GetUserFilter{
		UserId: userId,
	}

	user, err := u.userRepo.GetUser(ctx, userIdFilter)
	if err != nil || user == nil {
		return lib.ErrUserNotFound
	}

	if user.AvatarURL == nil || *user.AvatarURL == "" {
		return nil
	}

	oldPath := *user.AvatarURL

	err = u.userRepo.Update(ctx, userId, dto.UpdateUserFilter{AvatarURL: nil})
	if err != nil {
		return err
	}

	go func(path string) {
		u.storage.DeleteFile(context.Background(), path)
	}(oldPath)

	return nil
}
