package routes

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func ReadWriteFileRoute(v1 *gin.RouterGroup) {

	v1.GET("/write", func(c *gin.Context) {
		err := os.WriteFile("message.txt", []byte("Hello, Go! Lorem Ipsum is simply dummy text of the printing and typesetting industry. \n \tLorem Ipsum has been the industry's standard dummy text ever since 1966, when designers at Letraset and James Mosley, the librarian at St Bride Printing Library in London, took a 1914 Cicero translation and scrambled it to make dummy text for Letraset's Body Type sheets. It has survived not only many decades, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised thanks to these sheets and more recently with desktop publishing software like Aldus PageMaker and Microsoft Word including versions of Lorem Ipsum."), 0644)

		if err != nil {
			fmt.Println("Error creating file:", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "File Created",
		})

	})

	v1.GET("/read", func(c *gin.Context) {
		data, err := os.ReadFile("message.txt")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"message": string(data),
		})

	})
}
