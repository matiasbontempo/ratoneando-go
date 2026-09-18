package scrapers

import (
	"ratoneando/cores/vtex"
	"ratoneando/products"
)

func Jumbo(query string) ([]products.Schema, error) {
	return vtex.Core(vtex.CoreProps{
		Query:   query,
		BaseUrl: "https://www.jumbo.com.ar",
		Source:  "jumbo",
	})
}
