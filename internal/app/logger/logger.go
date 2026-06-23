package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var logger *zap.Logger

func ensureLogger() *zap.Logger {
	if logger == nil {
		logger = zap.NewNop()
	}
	return logger
}

// InitLogger initializes the logger
func InitLogger(level string) {
	// Set the log level
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zap.DebugLevel
	case "info":
		zapLevel = zap.InfoLevel
	case "warn":
		zapLevel = zap.WarnLevel
	case "error":
		zapLevel = zap.ErrorLevel
	default:
		zapLevel = zap.InfoLevel
	}

	// Configure log rotation
	logWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "./logs/tairitsu.log",
		MaxSize:    10,   // Max 10MB per log file
		MaxBackups: 5,    // Keep at most 5 backup files
		MaxAge:     30,   // Retain logs for up to 30 days
		Compress:   true, // Compress old log files
	})

	// Configure the encoder
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create the core
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		logWriter,
		zapLevel,
	)

	// Add console output based on the environment
	var cores []zapcore.Core
	cores = append(cores, core)

	if os.Getenv("APP_ENV") != "production" {
		consoleEncoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseColorLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		})

		cores = append(cores, zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapLevel))
	}

	// Create the combined core
	opts := []zap.Option{zap.AddCaller()}
	if os.Getenv("APP_ENV") != "production" {
		opts = append(opts, zap.Development())
	}
	logger = zap.New(zapcore.NewTee(cores...), opts...)
}

// Sync flushes any buffered log entries. Call this at application shutdown.
func Sync() {
	if logger != nil {
		_ = logger.Sync()
	}
}

// Debug logs a message at Debug level
func Debug(msg string, fields ...zap.Field) {
	ensureLogger().WithOptions(zap.AddCallerSkip(1)).Debug(msg, fields...)
}

// Info logs a message at Info level
func Info(msg string, fields ...zap.Field) {
	ensureLogger().WithOptions(zap.AddCallerSkip(1)).Info(msg, fields...)
}

// Warn logs a message at Warn level
func Warn(msg string, fields ...zap.Field) {
	ensureLogger().WithOptions(zap.AddCallerSkip(1)).Warn(msg, fields...)
}

// Error logs a message at Error level
func Error(msg string, fields ...zap.Field) {
	ensureLogger().WithOptions(zap.AddCallerSkip(1)).Error(msg, fields...)
}

// Fatal logs a message at Fatal level
func Fatal(msg string, fields ...zap.Field) {
	ensureLogger().WithOptions(zap.AddCallerSkip(1)).Fatal(msg, fields...)
}
