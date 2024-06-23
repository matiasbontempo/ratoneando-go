package scrapers

import (
	"ratoneando/cores/vtex"
	"ratoneando/products"
)

func DiaOnline(query string) ([]products.Schema, error) {
	return vtex.Core(vtex.CoreProps{
		Query:   query,
		BaseUrl: "https://diaonline.supermercadosdia.com.ar",
		Source:  "diaonline",
	})
}
