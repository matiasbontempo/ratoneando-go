package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"ratoneando/products"
	"ratoneando/unit"
	"ratoneando/utils/httpclient"
	"ratoneando/utils/logger"
)

const (
	maxAttempts   = 2
	snippetLength = 160

	persistedQueryNotFound = "PERSISTED_QUERY_NOT_FOUND"
)

// retryDelay is the pause before the second attempt. A variable so tests can shorten it.
var retryDelay = 300 * time.Millisecond

// failure describes why one attempt did not give a usable response.
type failure struct {
	message   string // starts with the same words the logs always used, so searches keep working
	retryable bool
}

// snippet is the start of a body on a single line, to tell an HTML error page from a truncated JSON.
func snippet(body []byte) string {
	text := string(body)
	if len(text) > snippetLength {
		text = text[:snippetLength]
	}
	return strings.Join(strings.Fields(text), " ")
}

type apiErrors struct {
	Errors []struct {
		Message    string `json:"message"`
		Extensions struct {
			Code string `json:"code"`
		} `json:"extensions"`
		Name string `json:"name"`
	} `json:"errors"`
}

func fetchAndParse[ResponseStructure any](searchUrl string) (ResponseStructure, *failure) {
	var responseStructure ResponseStructure

	resp, err := httpclient.Client.Get(searchUrl)
	if err != nil {
		return responseStructure, &failure{message: "Failed to fetch the URL (" + err.Error() + ")", retryable: true}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return responseStructure, &failure{
			message:   fmt.Sprintf("Failed to read the response body (status %d: %s)", resp.StatusCode, err),
			retryable: true,
		}
	}

	if err := json.Unmarshal(body, &responseStructure); err != nil {
		return responseStructure, &failure{
			message: fmt.Sprintf(
				"Failed to unmarshal the response body (status %d, content-type %q, %d bytes, starts with %q)",
				resp.StatusCode, resp.Header.Get("Content-Type"), len(body), snippet(body),
			),
			retryable: true,
		}
	}

	var errorCheck apiErrors
	if err := json.Unmarshal(body, &errorCheck); err == nil && len(errorCheck.Errors) > 0 {
		first := errorCheck.Errors[0]
		return responseStructure, &failure{
			message:   "API returned error: " + first.Message,
			retryable: first.Extensions.Code != persistedQueryNotFound,
		}
	}

	return responseStructure, nil
}

func Core[ResponseStructure any, RawProduct any](props CoreProps[ResponseStructure, RawProduct]) ([]products.Schema, error) {
	escapedQuery := url.PathEscape(props.Query)
	searchUrl := props.BaseUrl + props.SearchPattern(escapedQuery)
	label := escapedQuery + "@" + props.Source

	var responseStructure ResponseStructure
	for attempt := 1; ; attempt++ {
		parsed, problem := fetchAndParse[ResponseStructure](searchUrl)
		if problem == nil {
			responseStructure = parsed
			break
		}

		logger.LogError(fmt.Sprintf("%s: %s [attempt %d of %d]", problem.message, label, attempt, maxAttempts))
		if !problem.retryable || attempt >= maxAttempts {
			return nil, errors.New(props.Source)
		}
		time.Sleep(retryDelay)
	}

	normalizedProducts := props.Normalizer(responseStructure)

	products := make([]products.Schema, len(normalizedProducts))
	for i, product := range normalizedProducts {
		extractedProduct := props.Extractor(product)
		if extractedProduct.Unavailable {
			continue
		}
		products[i] = unit.CalculateUnitInfo(extractedProduct)
	}

	return products, nil
}
