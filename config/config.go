package config

// AppConfig defines the settings available throughout the application.
type AppConfig struct {
	Name         string // The name of the application.
	Version      string // The current application version.
	Environment  string // The environment where the app is running.
	APIBaseURL   string // The base URL of an external API.
	CacheSeconds int    // How long cached data should remain valid.
}

// App stores the default configuration values for the application.
// Other packages can access these values by importing this package.
var App = AppConfig{
	Name:         "Victech",
	Version:      "1.0.0",
	Environment:  "development",
	APIBaseURL:   "https://jsonplaceholder.typicode.com",
	CacheSeconds: 30,
}
