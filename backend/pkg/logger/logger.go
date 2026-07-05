package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func Init(level string) {
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

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		lvl,
	)

	Logger = zap.New(core)
}

func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, sanitizeFields(fields)...)
}

func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, sanitizeFields(fields)...)
}

func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, sanitizeFields(fields)...)
}

func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, sanitizeFields(fields)...)
}

// sensitiveFields contains field keys that should be masked in logs.
var sensitiveFields = map[string]bool{
	"password":     true,
	"passwordhash": true,
	"token":        true,
	"jwt":          true,
	"secret":       true,
}

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
