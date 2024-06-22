package middlewares

import (
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORS(router *gin.Engine) {
	corsConfig := cors.Config{
		AllowOrigins: []string{"*"},
	}
	if os.Getenv("NODE_ENV") == "production" {
		corsConfig.AllowOrigins = []string{os.Getenv("WEB_URL")}
	}
	router.Use(cors.New(corsConfig))
}
