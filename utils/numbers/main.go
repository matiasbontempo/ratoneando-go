package numbers

import (
	"strconv"
	"strings"
	"unicode"
)

func ParseMoney(input string) (float64, error) {
	var builder strings.Builder
	for _, char := range input {
		if unicode.IsDigit(char) || char == ',' || char == '.' {
			builder.WriteRune(char)
		}
	}

	filtered := builder.String()
	filtered = strings.Replace(filtered, ".", "", -1)
	filtered = strings.Replace(filtered, ",", ".", -1)

	result, err := strconv.ParseFloat(filtered, 64)
	if err != nil {
		return 0, err
	}

	return result, nil
}
