package client

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	httpClient := &http.Client{Timeout: 30 * time.Second}
	client, err := New(
		httpClient,
		"test-api-key",
		"https://api.example.com",
		"test-agent/1.0",
		10.0,
		3,
		time.Second,
		2.0,
		30*time.Second,
		NoOpLogger{}, // logger
		false,        // debug
	)

	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	if client == nil {
		t.Fatal("New() returned nil client")
	}

	if client.apiKey != "test-api-key" {
		t.Errorf("Expected apiKey to be 'test-api-key', got %q", client.apiKey)
	}

	if client.baseURL != "https://api.example.com" {
		t.Errorf("Expected baseURL to be 'https://api.example.com', got %q", client.baseURL)
	}

	if client.userAgent != "test-agent/1.0" {
		t.Errorf("Expected userAgent to be 'test-agent/1.0', got %q", client.userAgent)
	}

	if client.maxRetries != 3 {
		t.Errorf("Expected maxRetries to be 3, got %d", client.maxRetries)
	}

	if client.initialBackoff != time.Second {
		t.Errorf("Expected initialBackoff to be 1s, got %v", client.initialBackoff)
	}

	if client.backoffMultiplier != 2.0 {
		t.Errorf("Expected backoffMultiplier to be 2.0, got %f", client.backoffMultiplier)
	}

	if client.maxBackoff != 30*time.Second {
		t.Errorf("Expected maxBackoff to be 30s, got %v", client.maxBackoff)
	}
}

