package products

import (
	"sort"
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

const FUZZY_SCORE_THRESHOLD = 3

// normalizeText makes the query and the product names comparable: hyphens are spaces.
func normalizeText(text string) string {
	return strings.ReplaceAll(text, "-", " ")
}

func Fuzzy(productsList []Schema, query string) []Schema {
	var filteredProducts []Schema
	normalizedQuery := normalizeText(query)
	for _, product := range productsList {
		normalizedProductName := normalizeText(product.Name)
		score := fuzzy.RankMatchNormalizedFold(normalizedQuery, normalizedProductName)
		if score > FUZZY_SCORE_THRESHOLD {
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
