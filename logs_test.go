package fopost

import (
	"context"
	"io"
	"net/http"
	"testing"
)

func TestLogsListDecodesSummaryAndCursor(t *testing.T) {
	var method, path, query string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, path, query = r.Method, r.URL.Path, r.URL.RawQuery
		_, _ = io.WriteString(w, `{"data":[{"id":12,"workspace_id":"ws_1","time":"2026-10-18T09:00:00.000Z","auth_method":"api_key","method":"POST","route":"/v1/posts","path":"/v1/posts","status_code":422,"duration_ms":31,"platform":null,"message":"content is required"}],"meta":{"next_cursor":"11","has_more":true},"summary":{"since":"2026-10-11T09:00:00.000Z","total":40,"errors":3,"buckets":[{"at":"2026-10-18T09:00:00.000Z","total":4,"errors":1}]}}`)
	})

	page, err := client.Logs.List(context.Background(), &ListLogsParams{WorkspaceID: "ws_1", Status: "failed", Days: 7, Limit: 1})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if method != "GET" || path != "/logs" || query != "days=7&limit=1&status=failed&workspace_id=ws_1" {
		t.Fatalf("List sent %s %s?%s", method, path, query)
	}
	if len(page.Data) != 1 || page.Data[0].StatusCode != 422 || *page.Data[0].Message != "content is required" || page.Data[0].Platform != nil {
		t.Fatalf("data = %+v", page.Data)
	}
	if *page.Meta.NextCursor != "11" || !page.Meta.HasMore || page.Summary.Errors != 3 || len(page.Summary.Buckets) != 1 {
		t.Fatalf("meta = %+v summary = %+v", page.Meta, page.Summary)
	}
}
