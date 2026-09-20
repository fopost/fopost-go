package fopost

import (
	"context"
	"net/url"
)

// AdsService covers ads across the ad networks: boosts and ads created through FoPost, the ad
// connections they run on, audiences, targeting search and lead forms. Every
// method needs the `ads` scope; Boost, Create, SetStatus, Delete, the create,
// update, delete and duplicate calls on campaigns, ad sets and network ads, and
// SetStatuses spend money and also need `publish`.
type AdsService struct{ client *Client }

// Ad goals.
const (
	AdGoalEngagement = "engagement"
	AdGoalTraffic    = "traffic"
	AdGoalAwareness  = "awareness"
	AdGoalVideoViews = "video_views"
)

// Ad budget types.
const (
	AdBudgetDaily    = "daily"
	AdBudgetLifetime = "lifetime"
)

// Ad statuses for SetStatus.
const (
	AdStatusActive = "active"
	AdStatusPaused = "paused"
)

// Targeting search types for SearchTargeting.
const (
	TargetingCountry  = "country"
	TargetingRegion   = "region"
	TargetingCity     = "city"
	TargetingZip      = "zip"
	TargetingMetro    = "metro"
	TargetingInterest = "interest"
	TargetingBehavior = "behavior"
	TargetingIncome   = "income"
)

// Audience subtypes for CreateAudience.
const (
	AudienceCustom    = "CUSTOM"
	AudienceLookalike = "LOOKALIKE"
	AudienceWebsite   = "WEBSITE"
)

// AdTargetingItem is one interest, behaviour or income bracket as Meta names
// it, from SearchTargeting.
type AdTargetingItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AdTargetingLocation is a location below country level, from
// SearchTargeting. Type is "region", "city", "zip" or "geo_market".
type AdTargetingLocation struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// AdTargeting is who an ad is shown to. At least one country or one location
// is required.
type AdTargeting struct {
	// Countries are ISO 3166-1 alpha-2 codes.
	Countries []string `json:"countries"`
	AgeMin    int      `json:"ageMin"`
	AgeMax    int      `json:"ageMax"`
	// Gender is "all", "male" or "female".
	Gender      string                `json:"gender"`
	AudienceIDs []string              `json:"audienceIds,omitempty"`
	Locations   []AdTargetingLocation `json:"locations,omitempty"`
	Interests   []AdTargetingItem     `json:"interests,omitempty"`
	Behaviors   []AdTargetingItem     `json:"behaviors,omitempty"`
	Income      []AdTargetingItem     `json:"income,omitempty"`
}

// AdBudget is what an ad may spend. Minor is in the ad account currency's
// minor units.
type AdBudget struct {
	Minor int `json:"minor"`
	// Type is one of the AdBudget constants.
	Type  string `json:"type"`
	EndAt *Time  `json:"endAt,omitempty"`
}

// AdInsights are lifetime delivery numbers from the last refresh. SpendMinor
// is in the ad account currency's minor units.
type AdInsights struct {
	Impressions int `json:"impressions"`
	Reach       int `json:"reach"`
	Clicks      int `json:"clicks"`
	SpendMinor  int `json:"spendMinor"`
}

// AdCreative is the text and media of an ad created from scratch.
type AdCreative struct {
	Text           string `json:"text,omitempty"`
	Headline       string `json:"headline,omitempty"`
	DestinationURL string `json:"destinationUrl,omitempty"`
	MediaURL       string `json:"mediaUrl,omitempty"`
	URLTags        string `json:"urlTags,omitempty"`
}

// Ad is a boost or an ad created through FoPost.
type Ad struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspaceId"`
	// Kind is "boost" or "ad".
	Kind string `json:"kind"`
	Name string `json:"name"`
	// Goal is one of the AdGoal constants.
	Goal string `json:"goal"`
	// Status is what was asked for: "active" or "paused".
	Status string `json:"status"`
	// EffectiveStatus is Meta's own delivery status, from the last refresh.
	EffectiveStatus string `json:"effectiveStatus"`
	ConnectionID    string `json:"connectionId"`
	// AccountID is the connected account a boost was built from.
	AccountID   string `json:"accountId"`
	Platform    string `json:"platform"`
	AdAccountID string `json:"adAccountId"`
	// SourcePostID is the FoPost post a boost promotes.
	SourcePostID string      `json:"sourcePostId"`
	BudgetMinor  int         `json:"budgetMinor"`
	BudgetType   string      `json:"budgetType"`
	Currency     string      `json:"currency"`
	EndAt        Time        `json:"endAt"`
	Targeting    AdTargeting `json:"targeting"`
	Creative     *AdCreative `json:"creative"`
	Insights     *AdInsights `json:"insights"`
	InsightsAt   Time        `json:"insightsAt"`
	LastError    string      `json:"lastError"`
	CreatedAt    Time        `json:"createdAt"`
}

// ExternalAd is an ad on a connected ad account that was made elsewhere. It is
// read live from Meta, never stored.
type ExternalAd struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	EffectiveStatus string `json:"effectiveStatus"`
	CampaignID      string `json:"campaignId"`
	CampaignName    string `json:"campaignName"`
	Objective       string `json:"objective"`
	BudgetMinor     *int   `json:"budgetMinor"`
	BudgetType      string `json:"budgetType"`
	EndAt           Time   `json:"endAt"`
	CreatedAt       Time   `json:"createdAt"`
	ConnectionID    string `json:"connectionId"`
	AdAccountID     string `json:"adAccountId"`
	Currency        string `json:"currency"`
	WorkspaceID     string `json:"workspaceId"`
}

// AdConnection is one Meta Ads grant in a workspace.
type AdConnection struct {
	ID       string `json:"id"`
	Provider string `json:"provider"`
	// AuthType is "business" or "user".
	AuthType    string `json:"authType"`
	Name        string `json:"name"`
	BusinessID  string `json:"businessId"`
	CreatedAt   Time   `json:"createdAt"`
	WorkspaceID string `json:"workspaceId"`
}

// AdAccountRef is one ad account a connection reaches. ID is `act_…`; Status
// is Meta's account status code, 1 being active.
type AdAccountRef struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Currency string `json:"currency"`
	Status   int    `json:"status"`
}

// AdPageRef is one Page a connection reaches.
type AdPageRef struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	InstagramUserID string `json:"instagramUserId"`
}

