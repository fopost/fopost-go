// Package fopost is the official Go SDK for the FoPost API — schedule,
// publish, and analyze social media content across every connected platform.
//
//	client, err := fopost.New("fp_...")
//	if err != nil {
//		log.Fatal(err)
//	}
//	post, err := client.Posts.Create(ctx, &fopost.CreatePostRequest{
//		WorkspaceID: workspaceID,
//		Accounts:    []string{accountID},
//		Content:     fopost.Text("Hello from Go"),
//	})
package fopost

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the production API, including its version prefix.
	DefaultBaseURL = "https://api.fopost.com/v1"
	// DefaultTimeout bounds a single request, including its body.
	DefaultTimeout = 30 * time.Second
	// DefaultMaxRetries counts total attempts, so 3 means two retries.
	DefaultMaxRetries = 3
	// Version is the SDK version, reported in the User-Agent.
	Version = "0.3.0"

	maxRetryWait  = 60 * time.Second
	baseRetryWait = 500 * time.Millisecond
)

// Client talks to the FoPost API. It is safe for concurrent use.
type Client struct {
	apiKey     string
	baseURL    string
	userAgent  string
	maxRetries int
	httpClient *http.Client

	Posts         *PostsService
	Workspaces    *WorkspacesService
	Accounts      *AccountsService
	AccountGroups *AccountGroupsService
	Communities   *CommunitiesService
	Labels        *LabelsService
	Webhooks      *WebhooksService
	Analytics     *AnalyticsService
	Automations   *AutomationsService
	Media         *MediaService
	Inbox         *InboxService
	Ads           *AdsService
	Validate      *ValidateService
	Blogs         *BlogsService
}

// Option configures a Client.
type Option func(*Client)

// WithBaseURL points the client at another deployment, e.g. a local API.
// It falls back to the FOPOST_BASE_URL environment variable.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		if baseURL != "" {
			c.baseURL = strings.TrimRight(baseURL, "/")
		}
	}
}

// WithHTTPClient supplies your own transport. It overrides WithTimeout.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

// WithTimeout bounds a single request. Ignored when WithHTTPClient is given.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		if timeout > 0 {
			c.httpClient.Timeout = timeout
		}
	}
}

// WithMaxRetries sets the total number of attempts per request. Anything below
// 1 is treated as 1, which disables retrying.
func WithMaxRetries(attempts int) Option {
	return func(c *Client) {
		if attempts < 1 {
			attempts = 1
		}
		c.maxRetries = attempts
	}
}

// WithUserAgent appends your own identifier to the SDK's User-Agent.
func WithUserAgent(userAgent string) Option {
	return func(c *Client) {
		if userAgent != "" {
			c.userAgent = userAgent + " " + c.userAgent
		}
	}
}

// New builds a client for the given API key, created in the FoPost dashboard
// under Settings → API Keys. An empty key falls back to the FOPOST_API_KEY
// environment variable.
func New(apiKey string, opts ...Option) (*Client, error) {
	if apiKey == "" {
		apiKey = os.Getenv("FOPOST_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("fopost: an API key is required — pass one to New or set FOPOST_API_KEY")
	}

	baseURL := DefaultBaseURL
	if fromEnv := os.Getenv("FOPOST_BASE_URL"); fromEnv != "" {
		baseURL = strings.TrimRight(fromEnv, "/")
	}

	c := &Client{
		apiKey:     apiKey,
		baseURL:    baseURL,
		userAgent:  "fopost-go/" + Version,
		maxRetries: DefaultMaxRetries,
		httpClient: &http.Client{Timeout: DefaultTimeout},
	}
	for _, opt := range opts {
		opt(c)
	}

	c.Posts = &PostsService{client: c}
	c.Workspaces = &WorkspacesService{client: c}
	c.Accounts = &AccountsService{client: c}
	c.AccountGroups = &AccountGroupsService{client: c}
	c.Communities = &CommunitiesService{client: c}
	c.Labels = &LabelsService{client: c}
	c.Webhooks = &WebhooksService{client: c}
	c.Analytics = &AnalyticsService{client: c}
	c.Automations = &AutomationsService{client: c}
	c.Media = &MediaService{client: c}
	c.Inbox = &InboxService{client: c}
	c.Ads = &AdsService{client: c}
	c.Validate = &ValidateService{client: c}
	c.Blogs = &BlogsService{client: c}

	return c, nil
}

// BaseURL is the API root every request is sent to.
func (c *Client) BaseURL() string { return c.baseURL }

// request is one call, prepared so a retry can replay it byte for byte.
type request struct {
	method      string
	path        string
	query       url.Values
	body        []byte
	contentType string
	// unwrap peels a {"data": ...} envelope off the response before decoding.
	unwrap bool
}

// Do sends an authenticated request to an endpoint the SDK does not wrap yet
// and decodes the response into out, which may be nil.
//
//	var body map[string]any
//	err := client.Do(ctx, "GET", "/platforms", nil, nil, &body)
func (c *Client) Do(ctx context.Context, method, path string, body any, query url.Values, out any) error {
	req := &request{method: method, path: path, query: query}
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("fopost: encoding request body: %w", err)
		}
		req.body = encoded
		req.contentType = "application/json"
	}
	return c.do(ctx, req, out)
}

func (c *Client) json(ctx context.Context, method, path string, body any, query url.Values, out any) error {
	req := &request{method: method, path: path, query: query, unwrap: true}
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("fopost: encoding request body: %w", err)
		}
		req.body = encoded
		req.contentType = "application/json"
	}
	return c.do(ctx, req, out)
}

