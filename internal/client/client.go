// Package client is a small HTTP wrapper around the VoIP.ms REST/JSON API.
//
// Every call is an HTTP GET to https://voip.ms/api/v1/rest.php with
// api_username, api_password, and method as query parameters. The API also
// requires the caller's public IP to be allow-listed in the VoIP.ms portal.
package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ErrNotFound is returned when a lookup for a specific account object fails.
var ErrNotFound = errors.New("voip.ms object not found")

// DefaultBaseURL is the VoIP.ms REST endpoint that returns JSON directly.
const DefaultBaseURL = "https://voip.ms/api/v1/rest.php"

const (
	// VoIP.ms routinely takes tens of seconds under load, and a write that
	// times out client-side may still have been applied server-side.
	defaultTimeout     = 60 * time.Second
	defaultMaxAttempts = 5
)

// Cloudflare / edge timeouts while voip.ms origin is slow or unreachable.
const (
	statusCloudflareTimeout           = 522
	statusCloudflareOriginUnreachable = 523
	statusCloudflareEdgeTimeout       = 524
)

// Client talks to the VoIP.ms REST API.
type Client struct {
	baseURL      string
	username     string
	password     string
	httpClient   *http.Client
	userAgent    string
	cache        listCache
	maxAttempts  int
	retryBackoff func(ctx context.Context, d time.Duration) error
}

// Config is used to construct a Client.
type Config struct {
	BaseURL    string
	Username   string
	Password   string
	HTTPClient *http.Client
	UserAgent  string
}

// New returns a Client ready to call the VoIP.ms API.
func New(cfg Config) *Client {
	baseURL := strings.TrimSpace(cfg.BaseURL)
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultTimeout}
	}

	userAgent := cfg.UserAgent
	if userAgent == "" {
		userAgent = "terraform-provider-voipms"
	}

	return &Client{
		baseURL:      baseURL,
		username:     cfg.Username,
		password:     cfg.Password,
		httpClient:   httpClient,
		userAgent:    userAgent,
		maxAttempts:  defaultMaxAttempts,
		retryBackoff: sleepContext,
	}
}

func sleepContext(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func retryDelay(attempt int) time.Duration {
	d := time.Second << (attempt - 1)
	if d > 8*time.Second {
		return 8 * time.Second
	}
	return d
}

func retryableStatus(code int) bool {
	switch code {
	case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout,
		statusCloudflareTimeout, statusCloudflareOriginUnreachable, statusCloudflareEdgeTimeout:
		return true
	default:
		return false
	}
}

// HTTPError is a non-2xx response from the VoIP.ms endpoint (or Cloudflare in front of it).
type HTTPError struct {
	Method string
	Status int
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("call %s: unexpected HTTP %d", e.Method, e.Status)
}

func (e *HTTPError) Retryable() bool {
	return retryableStatus(e.Status)
}

func retryableCallError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.Retryable()
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return false
	}
	return true
}

// APIError is returned when VoIP.ms responds with a non-success status.
type APIError struct {
	Method  string
	Status  string
	Message string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("voip.ms API method %q returned status %q: %s", e.Method, e.Status, e.Message)
	}
	return fmt.Sprintf("voip.ms API method %q returned status %q", e.Method, e.Status)
}

// EmptyResult reports whether the status means “no rows” rather than a hard failure.
func (e *APIError) EmptyResult() bool {
	return strings.HasPrefix(strings.ToLower(e.Status), "no_")
}

func emptyResult(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.EmptyResult()
}

// Call invokes a VoIP.ms API method and decodes the JSON body into dest.
// Authentication parameters are added automatically. The password is never
// included in returned errors. Empty parameter values are omitted.
func (c *Client) Call(ctx context.Context, method string, params map[string]string, dest any) error {
	return c.call(ctx, method, params, dest, true)
}

// CallWrite is like Call but sends empty parameter values so fields can be cleared.
// Cached lists are dropped either way: a write that times out client-side may
// still have been applied, so the caches are suspect regardless of the outcome.
func (c *Client) CallWrite(ctx context.Context, method string, params map[string]string, dest any) error {
	defer c.invalidate()
	return c.call(ctx, method, params, dest, false)
}

func (c *Client) call(ctx context.Context, method string, params map[string]string, dest any, omitEmpty bool) error {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return fmt.Errorf("parse API URL: %w", err)
	}

	q := u.Query()
	q.Set("api_username", c.username)
	q.Set("api_password", c.password)
	q.Set("method", method)
	for key, value := range params {
		if omitEmpty && value == "" {
			continue
		}
		q.Set(key, value)
	}
	u.RawQuery = q.Encode()

	attempts := c.maxAttempts
	if attempts < 1 {
		attempts = 1
	}

	var last error
	for attempt := 1; attempt <= attempts; attempt++ {
		last = c.doOnce(ctx, method, u.String(), dest)
		if last == nil {
			return nil
		}
		if attempt == attempts || !retryableCallError(last) {
			return last
		}
		if err := c.retryBackoff(ctx, retryDelay(attempt)); err != nil {
			return last
		}
	}
	return last
}

func (c *Client) doOnce(ctx context.Context, method, rawURL string, dest any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return fmt.Errorf("build request for %s: %w", method, redactRequestError(err))
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call %s: %w", method, redactRequestError(err))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read %s response: %w", method, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &HTTPError{Method: method, Status: resp.StatusCode}
	}

	var envelope struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return fmt.Errorf("decode %s status: %w", method, err)
	}
	if !strings.EqualFold(envelope.Status, "success") {
		return &APIError{Method: method, Status: envelope.Status, Message: envelope.Message}
	}

	if dest == nil {
		return nil
	}
	if err := json.Unmarshal(body, dest); err != nil {
		return fmt.Errorf("decode %s body: %w", method, err)
	}
	return nil
}

// Balance is the payload returned by getBalance with advanced=true.
type Balance struct {
	CurrentBalance string `json:"current_balance"`
	SpentTotal     string `json:"spent_total"`
	CallsTotal     string `json:"calls_total"`
	TimeTotal      string `json:"time_total"`
	SpentToday     string `json:"spent_today"`
	CallsToday     string `json:"calls_today"`
	TimeToday      string `json:"time_today"`
}

type balanceResponse struct {
	Status  string  `json:"status"`
	Balance Balance `json:"balance"`
}

// GetBalance calls getBalance with the advanced breakdown enabled.
func (c *Client) GetBalance(ctx context.Context) (Balance, error) {
	var resp balanceResponse
	err := c.Call(ctx, "getBalance", map[string]string{"advanced": "true"}, &resp)
	if err != nil {
		return Balance{}, err
	}
	return resp.Balance, nil
}

// redactRequestError strips the query string out of a *url.Error. api_username
// and api_password travel in the query, and net/http puts the whole URL in the
// error it returns, so wrapping one verbatim prints the API password into
// Terraform output and CI logs.
func redactRequestError(err error) error {
	var uerr *url.Error
	if !errors.As(err, &uerr) {
		return err
	}
	safe := "(url redacted)"
	if parsed, perr := url.Parse(uerr.URL); perr == nil {
		parsed.RawQuery = ""
		safe = parsed.String()
	}
	return fmt.Errorf("%s %s: %w", uerr.Op, safe, uerr.Err)
}

func filterList[T any](items []T, keep func(*T) bool) []T {
	out := make([]T, 0, 1)
	for i := range items {
		if keep(&items[i]) {
			out = append(out, items[i])
		}
	}
	return out
}