// AdSource is a connection with the ad accounts and Pages its grant reaches.
// Error is set when Meta refused the listing, usually a revoked grant.
type AdSource struct {
	ConnectionID string         `json:"connectionId"`
	Name         string         `json:"name"`
	WorkspaceID  string         `json:"workspaceId"`
	AdAccounts   []AdAccountRef `json:"adAccounts"`
	Pages        []AdPageRef    `json:"pages"`
	Error        string         `json:"error"`
}

// BoostableDelivery is one published delivery a boost can be built from.
type BoostableDelivery struct {
	AccountID   string `json:"accountId"`
	Platform    string `json:"platform"`
	Username    string `json:"username"`
	ExternalURL string `json:"externalUrl"`
	PostedAt    Time   `json:"postedAt"`
}

// BoostablePost is a published post that can be boosted.
type BoostablePost struct {
	ID           string              `json:"id"`
	WorkspaceID  string              `json:"workspaceId"`
	Text         string              `json:"text"`
	ThumbnailURL string              `json:"thumbnailUrl"`
	Deliveries   []BoostableDelivery `json:"deliveries"`
}

// Audience is one custom, lookalike or website audience on an ad account.
type Audience struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Subtype        string `json:"subtype"`
	Description    string `json:"description"`
	SizeLower      *int   `json:"sizeLower"`
	SizeUpper      *int   `json:"sizeUpper"`
	DeliveryStatus string `json:"deliveryStatus"`
	CreatedAt      Time   `json:"createdAt"`
}

// AdPixel is one pixel on an ad account, for website audiences.
type AdPixel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AudiencesResult is an ad account's audiences and pixels.
type AudiencesResult struct {
	Audiences   []Audience `json:"audiences"`
	Pixels      []AdPixel  `json:"pixels"`
	WorkspaceID string     `json:"workspaceId"`
}

// TargetingOption is one match from SearchTargeting.
type TargetingOption struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Detail string `json:"detail"`
}

// LeadForm is one lead form on a Page.
type LeadForm struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Status     string   `json:"status"`
	LeadsCount int      `json:"leadsCount"`
	CreatedAt  Time     `json:"createdAt"`
	Questions  []string `json:"questions"`
}

// LeadFormSource is a connection's Page with its lead forms.
type LeadFormSource struct {
	ConnectionID   string     `json:"connectionId"`
	ConnectionName string     `json:"connectionName"`
	PageID         string     `json:"pageId"`
	PageName       string     `json:"pageName"`
	Forms          []LeadForm `json:"forms"`
	Error          string     `json:"error"`
	WorkspaceID    string     `json:"workspaceId"`
}

// LeadField is one answered question on a lead.
type LeadField struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

// Lead is one submission of a lead form.
type Lead struct {
	ID           string      `json:"id"`
	CreatedAt    Time        `json:"createdAt"`
	Fields       []LeadField `json:"fields"`
	AdName       string      `json:"adName"`
	CampaignName string      `json:"campaignName"`
	Platform     string      `json:"platform"`
	IsOrganic    bool        `json:"isOrganic"`
}

// LeadsPage is one page of leads. Pass NextCursor back as After for the next.
type LeadsPage struct {
	Leads      []Lead `json:"leads"`
	NextCursor string `json:"nextCursor"`
}

