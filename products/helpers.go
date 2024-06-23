package products

import (
	"sort"
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

const FUZZY_SCORE_THRESHOLD = 3

func Fuzzy(productsList []Schema, query string) []Schema {
	var filteredProducts []Schema
	for _, product := range productsList {
		normalizedProductName := strings.ReplaceAll(product.Name, "-", " ")
		score := fuzzy.RankMatchNormalizedFold(query, normalizedProductName)
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
