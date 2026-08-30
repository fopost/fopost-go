package fopost

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Error is returned for every non-2xx response. The API answers with an
// {"error": "<code>", "message": "<explanation>"} envelope, which maps onto
// Code and Message.
type Error struct {
	// Status is the HTTP status code.
	Status int
	// Code is the machine-readable error code, e.g. "subscription_required".
	Code string
	// Message is the human-readable explanation.
	Message string
	// Body is the raw response body, for fields this type does not model.
	Body []byte
	// RetryAfter is set on 429 when the API asked for a specific wait.
	RetryAfter time.Duration
	// RateLimit carries the X-RateLimit-* headers that came with the response.
	RateLimit RateLimit
}

func (e *Error) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("fopost: %d %s: %s", e.Status, e.Code, e.Message)
	}
	return fmt.Sprintf("fopost: %d: %s", e.Status, e.Message)
}

// UpgradeURL is the path a 402 suggests sending the user to, when it carries one.
func (e *Error) UpgradeURL() string {
	var body struct {
		UpgradeURL string `json:"upgrade_url"`
	}
	if json.Unmarshal(e.Body, &body) != nil {
		return ""
	}
	return body.UpgradeURL
}

// Field decodes an extra field some errors carry alongside error and message.
func (e *Error) Field(name string, out any) error {
	var body map[string]json.RawMessage
	if err := json.Unmarshal(e.Body, &body); err != nil {
		return err
	}
	raw, ok := body[name]
	if !ok {
		return fmt.Errorf("fopost: error body has no field %q", name)
	}
	return json.Unmarshal(raw, out)
}

// RateLimit is the per-key, per-minute budget reported on every response.
type RateLimit struct {
	Limit     int
	Remaining int
	Reset     time.Time
}

// APIError unwraps err into *Error, reporting whether it was one.
func APIError(err error) (*Error, bool) {
	var apiErr *Error
	ok := errors.As(err, &apiErr)
	return apiErr, ok
}

// StatusOf is the HTTP status behind err, or 0 if it is not an API error.
func StatusOf(err error) int {
	if apiErr, ok := APIError(err); ok {
		return apiErr.Status
	}
	return 0
}

// CodeOf is the API error code behind err, or "" if it is not an API error.
func CodeOf(err error) string {
	if apiErr, ok := APIError(err); ok {
		return apiErr.Code
	}
	return ""
}

// IsUnauthorized reports a 401 — missing, invalid, or expired API key.
func IsUnauthorized(err error) bool { return StatusOf(err) == http.StatusUnauthorized }

// IsPaymentRequired reports a 402 — no active subscription, or AI credits exhausted.
func IsPaymentRequired(err error) bool { return StatusOf(err) == http.StatusPaymentRequired }

// IsForbidden reports a 403 — the key is valid but lacks the scope or workspace access.
func IsForbidden(err error) bool { return StatusOf(err) == http.StatusForbidden }

// IsNotFound reports a 404 — no such resource, or it is outside the key's reach.
func IsNotFound(err error) bool { return StatusOf(err) == http.StatusNotFound }

// IsConflict reports a 409 — the resource is in a state that forbids the change.
func IsConflict(err error) bool { return StatusOf(err) == http.StatusConflict }

// IsRateLimited reports a 429. The wait the API asked for is in Error.RetryAfter.
func IsRateLimited(err error) bool { return StatusOf(err) == http.StatusTooManyRequests }
