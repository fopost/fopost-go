package fopost

import (
	"context"
	"net/url"
	"strings"
)

// Company-list audiences, forecasts, conversion rules and the public ad
// library. Which of these a connection answers follows its network's
// Capabilities; one that does not support a call answers "unsupported".

// AdCompany is one row of a company-list upload. At least one of Name,
// Domain, PageURL or Ticker is required. The rows travel with the request and
// are never stored.
type AdCompany struct {
	Name   string `json:"name,omitempty"`
	Domain string `json:"domain,omitempty"`
	// PageURL is the company's page on the network.
	PageURL string `json:"pageUrl,omitempty"`
	// Ticker is the stock symbol, where the network matches on one.
	Ticker  string `json:"ticker,omitempty"`
	Country string `json:"country,omitempty"`
}

// AddAudienceCompanies adds companies to a company-list audience and returns
// how many the network took.
func (s *AdsService) AddAudienceCompanies(ctx context.Context, id string, params *AdObjectParams, companies []AdCompany) (int, error) {
	body := map[string][]AdCompany{"companies": companies}
	var out struct {
		Added int `json:"added"`
	}
	path := "/ads/audiences/" + url.PathEscape(id) + "/companies"
	if err := s.client.json(ctx, "POST", path, body, params.query(), &out); err != nil {
		return 0, err
	}
	return out.Added, nil
}

// AdForecastRequest is the shared body of BidPricing and SupplyForecast.
type AdForecastRequest struct {
	WorkspaceID  string `json:"workspaceId"`
	ConnectionID string `json:"connectionId"`
	// AdAccountID is the ad account as the network addresses it.
	AdAccountID string `json:"adAccountId"`
	// Goal is one of the AdGoal constants.
	Goal       string      `json:"goal"`
	Targeting  AdTargeting `json:"targeting"`
	Placements []string    `json:"placements,omitempty"`
	// BidType is "CPC", "CPM" or "CPV"; BidPricing only.
	BidType string `json:"bidType,omitempty"`
	// BudgetMinor is the budget for the forecast window; SupplyForecast only.
	BudgetMinor int `json:"budgetMinor,omitempty"`
}

// BidPricing is what the auction costs, in minor units of the ad account
// currency.
type BidPricing struct {
	Currency              string `json:"currency"`
	SuggestedBidMinor     *int   `json:"suggestedBidMinor"`
	MinBidMinor           *int   `json:"minBidMinor"`
	MaxBidMinor           *int   `json:"maxBidMinor"`
	DailyBudgetFloorMinor *int   `json:"dailyBudgetFloorMinor"`
}

