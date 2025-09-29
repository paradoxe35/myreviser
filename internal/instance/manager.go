package instance

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

type Manager struct {
	lockFile *os.File
	lockPath string
}

func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	appDir := filepath.Join(homeDir, ".myreviser")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create app directory: %w", err)
	}

	lockPath := filepath.Join(appDir, ".lock")

	// Try to open the lock file
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open lock file: %w", err)
	}

	// Try to acquire exclusive lock (non-blocking)
	err = syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		lockFile.Close()
		return nil, fmt.Errorf("another instance is already running")
	}

	// Write PID to lock file
	lockFile.WriteString(fmt.Sprintf("%d\n", os.Getpid()))

	return &Manager{
		lockFile: lockFile,
		lockPath: lockPath,
	}, nil
}

func (m *Manager) Close() {
	if m.lockFile != nil {
		// Release the lock
		syscall.Flock(int(m.lockFile.Fd()), syscall.LOCK_UN)
		m.lockFile.Close()

		// Remove lock file
		os.Remove(m.lockPath)
	}
}