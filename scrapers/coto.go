package scrapers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"ratoneando/cores/api"
	"ratoneando/products"
	"ratoneando/utils/logger"
)

const (
	cotoBaseUrl = "https://www.cotodigital.com.ar"
	// The record state is a path relative to the products section of the site.
	cotoProductsUrl = cotoBaseUrl + "/sitios/cdigi/productos"
)

type CotoResponseProduct struct {
	DetailsAction struct {
		RecordState string `json:"recordState"`
	} `json:"detailsAction"`
	Attributes struct {
		ProductDisplayName  []string `json:"product.displayName"`
		ProductRepositoryId []string `json:"product.repositoryId"`
	} `json:"attributes"`
	Records []struct {
		Attributes struct {
			SkuReferencePrice     []string `json:"sku.referencePrice"`
			SkuActivePrice        []string `json:"sku.activePrice"`
			ProductContent        []string `json:"product.CONTENIDO"`
			ProductMediumImageUrl []string `json:"product.mediumImage.url"`
			SkuQuantity           []string `json:"sku.quantity"`
			ProductDiscounts      []string `json:"product.dtoDescuentos"`
		} `json:"attributes"`
	} `json:"records"`
}

// CotoProductDiscounts is the JSON that Coto sends, as a string, in product.dtoDescuentos. PrecioDesc is the
// discounted price as a plain number ("2141.28"); precioDescuento carries a "$" and is not used.
type CotoProductDiscounts []struct {
	PrecioDesc string `json:"precioDesc"`
}

type CotoRawProduct struct {
	CotoResponseProduct
	CotoProductDiscounts
}

// CotoResponseStructure keeps every block under Main. The block that holds the products is not always at
// the same position (a banner or the search adjustments can come before it), so it is not picked by index.
type CotoResponseStructure struct {
	Contents []struct {
		Main []struct {
			Contents []struct {
				Records []CotoResponseProduct `json:"records"`
			} `json:"contents"`
		} `json:"Main"`
	} `json:"contents"`
}

func first(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func cotoNormalize(response CotoResponseStructure) []CotoRawProduct {
	var normalizedProducts []CotoRawProduct

	for _, content := range response.Contents {
		for _, main := range content.Main {
			for _, block := range main.Contents {
				for _, rawProduct := range block.Records {
					if len(rawProduct.Records) == 0 || first(rawProduct.Attributes.ProductRepositoryId) == "" {
						continue
					}

					var productData CotoProductDiscounts
					if discounts := first(rawProduct.Records[0].Attributes.ProductDiscounts); discounts != "" {
						if err := json.Unmarshal([]byte(discounts), &productData); err != nil {
							logger.LogWarn(fmt.Sprintf("Error unmarshalling product data: %s", err))
						}
					}

					normalizedProducts = append(normalizedProducts, CotoRawProduct{
						CotoResponseProduct:  rawProduct,
						CotoProductDiscounts: productData,
					})
				}
			}
		}
	}

	return normalizedProducts
}

func cotoExtract(rawProduct CotoRawProduct) products.ExtendedSchema {
	attributes := rawProduct.Records[0].Attributes

	listPrice, _ := strconv.ParseFloat(first(attributes.SkuActivePrice), 64)
	price := listPrice

	if len(rawProduct.CotoProductDiscounts) > 0 {
		discountPrice, _ := strconv.ParseFloat(rawProduct.CotoProductDiscounts[0].PrecioDesc, 64)
		if discountPrice > 0 && discountPrice < listPrice {
			price = discountPrice
		}
	}

	return products.ExtendedSchema{
		ID:          first(rawProduct.CotoResponseProduct.Attributes.ProductRepositoryId),
		Source:      "coto",
		Name:        first(rawProduct.CotoResponseProduct.Attributes.ProductDisplayName),
		Link:        cotoProductsUrl + strings.TrimSuffix(rawProduct.DetailsAction.RecordState, "?format=json"),
		Image:       first(attributes.ProductMediumImageUrl),
		Unavailable: first(attributes.SkuQuantity) == "0" || listPrice == 0,
		Price:       price,
		ListPrice:   listPrice,
	}
}

func Coto(query string) ([]products.Schema, error) {
	return api.Core(api.CoreProps[CotoResponseStructure, CotoRawProduct]{
		Query:         query,
		BaseUrl:       cotoBaseUrl,
		SearchPattern: func(q string) string { return "/sitios/cdigi/categoria?Ntt=" + q + "&format=json" },
		Source:        "coto",
		Normalizer:    cotoNormalize,
		Extractor:     cotoExtract,
	})
}
