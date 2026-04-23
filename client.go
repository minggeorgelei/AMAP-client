package amapclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	Key        string
	BaseUrl    string
	HttpClient *http.Client
}

const (
	DefaultBaseUrl = "https://restapi.amap.com"
	DefaultTimeout = 15 * time.Second
)

// Options configures a Client. Any zero-valued field falls back to a sensible
// default (DefaultBaseUrl, DefaultTimeout, http.DefaultClient-like behavior).
type Options struct {
	Key        string
	BaseUrl    string
	Timeout    time.Duration
	HttpClient *http.Client
}

func NewClient(opts Options) *Client {
	if opts.BaseUrl == "" {
		opts.BaseUrl = DefaultBaseUrl
	}
	if opts.Timeout == 0 {
		opts.Timeout = DefaultTimeout
	}
	if opts.HttpClient == nil {
		opts.HttpClient = &http.Client{Timeout: opts.Timeout}
	}
	return &Client{
		Key:        opts.Key,
		BaseUrl:    opts.BaseUrl,
		HttpClient: opts.HttpClient,
	}
}

func (c *Client) httpClient() *http.Client {
	if c.HttpClient != nil {
		return c.HttpClient
	}
	return http.DefaultClient
}

func (c *Client) baseUrl() string {
	if c.BaseUrl != "" {
		return c.BaseUrl
	}
	return DefaultBaseUrl
}

// Request performs an HTTP request against the AMAP API and decodes the JSON
// response into out. The path is joined with the client's BaseUrl, and the
// API key is automatically injected into the query string.
func (c *Client) Request(ctx context.Context, method, path string, params url.Values, body io.Reader, out any) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if params == nil {
		params = url.Values{}
	}
	if c.Key != "" && params.Get("key") == "" {
		params.Set("key", c.Key)
	}

	endpoint := strings.TrimRight(c.baseUrl(), "/") + "/" + strings.TrimLeft(path, "/")
	if encoded := params.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("amap api %s %s: status %d: %s", method, path, resp.StatusCode, string(data))
	}

	if out == nil {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func (c *Client) Get(ctx context.Context, path string, params url.Values, out any) error {
	return c.Request(ctx, http.MethodGet, path, params, nil, out)
}
