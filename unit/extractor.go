package unit

import (
	"fmt"
	"strconv"
	"strings"

	"ratoneando/products"

	"github.com/dlclark/regexp2"
)

// A size is a number followed by a unit. Two rules keep it from matching things that merely look like one:
//   - the number cannot start inside another number or word ("285gr" must not become "85g", "N28 Lucchetti"
//     must not become "8 l"). The one exception is a standalone x right before it ("X400ml", "x170g");
//   - the unit cannot be followed by a letter ("7 Up" is not 7 units, "00000 Carrefour" is not 0 cc).
const (
	sizeStart  = `(?:(?<![\p{L}\d.,])|(?<=(?<![\p{L}\d])x))([0-9]+(?:[.,][0-9]{1,3})?) ?`
	sizeUnits  = `litros?|lts?|l|cc|ml|kilos?|kgs?|k|gramos?|grs?|g|c|unidades|unidad|unid|uni|uds?|un|u`
	liquidOnly = `litros?|lts?|l|cc|ml`
	wordEnd    = `(?![\p{L}])`
)

var (
	unitRegex = regexp2.MustCompile(sizeStart+`(`+sizeUnits+`)`+wordEnd, 0)

	// "473 ml x 6 un": the size is per unit, so the pack holds count times that size. Only for liquids,
	// where it is unambiguous. For weights the same wording is sometimes the total ("100 g x 50 un").
	sizeTimesCountRegex = regexp2.MustCompile(sizeStart+`(`+liquidOnly+`)`+wordEnd+` ?(?:x|por) ?([0-9]{1,2}) ?(?:unidades|unid|uni|un|u)`+wordEnd, 0)

	perKiloRegex = regexp2.MustCompile(`(?<![\p{L}])(?:x|por) ?(?:kg|kilos?)`+wordEnd, 0)
)

const maxPackCount = 24

var unitAliases = map[string]string{
	"litro": "l", "litros": "l", "lt": "l", "lts": "l",
	"kilo": "k", "kilos": "k", "kg": "k", "kgs": "k",
	"gramo": "g", "gramos": "g", "gr": "g", "grs": "g",
	"unidad": "un", "unidades": "un", "unid": "un",
	"uds": "ud",
}

func canonicalUnit(unit string) string {
	if alias, ok := unitAliases[unit]; ok {
		return alias
	}
	return unit
}

func ExtractUnit(prod products.ExtendedSchema) (string, float64) {
	if prod.Unit != "" && prod.Unit != "un" {
		return prod.Unit, 1
	}

	title := strings.ToLower(prod.Name)

	if pack, err := sizeTimesCountRegex.FindStringMatch(title); err == nil && pack != nil {
		size, sizeErr := parseNumber(pack.Groups()[1].String())
		count, countErr := strconv.Atoi(pack.Groups()[3].String())
		if sizeErr == nil && countErr == nil && count >= 2 && count <= maxPackCount {
			unit := canonicalUnit(pack.Groups()[2].String())
			return unitMapper[strings.ToUpper(unit)], computeUnitFactor(size, unit) * float64(count)
		}
	}

	matches, err := unitRegex.FindStringMatch(title)

	if err != nil || matches == nil {
		if perKilo, perKiloErr := perKiloRegex.MatchString(title); perKiloErr == nil && perKilo {
			return kilo, 1
		}
		return prod.Unit, 1
	}

	parsedValue, err := parseNumber(matches.Groups()[1].String())
	if err != nil {
		fmt.Println("Error parsing content:", err)
		return prod.Unit, 1
	}

	unit := canonicalUnit(matches.Groups()[2].String())
	unitFactor := computeUnitFactor(parsedValue, unit)
	return unitMapper[strings.ToUpper(unit)], unitFactor
}

func parseNumber(value string) (float64, error) {
	return strconv.ParseFloat(strings.ReplaceAll(value, ",", "."), 64)
}

func computeUnitFactor(content float64, unit string) float64 {
	if unit == "cc" || unit == "c" || unit == "ml" || unit == "g" || unit == "gr" {
		return content / 1000
	}
	return content
}
