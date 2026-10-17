package fopost

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
)

// GoogleBusinessService manages a connected Google Business Profile location:
// the profile itself, attributes, food menus, services, photos, place action
// links, verification and performance.
//
// Google grants Business Profile API access per project. Until that grant
// lands on a deployment every call here returns a 503 "configuration_error".
//
// Responses relay Google's own shape, field for field, so they come back as
// GoogleBusinessPayload rather than structs we would have to keep chasing.
type GoogleBusinessService struct {
	client *Client
}

// GoogleBusinessPayload is one relayed Business Profile response.
type GoogleBusinessPayload map[string]any

// DefaultGoogleBusinessDailyMetrics is the set fetched when a caller names none.
var DefaultGoogleBusinessDailyMetrics = []string{
	"BUSINESS_IMPRESSIONS_DESKTOP_MAPS",
	"BUSINESS_IMPRESSIONS_DESKTOP_SEARCH",
	"BUSINESS_IMPRESSIONS_MOBILE_MAPS",
	"BUSINESS_IMPRESSIONS_MOBILE_SEARCH",
	"CALL_CLICKS",
	"WEBSITE_CLICKS",
	"BUSINESS_DIRECTION_REQUESTS",
}

func gbpPath(id, suffix string) string {
	return "/accounts/" + url.PathEscape(id) + "/gbp" + suffix
}

// GoogleBusinessHoursPeriod is one opening stretch on the location's own clock.
type GoogleBusinessHoursPeriod struct {
	OpenDay   string `json:"open_day"`
	OpenTime  string `json:"open_time"`
	CloseDay  string `json:"close_day"`
	CloseTime string `json:"close_time"`
}

// UpdateGoogleBusinessLocationRequest patches the profile. A nil field is left
// alone; a pointer to the empty string clears the value.
type UpdateGoogleBusinessLocationRequest struct {
	Title            *string
	Description      *string
	WebsiteURI       *string
	PrimaryPhone     *string
	AdditionalPhones []string
	StoreCode        *string
	RegularHours     []GoogleBusinessHoursPeriod
}

// MarshalJSON omits nil fields and sends "" as null.
func (r UpdateGoogleBusinessLocationRequest) MarshalJSON() ([]byte, error) {
	body := map[string]any{}
	for key, value := range map[string]*string{
		"title":         r.Title,
		"description":   r.Description,
		"website_uri":   r.WebsiteURI,
		"primary_phone": r.PrimaryPhone,
		"store_code":    r.StoreCode,
	} {
		if value == nil {
			continue
		}
		if *value == "" && key != "title" {
			body[key] = nil
		} else {
			body[key] = *value
		}
	}
	if r.AdditionalPhones != nil {
		body["additional_phones"] = r.AdditionalPhones
	}
	if r.RegularHours != nil {
		body["regular_hours"] = r.RegularHours
	}
	return json.Marshal(body)
}

// GetLocation returns the connected location, in the Business Information shape.
func (s *GoogleBusinessService) GetLocation(ctx context.Context, id string) (GoogleBusinessPayload, error) {
	return s.read(ctx, "GET", gbpPath(id, "/location"), nil, nil)
}

// UpdateLocation patches only the fields the request carries.
func (s *GoogleBusinessService) UpdateLocation(ctx context.Context, id string, body *UpdateGoogleBusinessLocationRequest) (GoogleBusinessPayload, error) {
	if body == nil {
		body = &UpdateGoogleBusinessLocationRequest{}
	}
	return s.read(ctx, "PATCH", gbpPath(id, "/location"), body, nil)
}

// GoogleBusinessAttributesOptions narrows an attribute read.
type GoogleBusinessAttributesOptions struct {
	// Available lists what Google offers the location instead of what is set.
	Available    bool
	CategoryName string
	RegionCode   string
	LanguageCode string
}

// GetAttributes returns the attribute values set on the location, or, with
// Available, the attributes Google offers for its category and region.
func (s *GoogleBusinessService) GetAttributes(ctx context.Context, id string, opts *GoogleBusinessAttributesOptions) (GoogleBusinessPayload, error) {
	query := url.Values{}
	if opts != nil {
		if opts.Available {
			query.Set("available", "true")
		}
		setIfNotEmpty(query, "category_name", opts.CategoryName)
		setIfNotEmpty(query, "region_code", opts.RegionCode)
		setIfNotEmpty(query, "language_code", opts.LanguageCode)
	}
	return s.read(ctx, "GET", gbpPath(id, "/attributes"), nil, query)
}

