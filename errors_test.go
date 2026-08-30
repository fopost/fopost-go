package fopost

import (
	"context"
	"io"
	"net/http"
	"testing"
)

func TestErrorCarriesCodeAndMessage(t *testing.T) {
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = io.WriteString(w, `{"error":"subscription_required","message":"An active subscription is required"}`)
	})

	_, err := client.Posts.Create(context.Background(), &CreatePostRequest{WorkspaceID: "ws_1"})
	apiErr, ok := APIError(err)
	if !ok {
		t.Fatalf("err = %v, want an API error", err)
	}
	if apiErr.Status != 403 || apiErr.Code != "subscription_required" {
		t.Fatalf("err = %+v", apiErr)
	}
	if apiErr.Message != "An active subscription is required" {
		t.Fatalf("message = %q", apiErr.Message)
	}
	if !IsForbidden(err) || CodeOf(err) != "subscription_required" || StatusOf(err) != 403 {
		t.Fatalf("predicates disagree with %+v", apiErr)
	}
	if got := apiErr.Error(); got == "" {
		t.Fatal("Error() should describe the failure")
	}
}

func TestPaymentRequiredExposesUpgradeURL(t *testing.T) {
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusPaymentRequired)
		_, _ = io.WriteString(w, `{"error":"insufficient_credits","message":"Out of AI credits","upgrade_url":"https://app.fopost.com/settings/billing"}`)
	})

	_, err := client.Posts.Publish(context.Background(), "post_1", nil)
	if !IsPaymentRequired(err) {
		t.Fatalf("err = %v, want a 402", err)
	}
	apiErr, _ := APIError(err)
	if apiErr.UpgradeURL() != "https://app.fopost.com/settings/billing" {
		t.Fatalf("UpgradeURL() = %q", apiErr.UpgradeURL())
	}
}

func TestErrorFieldDecodesExtraContext(t *testing.T) {
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = io.WriteString(w, `{"error":"validation_failed","message":"Bad request","issues":["content is required"]}`)
	})

	_, err := client.Posts.Publish(context.Background(), "post_1", nil)
	apiErr, ok := APIError(err)
	if !ok {
		t.Fatalf("err = %v", err)
	}
	var issues []string
	if err := apiErr.Field("issues", &issues); err != nil {
		t.Fatalf("Field: %v", err)
	}
	if len(issues) != 1 || issues[0] != "content is required" {
		t.Fatalf("issues = %v", issues)
	}
}

func TestErrorWithoutJSONBodyFallsBackToStatus(t *testing.T) {
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})

	_, err := client.Workspaces.List(context.Background())
	if !IsUnauthorized(err) {
		t.Fatalf("err = %v, want a 401", err)
	}
	apiErr, _ := APIError(err)
	if apiErr.Message == "" {
		t.Fatal("an empty body should still produce a message")
	}
}

func TestPredicatesIgnoreNonAPIErrors(t *testing.T) {
	if IsNotFound(io.EOF) || StatusOf(io.EOF) != 0 || CodeOf(io.EOF) != "" {
		t.Fatal("a plain error is not an API error")
	}
}
