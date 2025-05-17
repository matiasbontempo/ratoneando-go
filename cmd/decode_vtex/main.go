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

	fmt.Println("Found sha256Hash:", sha256Hash)
}
