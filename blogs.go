package fopost

import (
	"context"
	"net/url"
)

// BlogsService reaches content a connected site already owns: the articles on
// a WordPress site or a Shopify store's blog, and a Shopify store's products.
//
// Every id here is the platform's own, never a FoPost id. Reads need the posts
// scope; anything that changes the site needs posts and publish, because a
// change here is visible to the site's own readers.
type BlogsService struct{ client *Client }

// RemoteBlog is a blog on a connected site. Shopify reports every blog on the
// store; WordPress has one implicit blog and reports it under the id "default",
// so both answer the same shape.
type RemoteBlog struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Handle *string `json:"handle"`
	URL    *string `json:"url"`
}

// RemoteArticle is an article that already lives on a connected site. Status is
// one of published, draft, pending or scheduled.
type RemoteArticle struct {
	ID          string   `json:"id"`
	BlogID      *string  `json:"blog_id"`
	Title       string   `json:"title"`
	BodyHTML    *string  `json:"body_html"`
	Excerpt     *string  `json:"excerpt"`
	Status      string   `json:"status"`
	AuthorName  *string  `json:"author_name"`
	Tags        []string `json:"tags"`
	ImageURL    *string  `json:"image_url"`
	URL         *string  `json:"url"`
	PublishedAt *string  `json:"published_at"`
	UpdatedAt   *string  `json:"updated_at"`
}

// RemoteProduct is a product on a connected store. Price is the lowest variant
// price, as a decimal string. Status is active, draft or archived.
type RemoteProduct struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Handle      *string  `json:"handle"`
	Status      string   `json:"status"`
	Description *string  `json:"description"`
	Vendor      *string  `json:"vendor"`
	ProductType *string  `json:"product_type"`
	Tags        []string `json:"tags"`
	ImageURL    *string  `json:"image_url"`
	URL         *string  `json:"url"`
	Price       *string  `json:"price"`
	Currency    *string  `json:"currency"`
	UpdatedAt   *string  `json:"updated_at"`
}

// ListArticlesParams narrows ListArticles. Zero fields are not sent.
type ListArticlesParams struct {
	// Limit is 1 to 50; the API defaults to 20.
	Limit int
	// Status is published, draft, pending or scheduled.
	Status string
	// Q matches the article title.
	Q string
}

func (p *ListArticlesParams) values() url.Values {
	q := newQuery()
	if p == nil {
		return q.values()
	}
	q.num("limit", p.Limit)
	q.str("status", p.Status)
	q.str("q", p.Q)
	return q.values()
}

// ListProductsParams narrows ListProducts. Zero fields are not sent.
type ListProductsParams struct {
	Limit int
	// Status is active, draft or archived.
	Status string
	// Q matches the product title.
	Q string
}

func (p *ListProductsParams) values() url.Values {
	q := newQuery()
	if p == nil {
		return q.values()
	}
	q.num("limit", p.Limit)
	q.str("status", p.Status)
	q.str("q", p.Q)
	return q.values()
}

// ArticleRequest is the body of CreateArticle and UpdateArticle. On an update
// every field is optional and only what is set travels, so an omitted field
// keeps whatever the site already had.
type ArticleRequest struct {
	Title string `json:"title,omitempty"`
	// Body is FoPost body markup; the site's own format is rendered from it.
	Body       string   `json:"body,omitempty"`
	Excerpt    string   `json:"excerpt,omitempty"`
	Status     string   `json:"status,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	AuthorName string   `json:"author_name,omitempty"`
	ImageURL   string   `json:"image_url,omitempty"`
}

// ProductRequest is the body of UpdateProduct. Only what is set travels.
type ProductRequest struct {
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Status      string   `json:"status,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	ProductType string   `json:"product_type,omitempty"`
	Vendor      string   `json:"vendor,omitempty"`
}

func articlePath(accountID, blogID string) string {
	return "/accounts/" + url.PathEscape(accountID) + "/blogs/" + url.PathEscape(blogID) + "/articles"
}

// ListBlogs returns the blogs the account can write to.
func (s *BlogsService) ListBlogs(ctx context.Context, accountID string) ([]RemoteBlog, error) {
	var out []RemoteBlog
	path := "/accounts/" + url.PathEscape(accountID) + "/blogs"
	if err := s.client.json(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListArticles returns the blog's articles, newest first, drafts included.
func (s *BlogsService) ListArticles(ctx context.Context, accountID, blogID string, params *ListArticlesParams) ([]RemoteArticle, error) {
	var out []RemoteArticle
	if err := s.client.json(ctx, "GET", articlePath(accountID, blogID), nil, params.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetArticle returns one article in full.
func (s *BlogsService) GetArticle(ctx context.Context, accountID, blogID, articleID string) (*RemoteArticle, error) {
	out := &RemoteArticle{}
	path := articlePath(accountID, blogID) + "/" + url.PathEscape(articleID)
	if err := s.client.json(ctx, "GET", path, nil, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateArticle writes a new article to the blog. Needs the publish scope.
func (s *BlogsService) CreateArticle(ctx context.Context, accountID, blogID string, body *ArticleRequest) (*RemoteArticle, error) {
	out := &RemoteArticle{}
	if err := s.client.json(ctx, "POST", articlePath(accountID, blogID), body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateArticle changes the live article in place. Needs the publish scope.
//
// The article is addressed by its own id and only the fields set on body
// travel, so an edit never creates a second post on the site.
func (s *BlogsService) UpdateArticle(ctx context.Context, accountID, blogID, articleID string, body *ArticleRequest) (*RemoteArticle, error) {
	out := &RemoteArticle{}
	path := articlePath(accountID, blogID) + "/" + url.PathEscape(articleID)
	if err := s.client.json(ctx, "PATCH", path, body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteArticle removes the article from the site. Needs the publish scope;
// this cannot be undone.
func (s *BlogsService) DeleteArticle(ctx context.Context, accountID, blogID, articleID string) error {
	path := articlePath(accountID, blogID) + "/" + url.PathEscape(articleID)
	return s.client.json(ctx, "DELETE", path, nil, nil, nil)
}

// ListProducts returns the store's products.
func (s *BlogsService) ListProducts(ctx context.Context, accountID string, params *ListProductsParams) ([]RemoteProduct, error) {
	var out []RemoteProduct
	path := "/accounts/" + url.PathEscape(accountID) + "/products"
	if err := s.client.json(ctx, "GET", path, nil, params.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateProduct changes a product on the store. Needs the publish scope; only
// the fields set on body change.
func (s *BlogsService) UpdateProduct(ctx context.Context, accountID, productID string, body *ProductRequest) (*RemoteProduct, error) {
	out := &RemoteProduct{}
	path := "/accounts/" + url.PathEscape(accountID) + "/products/" + url.PathEscape(productID)
	if err := s.client.json(ctx, "PATCH", path, body, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}
