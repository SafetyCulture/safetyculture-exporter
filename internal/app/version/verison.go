package version

// This variable should be overridden at build time using ldflags.
var version string = "0.0.0-dev"

func GetVersion() string {
	return version
}
