package fopost

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestQueueSlotsRoundTrip(t *testing.T) {
	var method, path, query, raw string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, path, query = r.Method, r.URL.Path, r.URL.RawQuery
		b, _ := io.ReadAll(r.Body)
		raw = string(b)
		switch {
		case r.Method == "POST":
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, `{"data":{"id":"slot_1","workspace_id":"ws_1","account_id":null,"weekday":1,"time":"09:00","timezone":"Europe/Zagreb","created_at":"2026-10-18T09:00:00.000Z"}}`)
		case r.Method == "PUT":
			_, _ = io.WriteString(w, `{"data":{"id":"slot_1","workspace_id":"ws_1","account_id":"acc_1","weekday":0,"time":"10:30","timezone":"Europe/Zagreb","created_at":"2026-10-18T09:00:00.000Z","moved":[{"id":"post_1","from":"2026-10-19T07:00:00.000Z","to":"2026-10-25T08:30:00.000Z"}]}}`)
		case r.Method == "DELETE":
			_, _ = io.WriteString(w, `{"data":{"deleted":true,"moved":[]}}`)
		case strings.HasSuffix(r.URL.Path, "/next"):
			_, _ = io.WriteString(w, `{"data":{"next":"2026-10-19T07:00:00.000Z"}}`)
		case strings.HasSuffix(r.URL.Path, "/preview"):
			_, _ = io.WriteString(w, `{"data":["2026-10-19T07:00:00.000Z","2026-10-26T07:00:00.000Z"]}`)
		default:
			_, _ = io.WriteString(w, `{"data":[{"id":"slot_1","workspace_id":"ws_1","account_id":null,"weekday":1,"time":"09:00","timezone":"Europe/Zagreb","created_at":"2026-10-18T09:00:00.000Z"}]}`)
		}
	})
	ctx := context.Background()

	slots, err := client.Queue.ListSlots(ctx, "ws_1")
	if err != nil {
		t.Fatalf("ListSlots: %v", err)
	}
	if method != "GET" || path != "/queue/slots" || query != "workspace_id=ws_1" {
		t.Fatalf("ListSlots sent %s %s?%s", method, path, query)
	}
	if len(slots) != 1 || slots[0].ID != "slot_1" || slots[0].AccountID != nil || slots[0].Weekday != 1 {
		t.Fatalf("slots = %+v", slots)
	}

	created, err := client.Queue.CreateSlot(ctx, &CreateQueueSlotRequest{WorkspaceID: "ws_1", Weekday: 1, Time: "09:00", Timezone: "Europe/Zagreb"})
	if err != nil {
		t.Fatalf("CreateSlot: %v", err)
	}
	if method != "POST" || path != "/queue/slots" {
		t.Fatalf("CreateSlot sent %s %s", method, path)
	}
	if !strings.Contains(raw, `"weekday":1`) || !strings.Contains(raw, `"time":"09:00"`) || strings.Contains(raw, "account_id") {
		t.Fatalf("CreateSlot body = %s", raw)
	}
	if created.ID != "slot_1" || created.Timezone != "Europe/Zagreb" {
		t.Fatalf("created = %+v", created)
	}

	updated, err := client.Queue.UpdateSlot(ctx, "slot_1", &UpdateQueueSlotRequest{Weekday: Int(0), Time: "10:30", AccountID: String("acc_1")})
	if err != nil {
		t.Fatalf("UpdateSlot: %v", err)
	}
	if method != "PUT" || path != "/queue/slots/slot_1" {
		t.Fatalf("UpdateSlot sent %s %s", method, path)
	}
	if !strings.Contains(raw, `"weekday":0`) || strings.Contains(raw, "timezone") {
		t.Fatalf("UpdateSlot body = %s", raw)
	}
	if updated.Weekday != 0 || *updated.AccountID != "acc_1" || len(updated.Moved) != 1 || updated.Moved[0].ID != "post_1" {
		t.Fatalf("updated = %+v", updated)
	}

	deleted, err := client.Queue.DeleteSlot(ctx, "slot_1")
	if err != nil {
		t.Fatalf("DeleteSlot: %v", err)
	}
	if method != "DELETE" || path != "/queue/slots/slot_1" || !deleted.Deleted || len(deleted.Moved) != 0 {
		t.Fatalf("DeleteSlot sent %s %s, got %+v", method, path, deleted)
	}

	next, err := client.Queue.Next(ctx, "ws_1", "acc_1")
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if method != "GET" || path != "/queue/next" || query != "account_id=acc_1&workspace_id=ws_1" {
		t.Fatalf("Next sent %s %s?%s", method, path, query)
	}
	if next == nil || next.String() != "2026-10-19T07:00:00Z" {
		t.Fatalf("next = %v", next)
	}

	preview, err := client.Queue.Preview(ctx, "ws_1", "", 2)
	if err != nil {
		t.Fatalf("Preview: %v", err)
	}
	if method != "GET" || path != "/queue/preview" || query != "count=2&workspace_id=ws_1" {
		t.Fatalf("Preview sent %s %s?%s", method, path, query)
	}
	if len(preview) != 2 || preview[1].String() != "2026-10-26T07:00:00Z" {
		t.Fatalf("preview = %v", preview)
	}
}

func TestQueueNextIsNilWithoutASlot(t *testing.T) {
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"data":{"next":null}}`)
	})
	next, err := client.Queue.Next(context.Background(), "ws_1", "")
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if next != nil {
		t.Fatalf("next = %v, want nil", next)
	}
}

func TestCreatePostCanJoinTheQueue(t *testing.T) {
	var raw string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		raw = string(b)
		_, _ = io.WriteString(w, `{"data":{"id":"post_1","status":"scheduled","schedule_at":"2026-10-19T07:00:00.000Z"}}`)
	})
	post, err := client.Posts.Create(context.Background(), &CreatePostRequest{WorkspaceID: "ws_1", Accounts: []string{"acc_1"}, Content: Text("Hello"), Schedule: ScheduleQueue})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if !strings.Contains(raw, `"schedule":"queue"`) || strings.Contains(raw, "schedule_at") {
		t.Fatalf("body = %s", raw)
	}
	if post.Status != "scheduled" {
		t.Fatalf("post = %+v", post)
	}

	if _, err := client.Posts.Update(context.Background(), "post_1", &UpdatePostRequest{Schedule: ScheduleQueue}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if raw != `{"schedule":"queue"}` {
		t.Fatalf("update body = %s", raw)
	}
}
