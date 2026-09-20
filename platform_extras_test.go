package fopost

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestCreatePinterestBoardSendsOnlyWhatWasGiven(t *testing.T) {
	var method, path, raw string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		b, _ := io.ReadAll(r.Body)
		raw = string(b)
		_, _ = io.WriteString(w, `{"data":{"id":"b1","name":"Recipes","privacy":"PUBLIC","description":null,"image":null}}`)
	})

	board, err := client.Accounts.CreatePinterestBoard(context.Background(), "acc_1", &CreatePinterestBoardRequest{Name: "Recipes"})
	if err != nil {
		t.Fatalf("CreatePinterestBoard: %v", err)
	}
	if method != "POST" || path != "/accounts/acc_1/pinterest/boards" {
		t.Fatalf("method = %q path = %q", method, path)
	}
	if strings.Contains(raw, "description") || strings.Contains(raw, "privacy") {
		t.Fatalf("empty optionals reached the wire: %s", raw)
	}
	if board.ID != "b1" {
		t.Fatalf("board = %+v", board)
	}
}

func TestSetDefaultYouTubePlaylistSendsNullToClear(t *testing.T) {
	var raw string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		raw = string(b)
		_, _ = io.WriteString(w, `{"data":{"playlist_id":null}}`)
	})

	stored, err := client.Accounts.SetDefaultYouTubePlaylist(context.Background(), "acc_1", "")
	if err != nil {
		t.Fatalf("SetDefaultYouTubePlaylist: %v", err)
	}
	if !strings.Contains(raw, `"playlist_id":null`) {
		t.Fatalf("body = %s", raw)
	}
	if stored != "" {
		t.Fatalf("stored = %q", stored)
	}
}

func TestBlueskyLanguagesRoundTrip(t *testing.T) {
	var method, raw string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		b, _ := io.ReadAll(r.Body)
		raw = string(b)
		_, _ = io.WriteString(w, `{"data":{"languages":["en","pt-BR"]}}`)
	})

	result, err := client.Accounts.SetBlueskyLanguages(context.Background(), "acc_1", []string{"en", "pt-BR"})
	if err != nil {
		t.Fatalf("SetBlueskyLanguages: %v", err)
	}
	if method != "PUT" || !strings.Contains(raw, `["en","pt-BR"]`) {
		t.Fatalf("method = %q body = %s", method, raw)
	}
	if len(result.Languages) != 2 {
		t.Fatalf("languages = %+v", result.Languages)
	}
}

func TestTikTokCreatorInfoReportsTheAccountsOwnSwitches(t *testing.T) {
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"data":{"privacy_level_options":["PUBLIC_TO_EVERYONE"],"comment_disabled":false,"duet_disabled":true,"stitch_disabled":false,"max_video_post_duration_sec":600}}`)
	})

	info, err := client.Accounts.GetTikTokCreatorInfo(context.Background(), "acc_1")
	if err != nil {
		t.Fatalf("GetTikTokCreatorInfo: %v", err)
	}
	if !info.DuetDisabled || info.StitchDisabled {
		t.Fatalf("info = %+v", info)
	}
	if info.MaxVideoPostDurationSec == nil || *info.MaxVideoPostDurationSec != 600 {
		t.Fatalf("duration = %+v", info.MaxVideoPostDurationSec)
	}
}

func TestInstagramStoriesAskForInsightsOnlyWhenRequested(t *testing.T) {
	var query string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		_, _ = io.WriteString(w, `{"data":[{"id":"s1","media_type":"IMAGE","insights":{"views":40}}]}`)
	})
	ctx := context.Background()

	if _, err := client.Accounts.ListInstagramStories(ctx, "acc_1", false); err != nil {
		t.Fatalf("ListInstagramStories: %v", err)
	}
	if query != "" {
		t.Fatalf("query = %q", query)
	}

	stories, err := client.Accounts.ListInstagramStories(ctx, "acc_1", true)
	if err != nil {
		t.Fatalf("ListInstagramStories: %v", err)
	}
	if query != "insights=true" {
		t.Fatalf("query = %q", query)
	}
	if stories[0].Insights["views"] != 40 {
		t.Fatalf("stories = %+v", stories)
	}
}

func TestSearchLinkedInMentionsCarriesTheAnnotation(t *testing.T) {
	var query string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		_, _ = io.WriteString(w, `{"data":[{"urn":"urn:li:organization:2414183","name":"Devtestco","vanity_name":"devtestco","logo_url":null,"type":"none","annotation":"@[Devtestco](urn:li:organization:2414183)"}]}`)
	})

	mentions, err := client.Accounts.SearchLinkedInMentions(context.Background(), "acc_1", "devtestco")
	if err != nil {
		t.Fatalf("SearchLinkedInMentions: %v", err)
	}
	if query != "q=devtestco" {
		t.Fatalf("query = %q", query)
	}
	if mentions[0].Annotation != "@[Devtestco](urn:li:organization:2414183)" {
		t.Fatalf("mentions = %+v", mentions)
	}
}
