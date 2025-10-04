// Package client provides a Wormly API client with rate limiting and retry capabilities.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"syscall"
	"time"

	"golang.org/x/time/rate"
)

// Logger defines the interface for logging within the client.
type Logger interface {
	Printf(format string, v ...interface{})
}

// NoOpLogger is a logger that does nothing.
type NoOpLogger struct{}

// Printf implements the Logger interface but does nothing.
func (l NoOpLogger) Printf(format string, v ...interface{}) {}

// StdLogger wraps the standard library logger.
type StdLogger struct {
	logger *log.Logger
}

// Printf implements the Logger interface using the standard library logger.
func (l StdLogger) Printf(format string, v ...interface{}) {
	l.logger.Printf(format, v...)
}

// NewStdLogger creates a new standard library logger wrapper.
func NewStdLogger(logger *log.Logger) *StdLogger {
	return &StdLogger{logger: logger}
}

// Client wraps an HTTP client with Wormly-specific functionality.
type Client struct {
	httpClient        *http.Client
	apiKey            string
	baseURL           string
	userAgent         string
	limiter           *rate.Limiter
	maxRetries        int
	initialBackoff    time.Duration
	backoffMultiplier float64
	maxBackoff        time.Duration
	logger            Logger
	debugEnabled      bool
}

// New creates a new Wormly API client.
func New(httpClient *http.Client, apiKey, baseURL, userAgent string,
	requestsPerSecond float64, maxRetries int, initialBackoff time.Duration,
	backoffMultiplier float64, maxBackoff time.Duration, logger Logger, debugEnabled bool) (*Client, error) {

	// Create rate limiter
	limiter := rate.NewLimiter(rate.Limit(requestsPerSecond), 1)

	if logger == nil {
		logger = NoOpLogger{}
	}

	return &Client{
		httpClient:        httpClient,
		apiKey:            apiKey,
		baseURL:           baseURL,
		userAgent:         userAgent,
		limiter:           limiter,
		maxRetries:        maxRetries,
		initialBackoff:    initialBackoff,
		backoffMultiplier: backoffMultiplier,
		maxBackoff:        maxBackoff,
		logger:            logger,
		debugEnabled:      debugEnabled,
	}, nil
}

// Do executes an HTTP request with rate limiting and retry logic.
func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	// Apply rate limiting
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter wait failed: %w", err)
	}

	// Inject headers if not already set
	if req.Header.Get("Authorization") == "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", c.userAgent)
	}
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	var lastErr error
	backoff := c.initialBackoff

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if c.debugEnabled {
			c.logger.Printf("Attempt %d: Making request to %s", attempt, req.URL)
		}

		// Make the request
		resp, err := c.httpClient.Do(req)
		if err != nil {
			// Check if it's a transient network error
			if isTransientNetworkError(err) {
				lastErr = err
				if attempt < c.maxRetries {
					if c.debugEnabled {
						c.logger.Printf("Transient network error: %v. Retrying in %v", err, backoff)
					}
					time.Sleep(backoff)
					backoff = c.calculateNextBackoff(backoff)
					continue
				}
			}
			return nil, err
		}

		// Check for transient HTTP errors
		if isTransientHTTPError(resp.StatusCode) {
			resp.Body.Close()
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
			if attempt < c.maxRetries {
				if c.debugEnabled {
					c.logger.Printf("Transient HTTP error: %v. Retrying in %v", lastErr, backoff)
				}
				time.Sleep(backoff)
				backoff = c.calculateNextBackoff(backoff)
				continue
			}
			return nil, lastErr
		}

		// Success or non-retryable error
		return resp, nil
	}

	return nil, fmt.Errorf("request failed after %d retries: %w", c.maxRetries, lastErr)
}

// calculateNextBackoff calculates the next backoff duration with exponential backoff.
func (c *Client) calculateNextBackoff(current time.Duration) time.Duration {
	next := time.Duration(float64(current) * c.backoffMultiplier)
	if next > c.maxBackoff {
		return c.maxBackoff
	}
	return next
}

