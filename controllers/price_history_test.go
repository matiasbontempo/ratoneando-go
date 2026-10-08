package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"

	"ratoneando/config"
	"ratoneando/history"
)

func historyRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/history", PriceHistory)
	return router
}

func get(router *gin.Engine, url, referer string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, url, nil)
	if referer != "" {
		request.Header.Set("Referer", referer)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

// The disabled case must run before any test enables history, because the switch is global.
func TestHistoryEndpointIsNotFoundWhenHistoryIsDisabled(t *testing.T) {
	config.ENV = "debug"

	response := get(historyRouter(), "/history?source=disco&id=1", "")

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", response.Code)
	}
}

func TestHistoryEndpoint(t *testing.T) {
	config.ENV = "debug"
	config.HISTORY_WINDOW_DAYS = 60

	store, err := history.Open(filepath.Join(t.TempDir(), "history.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	if _, err := store.Record([]history.Item{
		{Source: "disco", ID: "1", Name: "Leche 1 L", Price: 1500},
	}); err != nil {
		t.Fatal(err)
	}
	history.UseStore(store)

	router := historyRouter()

	t.Run("returns the series of a known product", func(t *testing.T) {
		response := get(router, "/history?source=disco&id=1&days=5", "")
		if response.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
		}
		if got := response.Header().Get("Cache-Control"); got != "public, max-age=3600" {
			t.Fatalf("unexpected Cache-Control %q", got)
		}

		var body struct {
			Source string `json:"source"`
			ID     string `json:"id"`
			Name   string `json:"name"`
			Series []struct {
				Day   string  `json:"d"`
				Price float64 `json:"p"`
			} `json:"series"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body.Source != "disco" || body.ID != "1" || body.Name != "Leche 1 L" {
			t.Fatalf("unexpected body %+v", body)
		}
		if len(body.Series) != 1 || body.Series[0].Price != 1500 || body.Series[0].Day == "" {
			t.Fatalf("unexpected series %+v", body.Series)
		}
	})

	t.Run("unknown product", func(t *testing.T) {
		if code := get(router, "/history?source=disco&id=999", "").Code; code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", code)
		}
	})

	t.Run("missing parameters", func(t *testing.T) {
		for _, url := range []string{"/history", "/history?source=disco", "/history?id=1"} {
			if code := get(router, url, "").Code; code != http.StatusBadRequest {
				t.Fatalf("%s: expected 400, got %d", url, code)
			}
		}
	})

	t.Run("invalid days", func(t *testing.T) {
		for _, days := range []string{"abc", "0", "1", "181", "-5"} {
			if code := get(router, "/history?source=disco&id=1&days="+days, "").Code; code != http.StatusBadRequest {
				t.Fatalf("days=%s: expected 400, got %d", days, code)
			}
		}
	})

	t.Run("values are treated as data, not SQL", func(t *testing.T) {
		url := "/history?source=disco&id=1%27%20OR%20%271%27=%271"
		if code := get(router, url, "").Code; code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", code)
		}
	})

	t.Run("release mode requires the web app as referer", func(t *testing.T) {
		config.ENV = "release"
		config.WEB_URL = "https://ratoneando.ar"
		t.Cleanup(func() { config.ENV = "debug" })

		if code := get(router, "/history?source=disco&id=1", "").Code; code != http.StatusForbidden {
			t.Fatalf("no referer: expected 403, got %d", code)
		}
		if code := get(router, "/history?source=disco&id=1", "https://evil.example/").Code; code != http.StatusForbidden {
			t.Fatalf("foreign referer: expected 403, got %d", code)
		}
		if code := get(router, "/history?source=disco&id=1", "https://ratoneando.ar/?q=leche").Code; code != http.StatusOK {
			t.Fatalf("web referer: expected 200, got %d", code)
		}
	})
}
