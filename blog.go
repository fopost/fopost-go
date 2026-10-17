package fopost

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
)

// BlogService plans keywords, writes articles and publishes them to a
// website. A post goes to a social network; an article goes to a website.
//
// Reads, settings, keyword edits and SubmitArticle need the `posts` scope.
// SuggestSeeds, PlanKeywords, RequestArticle and RegenerateArticle spend
// credits and need `ai`; an empty balance answers 402 (IsPaymentRequired,
// UpgradeURL). PublishArticle needs `publish` and answers 409 `no_target`
// until the settings name a target.
type BlogService struct{ client *Client }

// Article statuses.
const (
	BlogArticleStatusQueued     = "queued"
	BlogArticleStatusWriting    = "writing"
	BlogArticleStatusDraft      = "draft"
	BlogArticleStatusInReview   = "in_review"
	BlogArticleStatusScheduled  = "scheduled"
	BlogArticleStatusPublishing = "publishing"
	BlogArticleStatusPublished  = "published"
	BlogArticleStatusFailed     = "failed"
	BlogArticleStatusRejected   = "rejected"
)

// Keyword statuses.
const (
	BlogKeywordStatusSuggested = "suggested"
	BlogKeywordStatusPlanned   = "planned"
	BlogKeywordStatusWritten   = "written"
	BlogKeywordStatusDismissed = "dismissed"
)

// What the blog is for.
const (
	BlogGoalTraffic   = "traffic"
	BlogGoalLeads     = "leads"
	BlogGoalSales     = "sales"
	BlogGoalSignups   = "signups"
	BlogGoalAuthority = "authority"
)

// Where articles publish.
const (
	BlogTargetProductBlog = "product_blog"
	BlogTargetAccount     = "account"
	BlogTargetWebsite     = "website"
)

// BlogSettings is a workspace's blog configuration.
type BlogSettings struct {
	SiteURL  *string `json:"siteUrl"`
	Topic    *string `json:"topic"`
	Audience *string `json:"audience"`
	// Goal is one of the BlogGoal constants, or nil when unset.
	Goal           *string `json:"goal"`
	FocusTopics    *string `json:"focusTopics"`
	AvoidTopics    *string `json:"avoidTopics"`
	Language       string  `json:"language"`
	LocationCode   int     `json:"locationCode"`
	CadencePerWeek int     `json:"cadencePerWeek"`
	// ReviewHoldDays is how many days before its planned date a picked
	// keyword starts being written.
	ReviewHoldDays int     `json:"reviewHoldDays"`
	ImageTier      string  `json:"imageTier"`
	InBodyImages   int     `json:"inBodyImages"`
	MinWords       int     `json:"minWords"`
	MaxWords       int     `json:"maxWords"`
	BrandVoiceID   *string `json:"brandVoiceId"`
	// TargetKind is one of the BlogTarget constants, or nil when unset.
	TargetKind      *string `json:"targetKind"`
	TargetAccountID *string `json:"targetAccountId"`
	TargetPlatform  *string `json:"targetPlatform"`
	// WebsiteURL receives published articles as a signed webhook.
	WebsiteURL       *string `json:"websiteUrl"`
	HasWebsiteSecret bool    `json:"hasWebsiteSecret"`
	// PlanRun is the last keyword research run, nil before the first.
	PlanRun *BlogPlanRun `json:"planRun"`
	// KeywordDataAvailable reports whether search volume and difficulty can
	// be fetched on this deployment.
	KeywordDataAvailable bool `json:"keywordDataAvailable"`
	UpdatedAt            Time `json:"updatedAt"`
}

// BlogPlanRun is the state of a keyword research run.
type BlogPlanRun struct {
	// Status is queued, running, done or failed.
	Status string  `json:"status"`
	Error  *string `json:"error"`
	// Added is how many new keywords the run found.
	Added      *int  `json:"added"`
	FinishedAt *Time `json:"finishedAt"`
}

