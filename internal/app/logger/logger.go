package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var logger *zap.Logger

// InitLogger Initialize logger
func InitLogger(level string) {
	// Set log level
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

	// Create log rotation configuration
	logWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "./logs/tairitsu.log",
		MaxSize:    10, // Each log file max 10MB
		MaxBackups: 5,  // Max 5 backup files
		MaxAge:     30, // Max 30 days retention
		Compress:   true, // Compress old log files
	})

	// Create encoder configuration
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

	// Create core configuration
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		logWriter,
		zapLevel,
	)

	// Add console output based on environment variable
	var cores []zapcore.Core
	cores = append(cores, core)
	
	if os.Getenv("NODE_ENV") != "production" {
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

	// Create combined core
	logger = zap.New(zapcore.NewTee(cores...), zap.AddCaller(), zap.Development())

	// Ensure logs are flushed
	defer logger.Sync()
}

// GetLogger Get global logger instance
func GetLogger() *zap.Logger {
	return logger
}

// Debug Log Debug level log
func Debug(msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}

// Info Log Info level log
func Info(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

// Warn Log Warn level log
func Warn(msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}

// Error Log Error level log
func Error(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}

// Fatal Log Fatal level log
func Fatal(msg string, fields ...zap.Field) {
	logger.Fatal(msg, fields...)
}
