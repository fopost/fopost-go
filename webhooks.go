package fopost

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// WebhooksService covers outbound webhooks, the push counterpart to polling
// a post's deliveries.
type WebhooksService struct{ client *Client }

// Webhook is one subscription.
type Webhook struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspaceId"`
	URL         string `json:"url"`
	// Events are the Event constants this subscription asked for.
	Events          []string `json:"events"`
	Active          bool     `json:"active"`
	LastTriggeredAt Time     `json:"lastTriggeredAt"`
	FailureCount    int      `json:"failureCount"`
	CreatedAt       Time     `json:"createdAt"`
}

// CreatedWebhook is a new subscription. Secret is shown once, at creation, and
// signs every delivery — store it now.
type CreatedWebhook struct {
	ID          string   `json:"id"`
	WorkspaceID string   `json:"workspaceId"`
	URL         string   `json:"url"`
	Secret      string   `json:"secret"`
	Events      []string `json:"events"`
	Active      bool     `json:"active"`
	CreatedAt   Time     `json:"createdAt"`
}

// List returns the webhooks the key can reach.
func (s *WebhooksService) List(ctx context.Context) ([]Webhook, error) {
	var out []Webhook
	if err := s.client.json(ctx, "GET", "/webhooks", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateWebhookRequest is the body of Create.
type CreateWebhookRequest struct {
	WorkspaceID string `json:"workspaceId"`
	URL         string `json:"url"`
	// Events are the Event constants to subscribe to.
	Events []string `json:"events"`
}

// Create subscribes an endpoint to a workspace's events.
func (s *WebhooksService) Create(ctx context.Context, body *CreateWebhookRequest) (*CreatedWebhook, error) {
	out := &CreatedWebhook{}
	if err := s.client.json(ctx, "POST", "/webhooks", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateWebhookRequest is the body of Update. Only the fields you set are sent.
type UpdateWebhookRequest struct {
	URL    string   `json:"url,omitempty"`
	Events []string `json:"events,omitempty"`
	Active *bool    `json:"active,omitempty"`
}

// Update changes a subscription's endpoint, events, or active flag.
func (s *WebhooksService) Update(ctx context.Context, id string, body *UpdateWebhookRequest) (*Webhook, error) {
	out := &Webhook{}
	if err := s.client.json(ctx, "PUT", "/webhooks/"+url.PathEscape(id), body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Delete removes a subscription.
func (s *WebhooksService) Delete(ctx context.Context, id string) error {
	return s.client.Do(ctx, "DELETE", "/webhooks/"+url.PathEscape(id), nil, nil, nil)
}

// Test sends a sample event to the subscribed endpoint.
func (s *WebhooksService) Test(ctx context.Context, id string) (*Message, error) {
	out := &Message{}
	if err := s.client.Do(ctx, "POST", "/webhooks/"+url.PathEscape(id)+"/test", nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Headers on every delivery FoPost sends, to a webhook or to a blog website URL.
const (
	// WebhookSignatureHeader is `sha256=<hex HMAC-SHA256(secret, body)>`.
	WebhookSignatureHeader = "X-FoPost-Signature"
	// WebhookTimestampHeader is the unix seconds the delivery was sent.
	WebhookTimestampHeader = "X-FoPost-Timestamp"
	// WebhookTimestampedSignatureHeader is
	// `sha256=<hex HMAC-SHA256(secret, "<timestamp>.<body>")>`.
	WebhookTimestampedSignatureHeader = "X-FoPost-Signature-256"
	WebhookEventHeader                = "X-FoPost-Event"
	WebhookDeliveryHeader             = "X-FoPost-Delivery"
)

// WebhookTolerance is how far a delivery's timestamp may sit from now before
// VerifyWebhook refuses it.
const WebhookTolerance = 300 * time.Second

// ErrWebhookSignature is returned when no signature header is present or
// none matches the body.
var ErrWebhookSignature = errors.New("fopost: webhook signature mismatch")

// ErrWebhookTimestamp is returned when the delivery's timestamp is missing,
// unreadable or older than WebhookTolerance.
var ErrWebhookTimestamp = errors.New("fopost: webhook timestamp outside tolerance")

// WebhookEnvelope is one verified delivery. Data is the event body, still
// encoded; decode it with Decode or json.Unmarshal.
type WebhookEnvelope struct {
	Event     string          `json:"event"`
	Data      json.RawMessage `json:"data"`
	Timestamp Time            `json:"timestamp"`
}

// Decode unmarshals the event body into out.
func (e *WebhookEnvelope) Decode(out any) error {
	if err := json.Unmarshal(e.Data, out); err != nil {
		return fmt.Errorf("fopost: decoding webhook data: %w", err)
	}
	return nil
}

// BlogArticlePublished is the Data of an EventArticlePublished delivery.
// Image URLs stay fetchable for six hours; download and store your own copy.
type BlogArticlePublished struct {
	ID              string             `json:"id"`
	Title           string             `json:"title"`
	Slug            string             `json:"slug"`
	Excerpt         *string            `json:"excerpt"`
	MetaTitle       *string            `json:"metaTitle"`
	MetaDescription *string            `json:"metaDescription"`
	BodyMarkdown    string             `json:"bodyMarkdown"`
	BodyHTML        string             `json:"bodyHtml"`
	FeaturedImage   *BlogArticleImage  `json:"featuredImage"`
	Images          []BlogArticleImage `json:"images"`
	Tags            []string           `json:"tags"`
	Keyword         *string            `json:"keyword"`
	WordCount       int                `json:"wordCount"`
	PublishedAt     Time               `json:"publishedAt"`
	PublishedURL    string             `json:"publishedUrl"`
}

// VerifyWebhook checks a delivery against the secret and returns its
// envelope. rawBody is the request body exactly as received: re-encoding
// parsed JSON changes the bytes and the signature no longer matches.
//
// The timestamped signature is checked when its header is present and the
// timestamp has to sit within WebhookTolerance of now. Without a timestamp
// header the plain body signature is checked instead.
func VerifyWebhook(rawBody []byte, headers http.Header, secret string) (*WebhookEnvelope, error) {
	return verifyWebhookAt(rawBody, headers, secret, time.Now())
}

func verifyWebhookAt(rawBody []byte, headers http.Header, secret string, now time.Time) (*WebhookEnvelope, error) {
	if secret == "" {
		return nil, ErrWebhookSignature
	}
	if ts := headers.Get(WebhookTimestampHeader); ts != "" {
		seconds, err := strconv.ParseInt(strings.TrimSpace(ts), 10, 64)
		if err != nil {
			return nil, ErrWebhookTimestamp
		}
		if drift := now.Sub(time.Unix(seconds, 0)); drift > WebhookTolerance || drift < -WebhookTolerance {
			return nil, ErrWebhookTimestamp
		}
		signed := append([]byte(ts+"."), rawBody...)
		if !signatureMatches(headers.Get(WebhookTimestampedSignatureHeader), signed, secret) {
			return nil, ErrWebhookSignature
		}
	} else if !signatureMatches(headers.Get(WebhookSignatureHeader), rawBody, secret) {
		return nil, ErrWebhookSignature
	}

	out := &WebhookEnvelope{}
	if err := json.Unmarshal(rawBody, out); err != nil {
		return nil, fmt.Errorf("fopost: decoding webhook body: %w", err)
	}
	return out, nil
}

func signatureMatches(header string, signed []byte, secret string) bool {
	header = strings.TrimSpace(header)
	if header == "" {
		return false
	}
	got, err := hex.DecodeString(strings.TrimPrefix(header, "sha256="))
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(signed)
	return hmac.Equal(got, mac.Sum(nil))
}
