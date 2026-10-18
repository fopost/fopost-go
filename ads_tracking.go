package fopost

import (
	"context"
	"net/http"
	"net/url"
)

// Where a conversion happened, for ConversionEvent.ActionSource.
const (
	ConversionActionSourceWebsite           = "website"
	ConversionActionSourceApp               = "app"
	ConversionActionSourceEmail             = "email"
	ConversionActionSourcePhoneCall         = "phone_call"
	ConversionActionSourceChat              = "chat"
	ConversionActionSourcePhysicalStore     = "physical_store"
	ConversionActionSourceSystemGenerated   = "system_generated"
	ConversionActionSourceBusinessMessaging = "business_messaging"
	ConversionActionSourceOther             = "other"
)

// What a pixel's data may be used for, for AdPixelDetail.DataUseSetting.
const (
	PixelDataUseAdvertisingAndAnalytics = "advertising_and_analytics"
	PixelDataUseAnalyticsOnly           = "analytics_only"
)

// The standard event a custom conversion counts as, for CustomConversion.EventType.
const (
	CustomConversionAddPaymentInfo       = "ADD_PAYMENT_INFO"
	CustomConversionAddToCart            = "ADD_TO_CART"
	CustomConversionAddToWishlist        = "ADD_TO_WISHLIST"
	CustomConversionCompleteRegistration = "COMPLETE_REGISTRATION"
	CustomConversionContentView          = "CONTENT_VIEW"
	CustomConversionInitiatedCheckout    = "INITIATED_CHECKOUT"
	CustomConversionLead                 = "LEAD"
	CustomConversionPurchase             = "PURCHASE"
	CustomConversionSearch               = "SEARCH"
	CustomConversionContact              = "CONTACT"
	CustomConversionCustomizeProduct     = "CUSTOMIZE_PRODUCT"
	CustomConversionDonate               = "DONATE"
	CustomConversionFindLocation         = "FIND_LOCATION"
	CustomConversionSchedule             = "SCHEDULE"
	CustomConversionStartTrial           = "START_TRIAL"
	CustomConversionSubmitApplication    = "SUBMIT_APPLICATION"
	CustomConversionSubscribe            = "SUBSCRIBE"
	CustomConversionOther                = "OTHER"
)

// ConversionUploadResult is what a conversion batch came back with.
type ConversionUploadResult struct {
	// ID is the stored batch.
	ID string `json:"id"`
	// Accepted is how many events the network received.
	Accepted   int `json:"accepted"`
	EventCount int `json:"eventCount"`
	// Replayed is true when the Idempotency-Key had been used before and nothing was sent.
	Replayed bool `json:"replayed"`
}

// HashedConversionEvent is one event as a batch stored it: identity fields
// are SHA-256 hashes, the raw values never kept.
type HashedConversionEvent struct {
	EventName      string  `json:"eventName"`
	OccurredAt     string  `json:"occurredAt"`
	EventID        *string `json:"eventId"`
	ActionSource   string  `json:"actionSource"`
	EmailHash      *string `json:"emailHash"`
	PhoneHash      *string `json:"phoneHash"`
	ExternalIDHash *string `json:"externalIdHash"`
	ValueMinor     *int    `json:"valueMinor"`
	Currency       *string `json:"currency"`
	OrderID        *string `json:"orderId"`
}

// ConversionBatch is one stored conversion send.
type ConversionBatch struct {
	ID             string                  `json:"id"`
	WorkspaceID    string                  `json:"workspaceId"`
	ConnectionID   string                  `json:"connectionId"`
	AdAccountID    string                  `json:"adAccountId"`
	PixelID        string                  `json:"pixelId"`
	IdempotencyKey *string                 `json:"idempotencyKey"`
	EventCount     int                     `json:"eventCount"`
	Accepted       int                     `json:"accepted"`
	Events         []HashedConversionEvent `json:"events"`
	CreatedAt      string                  `json:"createdAt"`
}

// ListConversionEventsParams narrows a batch list. All fields are optional.
type ListConversionEventsParams struct {
	WorkspaceID  string
	ConnectionID string
	// Limit is up to 100, the default.
	Limit int
}

// AdPixelDetail is one pixel in full.
type AdPixelDetail struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// Code is the script to place on the site.
	Code          *string `json:"code"`
	CreatedAt     *string `json:"createdAt"`
	LastFiredAt   *string `json:"lastFiredAt"`
	IsUnavailable bool    `json:"isUnavailable"`
	// DataUseSetting is one of the PixelDataUse* constants.
	DataUseSetting *string `json:"dataUseSetting"`
}

// AdPixelStats is what a pixel fired over the last 28 days.
type AdPixelStats struct {
	Since  string              `json:"since"`
	Until  string              `json:"until"`
	Events []AdPixelEventCount `json:"events"`
}

