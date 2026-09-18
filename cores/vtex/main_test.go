package vtex

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

const responseFixture = `[
  {
    "productId": "123",
    "productName": "Mayonesa Hellmanns 475 Gr",
    "linkText": "mayonesa-hellmanns-475-gr",
    "link": "https://store.test/mayonesa-hellmanns-475-gr/p",
    "ProductData": ["{\"MeasurementUnit\":\"kg\",\"UnitMultiplier\":0.475}"],
    "items": [
      {
        "images": [{ "imageUrl": "https://store.test/mayonesa.jpg" }],
        "sellers": [
          { "commertialOffer": { "Price": 1900, "ListPrice": 2000, "IsAvailable": true } }
        ]
      }
    ]
  },
  {
    "productId": "456",
    "productName": "Mayonesa Natura 237 Gr",
    "linkText": "mayonesa-natura-237-gr",
    "items": [
      {
        "images": [{ "imageUrl": "https://store.test/natura.jpg" }],
        "sellers": [
          { "commertialOffer": { "Price": 0, "ListPrice": 0, "IsAvailable": false } }
        ]
      }
    ]
  },
  {
    "productId": "789",
    "productName": "Mayonesa sin items"
  }
]`

func TestCore(t *testing.T) {
	var requestedPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(responseFixture))
	}))
	defer server.Close()

	result, err := Core(CoreProps{
		Query:   "mayonesa",
		BaseUrl: server.URL,
		Source:  "test-store",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No hash or token should be part of the request
	if requestedPath != "/api/catalog_system/pub/products/search/?_from=0&_to=49&ft=mayonesa" {
		t.Errorf("unexpected request path: %s", requestedPath)
	}

	// Unavailable products and products without items are dropped
	if len(result) != 1 {
		t.Fatalf("expected 1 product, got %d", len(result))
	}

	product := result[0]

	if product.ID != "123" {
		t.Errorf("unexpected id: %s", product.ID)
	}
	if product.Source != "test-store" {
		t.Errorf("unexpected source: %s", product.Source)
	}
	if product.Price != 1900 {
		t.Errorf("unexpected price: %v", product.Price)
	}
	if product.Link != "https://store.test/mayonesa-hellmanns-475-gr/p" {
		t.Errorf("unexpected link: %s", product.Link)
	}
	if product.Image != "https://store.test/mayonesa.jpg" {
		t.Errorf("unexpected image: %s", product.Image)
	}
	if product.Unit != "KG" {
		t.Errorf("unexpected unit: %s", product.Unit)
	}
	if product.UnitPrice != 4000 {
		t.Errorf("unexpected unit price: %v", product.UnitPrice)
	}
}

func TestCoreBuildsLinkFromLinkText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"productId":"1","productName":"Producto 1 Kg","linkText":"producto","items":[{"images":[{"imageUrl":"i.jpg"}],"sellers":[{"commertialOffer":{"Price":100,"IsAvailable":true}}]}]}]`))
	}))
	defer server.Close()

	result, err := Core(CoreProps{Query: "producto", BaseUrl: server.URL, Source: "test-store"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 product, got %d", len(result))
	}
	if result[0].Link != server.URL+"/producto/p" {
		t.Errorf("unexpected link: %s", result[0].Link)
	}
}
