package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealthReportsStatusAndShortCommit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("RAILWAY_GIT_COMMIT_SHA", "2c14d9a5e1f0b7a6c3d2e1f0a9b8c7d6e5f4a3b2")
	router := gin.New()
	router.GET("/health", Health)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" || body["commit"] != "2c14d9a" {
		t.Fatalf("unexpected body %v", body)
	}
	if _, ok := body["history"].(bool); !ok {
		t.Fatalf("expected a boolean history flag, got %v", body["history"])
	}
}
