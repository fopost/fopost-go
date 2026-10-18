package fopost

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestConnectSessionsRoundTrip(t *testing.T) {
	var method, path, raw string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		b, _ := io.ReadAll(r.Body)
		raw = string(b)
		if r.Method == "POST" {
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, `{"data":{"id":"cs_1","status":"pending","external_user_id":"user-42","platforms":["linkedin","threads"],"redirect_url":"https://yourbrand.com/connected","branding":{"name":"Your Brand","logo_url":null},"account_id":null,"url":"https://connect.fopost.com/s/cs_1","expires_at":"2026-10-18T10:00:00.000Z","used_at":null,"created_at":"2026-10-18T09:00:00.000Z"}}`)
			return
		}
		_, _ = io.WriteString(w, `{"data":{"id":"cs_1","status":"used","external_user_id":"user-42","platforms":["linkedin","threads"],"redirect_url":null,"branding":{"name":null,"logo_url":null},"account_id":"acc_9","expires_at":"2026-10-18T10:00:00.000Z","used_at":"2026-10-18T09:12:00.000Z","created_at":"2026-10-18T09:00:00.000Z"}}`)
	})
	ctx := context.Background()

	session, err := client.ConnectSessions.Create(ctx, &CreateConnectSessionRequest{
		WorkspaceID:    "ws_1",
		ExternalUserID: "user-42",
		Platforms:      []string{"linkedin", "threads"},
		RedirectURL:    "https://yourbrand.com/connected",
		Branding:       &CreateConnectSessionBranding{Name: "Your Brand"},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if method != "POST" || path != "/connect-sessions" {
		t.Fatalf("Create sent %s %s", method, path)
	}
	for _, want := range []string{`"workspaceId":"ws_1"`, `"externalUserId":"user-42"`, `"platforms":["linkedin","threads"]`, `"redirectUrl":"https://yourbrand.com/connected"`, `"branding":{"name":"Your Brand"}`} {
		if !strings.Contains(raw, want) {
			t.Fatalf("body %s lacks %s", raw, want)
		}
	}
	if session.ID != "cs_1" || session.Status != ConnectSessionPending || session.URL != "https://connect.fopost.com/s/cs_1" || session.AccountID != nil || *session.Branding.Name != "Your Brand" {
		t.Fatalf("session = %+v", session)
	}

	fetched, err := client.ConnectSessions.Get(ctx, "cs_1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if method != "GET" || path != "/connect-sessions/cs_1" {
		t.Fatalf("Get sent %s %s", method, path)
	}
	if fetched.Status != ConnectSessionUsed || *fetched.AccountID != "acc_9" || fetched.URL != "" || fetched.UsedAt.IsZero() {
		t.Fatalf("fetched = %+v", fetched)
	}
}
