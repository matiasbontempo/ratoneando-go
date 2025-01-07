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
		"count":                4,
		"shippingOptions":      []string{},
		"variant":              nil,
	}
}

func GetExtensionsWithQuery(q string) map[string]interface{} {
	variables, _ := json.Marshal(GetVariablesWithQuery(q))
	return map[string]interface{}{
		"persistedQuery": map[string]interface{}{
			"version":    1,
			"sha256Hash": "0ef2c56d9518b51f912c2305ac4b07851c265b645dcbece6843c568bb91d39ff",
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
		"appsEtag":      "remove",
		"domain":        "store",
		"locale":        "es-AR",
		"operationName": "productSuggestions",
		"variables":     EncodeUrl("{}"),
		"extensions":    EncodeUrl(string(extensions)),
	}
	return EncodeQueryParams(queryParams)
}