// AdPixelEventCount is one event name and how often it fired.
type AdPixelEventCount struct {
	EventName string `json:"eventName"`
	Count     int    `json:"count"`
}

// EventMatchQuality is how well a pixel event's identity fields match.
type EventMatchQuality struct {
	EventName string `json:"eventName"`
	// Score is the network's match score out of 10.
	Score     *float64        `json:"score"`
	MatchKeys []MatchKeyShare `json:"matchKeys"`
}

// MatchKeyShare is how often one identifier was present on an event.
type MatchKeyShare struct {
	Identifier      string  `json:"identifier"`
	CoveragePercent float64 `json:"coveragePercent"`
}

// PixelParams names the pixel's ad account on a read.
type PixelParams struct {
	WorkspaceID  string
	ConnectionID string
	AdAccountID  string
}

func (p *PixelParams) query() url.Values {
	q := newQuery()
	if p != nil {
		q.str("workspace_id", p.WorkspaceID)
		q.str("connection_id", p.ConnectionID)
		q.str("ad_account_id", p.AdAccountID)
	}
	return q.values()
}

// CreatePixelRequest is the body of CreatePixel.
type CreatePixelRequest struct {
	WorkspaceID  string `json:"workspaceId"`
	ConnectionID string `json:"connectionId"`
	AdAccountID  string `json:"adAccountId"`
	Name         string `json:"name"`
}

// UpdatePixelRequest is the body of UpdatePixel.
type UpdatePixelRequest struct {
	WorkspaceID  string `json:"workspaceId"`
	ConnectionID string `json:"connectionId"`
	AdAccountID  string `json:"adAccountId"`
	Name         string `json:"name,omitempty"`
	// DataUseSetting is one of the PixelDataUse* constants.
	DataUseSetting string `json:"dataUseSetting,omitempty"`
}

// SharePixelRequest is the body of SharePixel.
type SharePixelRequest struct {
	WorkspaceID  string `json:"workspaceId"`
	ConnectionID string `json:"connectionId"`
	AdAccountID  string `json:"adAccountId"`
	// TargetAdAccountID is another ad account under the same business portfolio.
	TargetAdAccountID string `json:"targetAdAccountId"`
}

// CustomConversionRule is what a pixel event has to look like to count.
type CustomConversionRule struct {
	// EventName is the pixel event the rule reads; `PageView` when empty.
	EventName string `json:"eventName,omitempty"`
	// Pick one of URLContains and URLEquals.
	URLContains string `json:"urlContains,omitempty"`
	URLEquals   string `json:"urlEquals,omitempty"`
}

// CustomConversion is a pixel-backed conversion the network attributes.
type CustomConversion struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	PixelID *string `json:"pixelId"`
	// EventType is one of the CustomConversion* constants.
	EventType   string                `json:"eventType"`
	Rule        *CustomConversionRule `json:"rule"`
	Description *string               `json:"description"`
	// DefaultValueMinor is the account currency in minor units.
	DefaultValueMinor *int    `json:"defaultValueMinor"`
	IsArchived        bool    `json:"isArchived"`
	IsUnavailable     bool    `json:"isUnavailable"`
	CreatedAt         *string `json:"createdAt"`
	LastFiredAt       *string `json:"lastFiredAt"`
}

// CreateCustomConversionRequest is the body of CreateCustomConversion.
type CreateCustomConversionRequest struct {
	WorkspaceID  string `json:"workspaceId"`
	ConnectionID string `json:"connectionId"`
	AdAccountID  string `json:"adAccountId"`
	PixelID      string `json:"pixelId"`
	Name         string `json:"name"`
	// EventType is one of the CustomConversion* constants.
	EventType   string               `json:"eventType"`
	Rule        CustomConversionRule `json:"rule"`
	Description string               `json:"description,omitempty"`
	// DefaultValueMinor is the account currency in minor units.
	DefaultValueMinor *int `json:"defaultValueMinor,omitempty"`
}

// UpdateCustomConversionRequest is the body of UpdateCustomConversion.
type UpdateCustomConversionRequest struct {
	WorkspaceID       string `json:"workspaceId"`
	ConnectionID      string `json:"connectionId"`
	AdAccountID       string `json:"adAccountId"`
	Name              string `json:"name,omitempty"`
	Description       string `json:"description,omitempty"`
	DefaultValueMinor *int   `json:"defaultValueMinor,omitempty"`
}

// DeleteCustomConversionRequest is the body of DeleteCustomConversion.
type DeleteCustomConversionRequest struct {
	WorkspaceID  string `json:"workspaceId"`
	ConnectionID string `json:"connectionId"`
	AdAccountID  string `json:"adAccountId"`
}

// ─── Conversions ───────────────────────────────────────────────────

