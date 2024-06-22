package routes

import (
	"github.com/gin-gonic/gin"

	"ratoneando/controllers"
)

func RegisterRoutes(router *gin.Engine) {
	// Register the routes
	router.GET("/", controllers.NormalizedScraper)
	router.GET("/raw", controllers.NormalizedScraper)

	// Health check route
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}
