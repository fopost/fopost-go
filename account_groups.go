package fopost

import (
	"context"
	"net/url"
)

// AccountGroupsService covers account groups, named sets of connected
// accounts a post can target in one go.
type AccountGroupsService struct{ client *Client }

// AccountGroup is a named set of connected accounts in one workspace.
type AccountGroup struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	AccountIDs []string `json:"account_ids"`
	CreatedAt  Time     `json:"created_at"`
	UpdatedAt  Time     `json:"updated_at"`
}

// List returns the account groups the key can reach, optionally narrowed to
// one workspace.
func (s *AccountGroupsService) List(ctx context.Context, workspaceID string) ([]AccountGroup, error) {
	q := newQuery()
	q.str("workspace_id", workspaceID)
	var out []AccountGroup
	if err := s.client.json(ctx, "GET", "/account-groups", nil, q.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Get returns one account group.
func (s *AccountGroupsService) Get(ctx context.Context, id string) (*AccountGroup, error) {
	out := &AccountGroup{}
	if err := s.client.json(ctx, "GET", "/account-groups/"+url.PathEscape(id), nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateAccountGroupRequest is the body of Create. AccountIDs is optional.
type CreateAccountGroupRequest struct {
	WorkspaceID string   `json:"workspace_id"`
	Name        string   `json:"name"`
	AccountIDs  []string `json:"account_ids,omitempty"`
}

// Create adds an account group to a workspace.
func (s *AccountGroupsService) Create(ctx context.Context, body *CreateAccountGroupRequest) (*AccountGroup, error) {
	out := &AccountGroup{}
	if err := s.client.json(ctx, "POST", "/account-groups", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateAccountGroupRequest is the body of Update.
type UpdateAccountGroupRequest struct {
	Name string `json:"name"`
}

// Update renames an account group.
func (s *AccountGroupsService) Update(ctx context.Context, id string, body *UpdateAccountGroupRequest) (*AccountGroup, error) {
	out := &AccountGroup{}
	if err := s.client.json(ctx, "PATCH", "/account-groups/"+url.PathEscape(id), body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Delete removes an account group. Its accounts stay connected.
func (s *AccountGroupsService) Delete(ctx context.Context, id string) error {
	return s.client.Do(ctx, "DELETE", "/account-groups/"+url.PathEscape(id), nil, nil, nil)
}

// SetMembers replaces an account group's members with accountIDs. An empty
// list empties the group.
func (s *AccountGroupsService) SetMembers(ctx context.Context, id string, accountIDs []string) (*AccountGroup, error) {
	if accountIDs == nil {
		accountIDs = []string{}
	}
	body := map[string]any{"account_ids": accountIDs}
	out := &AccountGroup{}
	if err := s.client.json(ctx, "PUT", "/account-groups/"+url.PathEscape(id)+"/members", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}
