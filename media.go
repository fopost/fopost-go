package fopost

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
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

// PresignInput describes the file a direct upload will carry.
type PresignInput struct {
	WorkspaceID string `json:"workspaceId"`
	Filename    string `json:"filename"`
	MimeType    string `json:"mimeType"`
	// Size is the exact byte length of the upload, at most 50 MB.
	Size int64 `json:"size"`
}

// PresignedUpload is a one-time upload slot. PUT the raw bytes to UploadURL
// with exactly Headers and a Content-Length of the declared size, then call
// Complete with UploadID.
type PresignedUpload struct {
	UploadID  string            `json:"uploadId"`
	UploadURL string            `json:"uploadUrl"`
	Method    string            `json:"method"`
	Headers   map[string]string `json:"headers"`
	ExpiresAt Time              `json:"expiresAt"`
}

// Presign reserves a direct-upload slot for one file.
func (s *MediaService) Presign(ctx context.Context, input *PresignInput) (*PresignedUpload, error) {
	if input == nil {
		return nil, fmt.Errorf("fopost: presign input is required")
	}
	var out PresignedUpload
	if err := s.client.json(ctx, "POST", "/media/presign", input, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Complete turns a finished direct upload into a library asset.
func (s *MediaService) Complete(ctx context.Context, uploadID string) (*UploadedMedia, error) {
	if uploadID == "" {
		return nil, fmt.Errorf("fopost: an upload id is required")
	}
	var out UploadedMedia
	path := "/media/presign/" + url.PathEscape(uploadID) + "/complete"
	if err := s.client.json(ctx, "POST", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UploadDirect presigns, sends the bytes to the upload URL, and completes the
// upload in one call. The PUT carries no API key.
func (s *MediaService) UploadDirect(ctx context.Context, workspaceID, filename, mimeType string, data []byte) (*UploadedMedia, error) {
	presigned, err := s.Presign(ctx, &PresignInput{
		WorkspaceID: workspaceID,
		Filename:    filename,
		MimeType:    mimeType,
		Size:        int64(len(data)),
	})
	if err != nil {
		return nil, err
	}

	method := presigned.Method
	if method == "" {
		method = http.MethodPut
	}
	req, err := http.NewRequestWithContext(ctx, method, presigned.UploadURL, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("fopost: building upload request: %w", err)
	}
	req.ContentLength = int64(len(data))
	for name, value := range presigned.Headers {
		req.Header.Set(name, value)
	}
	res, err := s.client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fopost: PUT %s: %w", presigned.UploadURL, err)
	}
	body, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, errorFromResponse(res, body)
	}

	return s.Complete(ctx, presigned.UploadID)
}

// Delete removes an asset from the library.
func (s *MediaService) Delete(ctx context.Context, id string) error {
	return s.client.Do(ctx, "DELETE", "/media/"+url.PathEscape(id), nil, nil, nil)
}
