package handler

import (
	"github/OfrenDialsa/go-gin-starter/internal/modules/user/dto"
	"github/OfrenDialsa/go-gin-starter/internal/modules/user/service"
	"github/OfrenDialsa/go-gin-starter/lib"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

type UserHandlerImpl struct {
	userService service.UserService
	validator   *validator.Validate
}

func NewUserHandler(userService service.UserService) UserHandler {
	return &UserHandlerImpl{
		userService: userService,
		validator:   validator.New(),
	}
}

// GetMe godoc
// @Summary      Get current user profile
// @Description  Fetches the detailed profile of the authenticated user
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  lib.APIResponse{data=dto.GetMeResponse} "Successfully retrieved profile"
// @Failure      401  {object}  lib.HTTPError "Unauthorized: Missing or invalid token"
// @Failure      403  {object}  lib.HTTPError "Forbidden: User is not associated with the requested organization"
// @Failure      404  {object}  lib.HTTPError "Not Found: User record does not exist"
// @Failure      500  {object}  lib.HTTPError "Internal Server Error: Something went wrong on our end"
// @Router       /api/v1/users/me [get]
func (u *UserHandlerImpl) GetMe(ctx *gin.Context) {
	claims := ctx.MustGet("user").(*lib.JWTClaims)
	if claims == nil {
		lib.RespondError(ctx, lib.ErrMissingContext)
		return
	}

	data, err := u.userService.GetMe(ctx, claims.UserId)
	if err != nil {
		log.Error().Err(err).Str("userId", claims.UserId).Msg("Error in GetMe")

		if appErr, ok := err.(*lib.AppError); ok {
			lib.RespondError(ctx, appErr)
			return
		}

		lib.RespondError(ctx, lib.ErrInternalServer)
		return
	}

	lib.RespondSuccess(ctx, http.StatusOK, "Get profile success", data)
}

// UpdateProfile godoc
// @Summary      Update user profile
// @Description  Updates the name and avatar URL of the authenticated user.
// @Tags         Users
// @Accept       mpfd
// @Produce      json
// @Security     BearerAuth
// @Param        username    formData  string  false  "New username of the user"
// @Param        name        formData  string  false  "New name of the user"
// @Param        avatar      formData  file    false  "User avatar image file"
// @Success      200  {object}  lib.APIResponse{data=dto.UpdateProfileResponse} "Profile updated successfully"
// @Failure      400  {object}  lib.HTTPError "Bad Request: Invalid input"
// @Failure      401  {object}  lib.HTTPError "Unauthorized"
// @Failure      404  {object}  lib.HTTPError "Not Found"
// @Failure      500  {object}  lib.HTTPError "Internal Server Error"
// @Router       /api/v1/users/me [put]
func (u *UserHandlerImpl) UpdateProfile(ctx *gin.Context) {
	claims := ctx.MustGet("user").(*lib.JWTClaims)
	if claims == nil {
		lib.RespondError(ctx, lib.ErrMissingContext)
		return
	}

	var req dto.UpdateProfileRequest
	if err := ctx.ShouldBind(&req); err != nil {
		lib.RespondValidationError(ctx, http.StatusBadRequest, "Bad payload", lib.ParseValidationErrors(err))
		return
	}

	if err := u.validator.Struct(&req); err != nil {
		lib.RespondValidationError(ctx, http.StatusBadRequest, "Validation failed", lib.ParseValidationErrors(err))
		return
	}

	req.UserId = claims.UserId
	req.Username = strings.TrimSpace(req.Username)
	req.Name = strings.TrimSpace(req.Name)

	if req.Username == "" && req.Name == "" && req.Avatar == nil {
		lib.RespondError(ctx, lib.ErrBadPayload)
		return
	}

	if req.Username != "" {
		req.Username = strings.Join(strings.Fields(req.Username), " ")
	}

	if req.Name != "" {
		req.Name = strings.Join(strings.Fields(req.Name), " ")
	}

	if req.Avatar != nil {
		file, err := req.Avatar.Open()
		if err != nil {
			lib.RespondError(ctx, lib.ErrInternalServer)
			return
		}
		defer file.Close()

		buf := make([]byte, 512)
		_, err = file.Read(buf)
		if err != nil {
			lib.RespondError(ctx, lib.ErrInternalServer)
			return
		}

		if _, err := file.Seek(0, 0); err != nil {
			lib.RespondError(ctx, lib.ErrInternalServer)
			return
		}

		contentType := http.DetectContentType(buf)

		allowedTypes := map[string]bool{
			"image/jpeg": true,
			"image/png":  true,
			"image/webp": true,
		}

		if !allowedTypes[contentType] {
			lib.RespondError(ctx, lib.ErrInvalidFileType)
			return
		}

		if req.Avatar.Size > 2*1024*1024 {
			lib.RespondError(ctx, lib.ErrFileTooLarge)
			return
		}
	}

	req.Ip = ctx.ClientIP()
	req.Ua = ctx.Request.UserAgent()

	data, err := u.userService.UpdateProfile(ctx.Request.Context(), req)
	if err != nil {
		log.Error().Err(err).Str("userId", claims.UserId).Msg("Error in UpdateProfile Service")

		if appErr, ok := err.(*lib.AppError); ok {
			lib.RespondError(ctx, appErr)
			return
		}

		lib.RespondError(ctx, lib.ErrInternalServer)
		return
	}

	lib.RespondSuccess(ctx, http.StatusOK, "Update profile success", data)
}

// ChangePassword godoc
// @Summary      Change user password
// @Description  Allows an authenticated user to change their password and optionally revoke other active sessions.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.ChangePasswordRequest  true  "Change Password Request"
// @Success      200   {object}  lib.APIResponse            "Password changed successfully"
// @Failure      400   {object}  lib.HTTPError              "Bad Request: Validation error or password mismatch"
// @Failure      401   {object}  lib.HTTPError              "Unauthorized: Missing or invalid token"
// @Failure      500   {object}  lib.HTTPError              "Internal Server Error"
// @Router       /api/v1/users/me/password [put]
func (u *UserHandlerImpl) ChangePassword(ctx *gin.Context) {
	claims := ctx.MustGet("user").(*lib.JWTClaims)
	if claims == nil {
		lib.RespondError(ctx, lib.ErrMissingContext)
		return
	}

	var req dto.ChangePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		lib.RespondValidationError(ctx, http.StatusBadRequest, "Bad payload", lib.ParseValidationErrors(err))
		return
	}

	if err := u.validator.Struct(&req); err != nil {
		lib.RespondValidationError(ctx, http.StatusBadRequest, "Validation failed", lib.ParseValidationErrors(err))
		return
	}

	req.UserId = claims.UserId
	req.SessionId = claims.SessionId
	err := u.userService.ChangePassword(ctx, req)
	if err != nil {
		log.Error().Err(err).Str("userId", req.UserId).Msg("Error in ChangePassword")

		if appErr, ok := err.(*lib.AppError); ok {
			lib.RespondError(ctx, appErr)
			return
		}

		lib.RespondError(ctx, lib.ErrInternalServer)
		return
	}

	lib.RespondSuccess(ctx, http.StatusOK, "Password updated successfully", nil)
}

