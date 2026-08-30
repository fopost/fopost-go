package fopost

import (
	"context"
	"net/url"
)

// AnalyticsService covers the cross-account reporting surface.
type AnalyticsService struct{ client *Client }

// AnalyticsParams is the window and scope shared by most analytics calls.
// Days and an explicit From/To range are alternatives; zero fields leave the
// API's defaults in place.
type AnalyticsParams struct {
	AccountID   string
	WorkspaceID string
	Days        int
	// From and To are YYYY-MM-DD.
	From string
	To   string
	// Limit caps the rows returned, where the endpoint supports it.
	Limit int
	// Sort is "recent" on top-posts, to order by date instead of performance.
	Sort string
	// Label narrows top-posts to one campaign label.
	Label string
	// Page and PerPage paginate the posts table.
	Page int
}

func (p *AnalyticsParams) values() url.Values {
	q := newQuery()
	if p == nil {
		return q.values()
	}
	q.str("accountId", p.AccountID)
	q.str("workspace_id", p.WorkspaceID)
	q.num("days", p.Days)
	q.str("from", p.From)
	q.str("to", p.To)
	q.num("limit", p.Limit)
	q.str("sort", p.Sort)
	q.str("label", p.Label)
	q.num("page", p.Page)
	return q.values()
}

// AnalyticsDeltas are period-over-period changes, as fractions. A nil field
// means there was no earlier period to compare against.
type AnalyticsDeltas struct {
	Followers    *float64 `json:"followers"`
	Posts        *float64 `json:"posts"`
	Engagement   *float64 `json:"engagement"`
	Impressions  *float64 `json:"impressions"`
	Likes        *float64 `json:"likes"`
	Comments     *float64 `json:"comments"`
	Shares       *float64 `json:"shares"`
	ProfileViews *float64 `json:"profileViews"`
}

// AnalyticsOverview is the headline roll-up across every account in scope.
type AnalyticsOverview struct {
	TotalAccounts     int             `json:"totalAccounts"`
	TotalFollowers    int             `json:"totalFollowers"`
	TotalPosts        int             `json:"totalPosts"`
	TotalEngagement   int             `json:"totalEngagement"`
	TotalImpressions  int             `json:"totalImpressions"`
	TotalReach        int             `json:"totalReach"`
	TotalLikes        int             `json:"totalLikes"`
	TotalComments     int             `json:"totalComments"`
	TotalShares       int             `json:"totalShares"`
	TotalReposts      int             `json:"totalReposts"`
	TotalSaves        int             `json:"totalSaves"`
	TotalClicks       int             `json:"totalClicks"`
	TotalVideoViews   int             `json:"totalVideoViews"`
	TotalProfileViews int             `json:"totalProfileViews"`
	EngagementRate    *float64        `json:"engagementRate"`
	Deltas            AnalyticsDeltas `json:"deltas"`
	TodayStats        struct {
		Posts          int `json:"posts"`
		FollowerChange int `json:"followerChange"`
		Engagement     int `json:"engagement"`
	} `json:"todayStats"`
	Platforms []struct {
		Platform  string `json:"platform"`
		Accounts  int    `json:"accounts"`
		Followers int    `json:"followers"`
	} `json:"platforms"`
	Accounts []struct {
		AccountID  string `json:"accountId"`
		Platform   string `json:"platform"`
		Username   string `json:"username"`
		Name       string `json:"name"`
		Avatar     string `json:"avatar"`
		Followers  *int   `json:"followers"`
		TotalPosts *int   `json:"totalPosts"`
		FetchedAt  Time   `json:"fetchedAt"`
		History    []struct {
			Date      string `json:"date"`
			Followers *int   `json:"followers"`
		} `json:"history"`
	} `json:"accounts"`
}

