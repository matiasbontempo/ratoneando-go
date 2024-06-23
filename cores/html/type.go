package html

import "ratoneando/products"

type CoreProps struct {
	BaseUrl           string
	Query             string
	Source            string
	Raw               bool
	SearchPattern     func(string) string
	ContainerSelector string
	ProductSelector   string
	SkipIfSelector    string
	Extractor         func(element *ElementWrapper, doc *DocumentWrapper) products.ExtendedSchema
}
