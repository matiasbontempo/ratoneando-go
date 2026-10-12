package controllers

import (
	"errors"
	"log"
	"sync"

	"ratoneando/products"
)

// scraper is one store a search goes through. The source is the name reported in failedScrapers,
// the same one the scraper puts on its products.
type scraper struct {
	source string
	run    func(string) ([]products.Schema, error)
}

type scraperResult struct {
	Products []products.Schema
	Error    error
}

// runScrapers runs every scraper at once and returns their results in the same order. A scraper
// that panics fails like one that returns an error: a panic must not look like a store with no
// products, because then the response is cached as complete and the failure goes unnoticed.
func runScrapers(scrapers []scraper, query string) []scraperResult {
	var wg sync.WaitGroup
	results := make([]scraperResult, len(scrapers)) // each goroutine only writes its own slot

	for i, s := range scrapers {
		wg.Add(1)
		go func(i int, s scraper) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Recovered from panic in scraper %s: %v", s.source, r)
					results[i] = scraperResult{Error: errors.New(s.source)}
				}
			}()

			found, err := s.run(query)
			if err != nil {
				log.Printf("Error executing scraper: %v", err)
			}

			results[i] = scraperResult{Products: found, Error: err}
		}(i, s)
	}

	wg.Wait()
	return results
}
