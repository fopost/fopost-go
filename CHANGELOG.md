# Changelog

All notable changes to this module are documented here. Releases are tags on `main`.

## Unreleased

- `Blog` service: blog settings (`GetSettings`, `UpdateSettings`, `RotateWebsiteSecret`, `TestWebsite`), keywords (`ListKeywords`, `PlanKeywords`, `AddKeyword`, `UpdateKeyword`, `DeleteKeyword`) and articles (`ListArticles`, `RequestArticle`, `GetArticle`, `UpdateArticle`, `RegenerateArticle`, `PublishArticle`, `RejectArticle`, `DeleteArticle`), with the `BlogSettings`, `BlogKeyword`, `BlogArticle`, `BlogArticleFull` and `BlogArticleImage` types and the `BlogArticleStatus*`, `BlogKeywordStatus*` and `BlogTarget*` constants.
- `VerifyWebhook(rawBody, headers, secret)` checks a delivery's HMAC-SHA256 signature in constant time, preferring the timestamped `X-FoPost-Signature-256` within a 300 second tolerance and falling back to `X-FoPost-Signature`, and returns the `WebhookEnvelope`. `BlogArticlePublished` types the `article.published` payload; `EventArticlePublished` and `EventArticleTest` name the events.
- `Media.Limits` reads `GET /media/limits`: the account's upload ceiling, the multipart per-file ceiling, the next tier and the allowed types.
- `Posts.Create` and `Posts.Update` take `PostFile`s. Files under the multipart ceiling travel with the post in one request; larger ones go through `UploadDirect` and are attached to the first content block. `OpenPostFile` opens a path.
- `PresignInput` gains `FolderID`, `Origin`, `BrandRole`, `BrandVoiceID` and `PersonConsent`; `PresignedUpload` gains `MaxFileBytes`. The 50 MB presign ceiling is gone; sizes go up to the account's `MaxFileBytes`.
- `IsPayloadTooLarge` and `(*Error).MaxFileBytes()` read a `413 file_too_large`.
- **Breaking:** `Media.UploadDirect(ctx, *PresignInput, io.Reader)` streams the file instead of taking a `[]byte`.