// UpdateAttributes changes only the named attributes; the rest are left alone.
func (s *GoogleBusinessService) UpdateAttributes(ctx context.Context, id string, attributes []map[string]any) (GoogleBusinessPayload, error) {
	if attributes == nil {
		attributes = []map[string]any{}
	}
	return s.read(ctx, "PATCH", gbpPath(id, "/attributes"), map[string]any{"attributes": attributes}, nil)
}

// GetMenus returns the location's food menus.
func (s *GoogleBusinessService) GetMenus(ctx context.Context, id string) (GoogleBusinessPayload, error) {
	return s.read(ctx, "GET", gbpPath(id, "/menus"), nil, nil)
}

// ReplaceMenus replaces the whole menu set; Google has no per-section patch.
func (s *GoogleBusinessService) ReplaceMenus(ctx context.Context, id string, menus []map[string]any) (GoogleBusinessPayload, error) {
	if menus == nil {
		menus = []map[string]any{}
	}
	return s.read(ctx, "PUT", gbpPath(id, "/menus"), map[string]any{"menus": menus}, nil)
}

// GetServices returns the location's service list.
func (s *GoogleBusinessService) GetServices(ctx context.Context, id string) (GoogleBusinessPayload, error) {
	return s.read(ctx, "GET", gbpPath(id, "/services"), nil, nil)
}

// ReplaceServices replaces the whole service list.
func (s *GoogleBusinessService) ReplaceServices(ctx context.Context, id string, serviceItems []map[string]any) (GoogleBusinessPayload, error) {
	if serviceItems == nil {
		serviceItems = []map[string]any{}
	}
	return s.read(ctx, "PUT", gbpPath(id, "/services"), map[string]any{"service_items": serviceItems}, nil)
}

// ListMedia returns the photos on the location.
func (s *GoogleBusinessService) ListMedia(ctx context.Context, id string, pageSize int, pageToken string) (GoogleBusinessPayload, error) {
	query := url.Values{}
	if pageSize > 0 {
		query.Set("page_size", strconv.Itoa(pageSize))
	}
	setIfNotEmpty(query, "page_token", pageToken)
	return s.read(ctx, "GET", gbpPath(id, "/media"), nil, query)
}

// AddGoogleBusinessMediaRequest adds a photo from the media library. The asset
// has to be in a workspace the caller can reach, and JPEG or PNG.
type AddGoogleBusinessMediaRequest struct {
	MediaID     string `json:"media_id"`
	Category    string `json:"category,omitempty"`
	Description string `json:"description,omitempty"`
}

// AddMedia adds a library photo to the location.
func (s *GoogleBusinessService) AddMedia(ctx context.Context, id string, body *AddGoogleBusinessMediaRequest) (GoogleBusinessPayload, error) {
	if body == nil {
		body = &AddGoogleBusinessMediaRequest{}
	}
	return s.read(ctx, "POST", gbpPath(id, "/media"), body, nil)
}

// DeleteMedia removes a photo by the media key Google returned.
func (s *GoogleBusinessService) DeleteMedia(ctx context.Context, id, mediaKey string) (GoogleBusinessPayload, error) {
	return s.read(ctx, "DELETE", gbpPath(id, "/media/"+url.PathEscape(mediaKey)), nil, nil)
}

// ListPlaceActions returns the Book, Order and Reserve links on the listing.
func (s *GoogleBusinessService) ListPlaceActions(ctx context.Context, id string) (GoogleBusinessPayload, error) {
	return s.read(ctx, "GET", gbpPath(id, "/place-actions"), nil, nil)
}

// CreateGoogleBusinessPlaceActionRequest adds one action link.
type CreateGoogleBusinessPlaceActionRequest struct {
	URI             string `json:"uri"`
	PlaceActionType string `json:"place_action_type"`
	IsPreferred     *bool  `json:"is_preferred,omitempty"`
}

// CreatePlaceAction adds an action link to the listing.
func (s *GoogleBusinessService) CreatePlaceAction(ctx context.Context, id string, body *CreateGoogleBusinessPlaceActionRequest) (GoogleBusinessPayload, error) {
	if body == nil {
		body = &CreateGoogleBusinessPlaceActionRequest{}
	}
	return s.read(ctx, "POST", gbpPath(id, "/place-actions"), body, nil)
}

