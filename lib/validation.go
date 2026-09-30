package lib

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func getValidationMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return e.Field() + " is required"
	case "email":
		return "Invalid email format"
	case "min":
		return e.Field() + " must be at least " + e.Param() + " characters"
	case "max":
		return e.Field() + " must be at most " + e.Param() + " characters"
	case "alphanum":
		return e.Field() + " must contain only letters and numbers"
	case "username":
		return e.Field() + " must only contain lowercase letters, numbers, dots, or underscores."
	default:
		return e.Field() + " is invalid"
	}
}

func ParseValidationErrors(err error) []gin.H {
	var errors []gin.H

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors = append(errors, gin.H{
				"field":   e.Field(),
				"message": getValidationMessage(e),
			})
		}
	} else {
		errors = append(errors, gin.H{
			"field":   "unknown",
			"message": err.Error(),
		})
	}

	return errors
}
