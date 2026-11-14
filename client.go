package sdp

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

const (
	defaultUserAgent = "sdp-go/1.0"
	defaultTimeout   = 30 * time.Second
)

// AuthType represents the authentication mechanism for SDP.
type AuthType int

const (
	AuthTechnicianKey AuthType = iota
	AuthBearerToken
	AuthZohoOAuth
)

// ClientOption represents an option to configure the Client.
type ClientOption func(*Client)

// Client is the main entry point to communicate with ManageEngine ServiceDesk Plus.
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	userAgent  string

	authType AuthType
	apiKey   string

	// Services
	Requests *RequestsService
}

// WithHTTPClient overrides the default http.Client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

// WithAuthToken sets the technician authentication token (passed in `authtoken` header).
func WithAuthToken(token string) ClientOption {
	return func(c *Client) {
		c.authType = AuthTechnicianKey
		c.apiKey = token
	}
}

// WithOAuthToken sets the OAuth bearer token (passed in `Authorization: Bearer <token>`).
func WithOAuthToken(token string) ClientOption {
	return func(c *Client) {
		c.authType = AuthBearerToken
		c.apiKey = token
	}
}

// WithZohoOAuthToken sets the Zoho OAuth token (passed in `Authorization: Zoho-oauthtoken <token>`).
func WithZohoOAuthToken(token string) ClientOption {
	return func(c *Client) {
		c.authType = AuthZohoOAuth
		c.apiKey = token
	}
}

// WithUserAgent overrides the default User-Agent header.
func WithUserAgent(userAgent string) ClientOption {
	return func(c *Client) {
		c.userAgent = userAgent
	}
}

// NewClient creates a new SDP API client with the given baseURL and options.
func NewClient(rawBaseURL string, opts ...ClientOption) (*Client, error) {
	if rawBaseURL == "" {
		return nil, fmt.Errorf("baseURL is required")
	}

	if !strings.HasPrefix(rawBaseURL, "http://") && !strings.HasPrefix(rawBaseURL, "https://") {
		rawBaseURL = "https://" + rawBaseURL
	}

	parsedURL, err := url.Parse(strings.TrimRight(rawBaseURL, "/"))
	if err != nil {
		return nil, fmt.Errorf("invalid baseURL: %w", err)
	}

	c := &Client{
		baseURL: parsedURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		userAgent: defaultUserAgent,
	}

	for _, opt := range opts {
		opt(c)
	}

	// Initialize Services
	c.Requests = NewRequestsService(c)

	return c, nil
}

func (c *Client) newRequest(ctx context.Context, method, path string, query url.Values, body io.Reader) (*http.Request, error) {
	rel, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	u := c.baseURL.ResolveReference(rel)
	if len(query) > 0 {
		q := u.Query()
		for k, vs := range query {
			for _, v := range vs {
				q.Add(k, v)
			}
		}
		u.RawQuery = q.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/vnd.manageengine.sdp.v3+json, application/json")

	// Apply Authentication
	if c.apiKey != "" {
		switch c.authType {
		case AuthTechnicianKey:
			req.Header.Set("authtoken", c.apiKey)
		case AuthBearerToken:
			req.Header.Set("Authorization", "Bearer "+c.apiKey)
		case AuthZohoOAuth:
			req.Header.Set("Authorization", "Zoho-oauthtoken "+c.apiKey)
		}
	}

	return req, nil
}

func (c *Client) do(req *http.Request, v any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check HTTP status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return c.parseErrorResponse(resp.StatusCode, bodyBytes)
	}

	if v != nil && len(bodyBytes) > 0 {
		if err := json.Unmarshal(bodyBytes, v); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

func (c *Client) parseErrorResponse(statusCode int, body []byte) error {
	var errEnvelope struct {
		ResponseStatus FlexibleResponseStatus `json:"response_status"`
	}

	if err := json.Unmarshal(body, &errEnvelope); err == nil && len(errEnvelope.ResponseStatus) > 0 {
		return &APIError{
			HTTPStatusCode: statusCode,
			ResponseStatus: errEnvelope.ResponseStatus,
			RawBody:        string(body),
		}
	}

	return &APIError{
		HTTPStatusCode: statusCode,
		RawBody:        string(body),
	}
}

// get performs an HTTP GET request.
func (c *Client) get(ctx context.Context, path string, query url.Values, v any) error {
	req, err := c.newRequest(ctx, http.MethodGet, path, query, nil)
	if err != nil {
		return err
	}
	return c.do(req, v)
}

// post performs an HTTP POST request sending input_data as application/x-www-form-urlencoded.
func (c *Client) post(ctx context.Context, path string, payload any, v any) error {
	return c.sendPayload(ctx, http.MethodPost, path, payload, v)
}

// put performs an HTTP PUT request sending input_data as application/x-www-form-urlencoded.
func (c *Client) put(ctx context.Context, path string, payload any, v any) error {
	return c.sendPayload(ctx, http.MethodPut, path, payload, v)
}

// delete performs an HTTP DELETE request.
func (c *Client) delete(ctx context.Context, path string) error {
	req, err := c.newRequest(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

func (c *Client) sendPayload(ctx context.Context, method, path string, payload any, v any) error {
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	formData := url.Values{}
	formData.Set("input_data", string(jsonBytes))

	body := strings.NewReader(formData.Encode())
	req, err := c.newRequest(ctx, method, path, nil, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return c.do(req, v)
}
