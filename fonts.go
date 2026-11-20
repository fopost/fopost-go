package fopost

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// FontsService covers the font library and a workspace's uploaded fonts.
// Needs the posts scope.
type FontsService struct{ client *Client }

// LibraryFontFile is one file of a library font.
type LibraryFontFile struct {
	Path         string `json:"path"`
	Subset       string `json:"subset"`
	Weight       int    `json:"weight"`
	Style        string `json:"style"`
	UnicodeRange string `json:"unicode_range"`
}

// LibraryFont is a font every workspace can use.
type LibraryFont struct {
	Slug     string            `json:"slug"`
	Family   string            `json:"family"`
	Category string            `json:"category"`
	Scripts  []string          `json:"scripts"`
	Weights  []int             `json:"weights"`
	Files    []LibraryFontFile `json:"files"`
}

// UploadedFont is a font a workspace uploaded.
type UploadedFont struct {
	ID     string `json:"id"`
	Family string `json:"family"`
	// Style is "normal" or "italic".
	Style     string `json:"style"`
	Weight    int    `json:"weight"`
	Format    string `json:"format"`
	FileURL   string `json:"file_url"`
	CreatedAt Time   `json:"created_at"`
}

// FontList is what List answers.
type FontList struct {
	Library  []LibraryFont  `json:"library"`
	Uploaded []UploadedFont `json:"uploaded"`
}

// List returns the font library, plus the workspace's uploads when
// workspaceID is set.
func (s *FontsService) List(ctx context.Context, workspaceID string) (*FontList, error) {
	q := newQuery()
	q.str("workspace_id", workspaceID)
	var out FontList
	if err := s.client.json(ctx, "GET", "/fonts", nil, q.values(), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UploadFontRequest describes one font upload. Family defaults to the file
// name, Weight to 400 and Style to "normal".
type UploadFontRequest struct {
	WorkspaceID string
	File        File
	Family      string
	// Weight is 100 to 1000.
	Weight int
	Style  string
}

// Upload stores a font file (up to 5 MB) in the workspace.
func (s *FontsService) Upload(ctx context.Context, req *UploadFontRequest) (*UploadedFont, error) {
	if req == nil || req.WorkspaceID == "" {
		return nil, fmt.Errorf("fopost: a workspace id is required")
	}
	name := req.File.Name
	if name == "" {
		name = "font"
	}
	fields := map[string]string{
		"workspace_id": req.WorkspaceID,
		"family":       req.Family,
		"style":        req.Style,
	}
	if req.Weight != 0 {
		fields["weight"] = strconv.Itoa(req.Weight)
	}
	form, err := buildMultipart(fields, []fileField{{field: "file", filename: name, reader: req.File.Content}})
	if err != nil {
		return nil, err
	}
	var out UploadedFont
	err = s.client.do(ctx, &request{
		method:      "POST",
		path:        "/fonts",
		body:        form.body,
		contentType: form.contentType,
		unwrap:      true,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes an uploaded font.
func (s *FontsService) Delete(ctx context.Context, id string) error {
	return s.client.Do(ctx, "DELETE", "/fonts/"+url.PathEscape(id), nil, nil, nil)
}
