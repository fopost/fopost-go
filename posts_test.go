package fopost

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestPostsListDecodesPageMeta(t *testing.T) {
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"data":[{"id":"post_1","status":"published"}],"meta":{"current_page":1,"per_page":30,"total":1,"last_page":1}}`)
	})

	page, err := client.Posts.List(context.Background(), &ListPostsParams{WorkspaceID: "ws_1"})
	if err != nil {
		t.Fatalf("Posts.List: %v", err)
	}
	if len(page.Data) != 1 || page.Data[0].ID != "post_1" {
		t.Fatalf("data = %+v", page.Data)
	}
	if page.Meta.Total != 1 || page.Meta.LastPage != 1 {
		t.Fatalf("meta = %+v", page.Meta)
	}
}

func TestPostsEachWalksEveryPage(t *testing.T) {
	var pages []string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		pages = append(pages, page)
		switch page {
		case "1":
			_, _ = io.WriteString(w, `{"data":[{"id":"a"},{"id":"b"}],"meta":{"current_page":1,"per_page":2,"total":3,"last_page":2}}`)
		default:
			_, _ = io.WriteString(w, `{"data":[{"id":"c"}],"meta":{"current_page":2,"per_page":2,"total":3,"last_page":2}}`)
		}
	})

	var ids []string
	err := client.Posts.Each(context.Background(), &ListPostsParams{PerPage: 2}, func(post Post) error {
		ids = append(ids, post.ID)
		return nil
	})
	if err != nil {
		t.Fatalf("Posts.Each: %v", err)
	}
	if strings.Join(ids, ",") != "a,b,c" {
		t.Fatalf("ids = %v", ids)
	}
	if strings.Join(pages, ",") != "1,2" {
		t.Fatalf("pages = %v, want it to stop at last_page", pages)
	}
}

func TestPostsEachStopsOnCallerError(t *testing.T) {
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"data":[{"id":"a"},{"id":"b"}],"meta":{"current_page":1,"per_page":2,"total":99,"last_page":50}}`)
	})

	stop := fmt.Errorf("enough")
	seen := 0
	err := client.Posts.Each(context.Background(), nil, func(post Post) error {
		seen++
		return stop
	})
	if err != stop {
		t.Fatalf("err = %v, want the caller's error", err)
	}
	if seen != 1 {
		t.Fatalf("seen = %d, want the walk to stop immediately", seen)
	}
}

func TestPostsCreateSendsScheduleAsRFC3339(t *testing.T) {
	var body map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"id":"post_1","status":"scheduled","schedule_at":"2026-09-01T10:00:00.000Z"}`)
	})

	at := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	post, err := client.Posts.Create(context.Background(), &CreatePostRequest{
		WorkspaceID: "ws_1",
		Accounts:    []string{"acc_1"},
		Content:     Thread("First", "Second"),
		Status:      PostStatusScheduled,
		ScheduleAt:  NewTime(at),
	})
	if err != nil {
		t.Fatalf("Posts.Create: %v", err)
	}
	if body["schedule_at"] != "2026-09-01T10:00:00Z" {
		t.Fatalf("schedule_at = %v", body["schedule_at"])
	}
	content, _ := body["content"].([]any)
	if len(content) != 2 {
		t.Fatalf("content = %v, want one block per thread entry", body["content"])
	}
	if !post.ScheduleAt.Time.Equal(at) {
		t.Fatalf("ScheduleAt = %v, want %v", post.ScheduleAt, at)
	}
}

func TestPublishDryRunSetsOption(t *testing.T) {
	var body map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = io.WriteString(w, `{"data":{"dryRun":true,"post":{"id":"post_1","status":"draft"},"accounts":[{"accountId":"acc_1","platform":"twitter"}],"healthWarnings":[]}}`)
	})

	result, err := client.Posts.Publish(context.Background(), "post_1", &PublishOptions{DryRun: true, AccountIDs: []string{"acc_1"}})
	if err != nil {
		t.Fatalf("Posts.Publish: %v", err)
	}
	options, _ := body["options"].(map[string]any)
	if options["dryRun"] != true {
		t.Fatalf("body = %v", body)
	}
	if !result.DryRun || len(result.Accounts) != 1 {
		t.Fatalf("result = %+v", result)
	}
}

func TestPublishDecodesDeliveries(t *testing.T) {
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = io.WriteString(w, `{"data":{"post_status":"publishing","deliveries":[{"id":"del_1","accountId":"acc_1","status":"queued","attempts":0}],"healthWarnings":[{"accountId":"acc_2","platform":"tiktok","healthStatus":"expired","message":"Reconnect"}]}}`)
	})

	result, err := client.Posts.Publish(context.Background(), "post_1", nil)
	if err != nil {
		t.Fatalf("Posts.Publish: %v", err)
	}
	if result.PostStatus != PostStatusPublishing || len(result.Deliveries) != 1 {
		t.Fatalf("result = %+v", result)
	}
	if len(result.HealthWarnings) != 1 || result.HealthWarnings[0].AccountID != "acc_2" {
		t.Fatalf("warnings = %+v", result.HealthWarnings)
	}
}

func TestPreflightDecodesSignals(t *testing.T) {
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"data":{"ready":false,"post":{"id":"post_1","status":"draft"},"accounts":[{"accountId":"acc_1","platform":"twitter","username":"acme","ready":false,"issues":["Content exceeds the limit"],"score":42,"signals":[{"level":"warn","code":"no_media","message":"Posts with media do better"}]}]}}`)
	})

	result, err := client.Posts.Preflight(context.Background(), "post_1")
	if err != nil {
		t.Fatalf("Posts.Preflight: %v", err)
	}
	if result.Ready {
		t.Fatal("ready should be false")
	}
	account := result.Accounts[0]
	if len(account.Issues) != 1 || len(account.Signals) != 1 || account.Signals[0].Level != "warn" {
		t.Fatalf("account = %+v", account)
	}
}

