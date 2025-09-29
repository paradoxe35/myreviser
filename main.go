// MyReviser - AI-powered text revision tool
// Author: Paradoxe Ng <contact@pngwasi.me>
// Repository: https://github.com/paradoxe35/myreviser-go

package main

import (
	"log"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2/app"
	singleinstance "github.com/allan-simon/go-singleinstance"
	"github.com/paradoxe35/myreviser-go/internal/config"
	"github.com/paradoxe35/myreviser-go/internal/logger"
	"github.com/paradoxe35/myreviser-go/ui"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	// Initialize logger first
	if err := logger.Init(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Check for single instance
	// Create lock file in ~/.myscript/ directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Error("Failed to get user home directory", "error", err)
		os.Exit(1)
	}

	lockDir := filepath.Join(homeDir, ".myscript")
	if err := os.MkdirAll(lockDir, 0755); err != nil {
		logger.Error("Failed to create lock directory", "error", err)
		os.Exit(1)
	}

	lockPath := filepath.Join(lockDir, "myscript.lock")
	lockFile, err := singleinstance.CreateLockFile(lockPath)
	if err != nil {
		logger.Error("Another instance is already running", "error", err)
		os.Exit(1)
	}
	defer lockFile.Close()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		cfg = config.Default()
	}

	// Create Fyne application
	myApp := app.NewWithID("me.pngwasi.myreviser")
	myApp.Settings().SetTheme(&ui.MyReviserTheme{})

	logger.Info("MyReviser starting", "version", Version, "build_time", BuildTime)

	// Create and start the application
	application, err := NewApplication(myApp, cfg)
	if err != nil {
		logger.Error("Failed to create application", "error", err)
		os.Exit(1)
	}

	// Run the application
	if err := application.Start(); err != nil {
		logger.Error("Failed to start application", "error", err)
		os.Exit(1)
	}
}
