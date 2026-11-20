package fopost

import (
	"context"
	"fmt"
	"net/url"
)

// DesignsService covers designs: stories, posts, carousels, thumbnails, cards
// and collages built in the Create editor. Needs the posts scope;
// TextToSlides also needs ai.
type DesignsService struct{ client *Client }

// Design kinds.
const (
	DesignKindStory     = "story"
	DesignKindPost      = "post"
	DesignKindCarousel  = "carousel"
	DesignKindThumbnail = "thumbnail"
	DesignKindCard      = "card"
	DesignKindCollage   = "collage"
)

// Design export formats.
const (
	DesignExportPNG = "png"
	DesignExportJPG = "jpg"
	DesignExportPDF = "pdf"
	DesignExportMP4 = "mp4"
	DesignExportGIF = "gif"
)

// DesignDocument is the editor document, passed through as the API sends it.
type DesignDocument map[string]any

// Design is one saved design. Document is empty on list rows.
type Design struct {
	ID               string         `json:"id"`
	Name             string         `json:"name"`
	Kind             string         `json:"kind"`
	Width            int            `json:"width"`
	Height           int            `json:"height"`
	Document         DesignDocument `json:"document,omitempty"`
	ThumbnailMediaID *string        `json:"thumbnail_media_id"`
	ThumbnailURL     *string        `json:"thumbnail_url"`
	CreatedAt        Time           `json:"created_at"`
	UpdatedAt        Time           `json:"updated_at"`
}

// DesignVersion is one saved version. Document is set only when asked for.
type DesignVersion struct {
	ID        string         `json:"id"`
	Version   int            `json:"version"`
	Document  DesignDocument `json:"document,omitempty"`
	CreatedAt Time           `json:"created_at"`
}

// DesignUpdate is what Update and RestoreVersion answer: the design and the
// version saved alongside it, nil when none was.
type DesignUpdate struct {
	Design  Design         `json:"data"`
	Version *DesignVersion `json:"version"`
}

// DeletedDesign confirms a delete.
type DeletedDesign struct {
	ID      string `json:"id"`
	Deleted bool   `json:"deleted"`
}

// ListDesignsParams narrows List. WorkspaceID is required.
type ListDesignsParams struct {
	WorkspaceID string
	Kind        string
	Limit       int
	Offset      int
}

