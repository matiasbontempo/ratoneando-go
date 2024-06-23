package scrapers

import (
	"ratoneando/cores/vtex"
	"ratoneando/product"
)

func MasOnline(query string) ([]product.Schema, error) {
	return vtex.Core(vtex.CoreProps{
		Query:   query,
		BaseUrl: "https://www.masonlinae.com.ar",
		Source:  "masonline",
	})
}
