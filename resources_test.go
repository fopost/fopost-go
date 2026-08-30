package fopost

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestWorkspacesListUnwrapsAndDecodesAccounts(t *testing.T) {
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"data":[{"id":"ws_1","name":"Acme","slug":"acme","type":"TEAM","accounts":[{"id":"acc_1","platform":"twitter","username":"acme"}]}]}`)
	})

	workspaces, err := client.Workspaces.List(context.Background())
	if err != nil {
		t.Fatalf("Workspaces.List: %v", err)
	}
	if len(workspaces) != 1 || len(workspaces[0].Accounts) != 1 {
		t.Fatalf("workspaces = %+v", workspaces)
	}
}

func TestAccountsHealthSendsRefresh(t *testing.T) {
	var query string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		_, _ = io.WriteString(w, `{"data":{"id":"acc_1","platform":"twitter","username":"acme","active":true,"healthStatus":"healthy"}}`)
	})

	health, err := client.Accounts.Health(context.Background(), "acc_1", true)
	if err != nil {
		t.Fatalf("Accounts.Health: %v", err)
	}
	if query != "refresh=true" {
		t.Fatalf("query = %q", query)
	}
	if health.HealthStatus != HealthHealthy {
		t.Fatalf("health = %+v", health)
	}
}

func TestAccountsHealthOmitsRefreshWhenFalse(t *testing.T) {
	var query string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		_, _ = io.WriteString(w, `{"data":{}}`)
	})

	if _, err := client.Accounts.Health(context.Background(), "acc_1", false); err != nil {
		t.Fatalf("Accounts.Health: %v", err)
	}
	if query != "" {
		t.Fatalf("query = %q, want nothing sent", query)
	}
}

func TestCommunitiesRemoveUsesTheRowID(t *testing.T) {
	var path, method string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		path, method = r.URL.Path, r.Method
		_, _ = io.WriteString(w, `{"success":true}`)
	})

	if err := client.Communities.Remove(context.Background(), "acc_1", 42); err != nil {
		t.Fatalf("Communities.Remove: %v", err)
	}
	if method != "DELETE" || path != "/accounts/acc_1/communities/42" {
		t.Fatalf("%s %s", method, path)
	}
}

func TestWebhookCreateReturnsTheSecretOnce(t *testing.T) {
	var body map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"data":{"id":"wh_1","workspaceId":"ws_1","url":"https://example.com/hook","secret":"whsec_x","events":["post.published"],"active":true}}`)
	})

	hook, err := client.Webhooks.Create(context.Background(), &CreateWebhookRequest{
		WorkspaceID: "ws_1",
		URL:         "https://example.com/hook",
		Events:      []string{EventPostPublished},
	})
	if err != nil {
		t.Fatalf("Webhooks.Create: %v", err)
	}
	if hook.Secret != "whsec_x" {
		t.Fatalf("hook = %+v", hook)
	}
	events, _ := body["events"].([]any)
	if len(events) != 1 || events[0] != "post.published" {
		t.Fatalf("body = %v", body)
	}
}

func TestAutomationRunsKeepTheirMeta(t *testing.T) {
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"data":[{"id":7,"status":"completed"}],"meta":{"current_page":1,"per_page":30,"total":1,"last_page":1}}`)
	})

	runs, err := client.Automations.Runs(context.Background(), "auto_1", 0, 0)
	if err != nil {
		t.Fatalf("Automations.Runs: %v", err)
	}
	if len(runs.Data) != 1 || runs.Data[0].ID != 7 || runs.Meta.Total != 1 {
		t.Fatalf("runs = %+v", runs)
	}
}

func TestAnalyticsOverviewDecodesNullableDeltas(t *testing.T) {
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"data":{"totalAccounts":2,"totalFollowers":100,"engagementRate":null,"deltas":{"followers":0.12,"posts":null},"todayStats":{"posts":1,"followerChange":3,"engagement":9},"platforms":[],"accounts":[]}}`)
	})

	overview, err := client.Analytics.Overview(context.Background(), &AnalyticsParams{Days: 30})
	if err != nil {
		t.Fatalf("Analytics.Overview: %v", err)
	}
	if overview.EngagementRate != nil {
		t.Fatal("a null rate should stay nil rather than becoming zero")
	}
	if overview.Deltas.Followers == nil || *overview.Deltas.Followers != 0.12 {
		t.Fatalf("deltas = %+v", overview.Deltas)
	}
	if overview.Deltas.Posts != nil {
		t.Fatal("a null delta should stay nil")
	}
}

func TestAnalyticsParamsMapToQuery(t *testing.T) {
	var query string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.Query().Encode()
		_, _ = io.WriteString(w, `{"data":[]}`)
	})

	_, err := client.Analytics.TopPosts(context.Background(), &AnalyticsParams{
		WorkspaceID: "ws_1", Limit: 10, Days: 30, Sort: "recent",
	})
	if err != nil {
		t.Fatalf("Analytics.TopPosts: %v", err)
	}
	if query != "days=30&limit=10&sort=recent&workspace_id=ws_1" {
		t.Fatalf("query = %q", query)
	}
}

func TestPostingStreakFlattensItsWrapper(t *testing.T) {
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"data":{"streak":[{"date":"2026-08-01","count":2,"publishedCount":2}]}}`)
	})

	streak, err := client.Analytics.PostingStreak(context.Background(), "ws_1")
	if err != nil {
		t.Fatalf("Analytics.PostingStreak: %v", err)
	}
	if len(streak) != 1 || streak[0].Count != 2 {
		t.Fatalf("streak = %+v", streak)
	}
}

func TestMediaUploadSendsEveryFile(t *testing.T) {
	var names []string
	var workspaceID string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Errorf("ParseMultipartForm: %v", err)
		}
		workspaceID = r.FormValue("workspaceId")
		for _, header := range r.MultipartForm.File["files"] {
			names = append(names, header.Filename)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"data":[{"id":"med_1","type":"image","name":"a.png","url":"https://cdn/a.png","size":12}]}`)
	})

	uploaded, err := client.Media.Upload(context.Background(), "ws_1",
		File{Name: "a.png", Content: strings.NewReader("png-a")},
		File{Name: "b.png", Content: strings.NewReader("png-b")},
	)
	if err != nil {
		t.Fatalf("Media.Upload: %v", err)
	}
	if workspaceID != "ws_1" || strings.Join(names, ",") != "a.png,b.png" {
		t.Fatalf("form = %q %v", workspaceID, names)
	}
	if len(uploaded) != 1 {
		t.Fatalf("uploaded = %+v", uploaded)
	}
	item := uploaded[0].AsMediaItem()
	if item.URL != "https://cdn/a.png" || item.Type != "image" {
		t.Fatalf("media item = %+v", item)
	}
}

func TestMediaUploadRejectsAnEmptyCall(t *testing.T) {
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("no request should be sent")
	})

	if _, err := client.Media.Upload(context.Background(), "ws_1"); err == nil {
		t.Fatal("expected an error with no files")
	}
}

func TestMediaListRequiresAWorkspace(t *testing.T) {
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("no request should be sent")
	})

	if _, err := client.Media.List(context.Background(), ""); err == nil {
		t.Fatal("expected an error without a workspace id")
	}
}
