package config

import (
	"os"
	"ratoneando/utils/logger"
	"strconv"

	"github.com/joho/godotenv"
)

var (
	PORT                      = "3000"
	ENV                       = "development"
	WEB_URL                   = "http://localhost:5173"
	REDIS_URL                 = "redis://localhost:6379"
	REDIS_CACHE_EXPIRATION    = 28800
	RESPONSE_CACHE_EXPIRATION = "3600"
	CORE_CACHE_EXPIRATION     = 0
	VTEX_SHA256_HASH          = "REPLACE_ME"
	PARTIAL_CACHE_EXPIRATION  = 60
	HISTORY_ENABLED           = false
	HISTORY_DB_PATH           = "./data/history.db"
	HISTORY_WINDOW_DAYS       = 60
	HISTORY_MIN_POINTS        = 5
	HISTORY_MIN_SPAN_DAYS     = 14
)

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

func Init() {
	err := godotenv.Load()
	if err != nil {
		logger.LogWarn("Error loading .env file")
	}

	PORT = getEnv("PORT", "3000")
	ENV = getEnv("ENV", "development")
	WEB_URL = getEnv("WEB_URL", "http://localhost:5173")
	REDIS_URL = getEnv("REDIS_URL", "redis://localhost:6379")
	REDIS_CACHE_EXPIRATION, _ = strconv.Atoi(getEnv("REDIS_CACHE_EXPIRATION", "28800"))
	RESPONSE_CACHE_EXPIRATION = getEnv("RESPONSE_CACHE_EXPIRATION", "3600")
	CORE_CACHE_EXPIRATION, _ = strconv.Atoi(getEnv("CORE_CACHE_EXPIRATION", "0"))
	VTEX_SHA256_HASH = getEnv("VTEX_SHA256_HASH", "REPLACE_ME")
	PARTIAL_CACHE_EXPIRATION, _ = strconv.Atoi(getEnv("PARTIAL_CACHE_EXPIRATION", "60"))
	HISTORY_ENABLED = getEnv("HISTORY_ENABLED", "false") == "true"
	HISTORY_DB_PATH = getEnv("HISTORY_DB_PATH", "./data/history.db")
	HISTORY_WINDOW_DAYS, _ = strconv.Atoi(getEnv("HISTORY_WINDOW_DAYS", "60"))
	HISTORY_MIN_POINTS, _ = strconv.Atoi(getEnv("HISTORY_MIN_POINTS", "5"))
	HISTORY_MIN_SPAN_DAYS, _ = strconv.Atoi(getEnv("HISTORY_MIN_SPAN_DAYS", "14"))
}
