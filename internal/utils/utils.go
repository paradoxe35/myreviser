package utils

import (
	"log"
	"os"
	"path/filepath"
)

func AppHomeDir(elem ...string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Failed to get user home directory", "error", err)
	}

	parts := append([]string{homeDir, ".myreviser"}, elem...)
	return filepath.Join(parts...)
}

// Ensure app home directory created
func EnsureAppHomeDir() {
	appDir := AppHomeDir()
	if err := os.MkdirAll(appDir, 0755); err != nil {
		log.Fatal("Failed to create lock directory", "error", err)
	}
}
