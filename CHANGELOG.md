# Changelog

All notable changes to this module are documented here. Releases are tags on `main`.

## Unreleased

- `Ads.SendConversions(ctx, idempotencyKey, body)` sends conversions against a pixel exactly once (`POST /ads/conversions/events`, `Idempotency-Key` required) and answers a `ConversionUploadResult` (`ID`, `Accepted`, `EventCount`, `Replayed`). `Ads.ListConversionEvents` reads the stored `ConversionBatch` rows with their hashed events. The LinkedIn rule method keeps the `SendConversionEvents` name.
- `ConversionEvent` gains `EventID`, `ActionSource` (`ConversionActionSource*`), `SourceURL`, `ExternalID`, `ClientIPAddress`, `ClientUserAgent`, `ClickID`, `BrowserID` and `ContentIDs`. `Ads.UploadConversions` keeps returning the accepted count.
- Pixels: `Ads.ListPixels`, `CreatePixel`, `GetPixel`, `UpdatePixel`, `SharePixel`, `GetPixelStats` and `GetEventMatchQuality`, with `AdPixelDetail`, `AdPixelStats`, `EventMatchQuality` and the `PixelDataUse*` constants.
- Custom conversions: `Ads.ListCustomConversions`, `CreateCustomConversion`, `GetCustomConversion`, `UpdateCustomConversion` and `DeleteCustomConversion`, with `CustomConversion`, `CustomConversionRule` and the `CustomConversion*` event type constants.
- X engagement: `Inbox.Repost`, `Inbox.Unrepost`, `Inbox.Bookmark` and `Inbox.Unbookmark` (`POST /inbox/{id}/repost|unrepost|bookmark|unbookmark`, need `inbox` and `publish`), with `Reposted`, `Bookmarked`, `CanRepost` and `CanBookmark` on `InboxItem`.
- `Accounts.FollowXUser` and `Accounts.UnfollowXUser` follow an X user by numeric id as the account (`POST /accounts/{id}/x/follow|unfollow`), answering an `XFollowResult`.
- `Inbox.GetXPost` and `Inbox.SearchXPosts` read X posts with a connected account's token (`GET /x/posts/{id}`, `GET /x/search`), with the `XPost`, `XSearchParams`, `XSearchMeta` and `XPostList` types; search pages through `Meta.NextToken`.
- Existing X accounts must be reconnected for repost, bookmark and follow: the grant now asks for the bookmark and follow permissions, and an older token answers `409` `reconnect_required`. Every one of these calls is an X API call metered to the workspace; a spent X balance is a `402` `platform_balance_spent`.

## 0.4.0 - 2026-10-18

- `Queue` service for the posting queue: `ListSlots`, `CreateSlot`, `UpdateSlot`, `DeleteSlot`, `Next` and `Preview`, with the `QueueSlot`, `UpdatedQueueSlot`, `DeletedQueueSlot` and `QueueMove` types. `CreatePostRequest` and `UpdatePostRequest` gain `Schedule`; set it to `ScheduleQueue` to land the post in the next free slot.
- `Posts.Unpublish` takes a live post down on every network (`POST /posts/{id}/unpublish`, needs `publish`), answering `UnpublishResult`; `PostStatusUnpublished`.
- `ConnectSessions` service for hosted connect links: `Create` and `Get`, with `ConnectSession` and the `ConnectSession*` status constants.
- `Webhooks.ListDeliveries`, `Webhooks.GetDelivery` and `Webhooks.Redeliver` read and replay the delivery log, with the `WebhookDelivery` type.
- `Automations.ListVersions`, `Automations.GetVersion` and `Automations.RestoreVersion` walk and restore an automation's saved graphs, with `AutomationVersion`, `AutomationVersionSummary`, `RestoredAutomationVersion`, `WorkflowGraph`, `WorkflowNode` and `WorkflowEdge`.
- `Logs` service: `List` reads recent API activity with its summary and cursor (`GET /logs`, needs `workspaces`).
- `Analytics.Inbox` reads the inbox roll-up (`GET /analytics/inbox`).

