package fopost

import (
	"context"
	"net/url"
	"strconv"
)

// CommunitiesService covers the X communities an account can post into.
type CommunitiesService struct{ client *Client }

// Community is an X community linked to an account.
type Community struct {
	// ID is the local row id, used to remove the link.
	ID        int    `json:"id"`
	AccountID string `json:"accountId"`
	// CommunityID is X's own id for the community.
	CommunityID  string `json:"communityId"`
	Name         string `json:"name"`
	MemberCount  *int   `json:"memberCount"`
	Description  string `json:"description"`
	ImageURL     string `json:"imageUrl"`
	LastSyncedAt Time   `json:"lastSyncedAt"`
	CreatedAt    Time   `json:"createdAt"`
}

// CommunitySearchResult is a community as X's search returns it, before it is
// linked to the account.
type CommunitySearchResult struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MemberCount int    `json:"member_count"`
}

// List returns the communities linked to an account.
func (s *CommunitiesService) List(ctx context.Context, accountID string) ([]Community, error) {
	var out []Community
	if err := s.client.json(ctx, "GET", "/accounts/"+url.PathEscape(accountID)+"/communities", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Sync pulls the account's communities from X and stores them.
func (s *CommunitiesService) Sync(ctx context.Context, accountID string) ([]Community, error) {
	var out []Community
	if err := s.client.json(ctx, "POST", "/accounts/"+url.PathEscape(accountID)+"/communities/sync", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Search looks a community up on X without linking it.
func (s *CommunitiesService) Search(ctx context.Context, accountID, query string) ([]CommunitySearchResult, error) {
	q := newQuery()
	q.str("q", query)
	var out []CommunitySearchResult
	if err := s.client.json(ctx, "GET", "/accounts/"+url.PathEscape(accountID)+"/communities/search", nil, q.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Add links a community to the account by its X id, for the case where search
// and sync do not surface it.
func (s *CommunitiesService) Add(ctx context.Context, accountID, communityID, name string) (*Community, error) {
	body := map[string]any{"communityId": communityID}
	if name != "" {
		body["name"] = name
	}
	out := &Community{}
	if err := s.client.json(ctx, "POST", "/accounts/"+url.PathEscape(accountID)+"/communities/manual", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Remove unlinks a community. id is Community.ID, not the X community id.
func (s *CommunitiesService) Remove(ctx context.Context, accountID string, id int) error {
	path := "/accounts/" + url.PathEscape(accountID) + "/communities/" + strconv.Itoa(id)
	return s.client.Do(ctx, "DELETE", path, nil, nil, nil)
}
