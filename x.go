package fopost

import (
	"context"
	"net/url"
)

// X engagement beyond the inbox: following from an account and reading X
// posts with its token. Every call here is an X API call metered to the
// workspace, and a grant that predates the follow and bookmark scopes answers
// a 409 with code "reconnect_required" until the account is reconnected.

// XFollowResult is the answer to FollowXUser and UnfollowXUser.
type XFollowResult struct {
	AccountID string `json:"account_id"`
	// UserID is the numeric X user id, as X reports it.
	UserID    string `json:"user_id"`
	Following bool   `json:"following"`
}

// XPost is one X post read with a connected account's token.
type XPost struct {
	ID              string  `json:"id"`
	Text            *string `json:"text"`
	Permalink       *string `json:"permalink"`
	AuthorName      *string `json:"author_name"`
	AuthorHandle    *string `json:"author_handle"`
	AuthorAvatarURL *string `json:"author_avatar_url"`
	AuthorFollowers *int    `json:"author_followers"`
	// ConversationID is the thread root.
	ConversationID *string `json:"conversation_id"`
	// ParentExternalID is the post this one replied to.
	ParentExternalID  *string `json:"parent_external_id"`
	Likes             *int    `json:"likes"`
	Replies           *int    `json:"replies"`
	Reposts           *int    `json:"reposts"`
	Quotes            *int    `json:"quotes"`
	PlatformCreatedAt *Time   `json:"platform_created_at"`
}

// XSearchParams drives SearchXPosts. AccountID and Q are required; Q is up to
// 512 characters. Zero optional fields are not sent.
type XSearchParams struct {
	AccountID string
	Q         string
	// MaxResults is 10 to 100; the API defaults to 10.
	MaxResults int
	// NextToken is the cursor from the previous page's Meta.
	NextToken string
}

func (p *XSearchParams) values() url.Values {
	q := newQuery()
	if p == nil {
		return q.values()
	}
	q.str("account_id", p.AccountID)
	q.str("q", p.Q)
	q.num("max_results", p.MaxResults)
	q.str("next_token", p.NextToken)
	return q.values()
}

// XSearchMeta carries the cursor for the next page; nil when there is none.
type XSearchMeta struct {
	NextToken *string `json:"next_token"`
}

// XPostList is one page of search results. Pass Meta.NextToken back as
// NextToken for the next page.
type XPostList struct {
	Data []XPost     `json:"data"`
	Meta XSearchMeta `json:"meta"`
}

// FollowXUser follows an X user as the account. userID is the numeric X user
// id; anything else is a 400. Needs the `inbox` and `publish` scopes. A 402
// with code "platform_balance_spent" means X's own pay-per-use balance is
// empty.
// Wraps the API's `followXUser` operation.
func (s *AccountsService) FollowXUser(ctx context.Context, id, userID string) (*XFollowResult, error) {
	return s.xFollow(ctx, id, "follow", userID)
}

// UnfollowXUser unfollows an X user as the account. Needs the `inbox` and
// `publish` scopes.
// Wraps the API's `unfollowXUser` operation.
func (s *AccountsService) UnfollowXUser(ctx context.Context, id, userID string) (*XFollowResult, error) {
	return s.xFollow(ctx, id, "unfollow", userID)
}

func (s *AccountsService) xFollow(ctx context.Context, id, action, userID string) (*XFollowResult, error) {
	out := &XFollowResult{}
	path := "/accounts/" + url.PathEscape(id) + "/x/" + action
	body := map[string]string{"user_id": userID}
	if err := s.client.json(ctx, "POST", path, body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetXPost reads one X post by its numeric id with the token of accountID.
// A 404 means X has no such post. Needs the `inbox` scope.
// Wraps the API's `getXPost` operation.
func (s *InboxService) GetXPost(ctx context.Context, postID, accountID string) (*XPost, error) {
	out := &XPost{}
	query := newQuery()
	query.str("account_id", accountID)
	path := "/x/posts/" + url.PathEscape(postID)
	if err := s.client.json(ctx, "GET", path, nil, query.values(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// SearchXPosts runs a recent search on X with the token of params.AccountID.
// Needs the `inbox` scope.
// Wraps the API's `searchXPosts` operation.
func (s *InboxService) SearchXPosts(ctx context.Context, params *XSearchParams) (*XPostList, error) {
	out := &XPostList{}
	if err := s.client.Do(ctx, "GET", "/x/search", nil, params.values(), out); err != nil {
		return nil, err
	}
	return out, nil
}
