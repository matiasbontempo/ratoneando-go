package controllers

import (
	"encoding/json"
	"testing"

	"github.com/gin-gonic/gin"

	"ratoneando/config"
)

func TestCacheExpirationIsShortWhenStoresFailed(t *testing.T) {
	config.REDIS_CACHE_EXPIRATION = 28800
	config.PARTIAL_CACHE_EXPIRATION = 60

	if got := cacheExpiration(nil); got != 28800 {
		t.Fatalf("complete response: expected 28800, got %d", got)
	}
	if got := cacheExpiration([]string{}); got != 28800 {
		t.Fatalf("empty failure list: expected 28800, got %d", got)
	}
	if got := cacheExpiration([]string{"vea"}); got != 60 {
		t.Fatalf("partial response: expected 60, got %d", got)
	}
}

// The response is stored as JSON and read back into a gin.H, so the check has to work on what
// json.Unmarshal produces, not on the original []string.
func roundTrip(t *testing.T, response gin.H) gin.H {
	t.Helper()
	raw, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	out := gin.H{}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestHitCacheControl(t *testing.T) {
	config.RESPONSE_CACHE_EXPIRATION = "3600"
	config.PARTIAL_CACHE_EXPIRATION = 60

	complete := roundTrip(t, gin.H{"products": []string{}, "failedScrapers": []string(nil)})
	if got := hitCacheControl(complete); got != "public, max-age=3600" {
		t.Fatalf("complete response: unexpected header %q", got)
	}

	partial := roundTrip(t, gin.H{"products": []string{}, "failedScrapers": []string{"vea", "disco"}})
	if got := hitCacheControl(partial); got != "public, max-age=60" {
		t.Fatalf("partial response: unexpected header %q", got)
	}
}
