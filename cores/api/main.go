package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"ratoneando/products"
	"ratoneando/unit"
	"ratoneando/utils/logger"
)

func Core[ResponseStructure any, RawProduct any](props CoreProps[ResponseStructure, RawProduct]) ([]products.Schema, error) {
	escapedQuery := url.PathEscape(props.Query)
	searchUrl := props.BaseUrl + props.SearchPattern(escapedQuery)

	resp, err := http.Get(searchUrl)
	if err != nil {
		logger.LogError("Failed to fetch the URL: " + escapedQuery + "@" + props.Source)
		return nil, fmt.Errorf(props.Source)
	}
	defer resp.Body.Close()

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