// UpdateGoogleBusinessPlaceActionRequest patches one action link; a nil field
// is left alone.
type UpdateGoogleBusinessPlaceActionRequest struct {
	URI         *string `json:"uri,omitempty"`
	IsPreferred *bool   `json:"is_preferred,omitempty"`
}

// UpdatePlaceAction patches one action link.
func (s *GoogleBusinessService) UpdatePlaceAction(ctx context.Context, id, linkID string, body *UpdateGoogleBusinessPlaceActionRequest) (GoogleBusinessPayload, error) {
	if body == nil {
		body = &UpdateGoogleBusinessPlaceActionRequest{}
	}
	return s.read(ctx, "PATCH", gbpPath(id, "/place-actions/"+url.PathEscape(linkID)), body, nil)
}

// DeletePlaceAction removes one action link.
func (s *GoogleBusinessService) DeletePlaceAction(ctx context.Context, id, linkID string) (GoogleBusinessPayload, error) {
	return s.read(ctx, "DELETE", gbpPath(id, "/place-actions/"+url.PathEscape(linkID)), nil, nil)
}

// GetVerificationOptions returns the ways Google will let this location be verified.
func (s *GoogleBusinessService) GetVerificationOptions(ctx context.Context, id, languageCode string) (GoogleBusinessPayload, error) {
	query := url.Values{}
	setIfNotEmpty(query, "language_code", languageCode)
	return s.read(ctx, "GET", gbpPath(id, "/verification"), nil, query)
}

// StartGoogleBusinessVerificationRequest begins a verification. The response
// names the pending verification to complete with the PIN.
type StartGoogleBusinessVerificationRequest struct {
	Method            string `json:"method"`
	LanguageCode      string `json:"language_code,omitempty"`
	PhoneNumber       string `json:"phone_number,omitempty"`
	EmailAddress      string `json:"email_address,omitempty"`
	MailerContactName string `json:"mailer_contact_name,omitempty"`
}

// StartVerification begins verifying the location.
func (s *GoogleBusinessService) StartVerification(ctx context.Context, id string, body *StartGoogleBusinessVerificationRequest) (GoogleBusinessPayload, error) {
	if body == nil {
		body = &StartGoogleBusinessVerificationRequest{}
	}
	return s.read(ctx, "POST", gbpPath(id, "/verification/start"), body, nil)
}

// CompleteVerification finishes a pending verification with the PIN Google sent.
func (s *GoogleBusinessService) CompleteVerification(ctx context.Context, id, verificationName, pin string) (GoogleBusinessPayload, error) {
	body := map[string]any{"verification_name": verificationName, "pin": pin}
	return s.read(ctx, "POST", gbpPath(id, "/verification/complete"), body, nil)
}

// GetPerformance returns daily impressions, calls, direction requests and
// clicks for the range. Nil dailyMetrics leaves the API's default set.
func (s *GoogleBusinessService) GetPerformance(ctx context.Context, id, startDate, endDate string, dailyMetrics []string) (GoogleBusinessPayload, error) {
	query := url.Values{"start_date": {startDate}, "end_date": {endDate}}
	for _, metric := range dailyMetrics {
		query.Add("daily_metrics", metric)
	}
	return s.read(ctx, "GET", gbpPath(id, "/performance"), nil, query)
}

// GetSearchKeywords returns the search terms people used to find the listing,
// by month.
func (s *GoogleBusinessService) GetSearchKeywords(ctx context.Context, id, startDate, endDate, pageToken string) (GoogleBusinessPayload, error) {
	query := url.Values{"keywords": {"true"}, "start_date": {startDate}, "end_date": {endDate}}
	setIfNotEmpty(query, "page_token", pageToken)
	return s.read(ctx, "GET", gbpPath(id, "/performance"), nil, query)
}

// Assign hands the location to another workspace the caller owns. The
// connection and every row keyed to it move in one transaction.
func (s *GoogleBusinessService) Assign(ctx context.Context, id, workspaceID string) (*MovedAccount, error) {
	out := &MovedAccount{}
	body := map[string]any{"workspace_id": workspaceID}
	if err := s.client.json(ctx, "POST", gbpPath(id, "/assign"), body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *GoogleBusinessService) read(ctx context.Context, method, path string, body any, query url.Values) (GoogleBusinessPayload, error) {
	out := GoogleBusinessPayload{}
	if err := s.client.json(ctx, method, path, body, query, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func setIfNotEmpty(query url.Values, key, value string) {
	if value != "" {
		query.Set(key, value)
	}
}
