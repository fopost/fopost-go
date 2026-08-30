package fopost

import (
	"context"
	"fmt"
	"io"
	"net/url"
)

// PostsService covers posts, publishing, deliveries, and bulk operations.
type PostsService struct{ client *Client }

// PostLabelRef is a label as it appears on a post.
type PostLabelRef struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// PostAccountResult is one account a post targets, with its delivery outcome.
type PostAccountResult struct {
	ID              string `json:"id"`
	Platform        string `json:"platform"`
	Username        string `json:"username"`
	Name            string `json:"name"`
	Avatar          string `json:"avatar"`
	PublishStatus   string `json:"publish_status"`
	PostedAt        Time   `json:"posted_at"`
	PlatformPostID  string `json:"platform_post_id"`
	ExternalURL     string `json:"external_url"`
	ErrorCode       string `json:"error_code"`
	ErrorMessage    string `json:"error_message"`
	RawErrorMessage string `json:"raw_error_message"`
	Attempts        int    `json:"attempts"`
	MaxAttempts     int    `json:"max_attempts"`
}

// Post is a composed post and everything scheduled or delivered from it.
type Post struct {
	ID                string              `json:"id"`
	WorkspaceID       string              `json:"workspace_id"`
	Status            string              `json:"status"`
	ContentType       string              `json:"content_type"`
	ScheduleAt        Time                `json:"schedule_at"`
	Repeatable        bool                `json:"repeatable"`
	RepeatableTimes   *int                `json:"repeatable_times"`
	RepeatableGap     *int                `json:"repeatable_gap"`
	RepeatableGapUnit string              `json:"repeatable_gap_unit"`
	RemainingPosts    *int                `json:"remaining_posts"`
	Title             string              `json:"title"`
	Summary           string              `json:"summary"`
	AutoPlug          bool                `json:"auto_plug"`
	AutoPlugContent   string              `json:"auto_plug_content"`
	Content           []ContentBlock      `json:"content"`
	Accounts          []PostAccountResult `json:"accounts"`
	Labels            []PostLabelRef      `json:"labels"`
	Settings          map[string]any      `json:"settings"`
	CreatedAt         Time                `json:"created_at"`
	UpdatedAt         Time                `json:"updated_at"`
}

// PostList is one page of posts.
type PostList struct {
	Data []Post   `json:"data"`
	Meta PageMeta `json:"meta"`
}

// ListPostsParams filters and paginates List. Zero fields are not sent, so the
// API applies its own defaults (page 1, 30 per page).
type ListPostsParams struct {
	Page        int
	PerPage     int
	WorkspaceID string
	// Status is one of the PostStatus constants.
	Status    string
	Search    string
	Platform  string
	Label     string
	AccountID string
	// Date, From and To are YYYY-MM-DD.
	Date string
	From string
	To   string
	// Sort is "oldest" to reverse the default newest-first order.
	Sort string
}

func (p *ListPostsParams) values() url.Values {
	q := newQuery()
	if p == nil {
		return q.values()
	}
	q.num("page", p.Page)
	q.num("per_page", p.PerPage)
	q.str("workspace_id", p.WorkspaceID)
	q.str("status", p.Status)
	q.str("search", p.Search)
	q.str("platform", p.Platform)
	q.str("label", p.Label)
	q.str("account_id", p.AccountID)
	q.str("date", p.Date)
	q.str("from", p.From)
	q.str("to", p.To)
	q.str("sort", p.Sort)
	return q.values()
}

