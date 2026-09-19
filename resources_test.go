package fopost

import (
	"context"
	"encoding/json"
	"fmt"
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

func TestMediaUploadDirectPresignsPutsAndCompletes(t *testing.T) {
	var calls []string
	var presignBody map[string]any
	var putKey, putType string
	var putLength int64
	var putBody []byte
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		switch {
		case r.Method == "POST" && r.URL.Path == "/media/presign":
			_ = json.NewDecoder(r.Body).Decode(&presignBody)
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, `{"data":{"uploadId":"up_1","uploadUrl":"`+"http://"+r.Host+`/bucket/up_1","method":"PUT","headers":{"Content-Type":"image/png"},"expiresAt":"2026-09-19T12:00:00Z"}}`)
		case r.Method == "PUT" && r.URL.Path == "/bucket/up_1":
			putKey = r.Header.Get("X-API-Key")
			putType = r.Header.Get("Content-Type")
			putLength = r.ContentLength
			putBody, _ = io.ReadAll(r.Body)
			w.WriteHeader(http.StatusOK)
		case r.Method == "POST" && r.URL.Path == "/media/presign/up_1/complete":
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, `{"data":{"id":"med_1","type":"image","name":"a.png","url":"https://cdn/a.png","previewUrl":"https://cdn/p/a.png","size":5}}`)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
	})

	item, err := client.Media.UploadDirect(context.Background(), "ws_1", "a.png", "image/png", []byte("png-a"))
	if err != nil {
		t.Fatalf("Media.UploadDirect: %v", err)
	}
	if strings.Join(calls, ",") != "POST /media/presign,PUT /bucket/up_1,POST /media/presign/up_1/complete" {
		t.Fatalf("calls = %v", calls)
	}
	if presignBody["workspaceId"] != "ws_1" || presignBody["filename"] != "a.png" || presignBody["mimeType"] != "image/png" || presignBody["size"] != float64(5) {
		t.Fatalf("presign body = %v", presignBody)
	}
	if putKey != "" || putType != "image/png" || putLength != 5 || string(putBody) != "png-a" {
		t.Fatalf("put = key %q type %q length %d body %q", putKey, putType, putLength, putBody)
	}
	if item.ID != "med_1" || item.AsMediaItem().URL != "https://cdn/a.png" {
		t.Fatalf("item = %+v", item)
	}
}

func TestMediaUploadDirectReturnsAPIErrorOnRejectedPut(t *testing.T) {
	var completed bool
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/media/presign":
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, `{"data":{"uploadId":"up_1","uploadUrl":"`+"http://"+r.Host+`/bucket/up_1","method":"PUT","headers":{"Content-Type":"image/png"},"expiresAt":"2026-09-19T12:00:00Z"}}`)
		case r.URL.Path == "/bucket/up_1":
			w.WriteHeader(http.StatusForbidden)
			_, _ = io.WriteString(w, `<Error><Code>AccessDenied</Code></Error>`)
		default:
			completed = true
		}
	})

	_, err := client.Media.UploadDirect(context.Background(), "ws_1", "a.png", "image/png", []byte("png-a"))
	if !IsForbidden(err) {
		t.Fatalf("err = %v, want a 403 *Error", err)
	}
	if completed {
		t.Fatal("complete must not be called after a rejected PUT")
	}
}

func TestMediaCompleteRequiresAnUploadID(t *testing.T) {
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("no request should be sent")
	})

	if _, err := client.Media.Complete(context.Background(), ""); err == nil {
		t.Fatal("expected an error without an upload id")
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

func TestInboxListSendsFiltersAndKeepsMeta(t *testing.T) {
	var path, query string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		path, query = r.URL.Path, r.URL.RawQuery
		_, _ = io.WriteString(w, `{"data":[{"id":"itm_1","platform":"instagram","type":"comment","state":"unread","direction":"inbound","text":"Love this","attachments":[],"canReply":true,"account":{"id":"acc_1","platform":"instagram","username":"yourbrand"}}],"meta":{"page":2,"perPage":20,"total":41}}`)
	})

	page, err := client.Inbox.List(context.Background(), &ListInboxParams{
		WorkspaceID: "ws_1",
		Type:        InboxTypeComment,
		Sort:        InboxSortUnanswered,
		Page:        2,
		PerPage:     20,
	})
	if err != nil {
		t.Fatalf("Inbox.List: %v", err)
	}
	if path != "/inbox" {
		t.Fatalf("path = %q", path)
	}
	if query != "page=2&per_page=20&sort=unanswered&type=comment&workspace_id=ws_1" {
		t.Fatalf("query = %q", query)
	}
	if len(page.Data) != 1 || page.Data[0].Account.Username != "yourbrand" || !page.Data[0].CanReply {
		t.Fatalf("data = %+v", page.Data)
	}
	if page.Meta.Page != 2 || page.Meta.PerPage != 20 || page.Meta.Total != 41 {
		t.Fatalf("meta = %+v", page.Meta)
	}
}

