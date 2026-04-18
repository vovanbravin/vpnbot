package logger

import (
	"io"
	"log/slog"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

var Log *slog.Logger

type Config struct {
	Level      string
	FilePath   string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
	Console    bool
}

func Init(config Config) error {

	var logLevel slog.Level

	switch config.Level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	logger := &lumberjack.Logger{
		MaxSize:    config.MaxSize,
		MaxAge:     config.MaxAge,
		MaxBackups: config.MaxBackups,
		Compress:   config.Compress,
		Filename:   config.FilePath,
	}

	var writers []io.Writer

	writers = append(writers, logger)

	if config.Console {
		writers = append(writers, os.Stdout)
	}

	writer := io.MultiWriter(writers...)

	opts := &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: true,
	}

	handler := slog.NewJSONHandler(writer, opts)

	Log = slog.New(handler)

	return nil
}

func Debug(message string, args ...any) {
	Log.Debug(message, args)
}

func Info(message string, args ...any) {
	Log.Info(message, args)
}

func Warn(message string, args ...any) {
	Log.Warn(message, args)
}

func Error(message string, args ...any) {
	Log.Error(message, args)
}
