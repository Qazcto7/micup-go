package micup

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// APIError represents an error returned by the Micup API.
type APIError struct {
	StatusCode int    `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("micup api error: status=%d code=%s message=%s", e.StatusCode, e.Code, e.Message)
}

// RateLimiter manages route-based rate limits.
type RateLimiter struct {
	mu      sync.Mutex
	resets  map[string]time.Time
	global  time.Time
}

// NewRateLimiter initializes a rate limiter.
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		resets: make(map[string]time.Time),
	}
}

// Wait blocks if rate limited.
func (r *RateLimiter) Wait(ctx context.Context, route string) error {
	r.mu.Lock()
	now := time.Now()

	var waitDuration time.Duration
	if r.global.After(now) {
		waitDuration = r.global.Sub(now)
	} else if reset, ok := r.resets[route]; ok && reset.After(now) {
		waitDuration = reset.Sub(now)
	}
	r.mu.Unlock()

	if waitDuration > 0 {
		select {
		case <-time.After(waitDuration):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

// Update records rate limit headers from API response.
func (r *RateLimiter) Update(route string, resp *http.Response) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 429 Too Many Requests
	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfterStr := resp.Header.Get("Retry-After")
		if retryAfterStr != "" {
			if seconds, err := strconv.ParseFloat(retryAfterStr, 64); err == nil {
				wait := time.Duration(seconds * float64(time.Second))
				if resp.Header.Get("X-RateLimit-Global") == "true" {
					r.global = time.Now().Add(wait)
				} else {
					r.resets[route] = time.Now().Add(wait)
				}
				return
			}
		}
	}

	remaining := resp.Header.Get("X-RateLimit-Remaining")
	reset := resp.Header.Get("X-RateLimit-Reset")

	if remaining == "0" && reset != "" {
		if resetUnix, err := strconv.ParseInt(reset, 10, 64); err == nil {
			r.resets[route] = time.Unix(resetUnix, 0)
		}
	}
}

// RESTClient handles all HTTP requests to Micup API.
type RESTClient struct {
	httpClient  *http.Client
	discovery   *DiscoveryManager
	token       string
	tokenPrefix string
	userAgent   string
	limiter     *RateLimiter
	apiEndpoint string
	mu          sync.RWMutex
}

// NewRESTClient initializes a REST client.
func NewRESTClient(token string, discovery *DiscoveryManager) *RESTClient {
	return &RESTClient{
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		discovery:   discovery,
		token:       token,
		tokenPrefix: "Bot ", // Official bot token prefix
		userAgent:   DefaultUserAgent,
		limiter:     NewRateLimiter(),
	}
}

// SetUserAgent updates the User-Agent header.
func (c *RESTClient) SetUserAgent(ua string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.userAgent = ua
}

// SetTokenPrefix sets token prefix (e.g. "Bot ", "Bearer ", or "").
func (c *RESTClient) SetTokenPrefix(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tokenPrefix = prefix
}

// getBaseURL resolves the API base URL.
func (c *RESTClient) getBaseURL(ctx context.Context) string {
	c.mu.RLock()
	if c.apiEndpoint != "" {
		ep := c.apiEndpoint
		c.mu.RUnlock()
		return ep
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.apiEndpoint != "" {
		return c.apiEndpoint
	}

	if c.discovery != nil {
		if eps, err := c.discovery.GetEndpoints(ctx); err == nil && eps.API != "" {
			c.apiEndpoint = eps.API
			return c.apiEndpoint
		}
	}

	c.apiEndpoint = DefaultBaseURL + "/api"
	return c.apiEndpoint
}

// Request performs an authorized HTTP request with rate limit compliance and retries.
func (c *RESTClient) Request(ctx context.Context, method, endpoint string, body interface{}) (*http.Response, error) {
	baseURL := c.getBaseURL(ctx)
	fullURL := baseURL + endpoint

	var bodyReader io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBytes)
	}

	// Retry loop for rate limits
	for retries := 0; retries < 3; retries++ {
		if err := c.limiter.Wait(ctx, endpoint); err != nil {
			return nil, err
		}

		req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("User-Agent", c.userAgent)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		if c.token != "" {
			req.Header.Set("Authorization", c.tokenPrefix+c.token)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("http request error: %w", err)
		}

		c.limiter.Update(endpoint, resp)

		if resp.StatusCode == http.StatusTooManyRequests {
			resp.Body.Close()
			continue
		}

		return resp, nil
	}

	return nil, errors.New("rate limit retry cap exceeded")
}

// RequestJSON performs a request and decodes the response body into target.
func (c *RESTClient) RequestJSON(ctx context.Context, method, endpoint string, body interface{}, target interface{}) error {
	resp, err := c.Request(ctx, method, endpoint, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr APIError
		if err := json.Unmarshal(bodyBytes, &apiErr); err == nil && apiErr.Code != "" {
			apiErr.StatusCode = resp.StatusCode
			return &apiErr
		}
		return &APIError{
			StatusCode: resp.StatusCode,
			Code:       "HTTP_ERROR",
			Message:    string(bodyBytes),
		}
	}

	if target != nil && len(bodyBytes) > 0 {
		if err := json.Unmarshal(bodyBytes, target); err != nil {
			return fmt.Errorf("failed to unmarshal JSON response: %w", err)
		}
	}

	return nil
}
