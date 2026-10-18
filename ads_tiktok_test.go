package fopost

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestAdsTikTokReadsUseTheRightPathsAndQuery(t *testing.T) {
	var path, query string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		path, query = r.URL.Path, r.URL.RawQuery
		_, _ = io.WriteString(w, `{"data":[{"id":"item_99","identityId":"idt_1","views":48213}]}`)
	})

	posts, err := client.Ads.SparkPosts(context.Background(), &ListSparkPostsParams{
		WorkspaceID: "ws_1", ConnectionID: "conn_1", AdAccountID: "7011", IdentityID: "idt_1",
	})
	if err != nil {
		t.Fatalf("Ads.SparkPosts: %v", err)
	}
	if path != "/ads/spark-posts" {
		t.Fatalf("path = %q", path)
	}
	if !strings.Contains(query, "identity_id=idt_1") || !strings.Contains(query, "ad_account_id=7011") {
		t.Fatalf("query = %q", query)
	}
	if len(posts) != 1 || posts[0].Views == nil || *posts[0].Views != 48213 {
		t.Fatalf("posts = %+v", posts)
	}
}

func TestAdsTikTokBusinessCentersAndIdentities(t *testing.T) {
	var path string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		if path == "/ads/tiktok/business-centers" {
			_, _ = io.WriteString(w, `{"data":[{"id":"bc1","name":"Brand HQ","role":"ADMIN"}]}`)
			return
		}
		_, _ = io.WriteString(w, `{"data":[{"id":"idt_1","type":"CUSTOMIZED_USER","name":"Your Brand"}]}`)
	})

	centers, err := client.Ads.TikTokBusinessCenters(context.Background(), &AdObjectParams{
		WorkspaceID: "ws_1", ConnectionID: "conn_1",
	})
	if err != nil || len(centers) != 1 || centers[0].Name != "Brand HQ" {
		t.Fatalf("centers = %+v, err = %v", centers, err)
	}

	identities, err := client.Ads.TikTokIdentities(context.Background(), &ListAudiencesParams{
		WorkspaceID: "ws_1", ConnectionID: "conn_1", AdAccountID: "7011",
	})
	if err != nil || len(identities) != 1 || identities[0].Type != "CUSTOMIZED_USER" {
		t.Fatalf("identities = %+v, err = %v", identities, err)
	}
	if path != "/ads/tiktok/identities" {
		t.Fatalf("path = %q", path)
	}
}

func TestAdsSparkPostIDAndSmartPlusTravelInTheBody(t *testing.T) {
	var body map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = io.WriteString(w, `{"data":{"id":"ad_1"}}`)
	})

	if _, err := client.Ads.Create(context.Background(), &CreateAdRequest{
		WorkspaceID: "ws_1", ConnectionID: "conn_1", AdAccountID: "7011", PageID: "idt_1",
		Name: "Spark", Goal: AdGoalTraffic, SparkPostID: "item_99",
	}); err != nil {
		t.Fatalf("Ads.Create: %v", err)
	}
	if body["sparkPostId"] != "item_99" {
		t.Fatalf("body = %+v", body)
	}

	if _, err := client.Ads.CreateCampaign(context.Background(), &CreateCampaignRequest{
		WorkspaceID: "ws_1", ConnectionID: "conn_1", AdAccountID: "7011",
		Name: "Smart", Goal: AdGoalTraffic, SmartPlus: Bool(true),
	}); err != nil {
		t.Fatalf("Ads.CreateCampaign: %v", err)
	}
	if body["smartPlus"] != true {
		t.Fatalf("body = %+v", body)
	}
}

func TestAdsConversionsReportWhatTheNetworkAccepted(t *testing.T) {
	var body map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = io.WriteString(w, `{"data":{"accepted":2}}`)
	})

	accepted, err := client.Ads.UploadConversions(context.Background(), &UploadConversionsRequest{
		WorkspaceID: "ws_1", ConnectionID: "conn_1", AdAccountID: "7011", PixelID: "px_1",
		Events: []ConversionEvent{{EventName: "CompletePayment", OccurredAt: "2026-09-18T10:04:00Z"}},
	})
	if err != nil || accepted != 2 {
		t.Fatalf("accepted = %d, err = %v", accepted, err)
	}
	if body["pixelId"] != "px_1" {
		t.Fatalf("body = %+v", body)
	}
}

func TestAdsCommentsPageAndTheThreeWrites(t *testing.T) {
	var method, path string
	var body map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		if r.Method != http.MethodGet {
			_ = json.NewDecoder(r.Body).Decode(&body)
		}
		if r.Method == http.MethodGet {
			_, _ = io.WriteString(w, `{"data":{"comments":[{"id":"cm_1","text":"nice","likes":3,"hidden":true}],"nextCursor":"2"}}`)
			return
		}
		_, _ = io.WriteString(w, `{"data":{"replyId":"cm_2"}}`)
	})

	page, err := client.Ads.Comments(context.Background(), &ListAdCommentsParams{
		WorkspaceID: "ws_1", ConnectionID: "conn_1", AdID: "ad_1",
	})
	if err != nil || page.NextCursor != "2" || !page.Comments[0].Hidden || page.Comments[0].Likes != 3 {
		t.Fatalf("page = %+v, err = %v", page, err)
	}

	scope := &AdCommentRequest{WorkspaceID: "ws_1", ConnectionID: "conn_1", AdID: "ad_1"}

	replyID, err := client.Ads.ReplyToComment(context.Background(), "cm_1",
		&AdCommentRequest{WorkspaceID: "ws_1", ConnectionID: "conn_1", AdID: "ad_1", Text: "Friday!"})
	if err != nil || replyID != "cm_2" {
		t.Fatalf("replyID = %q, err = %v", replyID, err)
	}
	if path != "/ads/comments/cm_1/reply" || body["text"] != "Friday!" {
		t.Fatalf("path = %q, body = %+v", path, body)
	}

	if err := client.Ads.SetCommentHidden(context.Background(), "cm_1",
		&AdCommentRequest{WorkspaceID: "ws_1", ConnectionID: "conn_1", AdID: "ad_1", Hidden: Bool(true)}); err != nil {
		t.Fatalf("Ads.SetCommentHidden: %v", err)
	}
	if path != "/ads/comments/cm_1/hide" || body["hidden"] != true {
		t.Fatalf("path = %q, body = %+v", path, body)
	}

	if err := client.Ads.DeleteComment(context.Background(), "cm_1", scope); err != nil {
		t.Fatalf("Ads.DeleteComment: %v", err)
	}
	// The ad travels in the body, because the path already carries the comment.
	if method != http.MethodDelete || path != "/ads/comments/cm_1" || body["adId"] != "ad_1" {
		t.Fatalf("method = %q, path = %q, body = %+v", method, path, body)
	}
}

