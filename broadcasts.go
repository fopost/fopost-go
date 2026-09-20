package fopost

import (
	"context"
	"net/url"
)

// BroadcastsService covers broadcasts: one message into every conversation
// the workspace already has with a segment of its contacts.
//
// A broadcast is not a post and not a cold DM — every message lands in a
// direct-message thread the contact already started.
//
// Nothing is sent into a closed messaging window. Messenger and Instagram
// take a business-initiated message only within 24 hours of the contact's
// last one, so recipients outside it come back Skipped with SkipReason
// SkipWindowClosed and nothing is attempted, which is why the number sent is
// often lower than the audience. Telegram, Slack, Bluesky and Reddit have no
// window.
//
// Reading needs the `inbox` scope; Send and Cancel also need `publish`.
type BroadcastsService struct{ client *Client }

// SequencesService covers drip sequences: a series of messages, each a delay
// after the one before, walked per enrolled contact.
//
// The messaging window applies to every step. A step that comes due outside
// it is skipped rather than sent, and the enrollment carries on — so someone
// can complete a sequence having received only some of its messages.
type SequencesService struct{ client *Client }

// Broadcast statuses.
const (
	BroadcastDraft     = "draft"
	BroadcastScheduled = "scheduled"
	BroadcastSending   = "sending"
	BroadcastSent      = "sent"
	BroadcastCancelled = "cancelled"
)

// Recipient statuses.
const (
	RecipientPending = "pending"
	RecipientSent    = "sent"
	RecipientSkipped = "skipped"
	RecipientFailed  = "failed"
)

// Why a recipient was skipped instead of written to.
const (
	// SkipWindowClosed means the network's messaging window had shut, so
	// nothing was attempted.
	SkipWindowClosed = "window_closed"
	// SkipNoConversation means this contact never wrote to the sending account.
	SkipNoConversation = "no_conversation"
	// SkipUnsupportedPlatform means the account's network takes no messages.
	SkipUnsupportedPlatform = "unsupported_platform"
)

// Sequence statuses.
const (
	SequenceActive = "active"
	SequencePaused = "paused"
)

// Enrollment statuses.
const (
	EnrollmentActive    = "active"
	EnrollmentCompleted = "completed"
	EnrollmentStopped   = "stopped"
	EnrollmentFailed    = "failed"
)

// Operators an AudienceField clause may use.
const (
	AudienceIs       = "is"
	AudienceIsNot    = "is_not"
	AudienceContains = "contains"
	AudienceIsSet    = "is_set"
	AudienceIsNotSet = "is_not_set"
)

// AudienceField is one custom-field clause in an audience filter.
type AudienceField struct {
	Key string `json:"key"`
	// Op is one of the Audience* constants; empty means AudienceIs.
	Op    string `json:"op,omitempty"`
	Value string `json:"value,omitempty"`
}

// AudienceFilter is who a broadcast or an enrollment resolves to, expressed
// over contacts. Every clause narrows: a contact has to match all of them.
type AudienceFilter struct {
	// Platforms matches contacts with a handle on at least one of these networks.
	Platforms []string `json:"platforms,omitempty"`
	LabelIDs  []string `json:"label_ids,omitempty"`
	// Source is one of the ContactSource constants.
	Source string          `json:"source,omitempty"`
	Fields []AudienceField `json:"fields,omitempty"`
}

// BroadcastCounts is what became of a broadcast's recipients, by status.
type BroadcastCounts struct {
	Total int `json:"total"`
	Sent  int `json:"sent"`
	// Skipped is usually the messaging window doing its job.
	Skipped int `json:"skipped"`
	Failed  int `json:"failed"`
	Pending int `json:"pending"`
}

// Broadcast is one message, sent into conversations the workspace already has.
type Broadcast struct {
	ID string `json:"id"`
	// Name is internal only; it is never sent to anyone.
	Name      string         `json:"name"`
	Text      string         `json:"text"`
	AccountID string         `json:"account_id"`
	Audience  AudienceFilter `json:"audience"`
	// Status is one of the Broadcast* constants.
	Status      string          `json:"status"`
	ScheduledAt Time            `json:"scheduled_at"`
	SentAt      Time            `json:"sent_at"`
	CreatedAt   Time            `json:"created_at"`
	Counts      BroadcastCounts `json:"counts"`
	// WorkspaceID is set only on a listing that spans workspaces.
	WorkspaceID string `json:"workspace_id"`
}

