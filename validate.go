package fopost

import (
	"context"
	"net/url"
)

// ValidateService checks content against platform rules without creating a
// post. Nothing is stored server-side; every method needs the posts scope.
type ValidateService struct{ client *Client }

// ValidateMediaItem is one attachment described for ValidatePost.
type ValidateMediaItem struct {
	URL      string `json:"url"`
	MimeType string `json:"mime_type"`
	Size     *int64 `json:"size,omitempty"`
}

// ValidatePostRequest is the body of Post. Platforms is required.
type ValidatePostRequest struct {
	Content   string              `json:"content,omitempty"`
	Media     []ValidateMediaItem `json:"media,omitempty"`
	Platforms []string            `json:"platforms"`
}

// ValidatePostResult is the per-platform readiness of a draft.
type ValidatePostResult struct {
	Ready     bool                   `json:"ready"`
	Platforms []ValidatePostPlatform `json:"platforms"`
}

// ValidatePostPlatform is one platform's verdict. Issues block publishing;
// Signals are advisory. Score is nil when the platform reports none.
type ValidatePostPlatform struct {
	Platform string          `json:"platform"`
	Ready    bool            `json:"ready"`
	Issues   []string        `json:"issues"`
	Score    *float64        `json:"score"`
	Signals  []ContentSignal `json:"signals"`
}

// Post checks text and media against each platform's rules.
func (s *ValidateService) Post(ctx context.Context, body *ValidatePostRequest) (*ValidatePostResult, error) {
	out := &ValidatePostResult{}
	if err := s.client.json(ctx, "POST", "/validate/post", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ValidateLengthRequest is the body of Length. Both fields are required.
type ValidateLengthRequest struct {
	Text      string   `json:"text"`
	Platforms []string `json:"platforms"`
}

// ValidateLengthResult is the per-platform length check.
type ValidateLengthResult struct {
	OK        bool                     `json:"ok"`
	Platforms []ValidateLengthPlatform `json:"platforms"`
}

// ValidateLengthPlatform is one platform's count. Limit is nil when the
// platform has no text limit; Unit is "chars" or "bytes".
type ValidateLengthPlatform struct {
	Platform string          `json:"platform"`
	Length   int             `json:"length"`
	Limit    *int            `json:"limit"`
	Unit     string          `json:"unit"`
	OK       bool            `json:"ok"`
	Signals  []ContentSignal `json:"signals"`
}

// Length counts text the way each platform does and compares it to the limit.
func (s *ValidateService) Length(ctx context.Context, body *ValidateLengthRequest) (*ValidateLengthResult, error) {
	out := &ValidateLengthResult{}
	if err := s.client.json(ctx, "POST", "/validate/length", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ValidateMediaResult is the verdict on one public file. MimeType and Type
// are set only when OK; Type is "image", "video", "audio" or "document".
type ValidateMediaResult struct {
	OK       bool     `json:"ok"`
	Issues   []string `json:"issues"`
	Name     string   `json:"name"`
	Size     int64    `json:"size"`
	MimeType string   `json:"mime_type"`
	Type     string   `json:"type"`
}

// Media fetches a public http(s) URL and checks the file. A file that fails a
// check still answers 200 with OK false; an unreachable URL is an error.
func (s *ValidateService) Media(ctx context.Context, fileURL string) (*ValidateMediaResult, error) {
	out := &ValidateMediaResult{}
	body := map[string]string{"url": fileURL}
	if err := s.client.json(ctx, "POST", "/validate/media", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// SubredditCheck is the verdict on one subreddit. OK is true when it exists and
// takes a post from the named account.
type SubredditCheck struct {
	Subreddit    string `json:"subreddit"`
	Exists       bool   `json:"exists"`
	CanPost      bool   `json:"can_post"`
	Over18       bool   `json:"over_18"`
	FlairEnabled bool   `json:"flair_enabled"`
	OK           bool   `json:"ok"`
}

// Subreddit reports whether a subreddit exists and takes a post from a
// connected Reddit account. The check runs with that account's own token, so
// accountID is required. A private, banned or missing subreddit answers 200
// with Exists false rather than an error.
func (s *ValidateService) Subreddit(ctx context.Context, accountID, name string) (*SubredditCheck, error) {
	out := &SubredditCheck{}
	query := url.Values{"account_id": {accountID}, "name": {name}}
	if err := s.client.json(ctx, "GET", "/validate/subreddit", nil, query, out); err != nil {
		return nil, err
	}
	return out, nil
}
