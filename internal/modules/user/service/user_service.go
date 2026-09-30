package service

import (
	"context"
	"github/OfrenDialsa/go-gin-starter/internal/modules/user/dto"
)

type UserService interface {
	Create(ctx context.Context, req dto.CreateUserRequest) error
	GetMe(ctx context.Context, userId string) (*dto.GetMeResponse, error)
	GetByUserId(ctx context.Context, userId string) (*dto.GetUserResponse, error)
	GetByEmail(ctx context.Context, email string) (*dto.GetUserResponse, error)
	GetByEmailOrUsername(ctx context.Context, email, username string) (*dto.GetUserResponse, error)
	MarkLastLogin(ctx context.Context, userId string) error
	MarkVerifiedEmail(ctx context.Context, userId string) error
	ChangePassword(ctx context.Context, req dto.ChangePasswordRequest) error
	UpdateProfile(ctx context.Context, req dto.UpdateProfileRequest) (*dto.UpdateProfileResponse, error)
	UpdatePassword(ctx context.Context, userId, password string) error
	DeleteAccount(ctx context.Context, req dto.UserDeleteAccountRequest) error
	DeleteAvatar(ctx context.Context, userId string) error
	CheckEmailExits(ctx context.Context, email string) (bool, error)
	CheckUsernameExits(ctx context.Context, username string) (bool, error)
}
