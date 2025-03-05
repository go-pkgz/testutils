package testutils

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"
	
	"github.com/stretchr/testify/require"
)

func TestMockHTTPServer(t *testing.T) {
	// create a simple handler using ResponseRecorder for write error checking
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// In real request handling, the error is typically ignored as it's a disconnect which HTTP server handles
		// ResponseWriter interface doesn't have a way to check for errors in tests
		// nolint:errcheck // http.ResponseWriter errors are handled by the HTTP server
		w.Write([]byte(`{"status":"ok"}`))
	})

	// get the server URL and cleanup function
	serverURL, cleanup := MockHTTPServer(t, handler)
	defer cleanup() // this is redundant as t.Cleanup is also used, but demonstrates the pattern

	// make a request to the server
	resp, err := http.Get(serverURL + "/test")
	require.NoError(t, err, "Failed to make request")
	defer resp.Body.Close()

	// check response
	require.Equal(t, http.StatusOK, resp.StatusCode, "Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	require.Equal(t, "application/json", resp.Header.Get("Content-Type"), "Wrong Content-Type header")

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")

	expectedBody := `{"status":"ok"}`
	require.Equal(t, expectedBody, string(body), "Wrong response body")
}

func TestHTTPRequestCaptor(t *testing.T) {
	// create a test handler that will receive forwarded requests
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// nolint:errcheck // http.ResponseWriter errors are handled by the HTTP server
		w.Write([]byte("response"))
	})

	// create the request captor
	captor, handler := HTTPRequestCaptor(t, testHandler)

	// create a test server with the captor handler
	serverURL, _ := MockHTTPServer(t, handler)

	// make GET request
	_, err := http.Get(serverURL + "/get-path?param=value")
	require.NoError(t, err, "Failed to make GET request")

	// make POST request with a body
	postBody := `{"key":"value"}`
	_, err = http.Post(serverURL+"/post-path", "application/json", strings.NewReader(postBody))
	require.NoError(t, err, "Failed to make POST request")

	// make PUT request with different content type
	req, _ := http.NewRequest(http.MethodPut, serverURL+"/put-path", bytes.NewBuffer([]byte("text data")))
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Authorization", "Bearer token123")
	_, err = http.DefaultClient.Do(req)
	require.NoError(t, err, "Failed to make PUT request")

	// check the captured requests
	require.Equal(t, 3, captor.Len(), "Wrong number of captured requests")

	// check GET request
	getReq, ok := captor.GetRequest(0)
	require.True(t, ok, "Failed to get first request")
	require.Equal(t, http.MethodGet, getReq.Method, "Wrong request method")
	require.Equal(t, "/get-path", getReq.Path, "Wrong request path")

	// check POST request
	postReq, ok := captor.GetRequest(1)
	require.True(t, ok, "Failed to get second request")
	require.Equal(t, http.MethodPost, postReq.Method, "Wrong request method")
	require.Equal(t, "/post-path", postReq.Path, "Wrong request path")
	require.Equal(t, postBody, string(postReq.Body), "Wrong request body")

	// check PUT request with headers
	putReq, ok := captor.GetRequest(2)
	require.True(t, ok, "Failed to get third request")
	require.Equal(t, http.MethodPut, putReq.Method, "Wrong request method")
	require.Equal(t, "Bearer token123", putReq.Headers.Get("Authorization"), "Wrong Authorization header")
	require.Equal(t, "text/plain", putReq.Headers.Get("Content-Type"), "Wrong Content-Type header")

	// test GetRequests
	allRequests := captor.GetRequests()
	require.Equal(t, 3, len(allRequests), "Wrong number of requests from GetRequests")

	// test Reset
	captor.Reset()
	require.Equal(t, 0, captor.Len(), "Reset didn't clear requests")
}
