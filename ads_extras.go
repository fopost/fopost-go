package fopost

import (
	"context"
	"net/url"
)

// Goals a network may run beyond the four every network runs. A goal the
// deployment is not set up for is absent from Goals and refused if sent.
const (
	AdGoalMessages = "messages"
	AdGoalCalls    = "calls"
	AdGoalWhatsApp = "whatsapp"
	AdGoalSales    = "sales"
)

// Where a messaging ad opens a conversation.
const (
	MessagingMessenger       = "messenger"
	MessagingInstagramDirect = "instagram_direct"
	MessagingWhatsApp        = "whatsapp"
)

// High-demand budget value types.
const (
	BudgetValueAbsolute   = "ABSOLUTE"
	BudgetValueMultiplier = "MULTIPLIER"
)

// AdAccountParams names the ad account an account-scoped read runs against.
type AdAccountParams struct {
	WorkspaceID  string
	ConnectionID string
	// AdAccountID is `act_…`.
	AdAccountID string
}

func (p *AdAccountParams) query() url.Values {
	q := newQuery()
	if p != nil {
		q.str("workspace_id", p.WorkspaceID)
		q.str("connection_id", p.ConnectionID)
		q.str("ad_account_id", p.AdAccountID)
	}
	return q.values()
}

// Goals returns the goals this connection's network can run right now. Ask
// rather than assume: a goal the deployment is not set up for is absent here
// and is refused if you send it anyway.
func (s *AdsService) Goals(ctx context.Context, params *AdObjectParams) ([]string, error) {
	var out []string
	if err := s.client.json(ctx, "GET", "/ads/goals", nil, params.query(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ─── Reach and frequency ────────────────────────────────────────────

// ReachFrequencyPrediction is a priced flight. Nothing is bought until it is
// reserved.
type ReachFrequencyPrediction struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Status       string `json:"status"`
	Reach        *int   `json:"reach"`
	Impressions  *int   `json:"impressions"`
	FrequencyCap *int   `json:"frequencyCap"`
	// BudgetMinor is account currency, minor units.
	BudgetMinor *int  `json:"budgetMinor"`
	StartAt     *Time `json:"startAt"`
	EndAt       *Time `json:"endAt"`
	// Reserved is true once the prediction holds inventory.
	Reserved bool `json:"reserved"`
}

// ReachFrequencyResult is the predictions on one ad account.
type ReachFrequencyResult struct {
	Predictions []ReachFrequencyPrediction `json:"predictions"`
	WorkspaceID string                     `json:"workspaceId"`
}

// CreateReachFrequencyRequest prices a flight.
type CreateReachFrequencyRequest struct {
	WorkspaceID  string `json:"workspaceId"`
	ConnectionID string `json:"connectionId"`
	// AdAccountID is `act_…`.
	AdAccountID string      `json:"adAccountId"`
	Name        string      `json:"name"`
	Targeting   AdTargeting `json:"targeting"`
	Placements  []string    `json:"placements"`
	BudgetMinor int         `json:"budgetMinor"`
	StartAt     Time        `json:"startAt"`
	EndAt       Time        `json:"endAt"`
	// FrequencyCap is how often one person should see the ad over the flight.
	FrequencyCap int `json:"frequencyCap,omitempty"`
}

// ReachFrequencyActionRequest reserves or cancels a prediction.
type ReachFrequencyActionRequest struct {
	WorkspaceID  string `json:"workspaceId"`
	ConnectionID string `json:"connectionId"`
	// AdAccountID is `act_…`.
	AdAccountID string `json:"adAccountId"`
}

// ReachFrequency returns the predictions on one ad account.
func (s *AdsService) ReachFrequency(ctx context.Context, params *AdAccountParams) (*ReachFrequencyResult, error) {
	out := &ReachFrequencyResult{}
	if err := s.client.json(ctx, "GET", "/ads/reach-frequency", nil, params.query(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateReachFrequency prices a flight. Nothing is bought until you reserve it.
func (s *AdsService) CreateReachFrequency(ctx context.Context, body *CreateReachFrequencyRequest) (*ReachFrequencyPrediction, error) {
	out := &ReachFrequencyPrediction{}
	if err := s.client.json(ctx, "POST", "/ads/reach-frequency", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ReachFrequencyPredictionByID reads one prediction.
func (s *AdsService) ReachFrequencyPredictionByID(ctx context.Context, id string, params *AdAccountParams) (*ReachFrequencyPrediction, error) {
	out := &ReachFrequencyPrediction{}
	if err := s.client.json(ctx, "GET", "/ads/reach-frequency/"+id, nil, params.query(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// ReserveReachFrequency holds the inventory the prediction priced. Needs
// `publish` as well as `ads`.
// Wraps the API's `reserveReachFrequency` operation.
func (s *AdsService) ReserveReachFrequency(ctx context.Context, id string, body *ReachFrequencyActionRequest) (*ReachFrequencyPrediction, error) {
	return s.reachFrequencyAction(ctx, id, "reserve", body)
}

// CancelReachFrequency cancels a reservation. Needs `publish` as well as `ads`.
// Wraps the API's `cancelReachFrequency` operation.
func (s *AdsService) CancelReachFrequency(ctx context.Context, id string, body *ReachFrequencyActionRequest) (*ReachFrequencyPrediction, error) {
	return s.reachFrequencyAction(ctx, id, "cancel", body)
}

func (s *AdsService) reachFrequencyAction(ctx context.Context, id, action string, body *ReachFrequencyActionRequest) (*ReachFrequencyPrediction, error) {
	out := &ReachFrequencyPrediction{}
	if err := s.client.json(ctx, "POST", "/ads/reach-frequency/"+id+"/"+action, body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ─── Partnership ads ────────────────────────────────────────────────

// PartnershipCreator is a creator who allowlisted this advertiser.
type PartnershipCreator struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	Name        string   `json:"name"`
	Status      string   `json:"status"`
	Permissions []string `json:"permissions"`
}

// PartnershipParams names the Page whose allowlist is read.
type PartnershipParams struct {
	WorkspaceID  string
	ConnectionID string
	PageID       string
}

func (p *PartnershipParams) query() url.Values {
	q := newQuery()
	if p != nil {
		q.str("workspace_id", p.WorkspaceID)
		q.str("connection_id", p.ConnectionID)
		q.str("page_id", p.PageID)
	}
	return q.values()
}

// PartnershipRequest asks a creator for permission.
type PartnershipRequest struct {
	WorkspaceID  string `json:"workspaceId"`
	ConnectionID string `json:"connectionId"`
	PageID       string `json:"pageId"`
	// CreatorID is the creator's account id.
	CreatorID string `json:"creatorId"`
}

// PartnershipCreators returns the creators who allowlisted this Page to run
// partnership ads on their posts.
func (s *AdsService) PartnershipCreators(ctx context.Context, params *PartnershipParams) ([]PartnershipCreator, error) {
	var out []PartnershipCreator
	if err := s.client.json(ctx, "GET", "/ads/partnership/creators", nil, params.query(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// RequestPartnership asks a creator for permission and returns the list as it
// now stands.
func (s *AdsService) RequestPartnership(ctx context.Context, body *PartnershipRequest) ([]PartnershipCreator, error) {
	var out []PartnershipCreator
	if err := s.client.json(ctx, "POST", "/ads/partnership/creators", body, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// RevokePartnership revokes a creator's permission.
func (s *AdsService) RevokePartnership(ctx context.Context, creatorID string, params *PartnershipParams) error {
	return s.client.json(ctx, "DELETE", "/ads/partnership/creators/"+creatorID, nil, params.query(), nil)
}

// ─── Ad account settings ────────────────────────────────────────────

// AdActivity is one change recorded on an ad account.
type AdActivity struct {
	ID         string `json:"id"`
	EventType  string `json:"eventType"`
	ActorName  string `json:"actorName"`
	ObjectName string `json:"objectName"`
	ObjectType string `json:"objectType"`
	ExtraData  string `json:"extraData"`
	CreatedAt  *Time  `json:"createdAt"`
}

// AdActivityResult is the change log of one ad account.
type AdActivityResult struct {
	Activity    []AdActivity `json:"activity"`
	WorkspaceID string       `json:"workspaceId"`
}

// AdLabel groups campaigns, ad sets and ads for reporting.
type AdLabel struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt *Time  `json:"createdAt"`
}

// AdStudy is an A/B study splitting traffic across its cells.
type AdStudy struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	StartAt     *Time  `json:"startAt"`
	EndAt       *Time  `json:"endAt"`
}

// IosCampaignLimits is how many iOS 14 campaigns an account may run at once.
type IosCampaignLimits struct {
	Limit *int   `json:"limit"`
	Used  *int   `json:"used"`
	AppID string `json:"appId"`
}

// HighDemandPeriod is a window the network should expect heavier spend over.
type HighDemandPeriod struct {
	ID              string   `json:"id"`
	StartAt         *Time    `json:"startAt"`
	EndAt           *Time    `json:"endAt"`
	BudgetValue     *float64 `json:"budgetValue"`
	BudgetValueType string   `json:"budgetValueType"`
}

// ValueRule weights one condition's conversions.
type ValueRule struct {
	Condition  string   `json:"condition"`
	Multiplier *float64 `json:"multiplier"`
}

// ValueRuleSet weights conversions so some audiences count for more.
type ValueRuleSet struct {
	ID     string      `json:"id"`
	Name   string      `json:"name"`
	Status string      `json:"status"`
	Rules  []ValueRule `json:"rules"`
}

// AccountActivityParams reads an ad account's change log. Since and Until are
// `YYYY-MM-DD`.
type AccountActivityParams struct {
	WorkspaceID  string
	ConnectionID string
	// AdAccountID is `act_…`.
	AdAccountID string
	Since       string
	Until       string
}

// AdLabelRequest creates or renames a label.
type AdLabelRequest struct {
	WorkspaceID  string `json:"workspaceId"`
	ConnectionID string `json:"connectionId"`
	// AdAccountID is `act_…`.
	AdAccountID string `json:"adAccountId"`
	Name        string `json:"name"`
}

// ApplyAdLabelRequest puts a label on an object. Level is "campaign",
// "ad_set" or "ad".
type ApplyAdLabelRequest struct {
	WorkspaceID  string `json:"workspaceId"`
	ConnectionID string `json:"connectionId"`
	// AdAccountID is `act_…`.
	AdAccountID string `json:"adAccountId"`
	ObjectID    string `json:"objectId"`
	Level       string `json:"level"`
}

// AdStudyCell is one arm of an A/B study.
type AdStudyCell struct {
	Name      string   `json:"name"`
	ObjectIDs []string `json:"objectIds"`
}

// CreateAdStudyRequest creates an A/B study over two to five cells.
type CreateAdStudyRequest struct {
	WorkspaceID  string `json:"workspaceId"`
	ConnectionID string `json:"connectionId"`
	// AdAccountID is `act_…`.
	AdAccountID string        `json:"adAccountId"`
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	StartAt     Time          `json:"startAt"`
	EndAt       Time          `json:"endAt"`
	Cells       []AdStudyCell `json:"cells"`
}

// CreateHighDemandPeriodRequest declares a heavier-spend window.
// BudgetValueType is "ABSOLUTE" or "MULTIPLIER".
type CreateHighDemandPeriodRequest struct {
	WorkspaceID  string `json:"workspaceId"`
	ConnectionID string `json:"connectionId"`
	// AdAccountID is `act_…`.
	AdAccountID     string  `json:"adAccountId"`
	StartAt         Time    `json:"startAt"`
	EndAt           Time    `json:"endAt"`
	BudgetValue     float64 `json:"budgetValue"`
	BudgetValueType string  `json:"budgetValueType"`
}

// CreateValueRuleSetRequest weights conversions across one to twenty rules.
type CreateValueRuleSetRequest struct {
	WorkspaceID  string `json:"workspaceId"`
	ConnectionID string `json:"connectionId"`
	// AdAccountID is `act_…`.
	AdAccountID string      `json:"adAccountId"`
	Name        string      `json:"name"`
	Rules       []ValueRule `json:"rules"`
}

// AccountActivity returns who changed what on the ad account, and when.
func (s *AdsService) AccountActivity(ctx context.Context, params *AccountActivityParams) (*AdActivityResult, error) {
	q := newQuery()
	if params != nil {
		q.str("workspace_id", params.WorkspaceID)
		q.str("connection_id", params.ConnectionID)
		q.str("ad_account_id", params.AdAccountID)
		q.str("since", params.Since)
		q.str("until", params.Until)
	}
	out := &AdActivityResult{}
	if err := s.client.json(ctx, "GET", "/ads/account/activity", nil, q.values(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// Labels returns the ad account's labels.
func (s *AdsService) Labels(ctx context.Context, params *AdAccountParams) ([]AdLabel, error) {
	var out []AdLabel
	if err := s.client.json(ctx, "GET", "/ads/account/labels", nil, params.query(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateLabel creates a label.
func (s *AdsService) CreateLabel(ctx context.Context, body *AdLabelRequest) (*AdLabel, error) {
	out := &AdLabel{}
	if err := s.client.json(ctx, "POST", "/ads/account/labels", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateLabel renames a label.
func (s *AdsService) UpdateLabel(ctx context.Context, id string, params *AdObjectParams, body *AdLabelRequest) (*AdLabel, error) {
	out := &AdLabel{}
	if err := s.client.json(ctx, "PATCH", "/ads/account/labels/"+id, body, params.query(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteLabel deletes a label.
func (s *AdsService) DeleteLabel(ctx context.Context, id string, params *AdAccountParams) error {
	return s.client.json(ctx, "DELETE", "/ads/account/labels/"+id, nil, params.query(), nil)
}

// ApplyLabel puts a label on a campaign, ad set or ad, keeping whatever labels
// the object already carries.
func (s *AdsService) ApplyLabel(ctx context.Context, id string, body *ApplyAdLabelRequest) error {
	return s.client.json(ctx, "POST", "/ads/account/labels/"+id+"/apply", body, nil, nil)
}

// Studies returns the ad account's A/B studies.
func (s *AdsService) Studies(ctx context.Context, params *AdAccountParams) ([]AdStudy, error) {
	var out []AdStudy
	if err := s.client.json(ctx, "GET", "/ads/account/studies", nil, params.query(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateStudy creates an A/B study, splitting traffic evenly across its cells.
func (s *AdsService) CreateStudy(ctx context.Context, body *CreateAdStudyRequest) (*AdStudy, error) {
	out := &AdStudy{}
	if err := s.client.json(ctx, "POST", "/ads/account/studies", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Study reads one A/B study.
func (s *AdsService) Study(ctx context.Context, id string, params *AdAccountParams) (*AdStudy, error) {
	out := &AdStudy{}
	if err := s.client.json(ctx, "GET", "/ads/account/studies/"+id, nil, params.query(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteStudy deletes an A/B study.
func (s *AdsService) DeleteStudy(ctx context.Context, id string, params *AdAccountParams) error {
	return s.client.json(ctx, "DELETE", "/ads/account/studies/"+id, nil, params.query(), nil)
}

// IosCampaignLimits returns how many iOS 14 campaigns the account may run at
// once, per app.
func (s *AdsService) IosCampaignLimits(ctx context.Context, params *AdAccountParams) ([]IosCampaignLimits, error) {
	var out []IosCampaignLimits
	if err := s.client.json(ctx, "GET", "/ads/account/ios-limits", nil, params.query(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// HighDemandPeriods returns the account's declared high-demand windows.
func (s *AdsService) HighDemandPeriods(ctx context.Context, params *AdAccountParams) ([]HighDemandPeriod, error) {
	var out []HighDemandPeriod
	if err := s.client.json(ctx, "GET", "/ads/account/high-demand-periods", nil, params.query(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateHighDemandPeriod tells the network to expect heavier spend over a
// window, so pacing allows for it.
func (s *AdsService) CreateHighDemandPeriod(ctx context.Context, body *CreateHighDemandPeriodRequest) (*HighDemandPeriod, error) {
	out := &HighDemandPeriod{}
	if err := s.client.json(ctx, "POST", "/ads/account/high-demand-periods", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteHighDemandPeriod deletes a high-demand window.
func (s *AdsService) DeleteHighDemandPeriod(ctx context.Context, id string, params *AdAccountParams) error {
	return s.client.json(ctx, "DELETE", "/ads/account/high-demand-periods/"+id, nil, params.query(), nil)
}

// ValueRuleSets returns the account's value rule sets.
func (s *AdsService) ValueRuleSets(ctx context.Context, params *AdAccountParams) ([]ValueRuleSet, error) {
	var out []ValueRuleSet
	if err := s.client.json(ctx, "GET", "/ads/account/value-rule-sets", nil, params.query(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateValueRuleSet weights conversions so some audiences count for more than
// others.
func (s *AdsService) CreateValueRuleSet(ctx context.Context, body *CreateValueRuleSetRequest) (*ValueRuleSet, error) {
	out := &ValueRuleSet{}
	if err := s.client.json(ctx, "POST", "/ads/account/value-rule-sets", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteValueRuleSet deletes a value rule set.
func (s *AdsService) DeleteValueRuleSet(ctx context.Context, id string, params *AdAccountParams) error {
	return s.client.json(ctx, "DELETE", "/ads/account/value-rule-sets/"+id, nil, params.query(), nil)
}