// Overview returns the headline numbers for the window.
func (s *AnalyticsService) Overview(ctx context.Context, params *AnalyticsParams) (*AnalyticsOverview, error) {
	out := &AnalyticsOverview{}
	if err := s.client.json(ctx, "GET", "/analytics/overview", nil, params.values(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// TimeSeriesPoint is one day of activity.
type TimeSeriesPoint struct {
	Date        string `json:"date"`
	Engagements int    `json:"engagements"`
	Impressions int    `json:"impressions"`
	Likes       int    `json:"likes"`
	Comments    int    `json:"comments"`
	Shares      int    `json:"shares"`
	Followers   int    `json:"followers"`
	Posts       int    `json:"posts"`
}

// TimeSeries is daily activity over the window.
type TimeSeries struct {
	Days   int               `json:"days"`
	Series []TimeSeriesPoint `json:"series"`
}

// TimeSeries returns one point per day in the window.
func (s *AnalyticsService) TimeSeries(ctx context.Context, params *AnalyticsParams) (*TimeSeries, error) {
	out := &TimeSeries{}
	if err := s.client.json(ctx, "GET", "/analytics/time-series", nil, params.values(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// TopPost is one high-performing post. Source is "fopost" for a post published
// from here, or "platform" for one found on the account.
type TopPost struct {
	Rank           int    `json:"rank"`
	PostID         string `json:"postId"`
	ExternalPostID string `json:"externalPostId"`
	Source         string `json:"source"`
	Preview        string `json:"preview"`
	Permalink      string `json:"permalink"`
	ThumbnailURL   string `json:"thumbnailUrl"`
	Status         string `json:"status"`
	CreatedAt      Time   `json:"createdAt"`
	Platforms      []struct {
		Platform string `json:"platform"`
		Username string `json:"username"`
		URL      string `json:"url"`
	} `json:"platforms"`
	Labels  []PostLabelRef `json:"labels"`
	Metrics struct {
		Engagements int `json:"engagements"`
		Impressions int `json:"impressions"`
		Reach       int `json:"reach"`
		Likes       int `json:"likes"`
		Comments    int `json:"comments"`
		Shares      int `json:"shares"`
		Reposts     int `json:"reposts"`
		Clicks      int `json:"clicks"`
		Saves       int `json:"saves"`
		VideoViews  int `json:"videoViews"`
	} `json:"metrics"`
}

// TopPosts returns the best performing posts in the window.
func (s *AnalyticsService) TopPosts(ctx context.Context, params *AnalyticsParams) ([]TopPost, error) {
	var out []TopPost
	if err := s.client.json(ctx, "GET", "/analytics/top-posts", nil, params.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// LabelAnalytics is one campaign label's performance.
type LabelAnalytics struct {
	LabelID        string   `json:"labelId"`
	Name           string   `json:"name"`
	Color          string   `json:"color"`
	PostCount      int      `json:"postCount"`
	Impressions    int      `json:"impressions"`
	Reach          int      `json:"reach"`
	Engagements    int      `json:"engagements"`
	Likes          int      `json:"likes"`
	Comments       int      `json:"comments"`
	Shares         int      `json:"shares"`
	EngagementRate *float64 `json:"engagementRate"`
	FollowerDelta  *int     `json:"followerDelta"`
}

// Labels returns a per-label campaign roll-up.
func (s *AnalyticsService) Labels(ctx context.Context, params *AnalyticsParams) ([]LabelAnalytics, error) {
	var out []LabelAnalytics
	if err := s.client.json(ctx, "GET", "/analytics/labels", nil, params.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// PostsTable is a page of posts with their per-platform delivery breakdown.
type PostsTable struct {
	Posts []struct {
		PostID      string `json:"postId"`
		Preview     string `json:"preview"`
		Status      string `json:"status"`
		CreatedAt   Time   `json:"createdAt"`
		ScheduledAt Time   `json:"scheduledAt"`
		Platforms   []struct {
			Platform       string `json:"platform"`
			Username       string `json:"username"`
			URL            string `json:"url"`
			DeliveryStatus string `json:"deliveryStatus"`
		} `json:"platforms"`
		DeliverySummary struct {
			Total     int `json:"total"`
			Published int `json:"published"`
			Failed    int `json:"failed"`
			Pending   int `json:"pending"`
		} `json:"deliverySummary"`
	} `json:"posts"`
	Total         int `json:"total"`
	Page          int `json:"page"`
	Limit         int `json:"limit"`
	StatusSummary struct {
		Draft     int `json:"draft"`
		Scheduled int `json:"scheduled"`
		Published int `json:"published"`
		Failed    int `json:"failed"`
		Pending   int `json:"pending"`
		Total     int `json:"total"`
	} `json:"statusSummary"`
}

// PostsTable returns posts with their delivery breakdown, paginated.
func (s *AnalyticsService) PostsTable(ctx context.Context, params *AnalyticsParams) (*PostsTable, error) {
	out := &PostsTable{}
	if err := s.client.json(ctx, "GET", "/analytics/posts-table", nil, params.values(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// StreakDay is one day of posting activity.
type StreakDay struct {
	Date           string `json:"date"`
	Count          int    `json:"count"`
	PublishedCount int    `json:"publishedCount"`
	FailedCount    int    `json:"failedCount"`
	ScheduledCount int    `json:"scheduledCount"`
}

// PostingStreak returns 365 days of posting activity.
func (s *AnalyticsService) PostingStreak(ctx context.Context, workspaceID string) ([]StreakDay, error) {
	q := newQuery()
	q.str("workspace_id", workspaceID)
	var out struct {
		Streak []StreakDay `json:"streak"`
	}
	if err := s.client.json(ctx, "GET", "/analytics/posting-streak", nil, q.values(), &out); err != nil {
		return nil, err
	}
	return out.Streak, nil
}

// Audiences a demographics breakdown can describe.
const (
	AudienceFollowers = "followers"
	AudienceEngaged   = "engaged"
	AudienceReached   = "reached"
)

// DemographicsBucket is one slice of an audience: a value and its share.
type DemographicsBucket struct {
	Key   string  `json:"key"`
	Value float64 `json:"value"`
	Share float64 `json:"share"`
}

// AccountRef names an account in a demographics response.
type AccountRef struct {
	AccountID string `json:"accountId"`
	Platform  string `json:"platform"`
	Username  string `json:"username"`
}

// Demographics is an audience breakdown. UnsupportedAccounts names the
// accounts whose platform does not report demographics.
type Demographics struct {
	Audience   string `json:"audience"`
	Dimensions struct {
		Age     []DemographicsBucket `json:"age"`
		Gender  []DemographicsBucket `json:"gender"`
		Country []DemographicsBucket `json:"country"`
		City    []DemographicsBucket `json:"city"`
	} `json:"dimensions"`
	ContributingAccounts []AccountRef `json:"contributingAccounts"`
	UnsupportedAccounts  []AccountRef `json:"unsupportedAccounts"`
}

// Demographics returns an audience breakdown. audience is one of the Audience
// constants; empty means followers.
func (s *AnalyticsService) Demographics(ctx context.Context, audience string, params *AnalyticsParams) (*Demographics, error) {
	values := params.values()
	if audience != "" {
		values.Set("audience", audience)
	}
	out := &Demographics{}
	if err := s.client.json(ctx, "GET", "/analytics/demographics", nil, values, out); err != nil {
		return nil, err
	}
	return out, nil
}

// CollectSummary reports what a collection run refreshed.
type CollectSummary struct {
	Accounts     int `json:"accounts"`
	Posts        int `json:"posts"`
	Demographics int `json:"demographics"`
	Errors       int `json:"errors"`
	ErrorDetails []struct {
		AccountID string `json:"accountId"`
		Platform  string `json:"platform"`
		Username  string `json:"username"`
		// Stage is "account", "demographics", "timeline", or "post".
		Stage   string `json:"stage"`
		Message string `json:"message"`
	} `json:"errorDetails"`
}

// Collect pulls fresh numbers from the platforms. It is rate limited harder
// than the read endpoints, since every call reaches out to a network.
func (s *AnalyticsService) Collect(ctx context.Context, accountID string) (*CollectSummary, error) {
	q := newQuery()
	q.str("accountId", accountID)
	out := &CollectSummary{}
	if err := s.client.json(ctx, "POST", "/analytics/collect", nil, q.values(), out); err != nil {
		return nil, err
	}
	return out, nil
}
