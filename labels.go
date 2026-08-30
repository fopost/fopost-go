package fopost

import (
	"context"
	"net/url"
)

// LabelsService covers labels, the campaign tags posts are grouped by.
type LabelsService struct{ client *Client }

// LabelWorkspaceRef is the workspace a label belongs to.
type LabelWorkspaceRef struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	Type     string `json:"type"`
	Logo     string `json:"logo"`
	Timezone string `json:"timezone"`
	Language string `json:"language"`
}

// Label is one campaign tag.
type Label struct {
	ID        string             `json:"id"`
	Name      string             `json:"name"`
	Color     string             `json:"color"`
	Workspace *LabelWorkspaceRef `json:"workspace"`
	CreatedAt Time               `json:"created_at"`
	UpdatedAt Time               `json:"updated_at"`
}

// List returns the labels the key can reach, optionally narrowed to one
// workspace.
func (s *LabelsService) List(ctx context.Context, workspaceID string) ([]Label, error) {
	q := newQuery()
	q.str("workspace_id", workspaceID)
	var out []Label
	if err := s.client.json(ctx, "GET", "/labels", nil, q.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Get returns one label.
func (s *LabelsService) Get(ctx context.Context, id string) (*Label, error) {
	out := &Label{}
	if err := s.client.json(ctx, "GET", "/labels/"+url.PathEscape(id), nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateLabelRequest is the body of Create. Color is a hex value, e.g. "#2563eb".
type CreateLabelRequest struct {
	WorkspaceID string `json:"workspace_id"`
	Name        string `json:"name"`
	Color       string `json:"color"`
}

// Create adds a label to a workspace.
func (s *LabelsService) Create(ctx context.Context, body *CreateLabelRequest) (*Label, error) {
	out := &Label{}
	if err := s.client.json(ctx, "POST", "/labels", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateLabelRequest is the body of Update. Both fields are required.
type UpdateLabelRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// Update renames or recolors a label.
func (s *LabelsService) Update(ctx context.Context, id string, body *UpdateLabelRequest) (*Label, error) {
	out := &Label{}
	if err := s.client.json(ctx, "PUT", "/labels/"+url.PathEscape(id), body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Delete removes a label and unlinks it from every post carrying it.
func (s *LabelsService) Delete(ctx context.Context, id string) error {
	return s.client.Do(ctx, "DELETE", "/labels/"+url.PathEscape(id), nil, nil, nil)
}
