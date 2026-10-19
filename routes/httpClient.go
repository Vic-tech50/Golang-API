package routes

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Data struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func HttpClientRoute(v1 *gin.RouterGroup) {

	v1.GET("/getdata", func(c *gin.Context) {
		// Create an HTTP GET request
		req, err := http.NewRequest(http.MethodGet, "https://jsonplaceholder.typicode.com/users/1", nil)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"error": err.Error(),
			})
			return
		}
		//set header data
		req.Header.Set("X-App-Name", "Victech")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"error": err.Error(),
			})
			return
		}
		defer resp.Body.Close()

		// Check the server's response status before processing its body.
		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			c.JSON(http.StatusBadGateway, gin.H{"status": resp.Status})
			return
		}

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"error": err.Error(),
			})
			return
		}

		var users Data //use this for single value
		// var users []Data //use this for multiple value
		if err := json.Unmarshal(body, &users); err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "invalid response from upstream"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user":   users,
			"header": resp.Header.Get("X-App-Name"), //read header data
		})
	})
}
