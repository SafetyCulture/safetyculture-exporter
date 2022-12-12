package main

import (
	"context"
	"fmt"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/httpapi"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	// settingsDir, err := logger.GetSettingsDirectory()
	// if err != nil {
	// 	panic("failed to get settings directory")
	// }
	//
	//
	//
	// _, err = os.Create(filepath.Join(settingsDir, "logs.log"))
	// if err != nil {
	// 	fmt.Printf("error while creating log file %v", err)
	// }

	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) ExportCSV() {

}

// ValidateApiKey validates the api, returns true if valid, false otherwise
func (a *App) ValidateApiKey(apiKey string) bool {
	var apiOpts []httpapi.Opt

	c := httpapi.NewClient("https://api.safetyculture.io", fmt.Sprintf("Bearer %s", apiKey), apiOpts...)
	res, err := c.WhoAmI(a.ctx)

	if err != nil {
		runtime.LogError(a.ctx, "something bad happened")
		return false
	}

	if res != nil && (res.UserID == "" || res.OrganisationID == "") {
		runtime.LogError(a.ctx, "something bad happened")
		return false
	}

	runtime.LogInfo(a.ctx, "api key is valid")
	return true
}
