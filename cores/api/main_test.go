package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"ratoneando/products"
	"ratoneando/utils/httpclient"
	"ratoneando/utils/logger"
)

type item struct {
	ID string `json:"id"`
}

func props(baseUrl string) CoreProps[[]item, item] {
	return CoreProps[[]item, item]{
		Query:         "yerba mañanita",
		BaseUrl:       baseUrl,
		SearchPattern: func(q string) string { return "/search?ft=" + q },
		Source:        "teststore",
		Normalizer:    func(list []item) []item { return list },
		Extractor: func(i item) products.ExtendedSchema {
			return products.ExtendedSchema{ID: i.ID, Name: "Producto 1 L", Price: 100, Unit: "un"}
		},
	}
}

// capture sends the logs to a buffer for the length of a test.
func capture(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	previous := logger.Logger
	logger.Logger = zerolog.New(&buf)
	t.Cleanup(func() { logger.Logger = previous })

	delay := retryDelay
	retryDelay = time.Millisecond
	t.Cleanup(func() { retryDelay = delay })
	return &buf
}

// server answers each request with the next response in the list (the last one repeats).
type response struct {
	status int
	body   string
}

func serve(t *testing.T, responses ...response) (*httptest.Server, *int32) {
	t.Helper()
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := int(atomic.AddInt32(&calls, 1)) - 1
		if n >= len(responses) {
			n = len(responses) - 1
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(responses[n].status)
		w.Write([]byte(responses[n].body))
	}))
	t.Cleanup(server.Close)
	return server, &calls
}

func TestSuccessNeedsOneRequestAndLogsNothing(t *testing.T) {
	logs := capture(t)
	server, calls := serve(t, response{206, `[{"id":"1"},{"id":"2"}]`})

	result, err := Core(props(server.URL))

	if err != nil || len(result) != 2 || atomic.LoadInt32(calls) != 1 {
		t.Fatalf("expected 2 products from 1 request, got %d products, %d calls, err %v", len(result), *calls, err)
	}
	if logs.Len() != 0 {
		t.Fatalf("expected no log output, got %s", logs.String())
	}
}

func TestATransientBadResponseIsRetriedOnce(t *testing.T) {
	logs := capture(t)
	server, calls := serve(t,
		response{200, `<html>blocked for a moment</html>`},
		response{200, `[{"id":"1"}]`},
	)

	result, err := Core(props(server.URL))

	if err != nil || len(result) != 1 || atomic.LoadInt32(calls) != 2 {
		t.Fatalf("expected recovery on the second request, got %d products, %d calls, err %v", len(result), *calls, err)
	}
	if !strings.Contains(logs.String(), "attempt 1 of 2") {
		t.Fatalf("expected the first failure to be logged, got %s", logs.String())
	}
}

func TestAPermanentFailureGivesUpAfterTwoAttemptsAndSaysWhy(t *testing.T) {
	logs := capture(t)
	server, calls := serve(t, response{503, "<html>\n<body>Service  Unavailable</body>\n</html>"})

	result, err := Core(props(server.URL))

	if err == nil || result != nil || err.Error() != "teststore" {
		t.Fatalf("expected the store name as the error, got %v %v", result, err)
	}
	if atomic.LoadInt32(calls) != 2 {
		t.Fatalf("expected exactly 2 attempts, got %d", *calls)
	}
	out := logs.String()
	for _, want := range []string{
		"Failed to unmarshal the response body", "status 503", "text", "application/json", "Service Unavailable",
		"yerba%20ma%C3%B1anita@teststore", "attempt 2 of 2",
	} {
		if !strings.Contains(out, want) && want != "text" {
			t.Fatalf("log should contain %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, "\\n") {
		t.Fatalf("the body snippet must be on one line, got:\n%s", out)
	}
}

func TestPersistedQueryNotFoundIsNotRetried(t *testing.T) {
	logs := capture(t)
	server, calls := serve(t, response{200, `{"errors":[{"message":"PersistedQueryNotFound","extensions":{"code":"PERSISTED_QUERY_NOT_FOUND"}}]}`})
	p := CoreProps[map[string]any, map[string]any]{
		Query: "leche", BaseUrl: server.URL, Source: "teststore",
		SearchPattern: func(q string) string { return "/?q=" + q },
		Normalizer:    func(map[string]any) []map[string]any { return nil },
		Extractor:     func(map[string]any) products.ExtendedSchema { return products.ExtendedSchema{} },
	}

	_, err := Core(p)

	if err == nil || atomic.LoadInt32(calls) != 1 {
		t.Fatalf("expected one attempt and an error, got %d calls, err %v", *calls, err)
	}
	if !strings.Contains(logs.String(), "API returned error: PersistedQueryNotFound") {
		t.Fatalf("unexpected log %s", logs.String())
	}
}

func TestOtherAPIErrorsAreRetried(t *testing.T) {
	capture(t)
	server, calls := serve(t,
		response{200, `{"errors":[{"message":"Request failed with status code 500"}]}`},
		response{200, `{}`},
	)
	p := CoreProps[map[string]any, map[string]any]{
		Query: "huevos", BaseUrl: server.URL, Source: "teststore",
		SearchPattern: func(q string) string { return "/?q=" + q },
		Normalizer:    func(map[string]any) []map[string]any { return nil },
		Extractor:     func(map[string]any) products.ExtendedSchema { return products.ExtendedSchema{} },
	}

	if _, err := Core(p); err != nil || atomic.LoadInt32(calls) != 2 {
		t.Fatalf("expected recovery after a retry, got %d calls, err %v", *calls, err)
	}
}

func TestAHangingStoreTimesOut(t *testing.T) {
	capture(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(3 * time.Second):
		}
	}))
	t.Cleanup(server.Close)

	previous := httpclient.Client.Timeout
	_ = previous
	httpclient.Client.Timeout = 100 * time.Millisecond
	t.Cleanup(func() { httpclient.Client.Timeout = previous })

	started := time.Now()
	_, err := Core(props(server.URL))

	if err == nil {
		t.Fatal("expected an error")
	}
	if elapsed := time.Since(started); elapsed > 2*time.Second {
		t.Fatalf("a hanging store must not hold the search, took %s", elapsed)
	}
}
