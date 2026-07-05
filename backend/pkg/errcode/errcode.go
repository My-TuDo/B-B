package errcode

var messages = map[int]string{
	OK:               "成功",
	BadRequest:       "请求参数错误",
	Unauthorized:     "未授权，请先登录",
	Forbidden:        "无权限访问",
	NotFound:         "资源不存在",
	Conflict:         "资源冲突",
	Internal:         "服务器内部错误",
	TooManyRequests:  "请求过于频繁，请稍后再试",
	InvalidParams:    "参数校验失败",
	UserExists:       "用户已存在",
	UserNotFound:     "用户不存在",
	PasswordWrong:    "密码错误",
	TokenExpired:     "Token已过期",
	TokenInvalid:     "Token无效",
	VideoNotFound:    "视频不存在",
	VideoUploadFail:  "视频上传失败",
	FileTooLarge:     "文件大小超出限制",
	InvalidFileType:  "文件类型不支持",
	CategoryNotFound: "分类不存在",
}

const (
	OK              = 200
	BadRequest      = 400
	Unauthorized    = 401
	Forbidden       = 403
	NotFound        = 404
	Conflict        = 409
	Internal        = 500
	TooManyRequests = 429

	InvalidParams   = 1001
	UserExists      = 2001
	UserNotFound    = 2002
	PasswordWrong   = 2003
	TokenExpired    = 3001
	TokenInvalid    = 3002
	VideoNotFound   = 4001
	VideoUploadFail = 4002
	FileTooLarge    = 4003
	InvalidFileType = 4004
	CategoryNotFound = 5001
)

func Message(code int) string {
	if msg, ok := messages[code]; ok {
		return msg
	}
	return "未知错误"
}
