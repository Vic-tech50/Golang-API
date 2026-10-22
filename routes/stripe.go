
package routes

import (
	"go-crud/config"

	"github.com/gin-gonic/gin"
)

// RegisterStripeRoutes adds the Stripe payment endpoints.
func RegisterStripeRoutes(router *gin.Engine) {
	// Group Stripe endpoints under one URL prefix.
	payments := router.Group("/api/stripe")
	{
		// Initialize a new Checkout Session.
		payments.POST("/initialize", config.InitializeStripePayment)

		// Verify an existing Checkout Session.
		payments.GET("/verify", config.VerifyStripePayment)
	}
}
