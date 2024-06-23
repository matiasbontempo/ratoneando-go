package scrapers

import (
	"ratoneando/cores/vtex"
	"ratoneando/products"
)

func MasOnline(query string) ([]products.Schema, error) {
	return vtex.Core(vtex.CoreProps{
		Query:   query,
		BaseUrl: "https://www.masonline.com.ar",
		Source:  "masonline",
	})
}
