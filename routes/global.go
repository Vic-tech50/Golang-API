package routes

import (
	"net/http" // Provides HTTP status codes.

	"go-crud/config" // Imports the shared application settings.

	"github.com/gin-gonic/gin" // Provides Gin routing and JSON responses.
)

// ConfigRoute registers endpoints for viewing public app configuration.
func GetConfigRoute(router *gin.Engine) {
	// Register GET /config under the supplied v1 route group.
	router.GET("/config", func(c *gin.Context) {
		// c.String(http.StatusOK, "This is good")
		// Return selected configuration values as a JSON response.

		// Do not include secrets, passwords, or private API keys here.
		c.JSON(http.StatusOK, gin.H{
			"app_name":      config.App.Name,
			"version":       config.App.Version,
			"environment":   config.App.Environment,
			"api_base_url":  config.App.APIBaseURL,
			"cache_seconds": config.App.CacheSeconds,
			"CEO":           config.App.Author,
		})
	})
}
