package controllers

import (
	"errors"
	"testing"

	"ratoneando/products"
)

func TestRunScrapersKeepsTheOrderAndTheProducts(t *testing.T) {
	results := runScrapers([]scraper{
		{"uno", func(string) ([]products.Schema, error) { return []products.Schema{{ID: "1"}}, nil }},
		{"dos", func(q string) ([]products.Schema, error) { return []products.Schema{{ID: q}}, nil }},
	}, "leche")

	if len(results) != 2 || results[0].Products[0].ID != "1" || results[1].Products[0].ID != "leche" {
		t.Fatalf("unexpected results %+v", results)
	}
	if results[0].Error != nil || results[1].Error != nil {
		t.Fatalf("no scraper failed, got %v and %v", results[0].Error, results[1].Error)
	}
}

// A panicking scraper used to leave an empty result with no error, so the response looked
// complete and was cached for hours without that store.
func TestRunScrapersReportsAPanicAsAFailureOfThatScraper(t *testing.T) {
	results := runScrapers([]scraper{
		{"coto", func(string) ([]products.Schema, error) {
			var empty []int
			_ = empty[0]
			return nil, nil
		}},
		{"vea", func(string) ([]products.Schema, error) { return []products.Schema{{ID: "ok"}}, nil }},
	}, "leche")

	if results[0].Error == nil || results[0].Error.Error() != "coto" {
		t.Fatalf("expected the panic to fail as %q, got %v", "coto", results[0].Error)
	}
	if len(results[0].Products) != 0 {
		t.Errorf("a panicked scraper has no products, got %v", results[0].Products)
	}
	if results[1].Error != nil || len(results[1].Products) != 1 {
		t.Errorf("the other scraper is not affected, got %+v", results[1])
	}
}

func TestRunScrapersKeepsTheErrorOfAFailingScraper(t *testing.T) {
	results := runScrapers([]scraper{
		{"jumbo", func(string) ([]products.Schema, error) { return nil, errors.New("jumbo") }},
	}, "leche")

	if results[0].Error == nil || results[0].Error.Error() != "jumbo" {
		t.Fatalf("unexpected error %v", results[0].Error)
	}
}