// DeleteAccount godoc
// @Summary      Delete user account
// @Description  Deletes the authenticated user's account.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.UserDeleteAccountRequest true "Delete account request"
// @Success      200  {object}  lib.APIResponse "Account deleted successfully"
// @Failure      400  {object}  lib.HTTPError "Invalid password or ownership transfer required"
// @Failure      404  {object}  lib.HTTPError "User not found"
// @Failure      500  {object}  lib.HTTPError "Internal server error"
// @Router       /api/v1/users/me [delete]
func (u *UserHandlerImpl) DeleteAccount(ctx *gin.Context) {
	claims := ctx.MustGet("user").(*lib.JWTClaims)
	if claims == nil {
		lib.RespondError(ctx, lib.ErrMissingContext)
		return
	}

	var req dto.UserDeleteAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		lib.RespondValidationError(ctx, http.StatusBadRequest, "Bad payload", lib.ParseValidationErrors(err))
		return
	}

	if err := u.validator.Struct(&req); err != nil {
		lib.RespondValidationError(ctx, http.StatusBadRequest, "Validation failed", lib.ParseValidationErrors(err))
		return
	}

	req.UserId = claims.UserId
	err := u.userService.DeleteAccount(ctx, req)
	if err != nil {
		log.Error().Err(err).Str("userId", claims.UserId).Msg("Error in DeleteAccount")

		if appErr, ok := err.(*lib.AppError); ok {
			lib.RespondError(ctx, appErr)
			return
		}

		lib.RespondError(ctx, lib.ErrInternalServer)
		return
	}

	lib.RespondSuccess(ctx, http.StatusOK, "Account deleted successfully", nil)
}

// DeleteAvatar godoc
// @Summary      Delete user avatar
// @Description  Deletes the current authenticated user's avatar image and sets the URL to null.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  lib.APIResponse "Avatar deleted successfully"
// @Failure      401  {object}  lib.HTTPError "Unauthorized"
// @Failure      404  {object}  lib.HTTPError "User not found"
// @Failure      500  {object}  lib.HTTPError "Internal Server Error"
// @Router       /api/v1/users/me/avatar [delete]
func (u *UserHandlerImpl) DeleteAvatar(ctx *gin.Context) {
	claims := ctx.MustGet("user").(*lib.JWTClaims)
	if claims == nil {
		lib.RespondError(ctx, lib.ErrMissingContext)
		return
	}

	err := u.userService.DeleteAvatar(ctx.Request.Context(), claims.UserId)
	if err != nil {
		log.Error().Err(err).Str("userId", claims.UserId).Msg("Error in DeleteAvatar Handler")

		if appErr, ok := err.(*lib.AppError); ok {
			lib.RespondError(ctx, appErr)
			return
		}

		lib.RespondError(ctx, lib.ErrInternalServer)
		return
	}

	lib.RespondSuccess(ctx, http.StatusOK, "Avatar deleted successfully", nil)
}
