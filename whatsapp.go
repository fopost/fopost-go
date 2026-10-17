package fopost

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// WhatsAppService reaches a WhatsApp Business connection.
//
// The platform owns templates, flows, the business profile and the commerce
// settings, so every method here is a live read or write against the customer's
// own WhatsApp Business Account. Nothing is cached, and all of it answers 503
// until WhatsApp is set up on the deployment. Every method needs the accounts
// scope, except the sandbox, which sends a template and needs publish.
type WhatsAppService struct{ client *Client }

// ─── Profile ────────────────────────────────────────────────────────────

// WhatsAppProfile is the business profile on a number, plus its standing.
type WhatsAppProfile struct {
	About             *string  `json:"about"`
	Address           *string  `json:"address"`
	Description       *string  `json:"description"`
	Email             *string  `json:"email"`
	Vertical          *string  `json:"vertical"`
	Websites          []string `json:"websites"`
	ProfilePictureURL *string  `json:"profilePictureUrl"`
	DisplayName       *string  `json:"displayName"`
	// DisplayNameStatus is the platform's review state for the display name.
	DisplayNameStatus  *string `json:"displayNameStatus"`
	Username           *string `json:"username"`
	QualityRating      *string `json:"qualityRating"`
	MessagingLimitTier *string `json:"messagingLimitTier"`
}

// UpdateWhatsAppProfileRequest is a partial update: omitted fields keep their value.
type UpdateWhatsAppProfileRequest struct {
	About       string   `json:"about,omitempty"`
	Address     string   `json:"address,omitempty"`
	Description string   `json:"description,omitempty"`
	Vertical    string   `json:"vertical,omitempty"`
	Websites    []string `json:"websites,omitempty"`
	// ProfilePictureMediaID is a library media id, uploaded first.
	ProfilePictureMediaID string `json:"profile_picture_media_id,omitempty"`
}

