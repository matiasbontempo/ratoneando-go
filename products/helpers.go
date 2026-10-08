package products

import (
	"sort"
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

// normalizeText makes the query and the product names comparable: hyphens are spaces.
func normalizeText(text string) string {
	return strings.ReplaceAll(text, "-", " ")
}

func Fuzzy(productsList []Schema, query string) []Schema {
	var filteredProducts []Schema
	normalizedQuery := normalizeText(query)
	for _, product := range productsList {
		normalizedProductName := normalizeText(product.Name)
		// The score is a distance (0 is an identical name) and -1 means no match at all.
		score := fuzzy.RankMatchNormalizedFold(normalizedQuery, normalizedProductName)
		if score >= 0 {
			filteredProducts = append(filteredProducts, product)
		}
	}
	return filteredProducts
}

func Sort(productsList []Schema) []Schema {
	sortedProducts := make([]Schema, len(productsList))
	copy(sortedProducts, productsList)

	sort.Slice(sortedProducts, func(i, j int) bool {
		return sortedProducts[i].UnitPrice < sortedProducts[j].UnitPrice
	})

	return sortedProducts
}
