package vtex

import (
	"fmt"

	"ratoneando/cores/api"
	"ratoneando/product"
)

func Core(props CoreProps) ([]product.Schema, error) {
	return api.Core(api.CoreProps[ResponseStructure, RawProduct]{
		Query:         props.Query,
		BaseUrl:       props.BaseUrl + "/_v/segment/graphql/v1/",
		SearchPattern: EncodeQuery,
		Source:        props.Source,
		Normalizer: func(response ResponseStructure) []RawProduct {
			products := response.Data.ProductSuggestions.Products
			expandedProducts := make([]RawProduct, len(products))

			for i, product := range products {
				properties := make(map[PropertyName]string)
				for _, property := range product.Properties {
					properties[property.Name] = property.Values[0]
				}

				expandedProducts[i] = RawProduct{
					ResponseProduct: product,
					Properties:      properties,
				}
			}

			return expandedProducts
		},
		Extractor: func(normalizedProduct RawProduct) product.ExtendedSchema {
			return product.ExtendedSchema{
				ID:        normalizedProduct.ProductId,
				Name:      normalizedProduct.ProductName,
				Link:      fmt.Sprintf("%s/%s/p", props.BaseUrl, normalizedProduct.LinkText),
				Image:     normalizedProduct.Items[0].Images[0].ImageUrl,
				Price:     normalizedProduct.PriceRange.SellingPrice.LowPrice,
				ListPrice: normalizedProduct.PriceRange.ListPrice.LowPrice,
				Source:    props.Source,
			}
		},
	})
}
