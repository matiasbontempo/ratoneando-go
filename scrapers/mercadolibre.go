package scrapers

import (
	"ratoneando/cores/html"
	"ratoneando/products"
	"strconv"
	"strings"
)

func MercadoLibre(query string) ([]products.Schema, error) {
	return html.Core(html.CoreProps{
		Query:             query,
		BaseUrl:           "https://listado.mercadolibre.com.ar",
		Source:            "mercadolibre",
		SearchPattern:     func(q string) string { return "/supermercado/" + q + "#D[A:" + q + ",on]" },
		ContainerSelector: "div.ui-search-main",
		ProductSelector:   ".andes-card.ui-search-result",
		SkipIfSelector:    ".ui-search-zrp-disclaimer",
		Extractor: func(element *html.ElementWrapper, doc *html.DocumentWrapper) products.ExtendedSchema {
			id, _ := element.Attr("id")
			name := element.Find("h2.ui-search-item__title").Text()
			rawPrice := strings.ReplaceAll(element.Find("div.ui-search-price__second-line span.andes-money-amount__fraction").Text(), ".", "")
			price, _ := strconv.ParseFloat(rawPrice, 64)
			link, _ := element.Find("a.ui-search-link").Attr("href")
			image, _ := element.Find("img.ui-search-result-image__element").Attr("data-src")
			unavailable, _ := element.Find("div.ui-search-card-add-to-cart").Attr("disabled")

			return products.ExtendedSchema{
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
