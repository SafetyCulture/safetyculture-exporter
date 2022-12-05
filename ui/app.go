package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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
	log.Printf("Api key -- >> %s", apiKey)
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.safetyculture.io/accounts/user/v1/user:WhoAmI", nil)
	req.Header.Set("Authorization", fmt.Sprintf("%s%s", "Bearer ", apiKey))
	if err != nil {
		log.Fatal(err)
		return false
	}

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
		return false
	}

	responseData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
		return false
	}

	var response *WhoAmIResponse
	err = json.Unmarshal(responseData, &response)
	if err != nil {
		log.Println(err)
		return false
	}

	if response == nil || len(response.UserID) == 0 || len(response.OrganisationID) == 0 {
		return false
	}

	log.Printf("User ID -- >> %s", response.UserID)
	log.Printf("Org ID -- >> %s", response.OrganisationID)
	log.Printf("First Name -- >> %s", response.FirstName)
	log.Printf("Last Name -- >> %s", response.LastName)

	return true
}

type WhoAmIResponse struct {
	UserID         string `json:"user_id"`
	OrganisationID string `json:"organisation_id"`
	FirstName      string `json:"firstname"`
	LastName       string `json:"lastname"`
}