// UpdateBlogSettingsRequest is the body of UpdateSettings. Only the fields
// you set are sent.
type UpdateBlogSettingsRequest struct {
	SiteURL         *string `json:"siteUrl,omitempty"`
	Topic           *string `json:"topic,omitempty"`
	Audience        *string `json:"audience,omitempty"`
	Goal            *string `json:"goal,omitempty"`
	FocusTopics     *string `json:"focusTopics,omitempty"`
	AvoidTopics     *string `json:"avoidTopics,omitempty"`
	Language        string  `json:"language,omitempty"`
	LocationCode    *int    `json:"locationCode,omitempty"`
	CadencePerWeek  *int    `json:"cadencePerWeek,omitempty"`
	ReviewHoldDays  *int    `json:"reviewHoldDays,omitempty"`
	ImageTier       string  `json:"imageTier,omitempty"`
	InBodyImages    *int    `json:"inBodyImages,omitempty"`
	MinWords        *int    `json:"minWords,omitempty"`
	MaxWords        *int    `json:"maxWords,omitempty"`
	BrandVoiceID    *string `json:"brandVoiceId,omitempty"`
	TargetKind      *string `json:"targetKind,omitempty"`
	TargetAccountID *string `json:"targetAccountId,omitempty"`
	// WebsiteURL is where published articles are POSTed as a signed webhook.
	// Setting it for the first time issues the secret, returned once.
	WebsiteURL *string `json:"websiteUrl,omitempty"`
}

// UpdatedBlogSettings is the result of UpdateSettings. WebsiteSecret is set
// only when the call set a website URL for the first time; store it then,
// it is never shown again.
type UpdatedBlogSettings struct {
	Settings      BlogSettings `json:"data"`
	WebsiteSecret *string      `json:"websiteSecret"`
}

// BlogWebsiteTest is the outcome of TestWebsite.
type BlogWebsiteTest struct {
	OK     bool    `json:"ok"`
	Status *int    `json:"status"`
	Error  *string `json:"error"`
}

// BlogKeyword is one keyword in the plan.
type BlogKeyword struct {
	ID           string   `json:"id"`
	Keyword      string   `json:"keyword"`
	SearchVolume *int     `json:"searchVolume"`
	Difficulty   *int     `json:"difficulty"`
	CPC          *float64 `json:"cpc"`
	Intent       *string  `json:"intent"`
	Source       string   `json:"source"`
	Status       string   `json:"status"`
	// PlannedFor is a YYYY-MM-DD date.
	PlannedFor *string `json:"plannedFor"`
	// FitReason says why the keyword serves the blog goal.
	FitReason *string `json:"fitReason"`
	CreatedAt Time    `json:"createdAt"`
}

// BlogArticleImage is an image on an article. Position is the index of the
// outline heading it sits under, and nil on the featured image.
type BlogArticleImage struct {
	MediaID  string `json:"mediaId"`
	URL      string `json:"url"`
	Alt      string `json:"alt"`
	Position *int   `json:"position,omitempty"`
}

// BlogArticle is one article as a list returns it.
type BlogArticle struct {
	ID              string            `json:"id"`
	Title           string            `json:"title"`
	Slug            string            `json:"slug"`
	Keyword         *string           `json:"keyword"`
	Excerpt         *string           `json:"excerpt"`
	MetaTitle       *string           `json:"metaTitle"`
	MetaDescription *string           `json:"metaDescription"`
	FeaturedImage   *BlogArticleImage `json:"featuredImage"`
	Tags            []string          `json:"tags"`
	WordCount       int               `json:"wordCount"`
	Status          string            `json:"status"`
	Origin          string            `json:"origin"`
	ReviewUntil     *Time             `json:"reviewUntil"`
	PublishAt       *Time             `json:"publishAt"`
	PublishedAt     *Time             `json:"publishedAt"`
	PublishedURL    *string           `json:"publishedUrl"`
	Error           *string           `json:"error"`
	CreatedAt       Time              `json:"createdAt"`
	UpdatedAt       Time              `json:"updatedAt"`
}

