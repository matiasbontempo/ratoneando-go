package scrapers

import (
	"ratoneando/cores/vtex"
	"ratoneando/products"
)

func Carrefour(query string) ([]products.Schema, error) {
	return vtex.Core(vtex.CoreProps{
		Query:   query,
		BaseUrl: "https://www.carrefour.com.ar",
		Source:  "carrefour",
	})
}
