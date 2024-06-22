package scrapers

import (
	"encoding/json"

	"ratoneando/cores/api"
	"ratoneando/product"
	"ratoneando/unit"
)

type ResponseProduct struct {
	ProductId   string   `json:"productId"`
	ProductName string   `json:"productName"`
	Link        string   `json:"link"`
	ProductData []string `json:"ProductData"`
	Items       []struct {
		Images []struct {
			ImageUrl string `json:"imageUrl"`
		} `json:"images"`
		Sellers []struct {
			CommertialOffer struct {
				Price                float64 `json:"Price"`
				ListPrice            float64 `json:"ListPrice"`
				PriceWithoutDiscount float64 `json:"PriceWithoutDiscount"`
				AvailableQuantity    int     `json:"AvailableQuantity"`
				IsAvailable          bool    `json:"IsAvailable"`
			} `json:"commertialOffer"`
		} `json:"sellers"`
	} `json:"items"`
}

type ResponseStructure []ResponseProduct

func JumboScraper(query string) ([]product.Schema, error) {
	return api.Core(api.CoreProps[ResponseStructure]{
		Query:         query,
		BaseUrl:       "https://www.jumbo.com.ar",
		SearchPattern: func(q string) string { return "/api/catalog_system/pub/products/search/?ft=" + q },
		Source:        "jumbo",
		Extractor: func(response ResponseStructure) []product.Schema {
			var normalizedProducts []product.Schema

			for _, rawProduct := range response {
				var productData struct {
					MeasurementUnitUn string  `json:"MeasurementUnit"`
					UnitMultiplierUn  float64 `json:"UnitMultiplier"`
				}

				json.Unmarshal([]byte(rawProduct.ProductData[0]), &productData)

				var extendedProduct product.ExtendedSchema = product.ExtendedSchema{
					ID:          rawProduct.ProductId,
					Source:      "jumbo",
					Name:        rawProduct.ProductName,
					Link:        rawProduct.Link,
					Image:       rawProduct.Items[0].Images[0].ImageUrl,
					Unavailable: !rawProduct.Items[0].Sellers[0].CommertialOffer.IsAvailable,
					Price:       rawProduct.Items[0].Sellers[0].CommertialOffer.Price,
					ListPrice:   rawProduct.Items[0].Sellers[0].CommertialOffer.ListPrice,
					Unit:        productData.MeasurementUnitUn,
					UnitFactor:  productData.UnitMultiplierUn,
				}

				normalizedProducts = append(normalizedProducts, unit.CalculateUnitInfo(extendedProduct))
			}

			return normalizedProducts
		},
	})
}
