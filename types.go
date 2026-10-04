package fopost

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Time is an API timestamp. It tolerates null, an empty string, and formats
// other than RFC 3339, keeping the original text in Raw so nothing is lost.
type Time struct {
	time.Time
	// Raw is the string the API sent, empty when the field was null.
	Raw string
}

// NewTime wraps a time.Time for a request body.
func NewTime(t time.Time) *Time { return &Time{Time: t} }

// UnmarshalJSON accepts a timestamp string or null.
func (t *Time) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*t = Time{}
		return nil
	}
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	t.Raw = raw
	if raw == "" {
		t.Time = time.Time{}
		return nil
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05.000Z", "2006-01-02 15:04:05", "2006-01-02"} {
		if parsed, err := time.Parse(layout, raw); err == nil {
			t.Time = parsed
			return nil
		}
	}
	// Unparseable is not fatal: Raw still carries what the API said.
	t.Time = time.Time{}
	return nil
}

// MarshalJSON writes RFC 3339, or the original text when it never parsed.
func (t Time) MarshalJSON() ([]byte, error) {
	if t.Time.IsZero() {
		if t.Raw == "" {
			return []byte("null"), nil
		}
		return json.Marshal(t.Raw)
	}
	return json.Marshal(t.Time.UTC().Format(time.RFC3339))
}

// IsZero reports whether the API sent no usable timestamp.
func (t Time) IsZero() bool { return t.Time.IsZero() }

func (t Time) String() string {
	if t.Time.IsZero() {
		return t.Raw
	}
	return t.Time.UTC().Format(time.RFC3339)
}

// PageMeta describes one page of a paginated list.
type PageMeta struct {
	CurrentPage int `json:"current_page"`
	PerPage     int `json:"per_page"`
	Total       int `json:"total"`
	LastPage    int `json:"last_page"`
	From        int `json:"from"`
	To          int `json:"to"`
}

// Message is the bare acknowledgement several deletes answer with.
type Message struct {
	Message string `json:"message"`
	Deleted int    `json:"deleted,omitempty"`
}

// MediaItem is one attachment on a content block.
type MediaItem struct {
	// Type is "image", "video", or "gif".
	Type      string  `json:"type"`
	Name      string  `json:"name"`
	URL       string  `json:"url"`
	Size      float64 `json:"size,omitempty"`
	Alt       string  `json:"alt,omitempty"`
	Thumbnail string  `json:"thumbnail,omitempty"`
}

// ContentBlock is one text-plus-media unit. A single post has one block; a
// thread has one per entry, in order.
type ContentBlock struct {
	ID       int         `json:"id,omitempty"`
	Text     string      `json:"text"`
	Media    []MediaItem `json:"media,omitempty"`
	Position int         `json:"position,omitempty"`
}

// Text builds the content of a single-block post.
func Text(text string) []ContentBlock {
	return []ContentBlock{{Text: text}}
}

// Thread builds one content block per string, in order.
func Thread(texts ...string) []ContentBlock {
	blocks := make([]ContentBlock, 0, len(texts))
	for _, text := range texts {
		blocks = append(blocks, ContentBlock{Text: text})
	}
	return blocks
}

// ContentSignal is advisory feedback from a preflight check. Blockers arrive
// as issues instead.
type ContentSignal struct {
	// Level is "info" or "warn".
	Level   string `json:"level"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// HealthWarning flags an account whose credentials look shaky.
type HealthWarning struct {
	AccountID    string `json:"accountId"`
	Platform     string `json:"platform"`
	HealthStatus string `json:"healthStatus"`
	Message      string `json:"message"`
}

// Post statuses.
const (
	PostStatusDraft           = "draft"
	PostStatusScheduled       = "scheduled"
	PostStatusPublishing      = "publishing"
	PostStatusPublished       = "published"
	PostStatusPartiallyFailed = "partially_failed"
	PostStatusFailed          = "failed"
	PostStatusCancelled       = "cancelled"
)

// Delivery statuses, one per account on a post.
const (
	DeliveryStatusPending    = "pending"
	DeliveryStatusQueued     = "queued"
	DeliveryStatusDelayed    = "delayed"
	DeliveryStatusPublishing = "publishing"
	DeliveryStatusPublished  = "published"
	DeliveryStatusFailed     = "failed"
	DeliveryStatusCancelled  = "cancelled"
)

// Account health statuses.
const (
	HealthHealthy  = "healthy"
	HealthDegraded = "degraded"
	HealthExpired  = "expired"
	HealthRevoked  = "revoked"
	HealthUnknown  = "unknown"
)

// Webhook events a subscription can ask for.
const (
	EventPostPublished        = "post.published"
	EventPostFailed           = "post.failed"
	EventPostPartiallyFailed  = "post.partially_failed"
	EventDeliveryPublished    = "delivery.published"
	EventDeliveryFailed       = "delivery.failed"
	EventDeliveryDelayed      = "delivery.delayed"
	EventAccountHealthChanged = "account.health_changed"
)

// Platforms recognised when connecting an account with credentials.
const (
	PlatformTwitter           = "twitter"
	PlatformInstagram         = "instagram"
	PlatformInstagramBusiness = "instagram-business"
	PlatformFacebook          = "facebook"
	PlatformLinkedIn          = "linkedin"
	PlatformTikTok            = "tiktok"
	PlatformYouTube           = "youtube"
	PlatformBluesky           = "bluesky"
	PlatformThreads           = "threads"
	PlatformMastodon          = "mastodon"
	PlatformLemmy             = "lemmy"
	PlatformPinterest         = "pinterest"
	PlatformSnapchat          = "snapchat"
	PlatformTelegram          = "telegram"
	PlatformTwitch            = "twitch"
	PlatformDiscord           = "discord"
	PlatformSlack             = "slack"
	PlatformReddit            = "reddit"
	PlatformTumblr            = "tumblr"
	PlatformDribbble          = "dribbble"
	PlatformMeWe              = "mewe"
	PlatformDevTo             = "devto"
	PlatformHashnode          = "hashnode"
	PlatformMedium            = "medium"
	PlatformSubstack          = "substack"
	PlatformGoogleBusiness    = "google-business"
	PlatformKick              = "kick"
	PlatformListmonk          = "listmonk"
	PlatformWordPress         = "wordpress"
	PlatformNostr             = "nostr"
	PlatformWhop              = "whop"
	PlatformSkool             = "skool"
)

// queryBuilder collects query parameters, skipping the zero values the API
// reads as "not sent" and answers with its own default for.
type queryBuilder url.Values

func newQuery() queryBuilder { return queryBuilder{} }

func (q queryBuilder) values() url.Values { return url.Values(q) }

func (q queryBuilder) str(key, value string) {
	if value != "" {
		q[key] = []string{value}
	}
}

func (q queryBuilder) num(key string, value int) {
	if value != 0 {
		q[key] = []string{strconv.Itoa(value)}
	}
}

func (q queryBuilder) boolPtr(key string, value *bool) {
	if value != nil {
		q[key] = []string{strconv.FormatBool(*value)}
	}
}

func (q queryBuilder) time(key string, value *Time) {
	if value == nil {
		return
	}
	if s := strings.TrimSpace(value.String()); s != "" {
		q[key] = []string{s}
	}
}

// Bool returns a pointer to v, for the optional booleans in request structs.
func Bool(v bool) *bool { return &v }

// String returns a pointer to v, for the optional strings in request structs.
func String(v string) *string { return &v }

// Int returns a pointer to v, for the optional integers in request structs.
func Int(v int) *int { return &v }
