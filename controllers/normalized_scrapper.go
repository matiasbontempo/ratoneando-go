package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"ratoneando/config"
	"ratoneando/products"
	"ratoneando/scrapers"
	"ratoneando/utils/cache"
)

func NormalizedScraper(c *gin.Context) {
	referer := c.Request.Referer()
	query := c.Query("q")

	// Get the client IP
	clientIp := c.Request.Header.Get("X-Envoy-External-Address")
	if clientIp == "" {
		clientIp = c.ClientIP()
	}

	// Check if the request is coming from a valid source
	if config.ENV == "release" && (referer == "" || !strings.Contains(referer, config.WEB_URL)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden."})
		return
	}

	// Check if the query is empty
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query is empty."})
		return
	}

	// Quick check to see if the query is valid
	if strings.ToLower(query) != query {
		fmt.Println("Uppercase query", query, c.Request.Header, clientIp)
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Forbidden",
		})
		return
	}

	// Cache
	cacheResponse, _ := cache.Get(query)
	if cacheResponse != "" {
		response := gin.H{}
		json.Unmarshal([]byte(cacheResponse), &response)

		c.Header("Cache-Control", "public, max-age="+config.RESPONSE_CACHE_EXPIRATION)
		c.Header("X-Cache", "HIT")

		c.JSON(http.StatusOK, response)

		return
	}

	// Middlewares
	// TODO

	// Run the scrapers
	var wg sync.WaitGroup
	var mu sync.Mutex

	scrappers := []func(string) ([]products.Schema, error){
		scrapers.Carrefour,
		scrapers.Coto,
		scrapers.DiaOnline,
		scrapers.Disco,
		scrapers.Farmacity,
		scrapers.Jumbo,
		scrapers.MasOnline,
		scrapers.MercadoLibre,
		scrapers.Vea,
	}

	type result struct {
		Products []products.Schema
		Error    error
	}

	results := make([]result, len(scrappers))

	for i, scrapper := range scrappers {
		wg.Add(1)
		go func(i int, scrapper func(string) ([]products.Schema, error)) {
			defer wg.Done()
			products, err := scrapper(query)
			mu.Lock()
			results[i] = result{Products: products, Error: err}
			mu.Unlock()
		}(i, scrapper)
	}

	wg.Wait()

	var failedScrappers []string
	var normalizedProducts []products.Schema

	for _, result := range results {
		if result.Error != nil {
			failedScrappers = append(failedScrappers, result.Error.Error())
		} else {
			normalizedProducts = append(normalizedProducts, result.Products...)
		}
	}

	filteredProducts := products.Fuzzy(normalizedProducts, query)
	sortedProducts := products.Sort(filteredProducts)

	response := gin.H{
		"products":       sortedProducts,
		"failedScrapers": failedScrappers,
		"timestamp":      time.Now(),
	}
	// Cache the response
	stringifiedResponse := []byte{}
	stringifiedResponse, _ = json.Marshal(response)

	cache.Set(query, string(stringifiedResponse), time.Duration(config.CORE_CACHE_EXPIRATION)*time.Second)

	// Return the products
	c.JSON(http.StatusOK, response)
}
