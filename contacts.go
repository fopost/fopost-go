package fopost

import (
	"context"
	"net/url"
)

// ContactsService covers contacts: the people behind the inbox, and the
// custom fields a workspace keeps about them.
//
// A contact is one human however many handles they write from. An inbound
// inbox item files its author, a reply files whoever you answered, and both
// fold into whatever is already on file, so the same person never becomes two
// rows. Every method needs the `inbox` scope, except ConversationAnalytics,
// which needs `analytics`.
type ContactsService struct{ client *Client }

// What first created a contact row.
const (
	ContactSourceInbox  = "inbox"
	ContactSourceRadar  = "radar"
	ContactSourceImport = "import"
)

// Custom field types.
const (
	ContactFieldText    = "text"
	ContactFieldNumber  = "number"
	ContactFieldDate    = "date"
	ContactFieldSelect  = "select"
	ContactFieldBoolean = "boolean"
)

// Sort orders for ConversationAnalytics.
const (
	ConversationSortVolume  = "volume"
	ConversationSortSlowest = "slowest"
	ConversationSortRecent  = "recent"
)

// ContactChannel is one handle on one network. Handle is lower-cased with no
// leading @. ExternalID is the platform's own id for this person when the
// network gave us one, and it is what a merge prefers: a handle can be
// changed, an id cannot.
type ContactChannel struct {
	Platform   string `json:"platform"`
	Handle     string `json:"handle"`
	ExternalID string `json:"externalId,omitempty"`
}

// ContactLabel is a workspace label put on a contact.
type ContactLabel struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// Contact is one person, however many handles they write from.
type Contact struct {
	ID          string           `json:"id"`
	DisplayName string           `json:"display_name"`
	Channels    []ContactChannel `json:"channels"`
	// Source is one of the ContactSource constants.
	Source      string `json:"source"`
	Note        string `json:"note"`
	FirstSeenAt Time   `json:"first_seen_at"`
	LastSeenAt  Time   `json:"last_seen_at"`
	// Fields holds the custom field values, keyed by field key.
	Fields map[string]string `json:"fields"`
	Labels []ContactLabel    `json:"labels"`
	// WorkspaceID is set only on a listing that spans workspaces.
	WorkspaceID string `json:"workspace_id"`
}

// ContactPageMeta describes one page of a contacts list.
type ContactPageMeta struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Total   int `json:"total"`
}

// ContactList is one page of contacts.
type ContactList struct {
	Data       []Contact       `json:"data"`
	Pagination ContactPageMeta `json:"pagination"`
}

// ContactConversation is one thread a contact appears in.
type ContactConversation struct {
	// Key is how the inbox groups the thread: the DM thread id, else the post
	// the comments hang off, else the handle.
	Key             string `json:"key"`
	AccountID       string `json:"account_id"`
	AccountUsername string `json:"account_username"`
	Platform        string `json:"platform"`
	Messages        int    `json:"messages"`
	Received        int    `json:"received"`
	Sent            int    `json:"sent"`
	LastMessageAt   Time   `json:"last_message_at"`
	// LastItemID is an inbox item id, readable through the inbox endpoints.
	LastItemID string `json:"last_item_id"`
}

// ContactImportSkip is one CSV row the import could not read.
type ContactImportSkip struct {
	Row    int    `json:"row"`
	Reason string `json:"reason"`
}

// ContactImportResult is what a CSV import did.
type ContactImportResult struct {
	Created int `json:"created"`
	// Merged counts rows that folded into a contact already on file.
	Merged  int                 `json:"merged"`
	Skipped []ContactImportSkip `json:"skipped"`
	// UnknownColumns names columns that matched neither a reserved field nor a
	// custom field. They are reported, never stored.
	UnknownColumns []string `json:"unknownColumns"`
}

// ContactField is a column the workspace invented.
type ContactField struct {
	ID string `json:"id"`
	// Key is the machine name and the CSV column header. Fixed once created.
	Key  string `json:"key"`
	Name string `json:"name"`
	// Type is one of the ContactField type constants.
	Type string `json:"type"`
	// Options are the allowed values when Type is select.
	Options  []string `json:"options"`
	Position int      `json:"position"`
}

