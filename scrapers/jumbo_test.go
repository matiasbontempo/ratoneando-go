package scrapers

import (
	"encoding/json"
	"testing"

	"ratoneando/products"
	"ratoneando/unit"
)

// The ProductData strings below carry the keys Jumbo really sends (checked against live responses): products sold
// by weight have "measurement_unit":"kg" and a price per kilo, the others have "measurement_unit":"un".
const (
	soldByKilo  = `{"allow_notes":false,"allow_substitute":false,"cart_limit":"0","brandName":"LARIO","measurement_unit":"kg","unit_multiplier":0.1,"measurement_unit_un":"KG","unit_multiplier_un":1}`
	soldByUnits = `{"allow_notes":true,"allow_substitute":true,"cart_limit":"0","brandName":"ESPUÑA","measurement_unit":"un","unit_multiplier":1,"measurement_unit_un":"gr.","unit_multiplier_un":100}`
)

func decodeProductData(t *testing.T, raw string) ProductData {
	t.Helper()
	var data ProductData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		t.Fatal(err)
	}
	return data
}

func TestStoreUnit(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"Sold by weight has a price per kilo", soldByKilo, "KG"},
		{"Sold by unit leaves the unit to the name", soldByUnits, ""},
		{"No measurement data leaves the unit to the name", `{"brandName":"X"}`, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := storeUnit(decodeProductData(t, tt.raw)); got != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

// Real products: the slice size in the name must not divide a price that is already per kilo.
func TestUnitPriceOfWeightSoldProducts(t *testing.T) {
	tests := []struct {
		name          string
		productName   string
		price         float64
		raw           string
		wantUnit      string
		wantUnitPrice float64
	}{
		{"Sold by weight, 100 Grs in the name", "Bondiola de Cerdo Feteada 100 Grs Lario", 73590, soldByKilo, "KG", 73590},
		{"Sold by weight, 150 Grs in the name", "Bondiola de Cerdo Feteada 150 Grs Campo Austral", 48890, soldByKilo, "KG", 48890},
		{"Sold by weight, no size in the name", "Bondiola de Cerdo Fresca", 10599, soldByKilo, "KG", 10599},
		{"Sold by unit keeps reading the size from the name", "Bondiola Cerdo Feteada 100 Grs Espuña", 8100, soldByUnits, "KG", 81000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := unit.CalculateUnitInfo(products.ExtendedSchema{
				Name:  tt.productName,
				Price: tt.price,
				Unit:  storeUnit(decodeProductData(t, tt.raw)),
			})
			if schema.Unit != tt.wantUnit {
				t.Errorf("expected unit %q, got %q", tt.wantUnit, schema.Unit)
			}
			if diff := schema.UnitPrice - tt.wantUnitPrice; diff > 0.01 || diff < -0.01 {
				t.Errorf("expected unit price %.2f, got %.2f", tt.wantUnitPrice, schema.UnitPrice)
			}
		})
	}
}
