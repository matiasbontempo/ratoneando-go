package config

import (
	"os"
	"ratoneando/utils/logger"
	"strconv"

	"github.com/joho/godotenv"
)

var (
	PORT                      string
	ENV                       string
	WEB_URL                   string
	REDIS_URL                 string
	REDIS_CACHE_EXPIRATION    string
	RESPONSE_CACHE_EXPIRATION string
	CORE_CACHE_EXPIRATION     int
)

func InitConfig() {
	err := godotenv.Load()
	if err != nil {
		logger.LogFatal("Error loading .env file")
	}

	PORT = getEnv("PORT", "3000")
	ENV = getEnv("ENV", "development")
	WEB_URL = getEnv("WEB_URL", "http://localhost:5173")
	REDIS_URL = getEnv("REDIS_URL", "redis://localhost:6379")
	REDIS_CACHE_EXPIRATION = getEnv("REDIS_CACHE_EXPIRATION", "28800")
	RESPONSE_CACHE_EXPIRATION = getEnv("RESPONSE_CACHE_EXPIRATION", "3600")
	CORE_CACHE_EXPIRATION, _ = strconv.Atoi(getEnv("CORE_CACHE_EXPIRATION", "0"))
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}
