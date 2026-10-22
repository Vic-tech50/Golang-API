package routes

import (
	"go-crud/config"

	"github.com/gin-gonic/gin"
)

// RegisterFlutterwaveRoutes registers the Flutterwave endpoints.
func RegisterFlutterwaveRoutes(router *gin.Engine) {
	// Group related payment routes together.
	payments := router.Group("/api/flutterwave")
	{
		// Create a hosted payment link.
		payments.POST("/initialize", config.InitializeFlutterwavePayment)

		// Verify a completed payment.
		payments.GET("/verify", config.VerifyFlutterwavePayment)
	}
}