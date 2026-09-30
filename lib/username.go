package lib

import (
	"regexp"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var usernameRegex = regexp.MustCompile(`^[a-z0-9_.]+$`)

func username(fl validator.FieldLevel) bool {
	str := fl.Field().String()
	if str != strings.ToLower(str) {
		return false
	}
	return usernameRegex.MatchString(str)
}

func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("username", username)
	}
}
