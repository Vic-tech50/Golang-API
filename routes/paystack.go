package routes

import (
	"go-crud/config"

	"github.com/gin-gonic/gin"
)

func RegisterPaymentRoutes(router *gin.Engine) {
	payments := router.Group("/api/payments")
	{
		// Initialize a new payment.
		payments.POST("/initialize", config.InitializePayment)

		// Verify a payment using its reference.
		payments.GET("/verify/:reference", config.VerifyPayment)
	}
}