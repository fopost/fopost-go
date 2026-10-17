package fopost

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

const broadcastBody = `{
  "id": "bc_1",
  "name": "September check-in",
  "text": "New colours just landed.",
  "account_id": "acc_1",
  "audience": {"platforms": ["instagram"]},
  "status": "sent",
  "scheduled_at": null,
  "sent_at": "2026-09-19T10:04:00.000Z",
  "created_at": "2026-09-19T09:58:00.000Z",
  "counts": {"total": 3, "sent": 2, "skipped": 1, "failed": 0, "pending": 0}
}`

func TestBroadcastsListReadsThePaginationBlock(t *testing.T) {
	var path, query string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		path, query = r.URL.Path, r.URL.RawQuery
		_, _ = io.WriteString(w, `{"data":[`+broadcastBody+`],"pagination":{"page":2,"per_page":10,"total":11}}`)
	})

	page, err := client.Broadcasts.List(context.Background(), &ListBroadcastsParams{
		WorkspaceID: "ws_1", Status: BroadcastSent, Page: 2, PerPage: 10,
	})
	if err != nil {
		t.Fatalf("Broadcasts.List: %v", err)
	}
	if path != "/broadcasts" {
		t.Fatalf("path = %q", path)
	}
	if query != "page=2&per_page=10&status=sent&workspace_id=ws_1" {
		t.Fatalf("query = %q", query)
	}
	if len(page.Data) != 1 || page.Data[0].Name != "September check-in" {
		t.Fatalf("data = %+v", page.Data)
	}
	if page.Data[0].Counts.Sent != 2 || page.Data[0].Counts.Skipped != 1 {
		t.Fatalf("counts = %+v", page.Data[0].Counts)
	}
	if page.Pagination.Total != 11 {
		t.Fatalf("pagination = %+v", page.Pagination)
	}
}

func TestBroadcastsCreateSendsTheSnakeCaseBody(t *testing.T) {
	var body map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"data":`+broadcastBody+`}`)
	})

	at := Time{}
	if err := at.UnmarshalJSON([]byte(`"2026-10-01T09:00:00.000Z"`)); err != nil {
		t.Fatalf("time: %v", err)
	}
	_, err := client.Broadcasts.Create(context.Background(), &CreateBroadcastRequest{
		WorkspaceID: "ws_1",
		AccountID:   "acc_1",
		Name:        "September check-in",
		Text:        "New colours just landed.",
		Audience:    &AudienceFilter{Platforms: []string{"instagram"}},
		ScheduledAt: &at,
	})
	if err != nil {
		t.Fatalf("Broadcasts.Create: %v", err)
	}
	if body["workspace_id"] != "ws_1" || body["account_id"] != "acc_1" {
		t.Fatalf("body = %+v", body)
	}
	audience, ok := body["audience"].(map[string]any)
	if !ok || audience["platforms"] == nil {
		t.Fatalf("audience = %+v", body["audience"])
	}
}

// A closed messaging window has to be readable, or a non-send is a mystery.
func TestBroadcastRecipientKeepsItsSkipReason(t *testing.T) {
	var query string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		_, _ = io.WriteString(w, `{"data":[{"contact_id":"con_1","display_name":"Sam Rivera",`+
			`"status":"skipped","skip_reason":"window_closed","sent_at":null,"error":null}],`+
			`"pagination":{"page":1,"per_page":50,"total":1}}`)
	})

	page, err := client.Broadcasts.Recipients(context.Background(), "bc_1", &ListRecipientsParams{
		Status: RecipientSkipped,
	})
	if err != nil {
		t.Fatalf("Broadcasts.Recipients: %v", err)
	}
	if query != "status=skipped" {
		t.Fatalf("query = %q", query)
	}
	if page.Data[0].Status != RecipientSkipped {
		t.Fatalf("status = %q", page.Data[0].Status)
	}
	if page.Data[0].SkipReason != SkipWindowClosed {
		t.Fatalf("skip reason = %q", page.Data[0].SkipReason)
	}
}

func TestBroadcastSendAndCancelPostToTheirOwnPaths(t *testing.T) {
	var path string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		if path == "/broadcasts/bc_1/send" {
			_, _ = io.WriteString(w, `{"data":{"id":"bc_1","status":"sending","recipients":3}}`)
			return
		}
		_, _ = io.WriteString(w, `{"data":{"id":"bc_1","status":"cancelled"}}`)
	})

	sent, err := client.Broadcasts.Send(context.Background(), "bc_1")
	if err != nil {
		t.Fatalf("Broadcasts.Send: %v", err)
	}
	if path != "/broadcasts/bc_1/send" || sent.Recipients != 3 {
		t.Fatalf("path = %q, result = %+v", path, sent)
	}

	cancelled, err := client.Broadcasts.Cancel(context.Background(), "bc_1")
	if err != nil {
		t.Fatalf("Broadcasts.Cancel: %v", err)
	}
	if path != "/broadcasts/bc_1/cancel" || cancelled.Status != BroadcastCancelled {
		t.Fatalf("path = %q, result = %+v", path, cancelled)
	}
}

func TestSequenceStepsTravelAsGiven(t *testing.T) {
	var body map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"data":{"id":"seq_1","name":"Welcome","account_id":"acc_1",`+
			`"steps":[{"delay_hours":0,"text":"Hi"},{"delay_hours":48,"text":"Still here?"}],`+
			`"status":"active","created_at":"2026-09-12T08:00:00.000Z"}}`)
	})

	sequence, err := client.Sequences.Create(context.Background(), &CreateSequenceRequest{
		WorkspaceID: "ws_1",
		AccountID:   "acc_1",
		Name:        "Welcome",
		Steps:       []SequenceStep{{DelayHours: 0, Text: "Hi"}},
	})
	if err != nil {
		t.Fatalf("Sequences.Create: %v", err)
	}
	if sequence.Steps[1].DelayHours != 48 {
		t.Fatalf("steps = %+v", sequence.Steps)
	}
	steps, ok := body["steps"].([]any)
	if !ok || len(steps) != 1 {
		t.Fatalf("body steps = %+v", body["steps"])
	}
	step, _ := steps[0].(map[string]any)
	if step["text"] != "Hi" || step["delay_hours"] != float64(0) {
		t.Fatalf("step = %+v", step)
	}
}

