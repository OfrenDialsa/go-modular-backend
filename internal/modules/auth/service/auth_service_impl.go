package service

import (
	"context"
	"errors"
	"fmt"
	"github/OfrenDialsa/go-gin-starter/config"
	"github/OfrenDialsa/go-gin-starter/database"
	mailerDto "github/OfrenDialsa/go-gin-starter/internal/dto"
	"github/OfrenDialsa/go-gin-starter/internal/infra/mailer"
	"github/OfrenDialsa/go-gin-starter/internal/infra/metrics"
	"github/OfrenDialsa/go-gin-starter/internal/modules/auth/dto"
	sessDto "github/OfrenDialsa/go-gin-starter/internal/modules/session/dto"
	"github/OfrenDialsa/go-gin-starter/internal/modules/session/service"
	userDto "github/OfrenDialsa/go-gin-starter/internal/modules/user/dto"
	userSvc "github/OfrenDialsa/go-gin-starter/internal/modules/user/service"
	"github/OfrenDialsa/go-gin-starter/lib"
	"github/OfrenDialsa/go-gin-starter/utils"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type authServiceImpl struct {
	env        *config.EnvironmentVariable
	mailer     mailer.MailerService
	txManager  database.TxManager
	userSvc    userSvc.UserService
	sessionSvc service.SessionService

	wg *sync.WaitGroup
}

func NewAuthService(
	env *config.EnvironmentVariable,
	txManager database.TxManager,
	mailer mailer.MailerService,
	userSvc userSvc.UserService,
	sessionSvc service.SessionService,
	wg *sync.WaitGroup,
) AuthService {
	return &authServiceImpl{
		env:        env,
		txManager:  txManager,
		mailer:     mailer,
		userSvc:    userSvc,
		sessionSvc: sessionSvc,
		wg:         wg,
	}
}

func (s *authServiceImpl) Register(ctx context.Context, userAgent, ipAddress string, req *dto.RegisterRequest) (res *dto.RegisterResponse, err error) {
	start := time.Now()
	defer func() {
		status := "success"
		if err != nil {
			status = "failed"
		}
		metrics.TrackAuth("register", status, time.Since(start))
	}()

	existingUser, err := s.userSvc.GetByEmailOrUsername(ctx, req.Email, req.Username)
	if err != nil {
		if !errors.Is(err, lib.ErrRecordNotFound) {
			return nil, fmt.Errorf("failed to check existing user: %w", err)
		}
	}
	if existingUser != nil {
		if existingUser.Email == req.Email {
			return nil, lib.ErrEmailAlreadyExists
		}
		if existingUser.Username == req.Username {
			return nil, lib.ErrUsernameNotAvailable
		}
	}

	userReq := userDto.CreateUserRequest{
		UserId:       utils.GenerateULID(),
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: req.Password,
		Name:         req.Name,
		AvatarURL:    nil,
	}

	verifToken, err := utils.GenerateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate verification token: %w", err)
	}

	expiresAt := time.Now().Add(time.Hour)
	sessionReq := sessDto.CreateSession{
		SessionId: utils.GenerateULID(),
		UserId:    userReq.UserId,
		TokenHash: verifToken,
		ExpiresAt: expiresAt,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Type:      "verify_email",
	}

	err = s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := s.userSvc.Create(txCtx, userReq); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		if err := s.sessionSvc.CreateSession(txCtx, sessionReq); err != nil {
			return fmt.Errorf("failed to create verify email session: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	userName := userReq.Name
	userEmail := userReq.Email

	if s.wg != nil {
		s.wg.Add(1)
	}

	go func() {
		defer func() {
			if s.wg != nil {
				s.wg.Done()
			}

			if r := recover(); r != nil {
				log.Error().Interface("panic", r).Msg("recovered from panic in verify user email goroutine")
			}
		}()

		verifLink := fmt.Sprintf("%s?token=%s", s.env.External.VerifyEmailURL, verifToken)
		emailBody, err := lib.BuildEmailBodyRegister(userName, verifLink)
		if err != nil {
			log.Error().Err(err).Msg("[x] failed to build email body in background")
			return
		}

		mailData := mailerDto.MailerRequest{
			To:          []string{userEmail},
			Subject:     lib.DefaultEmailSubjectRegister,
			Body:        emailBody,
			Attachments: []string{},
		}

		_, err = s.mailer.Send(mailData)
		if err != nil {
			log.Error().Err(err).Msg("[x] failed to send verification email in background")
			return
		}
		log.Info().Str("email", userEmail).Msg("[v] verification email sent successfully")
	}()

	return &dto.RegisterResponse{
		User: dto.UserData{
			UserId:          userReq.UserId,
			Email:           userReq.Email,
			Username:        userReq.Username,
			Name:            userReq.Name,
			Status:          userReq.Status,
			AvatarURL:       userReq.AvatarURL,
			EmailVerifiedAt: userReq.EmailVerifiedAt,
		},
	}, nil
}

func (s *authServiceImpl) Login(ctx context.Context, req dto.LoginRequest) (res *dto.LoginResponse, err error) {
	start := time.Now()
	defer func() {
		status := "success"
		if err != nil {
			status = "failed"
		}
		metrics.TrackAuth("login", status, time.Since(start))
	}()

	user, err := s.userSvc.GetByEmailOrUsername(ctx, req.Identifier, req.Identifier)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil || user.PasswordHash == nil {
		return nil, lib.ErrInvalidCredential
	}

	if user.Status != "active" {
		return nil, lib.ErrAccountInactive
	}

	if err := lib.Verify(*user.PasswordHash, req.Password); err != nil {
		return nil, lib.ErrInvalidCredential
	}

	now := time.Now()
	sessionId := utils.GenerateULID()

	accessExpiry := time.Duration(s.env.JWT.Token.AccessLifeTime)
	refreshExpiry := 7 * 24 * time.Hour

	tokens, err := lib.GenerateTokenPair(
		user.UserId,
		sessionId,
		user.Role,
		s.env.JWT.SecretKey.Access,
		s.env.JWT.SecretKey.Refresh,
		accessExpiry,
		refreshExpiry,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	session := sessDto.CreateSession{
		SessionId: sessionId,
		UserId:    user.UserId,
		TokenHash: tokens.RefreshToken,
		Type:      "refresh_token",
		IPAddress: req.IPAddress,
		UserAgent: req.UserAgent,
		ExpiresAt: now.Add(refreshExpiry),
	}

	err = s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := s.sessionSvc.CreateSession(txCtx, session); err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}

		if err := s.userSvc.MarkLastLogin(txCtx, user.UserId); err != nil {
			return fmt.Errorf("failed to update last login: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
		TokenType:    "Bearer",
		User: dto.UserData{
			UserId:          user.UserId,
			Email:           user.Email,
			Username:        user.Username,
			Name:            user.Name,
			Status:          user.Status,
			AvatarURL:       user.AvatarURL,
			EmailVerifiedAt: user.EmailVerifiedAt,
		},
	}, nil
}

func (s *authServiceImpl) Logout(ctx context.Context, sessionID string) error {
	err := s.sessionSvc.RevokeSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to revoke session during logout: %w", err)
	}
	return nil
}

func (s *authServiceImpl) ForgotPassword(ctx context.Context, email, userAgent, ipAddress string) error {
	user, err := s.userSvc.GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to get user by email: %w", err)
	}

	if user == nil {
		return nil
	}

	resetToken, err := utils.GenerateToken()
	if err != nil {
		return fmt.Errorf("failed to generate reset token: %w", err)
	}

	hashedToken := utils.HashTokenSHA256(resetToken)
	expiresAt := time.Now().Add(time.Hour)

	session := sessDto.CreateSession{
		SessionId: utils.GenerateULID(),
		UserId:    user.UserId,
		TokenHash: hashedToken,
		ExpiresAt: expiresAt,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Type:      "reset_password",
	}

	err = s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := s.sessionSvc.CreateSession(txCtx, session); err != nil {
			return fmt.Errorf("failed to create reset password session: %w", err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	userEmail := user.Email
	userName := user.Name

	if s.wg != nil {
		s.wg.Add(1)
	}

	go func() {
		defer func() {
			if s.wg != nil {
				s.wg.Done()
			}

			if r := recover(); r != nil {
				log.Error().Interface("panic", r).Msg("recovered from panic in forgot password email goroutine")
			}
		}()

		resetLink := fmt.Sprintf("%s?token=%s", s.env.External.ResetPasswordURL, resetToken)
		emailBody, err := lib.BuildEmailBodyResetPassword(userName, resetLink)
		if err != nil {
			log.Error().Err(err).Msg("[x] failed to build email body in background")
			return
		}

		mailData := mailerDto.MailerRequest{
			To:          []string{userEmail},
			Subject:     lib.DefaultEmailSubjectResetPassword,
			Body:        emailBody,
			Attachments: []string{},
		}

		_, err = s.mailer.Send(mailData)
		if err != nil {
			log.Error().Err(err).Str("email", userEmail).Msg("[x] error in sending forgot password email via goroutine")
			return
		}

		log.Info().Str("email", userEmail).Msg("[v] forgot password email sent successfully")
	}()

	return nil
}

func (s *authServiceImpl) RefreshToken(ctx context.Context, refreshToken string) (res *dto.RefreshTokenResponse, err error) {
	start := time.Now()
	defer func() {
		status := "success"
		if err != nil {
			status = "failed"
		}
		metrics.TrackAuth("refresh_token", status, time.Since(start))
	}()

	claims, err := lib.ValidateToken(refreshToken, s.env.JWT.SecretKey.Refresh)
	if err != nil {
		log.Warn().Err(err).Msg("failed to validate refresh token")
		return nil, lib.ErrInvalidToken
	}

	session, err := s.sessionSvc.GetBySessionId(ctx, claims.SessionId)
	if err != nil {
		log.Error().Err(err).Str("session_id", claims.SessionId).Msg("database error when fetching session")
		return nil, fmt.Errorf("failed to process session")
	}
	if session == nil {
		log.Warn().Str("session_id", claims.SessionId).Msg("refresh token session not found")
		return nil, lib.ErrUnauthorized
	}

	// 1. Deteksi Replay Attack
	if session.TokenHash != refreshToken {
		log.Error().
			Str("session_id", session.SessionId).
			Str("user_id", claims.UserId).
			Msg("REPLAY ATTACK DETECTED: refresh token already used, revoking session")

		// Cabut session dalam transaksi untuk memastikan perubahan status tersimpan
		_ = s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
			return s.sessionSvc.RevokeSession(txCtx, session.SessionId)
		})
		return nil, lib.ErrUnauthorized
	}

	if session.RevokedAt != nil || time.Now().After(session.ExpiresAt) {
		log.Info().Str("session_id", session.SessionId).Msg("attempted to use a revoked or expired session")
		return nil, lib.ErrUnauthorized
	}

	// 2. Buat Token Pair Baru
	accessExpiry := time.Duration(s.env.JWT.Token.AccessLifeTime)
	refreshExpiry := 7 * 24 * time.Hour

	tokens, err := lib.GenerateTokenPair(
		claims.UserId,
		claims.SessionId,
		claims.Role,
		s.env.JWT.SecretKey.Access,
		s.env.JWT.SecretKey.Refresh,
		accessExpiry,
		refreshExpiry,
	)
	if err != nil {
		log.Error().Err(err).Str("user_id", claims.UserId).Msg("failed to generate token pair")
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	now := time.Now()
	updatedSession := sessDto.UpdateSessionRequest{
		TokenHash: tokens.RefreshToken,
		ExpiresAt: now.Add(refreshExpiry),
	}

	err = s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := s.sessionSvc.Update(txCtx, updatedSession); err != nil {
			log.Error().Err(err).Str("session_id", session.SessionId).Msg("failed to update session in database")
			return fmt.Errorf("failed to update session: %w", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &dto.RefreshTokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
		TokenType:    "Bearer",
	}, nil
}

func (s *authServiceImpl) ResetPassword(ctx context.Context, token string, newPassword string) error {
	if len(newPassword) < 8 {
		return lib.ErrWeakPassword
	}

	hashedToken := utils.HashTokenSHA256(token)

	log.Debug().Str("token_hash", hashedToken).Msg("attempting to reset password")
	session, err := s.sessionSvc.GetByToken(ctx, hashedToken, "reset_password")
	if err != nil {
		return fmt.Errorf("failed to retrieve session: %w", err)
	}

	if session == nil || session.RevokedAt != nil || session.ExpiresAt.Before(time.Now()) {
		return lib.ErrInvalidToken
	}

	user, err := s.userSvc.GetByUserId(ctx, session.UserId)
	if err != nil {
		return fmt.Errorf("failed to retrieve user: %w", err)
	}

	err = s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := s.userSvc.UpdatePassword(txCtx, session.UserId, newPassword); err != nil {
			return fmt.Errorf("failed to update password: %w", err)
		}

		if err := s.sessionSvc.RevokeAllUserSessions(txCtx, session.UserId); err != nil {
			return fmt.Errorf("failed to revoke user sessions: %w", err)
		}

		if err := s.sessionSvc.DeleteSession(txCtx, session.SessionId); err != nil {
			return fmt.Errorf("failed to delete reset session: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	userName := user.Name
	userEmail := user.Email

	if s.wg != nil {
		s.wg.Add(1)
	}

	go func() {
		defer func() {
			if s.wg != nil {
				s.wg.Done()
			}

			if r := recover(); r != nil {
				log.Error().Interface("panic", r).Msg("recovered from panic in reset password email goroutine")
			}
		}()

		emailBody, err := lib.BuildEmailBodyPasswordResetSuccess(userName)
		if err != nil {
			log.Error().Err(err).Msg("[x] failed to build email body in background")
			return
		}

		mailData := mailerDto.MailerRequest{
			To:          []string{userEmail},
			Subject:     lib.DefaultEmailSubjectPasswordResetSuccess,
			Body:        emailBody,
			Attachments: []string{},
		}

		_, err = s.mailer.Send(mailData)
		if err != nil {
			log.Error().Err(err).Msg("[x] failed to send reset password email in background")
			return
		}

		log.Info().Str("email", userEmail).Msg("[v] password reset success email sent successfully")
	}()

	log.Info().
		Str("user_id", session.UserId).
		Msg("password reset successfully and sessions cleaned up")

	return nil
}

func (s *authServiceImpl) ResendVerificationEmail(ctx context.Context, userAgent, ipAddress string, req *dto.ResendVerificationRequest) (err error) {
	start := time.Now()
	defer func() {
		status := "success"
		if err != nil {
			status = "failed"
		}
		metrics.TrackAuth("resend_verification", status, time.Since(start))
	}()

	user, err := s.userSvc.GetByEmail(ctx, req.Email)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return lib.ErrUserNotFound
	}

	if user.EmailVerifiedAt != nil {
		return lib.ErrEmailAlreadyVerified
	}

	verifToken, err := utils.GenerateToken()
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}
	hashedToken := utils.HashTokenSHA256(verifToken)
	expiresAt := time.Now().Add(time.Hour)

	session := sessDto.CreateSession{
		SessionId: utils.GenerateULID(),
		UserId:    user.UserId,
		TokenHash: hashedToken,
		ExpiresAt: expiresAt,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Type:      "verify_email",
	}

	err = s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := s.sessionSvc.DeleteByType(txCtx, user.UserId, "verify_email"); err != nil {
			return fmt.Errorf("failed to cleanup old verification tokens: %w", err)
		}

		if err := s.sessionSvc.CreateSession(txCtx, session); err != nil {
			return fmt.Errorf("failed to create verification session: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	userName := user.Name
	userEmail := user.Email

	if s.wg != nil {
		s.wg.Add(1)
	}

	go func() {
		defer func() {
			if s.wg != nil {
				s.wg.Done()
			}

			if r := recover(); r != nil {
				log.Error().Interface("panic", r).Msg("recovered from panic in resend verification email goroutine")
			}
		}()

		verifLink := fmt.Sprintf("%s?token=%s", s.env.External.VerifyEmailURL, verifToken)
		emailBody, err := lib.BuildEmailBodyResendVerification(userName, verifLink)
		if err != nil {
			log.Error().Err(err).Msg("[x] failed to build email body in background")
			return
		}

		mailData := mailerDto.MailerRequest{
			To:          []string{userEmail},
			Subject:     lib.DefaultEmailSubjectResend,
			Body:        emailBody,
			Attachments: []string{},
		}

		_, err = s.mailer.Send(mailData)
		if err != nil {
			log.Error().Err(err).Msg("[x] failed to resend verification email in background")
			return
		}

		log.Info().Str("email", userEmail).Msg("[v] resend verification email sent successfully")
	}()

	return nil
}

func (s *authServiceImpl) VerifyEmail(ctx context.Context, token string) error {
	hashedToken := utils.HashTokenSHA256(token)

	session, err := s.sessionSvc.ValidateToken(ctx, hashedToken, "verify_email")
	if err != nil {
		return fmt.Errorf("failed to retrieve session: %w", err)
	}

	if session == nil || session.RevokedAt != nil || session.ExpiresAt.Before(time.Now()) {
		return lib.ErrInvalidToken
	}

	user, err := s.userSvc.GetByUserId(ctx, session.UserId)
	if err != nil {
		return fmt.Errorf("failed to retrieve user: %w", err)
	}

	err = s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := s.userSvc.MarkVerifiedEmail(txCtx, session.UserId); err != nil {
			return fmt.Errorf("failed to update verified email: %w", err)
		}

		if err := s.sessionSvc.DeleteSession(txCtx, session.SessionId); err != nil {
			return fmt.Errorf("failed to cleanup session: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	userName := user.Name
	userEmail := user.Email

	if s.wg != nil {
		s.wg.Add(1)
	}

	go func() {
		defer func() {
			if s.wg != nil {
				s.wg.Done()
			}

			if r := recover(); r != nil {
				log.Error().Interface("panic", r).Msg("recovered from panic in verify email success goroutine")
			}
		}()

		loginURL := fmt.Sprintf("%s/login", s.env.External.FrontendURL)
		emailBody, err := lib.BuildEmailBodyVerifyEmailSuccess(userName, loginURL)
		if err != nil {
			log.Error().Err(err).Msg("[x] failed to build email body in background")
			return
		}

		mailData := mailerDto.MailerRequest{
			To:          []string{userEmail},
			Subject:     lib.DefaultEmailSubjectVerifyEmailSuccess,
			Body:        emailBody,
			Attachments: []string{},
		}

		_, err = s.mailer.Send(mailData)
		if err != nil {
			log.Error().Err(err).Msg("[x] failed to send verify email success background")
			return
		}

		log.Info().Str("email", userEmail).Msg("[v] verify email success notification sent successfully")
	}()

	return nil
}

func (s *authServiceImpl) CheckEmail(ctx context.Context, email string) (bool, error) {
	exists, err := s.userSvc.CheckEmailExits(ctx, email)
	if err != nil {
		log.Error().Err(err).Str("email", email).Msg("failed to check email availability")
		return false, fmt.Errorf("failed to check email: %w", err)
	}
	return exists, nil
}

func (s *authServiceImpl) CheckUsername(ctx context.Context, username string) (bool, error) {
	exists, err := s.userSvc.CheckUsernameExits(ctx, username)
	if err != nil {
		log.Error().Err(err).Str("username", username).Msg("failed to check username availability")
		return false, fmt.Errorf("failed to check username: %w", err)
	}
	return exists, nil
}

func (s *authServiceImpl) SetWaitGroup(wg *sync.WaitGroup) {
	s.wg = wg
}