// BroadcastList is one page of broadcasts.
type BroadcastList struct {
	Data       []Broadcast     `json:"data"`
	Pagination ContactPageMeta `json:"pagination"`
}

// BroadcastRecipient is one contact on one broadcast, and what became of
// their message.
type BroadcastRecipient struct {
	ContactID   string `json:"contact_id"`
	DisplayName string `json:"display_name"`
	// Status is one of the Recipient* constants.
	Status string `json:"status"`
	// SkipReason is set when Status is RecipientSkipped: one of the Skip*
	// constants. SkipWindowClosed means nothing was attempted.
	SkipReason string `json:"skip_reason"`
	SentAt     Time   `json:"sent_at"`
	Error      string `json:"error"`
}

// RecipientList is one page of a broadcast's recipients.
type RecipientList struct {
	Data       []BroadcastRecipient `json:"data"`
	Pagination ContactPageMeta      `json:"pagination"`
}

// SequenceStep is one message and how long after the previous step it goes
// out. DelayHours on the first step is measured from the enrollment.
type SequenceStep struct {
	DelayHours float64 `json:"delay_hours"`
	Text       string  `json:"text"`
	MediaID    string  `json:"media_id,omitempty"`
}

// EnrollmentCounts is where a sequence's enrollments stand, by status.
type EnrollmentCounts struct {
	Total     int `json:"total"`
	Active    int `json:"active"`
	Completed int `json:"completed"`
	Stopped   int `json:"stopped"`
	Failed    int `json:"failed"`
}

// Sequence is a series of messages, each a delay after the one before.
type Sequence struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	AccountID string         `json:"account_id"`
	Steps     []SequenceStep `json:"steps"`
	// Status is SequenceActive or SequencePaused. A paused sequence fires nothing.
	Status      string           `json:"status"`
	CreatedAt   Time             `json:"created_at"`
	Enrollments EnrollmentCounts `json:"enrollments"`
	// WorkspaceID is set only on a listing that spans workspaces.
	WorkspaceID string `json:"workspace_id"`
}

// SequenceList is one page of sequences.
type SequenceList struct {
	Data       []Sequence      `json:"data"`
	Pagination ContactPageMeta `json:"pagination"`
}

// Enrollment is one contact walking one sequence.
type Enrollment struct {
	ID          string `json:"id"`
	ContactID   string `json:"contact_id"`
	DisplayName string `json:"display_name"`
	// Step counts the steps already sent, so it is also the index of the next one.
	Step   int  `json:"step"`
	NextAt Time `json:"next_at"`
	// Status is one of the Enrollment* constants.
	Status     string `json:"status"`
	LastSentAt Time   `json:"last_sent_at"`
	// Error carries the skip reason when a step was skipped rather than sent.
	Error string `json:"error"`
}

// EnrollmentList is one page of a sequence's enrollments.
type EnrollmentList struct {
	Data       []Enrollment    `json:"data"`
	Pagination ContactPageMeta `json:"pagination"`
}

// ListBroadcastsParams filters and paginates List. Zero fields are not sent.
type ListBroadcastsParams struct {
	WorkspaceID string
	// Status is one of the Broadcast* constants.
	Status  string
	Page    int
	PerPage int
}

func (p *ListBroadcastsParams) values() url.Values {
	q := newQuery()
	if p == nil {
		return q.values()
	}
	q.str("workspace_id", p.WorkspaceID)
	q.str("status", p.Status)
	q.num("page", p.Page)
	q.num("per_page", p.PerPage)
	return q.values()
}

// CreateBroadcastRequest creates a broadcast without sending it.
type CreateBroadcastRequest struct {
	WorkspaceID string `json:"workspace_id"`
	// AccountID is the connected account the messages go out from.
	AccountID string `json:"account_id"`
	Name      string `json:"name"`
	Text      string `json:"text"`
	MediaID   string `json:"media_id,omitempty"`
	// Audience omitted means every contact in the workspace.
	Audience *AudienceFilter `json:"audience,omitempty"`
	// ScheduledAt sends it at that time instead of on demand.
	ScheduledAt *Time `json:"scheduled_at,omitempty"`
}

