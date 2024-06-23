package scrapers

import (
	"ratoneando/cores/vtex"
	"ratoneando/product"
)

func Disco(query string) ([]product.Schema, error) {
	return vtex.Core(vtex.CoreProps{
		Query:   query,
		BaseUrl: "https://www.disco.com.ar",
		Source:  "disco",
	})
}
