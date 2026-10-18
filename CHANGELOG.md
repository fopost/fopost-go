# Changelog

All notable changes to this module are documented here. Releases are tags on `main`.

## Unreleased

- **Breaking:** the `Blog` service now speaks the live API contract: `workspace_id` on the query string, snake_case response fields, and an unwrapped `UpdateSettings` response with `website_secret` beside the settings. Request bodies are unchanged. `BlogWebsiteTest` is now `{HTTPStatus}`; a website that refuses the test is a `502` `*Error` whose `http_status` field holds its answer.
- **Breaking:** `BlogSettings.AutoPublish` and `UpdateBlogSettingsRequest.AutoPublish` are removed; the API no longer has the setting. Every article now goes through review.
- `Blog.SuggestSeeds` suggests about ten search terms from the blog goal and topics (`POST /blog/keywords/suggest`, needs `ai`).
- `Blog.SubmitArticle` moves a draft, failed or rejected article to `in_review` (`POST /blog/articles/{id}/submit`). `PublishArticle` and `PublishAt` now need the article in review; a draft answers `409`.
- `BlogSettings` gains `Goal`, `FocusTopics`, `AvoidTopics` and `PlanRun` (`BlogPlanRun`); `UpdateBlogSettingsRequest` gains `Goal`, `FocusTopics` and `AvoidTopics`. `ReviewHoldDays` now means how many days before its planned date a picked keyword starts being written.
- `BlogKeyword.FitReason` says why a keyword serves the blog goal. `PlanKeywords` no longer fills the calendar: setting a keyword to `planned` picks it.
- `BlogArticleStatusInReview` and the `BlogGoal*` constants.

- `Blog` service: blog settings (`GetSettings`, `UpdateSettings`, `RotateWebsiteSecret`, `TestWebsite`), keywords (`ListKeywords`, `PlanKeywords`, `AddKeyword`, `UpdateKeyword`, `DeleteKeyword`) and articles (`ListArticles`, `RequestArticle`, `GetArticle`, `UpdateArticle`, `RegenerateArticle`, `PublishArticle`, `RejectArticle`, `DeleteArticle`), with the `BlogSettings`, `BlogKeyword`, `BlogArticle`, `BlogArticleFull` and `BlogArticleImage` types and the `BlogArticleStatus*`, `BlogKeywordStatus*` and `BlogTarget*` constants.
- `VerifyWebhook(rawBody, headers, secret)` checks a delivery's HMAC-SHA256 signature in constant time, preferring the timestamped `X-FoPost-Signature-256` within a 300 second tolerance and falling back to `X-FoPost-Signature`, and returns the `WebhookEnvelope`. `BlogArticlePublished` types the `article.published` payload; `EventArticlePublished` and `EventArticleTest` name the events.
- `Media.Limits` reads `GET /media/limits`: the account's upload ceiling, the multipart per-file ceiling, the next tier and the allowed types.
- `Posts.Create` and `Posts.Update` take `PostFile`s. Files under the multipart ceiling travel with the post in one request; larger ones go through `UploadDirect` and are attached to the first content block. `OpenPostFile` opens a path.
- `PresignInput` gains `FolderID`, `Origin`, `BrandRole`, `BrandVoiceID` and `PersonConsent`; `PresignedUpload` gains `MaxFileBytes`. The 50 MB presign ceiling is gone; sizes go up to the account's `MaxFileBytes`.
- `IsPayloadTooLarge` and `(*Error).MaxFileBytes()` read a `413 file_too_large`.
- **Breaking:** `Media.UploadDirect(ctx, *PresignInput, io.Reader)` streams the file instead of taking a `[]byte`.
