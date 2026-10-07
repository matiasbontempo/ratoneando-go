package products

import (
	"encoding/json"
	"testing"
)

// The web app reads these exact field names.
func TestPriceHistoryJSONShape(t *testing.T) {
	product := Schema{
		ID: "1", Source: "disco", Name: "Leche", Price: 1000, ListPrice: 1200, UnitFactor: 1,
		History: &PriceHistory{
			Points: 8, WindowDays: 40, Typical: 1100, Low: 900, High: 1200, ChangePct: -9.1, ChangeDays: 30,
		},
	}

	raw, err := json.Marshal(product)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}

	history, ok := decoded["history"].(map[string]any)
	if !ok {
		t.Fatalf("expected a history object in %s", raw)
	}
	for _, field := range []string{"points", "windowDays", "typical", "low", "high", "changePct", "changeDays"} {
		if _, present := history[field]; !present {
			t.Fatalf("missing field %q in %s", field, raw)
		}
	}
	for _, internal := range []string{"ListPrice", "UnitFactor", "listPrice", "unitFactor"} {
		if _, leaked := decoded[internal]; leaked {
			t.Fatalf("internal field %q must not be serialized: %s", internal, raw)
		}
	}

	withoutHistory, _ := json.Marshal(Schema{ID: "1", Name: "Leche"})
	if string(withoutHistory) != `{"id":"1","name":"Leche"}` {
		t.Fatalf("a product without history must serialize as before, got %s", withoutHistory)
	}
}
