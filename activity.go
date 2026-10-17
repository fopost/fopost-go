package fopost

import (
	"context"
	"net/url"
)

// ActivityService covers the workspace activity log, including the security
// audit trail.
type ActivityService struct{ client *Client }

// Activity kinds. ActivityKindSecurity is the audit log: its rows are
// append-only and, unlike the rest, never expire.
const (
	ActivityKindPublish    = "publish"
	ActivityKindConnection = "connection"
	ActivityKindWebhook    = "webhook"
	ActivityKindInbox      = "inbox"
	ActivityKindAutomation = "automation"
	ActivityKindBilling    = "billing"
	ActivityKindSecurity   = "security"
)

// ActivityActor is who did it: user, api_key, agent or system. Name is empty
// for a system event.
type ActivityActor struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

// ActivityEvent is one thing that happened in a workspace.
type ActivityEvent struct {
	ID          string        `json:"id"`
	WorkspaceID string        `json:"workspace_id"`
	Kind        string        `json:"kind"`
	RefType     string        `json:"ref_type"`
	RefID       string        `json:"ref_id"`
	Summary     string        `json:"summary"`
	Actor       ActivityActor `json:"actor"`
	Time        Time          `json:"time"`
}

// ActivityPage is one page of activity. Pass NextCursor back as Cursor for the
// next one; it is empty at the end of the list.
type ActivityPage struct {
	Events     []ActivityEvent `json:"data"`
	NextCursor string          `json:"-"`
}

// ListActivityParams narrows the log. Leave WorkspaceID empty to read every
// workspace the key can reach.
type ListActivityParams struct {
	WorkspaceID string
	Kind        string
	From        *Time
	To          *Time
	Cursor      string
	Limit       int
}

// List returns activity newest first. Kind ActivityKindSecurity is the audit
// log: members joining, leaving or changing role and access, and changes to
// two-step verification, passkeys, single sign-on and signed-in devices.
func (s *ActivityService) List(ctx context.Context, params *ListActivityParams) (*ActivityPage, error) {
	q := newQuery()
	if params != nil {
		q.str("workspace_id", params.WorkspaceID)
		q.str("kind", params.Kind)
		q.time("from", params.From)
		q.time("to", params.To)
		q.str("cursor", params.Cursor)
		q.num("limit", params.Limit)
	}

	// The response carries meta beside data, so it is read whole rather than
	// unwrapped down to the array.
	var raw struct {
		Data []ActivityEvent `json:"data"`
		Meta struct {
			NextCursor string `json:"next_cursor"`
		} `json:"meta"`
	}
	req := &request{method: "GET", path: "/activity", query: url.Values(q)}
	if err := s.client.do(ctx, req, &raw); err != nil {
		return nil, err
	}
	return &ActivityPage{Events: raw.Data, NextCursor: raw.Meta.NextCursor}, nil
}
