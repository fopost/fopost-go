package fopost

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestBlogSettingsPatchKeepsTheSecretBesideTheData(t *testing.T) {
	var method, path string
	var body map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = io.WriteString(w, `{"site_url":null,"topic":null,"audience":null,"goal":"leads","focus_topics":"trail gear","avoid_topics":"politics","language":"en","location_code":2840,"cadence_per_week":2,"review_hold_days":3,"image_tier":"standard","in_body_images":1,"min_words":1200,"max_words":1800,"brand_voice_id":null,"target_kind":"website","target_account_id":null,"target_platform":null,"website_url":"https://yourbrand.com/fopost","has_website_secret":true,"plan_run":null,"updated_at":"2026-10-11T09:00:00.000Z","website_secret":"whsec_first"}`)
	})

	updated, err := client.Blog.UpdateSettings(context.Background(), "ws_1", &UpdateBlogSettingsRequest{
		TargetKind:     String(BlogTargetWebsite),
		WebsiteURL:     String("https://yourbrand.com/fopost"),
		CadencePerWeek: Int(2),
		Goal:           String(BlogGoalLeads),
		FocusTopics:    String("trail gear"),
		AvoidTopics:    String("politics"),
	})
	if err != nil {
		t.Fatalf("Blog.UpdateSettings: %v", err)
	}
	if method != http.MethodPatch || path != "/blog/settings" {
		t.Fatalf("%s %s", method, path)
	}
	if body["workspaceId"] != "ws_1" || body["targetKind"] != "website" || body["cadencePerWeek"] != float64(2) {
		t.Fatalf("body = %+v", body)
	}
	if body["goal"] != "leads" || body["focusTopics"] != "trail gear" || body["avoidTopics"] != "politics" {
		t.Fatalf("goal fields = %+v", body)
	}
	if _, ok := body["topic"]; ok {
		t.Fatalf("topic should be omitted: %+v", body)
	}
	if updated.WebsiteSecret == nil || *updated.WebsiteSecret != "whsec_first" {
		t.Fatalf("websiteSecret = %v", updated.WebsiteSecret)
	}
	if !updated.Settings.HasWebsiteSecret || *updated.Settings.TargetKind != BlogTargetWebsite || updated.Settings.CadencePerWeek != 2 {
		t.Fatalf("settings = %+v", updated.Settings)
	}
	if *updated.Settings.Goal != BlogGoalLeads || *updated.Settings.FocusTopics != "trail gear" || updated.Settings.PlanRun != nil {
		t.Fatalf("goal settings = %+v", updated.Settings)
	}
	if updated.Settings.LocationCode != 2840 || updated.Settings.MaxWords != 1800 || updated.Settings.UpdatedAt.IsZero() {
		t.Fatalf("snake_case settings did not decode: %+v", updated.Settings)
	}
	out, _ := json.Marshal(updated)
	var flat map[string]any
	_ = json.Unmarshal(out, &flat)
	if flat["website_secret"] != "whsec_first" || flat["cadence_per_week"] != float64(2) {
		t.Fatalf("marshalled = %s", out)
	}
}

func TestBlogSettingsReadAndSecretRoutes(t *testing.T) {
	var calls []string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		calls = append(calls, r.Method+" "+r.URL.Path+"?"+r.URL.RawQuery+" ws="+asString(body["workspaceId"]))
		switch r.URL.Path {
		case "/blog/settings":
			_, _ = io.WriteString(w, `{"goal":null,"language":"en","cadence_per_week":0,"review_hold_days":3,"has_website_secret":false,"website_url":null,"plan_run":{"status":"done","error":null,"added":12,"finished_at":"2026-10-17T08:00:00.000Z"},"keyword_data_available":true}`)
		case "/blog/settings/website-secret":
			_, _ = io.WriteString(w, `{"website_secret":"whsec_rotated"}`)
		case "/blog/settings/website-test":
			_, _ = io.WriteString(w, `{"http_status":204}`)
		}
	})

	settings, err := client.Blog.GetSettings(context.Background(), "ws_1")
	if err != nil {
		t.Fatalf("Blog.GetSettings: %v", err)
	}
	if settings.WebsiteURL != nil || settings.HasWebsiteSecret || settings.Goal != nil || !settings.KeywordDataAvailable {
		t.Fatalf("settings = %+v", settings)
	}
	if run := settings.PlanRun; run == nil || run.Status != "done" || run.Error != nil || *run.Added != 12 || run.FinishedAt == nil {
		t.Fatalf("planRun = %+v", settings.PlanRun)
	}
	secret, err := client.Blog.RotateWebsiteSecret(context.Background(), "ws_1")
	if err != nil {
		t.Fatalf("Blog.RotateWebsiteSecret: %v", err)
	}
	if secret != "whsec_rotated" {
		t.Fatalf("secret = %q", secret)
	}
	result, err := client.Blog.TestWebsite(context.Background(), "ws_1")
	if err != nil {
		t.Fatalf("Blog.TestWebsite: %v", err)
	}
	if result.HTTPStatus != 204 {
		t.Fatalf("result = %+v", result)
	}

	want := []string{
		"GET /blog/settings?workspace_id=ws_1 ws=",
		"POST /blog/settings/website-secret? ws=ws_1",
		"POST /blog/settings/website-test? ws=ws_1",
	}
	for i := range want {
		if calls[i] != want[i] {
			t.Fatalf("call %d = %q, want %q", i, calls[i], want[i])
		}
	}
}

