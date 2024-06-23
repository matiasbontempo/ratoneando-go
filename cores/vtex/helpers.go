package vtex

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"strings"
)

// EncodeBase64 encodes a string to base64.
func EncodeBase64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

// EncodeUrl encodes a string to URL encoding.
func EncodeUrl(str string) string {
	return url.QueryEscape(str)
}

// EncodeQueryParams encodes query parameters to URL query string.
func EncodeQueryParams(params map[string]string) string {
	var queryParams []string
	for key, value := range params {
		queryParams = append(queryParams, key+"="+value)
	}
	return "?" + strings.Join(queryParams, "&")
}

// GetVariablesWithQuery returns the variables with the query.
func GetVariablesWithQuery(q string) map[string]interface{} {
	return map[string]interface{}{
		"productOriginVtex":    true,
		"simulationBehavior":   "default",
		"hideUnavailableItems": true,
		"fullText":             q,
		"count":                6,
		"shippingOptions":      []string{},
		"variant":              nil,
	}
}

// GetExtensionsWithQuery returns the extensions with the query.
func GetExtensionsWithQuery(q string) map[string]interface{} {
	variables, _ := json.Marshal(GetVariablesWithQuery(q))
	return map[string]interface{}{
		"persistedQuery": map[string]interface{}{
			"version":    1,
			"sha256Hash": "38162aedddb0d0a8642b0fdb5beac3ff921e16d77701245aa71d464633a969b7",
			"sender":     "vtex.store-resources@0.x",
			"provider":   "vtex.search-graphql@0.x",
		},
		"variables": EncodeBase64(string(variables)),
	}
}

// EncodeQuery encodes the query into URL parameters.
func EncodeQuery(query string) string {
	extensions, _ := json.Marshal(GetExtensionsWithQuery(query))
	queryParams := map[string]string{
		"workspace":     "master",
		"maxAge":        "medium",
		"domain":        "store",
		"locale":        "es-AR",
		"operationName": "productSuggestions",
		"variables":     EncodeUrl("{}"),
		"extensions":    EncodeUrl(string(extensions)),
	}
	return EncodeQueryParams(queryParams)
}
