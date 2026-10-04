package fopost

import (
	"context"
	"net/url"
	"strconv"
)

// Per-network extras under /accounts/{id}/<platform>/…, all on the accounts scope.

// PinterestBoard is a board a Pin can land on. Pass ID as the board_id platform
// setting to pin to it.
type PinterestBoard struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Privacy     *string `json:"privacy"`
	Description *string `json:"description"`
	// Image is the board cover.
	Image *string `json:"image"`
}

// CreatePinterestBoardRequest creates a board on the connected account.
type CreatePinterestBoardRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	// Privacy is PUBLIC, PROTECTED or SECRET; empty means PUBLIC.
	Privacy string `json:"privacy,omitempty"`
}

// YouTubePlaylist is a playlist on the connected channel.
type YouTubePlaylist struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Description  *string `json:"description"`
	Privacy      *string `json:"privacy"`
	ItemCount    *int    `json:"item_count"`
	ThumbnailURL *string `json:"thumbnail_url"`
	// IsDefault marks the playlist a new video joins when the post picks none.
	IsDefault bool `json:"is_default"`
}

// CreateYouTubePlaylistRequest creates a playlist on the channel.
type CreateYouTubePlaylistRequest struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	// Privacy is public, unlisted or private; empty means private.
	Privacy string `json:"privacy,omitempty"`
}

// YouTubeCaptionTrack is a caption track on one of the channel's videos.
type YouTubeCaptionTrack struct {
	ID string `json:"id"`
	// Language is a BCP-47 tag.
	Language     string  `json:"language"`
	Name         string  `json:"name"`
	TrackKind    *string `json:"track_kind"`
	IsDraft      bool    `json:"is_draft"`
	IsAutoSynced bool    `json:"is_auto_synced"`
	LastUpdated  *string `json:"last_updated"`
}

// UploadYouTubeCaptionsRequest uploads a caption track. Body is the subtitle
// file itself; YouTube reads SRT and WebVTT and works out which from the bytes.
type UploadYouTubeCaptionsRequest struct {
	Language string `json:"language"`
	Body     string `json:"body"`
	Name     string `json:"name,omitempty"`
	IsDraft  bool   `json:"is_draft,omitempty"`
}

// YouTubeTranscript is one caption track read back as text, in SRT.
type YouTubeTranscript struct {
	CaptionID  string `json:"caption_id"`
	Transcript string `json:"transcript"`
}

// BlueskyLanguages is the default post languages for a connection.
type BlueskyLanguages struct {
	// Languages holds up to three BCP-47 tags.
	Languages []string `json:"languages"`
}

// TikTokCreatorInfo reports the switches TikTok enforces at publish time. They
// are set on the TikTok account itself, not in FoPost.
type TikTokCreatorInfo struct {
	Username  *string `json:"username"`
	Nickname  *string `json:"nickname"`
	AvatarURL *string `json:"avatar_url"`
	// PrivacyLevelOptions are the levels this creator may publish at right now.
	PrivacyLevelOptions     []string `json:"privacy_level_options"`
	CommentDisabled         bool     `json:"comment_disabled"`
	DuetDisabled            bool     `json:"duet_disabled"`
	StitchDisabled          bool     `json:"stitch_disabled"`
	MaxVideoPostDurationSec *int     `json:"max_video_post_duration_sec"`
}

// TikTokMusic is a track from TikTok's Commercial Music Library. Pass ID as the
// music_id platform setting to attach it.
type TikTokMusic struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Author      *string `json:"author"`
	DurationSec *int    `json:"duration_sec"`
	CoverURL    *string `json:"cover_url"`
	PreviewURL  *string `json:"preview_url"`
}

// TikTokPlace is a place a post can be tagged with. Pass ID as the location_id
// platform setting.
type TikTokPlace struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Address *string `json:"address"`
	City    *string `json:"city"`
	Country *string `json:"country"`
}

// TikTokSearchOptions narrows a music or place search. Query is required.
type TikTokSearchOptions struct {
	Query string
	// Limit is 1 to 50; the API defaults to 20 when it is zero.
	Limit int
}

