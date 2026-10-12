package scrapers

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func decodeCoto(t *testing.T, raw []byte) []CotoRawProduct {
	t.Helper()
	var response CotoResponseStructure
	if err := json.Unmarshal(raw, &response); err != nil {
		t.Fatal(err)
	}
	return cotoNormalize(response)
}

// coto_leche.json is a real search response (trimmed). The products are in Main[2]: a banner and the search
// adjustments come first. Reading the products from a fixed position once made every search panic.
func TestCotoFindsTheProductsWhateverTheirPositionUnderMain(t *testing.T) {
	raw, err := os.ReadFile("testdata/coto_leche.json")
	if err != nil {
		t.Fatal(err)
	}

	found := decodeCoto(t, raw)
	if len(found) != 2 {
		t.Fatalf("expected 2 products, got %d", len(found))
	}

	plain := cotoExtract(found[0])
	if plain.ID != "prod00253104" || plain.Name != "Leche Parcialmente Descremada Liviana Larga Vida LA SERENISIMA 1l" {
		t.Errorf("unexpected product %+v", plain)
	}
	if plain.Price != 3060 || plain.ListPrice != 3060 {
		t.Errorf("a product without discounts costs its list price, got price %v and list price %v", plain.Price, plain.ListPrice)
	}
	if plain.Unavailable {
		t.Error("the product has stock")
	}
	if want := "https://www.cotodigital.com.ar/sitios/cdigi/productos/leche-parcialmente-descremada-liviana-larga-vida-la-serenisima-1l/_/A-00253104-00253104-200"; plain.Link != want {
		t.Errorf("expected link %q, got %q", want, plain.Link)
	}
	if !strings.HasPrefix(plain.Image, "https://") {
		t.Errorf("unexpected image %q", plain.Image)
	}

	discounted := cotoExtract(found[1])
	if discounted.Price != 2141.28 || discounted.ListPrice != 2860 {
		t.Errorf("expected the discount price 2141.28 over the list price 2860, got %v over %v", discounted.Price, discounted.ListPrice)
	}
}

func TestCotoWithoutProductsGivesNothing(t *testing.T) {
	for name, raw := range map[string]string{
		"empty object":            `{}`,
		"no contents":             `{"contents":[]}`,
		"no main":                 `{"contents":[{}]}`,
		"blocks without products": `{"contents":[{"Main":[{"@type":"Category_SearchAdjustments"},{"contents":[{"records":[]}]}]}]}`,
	} {
		t.Run(name, func(t *testing.T) {
			if found := decodeCoto(t, []byte(raw)); len(found) != 0 {
				t.Errorf("expected no products, got %d", len(found))
			}
		})
	}
}

// A product that lacks parts must be skipped or extracted without panicking.
func TestCotoIncompleteProductsDoNotPanic(t *testing.T) {
	raw := `{"contents":[{"Main":[{"contents":[{"records":[
		{"attributes":{"product.repositoryId":["prod1"]}},
		{"attributes":{"product.displayName":["Sin id"]},"records":[{"attributes":{"sku.activePrice":["10"]}}]},
		{"attributes":{"product.displayName":["Sin precio"],"product.repositoryId":["prod2"]},"records":[{"attributes":{}}]},
		{"attributes":{"product.displayName":["Descuento roto"],"product.repositoryId":["prod3"]},"records":[{"attributes":{"sku.activePrice":["100"],"sku.quantity":["1"],"product.dtoDescuentos":["no es json"]}}]}
	]}]}]}]}`

	found := decodeCoto(t, []byte(raw))
	if len(found) != 2 {
		t.Fatalf("expected the 2 products that have an id and a record, got %d", len(found))
	}

	if noPrice := cotoExtract(found[0]); !noPrice.Unavailable {
		t.Errorf("a product without a price cannot be offered, got %+v", noPrice)
	}

	broken := cotoExtract(found[1])
	if broken.Unavailable || broken.Price != 100 {
		t.Errorf("an unreadable discount leaves the list price, got %+v", broken)
	}
}

func TestCotoOutOfStockProductIsUnavailable(t *testing.T) {
	raw := `{"contents":[{"Main":[{"contents":[{"records":[
		{"attributes":{"product.displayName":["Agotado"],"product.repositoryId":["prod1"]},"records":[{"attributes":{"sku.activePrice":["100"],"sku.quantity":["0"]}}]}
	]}]}]}]}`

	found := decodeCoto(t, []byte(raw))
	if len(found) != 1 || !cotoExtract(found[0]).Unavailable {
		t.Fatalf("expected one unavailable product, got %+v", found)
	}
}
