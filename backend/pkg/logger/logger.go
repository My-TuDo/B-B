// Package logger 基于 zap 提供结构化日志记录。
// 支持 debug / info / warn / error 四个级别，输出 JSON 格式到 stdout。
// 自动对敏感字段（password、token 等）和 AMQP DSN 进行脱敏。
package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger 全局日志实例，由 Init 初始化。
var Logger *zap.Logger

// Init 根据日志级别字符串初始化全局 Logger。
// level 取值：debug / info / warn / error，默认为 info。
func Init(level string) {
	// 解析日志级别
	var lvl zapcore.Level
	switch level {
	case "debug":
		lvl = zapcore.DebugLevel
	case "info":
		lvl = zapcore.InfoLevel
	case "warn":
		lvl = zapcore.WarnLevel
	case "error":
		lvl = zapcore.ErrorLevel
	default:
		lvl = zapcore.InfoLevel
	}

	// 使用 JSON 编码器，ISO8601 时间格式
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// 构建 Core：输出到 stdout
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		lvl,
	)

	Logger = zap.New(core)
}

// Debug 输出调试级别日志。
func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(sanitizeMsg(msg), sanitizeFields(fields)...)
}

// Info 输出信息级别日志。
func Info(msg string, fields ...zap.Field) {
	Logger.Info(sanitizeMsg(msg), sanitizeFields(fields)...)
}

// Warn 输出警告级别日志。
func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(sanitizeMsg(msg), sanitizeFields(fields)...)
}

// Error 输出错误级别日志。
func Error(msg string, fields ...zap.Field) {
	Logger.Error(sanitizeMsg(msg), sanitizeFields(fields)...)
}

// sensitiveFields 定义需要脱敏的日志字段名。
var sensitiveFields = map[string]bool{
	"password":     true,
	"passwordhash": true,
	"token":        true,
	"jwt":          true,
	"secret":       true,
}

// sanitizeFields 对敏感字段值进行脱敏处理，替换为 [REDACTED]。
func sanitizeFields(fields []zap.Field) []zap.Field {
	result := make([]zap.Field, len(fields))
	for i, f := range fields {
		if sensitiveFields[f.Key] {
			result[i] = zap.String(f.Key, "[REDACTED]")
		} else {
			result[i] = f
		}
	}
	return result
}

// sanitizeMsg 对日志消息中的敏感模式（如 AMQP DSN）进行脱敏。
// 将 amqp://user:pass@host:port/ 替换为 amqp://[REDACTED]@host:port/。
func sanitizeMsg(msg string) string {
	// 查找 amqp:// 前缀
	idx := strings.Index(msg, "amqp://")
	if idx < 0 {
		return msg
	}
	// 定位 @ 符号，确定认证信息边界
	rest := msg[idx+len("amqp://"):]
	atIdx := strings.Index(rest, "@")
	if atIdx < 0 {
		return msg
	}
	return msg[:idx+len("amqp://")] + "[REDACTED]" + rest[atIdx:]
}