## 0.3.0 - 2026-10-18

- **Breaking:** the `Blog` service now speaks the live API contract: `workspace_id` on the query string, snake_case response fields, and an unwrapped `UpdateSettings` response with `website_secret` beside the settings. Request bodies are unchanged. `BlogWebsiteTest` is now `{HTTPStatus}`; a website that refuses the test is a `502` `*Error` whose `http_status` field holds its answer.
- **Breaking:** `BlogSettings.AutoPublish` and `UpdateBlogSettingsRequest.AutoPublish` are removed; the API no longer has the setting. Every article now goes through review.
- `Blog.SuggestSeeds` suggests about ten search terms from the blog goal and topics (`POST /blog/keywords/suggest`, needs `ai`).
- `Blog.SubmitArticle` moves a draft, failed or rejected article to `in_review` (`POST /blog/articles/{id}/submit`). `PublishArticle` and `PublishAt` now need the article in review; a draft answers `409`.
- `BlogSettings` gains `Goal`, `FocusTopics`, `AvoidTopics` and `PlanRun` (`BlogPlanRun`); `UpdateBlogSettingsRequest` gains `Goal`, `FocusTopics` and `AvoidTopics`. `ReviewHoldDays` now means how many days before its planned date a picked keyword starts being written.
- `BlogKeyword.FitReason` says why a keyword serves the blog goal. `PlanKeywords` no longer fills the calendar: setting a keyword to `planned` picks it.
- `BlogArticleStatusInReview` and the `BlogGoal*` constants.

- Reddit destinations: `Accounts.ListRedditSubreddits`, `Accounts.ListRedditSubredditRules`, `Accounts.ListRedditFlairs` and `Accounts.SetRedditDefaultSubreddit`, with `Validate.Subreddit` answering the same existence check before a post exists, and `Inbox.Vote(ctx, id, direction)` with `up`, `down` or `none` to take a vote back. A flair id is valid only in its own subreddit; `Account.ReconnectRequired` says a grant predates the subreddit scopes.
- `Blog` service: blog settings (`GetSettings`, `UpdateSettings`, `RotateWebsiteSecret`, `TestWebsite`), keywords (`ListKeywords`, `PlanKeywords`, `AddKeyword`, `UpdateKeyword`, `DeleteKeyword`) and articles (`ListArticles`, `RequestArticle`, `GetArticle`, `UpdateArticle`, `RegenerateArticle`, `PublishArticle`, `RejectArticle`, `DeleteArticle`), with the `BlogSettings`, `BlogKeyword`, `BlogArticle`, `BlogArticleFull` and `BlogArticleImage` types and the `BlogArticleStatus*`, `BlogKeywordStatus*` and `BlogTarget*` constants.
- `VerifyWebhook(rawBody, headers, secret)` checks a delivery's HMAC-SHA256 signature in constant time, preferring the timestamped `X-FoPost-Signature-256` within a 300 second tolerance and falling back to `X-FoPost-Signature`, and returns the `WebhookEnvelope`. `BlogArticlePublished` types the `article.published` payload; `EventArticlePublished` and `EventArticleTest` name the events.
- `Media.Limits` reads `GET /media/limits`: the account's upload ceiling, the multipart per-file ceiling, the next tier and the allowed types.
- `Posts.Create` and `Posts.Update` take `PostFile`s. Files under the multipart ceiling travel with the post in one request; larger ones go through `UploadDirect` and are attached to the first content block. `OpenPostFile` opens a path.
- `PresignInput` gains `FolderID`, `Origin`, `BrandRole`, `BrandVoiceID` and `PersonConsent`; `PresignedUpload` gains `MaxFileBytes`. The 50 MB presign ceiling is gone; sizes go up to the account's `MaxFileBytes`.
- `IsPayloadTooLarge` and `(*Error).MaxFileBytes()` read a `413 file_too_large`.
- **Breaking:** `Media.UploadDirect(ctx, *PresignInput, io.Reader)` streams the file instead of taking a `[]byte`.