// BlogArticleFull is one article with its body, as GetArticle returns it.
type BlogArticleFull struct {
	BlogArticle
	Instructions *string  `json:"instructions"`
	Outline      []string `json:"outline"`
	// Brief is the research the writer worked from.
	Brief        json.RawMessage    `json:"brief"`
	BodyMarkdown string             `json:"bodyMarkdown"`
	BodyHTML     string             `json:"bodyHtml"`
	Images       []BlogArticleImage `json:"images"`
}

// ListBlogKeywordsParams narrows ListKeywords.
type ListBlogKeywordsParams struct {
	// Status keeps only keywords in these statuses.
	Status []string
	Limit  int
}

// AddBlogKeywordRequest is the body of AddKeyword.
type AddBlogKeywordRequest struct {
	Keyword string `json:"keyword"`
	// PlannedFor is a YYYY-MM-DD date.
	PlannedFor string `json:"plannedFor,omitempty"`
}

// SuggestBlogSeedsRequest is the body of SuggestSeeds.
type SuggestBlogSeedsRequest struct {
	// Seeds are the current terms to refine, at most 20.
	Seeds []string `json:"seeds,omitempty"`
	// Prompt steers the suggestions, at most 500 characters.
	Prompt string `json:"prompt,omitempty"`
}

// UpdateBlogKeywordRequest is the body of UpdateKeyword. Only the fields you
// set are sent. Setting Status to planned picks the keyword; it takes the
// next free date unless PlannedFor is set.
type UpdateBlogKeywordRequest struct {
	Status     string  `json:"status,omitempty"`
	PlannedFor *string `json:"plannedFor,omitempty"`
}

// ListBlogArticlesParams narrows ListArticles.
type ListBlogArticlesParams struct {
	// Status keeps only articles in these statuses.
	Status []string
	Limit  int
}

// RequestBlogArticleRequest is the body of RequestArticle. One of Title,
// Keyword or KeywordID is required.
type RequestBlogArticleRequest struct {
	Title   string `json:"title,omitempty"`
	Keyword string `json:"keyword,omitempty"`
	// KeywordID names a keyword from the plan; its term becomes the target.
	KeywordID    string `json:"keywordId,omitempty"`
	Instructions string `json:"instructions,omitempty"`
}

