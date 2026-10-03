package fopost

import (
	"context"
	"net/url"
)

// GoogleAdsService covers the Google Ads surface no other network has:
// keywords, assets, Performance Max asset groups, Local Services leads,
// conversions and raw GAQL. Campaigns, ad groups, ads, audiences and insights
// are on AdsService and dispatch by connection; a connection on another
// network answers 400 here.
//
// Every method needs the `ads` scope; anything that changes what a live
// account serves or bids also needs `publish`. CustomerID is digits only and
// has to name an account the connection's grant reaches: any other answers 404.
type GoogleAdsService struct{ client *Client }

// Google keyword match types.
const (
	GoogleMatchExact  = "EXACT"
	GoogleMatchPhrase = "PHRASE"
	GoogleMatchBroad  = "BROAD"
)

// Google portfolio bid strategy types.
const (
	GoogleBidTargetSpend             = "TARGET_SPEND"
	GoogleBidMaximizeConversions     = "MAXIMIZE_CONVERSIONS"
	GoogleBidMaximizeConversionValue = "MAXIMIZE_CONVERSION_VALUE"
	GoogleBidTargetCPA               = "TARGET_CPA"
	GoogleBidTargetROAS              = "TARGET_ROAS"
)

// Google asset field types, for attaching an asset.
const (
	GoogleFieldSitelink          = "SITELINK"
	GoogleFieldCallout           = "CALLOUT"
	GoogleFieldStructuredSnippet = "STRUCTURED_SNIPPET"
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

// GoogleKeyword is one keyword on an ad group. ID is
// `<customerID>~keyword~<adGroupID>~<criterionID>`.
type GoogleKeyword struct {
	ID        string `json:"id"`
	AdGroupID string `json:"adGroupId"`
	Text      string `json:"text"`
	MatchType string `json:"matchType"`
	Status    string `json:"status"`
	// CPCBidMinor is in the account's currency, minor units.
	CPCBidMinor *int `json:"cpcBidMinor"`
	Negative    bool `json:"negative"`
}

// GoogleKeywordIdea is a keyword idea or the historical metrics of one.
type GoogleKeywordIdea struct {
	Text                  string `json:"text"`
	AvgMonthlySearches    int    `json:"avgMonthlySearches"`
	Competition           string `json:"competition"`
	LowTopOfPageBidMinor  *int   `json:"lowTopOfPageBidMinor"`
	HighTopOfPageBidMinor *int   `json:"highTopOfPageBidMinor"`
}

// GoogleSearchTerm is what someone actually searched, with what it earned.
type GoogleSearchTerm struct {
	Term      string          `json:"term"`
	AdGroupID string          `json:"adGroupId"`
	Status    string          `json:"status"`
	Metrics   InsightsMetrics `json:"metrics"`
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

// GoogleSharedSet is a negative keyword list.
type GoogleSharedSet struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	MemberCount int    `json:"memberCount"`
}

// GoogleAsset is a sitelink, callout or structured snippet.
type GoogleAsset struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	// Text is what a sitelink, callout or snippet renders.
	Text     string `json:"text"`
	FinalURL string `json:"finalUrl"`
}

// GoogleAssetLink is where an asset is attached. An asset with no links is in
// the library and serving nowhere.
type GoogleAssetLink struct {
	ID        string `json:"id"`
	AssetID   string `json:"assetId"`
	Level     string `json:"level"`
	OwnerID   string `json:"ownerId"`
	FieldType string `json:"fieldType"`
	Status    string `json:"status"`
}

// GoogleAssetsResult is the account's assets with the links that place them.
type GoogleAssetsResult struct {
	Assets []GoogleAsset     `json:"assets"`
	Links  []GoogleAssetLink `json:"links"`
}

