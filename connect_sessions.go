package fopost

import (
	"context"
	"net/url"
)

// ConnectSessionsService mints hosted connect links: a single-use page, open
// for one hour, where one of your end users connects a social account into
// your workspace with no FoPost login. Needs the `accounts` scope.
type ConnectSessionsService struct{ client *Client }

// Connect session statuses.
const (
	ConnectSessionPending = "pending"
	ConnectSessionUsed    = "used"
	ConnectSessionExpired = "expired"
)

// ConnectSessionBranding is the name and logo the hosted page shows.
type ConnectSessionBranding struct {
	Name    *string `json:"name"`
	LogoURL *string `json:"logo_url"`
}

// ConnectSession is one hosted connect link. URL is present on creation only.
type ConnectSession struct {
	ID string `json:"id"`
	// Status is one of the ConnectSession constants.
	Status         string                  `json:"status"`
	ExternalUserID string                  `json:"external_user_id"`
	Platforms      []string                `json:"platforms"`
	RedirectURL    *string                 `json:"redirect_url"`
	Branding       *ConnectSessionBranding `json:"branding"`
	// AccountID is the connected account, once the session is used.
	AccountID *string `json:"account_id"`
	URL       string  `json:"url,omitempty"`
	ExpiresAt Time    `json:"expires_at"`
	UsedAt    Time    `json:"used_at"`
	CreatedAt Time    `json:"created_at"`
}

// CreateConnectSessionBranding is the Branding of CreateConnectSessionRequest.
type CreateConnectSessionBranding struct {
	Name    string `json:"name,omitempty"`
	LogoURL string `json:"logoUrl,omitempty"`
}

// CreateConnectSessionRequest is the body of Create. WorkspaceID may be empty
// for a key bound to one workspace. RedirectURL must sit on a domain in the
// workspace redirect allowlist.
type CreateConnectSessionRequest struct {
	WorkspaceID    string `json:"workspaceId,omitempty"`
	ExternalUserID string `json:"externalUserId"`
	// Platforms are the networks the user may pick from.
	Platforms   []string                      `json:"platforms"`
	RedirectURL string                        `json:"redirectUrl,omitempty"`
	Branding    *CreateConnectSessionBranding `json:"branding,omitempty"`
}

// Create mints a link; hand the returned URL to your user.
func (s *ConnectSessionsService) Create(ctx context.Context, body *CreateConnectSessionRequest) (*ConnectSession, error) {
	out := &ConnectSession{}
	if err := s.client.json(ctx, "POST", "/connect-sessions", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Get returns a session's status and, once used, the account it connected.
func (s *ConnectSessionsService) Get(ctx context.Context, id string) (*ConnectSession, error) {
	out := &ConnectSession{}
	if err := s.client.json(ctx, "GET", "/connect-sessions/"+url.PathEscape(id), nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}
