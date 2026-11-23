package fopost

import (
	"context"
	"net/url"
)

// GoogleAdsService covers the Google Ads surface no other network has: bid
// strategies, the ad schedule, Local Services leads, conversions and raw GAQL.
// Audiences and insights are on AdsService and dispatch by connection; a
// connection on another network answers 400 here.
//
// Every method needs the `ads` scope; anything that changes what a live
// account serves or bids also needs `publish`. CustomerID is digits only and
// has to name an account the connection's grant reaches: any other answers 404.
type GoogleAdsService struct{ client *Client }

// Google portfolio bid strategy types.
const (
	GoogleBidTargetSpend             = "TARGET_SPEND"
	GoogleBidMaximizeConversions     = "MAXIMIZE_CONVERSIONS"
	GoogleBidMaximizeConversionValue = "MAXIMIZE_CONVERSION_VALUE"
	GoogleBidTargetCPA               = "TARGET_CPA"
	GoogleBidTargetROAS              = "TARGET_ROAS"
)

// GoogleScope names the connection and the Google Ads account every call runs
// against. WorkspaceID may be empty on a read, and is required on a write.
type GoogleScope struct {
	WorkspaceID  string `json:"workspaceId,omitempty"`
	ConnectionID string `json:"connectionId"`
	CustomerID   string `json:"customerId"`
}

func (s GoogleScope) query(extra ...[2]string) url.Values {
	q := newQuery()
	q.str("workspace_id", s.WorkspaceID)
	q.str("connection_id", s.ConnectionID)
	q.str("customer_id", s.CustomerID)
	for _, pair := range extra {
		q.str(pair[0], pair[1])
	}
	return q.values()
}

// GoogleBidStrategy is a portfolio bid strategy on the account.
type GoogleBidStrategy struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	Status        string `json:"status"`
	CampaignCount int    `json:"campaignCount"`
}

// GoogleAdScheduleSlot is one slot of a campaign's ad schedule.
type GoogleAdScheduleSlot struct {
	ID          string   `json:"id"`
	DayOfWeek   string   `json:"dayOfWeek"`
	StartHour   int      `json:"startHour"`
	EndHour     int      `json:"endHour"`
	BidModifier *float64 `json:"bidModifier"`
}

