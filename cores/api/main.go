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

	normalizedProducts := props.Normalizer(responseStructure)

	products := make([]products.Schema, len(normalizedProducts))
	for i, product := range normalizedProducts {
		products[i] = unit.CalculateUnitInfo(props.Extractor(product))
	}

	return products, nil
}
