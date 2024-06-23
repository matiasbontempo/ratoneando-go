package scrapers

import (
	"ratoneando/cores/vtex"
	"ratoneando/product"
)

func Carrefour(query string) ([]product.Schema, error) {
	return vtex.Core(vtex.CoreProps{
		Query:   query,
		BaseUrl: "https://www.carrefour.com.ar",
		Source:  "carrefour",
	})
}
