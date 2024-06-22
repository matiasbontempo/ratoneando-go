package main

import (
	"github.com/gin-gonic/gin"

	"ratoneando/config"
	"ratoneando/middlewares"
	"ratoneando/routes"
	"ratoneando/utils/logger"
)

func main() {
	config.InitConfig()
	port := config.PORT

	router := gin.Default()

	middlewares.CORS(router)

	// Register routes
	routes.RegisterRoutes(router)

	// Start the server
	logger.Log("Starting server on port " + port)
	if err := router.Run(":" + port); err != nil {
		logger.LogFatal("Could not start server: " + err.Error())
	}
}
