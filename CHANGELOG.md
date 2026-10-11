# Changelog

All notable changes to this module are documented here. Releases are tags on `main`.

## Unreleased

### Added

- `Blog` service: blog settings (`GetSettings`, `UpdateSettings`, `RotateWebsiteSecret`, `TestWebsite`), keywords (`ListKeywords`, `PlanKeywords`, `AddKeyword`, `UpdateKeyword`, `DeleteKeyword`) and articles (`ListArticles`, `RequestArticle`, `GetArticle`, `UpdateArticle`, `RegenerateArticle`, `PublishArticle`, `RejectArticle`, `DeleteArticle`), with the `BlogSettings`, `BlogKeyword`, `BlogArticle`, `BlogArticleFull` and `BlogArticleImage` types and the `BlogArticleStatus*`, `BlogKeywordStatus*` and `BlogTarget*` constants.
- `VerifyWebhook(rawBody, headers, secret)` checks a delivery's HMAC-SHA256 signature in constant time, preferring the timestamped `X-FoPost-Signature-256` within a 300 second tolerance and falling back to `X-FoPost-Signature`, and returns the `WebhookEnvelope`. `BlogArticlePublished` types the `article.published` payload; `EventArticlePublished` and `EventArticleTest` name the events.
