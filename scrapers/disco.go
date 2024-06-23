package scrapers

import (
	"ratoneando/cores/vtex"
	"ratoneando/products"
)

func Disco(query string) ([]products.Schema, error) {
	return vtex.Core(vtex.CoreProps{
		Query:   query,
		BaseUrl: "https://www.disco.com.ar",
		Source:  "disco",
	})
}
