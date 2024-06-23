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
	if prod.Unit != "" && prod.Unit != "un" {
		return prod.Unit, 1
	}

	title := strings.ToLower(prod.Name)
	matches, _ := unitRegex.FindStringMatch(title)

	if matches == nil {
		if strings.Contains(title, "x kg") {
			return kilo, 1
		}
		return prod.Unit, 1
	}

	content := matches.Groups()[1].String()
	unit := matches.Groups()[2].String()

	parsedContent, err := strconv.ParseFloat(strings.ReplaceAll(content, ",", "."), 64)
	if err != nil {
		fmt.Println("Error parsing content:", err)
		return prod.Unit, 1
	}

	unitFactor := computeUnitFactor(parsedContent, unit)
	return unitMapper[strings.ToUpper(unit)], unitFactor
}

func computeUnitFactor(content float64, unit string) float64 {
	if unit == "cc" || unit == "c" || unit == "ml" || unit == "g" || unit == "gr" {
		return content / 1000
	}
	return content
}
