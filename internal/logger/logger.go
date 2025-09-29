package logger

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"
)

var defaultLogger *slog.Logger

func Init() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	logDir := filepath.Join(homeDir, ".myreviser", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	logFile := filepath.Join(logDir, "myreviser.log")

	// Create rotating file writer
	fileWriter := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     30, // days
		Compress:   true,
	}

	// Create text handler with options
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	handler := slog.NewTextHandler(fileWriter, opts)
	defaultLogger = slog.New(handler)

	// Set as default
	slog.SetDefault(defaultLogger)

	return nil
}

// Convenience functions for logging
func Info(msg string, args ...any) {
	if defaultLogger != nil {
		defaultLogger.Info(msg, args...)
	}
}

func Error(msg string, args ...any) {
	if defaultLogger != nil {
		defaultLogger.Error(msg, args...)
	}
}

func Debug(msg string, args ...any) {
	if defaultLogger != nil {
		defaultLogger.Debug(msg, args...)
	}
}

func Warn(msg string, args ...any) {
	if defaultLogger != nil {
		defaultLogger.Warn(msg, args...)
	}
}

// GetCurrentLogFile returns the path to the current log file
func GetCurrentLogFile() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".myreviser", "logs", "myreviser.log")
}

// GetLogDirectory returns the path to the logs directory
func GetLogDirectory() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".myreviser", "logs")
}