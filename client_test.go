package fopost

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// testClient points a client at a stub server with retries wound down so a
// test never waits on a real backoff.
func testClient(t *testing.T, handler http.HandlerFunc, opts ...Option) (*Client, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	opts = append([]Option{WithBaseURL(server.URL), WithMaxRetries(1)}, opts...)
	client, err := New("fp_test", opts...)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return client, server
}

func TestNewRequiresAPIKey(t *testing.T) {
	t.Setenv("FOPOST_API_KEY", "")
	if _, err := New(""); err == nil {
		t.Fatal("expected an error without an API key")
	}
}

func TestNewReadsAPIKeyFromEnv(t *testing.T) {
	t.Setenv("FOPOST_API_KEY", "fp_from_env")
	client, err := New("")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if client.apiKey != "fp_from_env" {
		t.Fatalf("apiKey = %q, want fp_from_env", client.apiKey)
	}
	if client.BaseURL() != DefaultBaseURL {
		t.Fatalf("BaseURL() = %q, want %q", client.BaseURL(), DefaultBaseURL)
	}
}

func TestBaseURLFromEnv(t *testing.T) {
	t.Setenv("FOPOST_BASE_URL", "http://localhost:8080/v1/")
	client, err := New("fp_test")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if client.BaseURL() != "http://localhost:8080/v1" {
		t.Fatalf("BaseURL() = %q", client.BaseURL())
	}
}

func TestRequestCarriesAuthAndAgent(t *testing.T) {
	var gotKey, gotAgent, gotAccept, gotPath, gotQuery string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("X-API-Key")
		gotAgent = r.Header.Get("User-Agent")
		gotAccept = r.Header.Get("Accept")
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data":[]}`)
	})

	if _, err := client.Labels.List(context.Background(), "ws_1"); err != nil {
		t.Fatalf("Labels.List: %v", err)
	}
	if gotKey != "fp_test" {
		t.Errorf("X-API-Key = %q", gotKey)
	}
	if !strings.HasPrefix(gotAgent, "fopost-go/") {
		t.Errorf("User-Agent = %q", gotAgent)
	}
	if gotAccept != "application/json" {
		t.Errorf("Accept = %q", gotAccept)
	}
	if gotPath != "/labels" {
		t.Errorf("path = %q", gotPath)
	}
	if gotQuery != "workspace_id=ws_1" {
		t.Errorf("query = %q", gotQuery)
	}
}

func TestWithUserAgentPrefixes(t *testing.T) {
	var got string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("User-Agent")
		_, _ = io.WriteString(w, `{"data":[]}`)
	}, WithUserAgent("my-app/2.0"))

	if _, err := client.Workspaces.List(context.Background()); err != nil {
		t.Fatalf("Workspaces.List: %v", err)
	}
	if !strings.HasPrefix(got, "my-app/2.0 fopost-go/") {
		t.Fatalf("User-Agent = %q", got)
	}
}

func TestZeroParamsAreNotSent(t *testing.T) {
	var query string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		_, _ = io.WriteString(w, `{"data":[],"meta":{}}`)
	})

	if _, err := client.Posts.List(context.Background(), &ListPostsParams{WorkspaceID: "ws_1"}); err != nil {
		t.Fatalf("Posts.List: %v", err)
	}
	if query != "workspace_id=ws_1" {
		t.Fatalf("query = %q, want only the workspace", query)
	}
}

func TestEnvelopeIsUnwrappedOnlyWhenPresent(t *testing.T) {
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		// Posts answer bare; most other resources are wrapped in {"data": ...}.
		_, _ = io.WriteString(w, `{"id":"post_1","status":"draft"}`)
	})

	post, err := client.Posts.Get(context.Background(), "post_1")
	if err != nil {
		t.Fatalf("Posts.Get: %v", err)
	}
	if post.ID != "post_1" || post.Status != "draft" {
		t.Fatalf("post = %+v", post)
	}
}

func TestPathIDsAreEscaped(t *testing.T) {
	var path string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.EscapedPath()
		_, _ = io.WriteString(w, `{"data":{}}`)
	})

	if _, err := client.Posts.Get(context.Background(), "a b/../c"); err != nil {
		t.Fatalf("Posts.Get: %v", err)
	}
	if path != "/posts/a%20b%2F..%2Fc" {
		t.Fatalf("path = %q", path)
	}
}

