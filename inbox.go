package fopost

import (
	"context"
	"net/url"
	"strconv"
)

// InboxService covers the social inbox: comments, mentions and DMs read from
// connected accounts, replies sent as them, and drafted replies awaiting a
// person. Every method needs the `inbox` scope.
type InboxService struct{ client *Client }

// Inbox item types.
const (
	InboxTypeComment = "comment"
	InboxTypeMention = "mention"
	InboxTypeDM      = "dm"
)

// Inbox item states.
const (
	InboxStateUnread   = "unread"
	InboxStateRead     = "read"
	InboxStateResolved = "resolved"
	InboxStateSnoozed  = "snoozed"
)

// Inbox sort orders.
const (
	InboxSortNewest     = "newest"
	InboxSortOldest     = "oldest"
	InboxSortUnanswered = "unanswered"
)

// Inbox thread kinds for Threads.
const (
	InboxThreadsComments = "comments"
	InboxThreadsMentions = "mentions"
)

// InboxAccountRef is the connected account an inbox row belongs to.
type InboxAccountRef struct {
	ID       string `json:"id"`
	Platform string `json:"platform"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
}

// InboxAttachment is one file, link or share on an inbox item. URL and
// PreviewURL are served by the API, never a platform URL.
type InboxAttachment struct {
	// Kind is "image", "video", "audio", "file", "link" or "share".
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	Width      *int   `json:"width"`
	Height     *int   `json:"height"`
	Link       string `json:"link"`
	URL        string `json:"url"`
	PreviewURL string `json:"previewUrl"`
}

// InboxPostRef is the FoPost post an inbox row was published from.
type InboxPostRef struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// InboxPostContext is the platform post an item sits under, whoever
// published it.
type InboxPostContext struct {
	ExternalID      string        `json:"externalId"`
	IsOwn           bool          `json:"isOwn"`
	Text            string        `json:"text"`
	AuthorName      string        `json:"authorName"`
	AuthorHandle    string        `json:"authorHandle"`
	AuthorAvatarURL string        `json:"authorAvatarUrl"`
	ThumbnailURL    string        `json:"thumbnailUrl"`
	Permalink       string        `json:"permalink"`
	PublishedAt     Time          `json:"publishedAt"`
	Published       *InboxPostRef `json:"published"`
}

// InboxItem is one comment, mention or DM.
type InboxItem struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspaceId"`
	Platform    string `json:"platform"`
	// Type is one of the InboxType constants.
	Type string `json:"type"`
	// State is one of the InboxState constants.
	State string `json:"state"`
	// Direction is "inbound" or "outbound".
	Direction        string            `json:"direction"`
	ConversationID   string            `json:"conversationId"`
	AuthorName       string            `json:"authorName"`
	AuthorHandle     string            `json:"authorHandle"`
	AuthorAvatarURL  string            `json:"authorAvatarUrl"`
	Text             string            `json:"text"`
	Attachments      []InboxAttachment `json:"attachments"`
	Permalink        string            `json:"permalink"`
	PostExternalID   string            `json:"postExternalId"`
	ParentExternalID string            `json:"parentExternalId"`
	PlatformCreated  Time              `json:"platformCreatedAt"`
	SnoozedUntil     Time              `json:"snoozedUntil"`
	RepliedAt        Time              `json:"repliedAt"`
	CreatedAt        Time              `json:"createdAt"`
	// CanReply is false on platforms whose API reads but cannot answer.
	CanReply bool `json:"canReply"`
	Hidden   bool `json:"hidden"`
	Liked    bool `json:"liked"`
	Pinned   bool `json:"pinned"`
	// Reaction is our reaction on a DM, empty when there is none.
	Reaction string `json:"reaction"`
	EditedAt Time   `json:"editedAt"`
	CanHide  bool   `json:"canHide"`
	// CanDelete covers a comment someone left and our own reply.
	CanDelete bool `json:"canDelete"`
	CanLike   bool `json:"canLike"`
	// CanPin and CanEdit apply to our own comment only.
	CanPin        bool `json:"canPin"`
	CanEdit       bool `json:"canEdit"`
	CanReact      bool `json:"canReact"`
	CanSendMedia  bool `json:"canSendMedia"`
	CanQuickReply bool `json:"canQuickReply"`
	// CanPrivateReply means StartConversation can answer this comment by DM.
	CanPrivateReply bool `json:"canPrivateReply"`
	// ModerationStatus is the platform's own state for a comment: "published",
	// "held", "spam" or "rejected". Empty where the platform does not report one.
	ModerationStatus string            `json:"moderationStatus"`
	Post             *InboxPostRef     `json:"post"`
	PostContext      *InboxPostContext `json:"postContext"`
	Account          *InboxAccountRef  `json:"account"`
}

