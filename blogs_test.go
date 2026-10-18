package fopost

import (
	"context"
	"io"
	"net/http"
	"testing"
)

const articleJSON = `{"id":"99","blog_id":"11","title":"Spring drop","body_html":"<p>Hello</p>",` +
	`"excerpt":"A short summary","status":"published","author_name":"Store Owner","tags":["news"],` +
	`"image_url":null,"url":"https://demo.myshopify.com/blogs/article/spring-drop",` +
	`"published_at":"2026-09-01T10:00:00.000Z","updated_at":null}`

func TestBlogsListBlogsDecodesEveryBlog(t *testing.T) {
	var path string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		_, _ = io.WriteString(w, `{"data":[{"id":"11","title":"News","handle":"news","url":null}]}`)
	})

	blogs, err := client.Blogs.ListBlogs(context.Background(), "acc_1")
	if err != nil {
		t.Fatalf("Blogs.ListBlogs: %v", err)
	}
	if path != "/accounts/acc_1/blogs" {
		t.Fatalf("path = %q", path)
	}
	if len(blogs) != 1 || blogs[0].ID != "11" || blogs[0].Title != "News" {
		t.Fatalf("blogs = %+v", blogs)
	}
}

func TestBlogsListArticlesSendsTheFilters(t *testing.T) {
	var query string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.Query().Encode()
		_, _ = io.WriteString(w, `{"data":[`+articleJSON+`]}`)
	})

	articles, err := client.Blogs.ListArticles(context.Background(), "acc_1", "11",
		&ListArticlesParams{Limit: 5, Status: "draft", Q: "spring"})
	if err != nil {
		t.Fatalf("Blogs.ListArticles: %v", err)
	}
	if query != "limit=5&q=spring&status=draft" {
		t.Fatalf("query = %q", query)
	}
	if len(articles) != 1 || articles[0].ID != "99" || len(articles[0].Tags) != 1 {
		t.Fatalf("articles = %+v", articles)
	}
}

// The article id is in the path, which is what stops an edit from creating a
// second post on the site.
func TestBlogsUpdateArticleChangesItInPlace(t *testing.T) {
	var method, path, body string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		body = string(raw)
		_, _ = io.WriteString(w, `{"data":`+articleJSON+`}`)
	})

	if _, err := client.Blogs.UpdateArticle(context.Background(), "acc_1", "11", "99",
		&ArticleRequest{Title: "Spring drop, restocked"}); err != nil {
		t.Fatalf("Blogs.UpdateArticle: %v", err)
	}

	if method != "PATCH" {
		t.Fatalf("method = %q", method)
	}
	if path != "/accounts/acc_1/blogs/11/articles/99" {
		t.Fatalf("path = %q", path)
	}
	// Only what the caller set travels, so nothing else on the article is blanked.
	if body != `{"title":"Spring drop, restocked"}`+"\n" && body != `{"title":"Spring drop, restocked"}` {
		t.Fatalf("body = %q", body)
	}
}

func TestBlogsDeleteArticleHitsTheArticleRoute(t *testing.T) {
	var method, path string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})

	if err := client.Blogs.DeleteArticle(context.Background(), "acc_1", "11", "99"); err != nil {
		t.Fatalf("Blogs.DeleteArticle: %v", err)
	}
	if method != "DELETE" || path != "/accounts/acc_1/blogs/11/articles/99" {
		t.Fatalf("%s %s", method, path)
	}
}

func TestBlogsUpdateProductSendsOnlyWhatChanged(t *testing.T) {
	var body string
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		body = string(raw)
		_, _ = io.WriteString(w, `{"data":{"id":"7","title":"Mug XL","handle":"mug","status":"draft",`+
			`"description":null,"vendor":null,"product_type":"Drinkware","tags":[],"image_url":null,`+
			`"url":null,"price":"12.00","currency":"USD","updated_at":null}}`)
	})

	product, err := client.Blogs.UpdateProduct(context.Background(), "acc_1", "7",
		&ProductRequest{Title: "Mug XL", ProductType: "Drinkware"})
	if err != nil {
		t.Fatalf("Blogs.UpdateProduct: %v", err)
	}
	if want := `{"title":"Mug XL","product_type":"Drinkware"}`; body != want && body != want+"\n" {
		t.Fatalf("body = %q", body)
	}
	if product.Price == nil || *product.Price != "12.00" {
		t.Fatalf("product = %+v", product)
	}
}
