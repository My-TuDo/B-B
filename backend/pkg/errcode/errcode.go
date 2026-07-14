// Package errcode 定义统一的业务错误码及其中文消息映射。
// 错误码分段：
//
//	1xxx — 参数/校验错误
//	2xxx — 用户相关错误
//	4xxx — 视频相关错误
//	5xxx — 分类相关错误
//	6xxx — 交互相关错误（弹幕、评论、投币、收藏、关注）
//
// HTTP 状态码 200/400/401/403/404/409/429/500 也在此统一定义。
package errcode

// messages 错误码到中文消息的映射表。
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
	CoverUploadFail:  "封面上传失败",
	AvatarUploadFail: "头像上传失败",

	DanmakuSendFail:   "弹幕发送失败",
	CommentNotFound:   "评论不存在",
	CommentDeleteFail: "评论删除失败",
	CoinLimitExceeded: "今日投币已达上限",
	CoinAlreadyCoined: "该视频已投过币，每视频仅可投一次",
	FavoriteNotFound:  "收藏夹不存在",
	CannotFollowSelf:  "不能关注自己",
}

// 错误码常量定义
const (
	// HTTP 标准状态码
	OK              = 200
	BadRequest      = 400
	Unauthorized    = 401
	Forbidden       = 403
	NotFound        = 404
	Conflict        = 409
	Internal        = 500
	TooManyRequests = 429

	// 业务错误码 — 参数/校验
	InvalidParams = 1001

	// 业务错误码 — 用户
	UserExists    = 2001
	UserNotFound  = 2002
	PasswordWrong = 2003
	TokenExpired  = 2004
	TokenInvalid  = 2005

	// 业务错误码 — 视频
	VideoNotFound    = 4001
	VideoUploadFail  = 4002
	FileTooLarge     = 4003
	InvalidFileType  = 4004
	CoverUploadFail  = 4005
	AvatarUploadFail = 4006

	// 业务错误码 — 分类
	CategoryNotFound = 5001

	// 业务错误码 — 交互
	DanmakuSendFail   = 6001
	CommentNotFound   = 6002
	CommentDeleteFail = 6003
	CoinLimitExceeded = 6004
	CoinAlreadyCoined = 6007
	FavoriteNotFound  = 6005
	CannotFollowSelf  = 6006
)

// Message 根据错误码返回对应的中文消息。
// 未注册的错误码返回"未知错误"。
func Message(code int) string {
	if msg, ok := messages[code]; ok {
		return msg
	}
	return "未知错误"
}