func TestBlogKeywordsRoutesAndBodies(t *testing.T) {
	var calls []string
	var bodies []map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		bodies = append(bodies, body)
		calls = append(calls, r.Method+" "+r.URL.Path+"?"+r.URL.RawQuery)
		switch {
		case r.Method == http.MethodGet:
			_, _ = io.WriteString(w, `{"data":[{"id":"kw_1","keyword":"running shoes","search_volume":1200,"difficulty":34,"cpc":1.5,"intent":"commercial","source":"provider","status":"suggested","planned_for":null,"fit_reason":"Buyers compare models before they sign up","created_at":"2026-10-01T00:00:00.000Z"}]}`)
		case r.URL.Path == "/blog/keywords/plan":
			w.WriteHeader(http.StatusAccepted)
			_, _ = io.WriteString(w, `{"status":"queued"}`)
		case r.Method == http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, `{"id":"kw_2","keyword":"trail shoes","search_volume":null,"difficulty":null,"cpc":null,"intent":null,"source":"user","status":"planned","planned_for":"2026-10-20","fit_reason":null,"created_at":"2026-10-11T00:00:00.000Z"}`)
		case r.Method == http.MethodPatch:
			_, _ = io.WriteString(w, `{"id":"kw_1","keyword":"running shoes","source":"provider","status":"dismissed","planned_for":null,"created_at":"2026-10-01T00:00:00.000Z"}`)
		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})
	ctx := context.Background()

	keywords, err := client.Blog.ListKeywords(ctx, "ws_1", &ListBlogKeywordsParams{
		Status: []string{BlogKeywordStatusSuggested, BlogKeywordStatusPlanned},
		Limit:  50,
	})
	if err != nil {
		t.Fatalf("Blog.ListKeywords: %v", err)
	}
	if len(keywords) != 1 || *keywords[0].SearchVolume != 1200 || *keywords[0].CPC != 1.5 || keywords[0].PlannedFor != nil {
		t.Fatalf("keywords = %+v", keywords)
	}
	if keywords[0].FitReason == nil || *keywords[0].FitReason != "Buyers compare models before they sign up" {
		t.Fatalf("fitReason = %v", keywords[0].FitReason)
	}
	if err := client.Blog.PlanKeywords(ctx, "ws_1", []string{"running"}); err != nil {
		t.Fatalf("Blog.PlanKeywords: %v", err)
	}
	added, err := client.Blog.AddKeyword(ctx, "ws_1", &AddBlogKeywordRequest{Keyword: "trail shoes", PlannedFor: "2026-10-20"})
	if err != nil {
		t.Fatalf("Blog.AddKeyword: %v", err)
	}
	if added.ID != "kw_2" || *added.PlannedFor != "2026-10-20" || added.CreatedAt.IsZero() {
		t.Fatalf("added = %+v", added)
	}
	updated, err := client.Blog.UpdateKeyword(ctx, "ws_1", "kw_1", &UpdateBlogKeywordRequest{Status: BlogKeywordStatusDismissed})
	if err != nil {
		t.Fatalf("Blog.UpdateKeyword: %v", err)
	}
	if updated.Status != BlogKeywordStatusDismissed {
		t.Fatalf("updated = %+v", updated)
	}
	if err := client.Blog.DeleteKeyword(ctx, "ws_1", "kw_1"); err != nil {
		t.Fatalf("Blog.DeleteKeyword: %v", err)
	}

	want := []string{
		"GET /blog/keywords?limit=50&status=suggested%2Cplanned&workspace_id=ws_1",
		"POST /blog/keywords/plan?",
		"POST /blog/keywords?",
		"PATCH /blog/keywords/kw_1?workspace_id=ws_1",
		"DELETE /blog/keywords/kw_1?workspace_id=ws_1",
	}
	for i := range want {
		if calls[i] != want[i] {
			t.Fatalf("call %d = %q, want %q", i, calls[i], want[i])
		}
	}
	if bodies[1]["workspaceId"] != "ws_1" || bodies[1]["seeds"].([]any)[0] != "running" {
		t.Fatalf("plan body = %+v", bodies[1])
	}
	if bodies[2]["workspaceId"] != "ws_1" || bodies[2]["keyword"] != "trail shoes" || bodies[2]["plannedFor"] != "2026-10-20" {
		t.Fatalf("add body = %+v", bodies[2])
	}
	if bodies[3]["status"] != "dismissed" {
		t.Fatalf("patch body = %+v", bodies[3])
	}
	if _, ok := bodies[3]["workspaceId"]; ok {
		t.Fatalf("patch body carries the workspace on the query, not the body: %+v", bodies[3])
	}
	if _, ok := bodies[3]["plannedFor"]; ok {
		t.Fatalf("plannedFor should be omitted: %+v", bodies[3])
	}
}

