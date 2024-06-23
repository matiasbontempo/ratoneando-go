package scrapers

import (
	"ratoneando/cores/vtex"
	"ratoneando/product"
)

func Farmacity(query string) ([]product.Schema, error) {
	return vtex.Core(vtex.CoreProps{
		Query:   query,
		BaseUrl: "https://www.farmacity.com",
		Source:  "disco",
	})
}
