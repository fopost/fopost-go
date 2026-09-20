package fopost

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestKnowledgeListDecodesCamelCaseFields(t *testing.T) {
	var query string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		_, _ = io.WriteString(w, `{"data":[{"id":"know_1","kind":"url","title":"Refund policy","status":"ready","url":"https://yourbrand.com/help/refunds","chunkCount":3,"lastSyncedAt":"2026-09-20T00:00:00.000Z"}]}`)
	})

	sources, err := client.Knowledge.List(context.Background(), "ws_1")
	if err != nil {
		t.Fatalf("Knowledge.List: %v", err)
	}
	if query != "workspace_id=ws_1" {
		t.Fatalf("query = %q", query)
	}
	if len(sources) != 1 || sources[0].ChunkCount != 3 || sources[0].Status != "ready" {
		t.Fatalf("sources = %+v", sources)
	}
}

func TestKnowledgeCreateSendsSnakeCaseBodyAndOmitsUnset(t *testing.T) {
	var body map[string]any
	var path string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = io.WriteString(w, `{"data":{"id":"know_1","kind":"file","title":"Price list","status":"pending"}}`)
	})

	_, err := client.Knowledge.Create(context.Background(), CreateKnowledgeSourceRequest{
		Kind:        "file",
		Title:       "Price list",
		MediaID:     "media_1",
		WorkspaceID: "ws_1",
	})
	if err != nil {
		t.Fatalf("Knowledge.Create: %v", err)
	}
	if path != "/knowledge/sources" {
		t.Fatalf("path = %q", path)
	}
	if body["media_id"] != "media_1" || body["workspace_id"] != "ws_1" {
		t.Fatalf("body = %+v", body)
	}
	// Nothing the caller left out reaches the wire.
	if _, ok := body["url"]; ok {
		t.Fatalf("url should be omitted: %+v", body)
	}
	if _, ok := body["content"]; ok {
		t.Fatalf("content should be omitted: %+v", body)
	}
}

func TestKnowledgeSearchSendsTopKAndDecodesMatches(t *testing.T) {
	var query string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		_, _ = io.WriteString(w, `{"data":[{"sourceId":"know_1","sourceTitle":"Refund policy","sourceKind":"url","text":"We refund within 30 days.","score":0.82}]}`)
	})

	matches, err := client.Knowledge.Search(context.Background(), "refunds", SearchKnowledgeParams{TopK: 3})
	if err != nil {
		t.Fatalf("Knowledge.Search: %v", err)
	}
	if query != "q=refunds&top_k=3" {
		t.Fatalf("query = %q", query)
	}
	if len(matches) != 1 || matches[0].SourceTitle != "Refund policy" || matches[0].Score != 0.82 {
		t.Fatalf("matches = %+v", matches)
	}
}

func TestKnowledgeSyncPostsToTheSyncPath(t *testing.T) {
	var method, path string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		_, _ = io.WriteString(w, `{"data":{"id":"know_1","status":"pending"}}`)
	})

	if err := client.Knowledge.Sync(context.Background(), "know_1"); err != nil {
		t.Fatalf("Knowledge.Sync: %v", err)
	}
	if method != "POST" || path != "/knowledge/sources/know_1/sync" {
		t.Fatalf("%s %s", method, path)
	}
}
