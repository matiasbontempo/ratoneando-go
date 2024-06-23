package scrapers

import (
	"ratoneando/cores/html"
	"ratoneando/product"
	"strconv"
)

func MercadoLibre(query string) ([]product.Schema, error) {
	return html.Core(html.CoreProps{
		Query:             query,
		BaseUrl:           "https://listado.mercadolibre.com.ar",
		Source:            "mercadolibre",
		SearchPattern:     func(q string) string { return "/supermercado/" + q + "#D[A:" + q + ",on]" },
		ContainerSelector: "div.ui-search-main",
		ProductSelector:   ".andes-card.ui-search-result",
		SkipIfSelector:    ".ui-search-zrp-disclaimer",
		Extractor: func(element *html.ElementWrapper, doc *html.DocumentWrapper) product.ExtendedSchema {
			id, _ := element.Attr("id")
			name := element.Find("h2.ui-search-item__title").Text()
			price, _ := strconv.ParseFloat(element.Find("div.ui-search-price__second-line span.andes-money-amount__fraction").Text(), 64)
			link, _ := element.Find("a.ui-search-link").Attr("href")
			image, _ := element.Find("img.ui-search-result-image__element").Attr("data-src")
			unavailable, _ := element.Find("div.ui-search-card-add-to-cart").Attr("disabled")

			return product.ExtendedSchema{
				ID:          id,
				Name:        name,
				Link:        link,
				Image:       image,
				Price:       price,
				Unavailable: unavailable == "disabled",
			}
		},
	})
}