// TikTokVideoSource is one of the account's own videos, resolved from a share
// link. TikTok serves no raw media file, so DownloadURL is the share address,
// which is what a repurpose run reads.
type TikTokVideoSource struct {
	VideoID       string  `json:"video_id"`
	Title         *string `json:"title"`
	Description   *string `json:"description"`
	DurationSec   *int    `json:"duration_sec"`
	CoverImageURL *string `json:"cover_image_url"`
	ShareURL      *string `json:"share_url"`
	EmbedLink     *string `json:"embed_link"`
	DownloadURL   *string `json:"download_url"`
}

// InstagramAudio is a track a Reel can carry. Pass ID as the audio_id platform
// setting to attach it.
type InstagramAudio struct {
	ID              string  `json:"id"`
	Title           *string `json:"title"`
	Artist          *string `json:"artist"`
	DurationMs      *int    `json:"duration_ms"`
	AudioType       *string `json:"audio_type"`
	CoverArtworkURL *string `json:"cover_artwork_url"`
	PreviewURL      *string `json:"preview_url"`
	Username        *string `json:"username"`
	IsAdsEligible   *bool   `json:"is_ads_eligible"`
}

// InstagramAudioSearchOptions narrows the audio search. An empty Query asks
// Instagram for what is trending.
type InstagramAudioSearchOptions struct {
	Query string
	// AudioType is music (the default) or original_sound.
	AudioType string
}

// InstagramPublishingLimit reports what this account has published in the
// rolling window and how much is left before Instagram refuses the next post.
type InstagramPublishingLimit struct {
	QuotaUsage       int  `json:"quota_usage"`
	QuotaTotal       *int `json:"quota_total"`
	QuotaDurationSec *int `json:"quota_duration_sec"`
	Remaining        *int `json:"remaining"`
}

// InstagramStory is a story still inside its 24 hours.
type InstagramStory struct {
	ID               string  `json:"id"`
	MediaType        *string `json:"media_type"`
	MediaProductType *string `json:"media_product_type"`
	Permalink        *string `json:"permalink"`
	MediaURL         *string `json:"media_url"`
	ThumbnailURL     *string `json:"thumbnail_url"`
	Caption          *string `json:"caption"`
	Timestamp        *string `json:"timestamp"`
	// Insights is present only when asked for, and empty for a story too young
	// or too small for Instagram to report on.
	Insights map[string]int `json:"insights"`
}

// InstagramStoryInsights is the insight set for one story.
type InstagramStoryInsights struct {
	StoryID  string         `json:"story_id"`
	Insights map[string]int `json:"insights"`
}

// LinkedInMention is an entity a post can mention. Annotation is what the post
// text carries for LinkedIn to render a link.
type LinkedInMention struct {
	URN        string  `json:"urn"`
	Name       string  `json:"name"`
	VanityName *string `json:"vanity_name"`
	LogoURL    *string `json:"logo_url"`
	Type       string  `json:"type"`
	Annotation string  `json:"annotation"`
}