func TestInboxMarkThreadReadSendsSnakeCaseBody(t *testing.T) {
	var method, path string
	var body map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = io.WriteString(w, `{"data":{"updated":3}}`)
	})

	updated, err := client.Inbox.MarkThreadRead(context.Background(), &MarkInboxThreadReadRequest{
		WorkspaceID:    "ws_1",
		AccountID:      "acc_1",
		PostExternalID: "18001",
	})
	if err != nil {
		t.Fatalf("Inbox.MarkThreadRead: %v", err)
	}
	if method != "POST" || path != "/inbox/read" {
		t.Fatalf("%s %s", method, path)
	}
	if body["workspace_id"] != "ws_1" || body["account_id"] != "acc_1" || body["post_external_id"] != "18001" {
		t.Fatalf("body = %v", body)
	}
	if _, present := body["conversation_id"]; present {
		t.Fatalf("conversation_id sent empty: %v", body)
	}
	if updated != 3 {
		t.Fatalf("updated = %d", updated)
	}
}

func TestInboxReplyPostsTextAndDecodesTheReply(t *testing.T) {
	var method, path string
	var body map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = io.WriteString(w, `{"data":{"item":{"id":"itm_1","state":"read","repliedAt":"2026-09-19T10:00:00Z"},"reply":{"externalId":"18002","externalUrl":"https://example.com/c/18002"}}}`)
	})

	result, err := client.Inbox.Reply(context.Background(), "itm_1", "Thanks!")
	if err != nil {
		t.Fatalf("Inbox.Reply: %v", err)
	}
	if method != "POST" || path != "/inbox/itm_1/reply" {
		t.Fatalf("%s %s", method, path)
	}
	if body["text"] != "Thanks!" {
		t.Fatalf("body = %v", body)
	}
	if result.Item.ID != "itm_1" || result.Item.RepliedAt.IsZero() || result.Reply.ExternalID != "18002" {
		t.Fatalf("result = %+v", result)
	}
}

func TestInboxReplyWithSendsMediaAndQuickRepliesWithoutText(t *testing.T) {
	var body map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = io.WriteString(w, `{"data":{"item":{"id":"itm_1"},"reply":{"externalId":null,"externalUrl":null}}}`)
	})

	_, err := client.Inbox.ReplyWith(context.Background(), "itm_1", &InboxReplyRequest{
		MediaIDs:     []string{"med_1"},
		QuickReplies: []string{"Yes", "No"},
	})
	if err != nil {
		t.Fatalf("Inbox.ReplyWith: %v", err)
	}
	if _, present := body["text"]; present {
		t.Fatalf("text sent empty: %v", body)
	}
	if fmt.Sprint(body["media_ids"]) != "[med_1]" || fmt.Sprint(body["quick_replies"]) != "[Yes No]" {
		t.Fatalf("body = %v", body)
	}
}

func TestInboxEditCommentPatchesText(t *testing.T) {
	var method, path string
	var body map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = io.WriteString(w, `{"data":{"id":"itm_1","text":"Fixed","editedAt":"2026-09-19T10:00:00Z","canEdit":true}}`)
	})

	item, err := client.Inbox.EditComment(context.Background(), "itm_1", "Fixed")
	if err != nil {
		t.Fatalf("Inbox.EditComment: %v", err)
	}
	if method != "PATCH" || path != "/inbox/itm_1" || len(body) != 1 || body["text"] != "Fixed" {
		t.Fatalf("%s %s %v", method, path, body)
	}
	if item.EditedAt.IsZero() || !item.CanEdit {
		t.Fatalf("item = %+v", item)
	}
}

