package fopost

import (
	"context"
	"net/url"
)

// LogsService reads recent API activity: every request members made against
// a workspace, kept for 30 days. Needs the `workspaces` scope.
type LogsService struct{ client *Client }

// LogEntry is one request.
type LogEntry struct {
	ID int `json:"id"`
	// WorkspaceID is nil for account-level requests.
	WorkspaceID *string `json:"workspace_id"`
	Time        Time    `json:"time"`
	// AuthMethod is "jwt" or "api_key".
	AuthMethod string `json:"auth_method"`
	Method     string `json:"method"`
	// Route is the pattern, e.g. /v1/posts/{id}; Path is what was requested.
	Route      string  `json:"route"`
	Path       string  `json:"path"`
	StatusCode int     `json:"status_code"`
	DurationMs int     `json:"duration_ms"`
	Platform   *string `json:"platform"`
	// Message is the error text of a failed request.
	Message *string `json:"message"`
}

// LogBucket is one hour of the filtered window.
type LogBucket struct {
	At     Time `json:"at"`
	Total  int  `json:"total"`
	Errors int  `json:"errors"`
}

// LogSummary totals the whole filtered window.
type LogSummary struct {
	Since   Time        `json:"since"`
	Total   int         `json:"total"`
	Errors  int         `json:"errors"`
	Buckets []LogBucket `json:"buckets"`
}

// LogPageMeta is the cursor the logs list pages with.
type LogPageMeta struct {
	// NextCursor goes back as Before for the next page; nil on the last one.
	NextCursor *string `json:"next_cursor"`
	HasMore    bool    `json:"has_more"`
}

// LogPage is one page of log entries with the window's summary.
type LogPage struct {
	Data    []LogEntry  `json:"data"`
	Meta    LogPageMeta `json:"meta"`
	Summary LogSummary  `json:"summary"`
}

// ListLogsParams filters List. Zero fields are not sent.
type ListLogsParams struct {
	// WorkspaceID narrows to one workspace; empty reads every workspace the
	// caller belongs to plus their account-level requests.
	WorkspaceID string
	// Days is the window, 1 to 30 (default 7).
	Days int
	// Status is "success" or "failed".
	Status     string
	Method     string
	StatusCode int
	Platform   string
	Route      string
	// Q searches route, path and message.
	Q string
	// Limit is 1 to 200 (default 50).
	Limit int
	// Before is the previous page's NextCursor.
	Before string
}

func (p *ListLogsParams) values() url.Values {
	q := newQuery()
	if p == nil {
		return q.values()
	}
	q.str("workspace_id", p.WorkspaceID)
	q.num("days", p.Days)
	q.str("status", p.Status)
	q.str("method", p.Method)
	q.num("status_code", p.StatusCode)
	q.str("platform", p.Platform)
	q.str("route", p.Route)
	q.str("q", p.Q)
	q.num("limit", p.Limit)
	q.str("before", p.Before)
	return q.values()
}

// List returns recent requests, newest first.
func (s *LogsService) List(ctx context.Context, params *ListLogsParams) (*LogPage, error) {
	out := &LogPage{}
	if err := s.client.Do(ctx, "GET", "/logs", nil, params.values(), out); err != nil {
		return nil, err
	}
	return out, nil
}
