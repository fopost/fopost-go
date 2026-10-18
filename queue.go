package fopost

import (
	"context"
	"net/url"
)

// QueueService covers the posting queue: weekly time slots a post lands in
// when it is created or updated with Schedule set to ScheduleQueue.
type QueueService struct{ client *Client }

// ScheduleQueue is the Schedule value that drops a post into the next free
// queue slot instead of a fixed ScheduleAt.
const ScheduleQueue = "queue"

// QueueSlot is one weekly posting time. A nil AccountID applies to every
// account in the workspace.
type QueueSlot struct {
	ID          string  `json:"id"`
	WorkspaceID string  `json:"workspace_id"`
	AccountID   *string `json:"account_id"`
	// Weekday is 0 (Sunday) through 6 (Saturday).
	Weekday int `json:"weekday"`
	// Time is HH:MM, 24-hour, in Timezone.
	Time string `json:"time"`
	// Timezone is an IANA zone name.
	Timezone  string `json:"timezone"`
	CreatedAt Time   `json:"created_at"`
}

// QueueMove is a queued post that changed time because a slot changed.
type QueueMove struct {
	// ID is the post id.
	ID   string `json:"id"`
	From Time   `json:"from"`
	To   Time   `json:"to"`
}

// UpdatedQueueSlot is the slot after Update, with the posts the change moved.
type UpdatedQueueSlot struct {
	QueueSlot
	Moved []QueueMove `json:"moved"`
}

// DeletedQueueSlot is the outcome of Delete, with the posts the removal moved.
type DeletedQueueSlot struct {
	Deleted bool        `json:"deleted"`
	Moved   []QueueMove `json:"moved"`
}

// ListSlots returns the workspace's slots.
func (s *QueueService) ListSlots(ctx context.Context, workspaceID string) ([]QueueSlot, error) {
	q := newQuery()
	q.str("workspace_id", workspaceID)
	var out []QueueSlot
	if err := s.client.json(ctx, "GET", "/queue/slots", nil, q.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateQueueSlotRequest is the body of CreateSlot. A nil AccountID makes the
// slot apply to every account.
type CreateQueueSlotRequest struct {
	WorkspaceID string  `json:"workspace_id"`
	AccountID   *string `json:"account_id,omitempty"`
	Weekday     int     `json:"weekday"`
	Time        string  `json:"time"`
	Timezone    string  `json:"timezone"`
}

// CreateSlot adds a weekly slot.
func (s *QueueService) CreateSlot(ctx context.Context, body *CreateQueueSlotRequest) (*QueueSlot, error) {
	out := &QueueSlot{}
	if err := s.client.json(ctx, "POST", "/queue/slots", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateQueueSlotRequest is the body of UpdateSlot. Only the fields you set
// are sent; Weekday is a pointer because Sunday is 0.
type UpdateQueueSlotRequest struct {
	AccountID *string `json:"account_id,omitempty"`
	Weekday   *int    `json:"weekday,omitempty"`
	Time      string  `json:"time,omitempty"`
	Timezone  string  `json:"timezone,omitempty"`
}

// UpdateSlot changes a slot; queued posts on it move to the new time.
func (s *QueueService) UpdateSlot(ctx context.Context, id string, body *UpdateQueueSlotRequest) (*UpdatedQueueSlot, error) {
	out := &UpdatedQueueSlot{}
	if err := s.client.json(ctx, "PUT", "/queue/slots/"+url.PathEscape(id), body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteSlot removes a slot; queued posts on it move to the next free one.
func (s *QueueService) DeleteSlot(ctx context.Context, id string) (*DeletedQueueSlot, error) {
	out := &DeletedQueueSlot{}
	if err := s.client.json(ctx, "DELETE", "/queue/slots/"+url.PathEscape(id), nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Next returns the time a post added to the queue now would get, or nil when
// no slot applies. An empty accountID asks across every slot.
func (s *QueueService) Next(ctx context.Context, workspaceID, accountID string) (*Time, error) {
	q := newQuery()
	q.str("workspace_id", workspaceID)
	q.str("account_id", accountID)
	out := struct {
		Next *Time `json:"next"`
	}{}
	if err := s.client.json(ctx, "GET", "/queue/next", nil, q.values(), &out); err != nil {
		return nil, err
	}
	if out.Next == nil || out.Next.IsZero() {
		return nil, nil
	}
	return out.Next, nil
}

// Preview returns the next free times, count of them (the API defaults to 5,
// at most 20). An empty accountID asks across every slot.
func (s *QueueService) Preview(ctx context.Context, workspaceID, accountID string, count int) ([]Time, error) {
	q := newQuery()
	q.str("workspace_id", workspaceID)
	q.str("account_id", accountID)
	q.num("count", count)
	var out []Time
	if err := s.client.json(ctx, "GET", "/queue/preview", nil, q.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}
