package main

import (
	"log"
	"os"

	"fyne.io/fyne/v2/app"
	"github.com/paradoxe35/myreviser-go/internal/config"
	"github.com/paradoxe35/myreviser-go/internal/instance"
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
	instanceMgr, err := instance.NewManager()
	if err != nil {
		logger.Error("Another instance is already running", "error", err)
		os.Exit(1)
	}
	defer instanceMgr.Close()

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
