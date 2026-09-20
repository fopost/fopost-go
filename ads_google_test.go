package fopost

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

var googleScope = GoogleScope{WorkspaceID: "ws_1", ConnectionID: "conn_1", CustomerID: "1234567890"}

func TestGoogleKeywordsNameTheConnectionAndTheCustomer(t *testing.T) {
	var path, query string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		path, query = r.URL.Path, r.URL.Query().Encode()
		_, _ = io.WriteString(w, `{"data":[{"id":"1234567890~keyword~77~99","adGroupId":"1234567890~adGroup~77","text":"running shoes","matchType":"EXACT","status":"ENABLED","cpcBidMinor":180,"negative":false}]}`)
	})

	keywords, err := client.GoogleAds.Keywords(context.Background(), googleScope, &ListGoogleKeywordsParams{
		AdGroupID: "1234567890~adGroup~77",
	})
	if err != nil {
		t.Fatalf("GoogleAds.Keywords: %v", err)
	}
	if len(keywords) != 1 || keywords[0].Text != "running shoes" {
		t.Fatalf("keywords = %+v", keywords)
	}
	if keywords[0].CPCBidMinor == nil || *keywords[0].CPCBidMinor != 180 {
		t.Fatalf("cpcBidMinor = %+v", keywords[0].CPCBidMinor)
	}
	if path != "/ads/google/keywords" {
		t.Fatalf("path = %q", path)
	}
	want := "ad_group_id=1234567890~adGroup~77&connection_id=conn_1&customer_id=1234567890&workspace_id=ws_1"
	if query != want {
		t.Fatalf("query = %q", query)
	}
}

func TestGoogleCreateKeywordSendsTheScopeInTheBody(t *testing.T) {
	var body map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = io.WriteString(w, `{"data":{"id":"1234567890~keyword~77~99"}}`)
	})

	id, err := client.GoogleAds.CreateKeyword(context.Background(), &CreateGoogleKeywordRequest{
		GoogleScope: googleScope,
		AdGroupID:   "1234567890~adGroup~77",
		Text:        "running shoes",
		MatchType:   GoogleMatchExact,
	})
	if err != nil {
		t.Fatalf("GoogleAds.CreateKeyword: %v", err)
	}
	if id != "1234567890~keyword~77~99" {
		t.Fatalf("id = %q", id)
	}
	if body["customerId"] != "1234567890" || body["matchType"] != "EXACT" {
		t.Fatalf("body = %+v", body)
	}
}

func TestGoogleDeleteCarriesTheScopeInTheBody(t *testing.T) {
	var method string
	var body map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.WriteHeader(http.StatusNoContent)
	})

	if err := client.GoogleAds.DeleteAsset(context.Background(), "1234567890~asset~4321", googleScope); err != nil {
		t.Fatalf("GoogleAds.DeleteAsset: %v", err)
	}
	if method != http.MethodDelete {
		t.Fatalf("method = %q", method)
	}
	if body["connectionId"] != "conn_1" || body["customerId"] != "1234567890" {
		t.Fatalf("body = %+v", body)
	}
}

func TestGoogleAdScheduleIsReplacedWithPut(t *testing.T) {
	var method string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		_, _ = io.WriteString(w, `{"data":{"slots":2}}`)
	})

	slots, err := client.GoogleAds.SetAdSchedule(context.Background(), &SetGoogleAdScheduleRequest{
		GoogleScope: googleScope,
		CampaignID:  "1234567890~campaign~55",
		Slots:       []GoogleAdScheduleInput{{DayOfWeek: "MONDAY", StartHour: 9, EndHour: 18}},
	})
	if err != nil {
		t.Fatalf("GoogleAds.SetAdSchedule: %v", err)
	}
	if slots != 2 || method != http.MethodPut {
		t.Fatalf("slots = %d, method = %q", slots, method)
	}
}

func TestGoogleQueryReturnsRowsAsGoogleSendsThem(t *testing.T) {
	var path string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		_, _ = io.WriteString(w, `{"data":{"rows":[{"campaign":{"id":"55"}}]}}`)
	})

	rows, err := client.GoogleAds.Query(context.Background(), &GoogleQueryRequest{
		GoogleScope: googleScope,
		Query:       "SELECT campaign.id FROM campaign",
	})
	if err != nil {
		t.Fatalf("GoogleAds.Query: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("rows = %+v", rows)
	}
	if path != "/ads/insights/query" {
		t.Fatalf("path = %q", path)
	}
}

func TestAuthorizeGoogleHasItsOwnRoute(t *testing.T) {
	var path string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		_, _ = io.WriteString(w, `{"data":{"url":"https://accounts.google.com/o/x"}}`)
	})

	url, err := client.Ads.AuthorizeGoogle(context.Background(), &AuthorizeGoogleAdsRequest{WorkspaceID: "ws_1"})
	if err != nil {
		t.Fatalf("Ads.AuthorizeGoogle: %v", err)
	}
	if url != "https://accounts.google.com/o/x" || path != "/ads/connections/google/authorize" {
		t.Fatalf("url = %q, path = %q", url, path)
	}
}