// SendConversions sends conversions against a pixel exactly once: a replay of
// the same idempotencyKey answers Replayed true and sends nothing. The key is
// required. Up to 1000 events per call.
func (s *AdsService) SendConversions(ctx context.Context, idempotencyKey string, body *UploadConversionsRequest) (*ConversionUploadResult, error) {
	headers := http.Header{"Idempotency-Key": {idempotencyKey}}
	out := &ConversionUploadResult{}
	if err := s.client.jsonWithHeaders(ctx, "POST", "/ads/conversions/events", body, nil, headers, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListConversionEvents lists stored conversion batches, newest first.
func (s *AdsService) ListConversionEvents(ctx context.Context, params *ListConversionEventsParams) ([]ConversionBatch, error) {
	q := newQuery()
	if params != nil {
		q.str("workspace_id", params.WorkspaceID)
		q.str("connection_id", params.ConnectionID)
		q.num("limit", params.Limit)
	}
	var out []ConversionBatch
	if err := s.client.json(ctx, "GET", "/ads/conversions/events", nil, q.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ─── Pixels ────────────────────────────────────────────────────────

// ListPixels lists an ad account's pixels.
func (s *AdsService) ListPixels(ctx context.Context, params *PixelParams) ([]AdPixel, error) {
	var out []AdPixel
	if err := s.client.json(ctx, "GET", "/ads/pixels", nil, params.query(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreatePixel adds a pixel; Code on the answer is the script to install.
func (s *AdsService) CreatePixel(ctx context.Context, body *CreatePixelRequest) (*AdPixelDetail, error) {
	out := &AdPixelDetail{}
	if err := s.client.json(ctx, "POST", "/ads/pixels", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetPixel reads one pixel.
func (s *AdsService) GetPixel(ctx context.Context, id string, params *PixelParams) (*AdPixelDetail, error) {
	out := &AdPixelDetail{}
	if err := s.client.json(ctx, "GET", pixelPath(id, ""), nil, params.query(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdatePixel renames a pixel or changes its data use setting.
func (s *AdsService) UpdatePixel(ctx context.Context, id string, body *UpdatePixelRequest) (*AdPixelDetail, error) {
	out := &AdPixelDetail{}
	if err := s.client.json(ctx, "PATCH", pixelPath(id, ""), body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// SharePixel shares a pixel with another ad account in the same portfolio.
func (s *AdsService) SharePixel(ctx context.Context, id string, body *SharePixelRequest) error {
	return s.client.json(ctx, "POST", pixelPath(id, "/share"), body, nil, nil)
}

// GetPixelStats reads what a pixel fired over the last 28 days.
func (s *AdsService) GetPixelStats(ctx context.Context, id string, params *PixelParams) (*AdPixelStats, error) {
	out := &AdPixelStats{}
	if err := s.client.json(ctx, "GET", pixelPath(id, "/stats"), nil, params.query(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetEventMatchQuality reads the match quality of each event a pixel sends.
func (s *AdsService) GetEventMatchQuality(ctx context.Context, id string, params *PixelParams) ([]EventMatchQuality, error) {
	var out []EventMatchQuality
	if err := s.client.json(ctx, "GET", pixelPath(id, "/match-quality"), nil, params.query(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func pixelPath(id, suffix string) string {
	return "/ads/pixels/" + url.PathEscape(id) + suffix
}

// ─── Custom conversions ────────────────────────────────────────────

// ListCustomConversions lists an ad account's custom conversions.
func (s *AdsService) ListCustomConversions(ctx context.Context, params *PixelParams) ([]CustomConversion, error) {
	var out []CustomConversion
	if err := s.client.json(ctx, "GET", "/ads/custom-conversions", nil, params.query(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateCustomConversion adds a custom conversion on a pixel.
func (s *AdsService) CreateCustomConversion(ctx context.Context, body *CreateCustomConversionRequest) (*CustomConversion, error) {
	out := &CustomConversion{}
	if err := s.client.json(ctx, "POST", "/ads/custom-conversions", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetCustomConversion reads one custom conversion.
func (s *AdsService) GetCustomConversion(ctx context.Context, id string, params *PixelParams) (*CustomConversion, error) {
	out := &CustomConversion{}
	if err := s.client.json(ctx, "GET", customConversionPath(id), nil, params.query(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateCustomConversion changes a custom conversion's name, description or
// default value.
func (s *AdsService) UpdateCustomConversion(ctx context.Context, id string, body *UpdateCustomConversionRequest) (*CustomConversion, error) {
	out := &CustomConversion{}
	if err := s.client.json(ctx, "PATCH", customConversionPath(id), body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteCustomConversion removes a custom conversion.
func (s *AdsService) DeleteCustomConversion(ctx context.Context, id string, body *DeleteCustomConversionRequest) error {
	return s.client.json(ctx, "DELETE", customConversionPath(id), body, nil, nil)
}

func customConversionPath(id string) string {
	return "/ads/custom-conversions/" + url.PathEscape(id)
}