// InboxThread is one platform post with comments, or one post the account was
// mentioned in.
type InboxThread struct {
	WorkspaceID       string            `json:"workspaceId"`
	AccountID         string            `json:"accountId"`
	PostExternalID    string            `json:"postExternalId"`
	CommentCount      int               `json:"commentCount"`
	UnreadCount       int               `json:"unreadCount"`
	LastCommentAt     Time              `json:"lastCommentAt"`
	LastCommentText   string            `json:"lastCommentText"`
	LastCommentAuthor string            `json:"lastCommentAuthor"`
	Post              *InboxPostContext `json:"post"`
	Account           *InboxAccountRef  `json:"account"`
}

// InboxParticipant is the other side of a DM thread.
type InboxParticipant struct {
	Name      string `json:"name"`
	Handle    string `json:"handle"`
	AvatarURL string `json:"avatarUrl"`
}

// InboxConversation is one DM thread.
type InboxConversation struct {
	WorkspaceID         string           `json:"workspaceId"`
	AccountID           string           `json:"accountId"`
	ConversationID      string           `json:"conversationId"`
	MessageCount        int              `json:"messageCount"`
	UnreadCount         int              `json:"unreadCount"`
	LastMessageAt       Time             `json:"lastMessageAt"`
	LastMessageText     string           `json:"lastMessageText"`
	LastMessageOutbound bool             `json:"lastMessageOutbound"`
	Participant         InboxParticipant `json:"participant"`
	Account             *InboxAccountRef `json:"account"`
}

// InboxAccount is a connected account and what the inbox can read from it.
type InboxAccount struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspaceId"`
	Platform    string `json:"platform"`
	Username    string `json:"username"`
	Name        string `json:"name"`
	Avatar      string `json:"avatar"`
	// InboxSupported means comments and mentions can be read for this account.
	InboxSupported  bool   `json:"inboxSupported"`
	PendingReason   string `json:"pendingReason"`
	DMSupported     bool   `json:"dmSupported"`
	DMPendingReason string `json:"dmPendingReason"`
	// CanStartConversation means a new DM can be opened from this account by handle.
	CanStartConversation bool `json:"canStartConversation"`
	// ReconnectRequired means the grant predates a permission the inbox read
	// needs; the account is not polled until someone reconnects it.
	ReconnectRequired bool `json:"reconnectRequired"`
}

// InboxPlatform is one network and its inbox support. Comments and DMs are
// each "live", "soon" or "none".
type InboxPlatform struct {
	Platform string `json:"platform"`
	Comments string `json:"comments"`
	DMs      string `json:"dms"`
}

// InboxApprovalItem is the inbox item a drafted reply answers.
type InboxApprovalItem struct {
	ID              string `json:"id"`
	Platform        string `json:"platform"`
	Type            string `json:"type"`
	State           string `json:"state"`
	AuthorName      string `json:"authorName"`
	AuthorHandle    string `json:"authorHandle"`
	AuthorAvatarURL string `json:"authorAvatarUrl"`
	Text            string `json:"text"`
	Permalink       string `json:"permalink"`
	PlatformCreated Time   `json:"platformCreatedAt"`
}

