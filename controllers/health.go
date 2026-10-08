package controllers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"ratoneando/history"
)

// Health tells a deploy is up, which commit it runs (Railway sets RAILWAY_GIT_COMMIT_SHA) and
// whether price history is recording. Nothing in it is secret.
func Health(c *gin.Context) {
	commit := os.Getenv("RAILWAY_GIT_COMMIT_SHA")
	if len(commit) > 7 {
		commit = commit[:7]
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"commit":  commit,
		"history": history.Active() != nil,
	})
}
