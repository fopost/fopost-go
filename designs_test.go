package fopost

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

const designJSON = `{"id":"des_1","name":"Launch","kind":"carousel","width":1080,"height":1350,"document":{"pages":[{"id":"p1"}]},"thumbnail_media_id":null,"thumbnail_url":"https://fopost.com/t.png","created_at":"2026-11-20T10:00:00.000Z","updated_at":"2026-11-20T10:00:00.000Z"}`

type capturedRequest struct {
	method string
	path   string
	query  string
	body   map[string]any
}

func captureClient(t *testing.T, status int, response string) (*Client, *capturedRequest) {
	t.Helper()
	got := &capturedRequest{}
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		got.method = r.Method
		got.path = r.URL.Path
		got.query = r.URL.RawQuery
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&got.body)
		}
		w.WriteHeader(status)
		_, _ = io.WriteString(w, response)
	})
	return client, got
}

func TestDesignsListSendsFiltersAndDecodes(t *testing.T) {
	client, got := captureClient(t, 200, `{"data":[`+designJSON+`]}`)
	designs, err := client.Designs.List(context.Background(), &ListDesignsParams{WorkspaceID: "ws_1", Kind: DesignKindCarousel, Limit: 10, Offset: 20})
	if err != nil {
		t.Fatalf("Designs.List: %v", err)
	}
	if got.method != "GET" || got.path != "/designs" || got.query != "kind=carousel&limit=10&offset=20&workspace_id=ws_1" {
		t.Fatalf("request = %+v", got)
	}
	if len(designs) != 1 || designs[0].Kind != "carousel" || designs[0].Width != 1080 || designs[0].ThumbnailMediaID != nil {
		t.Fatalf("designs = %+v", designs)
	}
	if *designs[0].ThumbnailURL != "https://fopost.com/t.png" || designs[0].CreatedAt.IsZero() {
		t.Fatalf("design = %+v", designs[0])
	}
}

func TestDesignsListRequiresWorkspace(t *testing.T) {
	client, _ := captureClient(t, 200, `{}`)
	if _, err := client.Designs.List(context.Background(), &ListDesignsParams{}); err == nil {
		t.Fatal("expected an error without a workspace id")
	}
}

func TestDesignsCreatePassesDocumentThrough(t *testing.T) {
	client, got := captureClient(t, 201, `{"data":`+designJSON+`}`)
	design, err := client.Designs.Create(context.Background(), &CreateDesignRequest{
		WorkspaceID: "ws_1",
		Kind:        DesignKindCarousel,
		Name:        "Launch",
		Document:    DesignDocument{"pages": []any{map[string]any{"id": "p1"}}},
	})
	if err != nil {
		t.Fatalf("Designs.Create: %v", err)
	}
	if got.method != "POST" || got.path != "/designs" {
		t.Fatalf("request = %+v", got)
	}
	if got.body["workspace_id"] != "ws_1" || got.body["kind"] != "carousel" || got.body["document"] == nil {
		t.Fatalf("body = %+v", got.body)
	}
	if _, ok := got.body["brand_voice_id"]; ok {
		t.Fatalf("brand_voice_id should be omitted: %+v", got.body)
	}
	pages, _ := design.Document["pages"].([]any)
	if design.ID != "des_1" || len(pages) != 1 {
		t.Fatalf("design = %+v", design)
	}
}

func TestDesignsGet(t *testing.T) {
	client, got := captureClient(t, 200, `{"data":`+designJSON+`}`)
	design, err := client.Designs.Get(context.Background(), "des 1")
	if err != nil {
		t.Fatalf("Designs.Get: %v", err)
	}
	if got.method != "GET" || got.path != "/designs/des 1" || design.Name != "Launch" {
		t.Fatalf("request = %+v design = %+v", got, design)
	}
}

func TestDesignsUpdateKeepsTheVersionBesideTheDesign(t *testing.T) {
	client, got := captureClient(t, 200, `{"data":`+designJSON+`,"version":{"id":"ver_1","version":3,"created_at":"2026-11-20T10:00:00.000Z"}}`)
	out, err := client.Designs.Update(context.Background(), "des_1", &UpdateDesignRequest{
		Name:          String("Launch v2"),
		Document:      DesignDocument{"pages": []any{}},
		CreateVersion: Bool(true),
	})
	if err != nil {
		t.Fatalf("Designs.Update: %v", err)
	}
	if got.method != "PATCH" || got.path != "/designs/des_1" {
		t.Fatalf("request = %+v", got)
	}
	if got.body["name"] != "Launch v2" || got.body["create_version"] != true {
		t.Fatalf("body = %+v", got.body)
	}
	if _, ok := got.body["kind"]; ok {
		t.Fatalf("kind should be omitted: %+v", got.body)
	}
	if out.Design.ID != "des_1" || out.Version == nil || out.Version.Version != 3 {
		t.Fatalf("update = %+v", out)
	}
}

