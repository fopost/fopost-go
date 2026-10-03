package fopost

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
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

// PlatformMetricRow is one metric a network reports under its own name. Key is
// the platform's own name and is stable; Label is ours and may be reworded.
// Value is a number for every Kind but "series", which is an array of points.
type PlatformMetricRow struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	// Kind is one of count, duration_ms, currency_usd, ratio, series.
	Kind  string          `json:"kind"`
	Value json.RawMessage `json:"value"`
}

// Number decodes Value as a number. It reports false for a series, or for a
// metric the network answered as anything but a number.
func (r PlatformMetricRow) Number() (float64, bool) {
	var n float64
	if err := json.Unmarshal(r.Value, &n); err != nil {
		return 0, false
	}
	return n, true
}

// PlatformMetricsBlock is one side of a metric set: the account itself, or its
// newest measured post. ExternalPostID is empty on the account side.
type PlatformMetricsBlock struct {
	FetchedAt      Time                `json:"fetched_at"`
	ExternalPostID string              `json:"external_post_id"`
	Metrics        []PlatformMetricRow `json:"metrics"`
}

// AccountPlatformMetrics is what only this network reports, in its own
// vocabulary: ad-break earnings, story taps, a retention curve, the search
// terms behind a listing.
type AccountPlatformMetrics struct {
	Platform string               `json:"platform"`
	Account  PlatformMetricsBlock `json:"account"`
	Post     PlatformMetricsBlock `json:"post"`
}

