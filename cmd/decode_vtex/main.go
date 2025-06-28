package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
)

func main() {
	fmt.Println("=== VTEX Hash Extractor ===")
	fmt.Println("This tool extracts the SHA256 hash from a VTEX GraphQL URL.")
	fmt.Println("")
	fmt.Println("Instructions:")
	fmt.Println("1. Go to a VTEX store (e.g., carrefour.com.ar)")
	fmt.Println("2. Open browser dev tools (F12)")
	fmt.Println("3. Go to Network tab")
	fmt.Println("4. Search for any product")
	fmt.Println("5. Look for a request with the param 'productSuggestions'")
	fmt.Println("6. Copy the full URL and paste it below")
	fmt.Println("")
	fmt.Print("VTEX URL: ")

	reader := bufio.NewReader(os.Stdin)
	vtexURL, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		os.Exit(1)
	}

	vtexURL = strings.TrimSpace(vtexURL)

	if vtexURL == "" {
		fmt.Println("Missing VTEX URL")
		os.Exit(1)
	}

	parsedURL, err := url.Parse(vtexURL)
	if err != nil {
		fmt.Printf("Error parsing URL: %v\n", err)
		os.Exit(1)
	}

	queryParams := parsedURL.Query()
	extensionsEncoded := queryParams.Get("extensions")
	if extensionsEncoded == "" {
		fmt.Println("No 'extensions' parameter found in the URL")
		os.Exit(1)
	}

	operationName := queryParams.Get("operationName")
	if operationName != "productSuggestions" {
		fmt.Println("Make sure the operationName is 'productSuggestions'")
		os.Exit(1)
	}

	extensionsDecoded, err := url.QueryUnescape(extensionsEncoded)
	if err != nil {
		fmt.Printf("Error decoding extensions parameter: %v\n", err)
		os.Exit(1)
	}

	var extensions map[string]interface{}
	err = json.Unmarshal([]byte(extensionsDecoded), &extensions)
	if err != nil {
		fmt.Printf("Error parsing extensions JSON: %v\n", err)
		os.Exit(1)
	}

	persistedQuery, ok := extensions["persistedQuery"].(map[string]interface{})
	if !ok {
		fmt.Println("No 'persistedQuery' found in extensions")
		os.Exit(1)
	}

	sha256Hash, ok := persistedQuery["sha256Hash"].(string)
	if !ok {
		fmt.Println("No 'sha256Hash' found in persistedQuery")
		os.Exit(1)
	}

	fmt.Println("\n=== HASH FOUND ===")
	fmt.Println("SHA256 Hash:", sha256Hash)
	fmt.Println("Hash length:", len(sha256Hash))
	fmt.Println("\n=== NEXT STEPS ===")
	fmt.Println("1. Copy the hash above")
	fmt.Println("2. Update your .env file:")
	fmt.Printf("   VTEX_SHA256_HASH = %s\n", sha256Hash)
	fmt.Println("3. Run the verify script to test:")
	fmt.Println("   go run ./cmd/verify_vtex")
	fmt.Println("\nNote: This hash may expire periodically and need to be updated.")
}