func TestInboxLikeUnlikePinUnpinPostToTheirPaths(t *testing.T) {
	var paths []string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.Method+" "+r.URL.Path)
		_, _ = io.WriteString(w, `{"data":{"id":"itm_1","liked":true,"pinned":true,"canLike":true,"canPin":true}}`)
	})

	ctx := context.Background()
	item, err := client.Inbox.Like(ctx, "itm_1")
	if err != nil || !item.Liked || !item.CanLike {
		t.Fatalf("Inbox.Like: %+v, %v", item, err)
	}
	if _, err := client.Inbox.Unlike(ctx, "itm_1"); err != nil {
		t.Fatalf("Inbox.Unlike: %v", err)
	}
	if item, err = client.Inbox.Pin(ctx, "itm_1"); err != nil || !item.Pinned || !item.CanPin {
		t.Fatalf("Inbox.Pin: %+v, %v", item, err)
	}
	if _, err := client.Inbox.Unpin(ctx, "itm_1"); err != nil {
		t.Fatalf("Inbox.Unpin: %v", err)
	}
	want := "[POST /inbox/itm_1/like POST /inbox/itm_1/unlike POST /inbox/itm_1/pin POST /inbox/itm_1/unpin]"
	if fmt.Sprint(paths) != want {
		t.Fatalf("paths = %v", paths)
	}
}

func TestInboxReactSendsNullToRemove(t *testing.T) {
	var raw []string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		raw = append(raw, string(b))
		_, _ = io.WriteString(w, `{"data":{"id":"itm_1","reaction":"❤️","canReact":true}}`)
	})

	ctx := context.Background()
	item, err := client.Inbox.React(ctx, "itm_1", String("❤️"))
	if err != nil || item.Reaction != "❤️" || !item.CanReact {
		t.Fatalf("Inbox.React: %+v, %v", item, err)
	}
	if _, err := client.Inbox.React(ctx, "itm_1", nil); err != nil {
		t.Fatalf("Inbox.React(nil): %v", err)
	}
	if raw[0] != `{"reaction":"❤️"}` || raw[1] != `{"reaction":null}` {
		t.Fatalf("bodies = %v", raw)
	}
}

func TestInboxStartConversationSendsSnakeCaseBody(t *testing.T) {
	var method, path string
	var body map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"data":{"conversationId":"conv_2","item":{"id":"itm_9","type":"dm"}}}`)
	})

	started, err := client.Inbox.StartConversation(context.Background(), &StartInboxConversationRequest{
		CommentID: "itm_1",
		Text:      "Sent you the details",
		MediaIDs:  []string{"med_1"},
	})
	if err != nil {
		t.Fatalf("Inbox.StartConversation: %v", err)
	}
	if method != "POST" || path != "/inbox/conversations" {
		t.Fatalf("%s %s", method, path)
	}
	if body["comment_id"] != "itm_1" || body["text"] != "Sent you the details" || fmt.Sprint(body["media_ids"]) != "[med_1]" {
		t.Fatalf("body = %v", body)
	}
	if _, present := body["handle"]; present {
		t.Fatalf("handle sent empty: %v", body)
	}
	if started.ConversationID != "conv_2" || started.Item == nil || started.Item.Type != InboxTypeDM {
		t.Fatalf("started = %+v", started)
	}
}

func TestInboxSetTypingAndAccountsReadCanStartConversation(t *testing.T) {
	var path string
	var body map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		if r.Method == "GET" {
			_, _ = io.WriteString(w, `{"data":[{"id":"acc_1","canStartConversation":true}]}`)
			return
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = io.WriteString(w, `{"data":{"typing":false}}`)
	})

	ctx := context.Background()
	typing, err := client.Inbox.SetTyping(ctx, "conv_1", "acc_1", false)
	if err != nil || typing {
		t.Fatalf("Inbox.SetTyping: %v, %v", typing, err)
	}
	if path != "/inbox/conversations/conv_1/typing" || body["account_id"] != "acc_1" || body["on"] != false {
		t.Fatalf("%s %v", path, body)
	}
	accounts, err := client.Inbox.Accounts(ctx, "")
	if err != nil || len(accounts) != 1 || !accounts[0].CanStartConversation {
		t.Fatalf("Inbox.Accounts: %+v, %v", accounts, err)
	}
}

func TestInboxApproveReplyOmitsEmptyText(t *testing.T) {
	var path string
	var body map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = io.WriteString(w, `{"data":{"id":7,"outcome":"sent"}}`)
	})

	decision, err := client.Inbox.ApproveReply(context.Background(), 7, "")
	if err != nil {
		t.Fatalf("Inbox.ApproveReply: %v", err)
	}
	if path != "/inbox/approvals/7/approve" {
		t.Fatalf("path = %q", path)
	}
	if len(body) != 0 {
		t.Fatalf("body = %v, want empty object", body)
	}
	if decision.ID != 7 || decision.Outcome != "sent" {
		t.Fatalf("decision = %+v", decision)
	}
}

func TestAdsBoostSendsCamelCaseBodyAndDecodesTheAd(t *testing.T) {
	var method, path string
	var body map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"data":{"id":"ad_1","workspaceId":"ws_1","kind":"boost","name":"Launch","goal":"engagement","status":"paused","adAccountId":"act_1","sourcePostId":"post_1","budgetMinor":5000,"budgetType":"daily","currency":"USD","targeting":{"countries":["US"],"ageMin":18,"ageMax":65,"gender":"all"},"insights":null,"createdAt":"2026-09-19T10:00:00Z"}}`)
	})

	ad, err := client.Ads.Boost(context.Background(), &BoostPostRequest{
		WorkspaceID:  "ws_1",
		ConnectionID: "conn_1",
		AdAccountID:  "act_1",
		PostID:       "post_1",
		AccountID:    "acc_1",
		Name:         "Launch",
		Goal:         AdGoalEngagement,
		Budget:       AdBudget{Minor: 5000, Type: AdBudgetDaily},
		Targeting:    AdTargeting{Countries: []string{"US"}, AgeMin: 18, AgeMax: 65, Gender: "all"},
		Paused:       Bool(false),
	})
	if err != nil {
		t.Fatalf("Ads.Boost: %v", err)
	}
	if method != "POST" || path != "/ads/boost" {
		t.Fatalf("%s %s", method, path)
	}
	if body["workspaceId"] != "ws_1" || body["connectionId"] != "conn_1" || body["postId"] != "post_1" || body["accountId"] != "acc_1" {
		t.Fatalf("body = %v", body)
	}
	if body["paused"] != false {
		t.Fatalf("paused = %v", body["paused"])
	}
	budget, _ := body["budget"].(map[string]any)
	if budget["minor"] != float64(5000) || budget["type"] != "daily" {
		t.Fatalf("budget = %v", budget)
	}
	if _, present := budget["endAt"]; present {
		t.Fatalf("endAt sent empty: %v", budget)
	}
	if ad.ID != "ad_1" || ad.Kind != "boost" || ad.Status != AdStatusPaused || ad.Insights != nil || ad.Targeting.Countries[0] != "US" {
		t.Fatalf("ad = %+v", ad)
	}
}

