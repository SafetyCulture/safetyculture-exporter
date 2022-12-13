package main

import (
	"embed"
	"fmt"

	extLogger "github.com/SafetyCulture/safetyculture-exporter/pkg/logger"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	settingsDir, err := CreateSettingsDirectory()
	if err != nil {
		panic(fmt.Sprintf("failed to get settings directory: %s", err.Error()))
	}

	// Create application with options
	err = wails.Run(&options.App{
		Title:  "SafetyCulture Exporter",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
		LogLevel: logger.INFO,
		Logger: extLogger.GetExporterLogger(settingsDir),
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
