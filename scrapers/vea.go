package scrapers

import (
	"ratoneando/cores/vtex"
	"ratoneando/product"
)

func Vea(query string) ([]product.Schema, error) {
	return vtex.Core(vtex.CoreProps{
		Query:   query,
		BaseUrl: "https://www.jumbo.com.ar",
		Source:  "vea",
	})
}
