package unit

import (
	"ratoneando/products"
	"testing"
)

func TestExtractUnit(t *testing.T) {
	tests := []struct {
		name           string
		prod           products.ExtendedSchema
		expectedUnit   string
		expectedFactor float64
	}{
		{
			name:           "Extracts unit from title",
			prod:           products.ExtendedSchema{Name: "Gaseosa Coca Cola 2,25 Lt", Unit: "un"},
			expectedUnit:   "LT",
			expectedFactor: 2.25,
		},
		{
			name:           "Keeps original values if units are missing from title",
			prod:           products.ExtendedSchema{Name: "Coca-Cola", Unit: "un"},
			expectedUnit:   "un",
			expectedFactor: 1,
		},
		{
			name:           "Normalizes unitFactor if unit is not the base one",
			prod:           products.ExtendedSchema{Name: "Coca-Cola 430 ml", Unit: "un"},
			expectedUnit:   "LT",
			expectedFactor: 0.43,
		},
		{
			name:           "Ignores other numbers in the title",
			prod:           products.ExtendedSchema{Name: "Coca-Cola 430 ml 2", Unit: "un"},
			expectedUnit:   "LT",
			expectedFactor: 0.43,
		},
		{
			name:           "Ignores number before the unit",
			prod:           products.ExtendedSchema{Name: "Gaseosa Coca Cola Creations Y3000 473 Ml", Unit: "un"},
			expectedUnit:   "LT",
			expectedFactor: 0.473,
		},
		{
			name:           "Allows no space between number and unit",
			prod:           products.ExtendedSchema{Name: "Gaseosa Coca Cola Creations Y3000 473Ml", Unit: "un"},
			expectedUnit:   "LT",
			expectedFactor: 0.473,
		},
		{
			name:           "Allows comma as decimal separator",
			prod:           products.ExtendedSchema{Name: "Coca-Cola 1,430kg x KG", Unit: "un"},
			expectedUnit:   "KG",
			expectedFactor: 1.43,
		},
		{
			name:           "Extracts UN as unit",
			prod:           products.ExtendedSchema{Name: "Coca-Cola 8 UN", Unit: "un"},
			expectedUnit:   "UN",
			expectedFactor: 8,
		},
		{
			name:           "Extracts UNI as unit",
			prod:           products.ExtendedSchema{Name: "Pañales Babysec talle XXG ultrasoft 8 uni", Unit: "un"},
			expectedUnit:   "UN",
			expectedFactor: 8,
		},
		{
			name:           "Extracts U as unit",
			prod:           products.ExtendedSchema{Name: "Pañales Pampers Babydry Xxg 54u", Unit: "un"},
			expectedUnit:   "UN",
			expectedFactor: 54,
		},
		{
			name:           "Extracts from X KG",
			prod:           products.ExtendedSchema{Name: "Asado X KG", Unit: "un"},
			expectedUnit:   "KG",
			expectedFactor: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unit, factor := ExtractUnit(tt.prod)
			if unit != tt.expectedUnit {
				t.Errorf("expected unit %s, got %s", tt.expectedUnit, unit)
			}
			if factor != tt.expectedFactor {
				t.Errorf("expected factor %f, got %f", tt.expectedFactor, factor)
			}
		})
	}
}
