package service

import (
	"context"
	"github/OfrenDialsa/go-gin-starter/internal/modules/auth/dto"
	"sync"
)

type AuthService interface {
	Register(ctx context.Context, userAgent, ipAddress string, req *dto.RegisterRequest) (res *dto.RegisterResponse, err error)
	VerifyEmail(ctx context.Context, token string) error
	ResendVerificationEmail(ctx context.Context, userAgent, ipAddress string, req *dto.ResendVerificationRequest) (err error)
	Login(ctx context.Context, req dto.LoginRequest) (res *dto.LoginResponse, err error)
	RefreshToken(ctx context.Context, refreshToken string) (res *dto.RefreshTokenResponse, err error)
	Logout(ctx context.Context, sessionId string) error
	ForgotPassword(ctx context.Context, email, userAgent, ipAddress string) error
	ResetPassword(ctx context.Context, token string, newPassword string) error
	CheckEmail(ctx context.Context, email string) (bool, error)
	CheckUsername(ctx context.Context, username string) (bool, error)
	SetWaitGroup(wg *sync.WaitGroup)
}
