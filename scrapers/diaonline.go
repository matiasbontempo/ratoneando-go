package scrapers

import (
	"ratoneando/cores/vtex"
	"ratoneando/product"
)

func DiaOnline(query string) ([]product.Schema, error) {
	return vtex.Core(vtex.CoreProps{
		Query:   query,
		BaseUrl: "https://diaonline.supermercadosdia.com.ar",
		Source:  "diaonline",
	})
}
