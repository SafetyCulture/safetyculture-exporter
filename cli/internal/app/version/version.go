package version

// This variable should be overridden at build time using ldflags.
var version string = "0.0.0-dev"
var integrationID string = "safetyculture-exporter"

// GetVersion returns current version of the app
func GetVersion() string {
	return version
}

// GetIntegrationID returns the integration id
func GetIntegrationID() string {
	return integrationID
}