func TestAdsSetStatusSendsWorkspaceQuery(t *testing.T) {
	var method, path, query string
	var body map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, path, query = r.Method, r.URL.Path, r.URL.RawQuery
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = io.WriteString(w, `{"data":{"id":"ad_1","status":"active"}}`)
	})

	ad, err := client.Ads.SetStatus(context.Background(), "ad_1", "ws_1", AdStatusActive)
	if err != nil {
		t.Fatalf("Ads.SetStatus: %v", err)
	}
	if method != "PATCH" || path != "/ads/ad_1" || query != "workspace_id=ws_1" {
		t.Fatalf("%s %s?%s", method, path, query)
	}
	if body["status"] != "active" {
		t.Fatalf("body = %v", body)
	}
	if ad.Status != AdStatusActive {
		t.Fatalf("ad = %+v", ad)
	}
}

func TestAdsAudiencesDecodesPixels(t *testing.T) {
	var query string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		_, _ = io.WriteString(w, `{"data":{"audiences":[{"id":"aud_1","name":"Buyers","subtype":"CUSTOM","sizeLower":1000,"sizeUpper":null}],"pixels":[{"id":"px_1","name":"Site"}],"workspaceId":"ws_1"}}`)
	})

	result, err := client.Ads.Audiences(context.Background(), &ListAudiencesParams{
		WorkspaceID:  "ws_1",
		ConnectionID: "conn_1",
		AdAccountID:  "act_1",
	})
	if err != nil {
		t.Fatalf("Ads.Audiences: %v", err)
	}
	if query != "ad_account_id=act_1&connection_id=conn_1&workspace_id=ws_1" {
		t.Fatalf("query = %q", query)
	}
	if len(result.Audiences) != 1 || *result.Audiences[0].SizeLower != 1000 || result.Audiences[0].SizeUpper != nil {
		t.Fatalf("audiences = %+v", result.Audiences)
	}
	if len(result.Pixels) != 1 || result.Pixels[0].ID != "px_1" {
		t.Fatalf("pixels = %+v", result.Pixels)
	}
}

