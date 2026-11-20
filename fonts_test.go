package fopost

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
)

func TestFontsListSendsWorkspaceAndDecodes(t *testing.T) {
	client, got := captureClient(t, 200, `{"library":[{"slug":"geist","family":"Geist","category":"sans-serif","scripts":["latin"],"weights":[400,700],"files":[{"path":"/fonts/geist.woff2","subset":"latin","weight":400,"style":"normal","unicode_range":"U+0000-00FF"}]}],"uploaded":[{"id":"font_1","family":"Brand Sans","style":"italic","weight":700,"format":"woff2","file_url":"/v1/fonts/font_1/file","created_at":"2026-11-20T10:00:00.000Z"}]}`)
	fonts, err := client.Fonts.List(context.Background(), "ws_1")
	if err != nil {
		t.Fatalf("Fonts.List: %v", err)
	}
	if got.method != "GET" || got.path != "/fonts" || got.query != "workspace_id=ws_1" {
		t.Fatalf("request = %+v", got)
	}
	if len(fonts.Library) != 1 || fonts.Library[0].Files[0].UnicodeRange != "U+0000-00FF" {
		t.Fatalf("library = %+v", fonts.Library)
	}
	if len(fonts.Uploaded) != 1 || fonts.Uploaded[0].FileURL != "/v1/fonts/font_1/file" || fonts.Uploaded[0].Weight != 700 {
		t.Fatalf("uploaded = %+v", fonts.Uploaded)
	}
}

func TestFontsListWithoutWorkspaceSendsNoQuery(t *testing.T) {
	client, got := captureClient(t, 200, `{"library":[],"uploaded":[]}`)
	if _, err := client.Fonts.List(context.Background(), ""); err != nil {
		t.Fatalf("Fonts.List: %v", err)
	}
	if got.query != "" {
		t.Fatalf("query = %q", got.query)
	}
}

func TestFontsUploadSendsMultipartForm(t *testing.T) {
	var form map[string]string
	var filename, content string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/fonts" {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Errorf("ParseMultipartForm: %v", err)
		}
		form = map[string]string{}
		for key, values := range r.MultipartForm.Value {
			form[key] = values[0]
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			t.Errorf("FormFile: %v", err)
		} else {
			filename = header.Filename
			raw, _ := io.ReadAll(file)
			content = string(raw)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"id":"font_1","family":"Brand Sans","style":"italic","weight":700,"format":"woff2","file_url":"/v1/fonts/font_1/file","created_at":"2026-11-20T10:00:00.000Z"}`)
	})

	font, err := client.Fonts.Upload(context.Background(), &UploadFontRequest{
		WorkspaceID: "ws_1",
		File:        File{Name: "brand.woff2", Content: bytes.NewReader([]byte("wOF2"))},
		Family:      "Brand Sans",
		Weight:      700,
		Style:       "italic",
	})
	if err != nil {
		t.Fatalf("Fonts.Upload: %v", err)
	}
	if form["workspace_id"] != "ws_1" || form["family"] != "Brand Sans" || form["weight"] != "700" || form["style"] != "italic" {
		t.Fatalf("form = %+v", form)
	}
	if filename != "brand.woff2" || content != "wOF2" {
		t.Fatalf("file = %q %q", filename, content)
	}
	if font.ID != "font_1" || font.Style != "italic" {
		t.Fatalf("font = %+v", font)
	}
}

func TestFontsUploadOmitsUnsetFields(t *testing.T) {
	var form map[string][]string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseMultipartForm(1 << 20)
		form = r.MultipartForm.Value
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"id":"font_1"}`)
	})
	_, err := client.Fonts.Upload(context.Background(), &UploadFontRequest{
		WorkspaceID: "ws_1",
		File:        File{Name: "a.ttf", Content: bytes.NewReader([]byte("ttf"))},
	})
	if err != nil {
		t.Fatalf("Fonts.Upload: %v", err)
	}
	for _, key := range []string{"family", "weight", "style"} {
		if _, ok := form[key]; ok {
			t.Fatalf("%s should be omitted: %+v", key, form)
		}
	}
}

func TestFontsUploadRequiresWorkspaceAndContent(t *testing.T) {
	client, _ := captureClient(t, 201, `{}`)
	if _, err := client.Fonts.Upload(context.Background(), &UploadFontRequest{File: File{Name: "a.ttf", Content: bytes.NewReader(nil)}}); err == nil {
		t.Fatal("expected an error without a workspace id")
	}
	if _, err := client.Fonts.Upload(context.Background(), &UploadFontRequest{WorkspaceID: "ws_1"}); err == nil {
		t.Fatal("expected an error without file content")
	}
}

func TestFontsDeleteAcceptsNoContent(t *testing.T) {
	client, got := captureClient(t, http.StatusNoContent, ``)
	if err := client.Fonts.Delete(context.Background(), "font_1"); err != nil {
		t.Fatalf("Fonts.Delete: %v", err)
	}
	if got.method != "DELETE" || got.path != "/fonts/font_1" {
		t.Fatalf("request = %+v", got)
	}
}
