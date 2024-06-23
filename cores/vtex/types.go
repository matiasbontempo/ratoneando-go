package vtex

type CoreProps struct {
	BaseUrl string
	Source  string
	Query   string
	Raw     bool
}

type PropertyName string

const (
	PricePerUnit               PropertyName = "pricePerUnit"
	GramajeDeUnidadDeConsumo   PropertyName = "Gramaje de unidad de consumo"
	GramajeLeyendaDeConversión PropertyName = "Gramaje leyenda de conversión"
	GramajeDeUnidadDeMedida    PropertyName = "Gramaje de unidad de medida"
)

type Specification struct {
	Name         string
	OriginalName string
	Values       []string
}

type SpecificationGroup struct {
	Name           string
	OriginalName   string
	Specifications []Specification
}

type Property struct {
	Name   PropertyName
	Values []string
}

type ResponseProduct struct {
	CacheId          string
	ProductId        string
	Description      string
	ProductName      string
	ProductReference string
	Brand            string
	LinkText         string
	Categories       []string
	CategoryId       string
	PriceRange       *struct {
		SellingPrice struct {
			HighPrice float64
			LowPrice  float64
		}
		ListPrice struct {
			HighPrice float64
			LowPrice  float64
		}
	}
	SpecificationGroups []SpecificationGroup
	Properties          []Property
	Items               []struct {
		Name             string
		Ean              string
		MeassurementUnit string
		UnitMultiplier   float64
		Images           []struct {
			ImageUrl string
		}
	}
}

type ResponseStructure struct {
	Data struct {
		ProductSuggestions struct {
			Count      int
			Misspelled *string
			Operators  string
			Products   []ResponseProduct
		}
	}
}

type RawProduct struct {
	ResponseProduct
	Properties map[PropertyName]string
}
