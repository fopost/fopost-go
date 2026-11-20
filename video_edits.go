package fopost

import (
	"context"
	"fmt"
	"net/url"
)

// VideoEditsService runs edit jobs on a media library video: trim, cut,
// subtitles, clips, reframe and the rest. A job runs in the background; Get
// it until Status is succeeded or failed. Needs the posts scope.
type VideoEditsService struct{ client *Client }

// Video edit job statuses.
const (
	VideoEditQueued    = "queued"
	VideoEditRunning   = "running"
	VideoEditSucceeded = "succeeded"
	VideoEditFailed    = "failed"
)

// VideoEditOp is one operation the API offers.
type VideoEditOp struct {
	Name string `json:"name"`
	// Available is false when the deployment cannot run it.
	Available bool `json:"available"`
	UsesAI    bool `json:"uses_ai"`
}

// VideoEditOperation is one step of a job. Params depend on Op.
type VideoEditOperation struct {
	Op     string         `json:"op"`
	Params map[string]any `json:"params,omitempty"`
}

// VideoEditResultMedia is the job's output in the media library.
type VideoEditResultMedia struct {
	ID         string  `json:"id"`
	PreviewURL *string `json:"preview_url"`
	MimeType   *string `json:"mime_type"`
}

// VideoEditOutput is an extra file a job produced, such as one clip.
type VideoEditOutput struct {
	MediaID    string  `json:"media_id"`
	Label      string  `json:"label"`
	Mime       string  `json:"mime"`
	PreviewURL *string `json:"preview_url"`
}

// VideoEdit is one edit job.
type VideoEdit struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	// Progress runs from 0 to 1.
	Progress      float64               `json:"progress"`
	SourceMediaID string                `json:"source_media_id"`
	Operations    []VideoEditOperation  `json:"operations"`
	ResultMediaID *string               `json:"result_media_id"`
	ResultMedia   *VideoEditResultMedia `json:"result_media"`
	ExtraOutputs  []VideoEditOutput     `json:"extra_outputs"`
	Result        map[string]any        `json:"result"`
	Error         *string               `json:"error"`
	CreatedAt     Time                  `json:"created_at"`
	UpdatedAt     Time                  `json:"updated_at"`
	CompletedAt   *Time                 `json:"completed_at"`
}

// Ops lists the operations a job may use.
func (s *VideoEditsService) Ops(ctx context.Context) ([]VideoEditOp, error) {
	var out []VideoEditOp
	if err := s.client.json(ctx, "GET", "/video-edits/ops", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListVideoEditsParams narrows List. Limit defaults to 20.
type ListVideoEditsParams struct {
	WorkspaceID string
	MediaID     string
	Limit       int
}

// List returns recent jobs, newest first.
func (s *VideoEditsService) List(ctx context.Context, params *ListVideoEditsParams) ([]VideoEdit, error) {
	q := newQuery()
	if params != nil {
		q.str("workspace_id", params.WorkspaceID)
		q.str("media_id", params.MediaID)
		q.num("limit", params.Limit)
	}
	var out []VideoEdit
	if err := s.client.json(ctx, "GET", "/video-edits", nil, q.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateVideoEditRequest is the body of Create: 1 to 10 operations, run in order.
type CreateVideoEditRequest struct {
	WorkspaceID string               `json:"workspace_id,omitempty"`
	MediaID     string               `json:"media_id"`
	Operations  []VideoEditOperation `json:"operations"`
}

// Create queues a job. It returns at once; poll Get for the result.
func (s *VideoEditsService) Create(ctx context.Context, req *CreateVideoEditRequest) (*VideoEdit, error) {
	if req == nil || req.MediaID == "" {
		return nil, fmt.Errorf("fopost: a media id is required")
	}
	if len(req.Operations) == 0 {
		return nil, fmt.Errorf("fopost: at least one operation is required")
	}
	var out VideoEdit
	if err := s.client.json(ctx, "POST", "/video-edits", req, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get returns one job. workspaceID may be empty.
func (s *VideoEditsService) Get(ctx context.Context, id, workspaceID string) (*VideoEdit, error) {
	q := newQuery()
	q.str("workspace_id", workspaceID)
	var out VideoEdit
	if err := s.client.json(ctx, "GET", "/video-edits/"+url.PathEscape(id), nil, q.values(), &out); err != nil {
		return nil, err
	}
	return &out, nil
}
