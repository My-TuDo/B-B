package validator

import (
	"regexp"

	vt "github.com/go-playground/validator/v10"
)

var Validate *vt.Validate

func Init() {
	Validate = vt.New()

	_ = Validate.RegisterValidation("username", validateUsername)
	_ = Validate.RegisterValidation("password", validatePassword)
}

func validateUsername(fl vt.FieldLevel) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]{3,20}$`, fl.Field().String())
	return matched
}

func validatePassword(fl vt.FieldLevel) bool {
	s := fl.Field().String()
	if len(s) < 8 {
		return false
	}
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(s)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(s)
	return hasLetter && hasDigit
}

func Struct(s interface{}) error {
	return Validate.Struct(s)
}