// List returns the workspace's designs, newest first, without documents.
func (s *DesignsService) List(ctx context.Context, params *ListDesignsParams) ([]Design, error) {
	if params == nil || params.WorkspaceID == "" {
		return nil, fmt.Errorf("fopost: a workspace id is required")
	}
	q := newQuery()
	q.str("workspace_id", params.WorkspaceID)
	q.str("kind", params.Kind)
	q.num("limit", params.Limit)
	q.num("offset", params.Offset)
	var out []Design
	if err := s.client.json(ctx, "GET", "/designs", nil, q.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateDesignRequest is the body of Create. Document is optional; without
// one the API starts from a blank page of the kind's size.
type CreateDesignRequest struct {
	WorkspaceID  string         `json:"workspace_id"`
	Kind         string         `json:"kind"`
	Name         string         `json:"name,omitempty"`
	Document     DesignDocument `json:"document,omitempty"`
	BrandVoiceID string         `json:"brand_voice_id,omitempty"`
}

// Create saves a new design.
func (s *DesignsService) Create(ctx context.Context, req *CreateDesignRequest) (*Design, error) {
	if req == nil {
		return nil, fmt.Errorf("fopost: create design request is required")
	}
	var out Design
	if err := s.client.json(ctx, "POST", "/designs", req, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get returns one design with its document.
func (s *DesignsService) Get(ctx context.Context, id string) (*Design, error) {
	var out Design
	if err := s.client.json(ctx, "GET", designPath(id), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateDesignRequest is the body of Update. Only the set fields are sent.
type UpdateDesignRequest struct {
	Name     *string        `json:"name,omitempty"`
	Kind     *string        `json:"kind,omitempty"`
	Document DesignDocument `json:"document,omitempty"`
	// CreateVersion saves the new document as a version too.
	CreateVersion *bool `json:"create_version,omitempty"`
}

// Update edits a design.
func (s *DesignsService) Update(ctx context.Context, id string, req *UpdateDesignRequest) (*DesignUpdate, error) {
	if req == nil {
		req = &UpdateDesignRequest{}
	}
	var out DesignUpdate
	if err := s.client.Do(ctx, "PATCH", designPath(id), req, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes a design.
func (s *DesignsService) Delete(ctx context.Context, id string) (*DeletedDesign, error) {
	var out DeletedDesign
	if err := s.client.json(ctx, "DELETE", designPath(id), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListVersions returns a design's saved versions. withDocument includes each
// version's document.
func (s *DesignsService) ListVersions(ctx context.Context, id string, withDocument bool) ([]DesignVersion, error) {
	q := newQuery()
	if withDocument {
		q.str("include", "document")
	}
	var out []DesignVersion
	if err := s.client.json(ctx, "GET", designPath(id)+"/versions", nil, q.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// RestoreVersion puts a saved version back as the design's document.
func (s *DesignsService) RestoreVersion(ctx context.Context, id, versionID string) (*DesignUpdate, error) {
	var out DesignUpdate
	path := designPath(id) + "/versions/" + url.PathEscape(versionID) + "/restore"
	if err := s.client.Do(ctx, "POST", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Duplicate copies a design. An empty name lets the API pick one.
func (s *DesignsService) Duplicate(ctx context.Context, id, name string) (*Design, error) {
	body := struct {
		Name string `json:"name,omitempty"`
	}{Name: name}
	var out Design
	if err := s.client.json(ctx, "POST", designPath(id)+"/duplicate", body, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ExportDesignRequest is the body of Export.
type ExportDesignRequest struct {
	// Format is one of the DesignExport* constants.
	Format string `json:"format"`
	// PageIndexes picks pages; empty exports every page.
	PageIndexes []int `json:"page_indexes,omitempty"`
	// DurationSec is 5, 10 or 15, for mp4 and gif.
	DurationSec int `json:"duration_sec,omitempty"`
}

// ExportedDesignFile is one rendered file, stored in the media library.
type ExportedDesignFile struct {
	MediaID    string `json:"media_id"`
	PageIndex  int    `json:"page_index"`
	MimeType   string `json:"mime_type"`
	PreviewURL string `json:"preview_url"`
}

// DesignExport is what Export answers.
type DesignExport struct {
	Format string               `json:"format"`
	Files  []ExportedDesignFile `json:"files"`
}

// Export renders a design into the media library.
func (s *DesignsService) Export(ctx context.Context, id string, req *ExportDesignRequest) (*DesignExport, error) {
	if req == nil || req.Format == "" {
		return nil, fmt.Errorf("fopost: an export format is required")
	}
	var out DesignExport
	if err := s.client.json(ctx, "POST", designPath(id)+"/export", req, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// BulkCreateDesignsRequest is the body of Bulk: one design per row, each
// row's values filling the template's matching text fields.
type BulkCreateDesignsRequest struct {
	Rows []map[string]string `json:"rows"`
	// NameColumn names each design from that column.
	NameColumn string `json:"name_column,omitempty"`
}

// Bulk creates designs from a template design, one per row (up to 50).
func (s *DesignsService) Bulk(ctx context.Context, id string, req *BulkCreateDesignsRequest) ([]Design, error) {
	if req == nil || len(req.Rows) == 0 {
		return nil, fmt.Errorf("fopost: at least one row is required")
	}
	var out []Design
	if err := s.client.json(ctx, "POST", designPath(id)+"/bulk", req, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// DesignPost is a post that used a design's exported media.
type DesignPost struct {
	ID          string   `json:"id"`
	Status      string   `json:"status"`
	Title       *string  `json:"title"`
	Text        *string  `json:"text"`
	ScheduleAt  *Time    `json:"schedule_at"`
	PublishedAt *Time    `json:"published_at"`
	Platforms   []string `json:"platforms"`
	Engagement  *int     `json:"engagement"`
	Views       *int     `json:"views"`
}

// Posts lists the posts that used this design.
func (s *DesignsService) Posts(ctx context.Context, id string) ([]DesignPost, error) {
	var out []DesignPost
	if err := s.client.json(ctx, "GET", designPath(id)+"/posts", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// TextToSlidesRequest is the body of TextToSlides. Send Text or PostID.
type TextToSlidesRequest struct {
	WorkspaceID string `json:"workspace_id"`
	Text        string `json:"text,omitempty"`
	PostID      string `json:"post_id,omitempty"`
	// MaxSlides is 2 to 18.
	MaxSlides    int    `json:"max_slides,omitempty"`
	BrandVoiceID string `json:"brand_voice_id,omitempty"`
}

// SlidePoint is one heading and body.
type SlidePoint struct {
	Heading string `json:"heading"`
	Body    string `json:"body"`
}

// SlidesCover is the opening slide.
type SlidesCover struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
}

// SlidesBrand is the brand kit the slides should wear.
type SlidesBrand struct {
	Colors []string `json:"colors"`
	Fonts  struct {
		Heading string `json:"heading"`
		Body    string `json:"body"`
	} `json:"fonts"`
	LogoMediaID *string `json:"logo_media_id"`
}

// SlidesOutline is what TextToSlides answers: carousel copy, not a design.
type SlidesOutline struct {
	Cover  SlidesCover  `json:"cover"`
	Slides []SlidePoint `json:"slides"`
	CTA    SlidePoint   `json:"cta"`
	Brand  SlidesBrand  `json:"brand"`
}

// TextToSlides writes carousel copy from text or a post. Uses AI credits.
func (s *DesignsService) TextToSlides(ctx context.Context, req *TextToSlidesRequest) (*SlidesOutline, error) {
	if req == nil || req.WorkspaceID == "" {
		return nil, fmt.Errorf("fopost: a workspace id is required")
	}
	var out SlidesOutline
	if err := s.client.json(ctx, "POST", "/designs/text-to-slides", req, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Design-from-post styles.
const (
	DesignFromPostQuote = "quote"
	DesignFromPostPost  = "post"
	DesignFromPostPlain = "plain"
)

// DesignFromPostRequest is the body of FromPost. Send Text or PostID.
type DesignFromPostRequest struct {
	WorkspaceID string `json:"workspace_id"`
	Text        string `json:"text,omitempty"`
	PostID      string `json:"post_id,omitempty"`
	// Style is one of the DesignFromPost* constants; the API defaults to quote.
	Style        string `json:"style,omitempty"`
	Preset       string `json:"preset,omitempty"`
	AccountID    string `json:"account_id,omitempty"`
	Carousel     *bool  `json:"carousel,omitempty"`
	Name         string `json:"name,omitempty"`
	BrandVoiceID string `json:"brand_voice_id,omitempty"`
}

// FromPost turns a post's text into a new design.
func (s *DesignsService) FromPost(ctx context.Context, req *DesignFromPostRequest) (*Design, error) {
	if req == nil || req.WorkspaceID == "" {
		return nil, fmt.Errorf("fopost: a workspace id is required")
	}
	var out Design
	if err := s.client.json(ctx, "POST", "/designs/from-post", req, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func designPath(id string) string { return "/designs/" + url.PathEscape(id) }
