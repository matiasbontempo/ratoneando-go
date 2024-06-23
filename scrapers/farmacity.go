package scrapers

import (
	"ratoneando/cores/vtex"
	"ratoneando/products"
)

func Farmacity(query string) ([]products.Schema, error) {
	return vtex.Core(vtex.CoreProps{
		Query:   query,
		BaseUrl: "https://www.farmacity.com",
		Source:  "disco",
	})
}
