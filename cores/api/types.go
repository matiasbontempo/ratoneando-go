package api

import "ratoneando/product"

type CoreProps[ResponseStructure any, NormalizedProduct any] struct {
	Query         string
	BaseUrl       string
	SearchPattern func(string) string
	Source        string
	Normalizer    func(ResponseStructure) []NormalizedProduct
	Extractor     func(NormalizedProduct) product.ExtendedSchema
	Raw           bool
}
