package api

import "ratoneando/product"

type CoreProps[ResponseStructure any] struct {
	Query         string
	BaseUrl       string
	SearchPattern func(string) string
	Source        string
	Extractor     func(ResponseStructure) []product.Schema
	Raw           bool
}
