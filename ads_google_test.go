package fopost

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

var googleScope = GoogleScope{WorkspaceID: "ws_1", ConnectionID: "conn_1", CustomerID: "1234567890"}

func TestGoogleBidStrategiesNameTheConnectionAndTheCustomer(t *testing.T) {
	var path, query string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		path, query = r.URL.Path, r.URL.Query().Encode()
		_, _ = io.WriteString(w, `{"data":[{"id":"1234567890~biddingStrategy~42","name":"Target CPA","type":"TARGET_CPA","status":"ENABLED","campaignCount":3}]}`)
	})

	strategies, err := client.GoogleAds.BidStrategies(context.Background(), googleScope)
	if err != nil {
		t.Fatalf("GoogleAds.BidStrategies: %v", err)
	}
	if len(strategies) != 1 || strategies[0].CampaignCount != 3 {
		t.Fatalf("strategies = %+v", strategies)
	}
	if path != "/ads/google/bid-strategies" {
		t.Fatalf("path = %q", path)
	}
	want := "connection_id=conn_1&customer_id=1234567890&workspace_id=ws_1"
	if query != want {
		t.Fatalf("query = %q", query)
	}
}

func TestGoogleCreateBidStrategySendsTheScopeInTheBody(t *testing.T) {
	var body map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = io.WriteString(w, `{"data":{"id":"1234567890~biddingStrategy~42"}}`)
	})

	id, err := client.GoogleAds.CreateBidStrategy(context.Background(), &CreateGoogleBidStrategyRequest{
		GoogleScope: googleScope,
		Name:        "Target CPA",
		Type:        GoogleBidTargetCPA,
		TargetMinor: Int(1500),
	})
	if err != nil {
		t.Fatalf("GoogleAds.CreateBidStrategy: %v", err)
	}
	if id != "1234567890~biddingStrategy~42" {
		t.Fatalf("id = %q", id)
	}
	if body["customerId"] != "1234567890" || body["type"] != "TARGET_CPA" {
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