func TestAdsLeadsPassesTheCursor(t *testing.T) {
	var path, query string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		path, query = r.URL.Path, r.URL.RawQuery
		_, _ = io.WriteString(w, `{"data":{"leads":[{"id":"lead_1","fields":[{"name":"email","values":["jamie@yourbrand.com"]}],"isOrganic":false}],"nextCursor":"cursor_2"}}`)
	})

	page, err := client.Ads.Leads(context.Background(), "form_1", &ListLeadsParams{
		ConnectionID: "conn_1",
		PageID:       "123",
		After:        "cursor_1",
	})
	if err != nil {
		t.Fatalf("Ads.Leads: %v", err)
	}
	if path != "/ads/lead-forms/form_1/leads" || query != "after=cursor_1&connection_id=conn_1&page_id=123" {
		t.Fatalf("%s?%s", path, query)
	}
	if len(page.Leads) != 1 || page.Leads[0].Fields[0].Values[0] != "jamie@yourbrand.com" || page.NextCursor != "cursor_2" {
		t.Fatalf("page = %+v", page)
	}
}

func TestValidatePostSendsBodyToValidatePost(t *testing.T) {
	var path, method string
	var body map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		path, method = r.URL.Path, r.Method
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = io.WriteString(w, `{"data":{"ready":false,"platforms":[{"platform":"twitter","ready":false,"issues":["too long"],"score":42,"signals":[{"level":"warn","code":"over_length","message":"Over the limit"}]},{"platform":"linkedin","ready":true,"issues":[],"signals":[]}]}}`)
	})

	size := int64(1024)
	out, err := client.Validate.Post(context.Background(), &ValidatePostRequest{
		Content:   "Hello",
		Media:     []ValidateMediaItem{{URL: "https://yourbrand.com/a.png", MimeType: "image/png", Size: &size}},
		Platforms: []string{"twitter", "linkedin"},
	})
	if err != nil {
		t.Fatalf("Validate.Post: %v", err)
	}
	if method != "POST" || path != "/validate/post" {
		t.Fatalf("%s %s", method, path)
	}
	if body["content"] != "Hello" {
		t.Fatalf("body = %v", body)
	}
	platforms, _ := body["platforms"].([]any)
	if len(platforms) != 2 || platforms[0] != "twitter" {
		t.Fatalf("platforms = %v", body["platforms"])
	}
	media, _ := body["media"].([]any)
	item, _ := media[0].(map[string]any)
	if item["url"] != "https://yourbrand.com/a.png" || item["mime_type"] != "image/png" || item["size"] != float64(1024) {
		t.Fatalf("media = %v", body["media"])
	}
	if out.Ready || len(out.Platforms) != 2 || out.Platforms[0].Score == nil || *out.Platforms[0].Score != 42 {
		t.Fatalf("out = %+v", out)
	}
	if out.Platforms[1].Score != nil || out.Platforms[0].Signals[0].Code != "over_length" {
		t.Fatalf("out = %+v", out)
	}
}

func TestValidateLengthSendsBodyToValidateLength(t *testing.T) {
	var path, method string
	var body map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		path, method = r.URL.Path, r.Method
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = io.WriteString(w, `{"data":{"ok":true,"platforms":[{"platform":"twitter","length":5,"limit":280,"unit":"chars","ok":true,"signals":[]},{"platform":"linkedin","length":5,"limit":null,"unit":"chars","ok":true,"signals":[]}]}}`)
	})

	out, err := client.Validate.Length(context.Background(), &ValidateLengthRequest{
		Text:      "Hello",
		Platforms: []string{"twitter", "linkedin"},
	})
	if err != nil {
		t.Fatalf("Validate.Length: %v", err)
	}
	if method != "POST" || path != "/validate/length" {
		t.Fatalf("%s %s", method, path)
	}
	platforms, _ := body["platforms"].([]any)
	if body["text"] != "Hello" || len(platforms) != 2 {
		t.Fatalf("body = %v", body)
	}
	if !out.OK || len(out.Platforms) != 2 || out.Platforms[0].Limit == nil || *out.Platforms[0].Limit != 280 {
		t.Fatalf("out = %+v", out)
	}
	if out.Platforms[1].Limit != nil || out.Platforms[1].Unit != "chars" {
		t.Fatalf("out = %+v", out)
	}
}

