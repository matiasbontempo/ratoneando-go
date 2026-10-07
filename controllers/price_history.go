package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"ratoneando/config"
	"ratoneando/history"
	"ratoneando/utils/logger"
)

const (
	maxHistoryDays   = 180
	maxIdentifierLen = 100
)

// PriceHistory returns the daily price series of one product, for the detail chart.
//
//	GET /history?source=disco&id=12345[&days=60]
func PriceHistory(c *gin.Context) {
	if !isAllowedReferer(c.Request.Referer()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden."})
		return
	}

	store := history.Active()
	if store == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "History is not available."})
		return
	}

	source := c.Query("source")
	id := c.Query("id")
	if source == "" || id == "" || len(source) > maxIdentifierLen || len(id) > maxIdentifierLen {
		c.JSON(http.StatusBadRequest, gin.H{"error": "source and id are required."})
		return
	}

	days := config.HISTORY_WINDOW_DAYS
	if raw := c.Query("days"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 2 || parsed > maxHistoryDays {
			c.JSON(http.StatusBadRequest, gin.H{"error": "days must be a number between 2 and 180."})
			return
		}
		days = parsed
	}

	name, series, ok, err := store.Series(source, id, days)
	if err != nil {
		logger.LogError("Price history read failed: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not read the history."})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Unknown product."})
		return
	}

	c.Header("Cache-Control", "public, max-age=3600")
	c.JSON(http.StatusOK, gin.H{
		"source": source,
		"id":     id,
		"name":   name,
		"series": series,
	})
}
