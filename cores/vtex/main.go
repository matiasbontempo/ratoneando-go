package vtex

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"ratoneando/cores/api"
	"ratoneando/products"
)

// MAX_RESULTS is the amount of products requested to each store. VTEX caps
// this endpoint at 50 items per page.
const MAX_RESULTS = 50

func SearchPattern(query string) string {
	params := url.Values{}
	params.Set("ft", query)
	params.Set("_from", "0")
	params.Set("_to", fmt.Sprintf("%d", MAX_RESULTS-1))

	return "/api/catalog_system/pub/products/search/?" + params.Encode()
}

// firstAvailableSeller returns the offer of the first seller with stock,
// falling back to the first seller when none of them is available.
func firstAvailableSeller(sellers []Seller) (CommertialOffer, bool) {
	if len(sellers) == 0 {
		return CommertialOffer{}, false
	}

	for _, seller := range sellers {
		if seller.CommertialOffer.IsAvailable && seller.CommertialOffer.Price > 0 {
			return seller.CommertialOffer, true
		}
	}

	return sellers[0].CommertialOffer, false
}

func Core(props CoreProps) ([]products.Schema, error) {
	return api.Core(api.CoreProps[ResponseStructure, RawProduct]{
		Query:         props.Query,
		BaseUrl:       props.BaseUrl,
		SearchPattern: SearchPattern,
		Source:        props.Source,
		Normalizer: func(response ResponseStructure) []RawProduct {
			normalizedProducts := make([]RawProduct, 0, len(response))

			for _, rawProduct := range response {
				var productData ProductData

				// ProductData is optional, products without it still have
				// their unit extracted from the name.
				if len(rawProduct.ProductData) > 0 {
					json.Unmarshal([]byte(rawProduct.ProductData[0]), &productData)
				}

				normalizedProducts = append(normalizedProducts, RawProduct{
					ResponseProduct: rawProduct,
					ProductData:     productData,
				})
			}

			return normalizedProducts
		},
		Extractor: func(rawProduct RawProduct) products.ExtendedSchema {
			if len(rawProduct.Items) == 0 {
				return products.ExtendedSchema{Unavailable: true}
			}

			item := rawProduct.Items[0]

			offer, available := firstAvailableSeller(item.Sellers)
			if !available {
				return products.ExtendedSchema{Unavailable: true}
			}

			image := ""
			if len(item.Images) > 0 {
				image = item.Images[0].ImageUrl
			}

			link := rawProduct.Link
			if link == "" && rawProduct.LinkText != "" {
				link = fmt.Sprintf("%s/%s/p", strings.TrimSuffix(props.BaseUrl, "/"), rawProduct.LinkText)
			}

			return products.ExtendedSchema{
				ID:         rawProduct.ProductId,
				Source:     props.Source,
				Name:       rawProduct.ProductName,
				Link:       link,
				Image:      image,
				Price:      offer.Price,
				ListPrice:  offer.ListPrice,
				Unit:       rawProduct.MeasurementUnit,
				UnitFactor: rawProduct.UnitMultiplier,
			}
		},
	})
}
