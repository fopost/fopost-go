package fopost

import (
	"context"
	"net/url"
	"strconv"
)

// AutomationsService covers automations: a trigger plus the steps it runs.
type AutomationsService struct{ client *Client }

// Automation trigger types.
const (
	TriggerCrossPost  = "cross_post"
	TriggerRSSFeed    = "rss_feed"
	TriggerAPIWebhook = "api_webhook"
	TriggerSchedule   = "schedule"
)

// Automation step action types.
const (
	ActionPublish   = "publish"
	ActionDelay     = "delay"
	ActionTransform = "transform"
)

// AutomationStep is one action in an automation, in position order.
type AutomationStep struct {
	ID       int `json:"id,omitempty"`
	Position int `json:"position,omitempty"`
	// ActionType is one of the Action constants.
	ActionType   string         `json:"actionType"`
	ActionConfig map[string]any `json:"actionConfig,omitempty"`
}

// Automation is one trigger and the steps behind it.
type Automation struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspaceId"`
	Name        string `json:"name"`
	// TriggerType is one of the Trigger constants.
	TriggerType     string           `json:"triggerType"`
	TriggerConfig   map[string]any   `json:"triggerConfig"`
	Active          bool             `json:"active"`
	LastTriggeredAt Time             `json:"lastTriggeredAt"`
	RunCount        int              `json:"runCount"`
	Steps           []AutomationStep `json:"steps"`
	// Secret is returned once, by Create, for an api_webhook trigger.
	Secret    string `json:"secret,omitempty"`
	CreatedAt Time   `json:"createdAt"`
	UpdatedAt Time   `json:"updatedAt"`
}

// List returns the automations the key can reach.
func (s *AutomationsService) List(ctx context.Context) ([]Automation, error) {
	var out []Automation
	if err := s.client.json(ctx, "GET", "/automations", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Get returns one automation with its steps.
func (s *AutomationsService) Get(ctx context.Context, id string) (*Automation, error) {
	out := &Automation{}
	if err := s.client.json(ctx, "GET", "/automations/"+url.PathEscape(id), nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateAutomationRequest is the body of Create.
type CreateAutomationRequest struct {
	WorkspaceID   string           `json:"workspaceId"`
	Name          string           `json:"name"`
	TriggerType   string           `json:"triggerType"`
	TriggerConfig map[string]any   `json:"triggerConfig,omitempty"`
	Steps         []AutomationStep `json:"steps"`
	Active        *bool            `json:"active,omitempty"`
}

// Create adds an automation. For an api_webhook trigger, the response carries
// the signing secret once — store it now.
func (s *AutomationsService) Create(ctx context.Context, body *CreateAutomationRequest) (*Automation, error) {
	out := &Automation{}
	if err := s.client.json(ctx, "POST", "/automations", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateAutomationRequest is the body of Update. Only the fields you set are
// sent, but Steps replaces the whole list when given.
type UpdateAutomationRequest struct {
	Name          string           `json:"name,omitempty"`
	TriggerConfig map[string]any   `json:"triggerConfig,omitempty"`
	Steps         []AutomationStep `json:"steps,omitempty"`
	Active        *bool            `json:"active,omitempty"`
}

// Update edits an automation.
func (s *AutomationsService) Update(ctx context.Context, id string, body *UpdateAutomationRequest) (*Automation, error) {
	out := &Automation{}
	if err := s.client.json(ctx, "PUT", "/automations/"+url.PathEscape(id), body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Delete removes an automation.
func (s *AutomationsService) Delete(ctx context.Context, id string) error {
	return s.client.Do(ctx, "DELETE", "/automations/"+url.PathEscape(id), nil, nil, nil)
}

// ToggleResult is the automation's active flag after toggling.
type ToggleResult struct {
	ID     string `json:"id"`
	Active bool   `json:"active"`
}

// Toggle switches an automation on or off.
func (s *AutomationsService) Toggle(ctx context.Context, id string) (*ToggleResult, error) {
	out := &ToggleResult{}
	if err := s.client.json(ctx, "POST", "/automations/"+url.PathEscape(id)+"/toggle", nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// AutomationRun is one execution of an automation.
type AutomationRun struct {
	ID           int            `json:"id"`
	AutomationID string         `json:"automationId,omitempty"`
	Status       string         `json:"status"`
	CurrentStep  *int           `json:"currentStep"`
	TriggerEvent map[string]any `json:"triggerEvent"`
	Context      map[string]any `json:"context,omitempty"`
	StartedAt    Time           `json:"startedAt"`
	CompletedAt  Time           `json:"completedAt"`
	ErrorMessage string         `json:"errorMessage"`
	Logs         []struct {
		ID             int            `json:"id"`
		StepPosition   int            `json:"stepPosition"`
		Status         string         `json:"status"`
		InputSnapshot  map[string]any `json:"inputSnapshot"`
		OutputSnapshot map[string]any `json:"outputSnapshot"`
		StartedAt      Time           `json:"startedAt"`
		CompletedAt    Time           `json:"completedAt"`
		DurationMs     *int           `json:"durationMs"`
		ErrorMessage   string         `json:"errorMessage"`
	} `json:"logs,omitempty"`
}

// AutomationRunList is one page of runs.
type AutomationRunList struct {
	Data []AutomationRun `json:"data"`
	Meta PageMeta        `json:"meta"`
}

// Runs lists an automation's executions. Zero page or perPage leaves the API's
// defaults in place.
func (s *AutomationsService) Runs(ctx context.Context, id string, page, perPage int) (*AutomationRunList, error) {
	q := newQuery()
	q.num("page", page)
	q.num("per_page", perPage)
	out := &AutomationRunList{}
	if err := s.client.Do(ctx, "GET", "/automations/"+url.PathEscape(id)+"/runs", nil, q.values(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// Run returns one execution, with a log per step.
func (s *AutomationsService) Run(ctx context.Context, id string, runID int) (*AutomationRun, error) {
	path := "/automations/" + url.PathEscape(id) + "/runs/" + strconv.Itoa(runID)
	out := &AutomationRun{}
	if err := s.client.json(ctx, "GET", path, nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// TriggerResult is the run a webhook trigger started.
type TriggerResult struct {
	RunID     int  `json:"runId"`
	Triggered bool `json:"triggered"`
}

// Trigger fires an api_webhook automation with a payload its steps can read.
func (s *AutomationsService) Trigger(ctx context.Context, id string, payload map[string]any) (*TriggerResult, error) {
	if payload == nil {
		payload = map[string]any{}
	}
	out := &TriggerResult{}
	if err := s.client.json(ctx, "POST", "/automations/"+url.PathEscape(id)+"/trigger", payload, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// AutomationStats is the roll-up behind the automations dashboard.
type AutomationStats struct {
	TotalAutomations  int            `json:"totalAutomations"`
	ActiveAutomations int            `json:"activeAutomations"`
	TotalRuns24h      int            `json:"totalRuns24h"`
	RunsByStatus      map[string]int `json:"runsByStatus"`
	RecentRuns        []struct {
		ID           int    `json:"id"`
		AutomationID string `json:"automationId"`
		Status       string `json:"status"`
		StartedAt    Time   `json:"startedAt"`
		CompletedAt  Time   `json:"completedAt"`
	} `json:"recentRuns"`
}

// Stats returns automation counts and recent runs.
func (s *AutomationsService) Stats(ctx context.Context) (*AutomationStats, error) {
	out := &AutomationStats{}
	if err := s.client.json(ctx, "GET", "/automations/stats", nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}
