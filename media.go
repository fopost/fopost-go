package fopost

import (
	"context"
	"fmt"
	"io"
	"net/url"
)

// MediaService covers the media library. Uploads count against the plan's
// storage allowance and are reachable with the posts scope.
type MediaService struct{ client *Client }

// MediaLibraryItem is one stored asset.
type MediaLibraryItem struct {
	ID          string `json:"id"`
	UserID      string `json:"userId"`
	WorkspaceID string `json:"workspaceId"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	// Type is "image", "video", "gif", or "document".
	Type      string `json:"type"`
	MimeType  string `json:"mimeType"`
	Size      int64  `json:"size"`
	AltText   string `json:"altText"`
	CreatedAt Time   `json:"createdAt"`
}

// UploadedMedia is an asset as it comes back from Upload, shaped to drop
// straight into a post's content block.
type UploadedMedia struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
	URL  string `json:"url"`
	Size int64  `json:"size"`
}

// AsMediaItem turns an uploaded asset into a content-block attachment.
func (m UploadedMedia) AsMediaItem() MediaItem {
	return MediaItem{Type: m.Type, Name: m.Name, URL: m.URL, Size: float64(m.Size)}
}

// List returns the workspace's media library.
func (s *MediaService) List(ctx context.Context, workspaceID string) ([]MediaLibraryItem, error) {
	if workspaceID == "" {
		return nil, fmt.Errorf("fopost: a workspace id is required")
	}
	q := newQuery()
	q.str("workspaceId", workspaceID)
	var out []MediaLibraryItem
	if err := s.client.json(ctx, "GET", "/media", nil, q.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// File is one upload: a name and the bytes behind it.
type File struct {
	Name    string
	Content io.Reader
}

// Upload stores files in the workspace's media library.
func (s *MediaService) Upload(ctx context.Context, workspaceID string, files ...File) ([]UploadedMedia, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("fopost: at least one file is required")
	}
	fields := make([]fileField, 0, len(files))
	for _, file := range files {
		name := file.Name
		if name == "" {
			name = "upload"
		}
		fields = append(fields, fileField{field: "files", filename: name, reader: file.Content})
	}
	form, err := buildMultipart(map[string]string{"workspaceId": workspaceID}, fields)
	if err != nil {
		return nil, err
	}

	var out []UploadedMedia
	err = s.client.do(ctx, &request{
		method:      "POST",
		path:        "/media/upload",
		body:        form.body,
		contentType: form.contentType,
		unwrap:      true,
	}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Delete removes an asset from the library.
func (s *MediaService) Delete(ctx context.Context, id string) error {
	return s.client.Do(ctx, "DELETE", "/media/"+url.PathEscape(id), nil, nil, nil)
}