func TestSequenceEnrollTakesIDsOrAnAudience(t *testing.T) {
	var body map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		body = nil
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = io.WriteString(w, `{"data":{"id":"seq_1","enrolled":2}}`)
	})

	if _, err := client.Sequences.Enroll(context.Background(), "seq_1", &EnrollRequest{
		ContactIDs: []string{"con_1", "con_2"},
	}); err != nil {
		t.Fatalf("Sequences.Enroll: %v", err)
	}
	ids, ok := body["contact_ids"].([]any)
	if !ok || len(ids) != 2 {
		t.Fatalf("contact_ids = %+v", body["contact_ids"])
	}

	if _, err := client.Sequences.Enroll(context.Background(), "seq_1", &EnrollRequest{
		Audience: &AudienceFilter{Platforms: []string{"telegram"}},
	}); err != nil {
		t.Fatalf("Sequences.Enroll: %v", err)
	}
	if body["contact_ids"] != nil {
		t.Fatalf("an audience-only enroll must not send contact_ids: %+v", body)
	}
	if body["audience"] == nil {
		t.Fatalf("audience = %+v", body)
	}
}

func TestSequenceUnenrollNamesTheContactsItStops(t *testing.T) {
	var path string
	var body map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = io.WriteString(w, `{"data":{"id":"seq_1","stopped":1}}`)
	})

	result, err := client.Sequences.Unenroll(context.Background(), "seq_1", []string{"con_1"})
	if err != nil {
		t.Fatalf("Sequences.Unenroll: %v", err)
	}
	if path != "/sequences/seq_1/unenroll" || result.Stopped != 1 {
		t.Fatalf("path = %q, result = %+v", path, result)
	}
	ids, ok := body["contact_ids"].([]any)
	if !ok || len(ids) != 1 || ids[0] != "con_1" {
		t.Fatalf("contact_ids = %+v", body["contact_ids"])
	}
}
