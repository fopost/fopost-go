package fopost

import (
	"context"
	"net/url"
)

// AdsService covers Meta ads: boosts and ads created through FoPost, the ad
// connections they run on, audiences, targeting search and lead forms. Every
// method needs the `ads` scope; Boost, Create, SetStatus and Delete spend
// money and also need `publish`.
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
	Paused   *bool  `json:"paused,omitempty"`
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