// GoogleAssetGroup is a Performance Max asset group.
type GoogleAssetGroup struct {
	ID         string   `json:"id"`
	CampaignID string   `json:"campaignId"`
	Name       string   `json:"name"`
	Status     string   `json:"status"`
	FinalURLs  []string `json:"finalUrls"`
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

// ─── Keywords ──────────────────────────────────────────────────────

// ListGoogleKeywordsParams narrows the listing to one ad group.
type ListGoogleKeywordsParams struct {
	AdGroupID string
}

// Keywords lists the keywords on the account, or on one ad group.
func (s *GoogleAdsService) Keywords(ctx context.Context, scope GoogleScope, params *ListGoogleKeywordsParams) ([]GoogleKeyword, error) {
	adGroupID := ""
	if params != nil {
		adGroupID = params.AdGroupID
	}
	var out []GoogleKeyword
	if err := s.client.json(ctx, "GET", "/ads/google/keywords", nil, scope.query([2]string{"ad_group_id", adGroupID}), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateGoogleKeywordRequest is the body of CreateKeyword.
type CreateGoogleKeywordRequest struct {
	GoogleScope
	AdGroupID string `json:"adGroupId"`
	Text      string `json:"text"`
	MatchType string `json:"matchType"`
	// CPCBidMinor is in the account's currency, minor units.
	CPCBidMinor *int `json:"cpcBidMinor,omitempty"`
}

// CreateKeyword adds a keyword to an ad group. It needs `publish` as well as
// `ads`: the keyword goes live.
func (s *GoogleAdsService) CreateKeyword(ctx context.Context, body *CreateGoogleKeywordRequest) (string, error) {
	return s.id(ctx, "POST", "/ads/google/keywords", body)
}

// UpdateGoogleKeywordRequest is the body of UpdateKeyword. Status is
// AdStatusActive or AdStatusPaused.
type UpdateGoogleKeywordRequest struct {
	GoogleScope
	Status      string `json:"status,omitempty"`
	CPCBidMinor *int   `json:"cpcBidMinor,omitempty"`
}

// UpdateKeyword pauses, resumes or rebids a keyword. It needs `publish`.
func (s *GoogleAdsService) UpdateKeyword(ctx context.Context, id string, body *UpdateGoogleKeywordRequest) (string, error) {
	return s.id(ctx, "PATCH", "/ads/google/keywords/"+url.PathEscape(id), body)
}

// DeleteKeyword removes a keyword. It needs `publish` as well as `ads`.
func (s *GoogleAdsService) DeleteKeyword(ctx context.Context, id string, scope GoogleScope) error {
	return s.client.Do(ctx, "DELETE", "/ads/google/keywords/"+url.PathEscape(id), scope, nil, nil)
}

// GoogleKeywordIdeasRequest asks for ideas from seed keywords, a landing page,
// or both.
type GoogleKeywordIdeasRequest struct {
	GoogleScope
	Seeds        []string `json:"seeds,omitempty"`
	URL          string   `json:"url,omitempty"`
	LanguageID   string   `json:"languageId,omitempty"`
	GeoTargetIDs []string `json:"geoTargetIds,omitempty"`
}

// KeywordIdeas suggests keywords to bid on.
func (s *GoogleAdsService) KeywordIdeas(ctx context.Context, body *GoogleKeywordIdeasRequest) ([]GoogleKeywordIdea, error) {
	var out []GoogleKeywordIdea
	if err := s.client.json(ctx, "POST", "/ads/google/keyword-ideas", body, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GoogleKeywordMetricsRequest reads historical metrics for keywords you have.
type GoogleKeywordMetricsRequest struct {
	GoogleScope
	Keywords []string `json:"keywords"`
}

// KeywordMetrics reads the historical metrics of the keywords given.
func (s *GoogleAdsService) KeywordMetrics(ctx context.Context, body *GoogleKeywordMetricsRequest) ([]GoogleKeywordIdea, error) {
	var out []GoogleKeywordIdea
	if err := s.client.json(ctx, "POST", "/ads/google/keyword-metrics", body, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SearchTerms reports what people actually searched, with the metrics each
// term earned over the range.
func (s *GoogleAdsService) SearchTerms(ctx context.Context, scope GoogleScope, rng GoogleDateRange) ([]GoogleSearchTerm, error) {
	var out []GoogleSearchTerm
	q := scope.query([2]string{"since", rng.Since}, [2]string{"until", rng.Until})
	if err := s.client.json(ctx, "GET", "/ads/google/search-terms", nil, q, &out); err != nil {
		return nil, err
	}
	return out, nil
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

// ─── Negative keyword lists ────────────────────────────────────────

// NegativeKeywordLists lists the account's negative keyword lists.
func (s *GoogleAdsService) NegativeKeywordLists(ctx context.Context, scope GoogleScope) ([]GoogleSharedSet, error) {
	var out []GoogleSharedSet
	if err := s.client.json(ctx, "GET", "/ads/google/negative-keywords", nil, scope.query(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateGoogleNegativeKeywordListRequest is the body of
// CreateNegativeKeywordList.
type CreateGoogleNegativeKeywordListRequest struct {
	GoogleScope
	Name string `json:"name"`
}

// CreateNegativeKeywordList adds a list. It needs `publish`.
func (s *GoogleAdsService) CreateNegativeKeywordList(ctx context.Context, body *CreateGoogleNegativeKeywordListRequest) (string, error) {
	return s.id(ctx, "POST", "/ads/google/negative-keywords", body)
}

// GoogleNegativeKeyword is one keyword in a negative list.
type GoogleNegativeKeyword struct {
	Text      string `json:"text"`
	MatchType string `json:"matchType"`
}

// AddGoogleNegativeKeywordsRequest is the body of AddNegativeKeywords.
type AddGoogleNegativeKeywordsRequest struct {
	GoogleScope
	SharedSetID string                  `json:"sharedSetId"`
	Keywords    []GoogleNegativeKeyword `json:"keywords"`
}

// AddNegativeKeywords adds keywords to a list and reports how many landed. It
// needs `publish`.
func (s *GoogleAdsService) AddNegativeKeywords(ctx context.Context, body *AddGoogleNegativeKeywordsRequest) (int, error) {
	var out struct {
		Added int `json:"added"`
	}
	if err := s.client.json(ctx, "POST", "/ads/google/negative-keywords/keywords", body, nil, &out); err != nil {
		return 0, err
	}
	return out.Added, nil
}

// AttachGoogleNegativeKeywordListRequest is the body of
// AttachNegativeKeywordList.
type AttachGoogleNegativeKeywordListRequest struct {
	GoogleScope
	SharedSetID string `json:"sharedSetId"`
	CampaignID  string `json:"campaignId"`
}

// AttachNegativeKeywordList puts a list on a campaign. It needs `publish`.
func (s *GoogleAdsService) AttachNegativeKeywordList(ctx context.Context, body *AttachGoogleNegativeKeywordListRequest) error {
	return s.client.Do(ctx, "POST", "/ads/google/negative-keywords/attach", body, nil, nil)
}

// ─── Assets ────────────────────────────────────────────────────────

// Assets lists the account's sitelinks, callouts and snippets, with the links
// that put each one under an ad.
func (s *GoogleAdsService) Assets(ctx context.Context, scope GoogleScope) (*GoogleAssetsResult, error) {
	out := &GoogleAssetsResult{}
	if err := s.client.json(ctx, "GET", "/ads/google/assets", nil, scope.query(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// GoogleAssetSpec describes the asset to create. Kind picks which other fields
// apply: Text and FinalURL for "sitelink", Text alone for "callout", Header and
// Values for "snippet".
type GoogleAssetSpec struct {
	Kind         string   `json:"kind"`
	Text         string   `json:"text,omitempty"`
	Description1 string   `json:"description1,omitempty"`
	Description2 string   `json:"description2,omitempty"`
	FinalURL     string   `json:"finalUrl,omitempty"`
	Header       string   `json:"header,omitempty"`
	Values       []string `json:"values,omitempty"`
}

// CreateGoogleAssetRequest is the body of CreateAsset.
type CreateGoogleAssetRequest struct {
	GoogleScope
	Spec GoogleAssetSpec `json:"spec"`
}

// CreateAsset adds an asset to the library. It needs `publish`.
func (s *GoogleAdsService) CreateAsset(ctx context.Context, body *CreateGoogleAssetRequest) (string, error) {
	return s.id(ctx, "POST", "/ads/google/assets", body)
}

// AttachGoogleAssetRequest attaches an asset to the account, or to one
// campaign when CampaignID is set.
type AttachGoogleAssetRequest struct {
	GoogleScope
	AssetID    string `json:"assetId"`
	FieldType  string `json:"fieldType"`
	CampaignID string `json:"campaignId,omitempty"`
}

// AttachAsset puts an asset under the ads it belongs to. It needs `publish`.
func (s *GoogleAdsService) AttachAsset(ctx context.Context, body *AttachGoogleAssetRequest) error {
	return s.client.Do(ctx, "POST", "/ads/google/assets/attach", body, nil, nil)
}

// DeleteAsset removes the links that put an asset under an ad; on Google the
// asset itself is permanent. It needs `publish`.
func (s *GoogleAdsService) DeleteAsset(ctx context.Context, id string, scope GoogleScope) error {
	return s.client.Do(ctx, "DELETE", "/ads/google/assets/"+url.PathEscape(id), scope, nil, nil)
}

// ─── Performance Max asset groups ──────────────────────────────────

// ListGoogleAssetGroupsParams narrows the listing to one campaign.
type ListGoogleAssetGroupsParams struct {
	CampaignID string
}

// AssetGroups lists the Performance Max asset groups on the account.
func (s *GoogleAdsService) AssetGroups(ctx context.Context, scope GoogleScope, params *ListGoogleAssetGroupsParams) ([]GoogleAssetGroup, error) {
	campaignID := ""
	if params != nil {
		campaignID = params.CampaignID
	}
	var out []GoogleAssetGroup
	q := scope.query([2]string{"campaign_id", campaignID})
	if err := s.client.json(ctx, "GET", "/ads/google/asset-groups", nil, q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateGoogleAssetGroupRequest is the body of CreateAssetGroup. The group
// starts paused unless Status is AdStatusActive.
type CreateGoogleAssetGroupRequest struct {
	GoogleScope
	CampaignID string   `json:"campaignId"`
	Name       string   `json:"name"`
	FinalURLs  []string `json:"finalUrls"`
	Status     string   `json:"status,omitempty"`
}

// CreateAssetGroup adds an asset group. It needs `publish`.
func (s *GoogleAdsService) CreateAssetGroup(ctx context.Context, body *CreateGoogleAssetGroupRequest) (string, error) {
	return s.id(ctx, "POST", "/ads/google/asset-groups", body)
}

// UpdateGoogleAssetGroupRequest is the body of UpdateAssetGroup.
type UpdateGoogleAssetGroupRequest struct {
	GoogleScope
	Name   string `json:"name,omitempty"`
	Status string `json:"status,omitempty"`
}

// UpdateAssetGroup renames, pauses or resumes an asset group. It needs
// `publish`.
func (s *GoogleAdsService) UpdateAssetGroup(ctx context.Context, id string, body *UpdateGoogleAssetGroupRequest) (string, error) {
	return s.id(ctx, "PATCH", "/ads/google/asset-groups/"+url.PathEscape(id), body)
}

// DeleteAssetGroup removes an asset group. It needs `publish`.
func (s *GoogleAdsService) DeleteAssetGroup(ctx context.Context, id string, scope GoogleScope) error {
	return s.client.Do(ctx, "DELETE", "/ads/google/asset-groups/"+url.PathEscape(id), scope, nil, nil)
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
