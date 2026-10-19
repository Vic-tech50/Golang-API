package middleware



import (
	"net/http" // Provides HTTP status codes.
	"strconv"  // Converts retry time into a string.
	"sync"     // Protects shared data from concurrent access.
	"time"     // Provides time windows for rate limiting.

	"github.com/gin-gonic/gin" // Provides Gin middleware support.
)

// LoginAttempt stores the login request count and its time window.
type LoginAttempt struct {
	Count      int       // Number of requests from this IP in the window.
	WindowStart time.Time // Time when the current window started.
}

// LoginRateLimit limits requests from each client IP address.
func LoginRateLimit(maxAttempts int, window time.Duration) gin.HandlerFunc {
	// Store login attempts separately for each client IP address.
	attempts := make(map[string]LoginAttempt)

	// Protect the map because multiple requests can arrive simultaneously.
	var mutex sync.Mutex

	// Periodically remove expired records to prevent the map growing forever.
	go func() {
		ticker := time.NewTicker(window)

		// Keep checking for expired records while the application runs.
		for range ticker.C {
			now := time.Now()

			mutex.Lock()

			// Remove records whose time windows have expired.
			for ip, attempt := range attempts {
				if now.Sub(attempt.WindowStart) >= window {
					delete(attempts, ip)
				}
			}

			mutex.Unlock()
		}
	}()

	// Return the middleware function that Gin executes for each request.
	return func(c *gin.Context) {
		// Identify the client making the request.
		ip := c.ClientIP()

		// Get the current time.
		now := time.Now()

		// Lock the map while checking and updating the request count.
		mutex.Lock()

		// Get the previous attempts recorded for this IP address.
		attempt, exists := attempts[ip]

		// Start a new window if this is a new IP or the old window expired.
		if !exists || now.Sub(attempt.WindowStart) >= window {
			attempt = LoginAttempt{
				Count:       0,
				WindowStart: now,
			}
		}

		// Reject the request if the client has used all allowed attempts.
		if attempt.Count >= maxAttempts {
			// Calculate the approximate number of seconds until retrying.
			retryAfter := int(window.Seconds() -
				now.Sub(attempt.WindowStart).Seconds())

			if retryAfter < 1 {
				retryAfter = 1
			}

			// Save the current state before releasing the lock.
			attempts[ip] = attempt
			mutex.Unlock()

			// Tell the client when they can try again.
			c.Header("Retry-After", strconv.Itoa(retryAfter))

			// Return HTTP 429 instead of running the login handler.
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many login attempts. Please try again later.",
				"retry_after": retryAfter,
			})

			return
		}

		// Count this request as an attempt.
		attempt.Count++

		// Save the updated attempt record.
		attempts[ip] = attempt

		// Release the lock before continuing to the login handler.
		mutex.Unlock()

		// Continue to the next Gin handler, which performs login.
		c.Next()
	}
}