func TestBlogArticlesRoutesAndBodies(t *testing.T) {
	var calls []string
	var bodies []map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		bodies = append(bodies, body)
		calls = append(calls, r.Method+" "+r.URL.Path+"?"+r.URL.RawQuery)
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/blog/articles":
			_, _ = io.WriteString(w, `{"data":[{"id":"art_1","title":"Trail shoes","slug":"trail-shoes","keyword":"trail shoes","excerpt":null,"meta_title":"Trail shoes","meta_description":null,"tags":["gear"],"word_count":1200,"status":"draft","origin":"plan","featured_image":{"media_id":"m_1","url":"https://cdn.yourbrand.com/1.jpg","alt":"Shoes"},"review_until":"2026-10-14T00:00:00.000Z","publish_at":null,"published_at":null,"published_url":null,"error":null,"created_at":"2026-10-11T00:00:00.000Z","updated_at":"2026-10-11T00:00:00.000Z"}]}`)
		case r.Method == http.MethodGet:
			_, _ = io.WriteString(w, `{"id":"art_1","title":"Trail shoes","slug":"trail-shoes","tags":[],"word_count":1200,"status":"draft","origin":"plan","created_at":"2026-10-11T00:00:00.000Z","updated_at":"2026-10-11T00:00:00.000Z","instructions":null,"outline":["Why","How"],"brief":{"questions":["Which fits?"],"search_volume":1200},"body_markdown":"# Trail shoes","body_html":"<h1>Trail shoes</h1>","images":[{"media_id":"m_2","url":"https://cdn.yourbrand.com/2.jpg","alt":"Trail","position":1}]}`)
		case r.Method == http.MethodPost && r.URL.Path == "/blog/articles":
			w.WriteHeader(http.StatusAccepted)
			_, _ = io.WriteString(w, `{"id":"art_2","title":"Running shoes","slug":"running-shoes","tags":[],"word_count":0,"status":"queued","origin":"user","created_at":"2026-10-11T00:00:00.000Z","updated_at":"2026-10-11T00:00:00.000Z"}`)
		case r.Method == http.MethodPatch:
			_, _ = io.WriteString(w, `{"id":"art_1","title":"Trail shoes, ranked","slug":"trail-shoes","tags":["gear"],"word_count":1200,"status":"scheduled","origin":"plan","publish_at":"2026-10-15T09:00:00.000Z","created_at":"2026-10-11T00:00:00.000Z","updated_at":"2026-10-11T00:00:00.000Z","body_markdown":"# Trail shoes","body_html":"","images":[]}`)
		case r.Method == http.MethodPost:
			w.WriteHeader(http.StatusAccepted)
			_, _ = io.WriteString(w, `{"id":"art_1","title":"Trail shoes","slug":"trail-shoes","tags":[],"word_count":1200,"status":"publishing","origin":"plan","created_at":"2026-10-11T00:00:00.000Z","updated_at":"2026-10-11T00:00:00.000Z"}`)
		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})
	ctx := context.Background()

	articles, err := client.Blog.ListArticles(ctx, "ws_1", &ListBlogArticlesParams{Status: []string{BlogArticleStatusDraft}})
	if err != nil {
		t.Fatalf("Blog.ListArticles: %v", err)
	}
	if len(articles) != 1 || articles[0].FeaturedImage.MediaID != "m_1" || articles[0].PublishAt != nil || articles[0].ReviewUntil == nil || articles[0].WordCount != 1200 || *articles[0].MetaTitle != "Trail shoes" {
		t.Fatalf("articles = %+v", articles)
	}
	requested, err := client.Blog.RequestArticle(ctx, "ws_1", &RequestBlogArticleRequest{Keyword: "running shoes", Instructions: "Compare three models"})
	if err != nil {
		t.Fatalf("Blog.RequestArticle: %v", err)
	}
	if requested.Status != BlogArticleStatusQueued {
		t.Fatalf("requested = %+v", requested)
	}
	full, err := client.Blog.GetArticle(ctx, "ws_1", "art_1")
	if err != nil {
		t.Fatalf("Blog.GetArticle: %v", err)
	}
	if full.BodyMarkdown != "# Trail shoes" || full.BodyHTML != "<h1>Trail shoes</h1>" || len(full.Outline) != 2 || string(full.Brief) != `{"questions":["Which fits?"],"search_volume":1200}` || full.Images[0].MediaID != "m_2" || *full.Images[0].Position != 1 {
		t.Fatalf("full = %+v", full)
	}
	edited, err := client.Blog.UpdateArticle(ctx, "ws_1", "art_1", &UpdateBlogArticleRequest{
		Title:     "Trail shoes, ranked",
		PublishAt: NewTime(time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC)),
	})
	if err != nil {
		t.Fatalf("Blog.UpdateArticle: %v", err)
	}
	if edited.Status != BlogArticleStatusScheduled || edited.PublishAt == nil || edited.Title != "Trail shoes, ranked" {
		t.Fatalf("edited = %+v", edited)
	}
	for name, action := range map[string]func(context.Context, string, string) (*BlogArticle, error){
		"regenerate": client.Blog.RegenerateArticle,
		"publish":    client.Blog.PublishArticle,
		"reject":     client.Blog.RejectArticle,
	} {
		if _, err := action(ctx, "ws_1", "art_1"); err != nil {
			t.Fatalf("Blog.%s: %v", name, err)
		}
	}
	if err := client.Blog.DeleteArticle(ctx, "ws_1", "art_1"); err != nil {
		t.Fatalf("Blog.DeleteArticle: %v", err)
	}

	for i, want := range []string{
		"GET /blog/articles?status=draft&workspace_id=ws_1",
		"POST /blog/articles?",
		"GET /blog/articles/art_1?workspace_id=ws_1",
		"PATCH /blog/articles/art_1?workspace_id=ws_1",
	} {
		if calls[i] != want {
			t.Fatalf("call %d = %q, want %q", i, calls[i], want)
		}
	}
	seen := map[string]bool{}
	for _, call := range calls[4:7] {
		seen[call] = true
	}
	for _, want := range []string{
		"POST /blog/articles/art_1/regenerate?workspace_id=ws_1",
		"POST /blog/articles/art_1/publish?workspace_id=ws_1",
		"POST /blog/articles/art_1/reject?workspace_id=ws_1",
	} {
		if !seen[want] {
			t.Fatalf("missing call %q in %v", want, calls[4:7])
		}
	}
	if calls[7] != "DELETE /blog/articles/art_1?workspace_id=ws_1" {
		t.Fatalf("delete call = %q", calls[7])
	}
	if bodies[1]["workspaceId"] != "ws_1" || bodies[1]["keyword"] != "running shoes" || bodies[1]["instructions"] != "Compare three models" {
		t.Fatalf("request body = %+v", bodies[1])
	}
	if _, ok := bodies[1]["title"]; ok {
		t.Fatalf("title should be omitted: %+v", bodies[1])
	}
	if bodies[3]["title"] != "Trail shoes, ranked" || bodies[3]["publishAt"] != "2026-10-15T09:00:00Z" {
		t.Fatalf("patch body = %+v", bodies[3])
	}
	if _, ok := bodies[3]["workspaceId"]; ok {
		t.Fatalf("patch body carries the workspace on the query, not the body: %+v", bodies[3])
	}
}

