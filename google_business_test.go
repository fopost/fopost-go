package fopost

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

// Every Business Profile method, with the request it must send.
func TestGoogleBusinessMethodsMapOntoTheirRoutes(t *testing.T) {
	var method, path string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		_, _ = io.WriteString(w, `{"data":{"ok":true}}`)
	})
	ctx := context.Background()
	gb := client.GoogleBusiness

	cases := []struct {
		name   string
		call   func() error
		method string
		path   string
	}{
		{"GetLocation", func() error { _, err := gb.GetLocation(ctx, "acc_1"); return err },
			"GET", "/accounts/acc_1/gbp/location"},
		{"UpdateLocation", func() error { _, err := gb.UpdateLocation(ctx, "acc_1", nil); return err },
			"PATCH", "/accounts/acc_1/gbp/location"},
		{"GetAttributes", func() error { _, err := gb.GetAttributes(ctx, "acc_1", nil); return err },
			"GET", "/accounts/acc_1/gbp/attributes"},
		{"UpdateAttributes", func() error { _, err := gb.UpdateAttributes(ctx, "acc_1", nil); return err },
			"PATCH", "/accounts/acc_1/gbp/attributes"},
		{"GetMenus", func() error { _, err := gb.GetMenus(ctx, "acc_1"); return err },
			"GET", "/accounts/acc_1/gbp/menus"},
		{"ReplaceMenus", func() error { _, err := gb.ReplaceMenus(ctx, "acc_1", nil); return err },
			"PUT", "/accounts/acc_1/gbp/menus"},
		{"GetServices", func() error { _, err := gb.GetServices(ctx, "acc_1"); return err },
			"GET", "/accounts/acc_1/gbp/services"},
		{"ReplaceServices", func() error { _, err := gb.ReplaceServices(ctx, "acc_1", nil); return err },
			"PUT", "/accounts/acc_1/gbp/services"},
		{"ListMedia", func() error { _, err := gb.ListMedia(ctx, "acc_1", 0, ""); return err },
			"GET", "/accounts/acc_1/gbp/media"},
		{"AddMedia", func() error {
			_, err := gb.AddMedia(ctx, "acc_1", &AddGoogleBusinessMediaRequest{MediaID: "m_1"})
			return err
		}, "POST", "/accounts/acc_1/gbp/media"},
		{"DeleteMedia", func() error { _, err := gb.DeleteMedia(ctx, "acc_1", "CAoSL"); return err },
			"DELETE", "/accounts/acc_1/gbp/media/CAoSL"},
		{"ListPlaceActions", func() error { _, err := gb.ListPlaceActions(ctx, "acc_1"); return err },
			"GET", "/accounts/acc_1/gbp/place-actions"},
		{"CreatePlaceAction", func() error {
			_, err := gb.CreatePlaceAction(ctx, "acc_1", &CreateGoogleBusinessPlaceActionRequest{
				URI: "https://example.com/book", PlaceActionType: "APPOINTMENT",
			})
			return err
		}, "POST", "/accounts/acc_1/gbp/place-actions"},
		{"UpdatePlaceAction", func() error {
			_, err := gb.UpdatePlaceAction(ctx, "acc_1", "links-1", nil)
			return err
		}, "PATCH", "/accounts/acc_1/gbp/place-actions/links-1"},
		{"DeletePlaceAction", func() error { _, err := gb.DeletePlaceAction(ctx, "acc_1", "links-1"); return err },
			"DELETE", "/accounts/acc_1/gbp/place-actions/links-1"},
		{"GetVerificationOptions", func() error { _, err := gb.GetVerificationOptions(ctx, "acc_1", ""); return err },
			"GET", "/accounts/acc_1/gbp/verification"},
		{"StartVerification", func() error {
			_, err := gb.StartVerification(ctx, "acc_1", &StartGoogleBusinessVerificationRequest{Method: "SMS"})
			return err
		}, "POST", "/accounts/acc_1/gbp/verification/start"},
		{"CompleteVerification", func() error {
			_, err := gb.CompleteVerification(ctx, "acc_1", "v1", "123456")
			return err
		}, "POST", "/accounts/acc_1/gbp/verification/complete"},
		{"GetPerformance", func() error {
			_, err := gb.GetPerformance(ctx, "acc_1", "2026-09-01", "2026-09-07", nil)
			return err
		}, "GET", "/accounts/acc_1/gbp/performance"},
		{"Assign", func() error { _, err := gb.Assign(ctx, "acc_1", "ws_2"); return err },
			"POST", "/accounts/acc_1/gbp/assign"},
	}

	for _, tc := range cases {
		if err := tc.call(); err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if method != tc.method || path != tc.path {
			t.Fatalf("%s sent %s %s, want %s %s", tc.name, method, path, tc.method, tc.path)
		}
	}
}

func TestGoogleBusinessUpdateLocationSendsOnlyTheFieldsSet(t *testing.T) {
	var body map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = io.WriteString(w, `{"data":{}}`)
	})

	title, website := "Corner Bakery", ""
	if _, err := client.GoogleBusiness.UpdateLocation(context.Background(), "acc_1",
		&UpdateGoogleBusinessLocationRequest{Title: &title, WebsiteURI: &website}); err != nil {
		t.Fatalf("UpdateLocation: %v", err)
	}

	if len(body) != 2 || body["title"] != "Corner Bakery" || body["website_uri"] != nil {
		t.Fatalf("body = %+v, want the title set and the website cleared", body)
	}
}

func TestGoogleBusinessPerformanceRepeatsTheMetricParameter(t *testing.T) {
	var metrics []string
	var start string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		metrics = r.URL.Query()["daily_metrics"]
		start = r.URL.Query().Get("start_date")
		_, _ = io.WriteString(w, `{"data":{}}`)
	})

	if _, err := client.GoogleBusiness.GetPerformance(context.Background(), "acc_1",
		"2026-09-01", "2026-09-07", []string{"CALL_CLICKS", "WEBSITE_CLICKS"}); err != nil {
		t.Fatalf("GetPerformance: %v", err)
	}

	if len(metrics) != 2 || metrics[0] != "CALL_CLICKS" || metrics[1] != "WEBSITE_CLICKS" {
		t.Fatalf("daily_metrics = %v", metrics)
	}
	if start != "2026-09-01" {
		t.Fatalf("start_date = %q", start)
	}
}

func TestGoogleBusinessSearchKeywordsAsksTheSameRoute(t *testing.T) {
	var keywords string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		keywords = r.URL.Query().Get("keywords")
		_, _ = io.WriteString(w, `{"data":{}}`)
	})

	if _, err := client.GoogleBusiness.GetSearchKeywords(context.Background(), "acc_1",
		"2026-08-01", "2026-09-01", ""); err != nil {
		t.Fatalf("GetSearchKeywords: %v", err)
	}
	if keywords != "true" {
		t.Fatalf("keywords = %q", keywords)
	}
}

func TestGoogleBusinessSurfacesAPendingAPIGrant(t *testing.T) {
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = io.WriteString(w, `{"error":"configuration_error","message":"Not available yet"}`)
	})

	_, err := client.GoogleBusiness.GetLocation(context.Background(), "acc_1")
	apiErr, ok := APIError(err)
	if !ok {
		t.Fatalf("err = %v, want a *fopost.Error", err)
	}
	if apiErr.Status != http.StatusServiceUnavailable || apiErr.Code != "configuration_error" {
		t.Fatalf("err = %+v", apiErr)
	}
}
