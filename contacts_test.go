package fopost

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

const contactBody = `{
  "id": "con_1",
  "display_name": "Ada Okafor",
  "channels": [
    {"platform": "instagram", "handle": "adaokafor", "externalId": "178414"},
    {"platform": "x", "handle": "ada_writes", "externalId": null}
  ],
  "source": "inbox",
  "note": null,
  "first_seen_at": "2026-04-02T09:14:00.000Z",
  "last_seen_at": "2026-09-18T14:30:00.000Z",
  "fields": {"plan_tier": "Pro"},
  "labels": [{"id": "lbl_1", "name": "VIP", "color": "#0070f3"}]
}`

func TestContactsListReadsThePaginationBlock(t *testing.T) {
	var path, query string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		path, query = r.URL.Path, r.URL.RawQuery
		_, _ = io.WriteString(w, `{"data":[`+contactBody+`],"pagination":{"page":2,"per_page":10,"total":11}}`)
	})

	page, err := client.Contacts.List(context.Background(), &ListContactsParams{
		WorkspaceID: "ws_1", Search: "ada", Page: 2, PerPage: 10,
	})
	if err != nil {
		t.Fatalf("Contacts.List: %v", err)
	}
	if path != "/contacts" {
		t.Fatalf("path = %q", path)
	}
	if query != "page=2&per_page=10&search=ada&workspace_id=ws_1" {
		t.Fatalf("query = %q", query)
	}
	if len(page.Data) != 1 || page.Data[0].DisplayName != "Ada Okafor" {
		t.Fatalf("data = %+v", page.Data)
	}
	if page.Data[0].Channels[0].ExternalID != "178414" {
		t.Fatalf("channel = %+v", page.Data[0].Channels[0])
	}
	if page.Data[0].Fields["plan_tier"] != "Pro" {
		t.Fatalf("fields = %+v", page.Data[0].Fields)
	}
	if page.Pagination.Total != 11 || page.Pagination.Page != 2 {
		t.Fatalf("pagination = %+v", page.Pagination)
	}
}

func TestContactsCreateSendsTheWireNames(t *testing.T) {
	var body map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = io.WriteString(w, `{"data":`+contactBody+`}`)
	})

	if _, err := client.Contacts.Create(context.Background(), &CreateContactRequest{
		WorkspaceID: "ws_1",
		Channels:    []ContactChannel{{Platform: "x", Handle: "ada_writes"}},
		DisplayName: "Ada Okafor",
		Fields:      map[string]string{"plan_tier": "Pro"},
	}); err != nil {
		t.Fatalf("Contacts.Create: %v", err)
	}
	if body["workspace_id"] != "ws_1" || body["display_name"] != "Ada Okafor" {
		t.Fatalf("body = %+v", body)
	}
	// An empty externalId must not travel: it would claim the platform gave us one.
	channel := body["channels"].([]any)[0].(map[string]any)
	if _, ok := channel["externalId"]; ok {
		t.Fatalf("channel carried an empty externalId: %+v", channel)
	}
}

func TestContactsUpdateSendsOnlyWhatWasSetAndKeepsANullField(t *testing.T) {
	var raw []byte
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ = io.ReadAll(r.Body)
		_, _ = io.WriteString(w, `{"data":`+contactBody+`}`)
	})

	if _, err := client.Contacts.Update(context.Background(), "con_1", &UpdateContactRequest{
		Fields: map[string]*string{"region": nil},
	}); err != nil {
		t.Fatalf("Contacts.Update: %v", err)
	}
	if string(raw) != `{"fields":{"region":null}}` {
		t.Fatalf("body = %s", raw)
	}
}

func TestContactsConversationsReadsTheThreads(t *testing.T) {
	var path, query string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		path, query = r.URL.Path, r.URL.RawQuery
		_, _ = io.WriteString(w, `{"data":[{"key":"t_182736","account_id":"acc_1","account_username":"yourbrand","platform":"instagram","messages":14,"received":9,"sent":5,"last_message_at":"2026-09-18T14:30:00.000Z","last_item_id":"inb_1"}]}`)
	})

	rows, err := client.Contacts.Conversations(context.Background(), "con_1", 10)
	if err != nil {
		t.Fatalf("Contacts.Conversations: %v", err)
	}
	if path != "/contacts/con_1/conversations" || query != "limit=10" {
		t.Fatalf("path = %q query = %q", path, query)
	}
	if len(rows) != 1 || rows[0].Key != "t_182736" || rows[0].Received != 9 {
		t.Fatalf("rows = %+v", rows)
	}
}

func TestContactsImportReportsMergedAndSkipped(t *testing.T) {
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"data":{"created":1,"merged":2,"skipped":[{"row":4,"reason":"platform and handle are both required"}],"unknownColumns":["lifetime_value"]}}`)
	})

	result, err := client.Contacts.Import(context.Background(), "ws_1", "platform,handle\nx,ada_writes")
	if err != nil {
		t.Fatalf("Contacts.Import: %v", err)
	}
	if result.Created != 1 || result.Merged != 2 {
		t.Fatalf("result = %+v", result)
	}
	if len(result.Skipped) != 1 || result.Skipped[0].Row != 4 {
		t.Fatalf("skipped = %+v", result.Skipped)
	}
	if len(result.UnknownColumns) != 1 || result.UnknownColumns[0] != "lifetime_value" {
		t.Fatalf("unknownColumns = %+v", result.UnknownColumns)
	}
}

func TestContactsCreateFieldPutsTheWorkspaceOnTheQuery(t *testing.T) {
	var path, query string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		path, query = r.URL.Path, r.URL.RawQuery
		_, _ = io.WriteString(w, `{"data":{"id":"fld_1","key":"plan_tier","name":"Plan Tier","type":"select","options":["Free","Pro"],"position":0}}`)
	})

	field, err := client.Contacts.CreateField(context.Background(), "ws_1", &CreateContactFieldRequest{
		Key: "plan_tier", Name: "Plan Tier", Type: ContactFieldSelect, Options: []string{"Free", "Pro"},
	})
	if err != nil {
		t.Fatalf("Contacts.CreateField: %v", err)
	}
	if path != "/contacts/fields" || query != "workspace_id=ws_1" {
		t.Fatalf("path = %q query = %q", path, query)
	}
	if field.Key != "plan_tier" || len(field.Options) != 2 {
		t.Fatalf("field = %+v", field)
	}
}

func TestContactsConversationAnalyticsReadsTheAnalyticsRoute(t *testing.T) {
	var path, query string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		path, query = r.URL.Path, r.URL.RawQuery
		_, _ = io.WriteString(w, `{"data":{"conversations":[{"key":"t_1","accountId":"acc_1","platform":"instagram","received":9,"sent":5,"answered":5,"open":1,"medianResponseMinutes":47,"firstMessageAt":null,"lastMessageAt":null}],"total":128,"page":1,"perPage":25}}`)
	})

	report, err := client.Contacts.ConversationAnalytics(context.Background(), &ListConversationAnalyticsParams{
		Days: 30, Sort: ConversationSortSlowest,
	})
	if err != nil {
		t.Fatalf("Contacts.ConversationAnalytics: %v", err)
	}
	if path != "/analytics/inbox/conversations" || query != "days=30&sort=slowest" {
		t.Fatalf("path = %q query = %q", path, query)
	}
	if report.Total != 128 {
		t.Fatalf("total = %d", report.Total)
	}
	if report.Conversations[0].MedianResponseMinutes == nil || *report.Conversations[0].MedianResponseMinutes != 47 {
		t.Fatalf("median = %+v", report.Conversations[0].MedianResponseMinutes)
	}
}