func TestValidateMediaSendsURLToValidateMedia(t *testing.T) {
	var path, method string
	var body map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		path, method = r.URL.Path, r.Method
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = io.WriteString(w, `{"data":{"ok":false,"issues":["file too large"],"name":"big.mp4","size":99}}`)
	})

	out, err := client.Validate.Media(context.Background(), "https://yourbrand.com/big.mp4")
	if err != nil {
		t.Fatalf("Validate.Media: %v", err)
	}
	if method != "POST" || path != "/validate/media" {
		t.Fatalf("%s %s", method, path)
	}
	if body["url"] != "https://yourbrand.com/big.mp4" || len(body) != 1 {
		t.Fatalf("body = %v", body)
	}
	if out.OK || out.Name != "big.mp4" || out.Size != 99 || out.MimeType != "" || len(out.Issues) != 1 {
		t.Fatalf("out = %+v", out)
	}
}

func TestAccountGroupsCreateSendsBodyAndDecodesTheGroup(t *testing.T) {
	var method, path string
	var body map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"data":{"id":"grp_1","name":"Clients","account_ids":["acc_1"]}}`)
	})

	group, err := client.AccountGroups.Create(context.Background(), &CreateAccountGroupRequest{
		WorkspaceID: "ws_1", Name: "Clients", AccountIDs: []string{"acc_1"},
	})
	if err != nil {
		t.Fatalf("AccountGroups.Create: %v", err)
	}
	if method != "POST" || path != "/account-groups" {
		t.Fatalf("%s %s", method, path)
	}
	if body["workspace_id"] != "ws_1" || body["name"] != "Clients" {
		t.Fatalf("body = %v", body)
	}
	if group.ID != "grp_1" || len(group.AccountIDs) != 1 {
		t.Fatalf("group = %+v", group)
	}
}

func TestAccountGroupsSetMembersSendsAnEmptyListNotNull(t *testing.T) {
	var method, path, raw string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		b, _ := io.ReadAll(r.Body)
		raw = string(b)
		_, _ = io.WriteString(w, `{"data":{"id":"grp_1","account_ids":[]}}`)
	})

	if _, err := client.AccountGroups.SetMembers(context.Background(), "grp_1", nil); err != nil {
		t.Fatalf("AccountGroups.SetMembers: %v", err)
	}
	if method != "PUT" || path != "/account-groups/grp_1/members" {
		t.Fatalf("%s %s", method, path)
	}
	if raw != `{"account_ids":[]}` {
		t.Fatalf("body = %s", raw)
	}
}

func TestAccountsListWithParamsSendsGroupID(t *testing.T) {
	var query string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		_, _ = io.WriteString(w, `{"data":[{"id":"acc_1","name":"Brand","platformName":"acme"}]}`)
	})

	accounts, err := client.Accounts.ListWithParams(context.Background(), &ListAccountsParams{GroupID: "grp_1"})
	if err != nil {
		t.Fatalf("Accounts.ListWithParams: %v", err)
	}
	if query != "group_id=grp_1" {
		t.Fatalf("query = %q", query)
	}
	if accounts[0].PlatformName != "acme" || accounts[0].Name != "Brand" {
		t.Fatalf("accounts = %+v", accounts)
	}
}

func TestAccountsRenameSendsNullToRestore(t *testing.T) {
	var method, path, raw string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		b, _ := io.ReadAll(r.Body)
		raw = string(b)
		_, _ = io.WriteString(w, `{"data":{"id":"acc_1","name":"acme","platform_name":"acme"}}`)
	})

	renamed, err := client.Accounts.Rename(context.Background(), "acc_1", "")
	if err != nil {
		t.Fatalf("Accounts.Rename: %v", err)
	}
	if method != "PATCH" || path != "/accounts/acc_1" || raw != `{"display_name":null}` {
		t.Fatalf("%s %s %s", method, path, raw)
	}
	if renamed.PlatformName != "acme" {
		t.Fatalf("renamed = %+v", renamed)
	}
}

func TestAccountsMoveSurfacesBlockingTables(t *testing.T) {
	var method, path string
	var body map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.WriteHeader(http.StatusConflict)
		_, _ = io.WriteString(w, `{"error":"move_blocked","message":"blocked","blocking_tables":["posts"]}`)
	})

	_, err := client.Accounts.Move(context.Background(), "acc_1", "ws_2")
	if method != "POST" || path != "/accounts/acc_1/move" || body["workspace_id"] != "ws_2" {
		t.Fatalf("%s %s %v", method, path, body)
	}
	if !IsConflict(err) || CodeOf(err) != "move_blocked" {
		t.Fatalf("err = %v", err)
	}
	var tables []string
	if apiErr, ok := APIError(err); !ok || apiErr.Field("blocking_tables", &tables) != nil || len(tables) != 1 {
		t.Fatalf("blocking_tables = %v", tables)
	}
}
