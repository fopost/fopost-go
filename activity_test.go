package fopost

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestActivityListReadsTheAuditLogAndKeepsTheCursor(t *testing.T) {
	var query string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		_, _ = io.WriteString(w, `{"data":[{"id":"evt_1","workspace_id":"ws_1","kind":"security","ref_type":"member_removed","ref_id":"usr_2","summary":"Removed sam@example.com","actor":{"type":"user","name":"Ada"},"time":"2026-09-20T10:00:00Z"}],"meta":{"next_cursor":"42"}}`)
	})

	page, err := client.Activity.List(context.Background(), &ListActivityParams{
		WorkspaceID: "ws_1",
		Kind:        ActivityKindSecurity,
		Limit:       1,
	})
	if err != nil {
		t.Fatalf("Activity.List: %v", err)
	}
	if !strings.Contains(query, "kind=security") || !strings.Contains(query, "workspace_id=ws_1") {
		t.Fatalf("query = %q", query)
	}
	if len(page.Events) != 1 {
		t.Fatalf("events = %+v", page.Events)
	}
	if page.Events[0].RefType != "member_removed" || page.Events[0].Actor.Name != "Ada" {
		t.Fatalf("event = %+v", page.Events[0])
	}
	if page.NextCursor != "42" {
		t.Fatalf("next cursor = %q", page.NextCursor)
	}
}

func TestActivityListEndsWithAnEmptyCursor(t *testing.T) {
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"data":[],"meta":{"next_cursor":null}}`)
	})

	page, err := client.Activity.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("Activity.List: %v", err)
	}
	if len(page.Events) != 0 || page.NextCursor != "" {
		t.Fatalf("page = %+v", page)
	}
}