func (c *Client) do(ctx context.Context, req *request, out any) error {
	endpoint := c.baseURL + "/" + strings.TrimLeft(req.path, "/")
	if len(req.query) > 0 {
		endpoint += "?" + req.query.Encode()
	}

	for attempt := 1; ; attempt++ {
		httpReq, err := http.NewRequestWithContext(ctx, req.method, endpoint, bodyReader(req.body))
		if err != nil {
			return fmt.Errorf("fopost: building request: %w", err)
		}
		httpReq.Header.Set("Accept", "application/json")
		httpReq.Header.Set("X-API-Key", c.apiKey)
		httpReq.Header.Set("User-Agent", c.userAgent)
		if req.contentType != "" {
			httpReq.Header.Set("Content-Type", req.contentType)
		}

		res, err := c.httpClient.Do(httpReq)
		if err != nil {
			// A cancelled context is the caller's decision, not a blip.
			if ctx.Err() != nil {
				return err
			}
			wrapped := fmt.Errorf("fopost: %s %s: %w", req.method, req.path, err)
			if attempt >= c.maxRetries {
				return wrapped
			}
			if err := sleep(ctx, backoff(attempt)); err != nil {
				return err
			}
			continue
		}

		err = c.decode(res, out, req.unwrap)
		if err == nil {
			return nil
		}

		apiErr, isAPIErr := APIError(err)
		if !isAPIErr || attempt >= c.maxRetries || !retryable(apiErr.Status) {
			return err
		}
		wait := backoff(attempt)
		if apiErr.RetryAfter > 0 {
			wait = min(apiErr.RetryAfter, maxRetryWait)
		}
		if err := sleep(ctx, wait); err != nil {
			return err
		}
	}
}

func (c *Client) decode(res *http.Response, out any, unwrap bool) error {
	defer res.Body.Close()
	body, readErr := io.ReadAll(res.Body)

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return errorFromResponse(res, body)
	}
	if readErr != nil {
		return fmt.Errorf("fopost: reading response body: %w", readErr)
	}
	if out == nil || len(bytes.TrimSpace(body)) == 0 || res.StatusCode == http.StatusNoContent {
		return nil
	}

	payload := body
	if unwrap {
		payload = unwrapEnvelope(body)
	}
	if err := json.Unmarshal(payload, out); err != nil {
		return fmt.Errorf("fopost: decoding response: %w", err)
	}
	return nil
}

// unwrapEnvelope peels the {"data": ...} wrapper the API puts around most
// resources. Some endpoints answer bare, so it only unwraps what is there.
func unwrapEnvelope(body []byte) []byte {
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(body, &envelope); err != nil {
		return body
	}
	if data, ok := envelope["data"]; ok {
		return data
	}
	return body
}

func errorFromResponse(res *http.Response, body []byte) error {
	apiErr := &Error{
		Status:     res.StatusCode,
		Message:    http.StatusText(res.StatusCode),
		Body:       body,
		RateLimit:  rateLimitFrom(res.Header),
		RetryAfter: retryAfter(res.Header),
	}

	var parsed struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	if json.Unmarshal(body, &parsed) == nil {
		apiErr.Code = parsed.Error
		if parsed.Message != "" {
			apiErr.Message = parsed.Message
		} else if parsed.Error != "" {
			apiErr.Message = parsed.Error
		}
	}
	if apiErr.Message == "" {
		apiErr.Message = fmt.Sprintf("HTTP %d", res.StatusCode)
	}
	return apiErr
}

func rateLimitFrom(header http.Header) RateLimit {
	limit := RateLimit{}
	if v, err := strconv.Atoi(header.Get("X-RateLimit-Limit")); err == nil {
		limit.Limit = v
	}
	if v, err := strconv.Atoi(header.Get("X-RateLimit-Remaining")); err == nil {
		limit.Remaining = v
	}
	if raw := header.Get("X-RateLimit-Reset"); raw != "" {
		if seconds, err := strconv.ParseInt(raw, 10, 64); err == nil {
			// The API sends a unix timestamp; tolerate a delta from a proxy.
			if seconds > 1_000_000_000 {
				limit.Reset = time.Unix(seconds, 0).UTC()
			} else {
				limit.Reset = time.Now().Add(time.Duration(seconds) * time.Second).UTC()
			}
		}
	}
	return limit
}

// retryAfter reads the header in either of its forms: delta-seconds or a date.
func retryAfter(header http.Header) time.Duration {
	raw := strings.TrimSpace(header.Get("Retry-After"))
	if raw == "" {
		return 0
	}
	if seconds, err := strconv.ParseFloat(raw, 64); err == nil {
		if seconds < 0 {
			return 0
		}
		return time.Duration(seconds * float64(time.Second))
	}
	if target, err := http.ParseTime(raw); err == nil {
		if wait := time.Until(target); wait > 0 {
			return wait
		}
	}
	return 0
}

func retryable(status int) bool {
	return status == http.StatusTooManyRequests || status >= 500
}

func backoff(attempt int) time.Duration {
	wait := time.Duration(math.Pow(2, float64(attempt-1))) * baseRetryWait
	return min(wait, maxRetryWait)
}

func sleep(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func bodyReader(body []byte) io.Reader {
	if body == nil {
		return nil
	}
	return bytes.NewReader(body)
}