// isTransientNetworkError checks if an error is a transient network error that should be retried.
func isTransientNetworkError(err error) bool {
	if err == nil {
		return false
	}

	// Check for network errors
	if netErr, ok := err.(net.Error); ok {
		return netErr.Timeout()
	}

	// Check for URL errors that wrap network errors
	if urlErr, ok := err.(*url.Error); ok {
		return isTransientNetworkError(urlErr.Err)
	}

	// Check for specific syscall errors
	if opErr, ok := err.(*net.OpError); ok {
		if syscallErr, ok := opErr.Err.(syscall.Errno); ok {
			switch syscallErr {
			case syscall.ECONNREFUSED, syscall.ECONNRESET, syscall.ETIMEDOUT:
				return true
			}
		}
	}

	return false
}

// isTransientHTTPError checks if an HTTP status code indicates a transient error.
func isTransientHTTPError(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	}
	return false
}

// makeFormRequest is a helper method for making form-encoded API requests (Wormly API style).
func (c *Client) makeFormRequest(ctx context.Context, command string, params map[string]string, result interface{}) error {
	// Apply rate limiting
	if err := c.limiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limiter wait failed: %w", err)
	}
	// Build form data
	data := url.Values{}
	data.Set("cmd", command)
	data.Set("key", c.apiKey)
	data.Set("response", "json")

	for key, value := range params {
		data.Set(key, value)
	}

	if c.debugEnabled {
		// Create a safe copy of params for logging (without API key)
		safeParams := make(map[string]string)
		for k, v := range params {
			safeParams[k] = v
		}
		c.logger.Printf("Wormly API request - command: %s, params: %+v", command, safeParams)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers for form data (don't use the generic headers from Do method)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", c.userAgent)

	var lastErr error
	backoff := c.initialBackoff

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if c.debugEnabled {
			c.logger.Printf("Attempt %d: Making form request to %s with command %s", attempt, c.baseURL, command)
		}

		// Make the request directly without using Do to avoid header conflicts
		resp, err := c.httpClient.Do(req)
		if err != nil {
			// Check if it's a transient network error
			if isTransientNetworkError(err) {
				lastErr = err
				if attempt < c.maxRetries {
					if c.debugEnabled {
						c.logger.Printf("Transient network error: %v. Retrying in %v", err, backoff)
					}
					time.Sleep(backoff)
					backoff = c.calculateNextBackoff(backoff)
					continue
				}
			}
			return err
		}

		// Check for transient HTTP errors
		if isTransientHTTPError(resp.StatusCode) {
			resp.Body.Close()
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
			if attempt < c.maxRetries {
				if c.debugEnabled {
					c.logger.Printf("Transient HTTP error: %v. Retrying in %v", lastErr, backoff)
				}
				time.Sleep(backoff)
				backoff = c.calculateNextBackoff(backoff)
				continue
			}
			return lastErr
		}

		// Success or non-retryable error
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			bodyBytes, _ := io.ReadAll(resp.Body)
			if c.debugEnabled {
				c.logger.Printf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
			}
			return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
		}

		if result != nil {
			// Read response body for potential debugging
			responseBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read response body: %w", err)
			}

			if c.debugEnabled {
				c.logger.Printf("Wormly API response: %s", string(responseBytes))
			}

			// Decode the response
			if err := json.Unmarshal(responseBytes, result); err != nil {
				return fmt.Errorf("failed to decode response: %w", err)
			}
		}

		return nil
	}

	return fmt.Errorf("request failed after %d retries: %w", c.maxRetries, lastErr)
}

// DebugLog logs a debug message if debug logging is enabled.
func (c *Client) DebugLog(format string, v ...interface{}) {
	if c.debugEnabled {
		c.logger.Printf("[DEBUG] "+format, v...)
	}
}