// ConversationAnalyticsRow is how one thread performed over the period.
type ConversationAnalyticsRow struct {
	Key       string `json:"key"`
	AccountID string `json:"accountId"`
	Platform  string `json:"platform"`
	Received  int    `json:"received"`
	Sent      int    `json:"sent"`
	Answered  int    `json:"answered"`
	Open      int    `json:"open"`
	// MedianResponseMinutes is nil when the thread was never answered.
	MedianResponseMinutes *float64 `json:"medianResponseMinutes"`
	FirstMessageAt        Time     `json:"firstMessageAt"`
	LastMessageAt         Time     `json:"lastMessageAt"`
}

// ConversationAnalytics is inbox analytics broken out per thread.
type ConversationAnalytics struct {
	Conversations []ConversationAnalyticsRow `json:"conversations"`
	Total         int                        `json:"total"`
	Page          int                        `json:"page"`
	PerPage       int                        `json:"perPage"`
}

// ListContactsParams filters and paginates List. Zero fields are not sent.
type ListContactsParams struct {
	WorkspaceID string
	// Search matches a display name or any of their handles.
	Search   string
	Platform string
	// Source is one of the ContactSource constants.
	Source  string
	Page    int
	PerPage int
}

func (p *ListContactsParams) values() url.Values {
	q := newQuery()
	if p == nil {
		return q.values()
	}
	q.str("workspace_id", p.WorkspaceID)
	q.str("search", p.Search)
	q.str("platform", p.Platform)
	q.str("source", p.Source)
	q.num("page", p.Page)
	q.num("per_page", p.PerPage)
	return q.values()
}

// CreateContactRequest is the body of Create.
type CreateContactRequest struct {
	WorkspaceID string            `json:"workspace_id"`
	Channels    []ContactChannel  `json:"channels"`
	DisplayName string            `json:"display_name,omitempty"`
	Note        string            `json:"note,omitempty"`
	Fields      map[string]string `json:"fields,omitempty"`
}

// UpdateContactRequest is the body of Update. Only non-nil fields are sent;
// a Fields entry set to nil clears that field.
type UpdateContactRequest struct {
	DisplayName *string            `json:"display_name,omitempty"`
	Channels    []ContactChannel   `json:"channels,omitempty"`
	Note        *string            `json:"note,omitempty"`
	Fields      map[string]*string `json:"fields,omitempty"`
}

// CreateContactFieldRequest is the body of CreateField.
type CreateContactFieldRequest struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	// Type is one of the ContactField type constants; defaults to text.
	Type string `json:"type,omitempty"`
	// Options are required when Type is select.
	Options []string `json:"options,omitempty"`
}

// UpdateContactFieldRequest is the body of UpdateField. The key and the type
// are fixed once created; the name and options are not.
type UpdateContactFieldRequest struct {
	Name     *string  `json:"name,omitempty"`
	Options  []string `json:"options,omitempty"`
	Position *int     `json:"position,omitempty"`
}

// ListConversationAnalyticsParams filters ConversationAnalytics.
type ListConversationAnalyticsParams struct {
	WorkspaceID string
	AccountID   string
	// Days is the reporting period, 1 to 365. Defaults to 7.
	Days int
	// Sort is one of the ConversationSort constants.
	Sort    string
	Page    int
	PerPage int
}

func (p *ListConversationAnalyticsParams) values() url.Values {
	q := newQuery()
	if p == nil {
		return q.values()
	}
	q.str("workspace_id", p.WorkspaceID)
	q.str("accountId", p.AccountID)
	q.num("days", p.Days)
	q.str("sort", p.Sort)
	q.num("page", p.Page)
	q.num("per_page", p.PerPage)
	return q.values()
}

