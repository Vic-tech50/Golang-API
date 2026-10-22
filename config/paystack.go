package config

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"os"
	"strings"
	"time"

	"go-crud/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Paystack API base URL.
const paystackBaseURL = "https://api.paystack.co"

// InitializeInput contains the data sent by Postman.
type InitializeInput struct {
	Email      string `json:"email" binding:"required,email"`
	AmountKobo int64  `json:"amount_kobo" binding:"required,gt=0"`
}

// PaystackInitializeResponse represents Paystack's initialization response.
type PaystackInitializeResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		AuthorizationURL string `json:"authorization_url"`
		AccessCode       string `json:"access_code"`
		Reference        string `json:"reference"`
	} `json:"data"`
}

// PaystackVerifyResponse represents Paystack's verification response.
type PaystackVerifyResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Status    string `json:"status"`
		Reference string `json:"reference"`
		Amount    int64  `json:"amount"`
		Currency  string `json:"currency"`
	} `json:"data"`
}

// createReference generates a unique, unpredictable payment reference.
func createReference() (string, error) {
	randomBytes := make([]byte, 16)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	return "PAY-" + hex.EncodeToString(randomBytes), nil
}

// callPaystack sends an authenticated HTTP request to Paystack.
func callPaystack(method, endpoint string, body []byte) ([]byte, error) {
	secretKey := os.Getenv("PAYSTACK_SECRET_KEY")
	if secretKey == "" {
		return nil, errors.New("PAYSTACK_SECRET_KEY is not configured")
	}

	var requestBody io.Reader
	if body != nil {
		requestBody = bytes.NewReader(body)
	}

	req, err := http.NewRequest(
		method,
		paystackBaseURL+endpoint,
		requestBody,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+secretKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Limit response size to avoid reading an unexpectedly large response.
	responseBytes, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Paystack returned HTTP %d", resp.StatusCode)
	}

	return responseBytes, nil
}

// POST /api/payments/initialize
func InitializePayment(c *gin.Context) {
	var input InitializeInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "A valid email and positive amount_kobo are required",
		})
		return
	}

	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	// Validate the normalized email.
	if _, err := mail.ParseAddress(input.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email address",
		})
		return
	}

	// Generate our own payment reference.
	reference, err := createReference()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not generate payment reference",
		})
		return
	}

	// Save the pending payment before contacting Paystack.
	payment := models.Payment{
		Email:     input.Email,
		Reference: reference,
		Amount:    input.AmountKobo,
		Currency:  "NGN",
		Status:    "pending",
	}

	if err := DB.Create(&payment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not create payment record",
		})
		return
	}

	// Prepare the payload required by Paystack.
	payload := map[string]interface{}{
		"email":     payment.Email,
		"amount":    payment.Amount,
		"currency": payment.Currency,
		"reference": payment.Reference,
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not prepare payment",
		})
		return
	}

	// Initialize the payment on Paystack's server.
	responseBytes, err := callPaystack(
		http.MethodPost,
		"/transaction/initialize",
		jsonBody,
	)
	if err != nil {
		// Keep the record pending if the result is uncertain.
		// The request may have reached Paystack despite a network error.
		c.JSON(http.StatusBadGateway, gin.H{
			"error":     "Could not initialize payment with Paystack",
			"reference": payment.Reference,
		})
		return
	}

	var paystackResponse PaystackInitializeResponse

	if err := json.Unmarshal(responseBytes, &paystackResponse); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": "Invalid response from Paystack",
		})
		return
	}

	if !paystackResponse.Status ||
		paystackResponse.Data.AuthorizationURL == "" ||
		paystackResponse.Data.Reference != payment.Reference {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": "Paystack did not initialize the payment",
		})
		return
	}

	// Send the checkout URL to the caller.
	c.JSON(http.StatusOK, gin.H{
		"message":          "Payment initialized successfully",
		"reference":        payment.Reference,
		"authorization_url": paystackResponse.Data.AuthorizationURL,
		"access_code":      paystackResponse.Data.AccessCode,
		"amount_kobo":      payment.Amount,
		"currency":         payment.Currency,
		"status":            payment.Status,
	})
}

// GET /api/payments/verify/:reference
func VerifyPayment(c *gin.Context) {
	reference := strings.TrimSpace(c.Param("reference"))

	if reference == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Payment reference is required",
		})
		return
	}

	// Find the payment we originally created.
	var payment models.Payment

	if err := DB.
		Where("reference = ?", reference).
		First(&payment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Payment reference not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not retrieve payment",
		})
		return
	}

	// Verify directly with Paystack, not with client-submitted data.
	responseBytes, err := callPaystack(
		http.MethodGet,
		"/transaction/verify/"+reference,
		nil,
	)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": "Could not verify payment with Paystack",
		})
		return
	}

	var paystackResponse PaystackVerifyResponse

	if err := json.Unmarshal(responseBytes, &paystackResponse); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": "Invalid verification response from Paystack",
		})
		return
	}

	verified := paystackResponse.Status &&
		paystackResponse.Data.Status == "success" &&
		paystackResponse.Data.Reference == payment.Reference &&
		paystackResponse.Data.Amount == payment.Amount &&
		strings.EqualFold(
			paystackResponse.Data.Currency,
			payment.Currency,
		)

	if !verified {
		c.JSON(http.StatusOK, gin.H{
			"message": "Payment has not been verified as successful",
			"status":  payment.Status,
		})
		return
	}

	// Never downgrade a previously successful payment.
	if payment.Status != "success" {
		result := DB.Model(&models.Payment{}).
			Where("id = ? AND status <> ?", payment.ID, "success").
			Update("status", "success")

		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Payment verified, but status could not be saved",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Payment verified successfully",
		"reference": payment.Reference,
		"amount_kobo": payment.Amount,
		"currency": payment.Currency,
		"status":   "success",
	})
}