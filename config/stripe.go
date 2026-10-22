package config

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	// "go-crud/config"
	"go-crud/models"

	"github.com/gin-gonic/gin"
)

// stripeRequest sends a request to Stripe's API.
func stripeRequest(method, endpoint string, form url.Values) ([]byte, error) {
	// Get the secret key from .env.
	secretKey := os.Getenv("STRIPE_SECRET_KEY")
	if secretKey == "" {
		return nil, fmt.Errorf("STRIPE_SECRET_KEY is not configured")
	}

	// Encode form data for Stripe.
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}

	// Create the HTTP request.
	req, err := http.NewRequest(
		method,
		"https://api.stripe.com/v1"+endpoint,
		body,
	)
	if err != nil {
		return nil, err
	}

	// Stripe uses Bearer token authentication.
	req.Header.Set("Authorization", "Bearer "+secretKey)

	// Stripe Checkout Session creation uses form-encoded data.
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	// Avoid requests hanging forever.
	client := &http.Client{Timeout: 20 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read a limited response body.
	responseData, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}

	// Stripe returns non-2xx status codes for API errors.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf(
			"Stripe returned HTTP %d: %s",
			resp.StatusCode,
			string(responseData),
		)
	}

	return responseData, nil
}

// InitializeStripePayment creates a Stripe Checkout Session.
func InitializeStripePayment(c *gin.Context) {
	// Fields received from Postman.
	var input struct {
		Email       string `json:"email" binding:"required,email"`
		ProductName string `json:"product_name" binding:"required"`
		AmountCents int64  `json:"amount_cents" binding:"required,gt=0"`
	}

	// Validate the incoming JSON.
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Provide a valid email, product name, and positive amount_cents.",
		})
		return
	}

	// Generate a unique payment reference.
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not generate payment reference.",
		})
		return
	}

	reference := "VIC-STRIPE-" + hex.EncodeToString(randomBytes)

	// This learning example uses USD.
	currency := "usd"

	// Save the pending payment before creating the Checkout Session.
	payment := models.StripePayment{
		Email:     input.Email,
		Reference: reference,
		Amount:    input.AmountCents,
		Currency:  currency,
		Status:    "pending",
	}

	if err := DB.Create(&payment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not save payment record.",
		})
		return
	}

	// Read redirect URLs from .env.
	successURL := os.Getenv("STRIPE_SUCCESS_URL")
	cancelURL := os.Getenv("STRIPE_CANCEL_URL")

	if successURL == "" || cancelURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Stripe success and cancel URLs are not configured.",
		})
		return
	}

	// Build the Checkout Session request.
	form := url.Values{}
	form.Set("mode", "payment")
	form.Set("success_url", successURL+"?session_id={CHECKOUT_SESSION_ID}")
	form.Set("cancel_url", cancelURL)
	form.Set("customer_email", input.Email)
	form.Set("client_reference_id", reference)

	// Tell Stripe what the customer is purchasing.
	form.Set("line_items[0][price_data][currency]", currency)
	form.Set("line_items[0][price_data][unit_amount]", strconv.FormatInt(input.AmountCents, 10))
	form.Set("line_items[0][price_data][product_data][name]", input.ProductName)
	form.Set("line_items[0][quantity]", "1")

	// Attach our reference to the Checkout Session.
	form.Set("metadata[reference]", reference)

	// Create the Checkout Session with Stripe.
	responseData, err := stripeRequest(
		http.MethodPost,
		"/checkout/sessions",
		form,
	)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": "Could not create Stripe Checkout Session.",
		})
		return
	}

	// Read the session ID and hosted checkout URL.
	var response struct {
		ID  string `json:"id"`
		URL string `json:"url"`
	}

	if err := json.Unmarshal(responseData, &response); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": "Invalid response from Stripe.",
		})
		return
	}

	if response.ID == "" || response.URL == "" {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": "Stripe did not return a checkout URL.",
		})
		return
	}

	// Save the Checkout Session ID in our database.
	if err := DB.Model(&payment).Update(
		"checkout_session_id",
		response.ID,
	).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Checkout was created, but its session ID could not be saved.",
		})
		return
	}

	// Return the URL to the client.
	c.JSON(http.StatusOK, gin.H{
		"message":      "Stripe checkout created successfully.",
		"reference":    reference,
		"session_id":   response.ID,
		"checkout_url": response.URL,
		"amount_cents": input.AmountCents,
		"currency":     currency,
		"status":       "pending",
	})
}

// VerifyStripePayment checks the Checkout Session directly with Stripe.
func VerifyStripePayment(c *gin.Context) {
	// Read the Checkout Session ID and our payment reference.
	sessionID := strings.TrimSpace(c.Query("session_id"))
	reference := strings.TrimSpace(c.Query("reference"))

	if sessionID == "" || reference == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "session_id and reference are required.",
		})
		return
	}

	// Find the matching payment in MySQL.
	var payment models.StripePayment
	if err := DB.
		Where("reference = ? AND checkout_session_id = ?", reference, sessionID).
		First(&payment).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Matching payment record was not found.",
		})
		return
	}

	// Avoid processing an already successful payment twice.
	if payment.Status == "successful" {
		c.JSON(http.StatusOK, gin.H{
			"message":   "Payment has already been verified.",
			"reference": payment.Reference,
			"status":    payment.Status,
		})
		return
	}

	// Retrieve the Checkout Session from Stripe.
	responseData, err := stripeRequest(
		http.MethodGet,
		"/checkout/sessions/"+url.PathEscape(sessionID),
		nil,
	)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": "Could not retrieve Checkout Session from Stripe.",
		})
		return
	}

	// Decode the important Stripe fields.
	var response struct {
		ID              string `json:"id"`
		ClientReference string `json:"client_reference_id"`
		PaymentStatus   string `json:"payment_status"`
		AmountTotal     int64  `json:"amount_total"`
		Currency        string `json:"currency"`
		Metadata        struct {
			Reference string `json:"reference"`
		} `json:"metadata"`
	}

	if err := json.Unmarshal(responseData, &response); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": "Invalid verification response from Stripe.",
		})
		return
	}

	// Check the session, payment status, reference, amount, and currency.
	if response.ID != payment.CheckoutSessionID ||
		response.PaymentStatus != "paid" ||
		response.ClientReference != payment.Reference ||
		response.Metadata.Reference != payment.Reference ||
		response.AmountTotal != payment.Amount ||
		strings.ToLower(response.Currency) != strings.ToLower(payment.Currency) ||
		math.IsNaN(float64(response.AmountTotal)) {

		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Stripe payment has not been verified as successful.",
		})
		return
	}

	// Update the database only after the checks pass.
	if err := DB.Model(&payment).Updates(map[string]any{
		"status": "successful",
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Payment verified, but the database update failed.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Stripe payment verified successfully.",
		"reference":    payment.Reference,
		"amount_cents": payment.Amount,
		"currency":     payment.Currency,
		"status":       "successful",
	})
}
