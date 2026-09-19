package fopost

import (
	"context"
	"net/url"
)

// AccountsService covers connected social accounts.
type AccountsService struct{ client *Client }

// Account is a connected social account.
type Account struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspaceId"`
	Platform    string `json:"platform"`
	Username    string `json:"username"`
	// Name is the display name override when set, else PlatformName.
	Name         string `json:"name"`
	PlatformName string `json:"platformName"`
	Avatar       string `json:"avatar"`
	IsPrimary    bool   `json:"isPrimary"`
	Active       bool   `json:"active"`
	// HealthStatus is one of the Health constants.
	HealthStatus    string `json:"healthStatus"`
	LastHealthCheck Time   `json:"lastHealthCheck"`
}

// AccountDetail adds the owning workspace to an account.
type AccountDetail struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
	Platform    string `json:"platform"`
	Username    string `json:"username"`
	// Name is the display name override when set, else PlatformName.
	Name         string `json:"name"`
	PlatformName string `json:"platform_name"`
	Avatar       string `json:"avatar"`
	Workspace    struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Slug string `json:"slug"`
		Type string `json:"type"`
	} `json:"workspace"`
	CreatedAt Time `json:"created_at"`
	UpdatedAt Time `json:"updated_at"`
}

// List returns the connected accounts the key can reach, optionally narrowed
// to one workspace.
func (s *AccountsService) List(ctx context.Context, workspaceID string) ([]Account, error) {
	q := newQuery()
	q.str("workspaceId", workspaceID)
	var out []Account
	if err := s.client.json(ctx, "GET", "/accounts", nil, q.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListAccountsParams narrows ListWithParams. Zero fields are not sent.
type ListAccountsParams struct {
	WorkspaceID string
	// GroupID keeps only the accounts in that account group.
	GroupID string
}

// ListWithParams returns the connected accounts matching params.
func (s *AccountsService) ListWithParams(ctx context.Context, params *ListAccountsParams) ([]Account, error) {
	q := newQuery()
	if params != nil {
		q.str("workspaceId", params.WorkspaceID)
		q.str("group_id", params.GroupID)
	}
	var out []Account
	if err := s.client.json(ctx, "GET", "/accounts", nil, q.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Get returns one account.
func (s *AccountsService) Get(ctx context.Context, id string) (*AccountDetail, error) {
	out := &AccountDetail{}
	if err := s.client.json(ctx, "GET", "/accounts/"+url.PathEscape(id), nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateAccountRequest connects an account from credentials you already hold.
// Platforms that use OAuth are connected in the dashboard instead.
type CreateAccountRequest struct {
	WorkspaceID string `json:"workspaceId"`
	// Platform is one of the Platform constants.
	Platform string  `json:"platform"`
	Username string  `json:"username"`
	Name     string  `json:"name"`
	Avatar   *string `json:"avatar,omitempty"`
	// Credentials are the platform's own fields, e.g. an API token.
	Credentials map[string]any `json:"credentials,omitempty"`
}

// CreatedAccount is the account Create connected.
type CreatedAccount struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
	Platform    string `json:"platform"`
	Username    string `json:"username"`
	Name        string `json:"name"`
	Avatar      string `json:"avatar"`
	CreatedAt   Time   `json:"created_at"`
	UpdatedAt   Time   `json:"updated_at"`
}

// Create connects an account with credentials.
func (s *AccountsService) Create(ctx context.Context, body *CreateAccountRequest) (*CreatedAccount, error) {
	out := &CreatedAccount{}
	if err := s.client.json(ctx, "POST", "/accounts", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Delete disconnects an account.
func (s *AccountsService) Delete(ctx context.Context, id string) error {
	return s.client.Do(ctx, "DELETE", "/accounts/"+url.PathEscape(id), nil, nil, nil)
}

// RenamedAccount is an account's names after Rename.
type RenamedAccount struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	PlatformName string `json:"platform_name"`
}

// Rename sets the name shown instead of the platform name. An empty
// displayName restores the platform name.
func (s *AccountsService) Rename(ctx context.Context, id, displayName string) (*RenamedAccount, error) {
	body := map[string]any{"display_name": nil}
	if displayName != "" {
		body["display_name"] = displayName
	}
	out := &RenamedAccount{}
	if err := s.client.json(ctx, "PATCH", "/accounts/"+url.PathEscape(id), body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// MovedAccount is the account's workspace after Move.
type MovedAccount struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
}

// Move transfers an account to another workspace the caller owns. The account
// leaves its account groups. A 409 with code "move_blocked" lists the blocking
// records in its "blocking_tables" field, readable with (*Error).Field.
func (s *AccountsService) Move(ctx context.Context, id, workspaceID string) (*MovedAccount, error) {
	body := map[string]string{"workspace_id": workspaceID}
	out := &MovedAccount{}
	if err := s.client.json(ctx, "POST", "/accounts/"+url.PathEscape(id)+"/move", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// PrimaryResult is the account's primary flag after toggling.
type PrimaryResult struct {
	ID        string `json:"id"`
	IsPrimary bool   `json:"isPrimary"`
}

// SetPrimary toggles which account leads its platform in the workspace.
func (s *AccountsService) SetPrimary(ctx context.Context, id string) (*PrimaryResult, error) {
	out := &PrimaryResult{}
	if err := s.client.json(ctx, "POST", "/accounts/"+url.PathEscape(id)+"/primary", nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ValidationResult reports whether an account's stored credentials still work.
type ValidationResult struct {
	AccountID    string `json:"accountId"`
	Platform     string `json:"platform"`
	Valid        bool   `json:"valid"`
	HealthStatus string `json:"healthStatus"`
}

// Validate checks an account's credentials against the platform.
func (s *AccountsService) Validate(ctx context.Context, id string) (*ValidationResult, error) {
	out := &ValidationResult{}
	if err := s.client.json(ctx, "POST", "/accounts/"+url.PathEscape(id)+"/validate", nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// AccountHealth is one account's connection health.
type AccountHealth struct {
	ID              string `json:"id"`
	Platform        string `json:"platform"`
	Username        string `json:"username"`
	Active          bool   `json:"active"`
	HealthStatus    string `json:"healthStatus"`
	LastHealthCheck Time   `json:"lastHealthCheck"`
}

// Health returns an account's health. Pass refresh to re-check it live rather
// than reading the last stored result.
func (s *AccountsService) Health(ctx context.Context, id string, refresh bool) (*AccountHealth, error) {
	q := newQuery()
	if refresh {
		q.str("refresh", "true")
	}
	out := &AccountHealth{}
	if err := s.client.json(ctx, "GET", "/accounts/"+url.PathEscape(id)+"/health", nil, q.values(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// HealthSummary is the health of every account the key can reach.
type HealthSummary struct {
	Accounts []struct {
		ID              string `json:"id"`
		WorkspaceID     string `json:"workspaceId"`
		Platform        string `json:"platform"`
		Username        string `json:"username"`
		HealthStatus    string `json:"healthStatus"`
		LastHealthCheck Time   `json:"lastHealthCheck"`
	} `json:"accounts"`
	Summary struct {
		Total    int `json:"total"`
		Healthy  int `json:"healthy"`
		Degraded int `json:"degraded"`
		Expired  int `json:"expired"`
		Revoked  int `json:"revoked"`
		Unknown  int `json:"unknown"`
	} `json:"summary"`
}

// HealthSummary returns the health of every account, optionally narrowed to
// one workspace.
func (s *AccountsService) HealthSummary(ctx context.Context, workspaceID string) (*HealthSummary, error) {
	q := newQuery()
	q.str("workspaceId", workspaceID)
	out := &HealthSummary{}
	if err := s.client.json(ctx, "GET", "/accounts/health", nil, q.values(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// RefreshedToken reports when the refreshed credential now expires.
type RefreshedToken struct {
	Message   string `json:"message"`
	ExpiresAt Time   `json:"expiresAt"`
}

// RefreshToken renews an account's OAuth token ahead of its expiry.
func (s *AccountsService) RefreshToken(ctx context.Context, id string) (*RefreshedToken, error) {
	out := &RefreshedToken{}
	if err := s.client.json(ctx, "POST", "/accounts/"+url.PathEscape(id)+"/refresh-token", nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// AccountAnalyticsSnapshot is one point in an account's history.
type AccountAnalyticsSnapshot struct {
	Followers    *int `json:"followers"`
	Following    *int `json:"following"`
	TotalPosts   *int `json:"totalPosts"`
	Reach        *int `json:"reach"`
	ProfileViews *int `json:"profileViews"`
	FetchedAt    Time `json:"fetchedAt"`
}

// AccountAnalyticsHistory is an account's followers over time.
type AccountAnalyticsHistory struct {
	AccountID string                     `json:"accountId"`
	Platform  string                     `json:"platform"`
	Username  string                     `json:"username"`
	History   []AccountAnalyticsSnapshot `json:"history"`
}

// Analytics returns an account's stored snapshots, newest first. A limit of 0
// leaves the API's default of 30 in place.
func (s *AccountsService) Analytics(ctx context.Context, id string, limit int) (*AccountAnalyticsHistory, error) {
	q := newQuery()
	q.num("limit", limit)
	out := &AccountAnalyticsHistory{}
	if err := s.client.json(ctx, "GET", "/accounts/"+url.PathEscape(id)+"/analytics", nil, q.values(), out); err != nil {
		return nil, err
	}
	return out, nil
}