// List returns one page of posts.
func (s *PostsService) List(ctx context.Context, params *ListPostsParams) (*PostList, error) {
	out := &PostList{}
	if err := s.client.Do(ctx, "GET", "/posts", nil, params.values(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// Each walks every matching post, one page at a time. Returning a non-nil
// error from fn stops the walk and returns that error.
func (s *PostsService) Each(ctx context.Context, params *ListPostsParams, fn func(Post) error) error {
	walk := ListPostsParams{}
	if params != nil {
		walk = *params
	}
	if walk.Page < 1 {
		walk.Page = 1
	}
	if walk.PerPage < 1 {
		walk.PerPage = 30
	}

	for {
		page, err := s.List(ctx, &walk)
		if err != nil {
			return err
		}
		for _, post := range page.Data {
			if err := fn(post); err != nil {
				return err
			}
		}
		if len(page.Data) == 0 {
			return nil
		}
		if page.Meta.LastPage > 0 && walk.Page >= page.Meta.LastPage {
			return nil
		}
		if page.Meta.LastPage == 0 && len(page.Data) < walk.PerPage {
			return nil
		}
		walk.Page++
	}
}

// ListAll collects every matching post. Prefer Each for large workspaces.
func (s *PostsService) ListAll(ctx context.Context, params *ListPostsParams) ([]Post, error) {
	var all []Post
	err := s.Each(ctx, params, func(post Post) error {
		all = append(all, post)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return all, nil
}

// Get returns one post.
func (s *PostsService) Get(ctx context.Context, id string) (*Post, error) {
	out := &Post{}
	if err := s.client.json(ctx, "GET", "/posts/"+url.PathEscape(id), nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreatePostRequest is the body of Create. Status is "draft" or "scheduled";
// a scheduled post needs ScheduleAt. To send a post out now, create it and
// call Publish.
type CreatePostRequest struct {
	WorkspaceID string         `json:"workspace_id"`
	Accounts    []string       `json:"accounts"`
	Content     []ContentBlock `json:"content"`
	// ContentType is "post", "thread", or "reel".
	ContentType string `json:"content_type,omitempty"`
	// ArtifactType is "text_post", "thread", "article", "carousel",
	// "short_video", or "link_share".
	ArtifactType      string   `json:"artifact_type,omitempty"`
	Status            string   `json:"status,omitempty"`
	ScheduleAt        *Time    `json:"schedule_at,omitempty"`
	Repeatable        *bool    `json:"repeatable,omitempty"`
	RepeatableTimes   *int     `json:"repeatable_times,omitempty"`
	RepeatableGap     *int     `json:"repeatable_gap,omitempty"`
	RepeatableGapUnit string   `json:"repeatable_gap_unit,omitempty"`
	Labels            []string `json:"labels,omitempty"`
	Title             *string  `json:"title,omitempty"`
	InternalTitle     *string  `json:"internal_title,omitempty"`
	Summary           *string  `json:"summary,omitempty"`
	AutoPlug          *bool    `json:"auto_plug,omitempty"`
	AutoPlugContent   *string  `json:"auto_plug_content,omitempty"`
	// Settings holds per-platform options, keyed by platform.
	Settings    map[string]any `json:"settings,omitempty"`
	SourceIDs   []string       `json:"source_ids,omitempty"`
	CompanionOf *string        `json:"companion_of,omitempty"`
}

// Create composes a draft or a scheduled post.
func (s *PostsService) Create(ctx context.Context, body *CreatePostRequest) (*Post, error) {
	out := &Post{}
	if err := s.client.json(ctx, "POST", "/posts", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdatePostRequest is the body of Update. Only the fields you set are sent,
// so an update is a partial one.
type UpdatePostRequest struct {
	Accounts          []string       `json:"accounts,omitempty"`
	Content           []ContentBlock `json:"content,omitempty"`
	ContentType       string         `json:"content_type,omitempty"`
	ArtifactType      string         `json:"artifact_type,omitempty"`
	Status            string         `json:"status,omitempty"`
	ScheduleAt        *Time          `json:"schedule_at,omitempty"`
	Repeatable        *bool          `json:"repeatable,omitempty"`
	RepeatableTimes   *int           `json:"repeatable_times,omitempty"`
	RepeatableGap     *int           `json:"repeatable_gap,omitempty"`
	RepeatableGapUnit string         `json:"repeatable_gap_unit,omitempty"`
	Labels            []string       `json:"labels,omitempty"`
	Title             *string        `json:"title,omitempty"`
	InternalTitle     *string        `json:"internal_title,omitempty"`
	Summary           *string        `json:"summary,omitempty"`
	AutoPlug          *bool          `json:"auto_plug,omitempty"`
	AutoPlugContent   *string        `json:"auto_plug_content,omitempty"`
	Settings          map[string]any `json:"settings,omitempty"`
}

// Update edits a post that has not been published.
func (s *PostsService) Update(ctx context.Context, id string, body *UpdatePostRequest) (*Post, error) {
	out := &Post{}
	if err := s.client.json(ctx, "PUT", "/posts/"+url.PathEscape(id), body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Delete removes a post.
func (s *PostsService) Delete(ctx context.Context, id string) error {
	return s.client.Do(ctx, "DELETE", "/posts/"+url.PathEscape(id), nil, nil, nil)
}

// Duplicate copies a post into a new draft.
func (s *PostsService) Duplicate(ctx context.Context, id string) (*DuplicatedPost, error) {
	out := &DuplicatedPost{}
	if err := s.client.json(ctx, "POST", "/posts/"+url.PathEscape(id)+"/duplicate", nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// DuplicatedPost is the id and status of the copy Duplicate created.
type DuplicatedPost struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// PublishOptions carries the flags Publish accepts.
type PublishOptions struct {
	// AccountIDs narrows publishing to a subset of the post's accounts.
	AccountIDs []string `json:"accountIds,omitempty"`
	// DryRun validates without sending anything to a platform.
	DryRun bool `json:"-"`
}

// PublishDelivery is one account's delivery as reported by publish or retry.
type PublishDelivery struct {
	ID                 string `json:"id"`
	AccountID          string `json:"accountId"`
	Status             string `json:"status"`
	ErrorCode          string `json:"errorCode"`
	ErrorMessage       string `json:"errorMessage"`
	PlatformPostID     string `json:"platformPostId"`
	ExternalURL        string `json:"externalUrl"`
	PostedAt           Time   `json:"postedAt"`
	Attempts           int    `json:"attempts"`
	ScheduledPublishAt Time   `json:"scheduledPublishAt"`
	DelayReason        string `json:"delayReason"`
	DelayMessage       string `json:"delayMessage"`
}

// PublishResult is the outcome of a publish. On a dry run, DryRun is true and
// Deliveries is empty — the per-account plan is in Accounts instead.
type PublishResult struct {
	DryRun         bool              `json:"dryRun"`
	PostStatus     string            `json:"post_status"`
	Deliveries     []PublishDelivery `json:"deliveries"`
	HealthWarnings []HealthWarning   `json:"healthWarnings"`
	Post           struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	} `json:"post"`
	Accounts []struct {
		AccountID string `json:"accountId"`
		Platform  string `json:"platform"`
	} `json:"accounts"`
}

// Publish queues a post for immediate delivery to its accounts. Nothing
// reaches a platform without this call or a schedule the user set.
func (s *PostsService) Publish(ctx context.Context, id string, opts *PublishOptions) (*PublishResult, error) {
	body := map[string]any{}
	if opts != nil {
		if len(opts.AccountIDs) > 0 {
			body["accountIds"] = opts.AccountIDs
		}
		if opts.DryRun {
			body["options"] = map[string]any{"dryRun": true}
		}
	}
	out := &PublishResult{}
	if err := s.client.json(ctx, "POST", "/posts/"+url.PathEscape(id)+"/publish", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// RetryOptions narrows what Retry re-sends.
type RetryOptions struct {
	AccountIDs       []string `json:"accountIds,omitempty"`
	IncludePublished bool     `json:"includePublished,omitempty"`
}

// RetryResult reports the deliveries retried and the ones out of attempts.
type RetryResult struct {
	PostStatus string            `json:"post_status"`
	Deliveries []PublishDelivery `json:"deliveries"`
	Exceeded   []struct {
		AccountID string `json:"accountId"`
		Platform  string `json:"platform"`
		Attempts  int    `json:"attempts"`
	} `json:"exceeded"`
}

// Retry re-sends the deliveries that failed, leaving successful ones alone.
func (s *PostsService) Retry(ctx context.Context, id string, opts *RetryOptions) (*RetryResult, error) {
	out := &RetryResult{}
	if err := s.client.json(ctx, "POST", "/posts/"+url.PathEscape(id)+"/retry", bodyOrEmpty(opts), nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// CancelOptions narrows what Cancel stops.
type CancelOptions struct {
	AccountIDs []string `json:"accountIds,omitempty"`
}

// CancelResult reports the deliveries that were stopped.
type CancelResult struct {
	PostStatus string `json:"post_status"`
	Deliveries []struct {
		ID        string `json:"id"`
		AccountID string `json:"accountId"`
		Status    string `json:"status"`
	} `json:"deliveries"`
}

// Cancel stops the deliveries that have not gone out yet.
func (s *PostsService) Cancel(ctx context.Context, id string, opts *CancelOptions) (*CancelResult, error) {
	out := &CancelResult{}
	if err := s.client.json(ctx, "POST", "/posts/"+url.PathEscape(id)+"/cancel", bodyOrEmpty(opts), nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// PreflightResult lists per-account blockers and advisory content signals.
type PreflightResult struct {
	Ready bool `json:"ready"`
	Post  struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	} `json:"post"`
	Accounts []PreflightAccount `json:"accounts"`
}

// PreflightAccount is one account's readiness. Issues block publishing;
// Signals are advisory.
type PreflightAccount struct {
	AccountID string          `json:"accountId"`
	Platform  string          `json:"platform"`
	Username  string          `json:"username"`
	Ready     bool            `json:"ready"`
	Issues    []string        `json:"issues"`
	Score     float64         `json:"score"`
	Signals   []ContentSignal `json:"signals"`
}

// Preflight checks a post against every target platform without publishing.
func (s *PostsService) Preflight(ctx context.Context, id string) (*PreflightResult, error) {
	out := &PreflightResult{}
	if err := s.client.json(ctx, "POST", "/posts/"+url.PathEscape(id)+"/preflight", nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Delivery is one account's delivery record for a post.
type Delivery struct {
	ID                 string `json:"id"`
	AccountID          string `json:"accountId"`
	Status             string `json:"status"`
	ErrorCode          string `json:"errorCode"`
	ErrorMessage       string `json:"errorMessage"`
	Attempts           int    `json:"attempts"`
	MaxAttempts        int    `json:"maxAttempts"`
	ScheduledPublishAt Time   `json:"scheduledPublishAt"`
	DelayReason        string `json:"delayReason"`
	DelayMessage       string `json:"delayMessage"`
	PostedAt           Time   `json:"postedAt"`
	LastAttemptAt      Time   `json:"lastAttemptAt"`
	PlatformPostID     string `json:"platformPostId"`
	ExternalURL        string `json:"externalUrl"`
	Platform           string `json:"platform"`
	Username           string `json:"username"`
	AccountName        string `json:"accountName"`
}

// Deliveries lists the current delivery record per account.
func (s *PostsService) Deliveries(ctx context.Context, id string) ([]Delivery, error) {
	var out []Delivery
	if err := s.client.json(ctx, "GET", "/posts/"+url.PathEscape(id)+"/deliveries", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// PublishRunDelivery is one account's outcome within a single publish run.
type PublishRunDelivery struct {
	AccountID      string `json:"account_id"`
	AccountName    string `json:"account_name"`
	Username       string `json:"username"`
	Platform       string `json:"platform"`
	Status         string `json:"status"`
	AttemptNumber  int    `json:"attempt_number"`
	ErrorCode      string `json:"error_code"`
	ErrorMessage   string `json:"error_message"`
	PlatformPostID string `json:"platform_post_id"`
	ExternalURL    string `json:"external_url"`
	StartedAt      Time   `json:"started_at"`
	CompletedAt    Time   `json:"completed_at"`
	DurationMs     *int   `json:"duration_ms"`
}

// PublishRun is one attempt at publishing a post, with its per-account results.
type PublishRun struct {
	ID          string               `json:"id"`
	RunNumber   int                  `json:"run_number"`
	Status      string               `json:"status"`
	StartedAt   Time                 `json:"started_at"`
	CompletedAt Time                 `json:"completed_at"`
	Deliveries  []PublishRunDelivery `json:"deliveries"`
}

// PublishRuns lists every publish attempt made for a post, newest first.
func (s *PostsService) PublishRuns(ctx context.Context, id string) ([]PublishRun, error) {
	var out []PublishRun
	if err := s.client.json(ctx, "GET", "/posts/"+url.PathEscape(id)+"/publish-runs", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// PostAnalyticsMetrics is one platform's numbers for a post. A nil field means
// the platform does not report that metric.
type PostAnalyticsMetrics struct {
	Impressions        *int           `json:"impressions"`
	Reach              *int           `json:"reach"`
	Engagements        *int           `json:"engagements"`
	Likes              *int           `json:"likes"`
	Comments           *int           `json:"comments"`
	Shares             *int           `json:"shares"`
	Reposts            *int           `json:"reposts"`
	Clicks             *int           `json:"clicks"`
	Saves              *int           `json:"saves"`
	VideoViews         *int           `json:"videoViews"`
	WatchTimeMs        *int           `json:"watchTimeMs"`
	AvgWatchTimeMs     *int           `json:"avgWatchTimeMs"`
	ReactionsBreakdown map[string]any `json:"reactionsBreakdown"`
	Follows            *int           `json:"follows"`
	FetchedAt          Time           `json:"fetchedAt"`
}

// PostAnalytics is a post's performance, totalled and per platform.
type PostAnalytics struct {
	PostID string `json:"postId"`
	Totals struct {
		Impressions int `json:"impressions"`
		Reach       int `json:"reach"`
		Engagements int `json:"engagements"`
		Likes       int `json:"likes"`
		Comments    int `json:"comments"`
		Shares      int `json:"shares"`
		Reposts     int `json:"reposts"`
		Clicks      int `json:"clicks"`
		Saves       int `json:"saves"`
		VideoViews  int `json:"videoViews"`
		Follows     int `json:"follows"`
	} `json:"totals"`
	Platforms []struct {
		Platform       string               `json:"platform"`
		Username       string               `json:"username"`
		ExternalPostID string               `json:"externalPostId"`
		Permalink      string               `json:"permalink"`
		ThumbnailURL   string               `json:"thumbnailUrl"`
		MediaType      string               `json:"mediaType"`
		PostedAt       Time                 `json:"postedAt"`
		Metrics        PostAnalyticsMetrics `json:"metrics"`
	} `json:"platforms"`
	LastFetchedAt Time `json:"lastFetchedAt"`
}

// Analytics returns a post's performance across the platforms it reached.
func (s *PostsService) Analytics(ctx context.Context, id string) (*PostAnalytics, error) {
	out := &PostAnalytics{}
	if err := s.client.json(ctx, "GET", "/posts/"+url.PathEscape(id)+"/analytics", nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// BulkResult reports how many posts a bulk action changed.
type BulkResult struct {
	Updated int    `json:"updated"`
	Action  string `json:"action"`
	Mode    string `json:"mode,omitempty"`
}

// BulkShift moves a selection's schedule by offsetMinutes, which may be
// negative but never zero. Only drafts and scheduled posts can be shifted, and
// one ineligible post in the selection changes nothing at all.
func (s *PostsService) BulkShift(ctx context.Context, workspaceID string, postIDs []string, offsetMinutes int) (*BulkResult, error) {
	body := map[string]any{
		"action":         "shift",
		"workspace_id":   workspaceID,
		"post_ids":       postIDs,
		"offset_minutes": offsetMinutes,
	}
	return s.bulk(ctx, body)
}

// BulkLabel relabels a selection. Mode is "replace" (the default, where an
// empty labelIDs clears them), "add", or "remove".
func (s *PostsService) BulkLabel(ctx context.Context, workspaceID string, postIDs, labelIDs []string, mode string) (*BulkResult, error) {
	if labelIDs == nil {
		labelIDs = []string{}
	}
	body := map[string]any{
		"action":       "label",
		"workspace_id": workspaceID,
		"post_ids":     postIDs,
		"label_ids":    labelIDs,
	}
	if mode != "" {
		body["mode"] = mode
	}
	return s.bulk(ctx, body)
}

// BulkDelete removes a selection of posts in one transaction.
func (s *PostsService) BulkDelete(ctx context.Context, workspaceID string, postIDs []string) (*BulkResult, error) {
	body := map[string]any{
		"action":       "delete",
		"workspace_id": workspaceID,
		"post_ids":     postIDs,
	}
	return s.bulk(ctx, body)
}

func (s *PostsService) bulk(ctx context.Context, body map[string]any) (*BulkResult, error) {
	out := &BulkResult{}
	if err := s.client.Do(ctx, "POST", "/posts/bulk", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// BulkImportRow is one CSV row as the validator read it.
type BulkImportRow struct {
	Row            int      `json:"row"`
	ContentPreview string   `json:"content_preview"`
	ScheduleAt     Time     `json:"schedule_at"`
	Accounts       []string `json:"accounts"`
	Labels         int      `json:"labels"`
	HasMedia       bool     `json:"has_media"`
	Errors         []string `json:"errors"`
}

// BulkImportValidation is the dry run of a CSV import.
type BulkImportValidation struct {
	TotalRows   int             `json:"total_rows"`
	ValidRows   int             `json:"valid_rows"`
	InvalidRows int             `json:"invalid_rows"`
	Rows        []BulkImportRow `json:"rows"`
}

// BulkImportResult is what a committed CSV import created. Keep BatchID to
// roll the whole batch back.
type BulkImportResult struct {
	BatchID string `json:"batch_id"`
	Created int    `json:"created"`
	Posts   []struct {
		ID         string `json:"id"`
		ScheduleAt Time   `json:"schedule_at"`
	} `json:"posts"`
}

// ValidateBulkImport checks a CSV without creating anything.
func (s *PostsService) ValidateBulkImport(ctx context.Context, workspaceID, filename string, csv io.Reader) (*BulkImportValidation, error) {
	out := &BulkImportValidation{}
	if err := s.uploadCSV(ctx, "/posts/bulk-import/validate", workspaceID, filename, csv, out); err != nil {
		return nil, err
	}
	return out, nil
}

// CommitBulkImport creates the posts a CSV describes.
func (s *PostsService) CommitBulkImport(ctx context.Context, workspaceID, filename string, csv io.Reader) (*BulkImportResult, error) {
	out := &BulkImportResult{}
	if err := s.uploadCSV(ctx, "/posts/bulk-import/commit", workspaceID, filename, csv, out); err != nil {
		return nil, err
	}
	return out, nil
}

// RollbackBulkImport deletes every post a committed batch created.
func (s *PostsService) RollbackBulkImport(ctx context.Context, batchID string) (*Message, error) {
	out := &Message{}
	if err := s.client.Do(ctx, "DELETE", "/posts/bulk-import/"+url.PathEscape(batchID), nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *PostsService) uploadCSV(ctx context.Context, path, workspaceID, filename string, csv io.Reader, out any) error {
	if workspaceID == "" {
		return fmt.Errorf("fopost: a workspace id is required")
	}
	if filename == "" {
		filename = "posts.csv"
	}
	form, err := buildMultipart(map[string]string{"workspace_id": workspaceID}, []fileField{
		{field: "file", filename: filename, reader: csv},
	})
	if err != nil {
		return err
	}
	return s.client.do(ctx, &request{
		method:      "POST",
		path:        path,
		body:        form.body,
		contentType: form.contentType,
		unwrap:      true,
	}, out)
}

// bodyOrEmpty keeps a nil options struct from becoming a literal "null" body.
func bodyOrEmpty(v any) any {
	switch typed := v.(type) {
	case *RetryOptions:
		if typed == nil {
			return map[string]any{}
		}
	case *CancelOptions:
		if typed == nil {
			return map[string]any{}
		}
	}
	return v
}
