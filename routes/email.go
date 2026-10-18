package routes

import (
	"go-crud/services"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type EmailInput struct {
	Email string `json:"email" binding:"required,email"`
}

func EmailRoute(v1 *gin.RouterGroup) {

	v1.POST("/email", func(c *gin.Context) {
		var input EmailInput

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Please provide a valid email address",
			})
			return
		}

		err := services.SendEmail(
			input.Email,
			"Test Email Notification",
			"Hello! Your Go email notification is working.",
		)

		if err != nil {
			// Log the technical error on the server, not in the response.
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Unable to send email",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Email sent successfully",
		})
	})

	//sending with html

	v1.POST("/emailHtml", func(c *gin.Context) {
		htmlBody, er := os.ReadFile("templates/email.html")
		if er != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Could not read HTML email file",
			})
			return
		}
		// To send Email from background task
		// Replace with the intended recipient.
		recipient := "recipient@example.com"
		subject := "Welcome to Our Platform"

		// Send the email in the background.
		go func(to, subject, body string) {
			err := services.SendHTMLEmail(to, subject, body)
			if err != nil {
				log.Printf("Failed to send email to %s: %v", to, err)
				return
			}

			log.Printf("Email sent successfully to %s", to)
		}(recipient, subject, string(htmlBody))

		// er = services.SendHTMLEmail(
		// 	"recipient@example.com",
		// 	"Welcome to Our Platform",
		// 	string(htmlBody),
		// )

		// if er != nil {
		// 	c.JSON(http.StatusInternalServerError, gin.H{
		// 		"error": "Failed to send HTML email",
		// 	})
		// 	return
		// }

		c.JSON(http.StatusOK, gin.H{
			"message": "HTML email sent successfully",
		})
	})
}