func TestDesignsUpdateWithoutVersion(t *testing.T) {
	client, _ := captureClient(t, 200, `{"data":`+designJSON+`,"version":null}`)
	out, err := client.Designs.Update(context.Background(), "des_1", &UpdateDesignRequest{Name: String("x")})
	if err != nil {
		t.Fatalf("Designs.Update: %v", err)
	}
	if out.Version != nil || out.Design.Name != "Launch" {
		t.Fatalf("update = %+v", out)
	}
}

func TestDesignsDelete(t *testing.T) {
	client, got := captureClient(t, 200, `{"data":{"id":"des_1","deleted":true}}`)
	out, err := client.Designs.Delete(context.Background(), "des_1")
	if err != nil {
		t.Fatalf("Designs.Delete: %v", err)
	}
	if got.method != "DELETE" || got.path != "/designs/des_1" || !out.Deleted || out.ID != "des_1" {
		t.Fatalf("request = %+v out = %+v", got, out)
	}
}

func TestDesignsListVersionsIncludesDocumentOnRequest(t *testing.T) {
	client, got := captureClient(t, 200, `{"data":[{"id":"ver_1","version":2,"document":{"pages":[]},"created_at":"2026-11-20T10:00:00.000Z"}]}`)
	versions, err := client.Designs.ListVersions(context.Background(), "des_1", true)
	if err != nil {
		t.Fatalf("Designs.ListVersions: %v", err)
	}
	if got.method != "GET" || got.path != "/designs/des_1/versions" || got.query != "include=document" {
		t.Fatalf("request = %+v", got)
	}
	if len(versions) != 1 || versions[0].Version != 2 || versions[0].Document == nil {
		t.Fatalf("versions = %+v", versions)
	}

	client, got = captureClient(t, 200, `{"data":[]}`)
	if _, err := client.Designs.ListVersions(context.Background(), "des_1", false); err != nil {
		t.Fatalf("Designs.ListVersions: %v", err)
	}
	if got.query != "" {
		t.Fatalf("query = %q", got.query)
	}
}

func TestDesignsRestoreVersion(t *testing.T) {
	client, got := captureClient(t, 200, `{"data":`+designJSON+`,"version":{"id":"ver_9","version":5}}`)
	out, err := client.Designs.RestoreVersion(context.Background(), "des_1", "ver_1")
	if err != nil {
		t.Fatalf("Designs.RestoreVersion: %v", err)
	}
	if got.method != "POST" || got.path != "/designs/des_1/versions/ver_1/restore" {
		t.Fatalf("request = %+v", got)
	}
	if out.Design.ID != "des_1" || out.Version.ID != "ver_9" {
		t.Fatalf("restore = %+v", out)
	}
}

func TestDesignsDuplicate(t *testing.T) {
	client, got := captureClient(t, 201, `{"data":`+designJSON+`}`)
	if _, err := client.Designs.Duplicate(context.Background(), "des_1", "Copy"); err != nil {
		t.Fatalf("Designs.Duplicate: %v", err)
	}
	if got.method != "POST" || got.path != "/designs/des_1/duplicate" || got.body["name"] != "Copy" {
		t.Fatalf("request = %+v", got)
	}

	client, got = captureClient(t, 201, `{"data":`+designJSON+`}`)
	if _, err := client.Designs.Duplicate(context.Background(), "des_1", ""); err != nil {
		t.Fatalf("Designs.Duplicate: %v", err)
	}
	if _, ok := got.body["name"]; ok {
		t.Fatalf("name should be omitted: %+v", got.body)
	}
}

func TestDesignsExport(t *testing.T) {
	client, got := captureClient(t, 200, `{"data":{"format":"png","files":[{"media_id":"med_1","page_index":0,"mime_type":"image/png","preview_url":"/v1/media/med_1/file"}]}}`)
	out, err := client.Designs.Export(context.Background(), "des_1", &ExportDesignRequest{Format: DesignExportPNG, PageIndexes: []int{0, 2}})
	if err != nil {
		t.Fatalf("Designs.Export: %v", err)
	}
	if got.method != "POST" || got.path != "/designs/des_1/export" || got.body["format"] != "png" {
		t.Fatalf("request = %+v", got)
	}
	if pages, _ := got.body["page_indexes"].([]any); len(pages) != 2 {
		t.Fatalf("body = %+v", got.body)
	}
	if _, ok := got.body["duration_sec"]; ok {
		t.Fatalf("duration_sec should be omitted: %+v", got.body)
	}
	if out.Format != "png" || len(out.Files) != 1 || out.Files[0].MediaID != "med_1" || out.Files[0].MimeType != "image/png" {
		t.Fatalf("export = %+v", out)
	}
	if _, err := client.Designs.Export(context.Background(), "des_1", &ExportDesignRequest{}); err == nil {
		t.Fatal("expected an error without a format")
	}
}

