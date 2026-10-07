package products

import "testing"

func names(list []Schema) []string {
	out := make([]string, len(list))
	for i, p := range list {
		out[i] = p.Name
	}
	return out
}

func TestFuzzyTreatsHyphensAsSpaces(t *testing.T) {
	list := []Schema{
		{ID: "1", Name: "Gaseosa Coca Cola 2,25 L"},
		{ID: "2", Name: "Gaseosa Coca-Cola Zero 1,5 L"},
		{ID: "3", Name: "Galletitas Tipo Cracker"},
	}

	for _, query := range []string{"coca cola", "coca-cola"} {
		got := Fuzzy(list, query)
		if len(got) != 2 {
			t.Fatalf("query %q: expected 2 matches, got %v", query, names(got))
		}
	}
}

func TestFuzzyDropsUnrelatedProducts(t *testing.T) {
	list := []Schema{{ID: "1", Name: "Galletitas Tipo Cracker"}}

	if got := Fuzzy(list, "aceite"); len(got) != 0 {
		t.Fatalf("expected no match, got %v", names(got))
	}
}

func TestFuzzyKeepsNearlyIdenticalNames(t *testing.T) {
	list := []Schema{
		{ID: "1", Name: "Oreo"},
		{ID: "2", Name: "Leche 1L"},
		{ID: "3", Name: "Leche Entera La Serenisima 1 L"},
	}

	if got := Fuzzy(list, "oreo"); len(got) != 1 {
		t.Fatalf("expected the exact name to be kept, got %v", names(got))
	}
	if got := Fuzzy(list, "leche"); len(got) != 2 {
		t.Fatalf("expected both milk products, got %v", names(got))
	}
}
