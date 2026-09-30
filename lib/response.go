package lib

import (
	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Success bool        `json:"success" default:"true"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type HTTPError struct {
	Success bool          `json:"success" default:"false"`
	Message string        `json:"message,omitempty"`
	Error   ErrorResponse `json:"error"`
}

type ErrorResponse struct {
	Code    string      `json:"code,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

func RespondSuccess(ctx *gin.Context, code int, message string, data interface{}) {
	if message == "" {
		message = MessageSuccess
	}
	ctx.JSON(code, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func RespondError(ctx *gin.Context, err *AppError) {
	ctx.JSON(err.HTTPStatus, HTTPError{
		Success: false,
		Message: err.Message,
		Error: ErrorResponse{
			Code: err.Code,
		},
	})
}

func RespondValidationError(ctx *gin.Context, code int, message string, details interface{}) {
	ctx.JSON(code, HTTPError{
		Success: false,
		Message: message,
		Error: ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Details: details,
		},
	})
}
