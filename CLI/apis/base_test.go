package apis

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetRequest(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request method is GET
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Respond with a JSON object
		response := map[string]string{"message": "Hello, World!"}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Fatalf("Error encoding response: %v", err)
		}
	}))
	defer server.Close()

	// Call the GetRequest function with the mock server's URL
	result, err := GetRequest(server.URL)
	if err != nil {
		t.Fatalf("GetRequest failed: %v", err)
	}

	// Assert the result
	expected := map[string]interface{}{"message": "Hello, World!"}
	if resultMap, ok := result.(map[string]interface{}); ok {
		if resultMap["message"] != expected["message"] {
			t.Errorf("Expected %v, got %v", expected, resultMap)
		}
	} else {
		t.Errorf("Expected a map[string]interface{}, got %T", result)
	}
}

func TestGetRequest_ErrorHandling(t *testing.T) {
	// Create a mock HTTP server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a server error
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Call the GetRequest function with the mock server's URL
	_, err := GetRequest(server.URL)
	if err == nil {
		t.Error("Expected an error, got nil")
	}
}

func TestGetRequest_InvalidJSON(t *testing.T) {
	// Create a mock HTTP server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Respond with invalid JSON
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, "invalid json")
	}))
	defer server.Close()

	// Call the GetRequest function with the mock server's URL
	_, err := GetRequest(server.URL)
	if err == nil {
		t.Error("Expected an error due to invalid JSON, got nil")
	}
}

func TestPostRequest_Success(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request method is POST
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check the Content-Type header
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type: application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Read the request body
		var requestBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Fatalf("Error decoding request body: %v", err)
		}

		// Respond with a JSON object
		response := map[string]interface{}{"message": "Hello, " + requestBody["name"].(string)}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Fatalf("Error encoding response: %v", err)
		}
	}))
	defer server.Close()

	// Call the PostRequest function with the mock server's URL and a request body
	requestBody := map[string]string{"name": "World"}
	result, err := PostRequest(server.URL, requestBody)
	if err != nil {
		t.Fatalf("PostRequest failed: %v", err)
	}

	// Assert the result
	expected := map[string]interface{}{"message": "Hello, World"}
	if resultMap, ok := result.(map[string]interface{}); ok {
		if resultMap["message"] != expected["message"] {
			t.Errorf("Expected %v, got %v", expected, resultMap)
		}
	} else {
		t.Errorf("Expected a map[string]interface{}, got %T", result)
	}
}

func TestPostRequest_InvalidRequestBody(t *testing.T) {
	// Call the PostRequest function with an invalid request body
	invalidRequestBody := func() {} // Functions cannot be marshaled to JSON
	_, err := PostRequest("http://example.com", invalidRequestBody)
	if err == nil {
		t.Error("Expected an error due to invalid request body, got nil")
	}
}

func TestPostRequest_ServerError(t *testing.T) {
	// Create a mock HTTP server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a server error
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Call the PostRequest function with the mock server's URL
	requestBody := map[string]string{"name": "World"}
	_, err := PostRequest(server.URL, requestBody)
	if err == nil {
		t.Error("Expected an error, got nil")
	}
}

func TestPostRequest_InvalidJSONResponse(t *testing.T) {
	// Create a mock HTTP server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Respond with invalid JSON
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, "invalid json")
	}))
	defer server.Close()

	// Call the PostRequest function with the mock server's URL
	requestBody := map[string]string{"name": "World"}
	_, err := PostRequest(server.URL, requestBody)
	if err == nil {
		t.Error("Expected an error due to invalid JSON response, got nil")
	}
}