func TestClient_Do_HeaderInjection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers are injected correctly
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-api-key" {
			t.Errorf("Expected Authorization header 'Bearer test-api-key', got %q", auth)
		}

		ua := r.Header.Get("User-Agent")
		if ua != "test-agent/1.0" {
			t.Errorf("Expected User-Agent header 'test-agent/1.0', got %q", ua)
		}

		ct := r.Header.Get("Content-Type")
		if ct != "application/json" {
			t.Errorf("Expected Content-Type header 'application/json', got %q", ct)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := New(
		&http.Client{},
		"test-api-key",
		server.URL,
		"test-agent/1.0",
		10.0,
		3,
		time.Millisecond,
		2.0,
		time.Second,
		NoOpLogger{}, // logger
		false,        // debug
	)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	req, err := http.NewRequest("GET", server.URL+"/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	ctx := t.Context()
	resp, err := client.Do(ctx, req)
	if err != nil {
		t.Fatalf("Do() returned error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestClient_Do_RateLimiting(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Set a very low rate limit (1 request per 100ms)
	client, err := New(
		&http.Client{},
		"test-api-key",
		server.URL,
		"test-agent/1.0",
		10.0, // 10 requests per second = 1 request per 100ms
		0,    // No retries for this test
		time.Millisecond,
		2.0,
		time.Second,
		NoOpLogger{}, // logger
		false,        // debug
	)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	ctx := t.Context()
	start := time.Now()

	// Make 3 requests
	for i := 0; i < 3; i++ {
		req, err := http.NewRequest("GET", server.URL+"/test", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		resp, err := client.Do(ctx, req)
		if err != nil {
			t.Fatalf("Do() returned error: %v", err)
		}
		resp.Body.Close()
	}

	elapsed := time.Since(start)

	// With 10 RPS, 3 requests should take at least 200ms (first is immediate, second after 100ms, third after 200ms)
	if elapsed < 190*time.Millisecond {
		t.Errorf("Rate limiting not working: 3 requests completed in %v, expected at least 190ms", elapsed)
	}

	if requestCount != 3 {
		t.Errorf("Expected 3 requests, got %d", requestCount)
	}
}

func TestClient_Do_RetryOnTransientErrors(t *testing.T) {
	tests := []struct {
		name         string
		statusCodes  []int
		expectRetry  bool
		finalSuccess bool
	}{
		{
			name:         "Retry on 429",
			statusCodes:  []int{429, 429, 200},
			expectRetry:  true,
			finalSuccess: true,
		},
		{
			name:         "Retry on 500",
			statusCodes:  []int{500, 500, 200},
			expectRetry:  true,
			finalSuccess: true,
		},
		{
			name:         "Retry on 502",
			statusCodes:  []int{502, 200},
			expectRetry:  true,
			finalSuccess: true,
		},
		{
			name:         "Retry on 503",
			statusCodes:  []int{503, 200},
			expectRetry:  true,
			finalSuccess: true,
		},
		{
			name:         "Retry on 504",
			statusCodes:  []int{504, 200},
			expectRetry:  true,
			finalSuccess: true,
		},
		{
			name:         "No retry on 404",
			statusCodes:  []int{404},
			expectRetry:  false,
			finalSuccess: true, // 404 should return response, not error
		},
		{
			name:         "No retry on 400",
			statusCodes:  []int{400},
			expectRetry:  false,
			finalSuccess: true, // 400 should return response, not error
		},
		{
			name:         "Exhaust retries",
			statusCodes:  []int{500, 500, 500, 500, 500},
			expectRetry:  true,
			finalSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if requestCount < len(tt.statusCodes) {
					w.WriteHeader(tt.statusCodes[requestCount])
				} else {
					w.WriteHeader(http.StatusOK)
				}
				requestCount++
			}))
			defer server.Close()

			client, err := New(
				&http.Client{},
				"test-api-key",
				server.URL,
				"test-agent/1.0",
				1000.0, // High rate limit to avoid rate limiting in tests
				3,      // 3 retries
				time.Millisecond,
				2.0,
				100*time.Millisecond,
				NoOpLogger{}, // logger
				false,        // debug
			)
			if err != nil {
				t.Fatalf("New() returned error: %v", err)
			}

			req, err := http.NewRequest("GET", server.URL+"/test", nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			ctx := t.Context()
			resp, err := client.Do(ctx, req)

			if tt.finalSuccess {
				if err != nil {
					t.Errorf("Expected success, got error: %v", err)
				}
				if resp == nil {
					t.Fatal("Expected response, got nil")
				}
				defer resp.Body.Close()

				// For non-retryable errors, check the actual status code returned
				if !tt.expectRetry {
					expectedStatus := tt.statusCodes[0]
					if resp.StatusCode != expectedStatus {
						t.Errorf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
					}
				} else {
					// For retryable errors that eventually succeed, expect 200
					if resp.StatusCode != http.StatusOK {
						t.Errorf("Expected status 200, got %d", resp.StatusCode)
					}
				}
			} else {
				if err == nil {
					if resp != nil {
						resp.Body.Close()
					}
					t.Error("Expected error, got nil")
				}
			}

			if tt.expectRetry {
				expectedRequests := len(tt.statusCodes)
				if tt.finalSuccess {
					// If it eventually succeeds, we expect one more request than status codes
					if requestCount != expectedRequests {
						t.Errorf("Expected %d requests with retries, got %d", expectedRequests, requestCount)
					}
				} else {
					// If it fails after retries, we expect maxRetries + 1 requests
					expectedRequests = 4 // 1 initial + 3 retries
					if requestCount != expectedRequests {
						t.Errorf("Expected %d requests (1 + 3 retries), got %d", expectedRequests, requestCount)
					}
				}
			} else {
				if requestCount != 1 {
					t.Errorf("Expected 1 request (no retries), got %d", requestCount)
				}
			}
		})
	}
}

func TestClient_Do_ExponentialBackoff(t *testing.T) {
	requestTimes := []time.Time{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestTimes = append(requestTimes, time.Now())
		w.WriteHeader(http.StatusInternalServerError) // Always fail to trigger retries
	}))
	defer server.Close()

	client, err := New(
		&http.Client{},
		"test-api-key",
		server.URL,
		"test-agent/1.0",
		1000.0,               // High rate limit
		3,                    // 3 retries
		50*time.Millisecond,  // 50ms initial backoff
		2.0,                  // Double each time
		500*time.Millisecond, // 500ms max backoff
		NoOpLogger{},         // logger
		false,                // debug
	)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	req, err := http.NewRequest("GET", server.URL+"/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	ctx := t.Context()
	start := time.Now()
	_, err = client.Do(ctx, req) // This should fail after retries

	if err == nil {
		t.Error("Expected error after retries, got nil")
	}

	if len(requestTimes) != 4 { // 1 initial + 3 retries
		t.Fatalf("Expected 4 requests, got %d", len(requestTimes))
	}

	// Check intervals between requests
	// First retry should be after ~50ms
	interval1 := requestTimes[1].Sub(requestTimes[0])
	if interval1 < 40*time.Millisecond || interval1 > 70*time.Millisecond {
		t.Errorf("First retry interval should be ~50ms, got %v", interval1)
	}

	// Second retry should be after ~100ms
	interval2 := requestTimes[2].Sub(requestTimes[1])
	if interval2 < 90*time.Millisecond || interval2 > 130*time.Millisecond {
		t.Errorf("Second retry interval should be ~100ms, got %v", interval2)
	}

	// Third retry should be after ~200ms
	interval3 := requestTimes[3].Sub(requestTimes[2])
	if interval3 < 180*time.Millisecond || interval3 > 230*time.Millisecond {
		t.Errorf("Third retry interval should be ~200ms, got %v", interval3)
	}

	totalDuration := time.Since(start)
	// Total should be at least 350ms (50+100+200)
	if totalDuration < 330*time.Millisecond {
		t.Errorf("Total duration should be at least 330ms, got %v", totalDuration)
	}
}

