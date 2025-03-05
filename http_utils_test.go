package testutils

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestMockHTTPServer(t *testing.T) {
	// create a simple handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// get the server URL and cleanup function
	serverURL, cleanup := MockHTTPServer(t, handler)
	defer cleanup() // this is redundant as t.Cleanup is also used, but demonstrates the pattern

	// make a request to the server
	resp, err := http.Get(serverURL + "/test")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// check response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type %q, got %q", "application/json", contentType)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	expectedBody := `{"status":"ok"}`
	if string(body) != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, string(body))
	}
}

func TestHTTPRequestCaptor(t *testing.T) {
	// create a test handler that will receive forwarded requests
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response"))
	})

	// create the request captor
	captor, handler := HTTPRequestCaptor(t, testHandler)

	// create a test server with the captor handler
	serverURL, _ := MockHTTPServer(t, handler)

	// make GET request
	_, err := http.Get(serverURL + "/get-path?param=value")
	if err != nil {
		t.Fatalf("Failed to make GET request: %v", err)
	}

	// make POST request with a body
	postBody := `{"key":"value"}`
	_, err = http.Post(serverURL+"/post-path", "application/json", strings.NewReader(postBody))
	if err != nil {
		t.Fatalf("Failed to make POST request: %v", err)
	}

	// make PUT request with different content type
	req, _ := http.NewRequest(http.MethodPut, serverURL+"/put-path", bytes.NewBuffer([]byte("text data")))
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Authorization", "Bearer token123")
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make PUT request: %v", err)
	}

	// check the captured requests
	if captor.Len() != 3 {
		t.Errorf("Expected 3 captured requests, got %d", captor.Len())
	}

	// check GET request
	getReq, ok := captor.GetRequest(0)
	if !ok {
		t.Fatalf("Failed to get first request")
	}

	if getReq.Method != http.MethodGet {
		t.Errorf("Expected method %s, got %s", http.MethodGet, getReq.Method)
	}

	if getReq.Path != "/get-path" {
		t.Errorf("Expected path %s, got %s", "/get-path", getReq.Path)
	}

	// check POST request
	postReq, ok := captor.GetRequest(1)
	if !ok {
		t.Fatalf("Failed to get second request")
	}

	if postReq.Method != http.MethodPost {
		t.Errorf("Expected method %s, got %s", http.MethodPost, postReq.Method)
	}

	if postReq.Path != "/post-path" {
		t.Errorf("Expected path %s, got %s", "/post-path", postReq.Path)
	}

	if string(postReq.Body) != postBody {
		t.Errorf("Expected body %q, got %q", postBody, string(postReq.Body))
	}

	// check PUT request with headers
	putReq, ok := captor.GetRequest(2)
	if !ok {
		t.Fatalf("Failed to get third request")
	}

	if putReq.Method != http.MethodPut {
		t.Errorf("Expected method %s, got %s", http.MethodPut, putReq.Method)
	}

	authHeader := putReq.Headers.Get("Authorization")
	if authHeader != "Bearer token123" {
		t.Errorf("Expected Authorization header %q, got %q", "Bearer token123", authHeader)
	}

	contentType := putReq.Headers.Get("Content-Type")
	if contentType != "text/plain" {
		t.Errorf("Expected Content-Type header %q, got %q", "text/plain", contentType)
	}

	// test GetRequests
	allRequests := captor.GetRequests()
	if len(allRequests) != 3 {
		t.Errorf("Expected 3 requests from GetRequests, got %d", len(allRequests))
	}

	// test Reset
	captor.Reset()
	if captor.Len() != 0 {
		t.Errorf("Expected 0 requests after Reset, got %d", captor.Len())
	}
}
