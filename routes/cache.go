package routes

// Import packages used for JSON decoding, HTTP requests, synchronization, and expiration times.
import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// UserData represents the subset of user fields returned by the external API.
type UserData struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// CacheItem stores a user record together with the time at which it expires.
type CacheItem struct {
	Data      UserData
	ExpiresAt time.Time
}

// userCache stores cached user records, while cacheMutex protects concurrent access.
var (
	userCache  = make(map[string]CacheItem)
	cacheMutex sync.RWMutex
)

// GetCache returns a cached record and reports whether a valid entry exists.
func GetCache(key string) (UserData, bool) {
	// Read the entry while holding a shared lock so concurrent reads can proceed.
	cacheMutex.RLock()
	item, exists := userCache[key]
	cacheMutex.RUnlock()

	if !exists {
		// No value is associated with this key.
		return UserData{}, false
	}

	if time.Now().After(item.ExpiresAt) {
		// Recheck under an exclusive lock before deleting an expired entry.
		cacheMutex.Lock()
		// Check again in case another request refreshed the entry.
		item, exists = userCache[key]
		if exists && time.Now().After(item.ExpiresAt) {
			// Remove the entry only if it is still expired.
			delete(userCache, key)
			cacheMutex.Unlock()
			return UserData{}, false
		}
		cacheMutex.Unlock()

		if !exists {
			// The entry was removed while this request was waiting.
			return UserData{}, false
		}
		// Another request refreshed the entry, so return its current value.
		return item.Data, true
	}

	// Return the unexpired cached value.
	return item.Data, true
}

// SetCache saves a user record with a 30-second time-to-live.
func SetCache(key string, data UserData) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	userCache[key] = CacheItem{
		Data:      data,
		ExpiresAt: time.Now().Add(30 * time.Second),
	}
}

// CacheRoute registers the endpoint that serves cached or externally fetched user data.
func CacheRoute(v1 *gin.RouterGroup) {
	// Handle requests to retrieve the fixed user record.
	v1.GET("/getcachedata", func(c *gin.Context) {
		// Use a stable key for the user requested by this endpoint.
		cacheKey := "user:3"

		// 1. Check the cache first.
		if user, found := GetCache(cacheKey); found {
			// Return immediately when a valid cached record is available.
			c.JSON(http.StatusOK, gin.H{
				"message": "Data retrieved from cache",
				"source":  "cache",
				"user":    user,
			})
			return
			// Build a GET request for the user record when the cache misses.
		}

		// 2. Cache miss: request data from the external API.
		req, err := http.NewRequest(
			http.MethodGet,
			"https://jsonplaceholder.typicode.com/users/1",
			nil,
		)
		if err != nil {
			// Report an internal error if the request cannot be constructed.
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		// Identify this application's request to the external service.
		req.Header.Set("X-App-Name", "Victech")

		// Send the request using the default HTTP client.
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			// Report a gateway error if the external service cannot be reached.
			c.JSON(http.StatusBadGateway, gin.H{
				"error": err.Error(),
			})
			return
		}
		defer resp.Body.Close()

		// Reject unsuccessful HTTP responses from the external API.
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			c.JSON(http.StatusBadGateway, gin.H{
				"error":  "External API request failed",
				"status": resp.Status,
			})
			return
		}

		// Read the complete response body for decoding.
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			// Report an error when the response body cannot be read.
			c.JSON(http.StatusBadGateway, gin.H{
				"error": "Could not read external API response",
			})
			return
		}

		// Decode the JSON response into the application's user structure.
		var user UserData
		if err := json.Unmarshal(body, &user); err != nil {
			// Report malformed or incompatible JSON from the external API.
			c.JSON(http.StatusBadGateway, gin.H{
				"error": "Invalid response from external API",
			})
			return
			// Cache the decoded record for subsequent requests.
		}

		// Return the fetched user along with response metadata.
		// 3. Save the result in the cache for 30 seconds.
		SetCache(cacheKey, user)

		// 4. Return the result.
		c.JSON(http.StatusOK, gin.H{
			"message": "Data fetched from external API and cached",
			"source":  "external API",
			"user":    user,
			"header":  resp.Header.Get("Content-Type"),
			"app":     req.Header.Get("X-App-Name"),
		})

		log.Println("Welcome Here")
	})
}
