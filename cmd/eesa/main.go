package main

import (
	"context"
	"log"
	"os"

	"fyne.io/fyne/v2/app"
	"github.com/company/eesa/internal/config"
	"github.com/company/eesa/internal/ui"
	"github.com/company/eesa/pkg/utils"
)

func main() {
	// Initialize configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialize logger
	logger := utils.NewLogger(cfg.LogLevel)
	logger.Info("Starting ESA application", utils.NewField("version", "1.0.0"))

	// Create Fyne application
	app := app.New()
	app.SetMetadata(&app.Metadata{
		ID:   "com.company.eesa",
		Name: "Executive Summary Automation",
	})

	// Create main window
	ctx := context.Background()
	mainWindow := ui.NewMainWindow(ctx, app, cfg, logger)

	// Show and run
	mainWindow.ShowAndRun()
}