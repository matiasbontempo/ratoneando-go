package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"ratoneando/products"
	"ratoneando/unit"
	"ratoneando/utils/logger"
)

const USER_AGENT = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36"

var client = &http.Client{Timeout: 15 * time.Second}

func Core[ResponseStructure any, RawProduct any](props CoreProps[ResponseStructure, RawProduct]) ([]products.Schema, error) {
	escapedQuery := url.PathEscape(props.Query)
	searchUrl := props.BaseUrl + props.SearchPattern(escapedQuery)

	request, err := http.NewRequest(http.MethodGet, searchUrl, nil)
	if err != nil {
		logger.LogError("Failed to build the request: " + escapedQuery + "@" + props.Source)
		return nil, fmt.Errorf(props.Source)
	}

	// Some stores reject the default Go user agent
	request.Header.Set("User-Agent", USER_AGENT)
	request.Header.Set("Accept", "application/json")

	resp, err := client.Do(request)
	if err != nil {
		logger.LogError("Failed to fetch the URL: " + escapedQuery + "@" + props.Source)
		return nil, fmt.Errorf(props.Source)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		logger.LogError("Unexpected status code " + resp.Status + " for " + escapedQuery + "@" + props.Source)
		return nil, fmt.Errorf(props.Source)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.LogError("Failed to read the response body: " + escapedQuery + "@" + props.Source)
		return nil, fmt.Errorf(props.Source)
	}

	var responseStructure ResponseStructure
	err = json.Unmarshal(body, &responseStructure)
	if err != nil {
		logger.LogError("Failed to unmarshal the response body: " + escapedQuery + "@" + props.Source)
		return nil, fmt.Errorf(props.Source)
	}

	var errorCheck struct {
		Errors []struct {
			Message    string `json:"message"`
			Extensions struct {
				Code string `json:"code"`
			} `json:"extensions"`
			Name string `json:"name"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(body, &errorCheck); err == nil && len(errorCheck.Errors) > 0 {
		logger.LogError("API returned error: " + errorCheck.Errors[0].Message + " for " + escapedQuery + "@" + props.Source)
		return nil, fmt.Errorf(props.Source)
	}

	normalizedProducts := props.Normalizer(responseStructure)

	parsedProducts := make([]products.Schema, 0, len(normalizedProducts))
	for _, product := range normalizedProducts {
		extractedProduct := props.Extractor(product)
		if extractedProduct.Unavailable || extractedProduct.ID == "" {
			continue
		}
		parsedProducts = append(parsedProducts, unit.CalculateUnitInfo(extractedProduct))
	}

	return parsedProducts, nil
}
