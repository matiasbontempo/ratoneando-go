package routes

import (
	"github.com/gin-gonic/gin"

	"ratoneando/controllers"
)

func RegisterRoutes(router *gin.Engine) {
	// Register the routes
	router.GET("/", controllers.NormalizedScraper)
	router.GET("/raw", controllers.NormalizedScraper)
	router.GET("/history", controllers.PriceHistory)

	// Health check route
	router.GET("/health", controllers.Health)
}
