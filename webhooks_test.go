package fopost

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

const webhookTestSecret = "whsec_test"

var webhookTestBody = []byte(`{"event":"article.published","data":{"id":"art_1","title":"Trail shoes","slug":"trail-shoes","bodyMarkdown":"# Trail shoes","bodyHtml":"<h1>Trail shoes</h1>","featuredImage":null,"images":[{"mediaId":"m_2","url":"https://cdn.yourbrand.com/2.jpg","alt":"Trail","position":1}],"tags":["gear"],"keyword":"trail shoes","wordCount":1200,"publishedAt":"2026-10-11T09:00:00.000Z","publishedUrl":"https://yourbrand.com/blog/trail-shoes"},"timestamp":"2026-10-11T09:00:01.000Z"}`)

func signWebhookTest(secret string, parts ...[]byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	for _, part := range parts {
		mac.Write(part)
	}
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func timestampedWebhookHeaders(body []byte, sentAt time.Time) http.Header {
	ts := strconv.FormatInt(sentAt.Unix(), 10)
	headers := http.Header{}
	headers.Set(WebhookTimestampHeader, ts)
	headers.Set(WebhookTimestampedSignatureHeader, signWebhookTest(webhookTestSecret, []byte(ts+"."), body))
	headers.Set(WebhookSignatureHeader, signWebhookTest(webhookTestSecret, body))
	headers.Set(WebhookEventHeader, EventArticlePublished)
	return headers
}

func TestVerifyWebhookAcceptsATimestampedSignature(t *testing.T) {
	now := time.Date(2026, 10, 11, 9, 0, 5, 0, time.UTC)
	envelope, err := verifyWebhookAt(webhookTestBody, timestampedWebhookHeaders(webhookTestBody, now.Add(-2*time.Second)), webhookTestSecret, now)
	if err != nil {
		t.Fatalf("verifyWebhookAt: %v", err)
	}
	if envelope.Event != EventArticlePublished || envelope.Timestamp.Time.IsZero() {
		t.Fatalf("envelope = %+v", envelope)
	}
	var article BlogArticlePublished
	if err := envelope.Decode(&article); err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if article.ID != "art_1" || article.PublishedURL != "https://yourbrand.com/blog/trail-shoes" || *article.Images[0].Position != 1 || article.Images[0].MediaID != "m_2" || *article.Keyword != "trail shoes" {
		t.Fatalf("article = %+v", article)
	}
}

func TestVerifyWebhookRejectsATamperedBody(t *testing.T) {
	now := time.Date(2026, 10, 11, 9, 0, 5, 0, time.UTC)
	headers := timestampedWebhookHeaders(webhookTestBody, now)
	tampered := append([]byte{}, webhookTestBody...)
	tampered[len(tampered)-3] = '2'

	if _, err := verifyWebhookAt(tampered, headers, webhookTestSecret, now); !errors.Is(err, ErrWebhookSignature) {
		t.Fatalf("expected ErrWebhookSignature, got %v", err)
	}
	// A valid plain signature does not rescue a bad timestamped one.
	headers.Set(WebhookSignatureHeader, signWebhookTest(webhookTestSecret, tampered))
	if _, err := verifyWebhookAt(tampered, headers, webhookTestSecret, now); !errors.Is(err, ErrWebhookSignature) {
		t.Fatalf("expected ErrWebhookSignature, got %v", err)
	}
	if _, err := verifyWebhookAt(webhookTestBody, timestampedWebhookHeaders(webhookTestBody, now), "whsec_other", now); !errors.Is(err, ErrWebhookSignature) {
		t.Fatalf("expected ErrWebhookSignature with the wrong secret, got %v", err)
	}
}

func TestVerifyWebhookRejectsAStaleTimestamp(t *testing.T) {
	now := time.Date(2026, 10, 11, 9, 0, 0, 0, time.UTC)
	stale := timestampedWebhookHeaders(webhookTestBody, now.Add(-WebhookTolerance-time.Second))
	if _, err := verifyWebhookAt(webhookTestBody, stale, webhookTestSecret, now); !errors.Is(err, ErrWebhookTimestamp) {
		t.Fatalf("expected ErrWebhookTimestamp, got %v", err)
	}
	future := timestampedWebhookHeaders(webhookTestBody, now.Add(WebhookTolerance+time.Second))
	if _, err := verifyWebhookAt(webhookTestBody, future, webhookTestSecret, now); !errors.Is(err, ErrWebhookTimestamp) {
		t.Fatalf("expected ErrWebhookTimestamp for a future timestamp, got %v", err)
	}
	edge := timestampedWebhookHeaders(webhookTestBody, now.Add(-WebhookTolerance))
	if _, err := verifyWebhookAt(webhookTestBody, edge, webhookTestSecret, now); err != nil {
		t.Fatalf("a timestamp exactly at the tolerance should pass: %v", err)
	}
}

func TestVerifyWebhookFallsBackToTheBodySignature(t *testing.T) {
	headers := http.Header{}
	headers.Set(WebhookSignatureHeader, signWebhookTest(webhookTestSecret, webhookTestBody))

	envelope, err := VerifyWebhook(webhookTestBody, headers, webhookTestSecret)
	if err != nil {
		t.Fatalf("VerifyWebhook: %v", err)
	}
	if envelope.Event != EventArticlePublished {
		t.Fatalf("event = %q", envelope.Event)
	}
	headers.Set(WebhookSignatureHeader, "sha256=00")
	if _, err := VerifyWebhook(webhookTestBody, headers, webhookTestSecret); !errors.Is(err, ErrWebhookSignature) {
		t.Fatalf("expected ErrWebhookSignature, got %v", err)
	}
	if _, err := VerifyWebhook(webhookTestBody, http.Header{}, webhookTestSecret); !errors.Is(err, ErrWebhookSignature) {
		t.Fatalf("expected ErrWebhookSignature without headers, got %v", err)
	}
}

func TestWebhookDeliveriesRoundTrip(t *testing.T) {
	var method, path, query string
	var status int
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, path, query = r.Method, r.URL.Path, r.URL.RawQuery
		switch {
		case r.Method == "POST":
			w.WriteHeader(http.StatusNoContent)
		case strings.HasSuffix(r.URL.Path, "/deliveries"):
			_, _ = io.WriteString(w, `{"data":[{"id":"del_1","webhook_id":"wh_1","event":"post.published","delivery_key":"dk_1","attempt":2,"status":"failed","response_status":502,"response_snippet":"bad gateway","duration_ms":120,"payload_hash":"abc","next_retry_at":"2026-10-18T09:05:00.000Z","created_at":"2026-10-18T09:00:00.000Z"}]}`)
		default:
			_, _ = io.WriteString(w, `{"data":{"id":"del_1","webhook_id":"wh_1","event":"post.published","delivery_key":"dk_1","attempt":2,"status":"success","response_status":200,"response_snippet":null,"duration_ms":80,"payload_hash":"abc","next_retry_at":null,"created_at":"2026-10-18T09:00:00.000Z"}}`)
		}
		status = http.StatusOK
	})
	ctx := context.Background()

	deliveries, err := client.Webhooks.ListDeliveries(ctx, "wh_1", &ListWebhookDeliveriesParams{Status: "failed", Limit: 10})
	if err != nil {
		t.Fatalf("ListDeliveries: %v", err)
	}
	if method != "GET" || path != "/webhooks/wh_1/deliveries" || query != "limit=10&status=failed" {
		t.Fatalf("ListDeliveries sent %s %s?%s", method, path, query)
	}
	if len(deliveries) != 1 || deliveries[0].Attempt != 2 || *deliveries[0].ResponseStatus != 502 || *deliveries[0].ResponseSnippet != "bad gateway" || deliveries[0].NextRetryAt.IsZero() {
		t.Fatalf("deliveries = %+v", deliveries)
	}

	delivery, err := client.Webhooks.GetDelivery(ctx, "wh_1", "del_1")
	if err != nil {
		t.Fatalf("GetDelivery: %v", err)
	}
	if method != "GET" || path != "/webhooks/wh_1/deliveries/del_1" {
		t.Fatalf("GetDelivery sent %s %s", method, path)
	}
	if delivery.Status != "success" || delivery.ResponseSnippet != nil || !delivery.NextRetryAt.IsZero() {
		t.Fatalf("delivery = %+v", delivery)
	}

	if err := client.Webhooks.Redeliver(ctx, "wh_1", "del_1"); err != nil {
		t.Fatalf("Redeliver: %v", err)
	}
	if method != "POST" || path != "/webhooks/wh_1/deliveries/del_1/redeliver" || status != http.StatusOK {
		t.Fatalf("Redeliver sent %s %s", method, path)
	}
}
