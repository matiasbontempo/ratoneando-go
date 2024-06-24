package vtex

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"strings"
)

func EncodeBase64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func EncodeUrl(str string) string {
	return url.QueryEscape(str)
}

func EncodeQueryParams(params map[string]string) string {
	var queryParams []string
	for key, value := range params {
		queryParams = append(queryParams, key+"="+value)
	}
	return "?" + strings.Join(queryParams, "&")
}

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