func TestDoIsTheEscapeHatch(t *testing.T) {
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"data":[{"name":"twitter"}]}`)
	})

	var body map[string]any
	if err := client.Do(context.Background(), "GET", "/platforms", nil, nil, &body); err != nil {
		t.Fatalf("Do: %v", err)
	}
	// Do hands back the whole body, envelope and all.
	if _, ok := body["data"]; !ok {
		t.Fatalf("body = %v, want the envelope kept", body)
	}
}

func TestRetriesRateLimitAndHonorsRetryAfter(t *testing.T) {
	var calls int32
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.Header().Set("Retry-After", "0")
			w.Header().Set("X-RateLimit-Limit", "100")
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = io.WriteString(w, `{"error":"too_many_requests","message":"slow down"}`)
			return
		}
		_, _ = io.WriteString(w, `{"data":[]}`)
	}, WithMaxRetries(3))

	if _, err := client.Workspaces.List(context.Background()); err != nil {
		t.Fatalf("Workspaces.List: %v", err)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
}

func TestGivesUpAfterMaxRetries(t *testing.T) {
	var calls int32
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Retry-After", "0")
		w.Header().Set("X-RateLimit-Limit", "100")
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = io.WriteString(w, `{"error":"too_many_requests","message":"slow down"}`)
	}, WithMaxRetries(2))

	_, err := client.Workspaces.List(context.Background())
	if !IsRateLimited(err) {
		t.Fatalf("err = %v, want a rate limit error", err)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
	apiErr, _ := APIError(err)
	if apiErr.RateLimit.Limit != 100 || apiErr.RateLimit.Remaining != 0 {
		t.Fatalf("rate limit = %+v", apiErr.RateLimit)
	}
}

func TestDoesNotRetryClientErrors(t *testing.T) {
	var calls int32
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"error":"not_found","message":"Post not found"}`)
	}, WithMaxRetries(3))

	_, err := client.Posts.Get(context.Background(), "post_1")
	if !IsNotFound(err) {
		t.Fatalf("err = %v, want a 404", err)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want no retry", calls)
	}
}

func TestRetriesServerErrors(t *testing.T) {
	var calls int32
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) < 3 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		_, _ = io.WriteString(w, `{"data":[]}`)
	}, WithMaxRetries(3))

	if _, err := client.Workspaces.List(context.Background()); err != nil {
		t.Fatalf("Workspaces.List: %v", err)
	}
	if calls != 3 {
		t.Fatalf("calls = %d, want 3", calls)
	}
}

func TestContextCancellationStopsRetrying(t *testing.T) {
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(http.StatusTooManyRequests)
	}, WithMaxRetries(3))

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	if _, err := client.Workspaces.List(ctx); err == nil {
		t.Fatal("expected the cancelled context to surface")
	}
}

func TestNonJSONBodyStillErrors(t *testing.T) {
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = io.WriteString(w, "<html>nginx</html>")
	})

	_, err := client.Workspaces.List(context.Background())
	apiErr, ok := APIError(err)
	if !ok {
		t.Fatalf("err = %v, want an API error", err)
	}
	if apiErr.Status != http.StatusBadGateway {
		t.Fatalf("status = %d", apiErr.Status)
	}
}

func TestRequestBodyIsJSON(t *testing.T) {
	var body map[string]any
	var contentType string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		contentType = r.Header.Get("Content-Type")
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"id":"post_1","status":"draft"}`)
	})

	_, err := client.Posts.Create(context.Background(), &CreatePostRequest{
		WorkspaceID: "ws_1",
		Accounts:    []string{"acc_1"},
		Content:     Text("Hello"),
	})
	if err != nil {
		t.Fatalf("Posts.Create: %v", err)
	}
	if contentType != "application/json" {
		t.Fatalf("Content-Type = %q", contentType)
	}
	if body["workspace_id"] != "ws_1" {
		t.Fatalf("body = %v", body)
	}
	// Unset optional fields stay out of the payload entirely.
	if _, ok := body["schedule_at"]; ok {
		t.Fatalf("schedule_at should be omitted, body = %v", body)
	}
}