// List returns one page of contacts, most recently active first. Omit
// WorkspaceID to span every workspace the key can reach; each contact then
// carries WorkspaceID.
func (s *ContactsService) List(ctx context.Context, params *ListContactsParams) (*ContactList, error) {
	out := &ContactList{}
	if err := s.client.Do(ctx, "GET", "/contacts", nil, params.values(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// Get returns one contact. A contact in a workspace the key cannot reach
// answers 404, exactly as an id that never existed does.
func (s *ContactsService) Get(ctx context.Context, id string) (*Contact, error) {
	out := &Contact{}
	if err := s.client.json(ctx, "GET", "/contacts/"+url.PathEscape(id), nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Create files a contact. It folds into the contact that already holds the
// first channel, so it cannot duplicate someone the inbox has already met.
func (s *ContactsService) Create(ctx context.Context, body *CreateContactRequest) (*Contact, error) {
	out := &Contact{}
	if err := s.client.json(ctx, "POST", "/contacts", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Update patches a contact. Only the fields set are sent.
func (s *ContactsService) Update(ctx context.Context, id string, body *UpdateContactRequest) (*Contact, error) {
	out := &Contact{}
	if err := s.client.json(ctx, "PATCH", "/contacts/"+url.PathEscape(id), body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Delete removes a contact and its field values. The messages they sent stay
// in the inbox, so a later message files them again.
func (s *ContactsService) Delete(ctx context.Context, id string) error {
	return s.client.Do(ctx, "DELETE", "/contacts/"+url.PathEscape(id), nil, nil, nil)
}

// Conversations returns the threads one contact appears in, newest first.
// Matched on their channels, so a contact merged from two handles brings both
// threads with it. A limit of 0 leaves the server default.
func (s *ContactsService) Conversations(ctx context.Context, id string, limit int) ([]ContactConversation, error) {
	q := newQuery()
	q.num("limit", limit)
	var out []ContactConversation
	if err := s.client.json(ctx, "GET", "/contacts/"+url.PathEscape(id)+"/conversations", nil, q.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Import files contacts from CSV text. `platform` and `handle` are required
// columns; any other column is read as a custom field key, and one matching
// no field comes back in UnknownColumns rather than being stored.
func (s *ContactsService) Import(ctx context.Context, workspaceID, csv string) (*ContactImportResult, error) {
	body := struct {
		WorkspaceID string `json:"workspace_id"`
		CSV         string `json:"csv"`
	}{WorkspaceID: workspaceID, CSV: csv}
	out := &ContactImportResult{}
	if err := s.client.json(ctx, "POST", "/contacts/import", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListFields returns the columns this workspace keeps about its contacts, in
// display order.
func (s *ContactsService) ListFields(ctx context.Context, workspaceID string) ([]ContactField, error) {
	q := newQuery()
	q.str("workspace_id", workspaceID)
	var out []ContactField
	if err := s.client.json(ctx, "GET", "/contacts/fields", nil, q.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateField adds a custom field. A duplicate key answers 409.
func (s *ContactsService) CreateField(ctx context.Context, workspaceID string, body *CreateContactFieldRequest) (*ContactField, error) {
	q := newQuery()
	q.str("workspace_id", workspaceID)
	out := &ContactField{}
	if err := s.client.json(ctx, "POST", "/contacts/fields", body, q.values(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateField renames a field or changes its options or position.
func (s *ContactsService) UpdateField(ctx context.Context, id string, body *UpdateContactFieldRequest) (*ContactField, error) {
	out := &ContactField{}
	if err := s.client.json(ctx, "PATCH", "/contacts/fields/"+url.PathEscape(id), body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteField removes the field and every answer to it.
func (s *ContactsService) DeleteField(ctx context.Context, id string) error {
	return s.client.Do(ctx, "DELETE", "/contacts/fields/"+url.PathEscape(id), nil, nil, nil)
}

// ConversationAnalytics returns volume and median reply time per thread.
// Counts and timings only: no message text and no author. Needs the
// `analytics` scope rather than `inbox`.
func (s *ContactsService) ConversationAnalytics(ctx context.Context, params *ListConversationAnalyticsParams) (*ConversationAnalytics, error) {
	out := &ConversationAnalytics{}
	if err := s.client.json(ctx, "GET", "/analytics/inbox/conversations", nil, params.values(), out); err != nil {
		return nil, err
	}
	return out, nil
}