func (s *WhatsAppService) Profile(ctx context.Context, accountID string) (*WhatsAppProfile, error) {
	out := &WhatsAppProfile{}
	path := fmt.Sprintf("/accounts/%s/whatsapp/profile", url.PathEscape(accountID))
	if err := s.client.json(ctx, "GET", path, nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *WhatsAppService) UpdateProfile(ctx context.Context, accountID string, body *UpdateWhatsAppProfileRequest) (*WhatsAppProfile, error) {
	out := &WhatsAppProfile{}
	path := fmt.Sprintf("/accounts/%s/whatsapp/profile", url.PathEscape(accountID))
	if err := s.client.json(ctx, "PATCH", path, body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// RequestDisplayName files a name change for review. The number keeps its old
// name until the review passes, so this is a request and not a write.
func (s *WhatsAppService) RequestDisplayName(ctx context.Context, accountID, displayName string) error {
	body := map[string]string{"display_name": displayName}
	path := fmt.Sprintf("/accounts/%s/whatsapp/profile/display-name", url.PathEscape(accountID))
	return s.client.json(ctx, "POST", path, body, nil, nil)
}

func (s *WhatsAppService) SetUsername(ctx context.Context, accountID, username string) (*WhatsAppProfile, error) {
	out := &WhatsAppProfile{}
	body := map[string]string{"username": username}
	path := fmt.Sprintf("/accounts/%s/whatsapp/profile/username", url.PathEscape(accountID))
	if err := s.client.json(ctx, "PUT", path, body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ─── Templates ──────────────────────────────────────────────────────────

// WhatsAppTemplate is a message template. Status is whatever the platform
// assigned in review, passed through unchanged.
type WhatsAppTemplate struct {
	ID             string           `json:"id"`
	Name           string           `json:"name"`
	Language       string           `json:"language"`
	Category       string           `json:"category"`
	Status         string           `json:"status"`
	RejectedReason *string          `json:"rejectedReason"`
	Components     []map[string]any `json:"components"`
	QualityScore   *string          `json:"qualityScore"`
}

// CreateWhatsAppTemplateRequest files a template for review.
type CreateWhatsAppTemplateRequest struct {
	// Name takes lowercase letters, digits and underscores.
	Name       string           `json:"name"`
	Language   string           `json:"language"`
	Category   string           `json:"category"`
	Components []map[string]any `json:"components"`
	// AllowCategoryChange lets the platform re-file a template it judges differently.
	AllowCategoryChange *bool `json:"allow_category_change,omitempty"`
}

// ImportWhatsAppTemplateRequest creates a template from the platform's library.
type ImportWhatsAppTemplateRequest struct {
	LibraryTemplateName         string           `json:"library_template_name"`
	Name                        string           `json:"name"`
	Language                    string           `json:"language"`
	Category                    string           `json:"category"`
	LibraryTemplateButtonInputs []map[string]any `json:"library_template_button_inputs,omitempty"`
}

// UpdateWhatsAppTemplateRequest edits a template. The name cannot change.
type UpdateWhatsAppTemplateRequest struct {
	Category   string           `json:"category,omitempty"`
	Components []map[string]any `json:"components,omitempty"`
}

func (s *WhatsAppService) Templates(ctx context.Context, accountID string, after string) ([]WhatsAppTemplate, error) {
	q := newQuery()
	q.str("after", after)
	out := []WhatsAppTemplate{}
	path := fmt.Sprintf("/accounts/%s/whatsapp/templates", url.PathEscape(accountID))
	if err := s.client.json(ctx, "GET", path, nil, q.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// TemplateLibrary lists the pre-written templates the platform offers.
func (s *WhatsAppService) TemplateLibrary(ctx context.Context, accountID, search string) ([]map[string]any, error) {
	q := newQuery()
	q.str("search", search)
	out := []map[string]any{}
	path := fmt.Sprintf("/accounts/%s/whatsapp/templates/library", url.PathEscape(accountID))
	if err := s.client.json(ctx, "GET", path, nil, q.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *WhatsAppService) Template(ctx context.Context, accountID, templateID string) (*WhatsAppTemplate, error) {
	out := &WhatsAppTemplate{}
	path := fmt.Sprintf("/accounts/%s/whatsapp/templates/%s", url.PathEscape(accountID), url.PathEscape(templateID))
	if err := s.client.json(ctx, "GET", path, nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateTemplate files a template for review. The result carries the status the
// platform assigned, which is PENDING on a normal submission.
func (s *WhatsAppService) CreateTemplate(ctx context.Context, accountID string, body *CreateWhatsAppTemplateRequest) (*WhatsAppTemplate, error) {
	out := &WhatsAppTemplate{}
	path := fmt.Sprintf("/accounts/%s/whatsapp/templates", url.PathEscape(accountID))
	if err := s.client.json(ctx, "POST", path, body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *WhatsAppService) ImportTemplate(ctx context.Context, accountID string, body *ImportWhatsAppTemplateRequest) (*WhatsAppTemplate, error) {
	out := &WhatsAppTemplate{}
	path := fmt.Sprintf("/accounts/%s/whatsapp/templates/import", url.PathEscape(accountID))
	if err := s.client.json(ctx, "POST", path, body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *WhatsAppService) UpdateTemplate(ctx context.Context, accountID, templateID string, body *UpdateWhatsAppTemplateRequest) (*WhatsAppTemplate, error) {
	out := &WhatsAppTemplate{}
	path := fmt.Sprintf("/accounts/%s/whatsapp/templates/%s", url.PathEscape(accountID), url.PathEscape(templateID))
	if err := s.client.json(ctx, "PATCH", path, body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteTemplate needs the template name: it is what the platform deletes by.
func (s *WhatsAppService) DeleteTemplate(ctx context.Context, accountID, templateID, name string) error {
	q := newQuery()
	q.str("name", name)
	path := fmt.Sprintf("/accounts/%s/whatsapp/templates/%s", url.PathEscape(accountID), url.PathEscape(templateID))
	return s.client.json(ctx, "DELETE", path, nil, q.values(), nil)
}

// ─── Groups ─────────────────────────────────────────────────────────────

// WhatsAppGroup is a group on the business number. Participation is
// invite-only: no endpoint adds someone, so the invite link is how they join.
type WhatsAppGroup struct {
	ID               string     `json:"id"`
	Subject          string     `json:"subject"`
	Description      *string    `json:"description"`
	ParticipantCount *int       `json:"participantCount"`
	InviteLink       *string    `json:"inviteLink"`
	CreatedAt        *time.Time `json:"createdAt"`
}

// WhatsAppGroupRequest creates or updates a group.
type WhatsAppGroupRequest struct {
	Subject     string `json:"subject,omitempty"`
	Description string `json:"description,omitempty"`
}

func (s *WhatsAppService) Groups(ctx context.Context, accountID string) ([]WhatsAppGroup, error) {
	out := []WhatsAppGroup{}
	path := fmt.Sprintf("/accounts/%s/whatsapp/groups", url.PathEscape(accountID))
	if err := s.client.json(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *WhatsAppService) CreateGroup(ctx context.Context, accountID string, body *WhatsAppGroupRequest) (*WhatsAppGroup, error) {
	out := &WhatsAppGroup{}
	path := fmt.Sprintf("/accounts/%s/whatsapp/groups", url.PathEscape(accountID))
	if err := s.client.json(ctx, "POST", path, body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *WhatsAppService) Group(ctx context.Context, accountID, groupID string) (*WhatsAppGroup, error) {
	out := &WhatsAppGroup{}
	path := fmt.Sprintf("/accounts/%s/whatsapp/groups/%s", url.PathEscape(accountID), url.PathEscape(groupID))
	if err := s.client.json(ctx, "GET", path, nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *WhatsAppService) UpdateGroup(ctx context.Context, accountID, groupID string, body *WhatsAppGroupRequest) (*WhatsAppGroup, error) {
	out := &WhatsAppGroup{}
	path := fmt.Sprintf("/accounts/%s/whatsapp/groups/%s", url.PathEscape(accountID), url.PathEscape(groupID))
	if err := s.client.json(ctx, "PATCH", path, body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *WhatsAppService) DeleteGroup(ctx context.Context, accountID, groupID string) error {
	path := fmt.Sprintf("/accounts/%s/whatsapp/groups/%s", url.PathEscape(accountID), url.PathEscape(groupID))
	return s.client.json(ctx, "DELETE", path, nil, nil, nil)
}

// GroupInviteLink is the link someone joins the group with.
func (s *WhatsAppService) GroupInviteLink(ctx context.Context, accountID, groupID string) (string, error) {
	out := struct {
		InviteLink *string `json:"inviteLink"`
	}{}
	path := fmt.Sprintf("/accounts/%s/whatsapp/groups/%s/invite-link", url.PathEscape(accountID), url.PathEscape(groupID))
	if err := s.client.json(ctx, "GET", path, nil, nil, &out); err != nil {
		return "", err
	}
	if out.InviteLink == nil {
		return "", nil
	}
	return *out.InviteLink, nil
}

// ResetGroupInviteLink issues a new link and invalidates the old one.
func (s *WhatsAppService) ResetGroupInviteLink(ctx context.Context, accountID, groupID string) (string, error) {
	out := struct {
		InviteLink *string `json:"inviteLink"`
	}{}
	path := fmt.Sprintf("/accounts/%s/whatsapp/groups/%s/invite-link", url.PathEscape(accountID), url.PathEscape(groupID))
	if err := s.client.json(ctx, "POST", path, nil, nil, &out); err != nil {
		return "", err
	}
	if out.InviteLink == nil {
		return "", nil
	}
	return *out.InviteLink, nil
}

func (s *WhatsAppService) RemoveGroupParticipants(ctx context.Context, accountID, groupID string, users []string) error {
	body := map[string][]string{"users": users}
	path := fmt.Sprintf("/accounts/%s/whatsapp/groups/%s/participants", url.PathEscape(accountID), url.PathEscape(groupID))
	return s.client.json(ctx, "DELETE", path, body, nil, nil)
}

// ─── Blocking ───────────────────────────────────────────────────────────

// WhatsAppBlockResult names what the platform took and what it refused.
type WhatsAppBlockResult struct {
	Blocked   []string `json:"blocked"`
	Unblocked []string `json:"unblocked"`
	Failed    []string `json:"failed"`
}

func (s *WhatsAppService) Blocked(ctx context.Context, accountID, after string) ([]string, error) {
	q := newQuery()
	q.str("after", after)
	out := []string{}
	path := fmt.Sprintf("/accounts/%s/whatsapp/block", url.PathEscape(accountID))
	if err := s.client.json(ctx, "GET", path, nil, q.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *WhatsAppService) BlockUsers(ctx context.Context, accountID string, users []string) (*WhatsAppBlockResult, error) {
	out := &WhatsAppBlockResult{}
	body := map[string][]string{"users": users}
	path := fmt.Sprintf("/accounts/%s/whatsapp/block", url.PathEscape(accountID))
	if err := s.client.json(ctx, "POST", path, body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *WhatsAppService) UnblockUsers(ctx context.Context, accountID string, users []string) (*WhatsAppBlockResult, error) {
	out := &WhatsAppBlockResult{}
	body := map[string][]string{"users": users}
	path := fmt.Sprintf("/accounts/%s/whatsapp/block", url.PathEscape(accountID))
	if err := s.client.json(ctx, "DELETE", path, body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ─── Commerce ───────────────────────────────────────────────────────────

// WhatsAppCommerceSettings is whether the cart and catalog show on the number.
type WhatsAppCommerceSettings struct {
	CartEnabled    *bool   `json:"cartEnabled"`
	CatalogVisible *bool   `json:"catalogVisible"`
	CatalogID      *string `json:"catalogId"`
}

// UpdateWhatsAppCommerceRequest turns the cart or the catalog on or off.
type UpdateWhatsAppCommerceRequest struct {
	CartEnabled    *bool `json:"is_cart_enabled,omitempty"`
	CatalogVisible *bool `json:"is_catalog_visible,omitempty"`
}

func (s *WhatsAppService) CommerceSettings(ctx context.Context, accountID string) (*WhatsAppCommerceSettings, error) {
	out := &WhatsAppCommerceSettings{}
	path := fmt.Sprintf("/accounts/%s/whatsapp/commerce", url.PathEscape(accountID))
	if err := s.client.json(ctx, "GET", path, nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *WhatsAppService) UpdateCommerceSettings(ctx context.Context, accountID string, body *UpdateWhatsAppCommerceRequest) (*WhatsAppCommerceSettings, error) {
	out := &WhatsAppCommerceSettings{}
	path := fmt.Sprintf("/accounts/%s/whatsapp/commerce", url.PathEscape(accountID))
	if err := s.client.json(ctx, "PATCH", path, body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// LinkCatalog points the number at a catalog the customer already owns.
func (s *WhatsAppService) LinkCatalog(ctx context.Context, accountID, catalogID string) (*WhatsAppCommerceSettings, error) {
	out := &WhatsAppCommerceSettings{}
	body := map[string]string{"catalog_id": catalogID}
	path := fmt.Sprintf("/accounts/%s/whatsapp/commerce/catalog", url.PathEscape(accountID))
	if err := s.client.json(ctx, "POST", path, body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ─── Flows ──────────────────────────────────────────────────────────────

// WhatsAppFlowValidationError is one problem the platform found in a flow.
type WhatsAppFlowValidationError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// WhatsAppFlow is an in-chat form. The platform validates it and owns its status.
type WhatsAppFlow struct {
	ID               string                        `json:"id"`
	Name             string                        `json:"name"`
	Status           string                        `json:"status"`
	Categories       []string                      `json:"categories"`
	ValidationErrors []WhatsAppFlowValidationError `json:"validationErrors"`
	EndpointURI      *string                       `json:"endpointUri"`
	JSONVersion      *string                       `json:"jsonVersion"`
	PreviewURL       *string                       `json:"previewUrl"`
	PreviewExpiresAt *time.Time                    `json:"previewExpiresAt"`
}

// CreateWhatsAppFlowRequest creates a draft flow.
type CreateWhatsAppFlowRequest struct {
	Name       string   `json:"name"`
	Categories []string `json:"categories"`
	// EndpointURI is where the platform calls back for a flow that reads live data.
	EndpointURI string `json:"endpoint_uri,omitempty"`
	CloneFlowID string `json:"clone_flow_id,omitempty"`
}

// UpdateWhatsAppFlowRequest changes a flow's metadata, not its screens.
type UpdateWhatsAppFlowRequest struct {
	Name        string   `json:"name,omitempty"`
	Categories  []string `json:"categories,omitempty"`
	EndpointURI string   `json:"endpoint_uri,omitempty"`
}

// WhatsAppFlowJSONResult is the platform's verdict on an uploaded definition.
type WhatsAppFlowJSONResult struct {
	Success          bool                          `json:"success"`
	ValidationErrors []WhatsAppFlowValidationError `json:"validationErrors"`
}

// WhatsAppFlowResponse is what one person submitted through a flow.
type WhatsAppFlowResponse struct {
	MessageID   string         `json:"messageId"`
	WaID        *string        `json:"waId"`
	FlowToken   *string        `json:"flowToken"`
	Answers     map[string]any `json:"answers"`
	RespondedAt *time.Time     `json:"respondedAt"`
}

// WhatsAppEncryptionKeyStatus reports whether a key is registered. The key
// itself never comes back.
type WhatsAppEncryptionKeyStatus struct {
	HasKey          bool    `json:"hasKey"`
	SignatureStatus *string `json:"signatureStatus"`
}

func (s *WhatsAppService) Flows(ctx context.Context, accountID string) ([]WhatsAppFlow, error) {
	out := []WhatsAppFlow{}
	path := fmt.Sprintf("/accounts/%s/whatsapp/flows", url.PathEscape(accountID))
	if err := s.client.json(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *WhatsAppService) Flow(ctx context.Context, accountID, flowID string) (*WhatsAppFlow, error) {
	out := &WhatsAppFlow{}
	path := fmt.Sprintf("/accounts/%s/whatsapp/flows/%s", url.PathEscape(accountID), url.PathEscape(flowID))
	if err := s.client.json(ctx, "GET", path, nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *WhatsAppService) CreateFlow(ctx context.Context, accountID string, body *CreateWhatsAppFlowRequest) (*WhatsAppFlow, error) {
	out := &WhatsAppFlow{}
	path := fmt.Sprintf("/accounts/%s/whatsapp/flows", url.PathEscape(accountID))
	if err := s.client.json(ctx, "POST", path, body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *WhatsAppService) UpdateFlow(ctx context.Context, accountID, flowID string, body *UpdateWhatsAppFlowRequest) (*WhatsAppFlow, error) {
	out := &WhatsAppFlow{}
	path := fmt.Sprintf("/accounts/%s/whatsapp/flows/%s", url.PathEscape(accountID), url.PathEscape(flowID))
	if err := s.client.json(ctx, "PATCH", path, body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteFlow removes a draft. A published flow is deprecated instead.
func (s *WhatsAppService) DeleteFlow(ctx context.Context, accountID, flowID string) error {
	path := fmt.Sprintf("/accounts/%s/whatsapp/flows/%s", url.PathEscape(accountID), url.PathEscape(flowID))
	return s.client.json(ctx, "DELETE", path, nil, nil, nil)
}

// UploadFlowJSON replaces the flow's screens. The platform answers with its
// validation errors rather than refusing, so they come back as data.
func (s *WhatsAppService) UploadFlowJSON(ctx context.Context, accountID, flowID string, flowJSON map[string]any) (*WhatsAppFlowJSONResult, error) {
	out := &WhatsAppFlowJSONResult{}
	body := map[string]any{"flow_json": flowJSON}
	path := fmt.Sprintf("/accounts/%s/whatsapp/flows/%s/json", url.PathEscape(accountID), url.PathEscape(flowID))
	if err := s.client.json(ctx, "PUT", path, body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *WhatsAppService) PublishFlow(ctx context.Context, accountID, flowID string) (*WhatsAppFlow, error) {
	out := &WhatsAppFlow{}
	path := fmt.Sprintf("/accounts/%s/whatsapp/flows/%s/publish", url.PathEscape(accountID), url.PathEscape(flowID))
	if err := s.client.json(ctx, "POST", path, nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *WhatsAppService) DeprecateFlow(ctx context.Context, accountID, flowID string) (*WhatsAppFlow, error) {
	out := &WhatsAppFlow{}
	path := fmt.Sprintf("/accounts/%s/whatsapp/flows/%s/deprecate", url.PathEscape(accountID), url.PathEscape(flowID))
	if err := s.client.json(ctx, "POST", path, nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *WhatsAppService) FlowResponses(ctx context.Context, accountID string) ([]WhatsAppFlowResponse, error) {
	out := []WhatsAppFlowResponse{}
	path := fmt.Sprintf("/accounts/%s/whatsapp/flows/responses", url.PathEscape(accountID))
	if err := s.client.json(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *WhatsAppService) EncryptionKeyStatus(ctx context.Context, accountID string) (*WhatsAppEncryptionKeyStatus, error) {
	out := &WhatsAppEncryptionKeyStatus{}
	path := fmt.Sprintf("/accounts/%s/whatsapp/flows/encryption-key", url.PathEscape(accountID))
	if err := s.client.json(ctx, "GET", path, nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// SetEncryptionKey registers the public half of the key the platform encrypts a
// flow endpoint's payloads with. The private half stays with the customer.
func (s *WhatsAppService) SetEncryptionKey(ctx context.Context, accountID, businessPublicKey string) (*WhatsAppEncryptionKeyStatus, error) {
	out := &WhatsAppEncryptionKeyStatus{}
	body := map[string]string{"business_public_key": businessPublicKey}
	path := fmt.Sprintf("/accounts/%s/whatsapp/flows/encryption-key", url.PathEscape(accountID))
	if err := s.client.json(ctx, "PUT", path, body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ─── Account state and sandbox ──────────────────────────────────────────

// AccountEvents reads the account review state and the number's standing.
func (s *WhatsAppService) AccountEvents(ctx context.Context, accountID string) (map[string]any, error) {
	out := map[string]any{}
	path := fmt.Sprintf("/accounts/%s/whatsapp/events", url.PathEscape(accountID))
	if err := s.client.json(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// WhatsAppSandboxSession is a sandbox invitation. Only the last four digits of
// the tester's number travel; the number itself is never stored.
type WhatsAppSandboxSession struct {
	ID               string     `json:"id"`
	Status           string     `json:"status"`
	PhoneNumberLast4 string     `json:"phoneNumberLast4"`
	InvitedAt        time.Time  `json:"invitedAt"`
	ActivatedAt      *time.Time `json:"activatedAt"`
	ExpiresAt        time.Time  `json:"expiresAt"`
}

func (s *WhatsAppService) SandboxSessions(ctx context.Context, workspaceID string) ([]WhatsAppSandboxSession, error) {
	q := newQuery()
	q.str("workspaceId", workspaceID)
	out := []WhatsAppSandboxSession{}
	if err := s.client.json(ctx, "GET", "/whatsapp/sandbox/sessions", nil, q.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateSandboxSession invites one tester to the platform-owned test number.
// Inviting sends a template, so it needs the publish scope.
func (s *WhatsAppService) CreateSandboxSession(ctx context.Context, workspaceID, phoneNumber string) (*WhatsAppSandboxSession, error) {
	out := &WhatsAppSandboxSession{}
	body := map[string]string{"workspaceId": workspaceID, "phoneNumber": phoneNumber}
	if err := s.client.json(ctx, "POST", "/whatsapp/sandbox/sessions", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}
