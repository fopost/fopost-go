package fopost

import (
	"context"
	"net/url"
	"strings"
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

// Creative formats beyond image, video and carousel.
const (
	CreativeCatalog     = "catalog"
	CreativePartnership = "partnership"
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

// ─── Product catalogs ───────────────────────────────────────────────

// ProductCatalog is a catalog on the connection's business portfolio, read live.
type ProductCatalog struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Vertical     string `json:"vertical"`
	ProductCount *int   `json:"productCount"`
}

// ProductCatalogsResult is the catalogs one connection reaches.
type ProductCatalogsResult struct {
	Catalogs    []ProductCatalog `json:"catalogs"`
	WorkspaceID string           `json:"workspaceId"`
}

// CatalogProduct is one product in a catalog.
type CatalogProduct struct {
	ID string `json:"id"`
	// RetailerID is your own key for the product.
	RetailerID   string `json:"retailerId"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Availability string `json:"availability"`
	Condition    string `json:"condition"`
	// PriceMinor is minor units of Currency.
	PriceMinor *int   `json:"priceMinor"`
	Currency   string `json:"currency"`
	ImageURL   string `json:"imageUrl"`
	URL        string `json:"url"`
}

// CatalogProductsPage is one page of products; pass NextCursor back as After.
type CatalogProductsPage struct {
	Products   []CatalogProduct `json:"products"`
	NextCursor string           `json:"nextCursor"`
}

// CatalogProductWrite is one upsert or delete in a product batch. Op is
// "upsert" or "delete"; a delete needs only RetailerID.
type CatalogProductWrite struct {
	Op           string `json:"op"`
	RetailerID   string `json:"retailerId"`
	Name         string `json:"name,omitempty"`
	Description  string `json:"description,omitempty"`
	URL          string `json:"url,omitempty"`
	ImageURL     string `json:"imageUrl,omitempty"`
	PriceMinor   int    `json:"priceMinor,omitempty"`
	Currency     string `json:"currency,omitempty"`
	Availability string `json:"availability,omitempty"`
	Condition    string `json:"condition,omitempty"`
	Brand        string `json:"brand,omitempty"`
}

// CatalogBatchResult is what a product batch was accepted as.
type CatalogBatchResult struct {
	Handles []string `json:"handles"`
	// Accepted is how many products were sent in this batch.
	Accepted int `json:"accepted"`
}

// ProductFeed keeps a catalog in step with a product file you host.
type ProductFeed struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// URL is set when the network fetches the file on a schedule.
	URL       string `json:"url"`
	Schedule  string `json:"schedule"`
	CreatedAt *Time  `json:"createdAt"`
}

// ProductFeedUpload is one run the network made of a feed.
type ProductFeedUpload struct {
	ID           string `json:"id"`
	StartedAt    *Time  `json:"startedAt"`
	EndedAt      *Time  `json:"endedAt"`
	Status       string `json:"status"`
	ErrorCount   *int   `json:"errorCount"`
	WarningCount *int   `json:"warningCount"`
}

// ProductSet is the slice of a catalog one catalog ad runs from.
type ProductSet struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ProductCount *int   `json:"productCount"`
	// Filter is the network's own product-set filter.
	Filter map[string]any `json:"filter"`
}

// CreateCatalogRequest creates a catalog on the connection's portfolio.
type CreateCatalogRequest struct {
	WorkspaceID  string `json:"workspaceId"`
	ConnectionID string `json:"connectionId"`
	Name         string `json:"name"`
	// Vertical is the network's catalog vertical; "commerce" when empty.
	Vertical string `json:"vertical,omitempty"`
}

// UpdateCatalogRequest renames a catalog.
type UpdateCatalogRequest struct {
	WorkspaceID  string `json:"workspaceId"`
	ConnectionID string `json:"connectionId"`
	Name         string `json:"name"`
}

// CatalogProductBatchRequest writes up to 500 products in one batch.
type CatalogProductBatchRequest struct {
	WorkspaceID  string                `json:"workspaceId"`
	ConnectionID string                `json:"connectionId"`
	Products     []CatalogProductWrite `json:"products"`
}

// CreateProductFeedRequest creates a feed. Schedule needs URL.
type CreateProductFeedRequest struct {
	WorkspaceID  string `json:"workspaceId"`
	ConnectionID string `json:"connectionId"`
	Name         string `json:"name"`
	URL          string `json:"url,omitempty"`
	// Schedule is "HOURLY", "DAILY" or "WEEKLY".
	Schedule string `json:"schedule,omitempty"`
}

// StartFeedUploadRequest fetches a feed now.
type StartFeedUploadRequest struct {
	WorkspaceID  string `json:"workspaceId"`
	ConnectionID string `json:"connectionId"`
	// URL overrides the feed's own url for this run.
	URL string `json:"url,omitempty"`
}

// ProductSetRequest creates or updates a product set. Without a Filter the set
// is the whole catalog.
type ProductSetRequest struct {
	WorkspaceID  string         `json:"workspaceId"`
	ConnectionID string         `json:"connectionId"`
	Name         string         `json:"name"`
	Filter       map[string]any `json:"filter,omitempty"`
}

// Catalogs returns the catalogs the connection's portfolios reach, read live.
func (s *AdsService) Catalogs(ctx context.Context, params *AdObjectParams) (*ProductCatalogsResult, error) {
	out := &ProductCatalogsResult{}
	if err := s.client.json(ctx, "GET", "/ads/catalogs", nil, params.query(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateCatalog creates a catalog. Needs the `publish` scope as well as `ads`.
func (s *AdsService) CreateCatalog(ctx context.Context, body *CreateCatalogRequest) (*ProductCatalog, error) {
	out := &ProductCatalog{}
	if err := s.client.json(ctx, "POST", "/ads/catalogs", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Catalog reads one catalog.
func (s *AdsService) Catalog(ctx context.Context, id string, params *AdObjectParams) (*ProductCatalog, error) {
	out := &ProductCatalog{}
	if err := s.client.json(ctx, "GET", "/ads/catalogs/"+id, nil, params.query(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateCatalog renames a catalog. Needs `publish` as well as `ads`.
func (s *AdsService) UpdateCatalog(ctx context.Context, id string, params *AdObjectParams, body *UpdateCatalogRequest) (*ProductCatalog, error) {
	out := &ProductCatalog{}
	if err := s.client.json(ctx, "PATCH", "/ads/catalogs/"+id, body, params.query(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteCatalog deletes a catalog with every product, feed and set in it.
// Needs `publish` as well as `ads`.
func (s *AdsService) DeleteCatalog(ctx context.Context, id string, params *AdObjectParams) error {
	return s.client.json(ctx, "DELETE", "/ads/catalogs/"+id, nil, params.query(), nil)
}

// CatalogProductsParams pages a catalog's products.
type CatalogProductsParams struct {
	WorkspaceID  string
	ConnectionID string
	// After is a NextCursor from a previous page.
	After string
}

// CatalogProducts returns one page of a catalog's products.
func (s *AdsService) CatalogProducts(ctx context.Context, id string, params *CatalogProductsParams) (*CatalogProductsPage, error) {
	q := newQuery()
	if params != nil {
		q.str("workspace_id", params.WorkspaceID)
		q.str("connection_id", params.ConnectionID)
		q.str("after", params.After)
	}
	out := &CatalogProductsPage{}
	if err := s.client.json(ctx, "GET", "/ads/catalogs/"+id+"/products", nil, q.values(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// WriteCatalogProducts upserts and deletes products in one batch, keyed by
// your own retailer id. Needs `publish` as well as `ads`.
func (s *AdsService) WriteCatalogProducts(ctx context.Context, id string, body *CatalogProductBatchRequest) (*CatalogBatchResult, error) {
	out := &CatalogBatchResult{}
	if err := s.client.json(ctx, "POST", "/ads/catalogs/"+id+"/products", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ProductFeeds returns a catalog's feeds.
func (s *AdsService) ProductFeeds(ctx context.Context, id string, params *AdObjectParams) ([]ProductFeed, error) {
	var out []ProductFeed
	if err := s.client.json(ctx, "GET", "/ads/catalogs/"+id+"/feeds", nil, params.query(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateProductFeed creates a feed. Needs `publish` as well as `ads`.
func (s *AdsService) CreateProductFeed(ctx context.Context, id string, body *CreateProductFeedRequest) (*ProductFeed, error) {
	out := &ProductFeed{}
	if err := s.client.json(ctx, "POST", "/ads/catalogs/"+id+"/feeds", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteProductFeed deletes a feed. Needs `publish` as well as `ads`.
func (s *AdsService) DeleteProductFeed(ctx context.Context, id, feedID string, params *AdObjectParams) error {
	return s.client.json(ctx, "DELETE", "/ads/catalogs/"+id+"/feeds/"+feedID, nil, params.query(), nil)
}

// FeedUploads returns each run the network made of a feed.
func (s *AdsService) FeedUploads(ctx context.Context, id, feedID string, params *AdObjectParams) ([]ProductFeedUpload, error) {
	var out []ProductFeedUpload
	path := "/ads/catalogs/" + id + "/feeds/" + feedID + "/uploads"
	if err := s.client.json(ctx, "GET", path, nil, params.query(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// StartFeedUpload fetches the feed now and returns the id of the run. Needs
// `publish` as well as `ads`.
func (s *AdsService) StartFeedUpload(ctx context.Context, id, feedID string, body *StartFeedUploadRequest) (string, error) {
	out := struct {
		ID string `json:"id"`
	}{}
	path := "/ads/catalogs/" + id + "/feeds/" + feedID + "/uploads"
	if err := s.client.json(ctx, "POST", path, body, nil, &out); err != nil {
		return "", err
	}
	return out.ID, nil
}

// ProductSets returns a catalog's product sets. A catalog ad runs from a set,
// not the whole catalog.
func (s *AdsService) ProductSets(ctx context.Context, id string, params *AdObjectParams) ([]ProductSet, error) {
	var out []ProductSet
	if err := s.client.json(ctx, "GET", "/ads/catalogs/"+id+"/product-sets", nil, params.query(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateProductSet creates a product set. Needs `publish` as well as `ads`.
func (s *AdsService) CreateProductSet(ctx context.Context, id string, body *ProductSetRequest) (*ProductSet, error) {
	out := &ProductSet{}
	if err := s.client.json(ctx, "POST", "/ads/catalogs/"+id+"/product-sets", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateProductSet renames or refilters a product set. Needs `publish`.
func (s *AdsService) UpdateProductSet(ctx context.Context, id, setID string, params *AdObjectParams, body *ProductSetRequest) (*ProductSet, error) {
	out := &ProductSet{}
	path := "/ads/catalogs/" + id + "/product-sets/" + setID
	if err := s.client.json(ctx, "PATCH", path, body, params.query(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteProductSet deletes a product set. Needs `publish` as well as `ads`.
func (s *AdsService) DeleteProductSet(ctx context.Context, id, setID string, params *AdObjectParams) error {
	path := "/ads/catalogs/" + id + "/product-sets/" + setID
	return s.client.json(ctx, "DELETE", path, nil, params.query(), nil)
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
func (s *AdsService) ReserveReachFrequency(ctx context.Context, id string, body *ReachFrequencyActionRequest) (*ReachFrequencyPrediction, error) {
	return s.reachFrequencyAction(ctx, id, "reserve", body)
}

// CancelReachFrequency cancels a reservation. Needs `publish` as well as `ads`.
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

// ─── Ad Library ─────────────────────────────────────────────────────

// LibraryEntry is one public archive entry. Read live on every search and
// stored nowhere.
type LibraryEntry struct {
	ID                 string   `json:"id"`
	PageID             string   `json:"pageId"`
	PageName           string   `json:"pageName"`
	Bodies             []string `json:"bodies"`
	Titles             []string `json:"titles"`
	LinkURLs           []string `json:"linkUrls"`
	SnapshotURL        string   `json:"snapshotUrl"`
	PublisherPlatforms []string `json:"publisherPlatforms"`
	StartedAt          *Time    `json:"startedAt"`
	EndedAt            *Time    `json:"endedAt"`
	// Currency and the ranges below are only on the archive's disclosure entries.
	Currency         string `json:"currency"`
	SpendLower       *int   `json:"spendLower"`
	SpendUpper       *int   `json:"spendUpper"`
	ImpressionsLower *int   `json:"impressionsLower"`
	ImpressionsUpper *int   `json:"impressionsUpper"`
}

// LibraryPage is one page of archive results.
type LibraryPage struct {
	Entries    []LibraryEntry `json:"entries"`
	NextCursor string         `json:"nextCursor"`
}

// LibraryParams searches the public archive. Countries is required; search by
// Q or by PageIDs.
type LibraryParams struct {
	WorkspaceID  string
	ConnectionID string
	// Countries are ISO 3166-1 alpha-2 codes the ad reached.
	Countries []string
	Q         string
	PageIDs   []string
	// ActiveStatus is "ACTIVE", "INACTIVE" or "ALL".
	ActiveStatus string
	Limit        int
	After        string
}

// Library searches the public ad archive: ads anyone is running. Read live on
// every call and stored nowhere, so an ad that stops running is simply absent
// from the next search.
func (s *AdsService) Library(ctx context.Context, params *LibraryParams) (*LibraryPage, error) {
	q := newQuery()
	if params != nil {
		q.str("workspace_id", params.WorkspaceID)
		q.str("connection_id", params.ConnectionID)
		q.str("countries", strings.Join(params.Countries, ","))
		q.str("q", params.Q)
		q.str("page_ids", strings.Join(params.PageIDs, ","))
		q.str("active_status", params.ActiveStatus)
		q.num("limit", params.Limit)
		q.str("after", params.After)
	}
	out := &LibraryPage{}
	if err := s.client.json(ctx, "GET", "/ads/library", nil, q.values(), out); err != nil {
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
