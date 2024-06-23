package html

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"

	"ratoneando/products"
	"ratoneando/unit"
	"ratoneando/utils/logger"
)

type ElementWrapper struct {
	*goquery.Selection
}

type DocumentWrapper struct {
	*goquery.Document
}

func Core(props CoreProps) ([]products.Schema, error) {
	escapedQuery := url.PathEscape(props.Query)
	searchUrl := props.BaseUrl + props.SearchPattern(escapedQuery)

	resp, err := http.Get(searchUrl)
	if err != nil {
		logger.LogError("Failed to fetch the URL: " + escapedQuery + "@" + props.Source)
		return nil, fmt.Errorf(props.Source)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		logger.LogError("Failed to parse the response body: " + escapedQuery + "@" + props.Source)
		return nil, fmt.Errorf(props.Source)
	}

	if props.Raw {
		html, _ := doc.Html()
		return []products.Schema{{Source: props.Source, Name: html}}, nil
	}

	if props.SkipIfSelector != "" && doc.Find(props.SkipIfSelector).Length() > 0 {
		return []products.Schema{}, nil
	}

	products := []products.Schema{}
	doc.Find(props.ContainerSelector).Find(props.ProductSelector).Each(func(i int, s *goquery.Selection) {
		element := &ElementWrapper{s}
		product := props.Extractor(element, &DocumentWrapper{doc})
		product.Source = props.Source

		products = append(products, unit.CalculateUnitInfo(product))
	})

	return products, nil
}
