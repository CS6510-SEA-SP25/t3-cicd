package apis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	// "time"
)

// Global HTTP client with timeout
var httpClient = &http.Client{
	// Timeout: time.Second * 10, // Set a timeout of 10 seconds
}

// GET requests
func GetRequest(url string) (interface{}, error) {
	// Make a GET request
	response, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error making GET request: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	return result, nil
}

// POST requests
func PostRequest(url string, requestBody interface{}) (interface{}, error) {
	// Convert the request body to JSON
	jsonData, err := json.Marshal(requestBody)

	if err != nil {
		return nil, fmt.Errorf("error marshaling JSON: %w", err)
	}

	response, err := httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error making POST request: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	return result, nil
}
