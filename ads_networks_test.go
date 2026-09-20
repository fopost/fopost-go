package fopost

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestAdsAuthorizeReachesWhicheverNetworkTheRegistryNamed(t *testing.T) {
	var path string
	var body map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = io.WriteString(w, `{"data":{"url":"https://www.linkedin.com/oauth"}}`)
	})

	url, err := client.Ads.Authorize(context.Background(), "linkedin", &AuthorizeAdsRequest{
		WorkspaceID: "ws_1",
		ReturnTo:    "/ads",
	})
	if err != nil {
		t.Fatalf("Ads.Authorize: %v", err)
	}
	if path != "/ads/connections/linkedin/authorize" {
		t.Fatalf("path = %s", path)
	}
	if body["workspaceId"] != "ws_1" || body["returnTo"] != "/ads" {
		t.Fatalf("body = %v", body)
	}
	if url != "https://www.linkedin.com/oauth" {
		t.Fatalf("url = %s", url)
	}
}

func TestAdsProvidersCarryWhatEachNetworkSupports(t *testing.T) {
	client, _ := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"data":[{"id":"linkedin","name":"LinkedIn Ads","configured":false,
			"capabilities":{"conversions":true},"targetingFacets":["country","job_title"]}]}`)
	})

	providers, err := client.Ads.Providers(context.Background())
	if err != nil {
		t.Fatalf("Ads.Providers: %v", err)
	}
	if len(providers) != 1 || providers[0].Configured || !providers[0].Capabilities["conversions"] {
		t.Fatalf("providers = %+v", providers)
	}
	if len(providers[0].TargetingFacets) != 2 || providers[0].TargetingFacets[1] != "job_title" {
		t.Fatalf("facets = %v", providers[0].TargetingFacets)
	}
}

func TestAdsCompanyRowsTravelWithTheRequest(t *testing.T) {
	var path string
	var body map[string][]AdCompany
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = io.WriteString(w, `{"data":{"added":2}}`)
	})

	added, err := client.Ads.AddAudienceCompanies(
		context.Background(),
		"urn:li:adSegment:44",
		&AdObjectParams{WorkspaceID: "ws_1", ConnectionID: "conn_1"},
		[]AdCompany{{Domain: "northwind.example"}, {Name: "Contoso"}},
	)
	if err != nil {
		t.Fatalf("Ads.AddAudienceCompanies: %v", err)
	}
	if path != "/ads/audiences/urn:li:adSegment:44/companies" {
		t.Fatalf("path = %s", path)
	}
	if added != 2 || len(body["companies"]) != 2 || body["companies"][0].Domain != "northwind.example" {
		t.Fatalf("added = %d body = %+v", added, body)
	}
}

func TestAdsConversionEventsSendTheIdentityTheAPIHashes(t *testing.T) {
	var path, query string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		path, query = r.URL.Path, r.URL.RawQuery
		_, _ = io.WriteString(w, `{"data":{"accepted":1}}`)
	})

	accepted, err := client.Ads.SendConversionEvents(
		context.Background(),
		"urn:li:conversion:9",
		&AdObjectParams{WorkspaceID: "ws_1", ConnectionID: "conn_1"},
		[]ConversionEvent{{HappenedAt: 1758326400000, Email: "buyer@example.test"}},
	)
	if err != nil {
		t.Fatalf("Ads.SendConversionEvents: %v", err)
	}
	if path != "/ads/linkedin/conversion-rules/urn:li:conversion:9/events" {
		t.Fatalf("path = %s", path)
	}
	if query != "connection_id=conn_1&workspace_id=ws_1" {
		t.Fatalf("query = %s", query)
	}
	if accepted != 1 {
		t.Fatalf("accepted = %d", accepted)
	}
}