// PlatformMetrics returns the account's per-network metric set, keyed by the
// platform's own metric names and read from the newest collected snapshot
// rather than fetched live. It needs the analytics scope.
//
// A network whose metric access has not been granted yet answers 503
// (platform_metrics_unavailable) rather than an empty set.
func (s *AccountsService) PlatformMetrics(ctx context.Context, id string) (*AccountPlatformMetrics, error) {
	q := newQuery()
	q.str("raw", "true")
	out := &AccountPlatformMetrics{}
	if err := s.client.json(ctx, "GET", "/accounts/"+url.PathEscape(id)+"/insights", nil, q.values(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// TelegramConnectCode is a one-time code that connects a Telegram chat.
type TelegramConnectCode struct {
	Code string `json:"code"`
	// Command is what to send in the chat: "/connect <code>".
	Command     string  `json:"command"`
	BotUsername *string `json:"bot_username"`
	DeepLink    *string `json:"deep_link"`
	GroupLink   *string `json:"group_link"`
	ExpiresAt   Time    `json:"expires_at"`
}

// CreateTelegramConnectCode mints a code valid for 15 minutes. Sending its
// Command to the bot in a chat connects that chat. workspaceID may be empty
// for a key bound to one workspace.
func (s *AccountsService) CreateTelegramConnectCode(ctx context.Context, workspaceID string) (*TelegramConnectCode, error) {
	body := map[string]string{}
	if workspaceID != "" {
		body["workspaceId"] = workspaceID
	}
	out := &TelegramConnectCode{}
	if err := s.client.json(ctx, "POST", "/accounts/telegram/connect-code", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// TelegramConnectStatus is where a connect code stands.
type TelegramConnectStatus struct {
	// Status is "pending", "connected", "failed" or "expired".
	Status string `json:"status"`
	// AccountID is set once Status is "connected".
	AccountID *string `json:"account_id"`
	// Reason is "card_required", "slot_taken" or "workspace_unavailable" once
	// Status is "failed".
	Reason *string `json:"reason"`
}

// GetTelegramConnectStatus reports whether a connect code has been used.
func (s *AccountsService) GetTelegramConnectStatus(ctx context.Context, code string) (*TelegramConnectStatus, error) {
	q := newQuery()
	q.str("code", code)
	out := &TelegramConnectStatus{}
	if err := s.client.json(ctx, "GET", "/accounts/telegram/connect-code/status", nil, q.values(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// TelegramBotCommand is one entry in the bot's command menu.
type TelegramBotCommand struct {
	// Command is 1-32 lowercase letters, digits or underscores, without the slash.
	Command     string `json:"command"`
	Description string `json:"description"`
}

// TelegramBotCommands is the command menu the bot shows in a chat.
type TelegramBotCommands struct {
	Commands []TelegramBotCommand `json:"commands"`
}

// GetTelegramBotCommands returns the command menu for a connected chat.
func (s *AccountsService) GetTelegramBotCommands(ctx context.Context, id string) (*TelegramBotCommands, error) {
	out := &TelegramBotCommands{}
	if err := s.client.json(ctx, "GET", "/accounts/"+url.PathEscape(id)+"/telegram/commands", nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// SetTelegramBotCommands replaces the command menu for a connected chat with
// 1-100 commands.
func (s *AccountsService) SetTelegramBotCommands(ctx context.Context, id string, commands []TelegramBotCommand) (*TelegramBotCommands, error) {
	body := TelegramBotCommands{Commands: commands}
	out := &TelegramBotCommands{}
	if err := s.client.json(ctx, "PUT", "/accounts/"+url.PathEscape(id)+"/telegram/commands", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteTelegramBotCommands clears the command menu for a connected chat.
func (s *AccountsService) DeleteTelegramBotCommands(ctx context.Context, id string) (*TelegramBotCommands, error) {
	out := &TelegramBotCommands{}
	if err := s.client.json(ctx, "DELETE", "/accounts/"+url.PathEscape(id)+"/telegram/commands", nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// SlackChannel is a channel a Slack account can post to.
type SlackChannel struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IsPrivate bool   `json:"is_private"`
	// IsMember reports whether the bot is in the channel.
	IsMember bool `json:"is_member"`
	// IsCurrent marks the channel the account posts to.
	IsCurrent bool `json:"is_current"`
}

// SlackMember is a person in the connected Slack workspace. Pass ID as the
// handle to start a DM.
type SlackMember struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	RealName    *string `json:"real_name"`
	DisplayName *string `json:"display_name"`
	Avatar      *string `json:"avatar"`
	IsBot       bool    `json:"is_bot"`
}

// SlackIdentity is the name and icon a Slack account posts under. Nil fields
// fall back to the app's own.
type SlackIdentity struct {
	Username  *string `json:"username"`
	IconURL   *string `json:"icon_url"`
	IconEmoji *string `json:"icon_emoji"`
}

// UpdateSlackIdentityRequest is the body of UpdateSlackIdentity. A nil field
// keeps its value and a pointer to "" clears it. Set IconURL or IconEmoji, not
// both; setting one clears the other.
type UpdateSlackIdentityRequest struct {
	// Username is 1-80 characters.
	Username *string
	// IconURL is an http(s) image URL.
	IconURL *string
	// IconEmoji is an emoji code such as ":rocket:".
	IconEmoji *string
}

// MarshalJSON omits nil fields and sends "" as null.
func (r UpdateSlackIdentityRequest) MarshalJSON() ([]byte, error) {
	body := map[string]any{}
	for key, value := range map[string]*string{"username": r.Username, "icon_url": r.IconURL, "icon_emoji": r.IconEmoji} {
		if value == nil {
			continue
		}
		if *value == "" {
			body[key] = nil
		} else {
			body[key] = *value
		}
	}
	return json.Marshal(body)
}

// ListSlackChannels returns the channels a Slack account can post to: every
// public channel, and private ones the app was invited to. A 409 with code
// "webhook_connection" means the account posts through a webhook.
func (s *AccountsService) ListSlackChannels(ctx context.Context, id string) ([]SlackChannel, error) {
	var out []SlackChannel
	if err := s.client.json(ctx, "GET", "/accounts/"+url.PathEscape(id)+"/slack/channels", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListSlackMembers returns the people in a Slack account's workspace.
func (s *AccountsService) ListSlackMembers(ctx context.Context, id string) ([]SlackMember, error) {
	var out []SlackMember
	if err := s.client.json(ctx, "GET", "/accounts/"+url.PathEscape(id)+"/slack/members", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetSlackIdentity returns the name and icon a Slack account posts under.
func (s *AccountsService) GetSlackIdentity(ctx context.Context, id string) (*SlackIdentity, error) {
	out := &SlackIdentity{}
	if err := s.client.json(ctx, "GET", "/accounts/"+url.PathEscape(id)+"/slack/identity", nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateSlackIdentity sets the name and icon a Slack account posts under.
func (s *AccountsService) UpdateSlackIdentity(ctx context.Context, id string, body *UpdateSlackIdentityRequest) (*SlackIdentity, error) {
	if body == nil {
		body = &UpdateSlackIdentityRequest{}
	}
	out := &SlackIdentity{}
	if err := s.client.json(ctx, "PATCH", "/accounts/"+url.PathEscape(id)+"/slack/identity", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ─── Meta messaging settings (Facebook Pages, Instagram) ─────────

// MetaIceBreaker is a tappable prompt shown before the first message.
type MetaIceBreaker struct {
	// Question is up to 80 characters.
	Question string `json:"question"`
	// Payload is what the webhook receives when the prompt is tapped.
	Payload string `json:"payload"`
}

// MetaIceBreakers is the set of ice breakers on one account.
type MetaIceBreakers struct {
	IceBreakers []MetaIceBreaker `json:"ice_breakers"`
}

// MetaMenuItem is a persistent-menu item: a "postback" carrying Payload, or a
// "web_url" carrying URL.
type MetaMenuItem struct {
	Type  string `json:"type"`
	Title string `json:"title"`
	// Payload is set on a postback item.
	Payload string `json:"payload,omitempty"`
	// URL is set on a web_url item and must be http(s).
	URL string `json:"url,omitempty"`
}

// MetaPersistentMenuEntry is one locale's menu; "default" is the fallback.
type MetaPersistentMenuEntry struct {
	Locale                string         `json:"locale"`
	CallToActions         []MetaMenuItem `json:"call_to_actions"`
	ComposerInputDisabled *bool          `json:"composer_input_disabled,omitempty"`
}

// MetaPersistentMenu is the menu on one account, one entry per locale.
type MetaPersistentMenu struct {
	PersistentMenu []MetaPersistentMenuEntry `json:"persistent_menu"`
}

// MetaGreetingText is one locale's greeting, up to 160 characters.
type MetaGreetingText struct {
	Locale string `json:"locale"`
	Text   string `json:"text"`
}

// MetaGreeting is the greeting on one account, one entry per locale.
type MetaGreeting struct {
	Greeting []MetaGreetingText `json:"greeting"`
}

// WebhookSubscription is what the network delivers to the FoPost webhook for
// one account. Subscribed is false when it lapsed or a field is missing.
type WebhookSubscription struct {
	Subscribed    bool     `json:"subscribed"`
	Fields        []string `json:"fields"`
	MissingFields []string `json:"missing_fields"`
}

// GetIceBreakers returns the prompts shown before the first message. A network
// without them answers 400.
func (s *AccountsService) GetIceBreakers(ctx context.Context, id string) (*MetaIceBreakers, error) {
	out := &MetaIceBreakers{}
	if err := s.client.json(ctx, "GET", "/accounts/"+url.PathEscape(id)+"/messaging/ice-breakers", nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// SetIceBreakers replaces the ice breakers, up to four.
func (s *AccountsService) SetIceBreakers(ctx context.Context, id string, iceBreakers []MetaIceBreaker) (*MetaIceBreakers, error) {
	body := MetaIceBreakers{IceBreakers: iceBreakers}
	out := &MetaIceBreakers{}
	if err := s.client.json(ctx, "PUT", "/accounts/"+url.PathEscape(id)+"/messaging/ice-breakers", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteIceBreakers clears the ice breakers.
func (s *AccountsService) DeleteIceBreakers(ctx context.Context, id string) (*MetaIceBreakers, error) {
	out := &MetaIceBreakers{}
	if err := s.client.json(ctx, "DELETE", "/accounts/"+url.PathEscape(id)+"/messaging/ice-breakers", nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetPersistentMenu returns the always-visible Messenger menu. Facebook Pages
// only; other networks answer 400.
func (s *AccountsService) GetPersistentMenu(ctx context.Context, id string) (*MetaPersistentMenu, error) {
	out := &MetaPersistentMenu{}
	if err := s.client.json(ctx, "GET", "/accounts/"+url.PathEscape(id)+"/messaging/persistent-menu", nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// SetPersistentMenu replaces the menu, one entry per locale, up to three items each.
func (s *AccountsService) SetPersistentMenu(ctx context.Context, id string, menu []MetaPersistentMenuEntry) (*MetaPersistentMenu, error) {
	body := MetaPersistentMenu{PersistentMenu: menu}
	out := &MetaPersistentMenu{}
	if err := s.client.json(ctx, "PUT", "/accounts/"+url.PathEscape(id)+"/messaging/persistent-menu", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeletePersistentMenu clears the menu.
func (s *AccountsService) DeletePersistentMenu(ctx context.Context, id string) (*MetaPersistentMenu, error) {
	out := &MetaPersistentMenu{}
	if err := s.client.json(ctx, "DELETE", "/accounts/"+url.PathEscape(id)+"/messaging/persistent-menu", nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetGreeting returns the text shown before a Messenger conversation starts.
// Facebook Pages only.
func (s *AccountsService) GetGreeting(ctx context.Context, id string) (*MetaGreeting, error) {
	out := &MetaGreeting{}
	if err := s.client.json(ctx, "GET", "/accounts/"+url.PathEscape(id)+"/messaging/greeting", nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// SetGreeting replaces the greeting, one entry per locale.
func (s *AccountsService) SetGreeting(ctx context.Context, id string, greeting []MetaGreetingText) (*MetaGreeting, error) {
	body := MetaGreeting{Greeting: greeting}
	out := &MetaGreeting{}
	if err := s.client.json(ctx, "PUT", "/accounts/"+url.PathEscape(id)+"/messaging/greeting", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteGreeting clears the greeting.
func (s *AccountsService) DeleteGreeting(ctx context.Context, id string) (*MetaGreeting, error) {
	out := &MetaGreeting{}
	if err := s.client.json(ctx, "DELETE", "/accounts/"+url.PathEscape(id)+"/messaging/greeting", nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetWebhookSubscription reports what the network is delivering to the FoPost
// webhook for this account.
func (s *AccountsService) GetWebhookSubscription(ctx context.Context, id string) (*WebhookSubscription, error) {
	out := &WebhookSubscription{}
	if err := s.client.json(ctx, "GET", "/accounts/"+url.PathEscape(id)+"/webhook-subscription", nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ResubscribeWebhook subscribes to every field this account needs, lapsed or not.
func (s *AccountsService) ResubscribeWebhook(ctx context.Context, id string) (*WebhookSubscription, error) {
	out := &WebhookSubscription{}
	if err := s.client.json(ctx, "POST", "/accounts/"+url.PathEscape(id)+"/webhook-subscription", nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ─── Discord (bot connections) ─────────────────────────────────────
//
// A Discord account connected with a webhook has no bot to act as: every route
// here answers 409 with code "webhook_connection" for one.

// DiscordChannel is a text channel the bot can post to.
type DiscordChannel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// Type is Discord's channel type: 0 text, 5 announcement, 15 forum.
	Type     int     `json:"type"`
	ParentID *string `json:"parent_id"`
	NSFW     bool    `json:"nsfw"`
	// CanPost is false when a channel permission in Discord shuts the bot out.
	CanPost bool `json:"can_post"`
	// IsCurrent marks the channel the account posts to.
	IsCurrent bool `json:"is_current"`
}

// DiscordIdentity is the nickname and avatar the bot wears in the server. Nil
// fields fall back to the application's own.
type DiscordIdentity struct {
	Username  *string `json:"username"`
	AvatarURL *string `json:"avatar_url"`
}

// UpdateDiscordIdentityRequest is the body of UpdateDiscordIdentity. A nil
// field keeps its value and a pointer to "" clears it.
type UpdateDiscordIdentityRequest struct {
	// Username is 1-32 characters.
	Username *string
	// AvatarURL is an http(s) image URL.
	AvatarURL *string
}

// MarshalJSON omits nil fields and sends "" as null.
func (r UpdateDiscordIdentityRequest) MarshalJSON() ([]byte, error) {
	body := map[string]any{}
	for key, value := range map[string]*string{"username": r.Username, "avatar_url": r.AvatarURL} {
		if value == nil {
			continue
		}
		if *value == "" {
			body[key] = nil
		} else {
			body[key] = *value
		}
	}
	return json.Marshal(body)
}

// DiscordMessage is a message in the connected channel.
type DiscordMessage struct {
	ID         string `json:"id"`
	ChannelID  string `json:"channel_id"`
	Content    string `json:"content"`
	AuthorID   string `json:"author_id"`
	AuthorName string `json:"author_name"`
	Pinned     bool   `json:"pinned"`
	CreatedAt  string `json:"created_at"`
}

// DiscordMessageRef points at a message the bot put somewhere.
type DiscordMessageRef struct {
	ID        string `json:"id"`
	ChannelID string `json:"channel_id"`
}

// DiscordThread is the thread CreateDiscordThread started.
type DiscordThread struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	ParentID *string `json:"parent_id"`
}

// DiscordScheduledEvent is an event on the server's calendar. ChannelID names a
// voice or stage channel; otherwise Location says where it happens.
type DiscordScheduledEvent struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	ChannelID   *string `json:"channel_id"`
	Location    *string `json:"location"`
	StartTime   string  `json:"start_time"`
	EndTime     *string `json:"end_time"`
	// Status is "scheduled", "active", "completed" or "canceled".
	Status    string `json:"status"`
	UserCount *int   `json:"user_count"`
}

// DiscordEventRequest is the body of CreateDiscordEvent and UpdateDiscordEvent.
// Give ChannelID, or Location with an EndTime. On an update, an empty field is
// left as it is.
type DiscordEventRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	// StartTime and EndTime are RFC 3339.
	StartTime string `json:"start_time,omitempty"`
	EndTime   string `json:"end_time,omitempty"`
	ChannelID string `json:"channel_id,omitempty"`
	Location  string `json:"location,omitempty"`
	// Status is only meaningful on an update.
	Status string `json:"status,omitempty"`
}

// DiscordMember is a person in the connected server. Pass ID as the member id
// to send a DM or assign a role.
type DiscordMember struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	DisplayName *string  `json:"display_name"`
	Nick        *string  `json:"nick"`
	Avatar      *string  `json:"avatar"`
	IsBot       bool     `json:"is_bot"`
	Roles       []string `json:"roles"`
	JoinedAt    *string  `json:"joined_at"`
}

// ListDiscordMembersOptions filters the roster. Query searches by username or
// nickname prefix.
type ListDiscordMembersOptions struct {
	Query string
	Limit int
}

// DiscordRole is a role in the connected server.
type DiscordRole struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// Color is an RGB integer; 0 is the default colour.
	Color int  `json:"color"`
	Hoist bool `json:"hoist"`
	// Managed roles belong to an integration and cannot be edited.
	Mentionable bool `json:"mentionable"`
	Managed     bool `json:"managed"`
	Position    int  `json:"position"`
	// Permissions is Discord's bitfield as a decimal string.
	Permissions string `json:"permissions"`
}

// DiscordRoleRequest is the body of CreateDiscordRole and UpdateDiscordRole.
type DiscordRoleRequest struct {
	Name        string `json:"name,omitempty"`
	Color       *int   `json:"color,omitempty"`
	Hoist       *bool  `json:"hoist,omitempty"`
	Mentionable *bool  `json:"mentionable,omitempty"`
	Permissions string `json:"permissions,omitempty"`
}

func (s *AccountsService) discordPath(id, suffix string) string {
	return "/accounts/" + url.PathEscape(id) + "/discord" + suffix
}

// ListDiscordChannels returns the text channels the bot can post to.
func (s *AccountsService) ListDiscordChannels(ctx context.Context, id string) ([]DiscordChannel, error) {
	var out []DiscordChannel
	if err := s.client.json(ctx, "GET", s.discordPath(id, "/channels"), nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SwitchDiscordChannel moves the account to another channel in the same server.
func (s *AccountsService) SwitchDiscordChannel(ctx context.Context, id, channelID string) (*DiscordChannel, error) {
	out := &DiscordChannel{}
	body := map[string]string{"channel_id": channelID}
	if err := s.client.json(ctx, "PATCH", s.discordPath(id, "/channels/current"), body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetDiscordIdentity returns the nickname and avatar the bot wears.
func (s *AccountsService) GetDiscordIdentity(ctx context.Context, id string) (*DiscordIdentity, error) {
	out := &DiscordIdentity{}
	if err := s.client.json(ctx, "GET", s.discordPath(id, "/identity"), nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateDiscordIdentity sets the nickname and avatar the bot wears.
func (s *AccountsService) UpdateDiscordIdentity(ctx context.Context, id string, body *UpdateDiscordIdentityRequest) (*DiscordIdentity, error) {
	if body == nil {
		body = &UpdateDiscordIdentityRequest{}
	}
	out := &DiscordIdentity{}
	if err := s.client.json(ctx, "PATCH", s.discordPath(id, "/identity"), body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListDiscordPins returns the pinned messages in the account's channel.
func (s *AccountsService) ListDiscordPins(ctx context.Context, id string) ([]DiscordMessage, error) {
	var out []DiscordMessage
	if err := s.client.json(ctx, "GET", s.discordPath(id, "/messages/pinned"), nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteDiscordMessage removes a message from the account's channel.
func (s *AccountsService) DeleteDiscordMessage(ctx context.Context, id, messageID string) error {
	path := s.discordPath(id, "/messages/"+url.PathEscape(messageID))
	return s.client.json(ctx, "DELETE", path, nil, nil, nil)
}

// PinDiscordMessage pins a message in the account's channel.
func (s *AccountsService) PinDiscordMessage(ctx context.Context, id, messageID string) error {
	path := s.discordPath(id, "/messages/"+url.PathEscape(messageID)+"/pin")
	return s.client.json(ctx, "POST", path, nil, nil, nil)
}

// UnpinDiscordMessage unpins a message in the account's channel.
func (s *AccountsService) UnpinDiscordMessage(ctx context.Context, id, messageID string) error {
	path := s.discordPath(id, "/messages/"+url.PathEscape(messageID)+"/pin")
	return s.client.json(ctx, "DELETE", path, nil, nil, nil)
}

// CrosspostDiscordMessage publishes an announcement-channel message to every
// server following the channel.
func (s *AccountsService) CrosspostDiscordMessage(ctx context.Context, id, messageID string) (*DiscordMessageRef, error) {
	out := &DiscordMessageRef{}
	path := s.discordPath(id, "/messages/"+url.PathEscape(messageID)+"/crosspost")
	if err := s.client.json(ctx, "POST", path, nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateDiscordThread starts a thread on a message. autoArchiveDuration is 60,
// 1440, 4320 or 10080 minutes, or 0 for the server's default.
func (s *AccountsService) CreateDiscordThread(ctx context.Context, id, messageID, name string, autoArchiveDuration int) (*DiscordThread, error) {
	body := map[string]any{"name": name}
	if autoArchiveDuration > 0 {
		body["auto_archive_duration"] = autoArchiveDuration
	}
	out := &DiscordThread{}
	path := s.discordPath(id, "/messages/"+url.PathEscape(messageID)+"/thread")
	if err := s.client.json(ctx, "POST", path, body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// SendDiscordDM sends one message to a member of the server.
func (s *AccountsService) SendDiscordDM(ctx context.Context, id, memberID, content string) (*DiscordMessageRef, error) {
	out := &DiscordMessageRef{}
	body := map[string]string{"member_id": memberID, "content": content}
	if err := s.client.json(ctx, "POST", s.discordPath(id, "/dm"), body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListDiscordEvents returns the server's scheduled events.
func (s *AccountsService) ListDiscordEvents(ctx context.Context, id string) ([]DiscordScheduledEvent, error) {
	var out []DiscordScheduledEvent
	if err := s.client.json(ctx, "GET", s.discordPath(id, "/events"), nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetDiscordEvent returns one scheduled event.
func (s *AccountsService) GetDiscordEvent(ctx context.Context, id, eventID string) (*DiscordScheduledEvent, error) {
	out := &DiscordScheduledEvent{}
	path := s.discordPath(id, "/events/"+url.PathEscape(eventID))
	if err := s.client.json(ctx, "GET", path, nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateDiscordEvent adds an event to the server's calendar.
func (s *AccountsService) CreateDiscordEvent(ctx context.Context, id string, body *DiscordEventRequest) (*DiscordScheduledEvent, error) {
	if body == nil {
		body = &DiscordEventRequest{}
	}
	out := &DiscordScheduledEvent{}
	if err := s.client.json(ctx, "POST", s.discordPath(id, "/events"), body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateDiscordEvent changes a scheduled event.
func (s *AccountsService) UpdateDiscordEvent(ctx context.Context, id, eventID string, body *DiscordEventRequest) (*DiscordScheduledEvent, error) {
	if body == nil {
		body = &DiscordEventRequest{}
	}
	out := &DiscordScheduledEvent{}
	path := s.discordPath(id, "/events/"+url.PathEscape(eventID))
	if err := s.client.json(ctx, "PATCH", path, body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteDiscordEvent removes a scheduled event.
func (s *AccountsService) DeleteDiscordEvent(ctx context.Context, id, eventID string) error {
	path := s.discordPath(id, "/events/"+url.PathEscape(eventID))
	return s.client.json(ctx, "DELETE", path, nil, nil, nil)
}

// ListDiscordMembers returns the server's roster, or the members matching a
// search when opts.Query is set.
func (s *AccountsService) ListDiscordMembers(ctx context.Context, id string, opts *ListDiscordMembersOptions) ([]DiscordMember, error) {
	query := url.Values{}
	if opts != nil {
		if opts.Query != "" {
			query.Set("q", opts.Query)
		}
		if opts.Limit > 0 {
			query.Set("limit", strconv.Itoa(opts.Limit))
		}
	}
	var out []DiscordMember
	if err := s.client.json(ctx, "GET", s.discordPath(id, "/members"), nil, query, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetDiscordMember returns one member of the server.
func (s *AccountsService) GetDiscordMember(ctx context.Context, id, memberID string) (*DiscordMember, error) {
	out := &DiscordMember{}
	path := s.discordPath(id, "/members/"+url.PathEscape(memberID))
	if err := s.client.json(ctx, "GET", path, nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListDiscordRoles returns the server's roles, highest first.
func (s *AccountsService) ListDiscordRoles(ctx context.Context, id string) ([]DiscordRole, error) {
	var out []DiscordRole
	if err := s.client.json(ctx, "GET", s.discordPath(id, "/roles"), nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateDiscordRole adds a role to the server.
func (s *AccountsService) CreateDiscordRole(ctx context.Context, id string, body *DiscordRoleRequest) (*DiscordRole, error) {
	if body == nil {
		body = &DiscordRoleRequest{}
	}
	out := &DiscordRole{}
	if err := s.client.json(ctx, "POST", s.discordPath(id, "/roles"), body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateDiscordRole changes a role on the server.
func (s *AccountsService) UpdateDiscordRole(ctx context.Context, id, roleID string, body *DiscordRoleRequest) (*DiscordRole, error) {
	if body == nil {
		body = &DiscordRoleRequest{}
	}
	out := &DiscordRole{}
	path := s.discordPath(id, "/roles/"+url.PathEscape(roleID))
	if err := s.client.json(ctx, "PATCH", path, body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteDiscordRole removes a role from the server.
func (s *AccountsService) DeleteDiscordRole(ctx context.Context, id, roleID string) error {
	path := s.discordPath(id, "/roles/"+url.PathEscape(roleID))
	return s.client.json(ctx, "DELETE", path, nil, nil, nil)
}

// AddDiscordMemberRole gives a member a role.
func (s *AccountsService) AddDiscordMemberRole(ctx context.Context, id, roleID, memberID string) error {
	path := s.discordPath(id, "/roles/"+url.PathEscape(roleID)+"/members/"+url.PathEscape(memberID))
	return s.client.json(ctx, "PUT", path, nil, nil, nil)
}

// RemoveDiscordMemberRole takes a role from a member.
func (s *AccountsService) RemoveDiscordMemberRole(ctx context.Context, id, roleID, memberID string) error {
	path := s.discordPath(id, "/roles/"+url.PathEscape(roleID)+"/members/"+url.PathEscape(memberID))
	return s.client.json(ctx, "DELETE", path, nil, nil, nil)
}
