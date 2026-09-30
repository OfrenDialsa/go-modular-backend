package handler

import (
	"github/OfrenDialsa/go-gin-starter/config"
	"github/OfrenDialsa/go-gin-starter/internal/modules/auth/dto"
	"github/OfrenDialsa/go-gin-starter/internal/modules/auth/service"
	"github/OfrenDialsa/go-gin-starter/lib"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

type AuthHandlerImpl struct {
	env         *config.EnvironmentVariable
	authService service.AuthService
	validator   *validator.Validate
}

func NewAuthHandler(env *config.EnvironmentVariable, authService service.AuthService) AuthHandler {
	return &AuthHandlerImpl{
		env:         env,
		authService: authService,
		validator:   validator.New(),
	}
}

// Register handles user registration
// @Summary Register a new user
// @Description Register a new user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Register Request"
// @Success 201 {object} lib.APIResponse{data=dto.RegisterResponse}
// @Failure 400 {object} lib.HTTPError
// @Failure 500 {object} lib.HTTPError
// @Router /api/v1/auth/register [post]
func (h *AuthHandlerImpl) Register(ctx *gin.Context) {
	var req dto.RegisterRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		lib.RespondValidationError(ctx, http.StatusBadRequest, "Bad payload", lib.ParseValidationErrors(err))
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		lib.RespondValidationError(ctx, http.StatusBadRequest, "Validation failed", lib.ParseValidationErrors(err))
		return
	}

	ia := ctx.ClientIP()
	ua := ctx.Request.UserAgent()

	resp, err := h.authService.Register(ctx.Request.Context(), ua, ia, &req)
	if err != nil {
		if appErr, ok := err.(*lib.AppError); ok {
			lib.RespondError(ctx, appErr)
			return
		}

		lib.RespondError(ctx, lib.ErrInternalServer)
		return
	}

	lib.RespondSuccess(ctx, http.StatusCreated, lib.MsgRegistrationSuccess, resp)
}

// VerifyEmail godoc
// @Summary      Verify user email
// @Description  Verifies the user's email address using a token sent via email.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        token    query     string  true  "Verification token"
// @Success      200      {object}  lib.APIResponse "Email verified successfully"
// @Failure      400      {object}  lib.HTTPError "Invalid or expired token"
// @Failure      500      {object}  lib.HTTPError "Internal server error"
// @Router       /api/v1/auth/verify-email [get]
func (h *AuthHandlerImpl) VerifyEmail(ctx *gin.Context) {
	token := ctx.Query("token")

	if token == "" {
		lib.RespondError(ctx, lib.ErrInvalidToken)
		return
	}

	err := h.authService.VerifyEmail(ctx.Request.Context(), token)
	if err != nil {
		if appErr, ok := err.(*lib.AppError); ok {
			lib.RespondError(ctx, appErr)
			return
		}
		lib.RespondError(ctx, lib.ErrInternalServer)
		return
	}

	lib.RespondSuccess(ctx, http.StatusOK, "Email verified successfully", nil)
}

// ResendVerification handles resending the verification email
// @Summary Resend verification email
// @Description Resend verification email to user
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.ResendVerificationRequest true "Resend Verification Request"
// @Success 200 {object} lib.APIResponse
// @Failure 400 {object} lib.HTTPError
// @Failure 429 {object} lib.HTTPError
// @Failure 500 {object} lib.HTTPError
// @Router /api/v1/auth/resend-verification [post]
func (h *AuthHandlerImpl) ResendVerification(ctx *gin.Context) {
	var req dto.ResendVerificationRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		lib.RespondValidationError(ctx, http.StatusBadRequest, "Invalid request payload", lib.ParseValidationErrors(err))
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		lib.RespondValidationError(ctx, http.StatusBadRequest, "Validation failed", lib.ParseValidationErrors(err))
		return
	}

	ia := ctx.ClientIP()
	ua := ctx.Request.UserAgent()

	err := h.authService.ResendVerificationEmail(ctx.Request.Context(), ua, ia, &req)
	if err != nil {
		if appErr, ok := err.(*lib.AppError); ok {
			lib.RespondError(ctx, appErr)
			return
		}

		log.Error().Err(err).Msg("resend verification failed")
		lib.RespondError(ctx, lib.ErrInternalServer)
		return
	}

	// 4. Respond Success
	lib.RespondSuccess(ctx, http.StatusOK, "Verification email has been resent successfully", nil)
}

// Login handles user login
// @Summary Login user
// @Description Login user using email or username
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login Request"
// @Success 200 {object} lib.APIResponse{data=dto.LoginResponse}
// @Failure 400 {object} lib.HTTPError
// @Failure 401 {object} lib.HTTPError
// @Failure 500 {object} lib.HTTPError
// @Router /api/v1/auth/login [post]
func (h *AuthHandlerImpl) Login(ctx *gin.Context) {
	var req dto.LoginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		lib.RespondValidationError(ctx, http.StatusBadRequest, "Bad payload", lib.ParseValidationErrors(err))
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		lib.RespondValidationError(ctx, http.StatusBadRequest, "Validation failed", lib.ParseValidationErrors(err))
		return
	}

	req.IPAddress = ctx.ClientIP()
	req.UserAgent = ctx.Request.UserAgent()

	resp, err := h.authService.Login(ctx.Request.Context(), req)
	if err != nil {
		if appErr, ok := err.(*lib.AppError); ok {
			lib.RespondError(ctx, appErr)
			return
		}

		lib.RespondError(ctx, lib.ErrInternalServer)
		return
	}

	lib.RespondSuccess(ctx, http.StatusOK, lib.MsgLoginSuccess, resp)
}

