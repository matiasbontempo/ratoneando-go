package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"ratoneando/config"
)

// cacheExpiration is how long a response stays in Redis. A response that is missing stores
// (because some scrapers failed) is kept only briefly, so a short outage does not leave users
// with incomplete results for hours.
func cacheExpiration(failedScrapers []string) int {
	if len(failedScrapers) > 0 {
		return config.PARTIAL_CACHE_EXPIRATION
	}
	return config.REDIS_CACHE_EXPIRATION
}

// hitCacheControl is the Cache-Control header for a response served from Redis. A partial
// response must not be cached downstream (for example by Cloudflare) for the usual period.
func hitCacheControl(response gin.H) string {
	if failed, ok := response["failedScrapers"].([]interface{}); ok && len(failed) > 0 {
		return "public, max-age=" + strconv.Itoa(config.PARTIAL_CACHE_EXPIRATION)
	}
	return "public, max-age=" + config.RESPONSE_CACHE_EXPIRATION
}
