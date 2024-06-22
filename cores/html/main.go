package html

import (
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"

	"ratoneando/product"
	"ratoneando/unit"
	"ratoneando/utils/logger"
)

type ElementWrapper struct {
	*goquery.Selection
}

type DocumentWrapper struct {
	*goquery.Document
}

func Core(props CoreProps) ([]product.Schema, error) {
	escapedQuery := url.PathEscape(props.Query)
	searchUrl := props.BaseUrl + props.SearchPattern(escapedQuery)

	resp, err := http.Get(searchUrl)
	if err != nil {
		logger.LogError("Failed to fetch the URL: " + props.Query + "@" + props.Source)
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		logger.LogError("Failed to parse the response body: " + props.Query + "@" + props.Source)
		return nil, err
	}

	if props.Raw {
		html, _ := doc.Html()
		return []product.Schema{{Source: props.Source, Name: html}}, nil
	}

	if props.SkipIfSelector != "" && doc.Find(props.SkipIfSelector).Length() > 0 {
		return []product.Schema{}, nil
	}

	products := []product.Schema{}
	doc.Find(props.ContainerSelector).Find(props.ProductSelector).Each(func(i int, s *goquery.Selection) {
		element := &ElementWrapper{s}
		product := props.Extractor(element, &DocumentWrapper{doc})
		product.Source = props.Source

		products = append(products, unit.CalculateUnitInfo(product))
	})

	return products, nil
}