// ListPinterestBoards returns the boards this connection can pin to.
func (s *AccountsService) ListPinterestBoards(ctx context.Context, id string) ([]PinterestBoard, error) {
	var out []PinterestBoard
	if err := s.client.json(ctx, "GET", "/accounts/"+url.PathEscape(id)+"/pinterest/boards", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreatePinterestBoard creates a board on the connected account.
func (s *AccountsService) CreatePinterestBoard(ctx context.Context, id string, body *CreatePinterestBoardRequest) (*PinterestBoard, error) {
	out := &PinterestBoard{}
	if err := s.client.json(ctx, "POST", "/accounts/"+url.PathEscape(id)+"/pinterest/boards", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListYouTubePlaylists returns the channel's playlists with the default marked.
func (s *AccountsService) ListYouTubePlaylists(ctx context.Context, id string) ([]YouTubePlaylist, error) {
	var out []YouTubePlaylist
	if err := s.client.json(ctx, "GET", "/accounts/"+url.PathEscape(id)+"/youtube/playlists", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateYouTubePlaylist creates a playlist on the channel.
func (s *AccountsService) CreateYouTubePlaylist(ctx context.Context, id string, body *CreateYouTubePlaylistRequest) (*YouTubePlaylist, error) {
	out := &YouTubePlaylist{}
	if err := s.client.json(ctx, "POST", "/accounts/"+url.PathEscape(id)+"/youtube/playlists", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// SetDefaultYouTubePlaylist stores the playlist a new video joins when the post
// picks none. An empty playlistID clears it.
func (s *AccountsService) SetDefaultYouTubePlaylist(ctx context.Context, id, playlistID string) (string, error) {
	body := map[string]any{"playlist_id": nil}
	if playlistID != "" {
		body["playlist_id"] = playlistID
	}
	var out struct {
		PlaylistID *string `json:"playlist_id"`
	}
	if err := s.client.json(ctx, "PUT", "/accounts/"+url.PathEscape(id)+"/youtube/playlists/default", body, nil, &out); err != nil {
		return "", err
	}
	if out.PlaylistID == nil {
		return "", nil
	}
	return *out.PlaylistID, nil
}

// ListYouTubeCaptions returns the caption tracks on one of the channel's videos.
func (s *AccountsService) ListYouTubeCaptions(ctx context.Context, id, videoID string) ([]YouTubeCaptionTrack, error) {
	var out []YouTubeCaptionTrack
	path := "/accounts/" + url.PathEscape(id) + "/youtube/videos/" + url.PathEscape(videoID) + "/captions"
	if err := s.client.json(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// UploadYouTubeCaptions uploads a caption track to a video.
func (s *AccountsService) UploadYouTubeCaptions(ctx context.Context, id, videoID string, body *UploadYouTubeCaptionsRequest) (*YouTubeCaptionTrack, error) {
	out := &YouTubeCaptionTrack{}
	path := "/accounts/" + url.PathEscape(id) + "/youtube/videos/" + url.PathEscape(videoID) + "/captions"
	if err := s.client.json(ctx, "POST", path, body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ReadYouTubeTranscript returns one caption track as text.
func (s *AccountsService) ReadYouTubeTranscript(ctx context.Context, id, captionID string) (*YouTubeTranscript, error) {
	out := &YouTubeTranscript{}
	path := "/accounts/" + url.PathEscape(id) + "/youtube/captions/" + url.PathEscape(captionID)
	if err := s.client.json(ctx, "GET", path, nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetBlueskyLanguages returns what a post from this connection is written in
// when the post itself does not say.
func (s *AccountsService) GetBlueskyLanguages(ctx context.Context, id string) (*BlueskyLanguages, error) {
	out := &BlueskyLanguages{}
	if err := s.client.json(ctx, "GET", "/accounts/"+url.PathEscape(id)+"/bluesky/languages", nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// SetBlueskyLanguages stores up to three BCP-47 tags. An empty slice clears the
// default.
func (s *AccountsService) SetBlueskyLanguages(ctx context.Context, id string, languages []string) (*BlueskyLanguages, error) {
	if languages == nil {
		languages = []string{}
	}
	out := &BlueskyLanguages{}
	body := map[string]any{"languages": languages}
	if err := s.client.json(ctx, "PUT", "/accounts/"+url.PathEscape(id)+"/bluesky/languages", body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetTikTokCreatorInfo returns the switches TikTok enforces at publish time.
func (s *AccountsService) GetTikTokCreatorInfo(ctx context.Context, id string) (*TikTokCreatorInfo, error) {
	out := &TikTokCreatorInfo{}
	if err := s.client.json(ctx, "GET", "/accounts/"+url.PathEscape(id)+"/tiktok/creator-info", nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// SearchTikTokMusic searches TikTok's Commercial Music Library. It needs the
// Marketing API product on the TikTok app; without it the call fails with 403
// rather than answering an empty list.
func (s *AccountsService) SearchTikTokMusic(ctx context.Context, id string, opts TikTokSearchOptions) ([]TikTokMusic, error) {
	var out []TikTokMusic
	path := "/accounts/" + url.PathEscape(id) + "/tiktok/music"
	if err := s.client.json(ctx, "GET", path, nil, tiktokSearchQuery(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SearchTikTokLocations searches the places a post can be tagged with. Same
// TikTok product as the music library.
func (s *AccountsService) SearchTikTokLocations(ctx context.Context, id string, opts TikTokSearchOptions) ([]TikTokPlace, error) {
	var out []TikTokPlace
	path := "/accounts/" + url.PathEscape(id) + "/tiktok/locations"
	if err := s.client.json(ctx, "GET", path, nil, tiktokSearchQuery(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// LookupTikTokVideo resolves a share link to one of this account's own videos,
// for repurposing. A link to someone else's video answers 404.
func (s *AccountsService) LookupTikTokVideo(ctx context.Context, id, shareURL string) (*TikTokVideoSource, error) {
	out := &TikTokVideoSource{}
	path := "/accounts/" + url.PathEscape(id) + "/tiktok/video-download"
	if err := s.client.json(ctx, "POST", path, map[string]any{"url": shareURL}, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

func tiktokSearchQuery(opts TikTokSearchOptions) url.Values {
	query := url.Values{}
	if opts.Query != "" {
		query.Set("q", opts.Query)
	}
	if opts.Limit > 0 {
		query.Set("limit", strconv.Itoa(opts.Limit))
	}
	return query
}

// SearchInstagramAudio returns tracks a Reel can carry.
func (s *AccountsService) SearchInstagramAudio(ctx context.Context, id string, opts *InstagramAudioSearchOptions) ([]InstagramAudio, error) {
	query := url.Values{}
	if opts != nil {
		if opts.Query != "" {
			query.Set("q", opts.Query)
		}
		if opts.AudioType != "" {
			query.Set("audio_type", opts.AudioType)
		}
	}
	var out []InstagramAudio
	if err := s.client.json(ctx, "GET", "/accounts/"+url.PathEscape(id)+"/instagram/audio", nil, query, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetInstagramPublishingLimit reports how many posts are left in the window.
func (s *AccountsService) GetInstagramPublishingLimit(ctx context.Context, id string) (*InstagramPublishingLimit, error) {
	out := &InstagramPublishingLimit{}
	path := "/accounts/" + url.PathEscape(id) + "/instagram/publishing-limit"
	if err := s.client.json(ctx, "GET", path, nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListInstagramStories returns stories still inside their 24 hours, whether or
// not they were posted through FoPost. Asking for insights costs one extra
// call per story.
func (s *AccountsService) ListInstagramStories(ctx context.Context, id string, insights bool) ([]InstagramStory, error) {
	query := url.Values{}
	if insights {
		query.Set("insights", strconv.FormatBool(insights))
	}
	var out []InstagramStory
	if err := s.client.json(ctx, "GET", "/accounts/"+url.PathEscape(id)+"/instagram/stories", nil, query, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetInstagramStoryInsights returns the insight set for one story.
func (s *AccountsService) GetInstagramStoryInsights(ctx context.Context, id, storyID string) (*InstagramStoryInsights, error) {
	out := &InstagramStoryInsights{}
	path := "/accounts/" + url.PathEscape(id) + "/instagram/stories/" + url.PathEscape(storyID) + "/insights"
	if err := s.client.json(ctx, "GET", path, nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// SearchLinkedInMentions returns organizations a post can mention. People are
// not searchable: LinkedIn has no public person search.
func (s *AccountsService) SearchLinkedInMentions(ctx context.Context, id, q string) ([]LinkedInMention, error) {
	query := url.Values{}
	query.Set("q", q)
	var out []LinkedInMention
	if err := s.client.json(ctx, "GET", "/accounts/"+url.PathEscape(id)+"/linkedin/mentions", nil, query, &out); err != nil {
		return nil, err
	}
	return out, nil
}