// InboxApproval is a reply an automation or the agent drafted that a person
// still has to send. ID is what ApproveReply and RejectReply take.
type InboxApproval struct {
	ID          int    `json:"id"`
	WorkspaceID string `json:"workspaceId"`
	// Source is what drafted the reply.
	Source string `json:"source"`
	// Reply is the drafted text.
	Reply     string             `json:"reply"`
	CreatedAt Time               `json:"createdAt"`
	Item      *InboxApprovalItem `json:"item"`
}

// InboxApprovalDecision is the outcome of approving or rejecting a draft.
type InboxApprovalDecision struct {
	ID      int    `json:"id"`
	Outcome string `json:"outcome"`
}

// InboxPageMeta describes one page of an inbox list.
type InboxPageMeta struct {
	Page    int `json:"page"`
	PerPage int `json:"perPage"`
	Total   int `json:"total"`
}

// InboxItemList is one page of inbox items.
type InboxItemList struct {
	Data []InboxItem   `json:"data"`
	Meta InboxPageMeta `json:"meta"`
}

// InboxThreadList is one page of threads.
type InboxThreadList struct {
	Data []InboxThread `json:"data"`
	Meta InboxPageMeta `json:"meta"`
}

// InboxConversationList is one page of DM threads.
type InboxConversationList struct {
	Data []InboxConversation `json:"data"`
	Meta InboxPageMeta       `json:"meta"`
}

// ListInboxParams filters and paginates List. Zero fields are not sent.
type ListInboxParams struct {
	WorkspaceID string
	// Type is one of the InboxType constants.
	Type string
	// State is one of the InboxState constants.
	State     string
	Platform  string
	AccountID string
	// PostID narrows to comments under one FoPost post.
	PostID string
	// PostExternalID narrows to comments under one platform post, including
	// posts not published through FoPost.
	PostExternalID string
	// ConversationID narrows to one DM thread.
	ConversationID string
	// Direction is "inbound" or "outbound".
	Direction string
	Q         string
	// Sort is one of the InboxSort constants.
	Sort    string
	Page    int
	PerPage int
}

func (p *ListInboxParams) values() url.Values {
	q := newQuery()
	if p == nil {
		return q.values()
	}
	q.str("workspace_id", p.WorkspaceID)
	q.str("type", p.Type)
	q.str("state", p.State)
	q.str("platform", p.Platform)
	q.str("account_id", p.AccountID)
	q.str("post_id", p.PostID)
	q.str("post_external_id", p.PostExternalID)
	q.str("conversation_id", p.ConversationID)
	q.str("direction", p.Direction)
	q.str("q", p.Q)
	q.str("sort", p.Sort)
	q.num("page", p.Page)
	q.num("per_page", p.PerPage)
	return q.values()
}

// ListInboxThreadsParams filters and paginates Threads. Zero fields are not
// sent.
type ListInboxThreadsParams struct {
	WorkspaceID string
	// Kind is InboxThreadsComments (the default) for threads under the
	// workspace's own posts, or InboxThreadsMentions for posts it was tagged in.
	Kind      string
	Platform  string
	AccountID string
	// State is one of the InboxState constants.
	State string
	Q     string
	// Sort is one of the InboxSort constants.
	Sort    string
	Page    int
	PerPage int
}

func (p *ListInboxThreadsParams) values() url.Values {
	q := newQuery()
	if p == nil {
		return q.values()
	}
	q.str("workspace_id", p.WorkspaceID)
	q.str("kind", p.Kind)
	q.str("platform", p.Platform)
	q.str("account_id", p.AccountID)
	q.str("state", p.State)
	q.str("q", p.Q)
	q.str("sort", p.Sort)
	q.num("page", p.Page)
	q.num("per_page", p.PerPage)
	return q.values()
}

// ListInboxConversationsParams filters and paginates Conversations. Zero
// fields are not sent.
type ListInboxConversationsParams struct {
	WorkspaceID string
	Platform    string
	AccountID   string
	// State is one of the InboxState constants.
	State string
	Q     string
	// Sort is one of the InboxSort constants.
	Sort    string
	Page    int
	PerPage int
}