func TestBlogSuggestSeedsSendsSeedsAndPrompt(t *testing.T) {
	var method, path, query string
	var body map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, path, query = r.Method, r.URL.Path, r.URL.RawQuery
		body = nil
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = io.WriteString(w, `{"seeds":["trail running shoes","best trail shoes 2026"]}`)
	})

	seeds, err := client.Blog.SuggestSeeds(context.Background(), "ws_1", &SuggestBlogSeedsRequest{
		Seeds:  []string{"trail shoes"},
		Prompt: "Lean towards beginners",
	})
	if err != nil {
		t.Fatalf("Blog.SuggestSeeds: %v", err)
	}
	if method != http.MethodPost || path != "/blog/keywords/suggest" || query != "" {
		t.Fatalf("%s %s?%s", method, path, query)
	}
	if body["workspaceId"] != "ws_1" || body["seeds"].([]any)[0] != "trail shoes" || body["prompt"] != "Lean towards beginners" {
		t.Fatalf("body = %+v", body)
	}
	if len(seeds) != 2 || seeds[0] != "trail running shoes" {
		t.Fatalf("seeds = %v", seeds)
	}

	if _, err := client.Blog.SuggestSeeds(context.Background(), "ws_1", nil); err != nil {
		t.Fatalf("Blog.SuggestSeeds(nil): %v", err)
	}
	if _, ok := body["seeds"]; ok {
		t.Fatalf("seeds should be omitted: %+v", body)
	}
	if _, ok := body["prompt"]; ok {
		t.Fatalf("prompt should be omitted: %+v", body)
	}
}

