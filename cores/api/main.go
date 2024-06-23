package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"ratoneando/product"
	"ratoneando/unit"
	"ratoneando/utils/logger"
)

func Core[ResponseStructure any, RawProduct any](props CoreProps[ResponseStructure, RawProduct]) ([]product.Schema, error) {
	escapedQuery := url.PathEscape(props.Query)
	searchUrl := props.BaseUrl + props.SearchPattern(escapedQuery)

	resp, err := http.Get(searchUrl)
	if err != nil {
		logger.LogError("Failed to fetch the URL: " + props.Query + "@" + props.Source)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.LogError("Failed to read the response body: " + props.Query + "@" + props.Source)
		return nil, err
	}

	var responseStructure ResponseStructure
	err = json.Unmarshal(body, &responseStructure)
	if err != nil {
		logger.LogError("Failed to unmarshal the response body: " + props.Query + "@" + props.Source)
		return nil, err
	}

	normalizedProducts := props.Normalizer(responseStructure)

	products := make([]product.Schema, len(normalizedProducts))
	for i, product := range normalizedProducts {
		products[i] = unit.CalculateUnitInfo(props.Extractor(product))
	}

	return products, nil
}