func (p *ListInboxConversationsParams) values() url.Values {
	q := newQuery()
	if p == nil {
		return q.values()
	}
	q.str("workspace_id", p.WorkspaceID)
	q.str("platform", p.Platform)
	q.str("account_id", p.AccountID)
	q.str("state", p.State)
	q.str("q", p.Q)
	q.str("sort", p.Sort)
	q.num("page", p.Page)
	q.num("per_page", p.PerPage)
	return q.values()
}

// List returns one page of comments, mentions and DMs, newest first.
func (s *InboxService) List(ctx context.Context, params *ListInboxParams) (*InboxItemList, error) {
	out := &InboxItemList{}
	if err := s.client.Do(ctx, "GET", "/inbox", nil, params.values(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// Threads returns one page of threads: one row per platform post with
// comments, or per post the account was mentioned in.
func (s *InboxService) Threads(ctx context.Context, params *ListInboxThreadsParams) (*InboxThreadList, error) {
	out := &InboxThreadList{}
	if err := s.client.Do(ctx, "GET", "/inbox/posts", nil, params.values(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// Conversations returns one page of DM threads, latest first.
func (s *InboxService) Conversations(ctx context.Context, params *ListInboxConversationsParams) (*InboxConversationList, error) {
	out := &InboxConversationList{}
	if err := s.client.Do(ctx, "GET", "/inbox/conversations", nil, params.values(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// UnreadCount returns how many inbound items are unread, optionally in one
// workspace.
func (s *InboxService) UnreadCount(ctx context.Context, workspaceID string) (int, error) {
	q := newQuery()
	q.str("workspace_id", workspaceID)
	var out struct {
		Count int `json:"count"`
	}
	if err := s.client.Do(ctx, "GET", "/inbox/unread-count", nil, q.values(), &out); err != nil {
		return 0, err
	}
	return out.Count, nil
}

// Accounts returns the connected accounts with what the inbox can read from
// each, optionally narrowed to one workspace.
func (s *InboxService) Accounts(ctx context.Context, workspaceID string) ([]InboxAccount, error) {
	q := newQuery()
	q.str("workspace_id", workspaceID)
	var out []InboxAccount
	if err := s.client.json(ctx, "GET", "/inbox/accounts", nil, q.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Platforms returns every network and its inbox support.
func (s *InboxService) Platforms(ctx context.Context) ([]InboxPlatform, error) {
	var out []InboxPlatform
	if err := s.client.json(ctx, "GET", "/inbox/platforms", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// MarkInboxThreadReadRequest is the body of MarkThreadRead. Set PostExternalID
// for a comment thread or ConversationID for a DM thread.
type MarkInboxThreadReadRequest struct {
	WorkspaceID    string `json:"workspace_id"`
	AccountID      string `json:"account_id"`
	PostExternalID string `json:"post_external_id,omitempty"`
	ConversationID string `json:"conversation_id,omitempty"`
}

// MarkThreadRead marks every unread item in a thread read and returns how
// many changed.
func (s *InboxService) MarkThreadRead(ctx context.Context, body *MarkInboxThreadReadRequest) (int, error) {
	var out struct {
		Updated int `json:"updated"`
	}
	if err := s.client.json(ctx, "POST", "/inbox/read", body, nil, &out); err != nil {
		return 0, err
	}
	return out.Updated, nil
}

// InboxDMReconnect names an account whose DM access needs reconnecting.
type InboxDMReconnect struct {
	Platform string `json:"platform"`
	Account  string `json:"account"`
}

// InboxRefreshResult is what one poll of a workspace's accounts found.
type InboxRefreshResult struct {
	AccountsPolled int `json:"accountsPolled"`
	NewItems       int `json:"newItems"`
	// RateLimited counts accounts the platform rate-limited during this poll.
	RateLimited int                `json:"rateLimited"`
	DMReconnect []InboxDMReconnect `json:"dmReconnect"`
}

// Refresh polls every inbox-capable account in the workspace now.
func (s *InboxService) Refresh(ctx context.Context, workspaceID string) (*InboxRefreshResult, error) {
	body := map[string]string{"workspace_id": workspaceID}
	out := &InboxRefreshResult{}
	if err := s.client.json(ctx, "POST", "/inbox/refresh", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateInboxItemRequest is the body of Update. SnoozedUntil is required when
// State is InboxStateSnoozed and must be in the future.
type UpdateInboxItemRequest struct {
	State        string `json:"state"`
	SnoozedUntil *Time  `json:"snoozedUntil,omitempty"`
}

// Update changes an item's state.
func (s *InboxService) Update(ctx context.Context, id string, body *UpdateInboxItemRequest) (*InboxItem, error) {
	out := &InboxItem{}
	if err := s.client.json(ctx, "PATCH", "/inbox/"+url.PathEscape(id), body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// EditComment edits a comment the account wrote, on the platform. Also needs
// the `publish` scope.
func (s *InboxService) EditComment(ctx context.Context, id, text string) (*InboxItem, error) {
	body := map[string]string{"text": text}
	out := &InboxItem{}
	if err := s.client.json(ctx, "PATCH", "/inbox/"+url.PathEscape(id), body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// InboxReplyRef is where a sent reply landed on the platform.
type InboxReplyRef struct {
	ExternalID  string `json:"externalId"`
	ExternalURL string `json:"externalUrl"`
}

// InboxReplyResult is the item after a reply, plus the reply itself.
type InboxReplyResult struct {
	Item  InboxItem     `json:"item"`
	Reply InboxReplyRef `json:"reply"`
}

// InboxReplyRequest is the body of ReplyWith. Text may be empty when
// MediaIDs is set; MediaIDs and QuickReplies also need the `publish` scope.
type InboxReplyRequest struct {
	Text string `json:"text,omitempty"`
	// MediaIDs are media library ids to attach to a DM, at most 10.
	MediaIDs []string `json:"media_ids,omitempty"`
	// QuickReplies are answer buttons under a DM, at most 13 of 20 characters each.
	QuickReplies []string `json:"quick_replies,omitempty"`
}

// Reply sends text on the platform as the connected account.
func (s *InboxService) Reply(ctx context.Context, id, text string) (*InboxReplyResult, error) {
	return s.ReplyWith(ctx, id, &InboxReplyRequest{Text: text})
}

// ReplyWith sends a reply that may carry media and quick replies.
func (s *InboxService) ReplyWith(ctx context.Context, id string, body *InboxReplyRequest) (*InboxReplyResult, error) {
	out := &InboxReplyResult{}
	if err := s.client.json(ctx, "POST", "/inbox/"+url.PathEscape(id)+"/reply", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Hide hides a comment on the platform.
func (s *InboxService) Hide(ctx context.Context, id string) (*InboxItem, error) {
	out := &InboxItem{}
	if err := s.client.json(ctx, "POST", "/inbox/"+url.PathEscape(id)+"/hide", nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Unhide shows a hidden comment again.
func (s *InboxService) Unhide(ctx context.Context, id string) (*InboxItem, error) {
	out := &InboxItem{}
	if err := s.client.json(ctx, "POST", "/inbox/"+url.PathEscape(id)+"/unhide", nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Delete deletes a comment on the platform, or our own reply (which also
// needs the `publish` scope).
func (s *InboxService) Delete(ctx context.Context, id string) error {
	return s.client.Do(ctx, "DELETE", "/inbox/"+url.PathEscape(id), nil, nil, nil)
}

func (s *InboxService) action(ctx context.Context, id, action string, body any) (*InboxItem, error) {
	out := &InboxItem{}
	if err := s.client.json(ctx, "POST", "/inbox/"+url.PathEscape(id)+"/"+action, body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Like likes an item: an upvote on Reddit, a favourite on Mastodon. Also needs
// the `publish` scope.
func (s *InboxService) Like(ctx context.Context, id string) (*InboxItem, error) {
	return s.action(ctx, id, "like", nil)
}

// Unlike removes our like. Also needs the `publish` scope.
func (s *InboxService) Unlike(ctx context.Context, id string) (*InboxItem, error) {
	return s.action(ctx, id, "unlike", nil)
}

// Pin pins our own comment. Also needs the `publish` scope.
func (s *InboxService) Pin(ctx context.Context, id string) (*InboxItem, error) {
	return s.action(ctx, id, "pin", nil)
}

// Unpin unpins our own comment. Also needs the `publish` scope.
func (s *InboxService) Unpin(ctx context.Context, id string) (*InboxItem, error) {
	return s.action(ctx, id, "unpin", nil)
}

// React reacts to a DM with an emoji; a nil reaction removes ours. Also needs
// the `publish` scope.
func (s *InboxService) React(ctx context.Context, id string, reaction *string) (*InboxItem, error) {
	body := struct {
		Reaction *string `json:"reaction"`
	}{reaction}
	return s.action(ctx, id, "react", body)
}

// StartInboxConversationRequest is the body of StartConversation. Set Handle
// and AccountID to message someone, or CommentID to answer an inbox comment
// privately.
type StartInboxConversationRequest struct {
	AccountID string `json:"account_id,omitempty"`
	Handle    string `json:"handle,omitempty"`
	CommentID string `json:"comment_id,omitempty"`
	Text      string `json:"text"`
	// MediaIDs are media library ids to attach, at most 10.
	MediaIDs []string `json:"media_ids,omitempty"`
}

// InboxStartedConversation is the DM StartConversation opened. Either field
// may be empty when the platform does not report it.
type InboxStartedConversation struct {
	ConversationID string     `json:"conversationId"`
	Item           *InboxItem `json:"item"`
}

// StartConversation opens a DM. Also needs the `publish` scope.
func (s *InboxService) StartConversation(ctx context.Context, body *StartInboxConversationRequest) (*InboxStartedConversation, error) {
	out := &InboxStartedConversation{}
	if err := s.client.json(ctx, "POST", "/inbox/conversations", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// SetTyping shows or clears the typing indicator in a DM thread and returns
// whether it is on. Also needs the `publish` scope.
func (s *InboxService) SetTyping(ctx context.Context, conversationID, accountID string, on bool) (bool, error) {
	body := map[string]any{"account_id": accountID, "on": on}
	var out struct {
		Typing bool `json:"typing"`
	}
	if err := s.client.json(ctx, "POST", "/inbox/conversations/"+url.PathEscape(conversationID)+"/typing", body, nil, &out); err != nil {
		return false, err
	}
	return out.Typing, nil
}

// ListApprovals returns drafted replies a person still has to send, optionally
// narrowed to one workspace.
func (s *InboxService) ListApprovals(ctx context.Context, workspaceID string) ([]InboxApproval, error) {
	q := newQuery()
	q.str("workspace_id", workspaceID)
	var out []InboxApproval
	if err := s.client.json(ctx, "GET", "/inbox/approvals", nil, q.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ApproveReply sends a drafted reply. A non-empty text is sent in place of the
// draft.
func (s *InboxService) ApproveReply(ctx context.Context, id int, text string) (*InboxApprovalDecision, error) {
	body := map[string]string{}
	if text != "" {
		body["text"] = text
	}
	out := &InboxApprovalDecision{}
	if err := s.client.json(ctx, "POST", "/inbox/approvals/"+strconv.Itoa(id)+"/approve", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// RejectReply discards a drafted reply.
func (s *InboxService) RejectReply(ctx context.Context, id int) (*InboxApprovalDecision, error) {
	out := &InboxApprovalDecision{}
	if err := s.client.json(ctx, "POST", "/inbox/approvals/"+strconv.Itoa(id)+"/reject", nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}
