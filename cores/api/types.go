package api

import "ratoneando/products"

type CoreProps[ResponseStructure any, NormalizedProduct any] struct {
	Query         string
	BaseUrl       string
	SearchPattern func(string) string
	Source        string
	Normalizer    func(ResponseStructure) []NormalizedProduct
	Extractor     func(NormalizedProduct) products.ExtendedSchema
	Raw           bool
}
