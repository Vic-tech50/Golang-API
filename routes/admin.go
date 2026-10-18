package routes

import "github.com/gin-gonic/gin"

func AdminRoute(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*.html")

	router.GET("/admin", func(c *gin.Context) {
		c.String(200, "Hello World from admin.go!")
	})
}