// List returns the boosts and ads created through FoPost, with insights from
// their last refresh, optionally narrowed to one workspace.
func (s *AdsService) List(ctx context.Context, workspaceID string) ([]Ad, error) {
	var out []Ad
	if err := s.client.json(ctx, "GET", "/ads", nil, workspaceQuery(workspaceID), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// External returns ads on the connected ad accounts that were made elsewhere.
func (s *AdsService) External(ctx context.Context, workspaceID string) ([]ExternalAd, error) {
	var out []ExternalAd
	if err := s.client.json(ctx, "GET", "/ads/external", nil, workspaceQuery(workspaceID), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Boostable returns the published posts a boost can be built from.
func (s *AdsService) Boostable(ctx context.Context, workspaceID string) ([]BoostablePost, error) {
	var out []BoostablePost
	if err := s.client.json(ctx, "GET", "/ads/boostable", nil, workspaceQuery(workspaceID), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Connections returns the Meta Ads connections the key can reach.
func (s *AdsService) Connections(ctx context.Context, workspaceID string) ([]AdConnection, error) {
	var out []AdConnection
	if err := s.client.json(ctx, "GET", "/ads/connections", nil, workspaceQuery(workspaceID), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Sources returns each connection with the ad accounts and Pages its grant
// reaches.
func (s *AdsService) Sources(ctx context.Context, workspaceID string) ([]AdSource, error) {
	var out []AdSource
	if err := s.client.json(ctx, "GET", "/ads/sources", nil, workspaceQuery(workspaceID), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// AuthorizeMetaAdsRequest is the body of AuthorizeMeta.
type AuthorizeMetaAdsRequest struct {
	WorkspaceID string `json:"workspaceId"`
	// Method is "business" (the default) for Facebook Login for Business, or
	// "user" for a personal login.
	Method string `json:"method,omitempty"`
	// ReturnTo is the dashboard path to land on after Meta redirects back.
	ReturnTo string `json:"returnTo,omitempty"`
}

// AuthorizeMeta returns the Meta login URL; the caller finishes it in a
// browser.
func (s *AdsService) AuthorizeMeta(ctx context.Context, body *AuthorizeMetaAdsRequest) (string, error) {
	var out struct {
		URL string `json:"url"`
	}
	if err := s.client.json(ctx, "POST", "/ads/connections/meta/authorize", body, nil, &out); err != nil {
		return "", err
	}
	return out.URL, nil
}

// AuthorizeGoogleAdsRequest is the body of AuthorizeGoogle.
type AuthorizeGoogleAdsRequest struct {
	WorkspaceID string `json:"workspaceId"`
	// ReturnTo is a dashboard path to land on after Google redirects back.
	ReturnTo string `json:"returnTo,omitempty"`
}

// AuthorizeGoogle returns the Google login URL. The caller finishes it in
// their own browser session: the callback checks the same user came back.
func (s *AdsService) AuthorizeGoogle(ctx context.Context, body *AuthorizeGoogleAdsRequest) (string, error) {
	var out struct {
		URL string `json:"url"`
	}
	if err := s.client.json(ctx, "POST", "/ads/connections/google/authorize", body, nil, &out); err != nil {
		return "", err
	}
	return out.URL, nil
}

// DeleteConnection removes a connection and every ad record created through
// it.
func (s *AdsService) DeleteConnection(ctx context.Context, id, workspaceID string) error {
	return s.client.Do(ctx, "DELETE", "/ads/connections/"+url.PathEscape(id), nil, workspaceQuery(workspaceID), nil)
}

// BoostPostRequest is the body of Boost. The boost starts paused unless
// Paused is Bool(false).
type BoostPostRequest struct {
	WorkspaceID string `json:"workspaceId"`
	// ConnectionID is a Meta Ads connection in the workspace.
	ConnectionID string `json:"connectionId"`
	// AdAccountID is the Meta ad account, `act_…`.
	AdAccountID string `json:"adAccountId"`
	// PostID is the FoPost post to promote; AccountID is the delivery to
	// build the boost from.
	PostID    string `json:"postId"`
	AccountID string `json:"accountId"`
	Name      string `json:"name"`
	// Goal is one of the AdGoal constants.
	Goal      string      `json:"goal"`
	Budget    AdBudget    `json:"budget"`
	Targeting AdTargeting `json:"targeting"`
	Paused    *bool       `json:"paused,omitempty"`
}

// Boost promotes a published post. Needs the `publish` scope as well as
// `ads`. The boost starts paused unless Paused is Bool(false).
func (s *AdsService) Boost(ctx context.Context, body *BoostPostRequest) (*Ad, error) {
	out := &Ad{}
	if err := s.client.json(ctx, "POST", "/ads/boost", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateAdRequest is the body of Create. The ad starts paused unless Paused
// is Bool(false).
type CreateAdRequest struct {
	WorkspaceID string `json:"workspaceId"`
	// ConnectionID is a Meta Ads connection in the workspace.
	ConnectionID string `json:"connectionId"`
	// AdAccountID is the Meta ad account, `act_…`.
	AdAccountID string `json:"adAccountId"`
	// PageID is the Facebook Page the ad is published from.
	PageID string `json:"pageId"`
	Name   string `json:"name"`
	// Goal is one of the AdGoal constants.
	Goal      string      `json:"goal"`
	Budget    AdBudget    `json:"budget"`
	Targeting AdTargeting `json:"targeting"`
	// Text is up to 125 characters.
	Text string `json:"text"`
	// Headline is up to 40 characters.
	Headline       string `json:"headline,omitempty"`
	DestinationURL string `json:"destinationUrl,omitempty"`
	// MediaURL is a media library asset url.
	MediaURL string `json:"mediaUrl,omitempty"`
	// URLTags is a query string appended to every link in the ad, e.g.
	// `utm_source=meta&utm_medium=paid`.
	URLTags string `json:"urlTags,omitempty"`
	Paused  *bool  `json:"paused,omitempty"`
	// MessagingDestination is required by the "messages" goal: one of the
	// Messaging constants.
	MessagingDestination string `json:"messagingDestination,omitempty"`
	// PhoneNumber is required by the "calls" goal, in E.164, e.g. "+14155550123".
	PhoneNumber string `json:"phoneNumber,omitempty"`
	// ProductSetID is required by the "sales" goal: makes this a catalog ad over
	// that product set.
	ProductSetID string `json:"productSetId,omitempty"`
}

// Create makes an ad from scratch. Needs the `publish` scope as well as `ads`.
// The ad starts paused unless Paused is Bool(false).
func (s *AdsService) Create(ctx context.Context, body *CreateAdRequest) (*Ad, error) {
	out := &Ad{}
	if err := s.client.json(ctx, "POST", "/ads", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Refresh reads the delivery status and lifetime insights from Meta.
func (s *AdsService) Refresh(ctx context.Context, id, workspaceID string) (*Ad, error) {
	out := &Ad{}
	if err := s.client.json(ctx, "POST", "/ads/"+url.PathEscape(id)+"/refresh", nil, workspaceQuery(workspaceID), out); err != nil {
		return nil, err
	}
	return out, nil
}

// SetStatus resumes or pauses an ad. Status is AdStatusActive or
// AdStatusPaused. Needs the `publish` scope as well as `ads`.
func (s *AdsService) SetStatus(ctx context.Context, id, workspaceID, status string) (*Ad, error) {
	body := map[string]string{"status": status}
	out := &Ad{}
	if err := s.client.json(ctx, "PATCH", "/ads/"+url.PathEscape(id), body, workspaceQuery(workspaceID), out); err != nil {
		return nil, err
	}
	return out, nil
}

// Delete ends delivery and deletes the ad on Meta. Needs the `publish` scope
// as well as `ads`.
func (s *AdsService) Delete(ctx context.Context, id, workspaceID string) error {
	return s.client.Do(ctx, "DELETE", "/ads/"+url.PathEscape(id), nil, workspaceQuery(workspaceID), nil)
}

// ListAudiencesParams names the ad account Audiences reads.
type ListAudiencesParams struct {
	WorkspaceID  string
	ConnectionID string
	// AdAccountID is `act_…`.
	AdAccountID string
}

// Audiences returns an ad account's audiences and pixels.
func (s *AdsService) Audiences(ctx context.Context, params *ListAudiencesParams) (*AudiencesResult, error) {
	q := newQuery()
	if params != nil {
		q.str("workspace_id", params.WorkspaceID)
		q.str("connection_id", params.ConnectionID)
		q.str("ad_account_id", params.AdAccountID)
	}
	out := &AudiencesResult{}
	if err := s.client.json(ctx, "GET", "/ads/audiences", nil, q.values(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// AudienceSpec describes the audience to build. Subtype picks which other
// fields apply: Emails for AudienceCustom; OriginAudienceID, Country and Ratio
// for AudienceLookalike; PixelID, RetentionDays and URLContains for
// AudienceWebsite.
type AudienceSpec struct {
	Subtype string `json:"subtype"`
	// Emails are a customer list, hashed before they leave the API.
	Emails           []string `json:"emails,omitempty"`
	OriginAudienceID string   `json:"originAudienceId,omitempty"`
	// Country is an ISO 3166-1 alpha-2 code.
	Country string `json:"country,omitempty"`
	// Ratio is 0.01 to 0.2, defaulting to 0.01.
	Ratio   float64 `json:"ratio,omitempty"`
	PixelID string  `json:"pixelId,omitempty"`
	// RetentionDays is 1 to 180, defaulting to 30.
	RetentionDays int    `json:"retentionDays,omitempty"`
	URLContains   string `json:"urlContains,omitempty"`
}

// CreateAudienceRequest is the body of CreateAudience.
type CreateAudienceRequest struct {
	WorkspaceID  string `json:"workspaceId"`
	ConnectionID string `json:"connectionId"`
	// AdAccountID is `act_…`.
	AdAccountID string       `json:"adAccountId"`
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Spec        AudienceSpec `json:"spec"`
}

// CreatedAudience is a new audience. Added counts the emails Meta accepted.
type CreatedAudience struct {
	ID    string `json:"id"`
	Added int    `json:"added"`
}

// CreateAudience builds a custom, lookalike or website audience.
func (s *AdsService) CreateAudience(ctx context.Context, body *CreateAudienceRequest) (*CreatedAudience, error) {
	out := &CreatedAudience{}
	if err := s.client.json(ctx, "POST", "/ads/audiences", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// SearchTargetingParams narrows SearchTargeting.
type SearchTargetingParams struct {
	WorkspaceID  string
	ConnectionID string
	// Type is one of the Targeting constants.
	Type string
	Q    string
}

// SearchTargeting finds locations, interests, behaviours and income brackets
// as Meta names them.
func (s *AdsService) SearchTargeting(ctx context.Context, params *SearchTargetingParams) ([]TargetingOption, error) {
	q := newQuery()
	if params != nil {
		q.str("workspace_id", params.WorkspaceID)
		q.str("connection_id", params.ConnectionID)
		q.str("type", params.Type)
		q.str("q", params.Q)
	}
	var out []TargetingOption
	if err := s.client.json(ctx, "GET", "/ads/targeting/search", nil, q.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// LeadForms returns each connection's Page with its lead forms.
func (s *AdsService) LeadForms(ctx context.Context, workspaceID string) ([]LeadFormSource, error) {
	var out []LeadFormSource
	if err := s.client.json(ctx, "GET", "/ads/lead-forms", nil, workspaceQuery(workspaceID), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Lead form questions.
const (
	LeadQuestionEmail    = "EMAIL"
	LeadQuestionFullName = "FULL_NAME"
	LeadQuestionPhone    = "PHONE"
)

// CreateLeadFormRequest is the body of CreateLeadForm.
type CreateLeadFormRequest struct {
	WorkspaceID  string `json:"workspaceId"`
	ConnectionID string `json:"connectionId"`
	// PageID is the Facebook Page id.
	PageID string `json:"pageId"`
	Name   string `json:"name"`
	// Questions are one to three of the LeadQuestion constants.
	Questions        []string `json:"questions"`
	PrivacyPolicyURL string   `json:"privacyPolicyUrl"`
	ThankYouMessage  string   `json:"thankYouMessage"`
	FollowUpURL      string   `json:"followUpUrl,omitempty"`
}

// CreateLeadForm creates a lead form on a Page and returns its id.
func (s *AdsService) CreateLeadForm(ctx context.Context, body *CreateLeadFormRequest) (string, error) {
	var out struct {
		ID string `json:"id"`
	}
	if err := s.client.json(ctx, "POST", "/ads/lead-forms", body, nil, &out); err != nil {
		return "", err
	}
	return out.ID, nil
}

// ListLeadsParams names the Page a form's leads are read from. After is the
// NextCursor of the previous page.
type ListLeadsParams struct {
	WorkspaceID  string
	ConnectionID string
	PageID       string
	After        string
}

// Leads returns one page of a form's leads.
func (s *AdsService) Leads(ctx context.Context, formID string, params *ListLeadsParams) (*LeadsPage, error) {
	q := newQuery()
	if params != nil {
		q.str("workspace_id", params.WorkspaceID)
		q.str("connection_id", params.ConnectionID)
		q.str("page_id", params.PageID)
		q.str("after", params.After)
	}
	out := &LeadsPage{}
	if err := s.client.json(ctx, "GET", "/ads/lead-forms/"+url.PathEscape(formID)+"/leads", nil, q.values(), out); err != nil {
		return nil, err
	}
	return out, nil
}

func workspaceQuery(workspaceID string) url.Values {
	q := newQuery()
	q.str("workspace_id", workspaceID)
	return q.values()
}

// Object levels for SetStatuses.
const (
	AdLevelCampaign = "campaign"
	AdLevelAdSet    = "ad_set"
	AdLevelAd       = "ad"
)

// Insights breakdowns.
const (
	InsightsByAge       = "age"
	InsightsByGender    = "gender"
	InsightsByPlacement = "placement"
	InsightsByCountry   = "country"
)

// Creative formats for CreateCreative.
const (
	CreativeImage    = "image"
	CreativeVideo    = "video"
	CreativeCarousel = "carousel"
)

// AdObjectParams names the connection a Meta object is read or changed
// through. WorkspaceID is required on writes.
type AdObjectParams struct {
	WorkspaceID  string
	ConnectionID string
}

func (p *AdObjectParams) query() url.Values {
	q := newQuery()
	if p != nil {
		q.str("workspace_id", p.WorkspaceID)
		q.str("connection_id", p.ConnectionID)
	}
	return q.values()
}

// Campaign is a Meta campaign, read live and never stored. Status is Meta's
// `ACTIVE`, `PAUSED`, `DELETED` or `ARCHIVED`.
type Campaign struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Status          string `json:"status"`
	EffectiveStatus string `json:"effectiveStatus"`
	Objective       string `json:"objective"`
	// BudgetMinor is nil when the budget lives on the ad sets.
	BudgetMinor *int   `json:"budgetMinor"`
	BudgetType  string `json:"budgetType"`
	CreatedAt   Time   `json:"createdAt"`
}

// AdSet is a Meta ad set, read live and never stored.
type AdSet struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	CampaignID       string `json:"campaignId"`
	Status           string `json:"status"`
	EffectiveStatus  string `json:"effectiveStatus"`
	BudgetMinor      *int   `json:"budgetMinor"`
	BudgetType       string `json:"budgetType"`
	EndAt            Time   `json:"endAt"`
	OptimizationGoal string `json:"optimizationGoal"`
	CreatedAt        Time   `json:"createdAt"`
}

// NetworkAd is an ad inside a Meta ad set, read live and never stored.
type NetworkAd struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	CampaignID      string `json:"campaignId"`
	AdSetID         string `json:"adSetId"`
	CreativeID      string `json:"creativeId"`
	Status          string `json:"status"`
	EffectiveStatus string `json:"effectiveStatus"`
	CreatedAt       Time   `json:"createdAt"`
}

// TreeAdSet is an ad set with its ads.
type TreeAdSet struct {
	AdSet
	Ads []NetworkAd `json:"ads"`
}

// TreeCampaign is a campaign with its ad sets.
type TreeCampaign struct {
	Campaign
	AdSets []TreeAdSet `json:"adSets"`
}

// AdAccountTree is every campaign on an ad account with its ad sets and ads.
type AdAccountTree struct {
	AdAccountID string         `json:"adAccountId"`
	Currency    string         `json:"currency"`
	WorkspaceID string         `json:"workspaceId"`
	Campaigns   []TreeCampaign `json:"campaigns"`
}

// Tree returns an ad account's campaigns, ad sets and ads, whoever made them.
func (s *AdsService) Tree(ctx context.Context, adAccountID string, params *AdObjectParams) (*AdAccountTree, error) {
	out := &AdAccountTree{}
	if err := s.client.json(ctx, "GET", "/ads/accounts/"+url.PathEscape(adAccountID)+"/tree", nil, params.query(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateCampaignRequest is the body of CreateCampaign. The campaign starts
// paused unless Paused is Bool(false).
type CreateCampaignRequest struct {
	WorkspaceID  string `json:"workspaceId"`
	ConnectionID string `json:"connectionId"`
	// AdAccountID is `act_…`.
	AdAccountID string `json:"adAccountId"`
	Name        string `json:"name"`
	// Goal is one of the AdGoal constants.
	Goal   string `json:"goal"`
	Paused *bool  `json:"paused,omitempty"`
}

// UpdateCampaignRequest is the body of UpdateCampaign. Status is
// AdStatusActive or AdStatusPaused.
type UpdateCampaignRequest struct {
	Name   string `json:"name,omitempty"`
	Status string `json:"status,omitempty"`
}

// CreateCampaign creates a campaign. Needs `publish` as well as `ads`.
func (s *AdsService) CreateCampaign(ctx context.Context, body *CreateCampaignRequest) (*Campaign, error) {
	out := &Campaign{}
	if err := s.client.json(ctx, "POST", "/ads/campaigns", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Campaign returns one campaign.
func (s *AdsService) Campaign(ctx context.Context, id string, params *AdObjectParams) (*Campaign, error) {
	out := &Campaign{}
	if err := s.client.json(ctx, "GET", "/ads/campaigns/"+url.PathEscape(id), nil, params.query(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateCampaign renames, pauses or resumes a campaign. Needs `publish` as
// well as `ads`.
func (s *AdsService) UpdateCampaign(ctx context.Context, id string, params *AdObjectParams, body *UpdateCampaignRequest) (*Campaign, error) {
	out := &Campaign{}
	if err := s.client.json(ctx, "PATCH", "/ads/campaigns/"+url.PathEscape(id), body, params.query(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteCampaign deletes a campaign on Meta. Needs `publish` as well as `ads`.
func (s *AdsService) DeleteCampaign(ctx context.Context, id string, params *AdObjectParams) error {
	return s.client.Do(ctx, "DELETE", "/ads/campaigns/"+url.PathEscape(id), nil, params.query(), nil)
}

// DuplicateCampaign copies a campaign and returns the copy's Meta id. The
// copy starts paused unless paused is Bool(false). Needs `publish` as well as
// `ads`.
func (s *AdsService) DuplicateCampaign(ctx context.Context, id string, params *AdObjectParams, paused *bool) (string, error) {
	return s.duplicate(ctx, "/ads/campaigns/"+url.PathEscape(id)+"/duplicate", params, paused)
}

// CreateAdSetRequest is the body of CreateAdSet. The ad set starts paused
// unless Paused is Bool(false).
type CreateAdSetRequest struct {
	WorkspaceID  string `json:"workspaceId"`
	ConnectionID string `json:"connectionId"`
	CampaignID   string `json:"campaignId"`
	// PageID is the Page the ads in this set run as.
	PageID string `json:"pageId"`
	Name   string `json:"name"`
	// Goal is one of the AdGoal constants.
	Goal      string      `json:"goal"`
	Budget    AdBudget    `json:"budget"`
	Targeting AdTargeting `json:"targeting"`
	Paused    *bool       `json:"paused,omitempty"`
	// MessagingDestination is required by the "messages" goal: one of the
	// Messaging constants.
	MessagingDestination string `json:"messagingDestination,omitempty"`
	// PhoneNumber is required by the "calls" goal, in E.164, e.g. "+14155550123".
	PhoneNumber string `json:"phoneNumber,omitempty"`
	// ProductSetID is required by the "sales" goal: the product set the catalog
	// ad runs from.
	ProductSetID string `json:"productSetId,omitempty"`
}

// UpdateAdSetRequest is the body of UpdateAdSet. BudgetMinor keeps the budget
// type set at creation.
type UpdateAdSetRequest struct {
	Name        string       `json:"name,omitempty"`
	Status      string       `json:"status,omitempty"`
	BudgetMinor *int         `json:"budgetMinor,omitempty"`
	EndAt       *Time        `json:"endAt,omitempty"`
	Targeting   *AdTargeting `json:"targeting,omitempty"`
}

// CreateAdSet creates an ad set in a campaign. Needs `publish` as well as
// `ads`.
func (s *AdsService) CreateAdSet(ctx context.Context, body *CreateAdSetRequest) (*AdSet, error) {
	out := &AdSet{}
	if err := s.client.json(ctx, "POST", "/ads/ad-sets", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// AdSet returns one ad set.
func (s *AdsService) AdSet(ctx context.Context, id string, params *AdObjectParams) (*AdSet, error) {
	out := &AdSet{}
	if err := s.client.json(ctx, "GET", "/ads/ad-sets/"+url.PathEscape(id), nil, params.query(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateAdSet changes an ad set's name, status, budget, end or targeting.
// Needs `publish` as well as `ads`.
func (s *AdsService) UpdateAdSet(ctx context.Context, id string, params *AdObjectParams, body *UpdateAdSetRequest) (*AdSet, error) {
	out := &AdSet{}
	if err := s.client.json(ctx, "PATCH", "/ads/ad-sets/"+url.PathEscape(id), body, params.query(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteAdSet deletes an ad set on Meta. Needs `publish` as well as `ads`.
func (s *AdsService) DeleteAdSet(ctx context.Context, id string, params *AdObjectParams) error {
	return s.client.Do(ctx, "DELETE", "/ads/ad-sets/"+url.PathEscape(id), nil, params.query(), nil)
}

// DuplicateAdSet copies an ad set and returns the copy's Meta id. Needs
// `publish` as well as `ads`.
func (s *AdsService) DuplicateAdSet(ctx context.Context, id string, params *AdObjectParams, paused *bool) (string, error) {
	return s.duplicate(ctx, "/ads/ad-sets/"+url.PathEscape(id)+"/duplicate", params, paused)
}

// CreateNetworkAdRequest is the body of CreateNetworkAd. The ad starts paused
// unless Paused is Bool(false).
type CreateNetworkAdRequest struct {
	WorkspaceID  string `json:"workspaceId"`
	ConnectionID string `json:"connectionId"`
	AdSetID      string `json:"adSetId"`
	// CreativeID is from CreateCreative or Creatives.
	CreativeID string `json:"creativeId"`
	Name       string `json:"name"`
	Paused     *bool  `json:"paused,omitempty"`
}

// UpdateNetworkAdRequest is the body of UpdateNetworkAd.
type UpdateNetworkAdRequest struct {
	Name       string `json:"name,omitempty"`
	Status     string `json:"status,omitempty"`
	CreativeID string `json:"creativeId,omitempty"`
}

// CreateNetworkAd creates an ad inside an ad set. Unlike Create it builds no
// campaign. Needs `publish` as well as `ads`.
func (s *AdsService) CreateNetworkAd(ctx context.Context, body *CreateNetworkAdRequest) (*NetworkAd, error) {
	out := &NetworkAd{}
	if err := s.client.json(ctx, "POST", "/ads/ads", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// NetworkAd returns one ad by its Meta id.
func (s *AdsService) NetworkAd(ctx context.Context, id string, params *AdObjectParams) (*NetworkAd, error) {
	out := &NetworkAd{}
	if err := s.client.json(ctx, "GET", "/ads/ads/"+url.PathEscape(id), nil, params.query(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateNetworkAd renames, pauses, resumes or swaps the creative of an ad.
// Needs `publish` as well as `ads`.
func (s *AdsService) UpdateNetworkAd(ctx context.Context, id string, params *AdObjectParams, body *UpdateNetworkAdRequest) (*NetworkAd, error) {
	out := &NetworkAd{}
	if err := s.client.json(ctx, "PATCH", "/ads/ads/"+url.PathEscape(id), body, params.query(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteNetworkAd deletes an ad on Meta. Needs `publish` as well as `ads`.
func (s *AdsService) DeleteNetworkAd(ctx context.Context, id string, params *AdObjectParams) error {
	return s.client.Do(ctx, "DELETE", "/ads/ads/"+url.PathEscape(id), nil, params.query(), nil)
}

// DuplicateNetworkAd copies an ad and returns the copy's Meta id. Needs
// `publish` as well as `ads`.
func (s *AdsService) DuplicateNetworkAd(ctx context.Context, id string, params *AdObjectParams, paused *bool) (string, error) {
	return s.duplicate(ctx, "/ads/ads/"+url.PathEscape(id)+"/duplicate", params, paused)
}

func (s *AdsService) duplicate(ctx context.Context, path string, params *AdObjectParams, paused *bool) (string, error) {
	var body any
	if paused != nil {
		body = map[string]bool{"paused": *paused}
	}
	var out struct {
		ID string `json:"id"`
	}
	if err := s.client.json(ctx, "POST", path, body, params.query(), &out); err != nil {
		return "", err
	}
	return out.ID, nil
}

// AdObjectRef names one campaign, ad set or ad for SetStatuses. Level is one
// of the AdLevel constants.
type AdObjectRef struct {
	ID    string `json:"id"`
	Level string `json:"level"`
}

// SetStatusesRequest is the body of SetStatuses: up to 50 objects.
type SetStatusesRequest struct {
	WorkspaceID  string `json:"workspaceId"`
	ConnectionID string `json:"connectionId"`
	// Status is AdStatusActive or AdStatusPaused.
	Status  string        `json:"status"`
	Objects []AdObjectRef `json:"objects"`
}

// AdStatusResult is the outcome for one object of SetStatuses.
type AdStatusResult struct {
	ID    string `json:"id"`
	Level string `json:"level"`
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

// SetStatuses pauses or resumes many campaigns, ad sets and ads; each
// succeeds or fails on its own. Needs `publish` as well as `ads`.
func (s *AdsService) SetStatuses(ctx context.Context, body *SetStatusesRequest) ([]AdStatusResult, error) {
	var out []AdStatusResult
	if err := s.client.json(ctx, "POST", "/ads/status", body, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Creative is one ad creative on an ad account. Format is "image", "video",
// "carousel", "post" or "other".
type Creative struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Format       string `json:"format"`
	Status       string `json:"status"`
	Title        string `json:"title"`
	Body         string `json:"body"`
	Link         string `json:"link"`
	ThumbnailURL string `json:"thumbnailUrl"`
	CallToAction string `json:"callToAction"`
	URLTags      string `json:"urlTags"`
}

// CreativesResult is an ad account's creative library.
type CreativesResult struct {
	Creatives   []Creative `json:"creatives"`
	WorkspaceID string     `json:"workspaceId"`
}

// ListCreativesParams names the ad account Creatives reads.
type ListCreativesParams struct {
	WorkspaceID  string
	ConnectionID string
	// AdAccountID is `act_…`.
	AdAccountID string
}

// Creatives returns an ad account's creative library.
func (s *AdsService) Creatives(ctx context.Context, params *ListCreativesParams) (*CreativesResult, error) {
	q := newQuery()
	if params != nil {
		q.str("workspace_id", params.WorkspaceID)
		q.str("connection_id", params.ConnectionID)
		q.str("ad_account_id", params.AdAccountID)
	}
	out := &CreativesResult{}
	if err := s.client.json(ctx, "GET", "/ads/creatives", nil, q.values(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// CarouselCard is one card of a carousel creative. MediaURL is a library
// image.
type CarouselCard struct {
	MediaURL       string `json:"mediaUrl"`
	DestinationURL string `json:"destinationUrl,omitempty"`
	Headline       string `json:"headline,omitempty"`
	Description    string `json:"description,omitempty"`
}

// CreateCreativeRequest is the body of CreateCreative. Format is one of the
// Creative constants: MediaURL is required for a video, Cards (2 to 10) for a
// carousel.
type CreateCreativeRequest struct {
	WorkspaceID  string `json:"workspaceId"`
	ConnectionID string `json:"connectionId"`
	// AdAccountID is `act_…`.
	AdAccountID string `json:"adAccountId"`
	PageID      string `json:"pageId"`
	Name        string `json:"name"`
	Format      string `json:"format"`
	// Text is the primary text.
	Text           string `json:"text"`
	Headline       string `json:"headline,omitempty"`
	DestinationURL string `json:"destinationUrl,omitempty"`
	// CallToAction is Meta's button type, e.g. "SHOP_NOW"; "LEARN_MORE" when
	// empty.
	CallToAction string `json:"callToAction,omitempty"`
	// URLTags is a query string appended to every link in the ad.
	URLTags string `json:"urlTags,omitempty"`
	// MediaURL is a media library asset url: the image, or the video.
	MediaURL string `json:"mediaUrl,omitempty"`
	// ThumbnailMediaURL is a video's poster frame, as a library image.
	ThumbnailMediaURL string         `json:"thumbnailMediaUrl,omitempty"`
	Cards             []CarouselCard `json:"cards,omitempty"`
	// ProductSetID is required for the "catalog" format: the network fills the
	// cards from this product set.
	ProductSetID string `json:"productSetId,omitempty"`
	// Description is the per-product line under the headline, "catalog" only.
	Description string `json:"description,omitempty"`
	// CreatorPostID is required for the "partnership" format: the creator's
	// media id, or their Page post as `{page}_{post}`.
	CreatorPostID string `json:"creatorPostId,omitempty"`
	// CreatorInstagramUserID is the creator's Instagram account, "partnership"
	// only.
	CreatorInstagramUserID string `json:"creatorInstagramUserId,omitempty"`
}

// CreateCreative adds an image, video or carousel creative to an ad account.
func (s *AdsService) CreateCreative(ctx context.Context, body *CreateCreativeRequest) (*Creative, error) {
	out := &Creative{}
	if err := s.client.json(ctx, "POST", "/ads/creatives", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Creative returns one creative.
func (s *AdsService) Creative(ctx context.Context, id string, params *AdObjectParams) (*Creative, error) {
	out := &Creative{}
	if err := s.client.json(ctx, "GET", "/ads/creatives/"+url.PathEscape(id), nil, params.query(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteCreative deletes a creative on Meta.
func (s *AdsService) DeleteCreative(ctx context.Context, id string, params *AdObjectParams) error {
	return s.client.Do(ctx, "DELETE", "/ads/creatives/"+url.PathEscape(id), nil, params.query(), nil)
}

// UpdateAudienceRequest is the body of UpdateAudience.
type UpdateAudienceRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// Audience returns one audience.
func (s *AdsService) Audience(ctx context.Context, id string, params *AdObjectParams) (*Audience, error) {
	out := &Audience{}
	if err := s.client.json(ctx, "GET", "/ads/audiences/"+url.PathEscape(id), nil, params.query(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateAudience renames an audience or changes its description.
func (s *AdsService) UpdateAudience(ctx context.Context, id string, params *AdObjectParams, body *UpdateAudienceRequest) (*Audience, error) {
	out := &Audience{}
	if err := s.client.json(ctx, "PATCH", "/ads/audiences/"+url.PathEscape(id), body, params.query(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteAudience deletes an audience on Meta.
func (s *AdsService) DeleteAudience(ctx context.Context, id string, params *AdObjectParams) error {
	return s.client.Do(ctx, "DELETE", "/ads/audiences/"+url.PathEscape(id), nil, params.query(), nil)
}

// AddAudienceUsers adds emails to a custom audience, hashed before they leave
// the API, and returns how many were sent to Meta.
func (s *AdsService) AddAudienceUsers(ctx context.Context, id string, params *AdObjectParams, emails []string) (int, error) {
	body := map[string][]string{"emails": emails}
	var out struct {
		Added int `json:"added"`
	}
	if err := s.client.json(ctx, "POST", "/ads/audiences/"+url.PathEscape(id)+"/users", body, params.query(), &out); err != nil {
		return 0, err
	}
	return out.Added, nil
}

// ReachEstimateRequest is the body of EstimateReach.
type ReachEstimateRequest struct {
	WorkspaceID  string `json:"workspaceId"`
	ConnectionID string `json:"connectionId"`
	// AdAccountID is `act_…`.
	AdAccountID string      `json:"adAccountId"`
	PageID      string      `json:"pageId"`
	Targeting   AdTargeting `json:"targeting"`
}

// ReachEstimate is Meta's audience size range for a targeting. Ready is false
// while Meta is still computing it.
type ReachEstimate struct {
	Lower *int `json:"lower"`
	Upper *int `json:"upper"`
	Ready bool `json:"ready"`
}

// EstimateReach returns how many people a targeting reaches.
func (s *AdsService) EstimateReach(ctx context.Context, body *ReachEstimateRequest) (*ReachEstimate, error) {
	out := &ReachEstimate{}
	if err := s.client.json(ctx, "POST", "/ads/reach-estimate", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// InsightsMetrics are delivery numbers. SpendMinor is in the account
// currency's minor units; CTR is clicks per impression, as a percentage.
type InsightsMetrics struct {
	Impressions int     `json:"impressions"`
	Reach       int     `json:"reach"`
	Clicks      int     `json:"clicks"`
	SpendMinor  int     `json:"spendMinor"`
	CTR         float64 `json:"ctr"`
	Leads       int     `json:"leads"`
}

// InsightsBreakdownRow is one value of the requested breakdown.
type InsightsBreakdownRow struct {
	Key     string          `json:"key"`
	Metrics InsightsMetrics `json:"metrics"`
}

// InsightsDay is one day of the timeline.
type InsightsDay struct {
	Date    string          `json:"date"`
	Metrics InsightsMetrics `json:"metrics"`
}

// InsightsReport is an object's delivery over a date range. Totals is nil when
// Meta has nothing for the range.
type InsightsReport struct {
	ObjectID    string                 `json:"objectId"`
	Currency    string                 `json:"currency"`
	Since       string                 `json:"since"`
	Until       string                 `json:"until"`
	BreakdownBy string                 `json:"breakdownBy"`
	Totals      *InsightsMetrics       `json:"totals"`
	Breakdown   []InsightsBreakdownRow `json:"breakdown"`
	Timeline    []InsightsDay          `json:"timeline"`
}

// InsightsParams narrows Insights. ObjectID is an ad account (`act_…`),
// campaign, ad set or ad; Since and Until are inclusive `YYYY-MM-DD` dates;
// Breakdown is one of the InsightsBy constants; Daily adds the timeline.
type InsightsParams struct {
	WorkspaceID  string
	ConnectionID string
	ObjectID     string
	Since        string
	Until        string
	Breakdown    string
	Daily        *bool
}

// Insights returns delivery totals for any Meta object over a date range.
func (s *AdsService) Insights(ctx context.Context, params *InsightsParams) (*InsightsReport, error) {
	q := newQuery()
	if params != nil {
		q.str("workspace_id", params.WorkspaceID)
		q.str("connection_id", params.ConnectionID)
		q.str("object_id", params.ObjectID)
		q.str("since", params.Since)
		q.str("until", params.Until)
		q.str("breakdown", params.Breakdown)
		q.boolPtr("daily", params.Daily)
	}
	out := &InsightsReport{}
	if err := s.client.json(ctx, "GET", "/ads/insights", nil, q.values(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// AdInsightsParams narrows AdInsights. Since and Until are inclusive
// `YYYY-MM-DD` dates.
type AdInsightsParams struct {
	WorkspaceID string
	Since       string
	Until       string
	Breakdown   string
	Daily       *bool
}

// AdInsights returns delivery over a date range for a boost or ad created
// through FoPost, by its FoPost id.
func (s *AdsService) AdInsights(ctx context.Context, id string, params *AdInsightsParams) (*InsightsReport, error) {
	q := newQuery()
	if params != nil {
		q.str("workspace_id", params.WorkspaceID)
		q.str("since", params.Since)
		q.str("until", params.Until)
		q.str("breakdown", params.Breakdown)
		q.boolPtr("daily", params.Daily)
	}
	out := &InsightsReport{}
	if err := s.client.json(ctx, "GET", "/ads/"+url.PathEscape(id)+"/insights", nil, q.values(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// LeadFormDetail is one lead form with its settings.
type LeadFormDetail struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Status           string   `json:"status"`
	LeadsCount       int      `json:"leadsCount"`
	CreatedAt        Time     `json:"createdAt"`
	Questions        []string `json:"questions"`
	PageID           string   `json:"pageId"`
	PrivacyPolicyURL string   `json:"privacyPolicyUrl"`
	Locale           string   `json:"locale"`
}

// LeadFormParams names the Page a lead form lives on.
type LeadFormParams struct {
	WorkspaceID  string
	ConnectionID string
	PageID       string
}

// LeadForm returns one lead form.
func (s *AdsService) LeadForm(ctx context.Context, formID string, params *LeadFormParams) (*LeadFormDetail, error) {
	q := newQuery()
	if params != nil {
		q.str("workspace_id", params.WorkspaceID)
		q.str("connection_id", params.ConnectionID)
		q.str("page_id", params.PageID)
	}
	out := &LeadFormDetail{}
	if err := s.client.json(ctx, "GET", "/ads/lead-forms/"+url.PathEscape(formID), nil, q.values(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// ArchiveLeadForm stops a lead form collecting; its leads stay readable.
func (s *AdsService) ArchiveLeadForm(ctx context.Context, formID string, params *LeadFormParams) (*LeadFormDetail, error) {
	body := map[string]string{}
	if params != nil {
		body["workspaceId"] = params.WorkspaceID
		body["connectionId"] = params.ConnectionID
		body["pageId"] = params.PageID
	}
	out := &LeadFormDetail{}
	if err := s.client.json(ctx, "POST", "/ads/lead-forms/"+url.PathEscape(formID)+"/archive", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// FeedLead is one lead collected from a subscribed Page.
type FeedLead struct {
	ID string `json:"id"`
	// LeadID is Meta's lead id.
	LeadID       string      `json:"leadId"`
	ConnectionID string      `json:"connectionId"`
	PageID       string      `json:"pageId"`
	FormID       string      `json:"formId"`
	AdID         string      `json:"adId"`
	AdName       string      `json:"adName"`
	CampaignName string      `json:"campaignName"`
	Platform     string      `json:"platform"`
	IsOrganic    bool        `json:"isOrganic"`
	Fields       []LeadField `json:"fields"`
	SubmittedAt  Time        `json:"submittedAt"`
	WorkspaceID  string      `json:"workspaceId"`
}

// LeadsFeedPage is one page of the leads feed. Pass NextCursor back as Cursor
// for the next; it is empty on the last page.
type LeadsFeedPage struct {
	Leads      []FeedLead `json:"leads"`
	NextCursor string     `json:"nextCursor"`
}

// LeadsFeedParams narrows LeadsFeed. Limit is 1 to 100.
type LeadsFeedParams struct {
	WorkspaceID string
	FormID      string
	PageID      string
	Cursor      string
	Limit       int
}

// LeadsFeed returns the leads collected from subscribed Pages, newest first.
func (s *AdsService) LeadsFeed(ctx context.Context, params *LeadsFeedParams) (*LeadsFeedPage, error) {
	q := newQuery()
	if params != nil {
		q.str("workspace_id", params.WorkspaceID)
		q.str("form_id", params.FormID)
		q.str("page_id", params.PageID)
		q.str("cursor", params.Cursor)
		q.num("limit", params.Limit)
	}
	out := &LeadsFeedPage{}
	if err := s.client.json(ctx, "GET", "/ads/leads", nil, q.values(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// LeadPage is a Page whose leads are collected into the feed.
type LeadPage struct {
	ConnectionID string `json:"connectionId"`
	PageID       string `json:"pageId"`
	PageName     string `json:"pageName"`
	CreatedAt    Time   `json:"createdAt"`
	WorkspaceID  string `json:"workspaceId"`
}

// LeadPages returns the Pages whose leads are collected.
func (s *AdsService) LeadPages(ctx context.Context, workspaceID string) ([]LeadPage, error) {
	var out []LeadPage
	if err := s.client.json(ctx, "GET", "/ads/lead-pages", nil, workspaceQuery(workspaceID), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SubscribeLeadPageRequest is the body of SubscribeLeadPage.
type SubscribeLeadPageRequest struct {
	WorkspaceID  string `json:"workspaceId"`
	ConnectionID string `json:"connectionId"`
	PageID       string `json:"pageId"`
}

// SubscribedLeadPage is a newly subscribed Page. Backfilled counts the recent
// leads copied into the feed.
type SubscribedLeadPage struct {
	PageID     string `json:"pageId"`
	Backfilled int    `json:"backfilled"`
}

// SubscribeLeadPage starts collecting a Page's leads and backfills its most
// recent ones.
func (s *AdsService) SubscribeLeadPage(ctx context.Context, body *SubscribeLeadPageRequest) (*SubscribedLeadPage, error) {
	out := &SubscribedLeadPage{}
	if err := s.client.json(ctx, "POST", "/ads/lead-pages", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// UnsubscribeLeadPage stops collecting a Page's leads.
func (s *AdsService) UnsubscribeLeadPage(ctx context.Context, pageID string, params *AdObjectParams) error {
	return s.client.Do(ctx, "DELETE", "/ads/lead-pages/"+url.PathEscape(pageID), nil, params.query(), nil)
}
