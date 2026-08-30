package fopost

import (
	"context"
	"net/url"
)

// WorkspacesService covers workspaces, the tenant boundary every other
// resource is scoped to.
type WorkspacesService struct{ client *Client }

// Workspace types.
const (
	WorkspacePersonal     = "PERSONAL"
	WorkspaceTeam         = "TEAM"
	WorkspaceOrganization = "ORGANIZATION"
	WorkspaceClient       = "CLIENT"
	WorkspaceProject      = "PROJECT"
	WorkspaceDepartment   = "DEPARTMENT"
	WorkspaceEvent        = "EVENT"
	WorkspaceTemporary    = "TEMPORARY"
	WorkspaceCommunity    = "COMMUNITY"
	WorkspaceBrand        = "BRAND"
	WorkspaceAgency       = "AGENCY"
)

// WorkspaceAccountRef is a connected account as listed on a workspace.
type WorkspaceAccountRef struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspaceId"`
	Platform    string `json:"platform"`
	Username    string `json:"username"`
	Name        string `json:"name"`
	Avatar      string `json:"avatar"`
}

// Workspace is one tenant. Accounts is populated by List and Get.
type Workspace struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Slug        string                `json:"slug"`
	Type        string                `json:"type"`
	Logo        string                `json:"logo"`
	Website     string                `json:"website"`
	Timezone    string                `json:"timezone"`
	Country     string                `json:"country"`
	Description string                `json:"description"`
	Language    string                `json:"language"`
	Accounts    []WorkspaceAccountRef `json:"accounts"`
	CreatedAt   Time                  `json:"created_at"`
	UpdatedAt   Time                  `json:"updated_at"`
}

// List returns every workspace the key can reach. A key bound to a single
// workspace sees only that one.
func (s *WorkspacesService) List(ctx context.Context) ([]Workspace, error) {
	var out []Workspace
	if err := s.client.json(ctx, "GET", "/workspaces", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Get returns one workspace with its connected accounts.
func (s *WorkspacesService) Get(ctx context.Context, id string) (*Workspace, error) {
	out := &Workspace{}
	if err := s.client.json(ctx, "GET", "/workspaces/"+url.PathEscape(id), nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateWorkspaceRequest is the body of Create.
type CreateWorkspaceRequest struct {
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Type        string  `json:"type,omitempty"`
	Logo        *string `json:"logo,omitempty"`
	Website     *string `json:"website,omitempty"`
	Timezone    string  `json:"timezone,omitempty"`
	Country     *string `json:"country,omitempty"`
	Description *string `json:"description,omitempty"`
	Language    string  `json:"language,omitempty"`
}

// Create adds a workspace. Plans cap how many an account may have, so this
// answers 402 once the limit is reached.
func (s *WorkspacesService) Create(ctx context.Context, body *CreateWorkspaceRequest) (*Workspace, error) {
	out := &Workspace{}
	if err := s.client.json(ctx, "POST", "/workspaces", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateWorkspaceRequest is the body of Update. Only the fields you set are sent.
type UpdateWorkspaceRequest struct {
	Name             string  `json:"name,omitempty"`
	Slug             string  `json:"slug,omitempty"`
	Type             string  `json:"type,omitempty"`
	Logo             *string `json:"logo,omitempty"`
	Website          *string `json:"website,omitempty"`
	Timezone         string  `json:"timezone,omitempty"`
	Country          *string `json:"country,omitempty"`
	Description      *string `json:"description,omitempty"`
	Language         string  `json:"language,omitempty"`
	RequireApproval  *bool   `json:"requireApproval,omitempty"`
	AIAltTextEnabled *bool   `json:"aiAltTextEnabled,omitempty"`
	BrandColor       *string `json:"brandColor,omitempty"`
}

// Update edits a workspace.
func (s *WorkspacesService) Update(ctx context.Context, id string, body *UpdateWorkspaceRequest) (*Workspace, error) {
	out := &Workspace{}
	if err := s.client.json(ctx, "PUT", "/workspaces/"+url.PathEscape(id), body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Delete removes a workspace and everything scoped to it.
func (s *WorkspacesService) Delete(ctx context.Context, id string) error {
	return s.client.Do(ctx, "DELETE", "/workspaces/"+url.PathEscape(id), nil, nil, nil)
}

// WorkspaceAnalytics is the follower and post roll-up for one workspace.
type WorkspaceAnalytics struct {
	WorkspaceID string `json:"workspaceId"`
	Accounts    []struct {
		AccountID  string `json:"accountId"`
		Platform   string `json:"platform"`
		Username   string `json:"username"`
		Followers  *int   `json:"followers"`
		Following  *int   `json:"following"`
		TotalPosts *int   `json:"totalPosts"`
		FetchedAt  Time   `json:"fetchedAt"`
	} `json:"accounts"`
	Totals struct {
		Followers  int `json:"followers"`
		TotalPosts int `json:"totalPosts"`
	} `json:"totals"`
}

// Analytics returns a workspace's follower and post totals.
func (s *WorkspacesService) Analytics(ctx context.Context, id string) (*WorkspaceAnalytics, error) {
	out := &WorkspaceAnalytics{}
	if err := s.client.json(ctx, "GET", "/workspaces/"+url.PathEscape(id)+"/analytics", nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}