// BidPricing quotes what the auction currently costs for that audience.
func (s *AdsService) BidPricing(ctx context.Context, body *AdForecastRequest) (*BidPricing, error) {
	var out BidPricing
	if err := s.client.json(ctx, "POST", "/ads/linkedin/bid-pricing", body, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SupplyForecast is what an audience would deliver at a budget, over the
// network's own window. Ready is false while the network has no answer.
type SupplyForecast struct {
	Currency    string `json:"currency"`
	Impressions *int   `json:"impressions"`
	Clicks      *int   `json:"clicks"`
	SpendMinor  *int   `json:"spendMinor"`
	// WindowDays is how many days the numbers cover.
	WindowDays *int `json:"windowDays"`
	Ready      bool `json:"ready"`
}

// SupplyForecast asks what that audience would deliver at that budget.
func (s *AdsService) SupplyForecast(ctx context.Context, body *AdForecastRequest) (*SupplyForecast, error) {
	var out SupplyForecast
	if err := s.client.json(ctx, "POST", "/ads/linkedin/supply-forecast", body, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Conversion rule types, as the API names them.
const (
	ConversionPurchase      = "purchase"
	ConversionLead          = "lead"
	ConversionSignUp        = "sign_up"
	ConversionAddToCart     = "add_to_cart"
	ConversionDownload      = "download"
	ConversionInstall       = "install"
	ConversionKeyPageView   = "key_page_view"
	ConversionOther         = "other"
	AttributionLastTouch    = "last_touch"
	AttributionEachCampaign = "each_campaign"
)

// ConversionRule is how the network attributes a sale or a sign-up back to an
// ad set.
type ConversionRule struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// Type is one of the Conversion constants.
	Type string `json:"type"`
	// Attribution is one of the Attribution constants.
	Attribution           string `json:"attribution"`
	PostClickWindowDays   int    `json:"postClickWindowDays"`
	ViewThroughWindowDays int    `json:"viewThroughWindowDays"`
	// ValueMinor is the default value of one conversion.
	ValueMinor *int   `json:"valueMinor"`
	Currency   string `json:"currency"`
	Enabled    bool   `json:"enabled"`
	CreatedAt  string `json:"createdAt"`
	// CampaignIDs are the ad sets this rule is attached to.
	CampaignIDs []string `json:"campaignIds"`
}

// CreateConversionRuleRequest is the body of CreateConversionRule.
type CreateConversionRuleRequest struct {
	WorkspaceID           string `json:"workspaceId"`
	ConnectionID          string `json:"connectionId"`
	AdAccountID           string `json:"adAccountId"`
	Name                  string `json:"name"`
	Type                  string `json:"type"`
	Attribution           string `json:"attribution"`
	PostClickWindowDays   int    `json:"postClickWindowDays,omitempty"`
	ViewThroughWindowDays int    `json:"viewThroughWindowDays,omitempty"`
	ValueMinor            int    `json:"valueMinor,omitempty"`
	Currency              string `json:"currency,omitempty"`
}

// UpdateConversionRuleRequest is the body of UpdateConversionRule. Only the
// fields you set are changed.
type UpdateConversionRuleRequest struct {
	Name                  string `json:"name,omitempty"`
	Type                  string `json:"type,omitempty"`
	Attribution           string `json:"attribution,omitempty"`
	PostClickWindowDays   int    `json:"postClickWindowDays,omitempty"`
	ViewThroughWindowDays int    `json:"viewThroughWindowDays,omitempty"`
	ValueMinor            int    `json:"valueMinor,omitempty"`
	Currency              string `json:"currency,omitempty"`
	Enabled               *bool  `json:"enabled,omitempty"`
}

// ListConversionRulesParams scopes a rule listing to one ad account.
type ListConversionRulesParams struct {
	WorkspaceID  string
	ConnectionID string
	AdAccountID  string
}

// ConversionRules lists the conversion rules on one ad account.
func (s *AdsService) ConversionRules(ctx context.Context, params *ListConversionRulesParams) ([]ConversionRule, error) {
	q := newQuery()
	if params != nil {
		q.str("workspace_id", params.WorkspaceID)
		q.str("connection_id", params.ConnectionID)
		q.str("ad_account_id", params.AdAccountID)
	}
	var out []ConversionRule
	if err := s.client.json(ctx, "GET", "/ads/linkedin/conversion-rules", nil, q.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateConversionRule returns the new rule's id.
func (s *AdsService) CreateConversionRule(ctx context.Context, body *CreateConversionRuleRequest) (string, error) {
	var out struct {
		ID string `json:"id"`
	}
	if err := s.client.json(ctx, "POST", "/ads/linkedin/conversion-rules", body, nil, &out); err != nil {
		return "", err
	}
	return out.ID, nil
}

// ConversionRule reads one rule with the ad sets it is attached to.
func (s *AdsService) ConversionRule(ctx context.Context, id string, params *AdObjectParams) (*ConversionRule, error) {
	var out ConversionRule
	if err := s.client.json(ctx, "GET", conversionRulePath(id, ""), nil, params.query(), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateConversionRule changes a rule and returns it.
func (s *AdsService) UpdateConversionRule(ctx context.Context, id string, params *AdObjectParams, body *UpdateConversionRuleRequest) (*ConversionRule, error) {
	var out ConversionRule
	if err := s.client.json(ctx, "PATCH", conversionRulePath(id, ""), body, params.query(), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteConversionRule turns the rule off; the network keeps the history.
func (s *AdsService) DeleteConversionRule(ctx context.Context, id string, params *AdObjectParams) error {
	return s.client.Do(ctx, "DELETE", conversionRulePath(id, ""), nil, params.query(), nil)
}

// AttachConversionRule attaches a rule to an ad set on the same connection.
func (s *AdsService) AttachConversionRule(ctx context.Context, id string, params *AdObjectParams, campaignID string) (*ConversionRule, error) {
	return s.association(ctx, "POST", id, params, campaignID)
}

// DetachConversionRule detaches a rule from an ad set.
func (s *AdsService) DetachConversionRule(ctx context.Context, id string, params *AdObjectParams, campaignID string) (*ConversionRule, error) {
	return s.association(ctx, "DELETE", id, params, campaignID)
}

func (s *AdsService) association(ctx context.Context, method, id string, params *AdObjectParams, campaignID string) (*ConversionRule, error) {
	body := map[string]string{"campaignId": campaignID}
	var out ConversionRule
	if err := s.client.json(ctx, method, conversionRulePath(id, "/associations"), body, params.query(), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ConversionMetrics is what a rule recorded over a date range.
type ConversionMetrics struct {
	Conversions            int  `json:"conversions"`
	PostClickConversions   int  `json:"postClickConversions"`
	ViewThroughConversions int  `json:"viewThroughConversions"`
	ValueMinor             int  `json:"valueMinor"`
	CostPerConversionMinor *int `json:"costPerConversionMinor"`
}

// ConversionMetricsParams is the date range a metrics read covers. Since and
// Until are YYYY-MM-DD, inclusive.
type ConversionMetricsParams struct {
	WorkspaceID  string
	ConnectionID string
	Since        string
	Until        string
}

// ConversionMetrics reads what a rule recorded.
func (s *AdsService) ConversionMetrics(ctx context.Context, id string, params *ConversionMetricsParams) (*ConversionMetrics, error) {
	q := newQuery()
	if params != nil {
		q.str("workspace_id", params.WorkspaceID)
		q.str("connection_id", params.ConnectionID)
		q.str("since", params.Since)
		q.str("until", params.Until)
	}
	var out ConversionMetrics
	if err := s.client.json(ctx, "GET", conversionRulePath(id, "/metrics"), nil, q.values(), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ConversionEvent is one conversion sent back to the network. It needs an
// Email or a ClickID. The address is hashed inside the API, so the network
// never receives it and nothing about an event is stored.
type ConversionEvent struct {
	// HappenedAt is epoch milliseconds.
	HappenedAt int64  `json:"happenedAt"`
	ValueMinor int    `json:"valueMinor,omitempty"`
	Currency   string `json:"currency,omitempty"`
	// EventID is your own id, so a replay is counted once.
	EventID string `json:"eventId,omitempty"`
	Email   string `json:"email,omitempty"`
	// ClickID is the network's click id, as the landing page received it.
	ClickID string `json:"clickId,omitempty"`
}

// SendConversionEvents sends conversions back to the network and returns how
// many it took. At most 100 per call.
func (s *AdsService) SendConversionEvents(ctx context.Context, id string, params *AdObjectParams, events []ConversionEvent) (int, error) {
	body := map[string][]ConversionEvent{"events": events}
	var out struct {
		Accepted int `json:"accepted"`
	}
	if err := s.client.json(ctx, "POST", conversionRulePath(id, "/events"), body, params.query(), &out); err != nil {
		return 0, err
	}
	return out.Accepted, nil
}

func conversionRulePath(id, suffix string) string {
	return "/ads/linkedin/conversion-rules/" + url.PathEscape(id) + suffix
}

// AdLibraryParams narrows an ad-library search. Countries are ISO 3166-1
// alpha-2 codes; Since and Until are YYYY-MM-DD.
type AdLibraryParams struct {
	WorkspaceID  string
	ConnectionID string
	Keyword      string
	Advertiser   string
	Countries    []string
	Since        string
	Until        string
	Cursor       string
}

// AdLibraryAd is a public ad from the network's own library, never a
// connection's own data.
type AdLibraryAd struct {
	ID                string   `json:"id"`
	AdvertiserName    string   `json:"advertiserName"`
	AdvertiserURL     string   `json:"advertiserUrl"`
	Headline          string   `json:"headline"`
	Body              string   `json:"body"`
	Type              string   `json:"type"`
	ThumbnailURL      string   `json:"thumbnailUrl"`
	FirstImpressionAt string   `json:"firstImpressionAt"`
	LastImpressionAt  string   `json:"lastImpressionAt"`
	Countries         []string `json:"countries"`
	DetailsURL        string   `json:"detailsUrl"`
	// Payer is the paying entity, where the network discloses one.
	Payer            string `json:"payer"`
	ImpressionsRange string `json:"impressionsRange"`
}

// AdLibraryPage is one page of results; pass NextCursor back as Cursor.
type AdLibraryPage struct {
	Ads        []AdLibraryAd `json:"ads"`
	NextCursor string        `json:"nextCursor"`
}

// AdLibrary searches the network's own public ad library, not the
// connection's ads.
func (s *AdsService) AdLibrary(ctx context.Context, params *AdLibraryParams) (*AdLibraryPage, error) {
	q := newQuery()
	if params != nil {
		q.str("workspace_id", params.WorkspaceID)
		q.str("connection_id", params.ConnectionID)
		q.str("keyword", params.Keyword)
		q.str("advertiser", params.Advertiser)
		q.str("countries", strings.Join(params.Countries, ","))
		q.str("since", params.Since)
		q.str("until", params.Until)
		q.str("cursor", params.Cursor)
	}
	var out AdLibraryPage
	if err := s.client.json(ctx, "GET", "/ads/ad-library", nil, q.values(), &out); err != nil {
		return nil, err
	}
	return &out, nil
}
