package fopost

import (
	"context"
	"io"
	"net/http"
	"testing"
)

func TestAnalyticsDecayDecodesBandsAndHalfLife(t *testing.T) {
	var query string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		_, _ = io.WriteString(w, `{"data":{"days":30,"postsMeasured":2,"halfLifeBucket":"1h_3h","bands":[
			{"bucket":"under_1h","label":"First hour","posts":2,"avgEngagements":25,"avgImpressions":300,"shareOfFinal":0.3},
			{"bucket":"6h_12h","label":"6-12 hours","posts":0,"avgEngagements":0,"avgImpressions":0,"shareOfFinal":null}
		]}}`)
	})

	decay, err := client.Analytics.Decay(context.Background(), &AnalyticsParams{Days: 30, AccountID: "acc_1"})
	if err != nil {
		t.Fatalf("Analytics.Decay: %v", err)
	}
	if query != "accountId=acc_1&days=30" {
		t.Fatalf("query = %q", query)
	}
	if decay.HalfLifeBucket != "1h_3h" || decay.PostsMeasured != 2 {
		t.Fatalf("decay = %+v", decay)
	}
	if decay.Bands[0].ShareOfFinal == nil || *decay.Bands[0].ShareOfFinal != 0.3 {
		t.Fatalf("share of final = %+v", decay.Bands[0].ShareOfFinal)
	}
	// A band nothing was measured in reports no share rather than zero
	if decay.Bands[1].ShareOfFinal != nil {
		t.Fatalf("empty band reported a share: %+v", decay.Bands[1])
	}
}

func TestAnalyticsFrequencyDecodesWeeksAndBest(t *testing.T) {
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"data":{"days":90,
			"weeks":[{"weekStart":"2026-03-02","posts":2,"engagements":240,"avgEngagementsPerPost":120}],
			"bands":[{"band":"under_3","label":"1-2 a week","weeks":1,"posts":2,"avgPostsPerWeek":2,"avgEngagementsPerPost":120,"engagementRate":0.12}],
			"best":{"band":"under_3","label":"1-2 a week","avgEngagementsPerPost":120}}}`)
	})

	frequency, err := client.Analytics.Frequency(context.Background(), &AnalyticsParams{Days: 90})
	if err != nil {
		t.Fatalf("Analytics.Frequency: %v", err)
	}
	if frequency.Weeks[0].WeekStart != "2026-03-02" {
		t.Fatalf("weeks = %+v", frequency.Weeks)
	}
	if frequency.Best == nil || frequency.Best.Label != "1-2 a week" {
		t.Fatalf("best = %+v", frequency.Best)
	}
	if frequency.Bands[0].EngagementRate == nil || *frequency.Bands[0].EngagementRate != 0.12 {
		t.Fatalf("engagement rate = %+v", frequency.Bands[0].EngagementRate)
	}
}

func TestAnalyticsTimelineEscapesAPermalinkIntoThePath(t *testing.T) {
	var path string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.EscapedPath()
		_, _ = io.WriteString(w, `{"data":{"postId":null,"deliveries":[{"accountId":"acc_1","platform":"twitter","username":"acme","externalPostId":"1","postedAt":"2026-03-02T00:00:00.000Z","points":[{"at":"2026-03-02T00:30:00.000Z","ageMinutes":30,"engagements":40,"impressions":400,"reach":null,"likes":30,"comments":null,"shares":null,"videoViews":null,"delta":{"impressions":400,"reach":0,"engagements":40,"likes":30,"comments":0,"shares":0}}]}]}}`)
	})

	timeline, err := client.Analytics.Timeline(context.Background(), "https://x.com/acme/status/1")
	if err != nil {
		t.Fatalf("Analytics.Timeline: %v", err)
	}
	if path != "/analytics/posts/https:%2F%2Fx.com%2Facme%2Fstatus%2F1/timeline" {
		t.Fatalf("path = %q", path)
	}
	if timeline.PostID != "" {
		t.Fatalf("a native post reported a FoPost id: %q", timeline.PostID)
	}
	point := timeline.Deliveries[0].Points[0]
	if point.AgeMinutes == nil || *point.AgeMinutes != 30 || point.Delta.Engagements != 40 {
		t.Fatalf("point = %+v", point)
	}
}

func TestAnalyticsChangesCarriesTheCursor(t *testing.T) {
	var query string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		_, _ = io.WriteString(w, `{"data":{"since":"2026-03-02T00:00:00.000Z","cursor":"2026-03-02T06:00:00.000Z","hasMore":true,"changes":[{"accountId":"acc_1","platform":"twitter","externalPostId":"1","postId":"post_1","postedAt":"2026-03-02T00:00:00.000Z","fetchedAt":"2026-03-02T06:00:00.000Z","impressions":900,"reach":null,"engagements":90,"likes":70,"comments":10,"shares":10}]}}`)
	})

	page, err := client.Analytics.Changes(context.Background(), &MetricChangesParams{
		Since: "2026-03-02T00:00:00Z",
		Limit: 100,
	})
	if err != nil {
		t.Fatalf("Analytics.Changes: %v", err)
	}
	if query != "limit=100&since=2026-03-02T00%3A00%3A00Z" {
		t.Fatalf("query = %q", query)
	}
	if !page.HasMore || page.Changes[0].PostID != "post_1" {
		t.Fatalf("page = %+v", page)
	}
}

func TestAnalyticsCollectPostReportsEachDelivery(t *testing.T) {
	var method, path string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.EscapedPath()
		_, _ = io.WriteString(w, `{"data":{"collected":1,"deliveries":[{"accountId":"acc_1","platform":"twitter","externalPostId":"1","collected":true,"fetchedAt":"2026-03-02T00:30:00.000Z","message":null}]}}`)
	})

	result, err := client.Analytics.CollectPost(context.Background(), "post_1")
	if err != nil {
		t.Fatalf("Analytics.CollectPost: %v", err)
	}
	if method != "POST" || path != "/posts/post_1/analytics/collect" {
		t.Fatalf("%s %s", method, path)
	}
	if result.Collected != 1 || !result.Deliveries[0].Collected {
		t.Fatalf("result = %+v", result)
	}
}

func TestAnalyticsNativePostsKeepsTheMetaEnvelope(t *testing.T) {
	var query string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		_, _ = io.WriteString(w, `{"data":[{"externalPostId":"1","text":"Posted by hand","permalink":"https://x.com/acme/status/1","thumbnailUrl":null,"mediaType":null,"postedAt":"2026-03-02T00:00:00.000Z","fetchedAt":"2026-03-02T06:00:00.000Z","metrics":{"impressions":900,"reach":null,"engagements":90,"likes":70,"comments":10,"shares":10,"videoViews":null}}],"meta":{"page":1,"perPage":20,"total":1}}`)
	})

	list, err := client.Analytics.NativePosts(context.Background(), "acc_1", &NativePostsParams{Page: 1, PerPage: 20})
	if err != nil {
		t.Fatalf("Analytics.NativePosts: %v", err)
	}
	if query != "page=1&per_page=20" {
		t.Fatalf("query = %q", query)
	}
	if len(list.Data) != 1 || list.Meta.Total != 1 {
		t.Fatalf("list = %+v", list)
	}
	if list.Data[0].Permalink != "https://x.com/acme/status/1" {
		t.Fatalf("permalink = %q", list.Data[0].Permalink)
	}
	if list.Data[0].Metrics.Engagements == nil || *list.Data[0].Metrics.Engagements != 90 {
		t.Fatalf("metrics = %+v", list.Data[0].Metrics)
	}
}