// UpdateBlogArticleRequest is the body of UpdateArticle. Only the fields you
// set are sent.
type UpdateBlogArticleRequest struct {
	Title           string   `json:"title,omitempty"`
	Slug            string   `json:"slug,omitempty"`
	Excerpt         *string  `json:"excerpt,omitempty"`
	MetaTitle       *string  `json:"metaTitle,omitempty"`
	MetaDescription *string  `json:"metaDescription,omitempty"`
	BodyMarkdown    string   `json:"bodyMarkdown,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	// PublishAt puts the article on the calendar. It must be in review
	// (or already scheduled or failed); a draft answers 409.
	PublishAt *Time `json:"publishAt,omitempty"`
}

func blogWorkspaceQuery(workspaceID string) url.Values {
	q := newQuery()
	q.str("workspaceId", workspaceID)
	return q.values()
}

func blogListQuery(workspaceID string, status []string, limit int) url.Values {
	q := newQuery()
	q.str("workspaceId", workspaceID)
	q.str("status", strings.Join(status, ","))
	q.num("limit", limit)
	return q.values()
}

type blogWorkspaceBody struct {
	WorkspaceID string `json:"workspaceId"`
}

// GetSettings returns the workspace's blog settings.
func (s *BlogService) GetSettings(ctx context.Context, workspaceID string) (*BlogSettings, error) {
	out := &BlogSettings{}
	if err := s.client.json(ctx, "GET", "/blog/settings", nil, blogWorkspaceQuery(workspaceID), out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateSettings patches the blog settings. The returned WebsiteSecret is
// non-nil only when this call set a website URL for the first time.
func (s *BlogService) UpdateSettings(ctx context.Context, workspaceID string, patch *UpdateBlogSettingsRequest) (*UpdatedBlogSettings, error) {
	if patch == nil {
		patch = &UpdateBlogSettingsRequest{}
	}
	body := struct {
		blogWorkspaceBody
		*UpdateBlogSettingsRequest
	}{blogWorkspaceBody{workspaceID}, patch}
	out := &UpdatedBlogSettings{}
	// Do keeps the envelope: websiteSecret sits beside data, not inside it.
	if err := s.client.Do(ctx, "PATCH", "/blog/settings", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// RotateWebsiteSecret issues a new website secret and returns it. The old
// one stops verifying at once.
func (s *BlogService) RotateWebsiteSecret(ctx context.Context, workspaceID string) (string, error) {
	var out struct {
		WebsiteSecret string `json:"websiteSecret"`
	}
	if err := s.client.json(ctx, "POST", "/blog/settings/website-secret", blogWorkspaceBody{workspaceID}, nil, &out); err != nil {
		return "", err
	}
	return out.WebsiteSecret, nil
}

// TestWebsite sends a signed `article.test` delivery to the website URL. A
// website that does not answer 2xx comes back as a 502 *Error.
func (s *BlogService) TestWebsite(ctx context.Context, workspaceID string) (*BlogWebsiteTest, error) {
	out := &BlogWebsiteTest{}
	if err := s.client.json(ctx, "POST", "/blog/settings/website-test", blogWorkspaceBody{workspaceID}, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListKeywords returns the keyword plan.
func (s *BlogService) ListKeywords(ctx context.Context, workspaceID string, params *ListBlogKeywordsParams) ([]BlogKeyword, error) {
	if params == nil {
		params = &ListBlogKeywordsParams{}
	}
	var out []BlogKeyword
	if err := s.client.json(ctx, "GET", "/blog/keywords", nil, blogListQuery(workspaceID, params.Status, params.Limit), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SuggestSeeds suggests about ten search terms from the blog goal and topics.
// Pass the current seeds to refine them and a prompt to steer.
func (s *BlogService) SuggestSeeds(ctx context.Context, workspaceID string, req *SuggestBlogSeedsRequest) ([]string, error) {
	if req == nil {
		req = &SuggestBlogSeedsRequest{}
	}
	body := struct {
		blogWorkspaceBody
		*SuggestBlogSeedsRequest
	}{blogWorkspaceBody{workspaceID}, req}
	var out struct {
		Seeds []string `json:"seeds"`
	}
	if err := s.client.json(ctx, "POST", "/blog/keywords/suggest", body, nil, &out); err != nil {
		return nil, err
	}
	return out.Seeds, nil
}

// PlanKeywords queues keyword research. Nil seeds plans from the settings'
// topic alone. It does not fill the calendar: results land as suggested
// keywords, and UpdateKeyword with status planned picks one.
func (s *BlogService) PlanKeywords(ctx context.Context, workspaceID string, seeds []string) error {
	body := struct {
		blogWorkspaceBody
		Seeds []string `json:"seeds,omitempty"`
	}{blogWorkspaceBody{workspaceID}, seeds}
	return s.client.json(ctx, "POST", "/blog/keywords/plan", body, nil, nil)
}

// AddKeyword adds a keyword to the plan.
func (s *BlogService) AddKeyword(ctx context.Context, workspaceID string, req *AddBlogKeywordRequest) (*BlogKeyword, error) {
	if req == nil {
		req = &AddBlogKeywordRequest{}
	}
	body := struct {
		blogWorkspaceBody
		*AddBlogKeywordRequest
	}{blogWorkspaceBody{workspaceID}, req}
	out := &BlogKeyword{}
	if err := s.client.json(ctx, "POST", "/blog/keywords", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateKeyword changes a keyword's status or planned date.
func (s *BlogService) UpdateKeyword(ctx context.Context, workspaceID, id string, patch *UpdateBlogKeywordRequest) (*BlogKeyword, error) {
	if patch == nil {
		patch = &UpdateBlogKeywordRequest{}
	}
	out := &BlogKeyword{}
	if err := s.client.json(ctx, "PATCH", "/blog/keywords/"+url.PathEscape(id), patch, blogWorkspaceQuery(workspaceID), out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteKeyword removes a keyword from the plan.
func (s *BlogService) DeleteKeyword(ctx context.Context, workspaceID, id string) error {
	return s.client.Do(ctx, "DELETE", "/blog/keywords/"+url.PathEscape(id), nil, blogWorkspaceQuery(workspaceID), nil)
}

// ListArticles returns the workspace's articles.
func (s *BlogService) ListArticles(ctx context.Context, workspaceID string, params *ListBlogArticlesParams) ([]BlogArticle, error) {
	if params == nil {
		params = &ListBlogArticlesParams{}
	}
	var out []BlogArticle
	if err := s.client.json(ctx, "GET", "/blog/articles", nil, blogListQuery(workspaceID, params.Status, params.Limit), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// RequestArticle queues a new article; it comes back `queued` and is written
// in the background.
func (s *BlogService) RequestArticle(ctx context.Context, workspaceID string, req *RequestBlogArticleRequest) (*BlogArticle, error) {
	if req == nil {
		req = &RequestBlogArticleRequest{}
	}
	body := struct {
		blogWorkspaceBody
		*RequestBlogArticleRequest
	}{blogWorkspaceBody{workspaceID}, req}
	out := &BlogArticle{}
	if err := s.client.json(ctx, "POST", "/blog/articles", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetArticle returns one article with its body.
func (s *BlogService) GetArticle(ctx context.Context, workspaceID, id string) (*BlogArticleFull, error) {
	out := &BlogArticleFull{}
	if err := s.client.json(ctx, "GET", "/blog/articles/"+url.PathEscape(id), nil, blogWorkspaceQuery(workspaceID), out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateArticle edits a draft. An article that is being written or is
// already published answers 409 (IsConflict).
func (s *BlogService) UpdateArticle(ctx context.Context, workspaceID, id string, patch *UpdateBlogArticleRequest) (*BlogArticleFull, error) {
	if patch == nil {
		patch = &UpdateBlogArticleRequest{}
	}
	out := &BlogArticleFull{}
	if err := s.client.json(ctx, "PATCH", "/blog/articles/"+url.PathEscape(id), patch, blogWorkspaceQuery(workspaceID), out); err != nil {
		return nil, err
	}
	return out, nil
}

// RegenerateArticle rewrites the article from its brief.
func (s *BlogService) RegenerateArticle(ctx context.Context, workspaceID, id string) (*BlogArticle, error) {
	return s.articleAction(ctx, workspaceID, id, "/regenerate")
}

// SubmitArticle moves a draft, failed or rejected article to in_review.
// Any other status answers 409 (IsConflict).
func (s *BlogService) SubmitArticle(ctx context.Context, workspaceID, id string) (*BlogArticle, error) {
	return s.articleAction(ctx, workspaceID, id, "/submit")
}

// PublishArticle approves the article and queues it for publishing to the
// configured target. It must be in review (or already scheduled or failed);
// a draft answers 409 until SubmitArticle.
func (s *BlogService) PublishArticle(ctx context.Context, workspaceID, id string) (*BlogArticle, error) {
	return s.articleAction(ctx, workspaceID, id, "/publish")
}

// RejectArticle marks an article rejected so it never publishes.
func (s *BlogService) RejectArticle(ctx context.Context, workspaceID, id string) (*BlogArticle, error) {
	return s.articleAction(ctx, workspaceID, id, "/reject")
}

// DeleteArticle removes an article.
func (s *BlogService) DeleteArticle(ctx context.Context, workspaceID, id string) error {
	return s.client.Do(ctx, "DELETE", "/blog/articles/"+url.PathEscape(id), nil, blogWorkspaceQuery(workspaceID), nil)
}

func (s *BlogService) articleAction(ctx context.Context, workspaceID, id, suffix string) (*BlogArticle, error) {
	out := &BlogArticle{}
	if err := s.client.json(ctx, "POST", "/blog/articles/"+url.PathEscape(id)+suffix, nil, blogWorkspaceQuery(workspaceID), out); err != nil {
		return nil, err
	}
	return out, nil
}
