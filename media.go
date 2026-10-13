package fopost

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// DefaultMultipartMaxFileBytes is the per-file ceiling of a multipart request
// when Limits cannot be read. Larger files go through UploadDirect.
const DefaultMultipartMaxFileBytes int64 = 100 << 20

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

// UploadLimitTier is the next account tier and the ceiling it unlocks.
type UploadLimitTier struct {
	// UpToAccounts is nil on the open-ended top tier.
	UpToAccounts *int  `json:"upToAccounts"`
	MaxFileBytes int64 `json:"maxFileBytes"`
}

// UploadLimits is what the account may upload, as reported by Limits.
type UploadLimits struct {
	// MaxFileBytes is the largest file any route accepts.
	MaxFileBytes int64 `json:"maxFileBytes"`
	// Source is "accounts", "add_on", or "bypass".
	Source                string           `json:"source"`
	ConnectedAccounts     int              `json:"connectedAccounts"`
	LargeUploadsEnabled   bool             `json:"largeUploadsEnabled"`
	NextTier              *UploadLimitTier `json:"nextTier"`
	MultipartMaxFileBytes int64            `json:"multipartMaxFileBytes"`
	MaxFilesPerUpload     int              `json:"maxFilesPerUpload"`
	DirectUploadAvailable bool             `json:"directUploadAvailable"`
	AllowedTypes          []string         `json:"allowedTypes"`
}

// Limits reports the upload ceilings for the key's account. workspaceID may
// be empty.
func (s *MediaService) Limits(ctx context.Context, workspaceID string) (*UploadLimits, error) {
	q := newQuery()
	q.str("workspaceId", workspaceID)
	var out UploadLimits
	if err := s.client.json(ctx, "GET", "/media/limits", nil, q.values(), &out); err != nil {
		return nil, err
	}
	return &out, nil
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

// PersonConsent attests that the person in a BrandRole "person" asset agreed
// to its use.
type PersonConsent struct {
	Subject   string `json:"subject"`
	Statement string `json:"statement"`
}

// PresignInput describes the file a direct upload will carry.
type PresignInput struct {
	WorkspaceID string `json:"workspaceId"`
	Filename    string `json:"filename"`
	MimeType    string `json:"mimeType"`
	// Size is the exact byte length of the upload, up to the account's
	// MaxFileBytes (see Limits).
	Size     int64  `json:"size"`
	FolderID string `json:"folderId,omitempty"`
	// Origin is "chat" when the asset belongs to an agent conversation.
	Origin string `json:"origin,omitempty"`
	// BrandRole is "logo", "product", "person", or "background".
	BrandRole    string `json:"brandRole,omitempty"`
	BrandVoiceID string `json:"brandVoiceId,omitempty"`
	// PersonConsent is required when BrandRole is "person".
	PersonConsent *PersonConsent `json:"personConsent,omitempty"`
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
	// MaxFileBytes is the account's ceiling at the time of the presign.
	MaxFileBytes int64 `json:"maxFileBytes"`
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

// UploadDirect presigns, streams content to the upload URL, and completes the
// upload in one call. input.Size must be the exact byte length of content;
// nothing is buffered, so a multi-gigabyte file is fine. The PUT carries no
// API key.
func (s *MediaService) UploadDirect(ctx context.Context, input *PresignInput, content io.Reader) (*UploadedMedia, error) {
	if input == nil {
		return nil, fmt.Errorf("fopost: presign input is required")
	}
	if content == nil {
		return nil, fmt.Errorf("fopost: file %q has no content", input.Filename)
	}
	if input.Size <= 0 {
		return nil, fmt.Errorf("fopost: file %q needs its byte size", input.Filename)
	}
	presigned, err := s.Presign(ctx, input)
	if err != nil {
		return nil, err
	}

	method := presigned.Method
	if method == "" {
		method = http.MethodPut
	}
	req, err := http.NewRequestWithContext(ctx, method, presigned.UploadURL, content)
	if err != nil {
		return nil, fmt.Errorf("fopost: building upload request: %w", err)
	}
	req.ContentLength = input.Size
	for name, value := range presigned.Headers {
		req.Header.Set(name, value)
	}
	// The client timeout is sized for API calls, not a multi-gigabyte PUT; ctx still bounds it.
	putClient := *s.client.httpClient
	putClient.Timeout = 0
	res, err := putClient.Do(req)
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
