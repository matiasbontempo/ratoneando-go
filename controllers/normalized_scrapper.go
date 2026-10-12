package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"ratoneando/history"
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
	if !isAllowedReferer(referer) {
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

		c.Header("Cache-Control", hitCacheControl(response))
		c.Header("X-Cache", "HIT")

		c.JSON(http.StatusOK, response)

		return
	}

	// Middlewares
	// TODO

	// Run the scrapers. The source is the name that shows up in failedScrapers.
	results := runScrapers([]scraper{
		{"carrefour", scrapers.Carrefour},
		{"coto", scrapers.Coto},
		{"diaonline", scrapers.DiaOnline},
		{"disco", scrapers.Disco},
		// {"farmacity", scrapers.Farmacity}, // disabled: a pharmacy, not part of the supermarket comparison
		{"jumbo", scrapers.Jumbo},
		{"masonline", scrapers.MasOnline},
		// {"mercadolibre", scrapers.MercadoLibre}, // disabled: a marketplace, prices are not comparable
		{"vea", scrapers.Vea},
	}, query)

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
	history.Record(filteredProducts)
	sortedProducts := products.Sort(filteredProducts)
	history.Annotate(sortedProducts)

	response := gin.H{
		"products":       sortedProducts,
		"failedScrapers": failedScrappers,
		"timestamp":      time.Now(),
	}
	// Cache the response
	stringifiedResponse, _ := json.Marshal(response)

	cache.Set(query, string(stringifiedResponse), cacheExpiration(failedScrappers))

	// Return the products
	c.JSON(http.StatusOK, response)
}
