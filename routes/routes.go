package routes

import "github.com/gin-gonic/gin"

func RegisterMe(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*.html")

	router.GET("/show", func(c *gin.Context) {
		c.HTML(200, "show.html", nil)
	})

	router.GET("/hello", func(c *gin.Context) {
		c.String(200, "Hello World from routes.go!")
	})
}