// RefreshToken handles token refresh request
// @Summary Refresh access token
// @Description Get a new access token using a valid refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh token Request"
// @Success 200 {object} lib.APIResponse{data=dto.RefreshTokenResponse}
// @Failure 401 {object} lib.HTTPError
// @Failure 500 {object} lib.HTTPError
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandlerImpl) RefreshToken(ctx *gin.Context) {
	var req dto.RefreshTokenRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		lib.RespondValidationError(ctx, http.StatusBadRequest, "Bad payload", lib.ParseValidationErrors(err))
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		lib.RespondValidationError(ctx, http.StatusBadRequest, "Validation failed", lib.ParseValidationErrors(err))
		return
	}

	resp, err := h.authService.RefreshToken(ctx.Request.Context(), req.RefreshToken)
	if err != nil {
		if appErr, ok := err.(*lib.AppError); ok {
			lib.RespondError(ctx, appErr)
			return
		}

		lib.RespondError(ctx, lib.ErrInternalServer)
		return
	}

	lib.RespondSuccess(ctx, http.StatusOK, "Token refreshed successfully", resp)
}

// Logout handles user logout
// @Summary Logout user
// @Description Revoke current user session and invalidate the refresh token in database
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} lib.APIResponse
// @Failure 401 {object} lib.HTTPError
// @Failure 500 {object} lib.HTTPError
// @Router /api/v1/auth/logout [post]
func (h *AuthHandlerImpl) Logout(ctx *gin.Context) {
	claims := ctx.MustGet("user").(*lib.JWTClaims)
	if claims == nil {
		lib.RespondError(ctx, lib.ErrMissingContext)
		return
	}

	err := h.authService.Logout(ctx.Request.Context(), claims.SessionId)
	if err != nil {
		if appErr, ok := err.(*lib.AppError); ok {
			lib.RespondError(ctx, appErr)
			return
		}

		lib.RespondError(ctx, lib.ErrInternalServer)
		return
	}

	lib.RespondSuccess(ctx, http.StatusOK, "Logged out successfully", nil)
}

// ResetPassword handles password reset using token
// @Summary Reset user password
// @Description Reset user password using a reset token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.ResetPasswordRequest true "Reset Password Request"
// @Success 200 {object} lib.APIResponse
// @Failure 400 {object} lib.HTTPError
// @Failure 500 {object} lib.HTTPError
// @Router  /api/v1/auth/reset-password [post]
func (h *AuthHandlerImpl) ResetPassword(ctx *gin.Context) {
	var req dto.ResetPasswordRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		lib.RespondValidationError(ctx, http.StatusBadRequest, "Bad payload", lib.ParseValidationErrors(err))
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		lib.RespondValidationError(ctx, http.StatusBadRequest, "Validation failed", lib.ParseValidationErrors(err))
		return
	}

	err := h.authService.ResetPassword(ctx.Request.Context(), req.Token, req.NewPassword)
	if err != nil {
		if appErr, ok := err.(*lib.AppError); ok {
			lib.RespondError(ctx, appErr)
			return
		}

		lib.RespondError(ctx, lib.ErrInternalServer)
		return
	}

	lib.RespondSuccess(ctx, http.StatusOK, lib.MsgPasswordResetSuccess, nil)
}

// ForgotPassword handles forgot password request
// @Summary Forgot password
// @Description Send reset password link to email
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.ForgotPasswordRequest true "Forgot Password Request"
// @Success 200 {object} lib.APIResponse{message=string}
// @Failure 400 {object} lib.HTTPError
// @Failure 429 {object} lib.HTTPError
// @Failure 500 {object} lib.HTTPError
// @Router /api/v1/auth/forgot-password [post]
func (h *AuthHandlerImpl) ForgotPassword(ctx *gin.Context) {
	var req dto.ForgotPasswordRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		lib.RespondValidationError(ctx, http.StatusBadRequest, "Bad payload", lib.ParseValidationErrors(err))
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		lib.RespondValidationError(ctx, http.StatusBadRequest, "Validation failed", lib.ParseValidationErrors(err))
		return
	}

	ipAddress := ctx.ClientIP()
	userAgent := ctx.Request.UserAgent()

	err := h.authService.ForgotPassword(ctx.Request.Context(), req.Email, userAgent, ipAddress)
	if err != nil {
		lib.RespondError(ctx, lib.ErrInternalServer)
		return
	}

	lib.RespondSuccess(ctx, http.StatusOK, lib.MsgPasswordForgotSuccess, nil)
}

// CheckAvailability handles checking availability of email and/or username
// @Summary Check email and username availability
// @Description Check if an email or username is already registered/taken
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.CheckAvailabilityRequest true "Email and/or username to check"
// @Success 200 {object} lib.APIResponse{data=map[string]bool}
// @Failure 400 {object} lib.HTTPError
// @Router /api/v1/auth/check-availability [post]
func (h *AuthHandlerImpl) CheckAvailability(ctx *gin.Context) {
	var req dto.CheckAvailabilityRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		lib.RespondValidationError(ctx, http.StatusBadRequest, "Bad payload", lib.ParseValidationErrors(err))
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		lib.RespondValidationError(ctx, http.StatusBadRequest, "Validation failed", lib.ParseValidationErrors(err))
		return
	}

	result := gin.H{}

	if req.Email != "" {
		exists, err := h.authService.CheckEmail(ctx, req.Email)
		if err != nil {
			lib.RespondError(ctx, lib.ErrInternalServer)
		}
		result["email_available"] = !exists
	}

	if req.Username != "" {
		exists, err := h.authService.CheckUsername(ctx, req.Username)
		if err != nil {
			lib.RespondError(ctx, lib.ErrInternalServer)
		}
		result["username_available"] = !exists
	}

	lib.RespondSuccess(ctx, http.StatusOK, "Checked", result)
}
