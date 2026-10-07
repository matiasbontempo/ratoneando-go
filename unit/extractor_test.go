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

// Real product names found while auditing unit extraction against live store results.
func TestExtractUnitRealNames(t *testing.T) {
	tests := []struct {
		name           string
		prod           products.ExtendedSchema
		expectedUnit   string
		expectedFactor float64
	}{
		{
			name:           "A letter right after the number is not a unit (7 Up)",
			prod:           products.ExtendedSchema{Name: "Gaseosa 7 Up Sin Azucar 1.5lt", Unit: "un"},
			expectedUnit:   "LT",
			expectedFactor: 1.5,
		},
		{
			name:           "A grade number followed by a word is not a size (00000 Carrefour)",
			prod:           products.ExtendedSchema{Name: "Arroz largo fino 00000 Carrefour Classic en bolsa 1 kg."},
			expectedUnit:   "KG",
			expectedFactor: 1,
		},
		{
			name:           "A grade number followed by a word is not a size (00000 Cuquets)",
			prod:           products.ExtendedSchema{Name: "Arroz Integral 00000 Cuquets 1 Kg."},
			expectedUnit:   "KG",
			expectedFactor: 1,
		},
		{
			name:           "A number cannot start inside another number or word (N28 Lucchetti)",
			prod:           products.ExtendedSchema{Name: "Fideos tirabuzon N28 Lucchetti 500 g."},
			expectedUnit:   "KG",
			expectedFactor: 0.5,
		},
		{
			name:           "Reads a size glued to a standalone X (X170g)",
			prod:           products.ExtendedSchema{Name: "Galletitas Chocolinas X170g", Unit: "un"},
			expectedUnit:   "KG",
			expectedFactor: 0.17,
		},
		{
			name:           "Reads a size glued to a standalone X (X400ml)",
			prod:           products.ExtendedSchema{Name: "Shampoo Dove Hidratacion X400ml", Unit: "un"},
			expectedUnit:   "LT",
			expectedFactor: 0.4,
		},
		{
			name:           "Does not cut digits out of a multipack (24x285gr stays unrecognized)",
			prod:           products.ExtendedSchema{Name: "Ens Atun Verde La Campagnola 24x285gr", Unit: "un"},
			expectedUnit:   "un",
			expectedFactor: 1,
		},
		{
			name:           "Liquid size times count multiplies (473 Ml x 6 Un)",
			prod:           products.ExtendedSchema{Name: "Cerveza Amber Lager 473 Ml x 6 Un Patagonia", Unit: "un"},
			expectedUnit:   "LT",
			expectedFactor: 2.838,
		},
		{
			name:           "Liquid size times count multiplies (473 Cc x 6 Un)",
			prod:           products.ExtendedSchema{Name: "Cerveza Vera Ipa 473 Cc x 6 Un Patagonia", Unit: "un"},
			expectedUnit:   "LT",
			expectedFactor: 2.838,
		},
		{
			name:           "Weights are not multiplied by a count, it is ambiguous (100 Grs x 50 Un is the total)",
			prod:           products.ExtendedSchema{Name: "Mate Cocido 100 Grs x 50 Un Nobleza Gaucha", Unit: "un"},
			expectedUnit:   "KG",
			expectedFactor: 0.1,
		},
		{
			name:           "Recognizes POR KG",
			prod:           products.ExtendedSchema{Name: "Queso Cremón Cremoso Paquete Por Kg"},
			expectedUnit:   "KG",
			expectedFactor: 1,
		},
		{
			name:           "Recognizes POR KILO",
			prod:           products.ExtendedSchema{Name: "Queso Cremoso Cuisine & Co Por Kilo"},
			expectedUnit:   "KG",
			expectedFactor: 1,
		},
		{
			name:           "Still reads the usual spellings (Lts)",
			prod:           products.ExtendedSchema{Name: "Aceite de Girasol 1,5 Lts Natura", Unit: "un"},
			expectedUnit:   "LT",
			expectedFactor: 1.5,
		},
		{
			name:           "Still reads grams written as Grs",
			prod:           products.ExtendedSchema{Name: "Oblea Leche 4 Fingers 41.5 Grs Kitkat", Unit: "un"},
			expectedUnit:   "KG",
			expectedFactor: 0.0415,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unit, factor := ExtractUnit(tt.prod)
			if unit != tt.expectedUnit {
				t.Errorf("expected unit %s, got %s", tt.expectedUnit, unit)
			}
			if diff := factor - tt.expectedFactor; diff > 1e-9 || diff < -1e-9 {
				t.Errorf("expected factor %f, got %f", tt.expectedFactor, factor)
			}
		})
	}
}
