package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"ratoneando/config"
	"ratoneando/cores/vtex"
)

func main() {
	config.Init()

	query := "mayonesa"
	baseUrl := "https://www.carrefour.com.ar"
	url := baseUrl + "/_v/segment/graphql/v1/" + vtex.EncodeQuery(query)

	fmt.Println("Testing VTEX hash with URL:")
	fmt.Println(url)
	fmt.Println("\nUsing hash:", config.VTEX_SHA256_HASH)
	fmt.Println("Hash length:", len(config.VTEX_SHA256_HASH))

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	var errorCheck struct {
		Errors []struct {
			Message    string `json:"message"`
			Extensions struct {
				Code string `json:"code"`
			} `json:"extensions"`
			Name string `json:"name"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(body, &errorCheck); err == nil && len(errorCheck.Errors) > 0 {
		fmt.Println("\n❌ Hash validation FAILED!")
		fmt.Println("Error message:", errorCheck.Errors[0].Message)
		fmt.Println("\nYou may need to update the VTEX_SHA256_HASH in your .env file or config package.")
		os.Exit(1)
	}

	// Parse the response to check if we got products
	var response vtex.ResponseStructure
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		os.Exit(1)
	}

	// Check if we got products
	productCount := len(response.Data.ProductSuggestions.Products)
	if productCount > 0 {
		fmt.Println("\n✅ Hash validation SUCCESSFUL!")
		fmt.Printf("Found %d products for query '%s'\n", productCount, query)

		// Print the first product name
		if productCount > 0 {
			fmt.Println("\nFirst product found:")
			fmt.Println("Name:", response.Data.ProductSuggestions.Products[0].ProductName)
			fmt.Println("ID:", response.Data.ProductSuggestions.Products[0].ProductId)
		}
	} else {
		fmt.Println("\n⚠️ Hash seems valid but no products were found.")
		fmt.Println("This could be normal if the store doesn't have products matching 'mayonesa'.")
		fmt.Println("Try another query or store to confirm.")
	}
}