func TestDesignsBulk(t *testing.T) {
	client, got := captureClient(t, 201, `{"data":[`+designJSON+`,`+designJSON+`]}`)
	out, err := client.Designs.Bulk(context.Background(), "des_1", &BulkCreateDesignsRequest{
		Rows:       []map[string]string{{"title": "A"}, {"title": "B"}},
		NameColumn: "title",
	})
	if err != nil {
		t.Fatalf("Designs.Bulk: %v", err)
	}
	if got.method != "POST" || got.path != "/designs/des_1/bulk" || got.body["name_column"] != "title" {
		t.Fatalf("request = %+v", got)
	}
	if rows, _ := got.body["rows"].([]any); len(rows) != 2 {
		t.Fatalf("body = %+v", got.body)
	}
	if len(out) != 2 {
		t.Fatalf("designs = %+v", out)
	}
	if _, err := client.Designs.Bulk(context.Background(), "des_1", &BulkCreateDesignsRequest{}); err == nil {
		t.Fatal("expected an error without rows")
	}
}

func TestDesignsPosts(t *testing.T) {
	client, got := captureClient(t, 200, `{"data":[{"id":"post_1","status":"published","title":null,"text":"Hi","schedule_at":null,"published_at":"2026-11-20T10:00:00.000Z","platforms":["x"],"engagement":12,"views":null}]}`)
	posts, err := client.Designs.Posts(context.Background(), "des_1")
	if err != nil {
		t.Fatalf("Designs.Posts: %v", err)
	}
	if got.method != "GET" || got.path != "/designs/des_1/posts" {
		t.Fatalf("request = %+v", got)
	}
	p := posts[0]
	if p.Status != "published" || *p.Text != "Hi" || p.Title != nil || *p.Engagement != 12 || p.Views != nil || p.PublishedAt == nil {
		t.Fatalf("post = %+v", p)
	}
}

func TestDesignsTextToSlides(t *testing.T) {
	client, got := captureClient(t, 200, `{"data":{"cover":{"title":"T","subtitle":"S"},"slides":[{"heading":"H","body":"B"}],"cta":{"heading":"Follow","body":"More"},"brand":{"colors":["#000"],"fonts":{"heading":"Geist"},"logo_media_id":null}}}`)
	out, err := client.Designs.TextToSlides(context.Background(), &TextToSlidesRequest{WorkspaceID: "ws_1", Text: "Long text", MaxSlides: 6})
	if err != nil {
		t.Fatalf("Designs.TextToSlides: %v", err)
	}
	if got.method != "POST" || got.path != "/designs/text-to-slides" || got.body["max_slides"] != float64(6) || got.body["text"] != "Long text" {
		t.Fatalf("request = %+v", got)
	}
	if _, ok := got.body["post_id"]; ok {
		t.Fatalf("post_id should be omitted: %+v", got.body)
	}
	if out.Cover.Title != "T" || len(out.Slides) != 1 || out.CTA.Heading != "Follow" || out.Brand.Fonts.Heading != "Geist" || out.Brand.LogoMediaID != nil {
		t.Fatalf("outline = %+v", out)
	}
}

func TestDesignsFromPost(t *testing.T) {
	client, got := captureClient(t, 201, `{"data":`+designJSON+`}`)
	out, err := client.Designs.FromPost(context.Background(), &DesignFromPostRequest{WorkspaceID: "ws_1", PostID: "post_1", Style: DesignFromPostQuote, Carousel: Bool(false)})
	if err != nil {
		t.Fatalf("Designs.FromPost: %v", err)
	}
	if got.method != "POST" || got.path != "/designs/from-post" {
		t.Fatalf("request = %+v", got)
	}
	if got.body["post_id"] != "post_1" || got.body["style"] != "quote" || got.body["carousel"] != false {
		t.Fatalf("body = %+v", got.body)
	}
	if out.ID != "des_1" {
		t.Fatalf("design = %+v", out)
	}
}
