package unit

import (
	"fmt"
	"strconv"
	"strings"

	"ratoneando/products"

	"github.com/dlclark/regexp2"
)

var unitRegex = regexp2.MustCompile(`(?<!\p{L})([0-9]+(?:[.,][0-9]{1,3})?) ?(l|lt|cc|ml|k|kg|g|c|u|un|uni|ud)`, 0)

func ExtractUnit(prod products.ExtendedSchema) (string, float64) {
	// When the store reports the unit, normalize it and keep the multiplier
	// it comes with (eg. "kg" with a 0.475 multiplier for a 475 gr product).
	if prod.Unit != "" && strings.ToLower(prod.Unit) != "un" {
		unitFactor := prod.UnitFactor
		if unitFactor == 0 {
			unitFactor = 1
		}

		if mappedUnit, ok := unitMapper[strings.ToUpper(prod.Unit)]; ok {
			return mappedUnit, unitFactor
		}

		return prod.Unit, unitFactor
	}

	title := strings.ToLower(prod.Name)
	matches, err := unitRegex.FindStringMatch(title)

	if err != nil || matches == nil {
		if strings.Contains(title, "x kg") {
			return kilo, 1
		}
		return prod.Unit, 1
	}

	value := matches.Groups()[1].String()
	unit := matches.Groups()[2].String()

	parsedValue, err := strconv.ParseFloat(strings.ReplaceAll(value, ",", "."), 64)
	if err != nil {
		fmt.Println("Error parsing content:", err)
		return prod.Unit, 1
	}

	unitFactor := computeUnitFactor(parsedValue, unit)
	return unitMapper[strings.ToUpper(unit)], unitFactor
}

func computeUnitFactor(content float64, unit string) float64 {
	if unit == "cc" || unit == "c" || unit == "ml" || unit == "g" || unit == "gr" {
		return content / 1000
	}
	return content
}