func TestBlogSubmitArticleMovesItToReview(t *testing.T) {
	var call string
	var hadBody bool
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		call = r.Method + " " + r.URL.Path + "?" + r.URL.RawQuery
		raw, _ := io.ReadAll(r.Body)
		hadBody = len(raw) > 0
		if r.URL.Path == "/blog/articles/art_2/submit" {
			w.WriteHeader(http.StatusConflict)
			_, _ = io.WriteString(w, `{"error":"invalid_status","message":"Only a draft, failed or rejected article can be submitted"}`)
			return
		}
		_, _ = io.WriteString(w, `{"id":"art_1","title":"Trail shoes","slug":"trail-shoes","tags":[],"word_count":1200,"status":"in_review","origin":"plan","created_at":"2026-10-11T00:00:00.000Z","updated_at":"2026-10-11T00:00:00.000Z"}`)
	})
	ctx := context.Background()

	article, err := client.Blog.SubmitArticle(ctx, "ws_1", "art_1")
	if err != nil {
		t.Fatalf("Blog.SubmitArticle: %v", err)
	}
	if call != "POST /blog/articles/art_1/submit?workspace_id=ws_1" || hadBody {
		t.Fatalf("call = %q, body sent = %v", call, hadBody)
	}
	if article.Status != BlogArticleStatusInReview {
		t.Fatalf("article = %+v", article)
	}
	if _, err := client.Blog.SubmitArticle(ctx, "ws_1", "art_2"); !IsConflict(err) {
		t.Fatalf("expected a 409, got %v", err)
	}
}

func TestBlogCreditAndTargetErrorsMapToPredicates(t *testing.T) {
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/blog/articles" {
			w.WriteHeader(http.StatusPaymentRequired)
			_, _ = io.WriteString(w, `{"error":"insufficient_credits","message":"Balance ran out","upgrade_url":"https://app.yourbrand.com/billing"}`)
			return
		}
		w.WriteHeader(http.StatusConflict)
		_, _ = io.WriteString(w, `{"error":"no_target","message":"Choose where articles publish first"}`)
	})
	ctx := context.Background()

	_, err := client.Blog.RequestArticle(ctx, "ws_1", &RequestBlogArticleRequest{Title: "Hello"})
	if !IsPaymentRequired(err) || CodeOf(err) != "insufficient_credits" {
		t.Fatalf("expected a 402 insufficient_credits, got %v", err)
	}
	apiErr, _ := APIError(err)
	if apiErr.UpgradeURL() != "https://app.yourbrand.com/billing" {
		t.Fatalf("upgrade_url = %q", apiErr.UpgradeURL())
	}
	_, err = client.Blog.PublishArticle(ctx, "ws_1", "art_1")
	if !IsConflict(err) || CodeOf(err) != "no_target" {
		t.Fatalf("expected a 409 no_target, got %v", err)
	}
}

func TestBlogWebsiteTestFailureCarriesTheHTTPStatus(t *testing.T) {
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = io.WriteString(w, `{"error":"website_test_failed","message":"The website answered 500","http_status":500}`)
	})

	_, err := client.Blog.TestWebsite(context.Background(), "ws_1")
	apiErr, ok := APIError(err)
	if !ok || apiErr.Status != http.StatusBadGateway || CodeOf(err) != "website_test_failed" {
		t.Fatalf("expected a 502 website_test_failed, got %v", err)
	}
	var status int
	if err := apiErr.Field("http_status", &status); err != nil || status != 500 {
		t.Fatalf("http_status = %d (%v)", status, err)
	}
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}
