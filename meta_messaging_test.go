package fopost

import (
	"context"
	"io"
	"net/http"
	"testing"
)

func TestAccountsIceBreakersRoundTrip(t *testing.T) {
	var method, path, raw string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		b, _ := io.ReadAll(r.Body)
		raw = string(b)
		if r.Method == "DELETE" {
			_, _ = io.WriteString(w, `{"data":{"ice_breakers":[]}}`)
			return
		}
		_, _ = io.WriteString(w, `{"data":{"ice_breakers":[{"question":"What are your hours?","payload":"HOURS"}]}}`)
	})
	ctx := context.Background()

	got, err := client.Accounts.GetIceBreakers(ctx, "acc_1")
	if err != nil || method != "GET" || path != "/accounts/acc_1/messaging/ice-breakers" || len(got.IceBreakers) != 1 {
		t.Fatalf("get: %v %s %s %+v", err, method, path, got)
	}

	set, err := client.Accounts.SetIceBreakers(ctx, "acc_1", []MetaIceBreaker{{Question: "What are your hours?", Payload: "HOURS"}})
	if err != nil || method != "PUT" || raw != `{"ice_breakers":[{"question":"What are your hours?","payload":"HOURS"}]}` {
		t.Fatalf("set: %v %s %s", err, method, raw)
	}
	if set.IceBreakers[0].Payload != "HOURS" {
		t.Fatalf("set = %+v", set)
	}

	cleared, err := client.Accounts.DeleteIceBreakers(ctx, "acc_1")
	if err != nil || method != "DELETE" || len(cleared.IceBreakers) != 0 {
		t.Fatalf("delete: %v %s %+v", err, method, cleared)
	}
}

func TestAccountsPersistentMenuAndGreetingRoutes(t *testing.T) {
	var method, path, raw string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		b, _ := io.ReadAll(r.Body)
		raw = string(b)
		if path == "/accounts/acc_1/messaging/greeting" {
			_, _ = io.WriteString(w, `{"data":{"greeting":[{"locale":"default","text":"Hi!"}]}}`)
			return
		}
		_, _ = io.WriteString(w, `{"data":{"persistent_menu":[{"locale":"default","call_to_actions":[{"type":"web_url","title":"Shop","url":"https://example.com/shop"}]}]}}`)
	})
	ctx := context.Background()

	menu := []MetaPersistentMenuEntry{{
		Locale:        "default",
		CallToActions: []MetaMenuItem{{Type: "web_url", Title: "Shop", URL: "https://example.com/shop"}},
	}}
	set, err := client.Accounts.SetPersistentMenu(ctx, "acc_1", menu)
	if err != nil || method != "PUT" || path != "/accounts/acc_1/messaging/persistent-menu" {
		t.Fatalf("set menu: %v %s %s", err, method, path)
	}
	// A postback-less item must not carry an empty payload key.
	if raw != `{"persistent_menu":[{"locale":"default","call_to_actions":[{"type":"web_url","title":"Shop","url":"https://example.com/shop"}]}]}` {
		t.Fatalf("set menu body = %s", raw)
	}
	if set.PersistentMenu[0].CallToActions[0].URL != "https://example.com/shop" {
		t.Fatalf("set menu = %+v", set)
	}

	greeting, err := client.Accounts.SetGreeting(ctx, "acc_1", []MetaGreetingText{{Locale: "default", Text: "Hi!"}})
	if err != nil || raw != `{"greeting":[{"locale":"default","text":"Hi!"}]}` {
		t.Fatalf("set greeting: %v %s", err, raw)
	}
	if greeting.Greeting[0].Text != "Hi!" {
		t.Fatalf("greeting = %+v", greeting)
	}
}

func TestAccountsWebhookSubscriptionReportsAndResubscribes(t *testing.T) {
	var method, path string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		if r.Method == "POST" {
			_, _ = io.WriteString(w, `{"data":{"subscribed":true,"fields":["feed","messages"],"missing_fields":[]}}`)
			return
		}
		_, _ = io.WriteString(w, `{"data":{"subscribed":false,"fields":["feed"],"missing_fields":["messages"]}}`)
	})
	ctx := context.Background()

	lapsed, err := client.Accounts.GetWebhookSubscription(ctx, "acc_1")
	if err != nil || method != "GET" || path != "/accounts/acc_1/webhook-subscription" {
		t.Fatalf("get: %v %s %s", err, method, path)
	}
	if lapsed.Subscribed || len(lapsed.MissingFields) != 1 || lapsed.MissingFields[0] != "messages" {
		t.Fatalf("lapsed = %+v", lapsed)
	}

	fixed, err := client.Accounts.ResubscribeWebhook(ctx, "acc_1")
	if err != nil || method != "POST" || !fixed.Subscribed {
		t.Fatalf("resubscribe: %v %s %+v", err, method, fixed)
	}
}

func TestInboxHandoverPassesAndTakesControl(t *testing.T) {
	var path, raw string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		raw = string(b)
		_, _ = io.WriteString(w, `{"data":{"app_id":"263902037430900","control":"passed"}}`)
	})
	ctx := context.Background()

	passed, err := client.Inbox.Handover(ctx, "t_1", "acc_1", &HandoverOptions{AppID: "263902037430900"})
	if err != nil || path != "/inbox/conversations/t_1/handover" {
		t.Fatalf("pass: %v %s", err, path)
	}
	if raw != `{"account_id":"acc_1","app_id":"263902037430900"}` {
		t.Fatalf("pass body = %s", raw)
	}
	if passed.Control != "passed" {
		t.Fatalf("passed = %+v", passed)
	}

	if _, err := client.Inbox.Handover(ctx, "t_1", "acc_1", nil); err != nil {
		t.Fatalf("take: %v", err)
	}
	if raw != `{"account_id":"acc_1"}` {
		t.Fatalf("take body = %s", raw)
	}
}
