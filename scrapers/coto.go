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

type CotoResponseProduct struct {
	DetailsAction struct {
		RecordState string `json:"recordState"`
	} `json:"detailsAction"`
	Attributes struct {
		ProductDisplayName []string `json:"product.displayName"`
		ProductRepositoryId []string `json:"product.repositoryId"`
	} `json:"attributes"`
	Records []struct {
		Attributes struct {
			SkuReferencePrice []string `json:"sku.referencePrice"`
			SkuActivePrice []string `json:"sku.activePrice"`
			ProductContent []string `json:"product.CONTENIDO"`
			ProductMediumImageUrl []string `json:"product.mediumImage.url"`
			SkuQuantity []string `json:"sku.quantity"`
			ProductDiscounts []string `json:"product.dtoDescuentos"`
		} `json:"attributes"`
	} `json:"records"`
}

type CotoProductDiscounts []struct {
	precioDescuento string
}

type CotoRawProduct struct {
	CotoResponseProduct
	CotoProductDiscounts
}

type CotoResponseStructure struct {
	Contents []struct {
		Main []struct {
			Contents []struct {
				Records []CotoResponseProduct `json:"records"`
			} `json:"contents"`
		} `json:"Main"`
	} `json:"contents"`
}

func Coto(query string) ([]products.Schema, error) {
	return api.Core(api.CoreProps[CotoResponseStructure, CotoRawProduct]{
		Query:         query,
		BaseUrl:       "https://www.cotodigital.com.ar",
		SearchPattern: func(q string) string { return "/sitios/cdigi/categoria?Ntt=" + q + "&format=json" },
		Source:        "coto",
		Normalizer: func(response CotoResponseStructure) []CotoRawProduct {
			var normalizedProducts []CotoRawProduct

			for _, rawProduct := range response.Contents[0].Main[1].Contents[0].Records {
        var productData CotoProductDiscounts

				if len(rawProduct.Records[0].Attributes.ProductDiscounts) == 0 {
					continue
				}

				err := json.Unmarshal([]byte(rawProduct.Records[0].Attributes.ProductDiscounts[0]), &productData)

				if err != nil {
					logger.LogWarn(fmt.Sprintf("Error unmarshalling product data: %s", err))
				}

				normalizedProducts = append(normalizedProducts, CotoRawProduct{
					CotoResponseProduct: rawProduct,
					CotoProductDiscounts: productData,
				})
			}

			return normalizedProducts
		},
		Extractor: func(rawProduct CotoRawProduct) products.ExtendedSchema {
      listPrice, _ := strconv.ParseFloat(rawProduct.Records[0].Attributes.SkuActivePrice[0], 64)
      var price float64 = listPrice
      if len(rawProduct.CotoProductDiscounts) > 0 {
          price, _ = strconv.ParseFloat(rawProduct.CotoProductDiscounts[0].precioDescuento, 64)
      }

      fmt.Println(price)
      fmt.Println(listPrice)
			return products.ExtendedSchema{
				ID:          rawProduct.CotoResponseProduct.Attributes.ProductRepositoryId[0],
				Source:      "coto",
				Name:        rawProduct.CotoResponseProduct.Attributes.ProductDisplayName[0],
				Link:        strings.Replace(rawProduct.DetailsAction.RecordState, "?format=json", "", -1),
				Image:       rawProduct.Records[0].Attributes.ProductMediumImageUrl[0],
				Unavailable: rawProduct.Records[0].Attributes.SkuQuantity[0] == "0",
				Price:       price,
				ListPrice:   listPrice,
			}
		},
	})
}
