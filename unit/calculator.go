package unit

import (
	"strings"

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
		Unit:      normalizeUnit(unit),
		UnitPrice: unitPrice,
	}
}

func computeUnitPrice(price, unitFactor float64, unit string, unitPrice float64) float64 {
	if unitPrice == 0 && price != 0 && unitFactor != 0 {
		unitPrice = price / unitFactor
	}

	rawUnit := normalizeRawUnit(unit)
	if rawUnit == "CC" || rawUnit == "ML" || rawUnit == "G" || rawUnit == "GR" {
		unitPrice *= 1000
	}

	if unitPrice == 0 {
		unitPrice = price
	}

	return unitPrice
}

func normalizeUnit(unit string) string {
	rawUnit := normalizeRawUnit(unit)
	if normalizedUnit, exists := unitMapper[rawUnit]; exists {
		return normalizedUnit
	}
	return units
}

func normalizeRawUnit(unit string) string {
	return strings.ToUpper(strings.ReplaceAll(unit, "[^A-Z]", ""))
}
