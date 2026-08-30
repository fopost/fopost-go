package fopost

import (
	"context"
	"net/url"
)

// WebhooksService covers outbound webhooks, the push counterpart to polling
// a post's deliveries.
type WebhooksService struct{ client *Client }

// Webhook is one subscription.
type Webhook struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspaceId"`
	URL         string `json:"url"`
	// Events are the Event constants this subscription asked for.
	Events          []string `json:"events"`
	Active          bool     `json:"active"`
	LastTriggeredAt Time     `json:"lastTriggeredAt"`
	FailureCount    int      `json:"failureCount"`
	CreatedAt       Time     `json:"createdAt"`
}

// CreatedWebhook is a new subscription. Secret is shown once, at creation, and
// signs every delivery — store it now.
type CreatedWebhook struct {
	ID          string   `json:"id"`
	WorkspaceID string   `json:"workspaceId"`
	URL         string   `json:"url"`
	Secret      string   `json:"secret"`
	Events      []string `json:"events"`
	Active      bool     `json:"active"`
	CreatedAt   Time     `json:"createdAt"`
}

// List returns the webhooks the key can reach.
func (s *WebhooksService) List(ctx context.Context) ([]Webhook, error) {
	var out []Webhook
	if err := s.client.json(ctx, "GET", "/webhooks", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateWebhookRequest is the body of Create.
type CreateWebhookRequest struct {
	WorkspaceID string `json:"workspaceId"`
	URL         string `json:"url"`
	// Events are the Event constants to subscribe to.
	Events []string `json:"events"`
}

// Create subscribes an endpoint to a workspace's events.
func (s *WebhooksService) Create(ctx context.Context, body *CreateWebhookRequest) (*CreatedWebhook, error) {
	out := &CreatedWebhook{}
	if err := s.client.json(ctx, "POST", "/webhooks", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateWebhookRequest is the body of Update. Only the fields you set are sent.
type UpdateWebhookRequest struct {
	URL    string   `json:"url,omitempty"`
	Events []string `json:"events,omitempty"`
	Active *bool    `json:"active,omitempty"`
}

// Update changes a subscription's endpoint, events, or active flag.
func (s *WebhooksService) Update(ctx context.Context, id string, body *UpdateWebhookRequest) (*Webhook, error) {
	out := &Webhook{}
	if err := s.client.json(ctx, "PUT", "/webhooks/"+url.PathEscape(id), body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Delete removes a subscription.
func (s *WebhooksService) Delete(ctx context.Context, id string) error {
	return s.client.Do(ctx, "DELETE", "/webhooks/"+url.PathEscape(id), nil, nil, nil)
}

// Test sends a sample event to the subscribed endpoint.
func (s *WebhooksService) Test(ctx context.Context, id string) (*Message, error) {
	out := &Message{}
	if err := s.client.Do(ctx, "POST", "/webhooks/"+url.PathEscape(id)+"/test", nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}
