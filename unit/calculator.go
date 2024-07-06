package unit

import (
	"ratoneando/products"
)

func CalculateUnitInfo(prod products.ExtendedSchema) products.Schema {
	unit, unitFactor := ExtractUnit(prod)
	unitPrice := computeUnitPrice(prod.Price, unitFactor, unit, prod.UnitPrice)

	return products.Schema{
		ID:        prod.ID,
		Name:      prod.Name,
		Link:      prod.Link,
		Image:     prod.Image,
		Source:    prod.Source,
		Price:     prod.Price,
		Unit:      unit,
		UnitPrice: unitPrice,
	}
}

func computeUnitPrice(price, unitFactor float64, unit string, unitPrice float64) float64 {
	if unitPrice == 0 && price != 0 && unitFactor != 0 {
		unitPrice = price / unitFactor
	}

	if unit == "CC" || unit == "ML" || unit == "G" || unit == "GR" {
		unitPrice *= 1000
	}

	if unitPrice == 0 {
		unitPrice = price
	}

	return unitPrice
}
