package ui

import (
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/company/eesa/internal/config"
	"github.com/company/eesa/pkg/utils"
)

// MainWindow represents the main application window
type MainWindow struct {
	app    fyne.App
	window fyne.Window
	config *config.Config
	logger utils.Logger
}

// NewMainWindow creates a new main window
func NewMainWindow(ctx context.Context, app fyne.App, config *config.Config, logger utils.Logger) *MainWindow {
	window := app.NewWindow("Executive Summary Automation")
	window.SetContent(widget.NewLabel("ESA - Coming Soon"))
	
	return &MainWindow{
		app:    app,
		window: window,
		config: config,
		logger: logger,
	}
}

// ShowAndRun shows the window and runs the application
func (w *MainWindow) ShowAndRun() {
	w.window.ShowAndRun()
}