func TestBulkActionsSendTheirDiscriminator(t *testing.T) {
	var body map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = io.WriteString(w, `{"updated":2,"action":"label","mode":"add"}`)
	})

	result, err := client.Posts.BulkLabel(context.Background(), "ws_1", []string{"p1", "p2"}, []string{"l1"}, "add")
	if err != nil {
		t.Fatalf("Posts.BulkLabel: %v", err)
	}
	if body["action"] != "label" || body["mode"] != "add" {
		t.Fatalf("body = %v", body)
	}
	if result.Updated != 2 {
		t.Fatalf("result = %+v", result)
	}
}

func TestBulkLabelSendsEmptyListToClearLabels(t *testing.T) {
	var body map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = io.WriteString(w, `{"updated":1,"action":"label"}`)
	})

	if _, err := client.Posts.BulkLabel(context.Background(), "ws_1", []string{"p1"}, nil, ""); err != nil {
		t.Fatalf("Posts.BulkLabel: %v", err)
	}
	labels, ok := body["label_ids"].([]any)
	if !ok || len(labels) != 0 {
		t.Fatalf("label_ids = %v, want an empty array", body["label_ids"])
	}
}

func TestRetryWithoutOptionsSendsAnObject(t *testing.T) {
	var raw []byte
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ = io.ReadAll(r.Body)
		_, _ = io.WriteString(w, `{"data":{"post_status":"publishing","deliveries":[]}}`)
	})

	if _, err := client.Posts.Retry(context.Background(), "post_1", nil); err != nil {
		t.Fatalf("Posts.Retry: %v", err)
	}
	if strings.TrimSpace(string(raw)) != "{}" {
		t.Fatalf("body = %q, want an empty object rather than null", raw)
	}
}

func TestDeliveriesUnwrapsTheEnvelope(t *testing.T) {
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"data":[{"id":"del_1","accountId":"acc_1","status":"published","platform":"twitter"}]}`)
	})

	deliveries, err := client.Posts.Deliveries(context.Background(), "post_1")
	if err != nil {
		t.Fatalf("Posts.Deliveries: %v", err)
	}
	if len(deliveries) != 1 || deliveries[0].Platform != "twitter" {
		t.Fatalf("deliveries = %+v", deliveries)
	}
}

func TestBulkImportPostsMultipart(t *testing.T) {
	var contentType, workspaceID, filename, contents string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		contentType = r.Header.Get("Content-Type")
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Errorf("ParseMultipartForm: %v", err)
		}
		workspaceID = r.FormValue("workspace_id")
		file, header, err := r.FormFile("file")
		if err != nil {
			t.Errorf("FormFile: %v", err)
		} else {
			filename = header.Filename
			body, _ := io.ReadAll(file)
			contents = string(body)
		}
		_, _ = io.WriteString(w, `{"total_rows":1,"valid_rows":1,"invalid_rows":0,"rows":[]}`)
	})

	validation, err := client.Posts.ValidateBulkImport(context.Background(), "ws_1", "posts.csv", strings.NewReader("content\nHello\n"))
	if err != nil {
		t.Fatalf("ValidateBulkImport: %v", err)
	}
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		t.Fatalf("Content-Type = %q", contentType)
	}
	if workspaceID != "ws_1" || filename != "posts.csv" || !strings.Contains(contents, "Hello") {
		t.Fatalf("form = %q %q %q", workspaceID, filename, contents)
	}
	if validation.ValidRows != 1 {
		t.Fatalf("validation = %+v", validation)
	}
}

func TestPostsCreateSendsAccountGroupWithoutAccounts(t *testing.T) {
	var body map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"data":{"id":"post_1","status":"draft"}}`)
	})

	if _, err := client.Posts.Create(context.Background(), &CreatePostRequest{
		WorkspaceID:    "ws_1",
		AccountGroupID: "grp_1",
		Content:        Text("Hello"),
	}); err != nil {
		t.Fatalf("Posts.Create: %v", err)
	}
	if body["account_group_id"] != "grp_1" {
		t.Fatalf("account_group_id = %v", body["account_group_id"])
	}
	if _, sent := body["accounts"]; sent {
		t.Fatalf("accounts = %v, want it omitted", body["accounts"])
	}
}
