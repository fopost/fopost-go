package fopost

import (
	"context"
	"io"
	"net/http"
	"testing"
)

func TestWhatsAppCreateTemplateReturnsTheReviewStatusThePlatformGaveIt(t *testing.T) {
	var path string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		_, _ = io.WriteString(w, `{"data":{"id":"tpl-1","name":"order_shipped","language":"en_US","category":"UTILITY","status":"PENDING","rejectedReason":null,"components":[],"qualityScore":null}}`)
	})

	template, err := client.WhatsApp.CreateTemplate(context.Background(), "acc_1", &CreateWhatsAppTemplateRequest{
		Name:       "order_shipped",
		Language:   "en_US",
		Category:   "UTILITY",
		Components: []map[string]any{{"type": "BODY", "text": "On its way."}},
	})
	if err != nil {
		t.Fatalf("WhatsApp.CreateTemplate: %v", err)
	}
	if path != "/accounts/acc_1/whatsapp/templates" {
		t.Fatalf("path = %q", path)
	}
	// Nothing marks a template approved but the platform.
	if template.Status != "PENDING" {
		t.Fatalf("status = %q", template.Status)
	}
}

func TestWhatsAppDeleteTemplateNamesItInTheQuery(t *testing.T) {
	var query string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		_, _ = io.WriteString(w, `{"data":{"deleted":true}}`)
	})

	if err := client.WhatsApp.DeleteTemplate(context.Background(), "acc_1", "tpl-1", "order_shipped"); err != nil {
		t.Fatalf("WhatsApp.DeleteTemplate: %v", err)
	}
	if query != "name=order_shipped" {
		t.Fatalf("query = %q", query)
	}
}

func TestWhatsAppSandboxSessionCarriesOnlyTheLastFourDigits(t *testing.T) {
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"data":{"id":"ses-1","status":"invited","phoneNumberLast4":"4567","invitedAt":"2026-09-20T10:00:00Z","activatedAt":null,"expiresAt":"2026-09-21T10:00:00Z"}}`)
	})

	session, err := client.WhatsApp.CreateSandboxSession(context.Background(), "ws_1", "+15551234567")
	if err != nil {
		t.Fatalf("WhatsApp.CreateSandboxSession: %v", err)
	}
	if session.PhoneNumberLast4 != "4567" || session.Status != "invited" {
		t.Fatalf("session = %+v", session)
	}
}

func TestWhatsAppIsAPlatformConstant(t *testing.T) {
	if PlatformWhatsApp != "whatsapp" {
		t.Fatalf("PlatformWhatsApp = %q", PlatformWhatsApp)
	}
}
