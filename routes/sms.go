package routes

import (
	"go-crud/config"

	"github.com/gin-gonic/gin"
)

// RegisterSMSRoutes registers SMS-related endpoints.
func SMSRoutes(router *gin.Engine) {
	router.POST("/sms/send", config.SendSMS)
}
