package fopost

import (
	"context"
	"net/url"
)

// KnowledgeService covers the workspace knowledge base: what the workspace has
// told FoPost about itself.
//
// A source is an FAQ, a note, a page on your own site, or a plain-text/CSV item
// from the media library. Retrieval over these is what grounds a drafted inbox
// reply in your own answers instead of an invented one. Needs the inbox scope.
type KnowledgeService struct{ client *Client }

// KnowledgeSource is one thing the workspace has told FoPost about itself.
type KnowledgeSource struct {
	ID string `json:"id"`
	// Kind is "faq", "text", "url" or "file".
	Kind  string `json:"kind"`
	Title string `json:"title"`
	// Status is "pending", "syncing", "ready" or "failed". Only a ready
	// source is searched.
	Status string `json:"status"`
	// StatusMessage says why the last sync failed, in plain words.
	StatusMessage string `json:"statusMessage"`
	// URL is set for "url" sources.
	URL string `json:"url"`
	// MediaID is set for "file" sources: the media library item read.
	MediaID string `json:"mediaId"`
	// BrandVoiceID empty means the source serves the whole workspace.
	BrandVoiceID string `json:"brandVoiceId"`
	// ChunkCount is the searchable passages the last sync produced.
	ChunkCount int `json:"chunkCount"`
	// Content is the typed text, for "faq" and "text" sources only.
	Content      string `json:"content"`
	LastSyncedAt Time   `json:"lastSyncedAt"`
	CreatedAt    Time   `json:"createdAt"`
	UpdatedAt    Time   `json:"updatedAt"`
}

// KnowledgeMatch is one retrieved passage, with the source it came from so a
// reply can cite it.
type KnowledgeMatch struct {
	SourceID    string `json:"sourceId"`
	SourceTitle string `json:"sourceTitle"`
	SourceKind  string `json:"sourceKind"`
	SourceURL   string `json:"sourceUrl"`
	Text        string `json:"text"`
	// Score is the similarity to the question, 0-1.
	Score float64 `json:"score"`
}

// List returns the workspace's knowledge sources.
func (s *KnowledgeService) List(ctx context.Context, workspaceID string) ([]KnowledgeSource, error) {
	q := newQuery()
	q.str("workspace_id", workspaceID)
	var out []KnowledgeSource
	if err := s.client.json(ctx, "GET", "/knowledge/sources", nil, q.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateKnowledgeSourceRequest is the body of Create.
//
// Kind is "faq", "text", "url" or "file". An "faq" or "text" source needs
// Content, a "url" source needs URL, and a "file" source needs MediaID pointing
// at a plain-text or CSV item in the same workspace.
type CreateKnowledgeSourceRequest struct {
	Kind         string `json:"kind"`
	Title        string `json:"title"`
	Content      string `json:"content,omitempty"`
	URL          string `json:"url,omitempty"`
	MediaID      string `json:"media_id,omitempty"`
	BrandVoiceID string `json:"brand_voice_id,omitempty"`
	WorkspaceID  string `json:"workspace_id,omitempty"`
}

// Create adds a source and queues it for indexing, so it comes back pending.
func (s *KnowledgeService) Create(ctx context.Context, req CreateKnowledgeSourceRequest) (*KnowledgeSource, error) {
	out := &KnowledgeSource{}
	if err := s.client.json(ctx, "POST", "/knowledge/sources", req, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateKnowledgeSourceRequest is the body of Update. Only the non-nil fields
// are sent, so it stays a partial update.
type UpdateKnowledgeSourceRequest struct {
	Title        *string `json:"title,omitempty"`
	Content      *string `json:"content,omitempty"`
	URL          *string `json:"url,omitempty"`
	BrandVoiceID *string `json:"brand_voice_id,omitempty"`
}

// Update edits a source. Changing the content or the URL returns it to pending
// and re-indexes it.
func (s *KnowledgeService) Update(ctx context.Context, id string, req UpdateKnowledgeSourceRequest) (*KnowledgeSource, error) {
	out := &KnowledgeSource{}
	if err := s.client.json(ctx, "PATCH", "/knowledge/sources/"+url.PathEscape(id), req, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Delete removes a source and every passage indexed from it.
func (s *KnowledgeService) Delete(ctx context.Context, id string) error {
	return s.client.json(ctx, "DELETE", "/knowledge/sources/"+url.PathEscape(id), nil, nil, nil)
}

// Sync reads the source again — a "url" source is re-fetched. It returns once
// the re-index is queued, not once it has finished.
func (s *KnowledgeService) Sync(ctx context.Context, id string) error {
	return s.client.json(ctx, "POST", "/knowledge/sources/"+url.PathEscape(id)+"/sync", struct{}{}, nil, nil)
}

// SearchKnowledgeParams narrows a search. TopK defaults to 5 and caps at 20.
type SearchKnowledgeParams struct {
	TopK         int
	BrandVoiceID string
	WorkspaceID  string
}

// Search returns the passages closest to a question, best first. An empty
// slice is the honest answer when nothing stored answers it.
func (s *KnowledgeService) Search(ctx context.Context, query string, params SearchKnowledgeParams) ([]KnowledgeMatch, error) {
	q := newQuery()
	q.str("q", query)
	q.num("top_k", params.TopK)
	q.str("brand_voice_id", params.BrandVoiceID)
	q.str("workspace_id", params.WorkspaceID)
	var out []KnowledgeMatch
	if err := s.client.json(ctx, "GET", "/knowledge/search", nil, q.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}
