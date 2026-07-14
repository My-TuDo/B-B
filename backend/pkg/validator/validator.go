// Package validator 基于 go-playground/validator 提供请求参数校验。
// 注册了两个自定义校验规则：
//   - username：3-20 位字母、数字或下划线
//   - password：至少 8 位，必须同时包含字母和数字
package validator

import (
	"regexp"

	vt "github.com/go-playground/validator/v10"
)

// Validate 全局校验器实例，由 Init 初始化。
var Validate *vt.Validate

// Init 初始化校验器并注册自定义校验规则。
func Init() {
	Validate = vt.New()

	// 注册自定义校验标签
	_ = Validate.RegisterValidation("username", validateUsername)
	_ = Validate.RegisterValidation("password", validatePassword)
}

// validateUsername 校验用户名：3-20 位字母、数字或下划线。
func validateUsername(fl vt.FieldLevel) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]{3,20}$`, fl.Field().String())
	return matched
}

// validatePassword 校验密码：至少 8 位，必须同时包含字母和数字。
func validatePassword(fl vt.FieldLevel) bool {
	s := fl.Field().String()
	if len(s) < 8 {
		return false
	}
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(s)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(s)
	return hasLetter && hasDigit
}

// Struct 对结构体执行所有已注册的校验规则。
func Struct(s interface{}) error {
	return Validate.Struct(s)
}