func TestAdsSendConversionsCarriesTheIdempotencyKey(t *testing.T) {
	var path, key string
	var body map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		path, key = r.URL.Path, r.Header.Get("Idempotency-Key")
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Idempotency-Replayed", "true")
		w.WriteHeader(http.StatusAccepted)
		_, _ = io.WriteString(w, `{"data":{"id":"cb_1","accepted":1,"eventCount":1,"replayed":true}}`)
	})

	result, err := client.Ads.SendConversions(context.Background(), "order-42", &UploadConversionsRequest{
		WorkspaceID: "ws_1", ConnectionID: "conn_1", AdAccountID: "act_1", PixelID: "px_1",
		Events: []ConversionEvent{{
			EventName: "Purchase", OccurredAt: "2026-10-18T10:04:00Z", EventID: "evt_1",
			ActionSource: ConversionActionSourceWebsite, Email: "buyer@example.test", ContentIDs: []string{"sku_1"},
		}},
	})
	if err != nil || result.ID != "cb_1" || !result.Replayed || result.Accepted != 1 || result.EventCount != 1 {
		t.Fatalf("result = %+v, err = %v", result, err)
	}
	if path != "/ads/conversions/events" || key != "order-42" {
		t.Fatalf("path = %s, Idempotency-Key = %q", path, key)
	}
	event := body["events"].([]any)[0].(map[string]any)
	if event["eventId"] != "evt_1" || event["actionSource"] != "website" || event["email"] != "buyer@example.test" {
		t.Fatalf("event = %+v", event)
	}
}

func TestAdsPixelReadsUseTheRightPathAndQuery(t *testing.T) {
	var path, query string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		path, query = r.URL.Path, r.URL.RawQuery
		_, _ = io.WriteString(w, `{"data":{"id":"px_1","name":"Site","code":"<script></script>","createdAt":null,"lastFiredAt":"2026-10-17T00:00:00Z","isUnavailable":false,"dataUseSetting":"analytics_only"}}`)
	})

	pixel, err := client.Ads.GetPixel(context.Background(), "px_1", &PixelParams{
		WorkspaceID: "ws_1", ConnectionID: "conn_1", AdAccountID: "act_1",
	})
	if err != nil || pixel.ID != "px_1" || pixel.Code == nil || *pixel.Code != "<script></script>" || pixel.CreatedAt != nil {
		t.Fatalf("pixel = %+v, err = %v", pixel, err)
	}
	if pixel.DataUseSetting == nil || *pixel.DataUseSetting != PixelDataUseAnalyticsOnly {
		t.Fatalf("dataUseSetting = %v", pixel.DataUseSetting)
	}
	if path != "/ads/pixels/px_1" || !strings.Contains(query, "ad_account_id=act_1") || !strings.Contains(query, "connection_id=conn_1") {
		t.Fatalf("path = %s, query = %s", path, query)
	}
}

func TestAdsCreateCustomConversionSendsTheRule(t *testing.T) {
	var method, path string
	var body map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"data":{"id":"cc_1","name":"Checkout","pixelId":"px_1","eventType":"PURCHASE","rule":{"eventName":"PageView","urlContains":"/thanks"},"description":null,"defaultValueMinor":1999,"isArchived":false,"isUnavailable":false,"createdAt":null,"lastFiredAt":null}}`)
	})

	cc, err := client.Ads.CreateCustomConversion(context.Background(), &CreateCustomConversionRequest{
		WorkspaceID: "ws_1", ConnectionID: "conn_1", AdAccountID: "act_1", PixelID: "px_1",
		Name: "Checkout", EventType: CustomConversionPurchase,
		Rule: CustomConversionRule{URLContains: "/thanks"}, DefaultValueMinor: Int(1999),
	})
	if err != nil || cc.ID != "cc_1" || cc.Rule == nil || cc.Rule.URLContains != "/thanks" || cc.DefaultValueMinor == nil || *cc.DefaultValueMinor != 1999 {
		t.Fatalf("cc = %+v, err = %v", cc, err)
	}
	if method != http.MethodPost || path != "/ads/custom-conversions" {
		t.Fatalf("%s %s", method, path)
	}
	rule := body["rule"].(map[string]any)
	if body["eventType"] != "PURCHASE" || rule["urlContains"] != "/thanks" || rule["eventName"] != nil {
		t.Fatalf("body = %+v", body)
	}
}