// GoogleLocalServicesLead is a lead from Local Services Ads, read live on
// every call and never stored by FoPost.
type GoogleLocalServicesLead struct {
	ID          string `json:"id"`
	Category    string `json:"category"`
	Service     string `json:"service"`
	ContactName string `json:"contactName"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Status      string `json:"status"`
	Type        string `json:"type"`
	CreatedAt   string `json:"createdAt"`
}

// GoogleConversionAction is a conversion action on the account.
type GoogleConversionAction struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Category     string `json:"category"`
	Status       string `json:"status"`
	Type         string `json:"type"`
	CountingType string `json:"countingType"`
	ValueMinor   *int   `json:"valueMinor"`
}

// GoogleDateRange is inclusive and in the account's time zone, YYYY-MM-DD.
type GoogleDateRange struct {
	Since string
	Until string
}

// ─── Bid strategies and ad schedule ────────────────────────────────

// BidStrategies lists the portfolio bid strategies on the account.
func (s *GoogleAdsService) BidStrategies(ctx context.Context, scope GoogleScope) ([]GoogleBidStrategy, error) {
	var out []GoogleBidStrategy
	if err := s.client.json(ctx, "GET", "/ads/google/bid-strategies", nil, scope.query(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateGoogleBidStrategyRequest is the body of CreateBidStrategy. Type is one
// of the GoogleBid* constants.
type CreateGoogleBidStrategyRequest struct {
	GoogleScope
	Name string `json:"name"`
	Type string `json:"type"`
	// TargetMinor is in the account's currency, where the strategy takes a target.
	TargetMinor *int `json:"targetMinor,omitempty"`
}

// CreateBidStrategy adds a portfolio bid strategy. It needs `publish`.
func (s *GoogleAdsService) CreateBidStrategy(ctx context.Context, body *CreateGoogleBidStrategyRequest) (string, error) {
	return s.id(ctx, "POST", "/ads/google/bid-strategies", body)
}

// AdSchedule reads a campaign's ad schedule.
func (s *GoogleAdsService) AdSchedule(ctx context.Context, scope GoogleScope, campaignID string) ([]GoogleAdScheduleSlot, error) {
	var out []GoogleAdScheduleSlot
	q := scope.query([2]string{"campaign_id", campaignID})
	if err := s.client.json(ctx, "GET", "/ads/google/ad-schedule", nil, q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GoogleAdScheduleInput is one slot to set.
type GoogleAdScheduleInput struct {
	DayOfWeek   string   `json:"dayOfWeek"`
	StartHour   int      `json:"startHour"`
	EndHour     int      `json:"endHour"`
	BidModifier *float64 `json:"bidModifier,omitempty"`
}

// SetGoogleAdScheduleRequest replaces every slot on the campaign: Google has
// no partial edit for a schedule.
type SetGoogleAdScheduleRequest struct {
	GoogleScope
	CampaignID string                  `json:"campaignId"`
	Slots      []GoogleAdScheduleInput `json:"slots"`
}

// SetAdSchedule replaces a campaign's schedule. It needs `publish`.
func (s *GoogleAdsService) SetAdSchedule(ctx context.Context, body *SetGoogleAdScheduleRequest) (int, error) {
	var out struct {
		Slots int `json:"slots"`
	}
	if err := s.client.json(ctx, "PUT", "/ads/google/ad-schedule", body, nil, &out); err != nil {
		return 0, err
	}
	return out.Slots, nil
}

// ─── Local Services leads ──────────────────────────────────────────

// LocalServicesLeads reads leads from Local Services Ads over the range. They
// are read live and never stored by FoPost.
func (s *GoogleAdsService) LocalServicesLeads(ctx context.Context, scope GoogleScope, rng GoogleDateRange) ([]GoogleLocalServicesLead, error) {
	var out []GoogleLocalServicesLead
	q := scope.query([2]string{"since", rng.Since}, [2]string{"until", rng.Until})
	if err := s.client.json(ctx, "GET", "/ads/google/local-services", nil, q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ─── Conversions ───────────────────────────────────────────────────

// ConversionActions lists the account's conversion actions.
func (s *GoogleAdsService) ConversionActions(ctx context.Context, scope GoogleScope) ([]GoogleConversionAction, error) {
	var out []GoogleConversionAction
	if err := s.client.json(ctx, "GET", "/ads/google/conversions", nil, scope.query(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateGoogleConversionActionRequest is the body of CreateConversionAction.
type CreateGoogleConversionActionRequest struct {
	GoogleScope
	Name         string `json:"name"`
	Category     string `json:"category"`
	ValueMinor   *int   `json:"valueMinor,omitempty"`
	CountingType string `json:"countingType,omitempty"`
}

// CreateConversionAction adds a conversion action. It needs `publish`.
func (s *GoogleAdsService) CreateConversionAction(ctx context.Context, body *CreateGoogleConversionActionRequest) (string, error) {
	return s.id(ctx, "POST", "/ads/google/conversions", body)
}

// GoogleClickConversion is one offline conversion. One of GCLID, GBRAID or
// WBRAID is required: it is what matches the click.
type GoogleClickConversion struct {
	GCLID  string `json:"gclid,omitempty"`
	GBRAID string `json:"gbraid,omitempty"`
	WBRAID string `json:"wbraid,omitempty"`
	// ConversionActionID names the action the conversion counts against.
	ConversionActionID string `json:"conversionActionId"`
	// ConversionDateTime is `yyyy-MM-dd HH:mm:ss+|-HH:mm`, the only shape Google accepts.
	ConversionDateTime string `json:"conversionDateTime"`
	ValueMinor         *int   `json:"valueMinor,omitempty"`
	CurrencyCode       string `json:"currencyCode,omitempty"`
	OrderID            string `json:"orderId,omitempty"`
}

// UploadGoogleConversionsRequest is the body of UploadConversions.
type UploadGoogleConversionsRequest struct {
	GoogleScope
	Conversions []GoogleClickConversion `json:"conversions"`
}

// UploadConversions sends offline conversions and reports how many landed. It
// needs `publish`.
func (s *GoogleAdsService) UploadConversions(ctx context.Context, body *UploadGoogleConversionsRequest) (int, error) {
	return s.uploaded(ctx, "/ads/google/conversions/upload", body)
}

// GoogleConversionAdjustment restates, retracts or enhances a conversion
// already counted.
type GoogleConversionAdjustment struct {
	ConversionActionID string `json:"conversionActionId"`
	// AdjustmentType is RESTATEMENT, RETRACTION or ENHANCEMENT.
	AdjustmentType        string `json:"adjustmentType"`
	AdjustmentDateTime    string `json:"adjustmentDateTime"`
	OrderID               string `json:"orderId,omitempty"`
	GCLID                 string `json:"gclid,omitempty"`
	ConversionDateTime    string `json:"conversionDateTime,omitempty"`
	RestatementValueMinor *int   `json:"restatementValueMinor,omitempty"`
	CurrencyCode          string `json:"currencyCode,omitempty"`
}

// UploadGoogleConversionAdjustmentsRequest is the body of
// UploadConversionAdjustments.
type UploadGoogleConversionAdjustmentsRequest struct {
	GoogleScope
	Adjustments []GoogleConversionAdjustment `json:"adjustments"`
}

// UploadConversionAdjustments sends adjustments and reports how many landed.
// It needs `publish`.
func (s *GoogleAdsService) UploadConversionAdjustments(ctx context.Context, body *UploadGoogleConversionAdjustmentsRequest) (int, error) {
	return s.uploaded(ctx, "/ads/google/conversions/adjustments", body)
}

// ─── GAQL ──────────────────────────────────────────────────────────

// GoogleQueryRequest is a raw read-only GAQL SELECT. The account read is
// CustomerID, never anything named inside Query.
type GoogleQueryRequest struct {
	GoogleScope
	Query string `json:"query"`
}

// Query runs a GAQL SELECT and returns the rows exactly as Google sends them.
func (s *GoogleAdsService) Query(ctx context.Context, body *GoogleQueryRequest) ([]map[string]any, error) {
	var out struct {
		Rows []map[string]any `json:"rows"`
	}
	if err := s.client.json(ctx, "POST", "/ads/insights/query", body, nil, &out); err != nil {
		return nil, err
	}
	return out.Rows, nil
}

func (s *GoogleAdsService) id(ctx context.Context, method, path string, body any) (string, error) {
	var out struct {
		ID string `json:"id"`
	}
	if err := s.client.json(ctx, method, path, body, nil, &out); err != nil {
		return "", err
	}
	return out.ID, nil
}

func (s *GoogleAdsService) uploaded(ctx context.Context, path string, body any) (int, error) {
	var out struct {
		Uploaded int `json:"uploaded"`
	}
	if err := s.client.json(ctx, "POST", path, body, nil, &out); err != nil {
		return 0, err
	}
	return out.Uploaded, nil
}
