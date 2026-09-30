package lib

import "net/http"

type AppError struct {
	Code       string
	Message    string
	HTTPStatus int
}

func (e *AppError) Error() string {
	return e.Message
}

var (
	// general
	ErrInternalServer = &AppError{
		Code:       "INTERNAL_SERVER_ERROR",
		Message:    "Internal Server Error",
		HTTPStatus: http.StatusInternalServerError,
	}

	ErrBadRequest = &AppError{
		Code:       "BAD_REQUEST",
		Message:    "Invalid request",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrInvalidInput = &AppError{
		Code:       "INVALID_INPUT",
		Message:    "invalid input",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrBadPayload = &AppError{
		Code:       "BAD_PAYLOAD",
		Message:    "please check your payload",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrMissingRequired = &AppError{
		Code:       "MISSING_REQUIRED_FIELD",
		Message:    "missing required field",
		HTTPStatus: http.StatusBadRequest,
	}

	// auth
	ErrUnauthorized = &AppError{
		Code:       "UNAUTHORIZED",
		Message:    "Unauthorized",
		HTTPStatus: http.StatusUnauthorized,
	}

	ErrForbidden = &AppError{
		Code:       "FORBIDDEN",
		Message:    "Forbidden",
		HTTPStatus: http.StatusForbidden,
	}

	ErrInvalidCredential = &AppError{
		Code:       "INVALID_CREDENTIAL",
		Message:    "invalid credential",
		HTTPStatus: http.StatusUnauthorized,
	}

	// user
	ErrUserNotFound = &AppError{
		Code:       "USER_NOT_FOUND",
		Message:    "user not found",
		HTTPStatus: http.StatusNotFound,
	}

	ErrEmailAlreadyExists = &AppError{
		Code:       "EMAIL_ALREADY_EXISTS",
		Message:    "Email already registered",
		HTTPStatus: http.StatusConflict,
	}

	ErrEmailAlreadyVerified = &AppError{
		Code:       "EMAIL_ALREADY_Verified",
		Message:    "Email already verified",
		HTTPStatus: http.StatusConflict,
	}

	ErrEmailNotVerified = &AppError{
		Code:       "EMAIL_NOT_VERIFIED",
		Message:    "Please check your email to verify your account",
		HTTPStatus: http.StatusForbidden,
	}

	ErrAccountInactive = &AppError{
		Code:       "ACCOUNT_INACTIVE",
		Message:    "Your account is currently inactive or suspended",
		HTTPStatus: http.StatusForbidden,
	}

	ErrUsernameNotAvailable = &AppError{
		Code:       "USERNAME_NOT_AVAILABLE",
		Message:    "Username not available",
		HTTPStatus: http.StatusConflict,
	}

	ErrInvalidAuthorizationFormat = &AppError{
		Code:       "INVALID_AUTHORIZATION_FORMAT",
		Message:    "Invalid authorization format",
		HTTPStatus: http.StatusConflict,
	}

	// password
	ErrWeakPassword = &AppError{
		Code:       "WEAK_PASSWORD",
		Message:    "password must be at least 8 characters",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrInvalidPassword = &AppError{
		Code:       "INVALID_PASSWORD",
		Message:    "Password incorrect",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrInvalidCurrentPassword = &AppError{
		Code:       "INVALID_CURRENT_PASSWORD",
		Message:    "Current password is incorrect",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrSamePassword = &AppError{
		Code:       "SAME_PASSWORD",
		Message:    "New password must be different",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrUserHasNoPassword = &AppError{
		Code:       "USER_HAS_NO_PASSWORD",
		Message:    "user has no password yet",
		HTTPStatus: http.StatusBadRequest,
	}

	// token
	ErrInvalidToken = &AppError{
		Code:       "INVALID_TOKEN",
		Message:    "Invalid or expired token",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrMissingContext = &AppError{
		Code:       "MISSING_CONTEXT",
		Message:    "Missing context",
		HTTPStatus: http.StatusUnauthorized,
	}

	// data
	ErrRecordNotFound = &AppError{
		Code:       "RECORD_NOT_FOUND",
		Message:    "record not found",
		HTTPStatus: http.StatusNotFound,
	}

	ErrInvalidFileType = &AppError{
		Code:       "INVALID_FILE_TYPE",
		Message:    "File type is not allowed",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrFileTooLarge = &AppError{
		Code:       "FILE_TOO_LARGE",
		Message:    "File size exceeds maximum limit",
		HTTPStatus: http.StatusBadRequest,
	}

	// audit
	ErrAuditLogNotFound = &AppError{
		Code:       "AUDIT_LOG_NOT_FOUND",
		Message:    "Audit Log not found",
		HTTPStatus: http.StatusNotFound,
	}
)