func TestClient_Do_BackoffCap(t *testing.T) {
	client, err := New(
		&http.Client{},
		"test-api-key",
		"https://api.example.com",
		"test-agent/1.0",
		1000.0,
		5,
		100*time.Millisecond, // 100ms initial
		3.0,                  // Triple each time
		200*time.Millisecond, // 200ms max (should cap the backoff)
		NoOpLogger{},         // logger
		false,                // debug
	)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	// Test backoff calculation
	backoff := client.initialBackoff
	expected := []time.Duration{
		100 * time.Millisecond, // initial
		200 * time.Millisecond, // 100 * 3 = 300, but capped at 200
		200 * time.Millisecond, // 200 * 3 = 600, but capped at 200
		200 * time.Millisecond, // still capped
	}

	for i, exp := range expected {
		if backoff != exp {
			t.Errorf("Backoff calculation %d: expected %v, got %v", i, exp, backoff)
		}
		backoff = client.calculateNextBackoff(backoff)
	}
}

func TestIsTransientNetworkError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "timeout error",
			err:      &net.OpError{Op: "dial", Err: &timeoutError{}},
			expected: true,
		},
		{
			name:     "temporary error",
			err:      &net.OpError{Op: "dial", Err: &temporaryError{}},
			expected: false, // Temporary errors are no longer considered transient per Go 1.18+ guidance
		},
		{
			name:     "connection refused",
			err:      &net.OpError{Op: "dial", Err: &net.DNSError{Err: "connection refused", IsTimeout: false, IsTemporary: false}},
			expected: false, // DNS errors are not syscall errors
		},
		{
			name:     "timeout in op error",
			err:      &net.OpError{Op: "dial", Err: &timeoutError{}},
			expected: true,
		},
		{
			name:     "generic error",
			err:      fmt.Errorf("some generic error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTransientNetworkError(tt.err)
			if result != tt.expected {
				t.Errorf("isTransientNetworkError(%v) = %v, expected %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestIsTransientHTTPError(t *testing.T) {
	tests := []struct {
		statusCode int
		expected   bool
	}{
		{http.StatusOK, false},
		{http.StatusBadRequest, false},
		{http.StatusNotFound, false},
		{http.StatusTooManyRequests, true},
		{http.StatusInternalServerError, true},
		{http.StatusBadGateway, true},
		{http.StatusServiceUnavailable, true},
		{http.StatusGatewayTimeout, true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("status_%d", tt.statusCode), func(t *testing.T) {
			result := isTransientHTTPError(tt.statusCode)
			if result != tt.expected {
				t.Errorf("isTransientHTTPError(%d) = %v, expected %v", tt.statusCode, result, tt.expected)
			}
		})
	}
}

// Test helper types for network error simulation.
type timeoutError struct{}

func (e *timeoutError) Error() string   { return "timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return false }

type temporaryError struct{}

func (e *temporaryError) Error() string   { return "temporary" }
func (e *temporaryError) Timeout() bool   { return false }
func (e *temporaryError) Temporary() bool { return true }