// UpdateBroadcastRequest is a partial update: nil fields are not sent. Only a
// draft or scheduled broadcast can be edited.
type UpdateBroadcastRequest struct {
	Name        *string         `json:"name,omitempty"`
	Text        *string         `json:"text,omitempty"`
	MediaID     *string         `json:"media_id,omitempty"`
	Audience    *AudienceFilter `json:"audience,omitempty"`
	ScheduledAt *Time           `json:"scheduled_at,omitempty"`
}

// SendBroadcastResult reports what the send started.
type SendBroadcastResult struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	// Recipients is how many contacts matched, not how many will be messaged —
	// the messaging window decides that.
	Recipients int `json:"recipients"`
}

// CancelBroadcastResult reports the broadcast's new status.
type CancelBroadcastResult struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// ListRecipientsParams filters and paginates Recipients.
type ListRecipientsParams struct {
	// Status is one of the Recipient* constants.
	Status  string
	Page    int
	PerPage int
}

func (p *ListRecipientsParams) values() url.Values {
	q := newQuery()
	if p == nil {
		return q.values()
	}
	q.str("status", p.Status)
	q.num("page", p.Page)
	q.num("per_page", p.PerPage)
	return q.values()
}

// List returns one page of broadcasts, newest first. Leave WorkspaceID empty
// to span every workspace the key can reach; each broadcast then carries one.
func (s *BroadcastsService) List(ctx context.Context, params *ListBroadcastsParams) (*BroadcastList, error) {
	out := &BroadcastList{}
	if err := s.client.Do(ctx, "GET", "/broadcasts", nil, params.values(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// Get returns one broadcast.
func (s *BroadcastsService) Get(ctx context.Context, id string) (*Broadcast, error) {
	out := &Broadcast{}
	if err := s.client.json(ctx, "GET", "/broadcasts/"+url.PathEscape(id), nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Create writes a broadcast without sending it. Give ScheduledAt to have it
// go out on its own at that time; otherwise call Send.
func (s *BroadcastsService) Create(ctx context.Context, body *CreateBroadcastRequest) (*Broadcast, error) {
	out := &Broadcast{}
	if err := s.client.json(ctx, "POST", "/broadcasts", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Update edits a draft or scheduled broadcast.
func (s *BroadcastsService) Update(ctx context.Context, id string, body *UpdateBroadcastRequest) (*Broadcast, error) {
	out := &Broadcast{}
	if err := s.client.json(ctx, "PATCH", "/broadcasts/"+url.PathEscape(id), body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Send freezes the audience into a recipient list and starts sending. Needs
// the `publish` scope as well as `inbox`.
func (s *BroadcastsService) Send(ctx context.Context, id string) (*SendBroadcastResult, error) {
	out := &SendBroadcastResult{}
	if err := s.client.json(ctx, "POST", "/broadcasts/"+url.PathEscape(id)+"/send", struct{}{}, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Cancel stops a broadcast where it stands. Anyone not yet written to stays
// unsent; messages already delivered are not recalled. Needs `publish`.
func (s *BroadcastsService) Cancel(ctx context.Context, id string) (*CancelBroadcastResult, error) {
	out := &CancelBroadcastResult{}
	if err := s.client.json(ctx, "POST", "/broadcasts/"+url.PathEscape(id)+"/cancel", struct{}{}, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Recipients returns one row per contact, with what became of their message.
// A skipped row carries SkipReason.
func (s *BroadcastsService) Recipients(ctx context.Context, id string, params *ListRecipientsParams) (*RecipientList, error) {
	out := &RecipientList{}
	path := "/broadcasts/" + url.PathEscape(id) + "/recipients"
	if err := s.client.Do(ctx, "GET", path, nil, params.values(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// Delete removes the broadcast and its recipient records. Messages already
// sent stay in the conversations they went to.
func (s *BroadcastsService) Delete(ctx context.Context, id string) error {
	return s.client.Do(ctx, "DELETE", "/broadcasts/"+url.PathEscape(id), nil, nil, nil)
}

// ListSequencesParams paginates the sequence list.
type ListSequencesParams struct {
	WorkspaceID string
	Page        int
	PerPage     int
}

func (p *ListSequencesParams) values() url.Values {
	q := newQuery()
	if p == nil {
		return q.values()
	}
	q.str("workspace_id", p.WorkspaceID)
	q.num("page", p.Page)
	q.num("per_page", p.PerPage)
	return q.values()
}

// CreateSequenceRequest creates a sequence. It enrolls nobody.
type CreateSequenceRequest struct {
	WorkspaceID string         `json:"workspace_id"`
	AccountID   string         `json:"account_id"`
	Name        string         `json:"name"`
	Steps       []SequenceStep `json:"steps"`
	// Status is SequenceActive or SequencePaused; empty means active.
	Status string `json:"status,omitempty"`
}

// UpdateSequenceRequest is a partial update: nil fields are not sent.
type UpdateSequenceRequest struct {
	Name   *string         `json:"name,omitempty"`
	Steps  *[]SequenceStep `json:"steps,omitempty"`
	Status *string         `json:"status,omitempty"`
}

// EnrollRequest names contacts outright, or the audience they are drawn from.
type EnrollRequest struct {
	ContactIDs []string        `json:"contact_ids,omitempty"`
	Audience   *AudienceFilter `json:"audience,omitempty"`
}

// EnrollResult reports how many contacts were put on the sequence.
type EnrollResult struct {
	ID       string `json:"id"`
	Enrolled int    `json:"enrolled"`
}

// UnenrollResult reports how many enrollments were stopped.
type UnenrollResult struct {
	ID      string `json:"id"`
	Stopped int    `json:"stopped"`
}

// ListEnrollmentsParams paginates the enrollment list.
type ListEnrollmentsParams struct {
	Page    int
	PerPage int
}

func (p *ListEnrollmentsParams) values() url.Values {
	q := newQuery()
	if p == nil {
		return q.values()
	}
	q.num("page", p.Page)
	q.num("per_page", p.PerPage)
	return q.values()
}

// List returns one page of sequences.
func (s *SequencesService) List(ctx context.Context, params *ListSequencesParams) (*SequenceList, error) {
	out := &SequenceList{}
	if err := s.client.Do(ctx, "GET", "/sequences", nil, params.values(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// Get returns one sequence.
func (s *SequencesService) Get(ctx context.Context, id string) (*Sequence, error) {
	out := &Sequence{}
	if err := s.client.json(ctx, "GET", "/sequences/"+url.PathEscape(id), nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Create writes a sequence. Creating one enrolls nobody.
func (s *SequencesService) Create(ctx context.Context, body *CreateSequenceRequest) (*Sequence, error) {
	out := &Sequence{}
	if err := s.client.json(ctx, "POST", "/sequences", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Update edits a sequence. Pausing it stops every enrollment from firing
// without ending any of them; resuming picks them up where they stood.
func (s *SequencesService) Update(ctx context.Context, id string, body *UpdateSequenceRequest) (*Sequence, error) {
	out := &Sequence{}
	if err := s.client.json(ctx, "PATCH", "/sequences/"+url.PathEscape(id), body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Enroll puts contacts on the sequence. Re-enrolling someone restarts their
// walk from the first step rather than running two in parallel. Needs the
// `publish` scope as well as `inbox`.
func (s *SequencesService) Enroll(ctx context.Context, id string, body *EnrollRequest) (*EnrollResult, error) {
	out := &EnrollResult{}
	if err := s.client.json(ctx, "POST", "/sequences/"+url.PathEscape(id)+"/enroll", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Unenroll takes contacts off the sequence. Nothing further fires for them.
// Needs the `publish` scope.
func (s *SequencesService) Unenroll(ctx context.Context, id string, contactIDs []string) (*UnenrollResult, error) {
	body := struct {
		ContactIDs []string `json:"contact_ids"`
	}{ContactIDs: contactIDs}
	out := &UnenrollResult{}
	if err := s.client.json(ctx, "POST", "/sequences/"+url.PathEscape(id)+"/unenroll", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Enrollments returns who is on the sequence, what step they are at, and when
// the next one is due.
func (s *SequencesService) Enrollments(ctx context.Context, id string, params *ListEnrollmentsParams) (*EnrollmentList, error) {
	out := &EnrollmentList{}
	path := "/sequences/" + url.PathEscape(id) + "/enrollments"
	if err := s.client.Do(ctx, "GET", path, nil, params.values(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// Delete removes the sequence and every enrollment on it.
func (s *SequencesService) Delete(ctx context.Context, id string) error {
	return s.client.Do(ctx, "DELETE", "/sequences/"+url.PathEscape(id), nil, nil, nil)
}
