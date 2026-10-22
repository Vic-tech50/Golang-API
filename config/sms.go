package config

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/twilio/twilio-go"
	api "github.com/twilio/twilio-go/rest/api/v2010"
)

// SMSInput represents the JSON request from Postman.
type SMSInput struct {
	Phone   string `json:"phone" binding:"required"`
	Message string `json:"message" binding:"required"`
}

// SendSMS sends an SMS using Twilio.
func SendSMS(c *gin.Context) {
	var input SMSInput

	// Read the JSON request body.
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Phone and message are required",
		})
		return
	}

	// Remove unwanted spaces.
	input.Phone = strings.TrimSpace(input.Phone)
	input.Message = strings.TrimSpace(input.Message)

	if input.Phone == "" || input.Message == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Phone and message cannot be empty",
		})
		return
	}

	if len(input.Message) > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Message is too long",
		})
		return
	}

	// Load Twilio credentials from environment variables.
	accountSID := os.Getenv("TWILIO_ACCOUNT_SID")
	authToken := os.Getenv("TWILIO_AUTH_TOKEN")
	fromNumber := os.Getenv("TWILIO_FROM_NUMBER")

	if accountSID == "" || authToken == "" || fromNumber == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Twilio is not configured",
		})
		return
	}

	// Create the Twilio client.
	client := twilio.NewRestClientWithParams(
		twilio.ClientParams{
			Username: accountSID,
			Password: authToken,
		},
	)

	// Prepare the SMS.
	params := &api.CreateMessageParams{}
	params.SetTo(input.Phone)
	params.SetFrom(fromNumber)
	params.SetBody(input.Message)

	// Ask Twilio to send the message.
	// resp, err := client.Api.CreateMessage(params)
	// if err != nil {
	// 	c.JSON(http.StatusBadGateway, gin.H{
	// 		"error": "Twilio could not send the SMS",
	// 	})
	// 	return
	// }

	resp, err := client.Api.CreateMessage(params)
	if err != nil {
		// Print the actual error in your VS Code terminal.
		fmt.Printf("Twilio SMS error: %v\n", err)

		// Return the error to Postman while debugging.
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "Twilio could not send the SMS",
			"details": err.Error(),
		})
		return
	}

	// Return the Twilio message ID when available.
	response := gin.H{
		"message": "SMS request accepted by Twilio",
	}

	if resp != nil && resp.Sid != nil {
		response["message_sid"] = *resp.Sid
	}

	c.JSON(http.StatusOK, response)
}
