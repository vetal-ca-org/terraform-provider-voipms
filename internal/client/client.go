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

const defaultTimeout = 30 * time.Second

// Client talks to the VoIP.ms REST API.
type Client struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
	userAgent  string
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
		baseURL:    baseURL,
		username:   cfg.Username,
		password:   cfg.Password,
		httpClient: httpClient,
		userAgent:  userAgent,
	}
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
func (c *Client) CallWrite(ctx context.Context, method string, params map[string]string, dest any) error {
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return fmt.Errorf("build request for %s: %w", method, err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call %s: %w", method, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read %s response: %w", method, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("call %s: unexpected HTTP %d", method, resp.StatusCode)
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
