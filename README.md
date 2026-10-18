# fopost-go

[![Go Reference](https://pkg.go.dev/badge/github.com/fopost/fopost-go.svg)](https://pkg.go.dev/github.com/fopost/fopost-go)
[![CI](https://github.com/fopost/fopost-go/actions/workflows/ci.yml/badge.svg)](https://github.com/fopost/fopost-go/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Official Go SDK for the [FoPost](https://fopost.com) API. Schedule and publish to +30 social platforms from your code.

```bash
go get github.com/fopost/fopost-go
```

Requires Go 1.22 or newer. No dependencies beyond the standard library.

> **0.x release.** The public API is still settling and minor versions may
> contain breaking changes. Pin an exact version if that matters to you.

## Quick start

```go
package main

import (
	"context"
	"log"

	"github.com/fopost/fopost-go"
)

func main() {
	client, err := fopost.New("fp_...") // or leave it empty and set FOPOST_API_KEY
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	workspaces, err := client.Workspaces.List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	workspace := workspaces[0]

	accounts, err := client.Accounts.List(ctx, workspace.ID)
	if err != nil {
		log.Fatal(err)
	}

	ids := make([]string, 0, len(accounts))
	for _, account := range accounts {
		ids = append(ids, account.ID)
	}

	post, err := client.Posts.Create(ctx, &fopost.CreatePostRequest{
		WorkspaceID: workspace.ID,
		Accounts:    ids,
		Content:     fopost.Text("Hello from Go"),
	})
	if err != nil {
		log.Fatal(err)
	}

	if _, err := client.Posts.Publish(ctx, post.ID, nil); err != nil {
		log.Fatal(err)
	}
}
```

Get a key from **Settings → API Keys** in the FoPost dashboard
(<https://fopost.com/dashboard/settings/api-keys>). Every method takes a
`context.Context` first, and the client is safe for concurrent use.

## Content

`Text` builds a single-block post and `Thread` builds one block per entry. Media
is attached per block:

```go
client.Posts.Create(ctx, &fopost.CreatePostRequest{
	WorkspaceID: workspace.ID,
	Accounts:    ids,
	Content: []fopost.ContentBlock{
		{Text: "First post in the thread"},
		{
			Text: "Second one, with an image",
			Media: []fopost.MediaItem{
				{Type: "image", Name: "chart.png", URL: "https://.../chart.png"},
			},
		},
	},
})
```

Uploading through the media library gives you an item that drops straight in:

```go
file, _ := os.Open("chart.png")
defer file.Close()

uploaded, err := client.Media.Upload(ctx, workspace.ID, fopost.File{Name: "chart.png", Content: file})
if err != nil {
	log.Fatal(err)
}

content := []fopost.ContentBlock{{
	Text:  "Numbers are in",
	Media: []fopost.MediaItem{uploaded[0].AsMediaItem()},
}}
```

Files can also ride along with the post itself. `Posts.Create` and `Posts.Update`
take `PostFile`s after the body; they land in the workspace library and attach
to the first content block. Small files travel in the same request as the post;
anything over the multipart ceiling is sent straight to storage first and
attached by URL. The caller never sees the split.

```go
clip, err := fopost.OpenPostFile("clip.mp4")
if err != nil {
	log.Fatal(err)
}
defer clip.Close()

post, err := client.Posts.Create(ctx, &fopost.CreatePostRequest{
	WorkspaceID: workspace.ID,
	Accounts:    []string{account.ID},
	Content:     fopost.Text("Watch this"),
}, *clip)
```

A post takes up to `fopost.MaxPostFiles` files. `Media.Limits` reports what the
account may upload: `MaxFileBytes` is the largest file any route accepts and
grows with the number of connected accounts (or the Large Uploads add-on), and
`MultipartMaxFileBytes` is what one multipart request carries per file.

```go
limits, err := client.Media.Limits(ctx, workspace.ID) // "" for the key's default
fmt.Println(limits.MaxFileBytes, limits.MultipartMaxFileBytes, limits.NextTier)
```

Larger files can go straight to storage instead of through the API. `UploadDirect`
presigns a slot, streams the bytes to it with no API key, and completes the upload.
It reads from an `io.Reader` with a declared size, so nothing is held in memory:

```go
file, _ := os.Open("clip.mp4")
defer file.Close()
info, _ := file.Stat()

uploaded, err := client.Media.UploadDirect(ctx, &fopost.PresignInput{
	WorkspaceID: workspace.ID,
	Filename:    "clip.mp4",
	MimeType:    "video/mp4",
	Size:        info.Size(),
}, file)
if err != nil {
	log.Fatal(err)
}
```

`Presign` and `Complete` are the two halves when you want to send the bytes
yourself. A file over the account's ceiling is a `413` with code `file_too_large`;
`fopost.IsPayloadTooLarge(err)` spots it and `(*Error).MaxFileBytes()` reads the
ceiling it reports.

## Scheduling and publishing

`Status` is `"draft"` or `"scheduled"`; a scheduled post needs `ScheduleAt`. To
send something out now, create it and call `Publish`. Nothing reaches a platform
without one of those two.

```go
at := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)

post, err := client.Posts.Create(ctx, &fopost.CreatePostRequest{
	WorkspaceID: workspace.ID,
	Accounts:    ids,
	Content:     fopost.Text("Scheduled with the SDK"),
	Status:      fopost.PostStatusScheduled,
	ScheduleAt:  fopost.NewTime(at),
})
```

`Validate` checks content against platform rules before a post exists, and
nothing is stored:

```go
check, _ := client.Validate.Post(ctx, &fopost.ValidatePostRequest{
	Content:   "Hello from Go",
	Platforms: []string{"twitter", "linkedin"},
})
for _, platform := range check.Platforms {
	fmt.Println(platform.Platform, platform.Ready, platform.Issues)
}

length, _ := client.Validate.Length(ctx, &fopost.ValidateLengthRequest{Text: "Hello", Platforms: []string{"bluesky"}})
media, _ := client.Validate.Media(ctx, "https://yourbrand.com/chart.png")
fmt.Println(length.OK, media.OK, media.Issues)
```

`WhatsApp` reaches a WhatsApp Business number the customer already owns. The
platform owns the templates, flows, profile and commerce settings, so every call
is live and all of it answers 503 until WhatsApp is set up:

```go
// Filing a template returns the review status the platform gave it, never an
// assumed one.
template, _ := client.WhatsApp.CreateTemplate(ctx, accountID, &fopost.CreateWhatsAppTemplateRequest{
	Name:       "order_shipped",
	Language:   "en_US",
	Category:   "UTILITY",
	Components: []map[string]any{{"type": "BODY", "text": "Your order is on its way."}},
})
fmt.Println(template.Status) // PENDING until the platform approves it

// A flow is created as a draft, its screens uploaded, then published.
flow, _ := client.WhatsApp.CreateFlow(ctx, accountID, &fopost.CreateWhatsAppFlowRequest{
	Name:       "Book a fitting",
	Categories: []string{"LEAD_GENERATION"},
})
client.WhatsApp.UploadFlowJSON(ctx, accountID, flow.ID, map[string]any{"version": "7.0", "screens": []any{}})
client.WhatsApp.PublishFlow(ctx, accountID, flow.ID)
answers, _ := client.WhatsApp.FlowResponses(ctx, accountID)
fmt.Println(len(answers))
```

`Preflight` reports per-account blockers and advisory signals without publishing,
and `Publish` with `DryRun` validates the whole delivery plan:

```go
check, _ := client.Posts.Preflight(ctx, post.ID)
for _, account := range check.Accounts {
	fmt.Println(account.Platform, account.Ready, account.Issues)
}

plan, _ := client.Posts.Publish(ctx, post.ID, &fopost.PublishOptions{DryRun: true})
fmt.Println(plan.DryRun, plan.HealthWarnings)
```

After publishing, `Deliveries` is the per-account state, `Retry` re-sends only
what failed, and `Cancel` stops what has not gone out yet.

## Pagination

`Posts.List` returns one page with its `Meta`. `Each` walks every page for you,
and returning an error from the callback stops the walk:

```go
page, _ := client.Posts.List(ctx, &fopost.ListPostsParams{
	WorkspaceID: workspace.ID,
	Status:      fopost.PostStatusPublished,
	PerPage:     50,
})
fmt.Printf("%d published posts\n", page.Meta.Total)

err := client.Posts.Each(ctx, &fopost.ListPostsParams{WorkspaceID: workspace.ID}, func(post fopost.Post) error {
	fmt.Println(post.ID, post.Status)
	return nil
})
```

Zero-valued parameters are not sent, so the API applies its own defaults.

## Analytics

```go
// How long a post keeps earning, from the repeated readings of each post
decay, err := client.Analytics.Decay(ctx, &fopost.AnalyticsParams{Days: 30})
fmt.Println(decay.HalfLifeBucket) // e.g. "1h_3h"

// Whether posting more earned more
cadence, err := client.Analytics.Frequency(ctx, &fopost.AnalyticsParams{Days: 90})
if cadence.Best != nil {
    fmt.Println(cadence.Best.Label) // e.g. "3-5 a week"
}

// Every reading held for one post, with what moved between them
timeline, err := client.Analytics.Timeline(ctx, post.ID)

// Mirror the metrics into your own store, without refetching everything
cursor := ""
for {
    page, err := client.Analytics.Changes(ctx, &fopost.MetricChangesParams{Since: cursor})
    if err != nil {
        return err
    }
    save(page.Changes)
    if !page.HasMore {
        break
    }
    cursor = page.Cursor.String()
}

// Refresh one post now instead of waiting for the next collection run
_, err = client.Analytics.CollectPost(ctx, post.ID)

// Posts on the account that never went out through FoPost
native, err := client.Analytics.NativePosts(ctx, accounts[0].ID, nil)
```

A post is addressed by its FoPost id or by its permalink, so a post made by
hand on the network works the same way:

```go
client.Analytics.Timeline(ctx, "https://x.com/acme/status/1")
```

## Configuration

```go
client, err := fopost.New(apiKey,
	fopost.WithBaseURL("http://localhost:8080/v1"), // point at another deployment
	fopost.WithTimeout(30*time.Second),             // per request
	fopost.WithMaxRetries(3),                       // total attempts
	fopost.WithHTTPClient(myClient),                // bring your own transport
	fopost.WithUserAgent("my-app/2.0"),             // prefixed to the SDK's
)
```

| Env var           | Used for                                        |
| ----------------- | ----------------------------------------------- |
| `FOPOST_API_KEY`  | API key, when the one passed to `New` is empty  |
| `FOPOST_BASE_URL` | API root, when `WithBaseURL` is not given       |

A `429` is retried automatically, waiting the interval the API asks for in
`Retry-After` (delta-seconds or an HTTP date, capped at 60s). `5xx` responses and
transport errors back off exponentially. `MaxRetries` counts total attempts, so
the default of 3 means two retries; a cancelled context stops the wait
immediately.

## Errors

Every non-2xx response is an `*fopost.Error` carrying the API's `Status`, `Code`,
and `Message`, plus the rate-limit headers that came with it.

```go
if _, err := client.Posts.Publish(ctx, postID, nil); err != nil {
	switch {
	case fopost.IsPaymentRequired(err):
		apiErr, _ := fopost.APIError(err)
		log.Printf("subscription needed — upgrade at %s", apiErr.UpgradeURL())
	case fopost.IsRateLimited(err):
		apiErr, _ := fopost.APIError(err)
		log.Printf("rate limited, retry in %s", apiErr.RetryAfter)
	default:
		log.Printf("api error: %v", err)
	}
}
```

| Status | Predicate            | Meaning                                        |
| ------ | -------------------- | ---------------------------------------------- |
| 401    | `IsUnauthorized`     | missing, invalid, or expired key                |
| 402    | `IsPaymentRequired`  | no active subscription, or credits exhausted    |
| 403    | `IsForbidden`        | valid key, but no scope or workspace access     |
| 404    | `IsNotFound`         | no such resource, or outside the key's reach    |
| 409    | `IsConflict`         | the resource's state forbids the change         |
| 413    | `IsPayloadTooLarge`  | a file over the account's ceiling               |
| 429    | `IsRateLimited`      | over the plan's per-minute ceiling              |

`StatusOf` and `CodeOf` read the same fields off any error, and
`(*Error).Field` decodes the extra context some responses carry.

## Services

| Service       | Covers                                                                                                                                                            |
| ------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Posts`       | `List`, `Each`, `ListAll`, `Get`, `Create`, `Update`, `Delete`, `Duplicate`, `Publish`, `Retry`, `Cancel`, `Preflight`, `Unpublish`, `Deliveries`, `PublishRuns`, `Analytics`, `BulkShift`, `BulkLabel`, `BulkDelete`, `ValidateBulkImport`, `CommitBulkImport`, `RollbackBulkImport` |
| `Workspaces`  | `List`, `Get`, `Create`, `Update`, `Delete`, `Analytics`                                                                                                          |
| `Accounts`    | `List`, `ListWithParams`, `Get`, `Create`, `Rename`, `Move`, `Delete`, `SetPrimary`, `Validate`, `Health`, `HealthSummary`, `RefreshToken`, `Analytics`, `CreateTelegramConnectCode`, `GetTelegramConnectStatus`, `GetTelegramBotCommands`, `SetTelegramBotCommands`, `DeleteTelegramBotCommands`, `ListSlackChannels`, `ListSlackMembers`, `GetSlackIdentity`, `UpdateSlackIdentity`, `GetIceBreakers`, `SetIceBreakers`, `DeleteIceBreakers`, `GetPersistentMenu`, `SetPersistentMenu`, `DeletePersistentMenu`, `GetGreeting`, `SetGreeting`, `DeleteGreeting`, `GetWebhookSubscription`, `ResubscribeWebhook`, `ListDiscordChannels`, `SwitchDiscordChannel`, `GetDiscordIdentity`, `UpdateDiscordIdentity`, `ListDiscordPins`, `DeleteDiscordMessage`, `PinDiscordMessage`, `UnpinDiscordMessage`, `CrosspostDiscordMessage`, `CreateDiscordThread`, `SendDiscordDM`, `ListDiscordEvents`, `GetDiscordEvent`, `CreateDiscordEvent`, `UpdateDiscordEvent`, `DeleteDiscordEvent`, `ListDiscordMembers`, `GetDiscordMember`, `ListDiscordRoles`, `CreateDiscordRole`, `UpdateDiscordRole`, `DeleteDiscordRole`, `AddDiscordMemberRole`, `RemoveDiscordMemberRole`, `PlatformMetrics`, `ListRedditSubreddits`, `ListRedditSubredditRules`, `ListRedditFlairs`, `SetRedditDefaultSubreddit`, `FollowXUser`, `UnfollowXUser` |
| `AccountGroups` | `List`, `Get`, `Create`, `Update`, `Delete`, `SetMembers`                                                                                                       |
| `Communities` | `List`, `Sync`, `Search`, `Add`, `Remove`                                                                                                                         |
| `Labels`      | `List`, `Get`, `Create`, `Update`, `Delete`                                                                                                                       |
| `Webhooks`    | `List`, `Create`, `Update`, `Delete`, `Test`, `ListDeliveries`, `GetDelivery`, `Redeliver`                                                                        |
| `Analytics`   | `Overview`, `TimeSeries`, `TopPosts`, `Labels`, `PostsTable`, `PostingStreak`, `Demographics`, `Collect`, `Decay`, `Frequency`, `Timeline`, `Changes`, `CollectPost`, `NativePosts`, `Inbox` |
| `Automations` | `List`, `Get`, `Create`, `Update`, `Delete`, `Toggle`, `Runs`, `Run`, `Trigger`, `Stats`, `ListVersions`, `GetVersion`, `RestoreVersion`                           |
| `Media`       | `List`, `Limits`, `Upload`, `Presign`, `Complete`, `UploadDirect`, `Delete`                                                                                               |
| `Inbox`       | `List`, `Threads`, `Conversations`, `UnreadCount`, `Accounts`, `Platforms`, `MarkThreadRead`, `Refresh`, `Update`, `EditComment`, `Reply`, `ReplyWith`, `Hide`, `Unhide`, `Delete`, `Like`, `Unlike`, `Repost`, `Unrepost`, `Bookmark`, `Unbookmark`, `Vote`, `Pin`, `Unpin`, `React`, `StartConversation`, `SetTyping`, `Handover`, `ListApprovals`, `ApproveReply`, `RejectReply`, `GetXPost`, `SearchXPosts` |
| `Contacts`    | `List`, `Get`, `Create`, `Update`, `Delete`, `Conversations`, `Import`, `ListFields`, `CreateField`, `UpdateField`, `DeleteField`, `ConversationAnalytics` |
| `Broadcasts`  | `List`, `Get`, `Create`, `Update`, `Delete`, `Send`, `Cancel`, `Recipients` |
| `Sequences`   | `List`, `Get`, `Create`, `Update`, `Delete`, `Enroll`, `Unenroll`, `Enrollments` |
| `Ads`         | `List`, `External`, `Boostable`, `Connections`, `Sources`, `AuthorizeMeta`, `DeleteConnection`, `Boost`, `Create`, `Refresh`, `SetStatus`, `Delete`, `Audiences`, `CreateAudience`, `SearchTargeting`, `LeadForms`, `CreateLeadForm`, `Leads`, `Tree`, `CreateCampaign`, `Campaign`, `UpdateCampaign`, `DeleteCampaign`, `DuplicateCampaign`, `CreateAdSet`, `AdSet`, `UpdateAdSet`, `DeleteAdSet`, `DuplicateAdSet`, `CreateNetworkAd`, `NetworkAd`, `UpdateNetworkAd`, `DeleteNetworkAd`, `DuplicateNetworkAd`, `SetStatuses`, `Creatives`, `CreateCreative`, `Creative`, `DeleteCreative`, `Audience`, `UpdateAudience`, `DeleteAudience`, `AddAudienceUsers`, `EstimateReach`, `Insights`, `AdInsights`, `LeadForm`, `ArchiveLeadForm`, `LeadsFeed`, `LeadPages`, `SubscribeLeadPage`, `UnsubscribeLeadPage`, `Goals`, `Catalogs`, `CreateCatalog`, `Catalog`, `UpdateCatalog`, `DeleteCatalog`, `CatalogProducts`, `WriteCatalogProducts`, `ProductFeeds`, `CreateProductFeed`, `DeleteProductFeed`, `FeedUploads`, `StartFeedUpload`, `ProductSets`, `CreateProductSet`, `UpdateProductSet`, `DeleteProductSet`, `ReachFrequency`, `CreateReachFrequency`, `ReachFrequencyPredictionByID`, `ReserveReachFrequency`, `CancelReachFrequency`, `Library`, `PartnershipCreators`, `RequestPartnership`, `RevokePartnership`, `AccountActivity`, `Labels`, `CreateLabel`, `UpdateLabel`, `DeleteLabel`, `ApplyLabel`, `Studies`, `CreateStudy`, `Study`, `DeleteStudy`, `IosCampaignLimits`, `HighDemandPeriods`, `CreateHighDemandPeriod`, `DeleteHighDemandPeriod`, `ValueRuleSets`, `CreateValueRuleSet`, `DeleteValueRuleSet` |
| `Knowledge`   | `List`, `Create`, `Update`, `Delete`, `Sync`, `Search`                                                                                                           |
| `Inbox`       | `List`, `Threads`, `Conversations`, `UnreadCount`, `Accounts`, `Platforms`, `MarkThreadRead`, `Refresh`, `Update`, `EditComment`, `Reply`, `ReplyWith`, `Hide`, `Unhide`, `Delete`, `Like`, `Unlike`, `Repost`, `Unrepost`, `Bookmark`, `Unbookmark`, `Vote`, `Pin`, `Unpin`, `React`, `StartConversation`, `SetTyping`, `ListApprovals`, `ApproveReply`, `RejectReply`, `GetXPost`, `SearchXPosts` |
| `Ads`         | `List`, `External`, `Boostable`, `Connections`, `Sources`, `Providers`, `Authorize`, `DeleteConnection`, `Boost`, `Create`, `Refresh`, `SetStatus`, `Delete`, `Audiences`, `CreateAudience`, `SearchTargeting`, `LeadForms`, `CreateLeadForm`, `Leads`, `Tree`, `CreateCampaign`, `Campaign`, `UpdateCampaign`, `DeleteCampaign`, `DuplicateCampaign`, `CreateAdSet`, `AdSet`, `UpdateAdSet`, `DeleteAdSet`, `DuplicateAdSet`, `CreateNetworkAd`, `NetworkAd`, `UpdateNetworkAd`, `DeleteNetworkAd`, `DuplicateNetworkAd`, `SetStatuses`, `Creatives`, `CreateCreative`, `Creative`, `DeleteCreative`, `Audience`, `UpdateAudience`, `DeleteAudience`, `AddAudienceUsers`, `AddAudienceCompanies`, `EstimateReach`, `Insights`, `AdInsights`, `LeadForm`, `ArchiveLeadForm`, `LeadsFeed`, `LeadPages`, `SubscribeLeadPage`, `UnsubscribeLeadPage`, `BidPricing`, `SupplyForecast`, `ConversionRules`, `CreateConversionRule`, `ConversionRule`, `UpdateConversionRule`, `DeleteConversionRule`, `AttachConversionRule`, `DetachConversionRule`, `ConversionMetrics`, `SendConversionEvents`, `SendConversions`, `ListConversionEvents`, `ListPixels`, `CreatePixel`, `GetPixel`, `UpdatePixel`, `SharePixel`, `GetPixelStats`, `GetEventMatchQuality`, `ListCustomConversions`, `CreateCustomConversion`, `GetCustomConversion`, `UpdateCustomConversion`, `DeleteCustomConversion` |
| `GoogleBusiness` | `GetLocation`, `UpdateLocation`, `GetAttributes`, `UpdateAttributes`, `GetMenus`, `ReplaceMenus`, `GetServices`, `ReplaceServices`, `ListMedia`, `AddMedia`, `DeleteMedia`, `ListPlaceActions`, `CreatePlaceAction`, `UpdatePlaceAction`, `DeletePlaceAction`, `GetVerificationOptions`, `StartVerification`, `CompleteVerification`, `GetPerformance`, `GetSearchKeywords`, `Assign` |
| `Validate`    | `Post`, `Length`, `Media`, `Subreddit` |
| `Activity`    | `List`                                                                                                                                                            |
| `Blog`        | `GetSettings`, `UpdateSettings`, `RotateWebsiteSecret`, `TestWebsite`, `ListKeywords`, `SuggestSeeds`, `PlanKeywords`, `AddKeyword`, `UpdateKeyword`, `DeleteKeyword`, `ListArticles`, `RequestArticle`, `GetArticle`, `UpdateArticle`, `RegenerateArticle`, `SubmitArticle`, `PublishArticle`, `RejectArticle`, `DeleteArticle` |
| `Blogs`       | `ListBlogs`, `ListArticles`, `GetArticle`, `CreateArticle`, `UpdateArticle`, `DeleteArticle`, `ListProducts`, `UpdateProduct`                            |
| `Queue`       | `ListSlots`, `CreateSlot`, `UpdateSlot`, `DeleteSlot`, `Next`, `Preview`                                                                                          |
| `ConnectSessions` | `Create`, `Get`                                                                                                                                               |
| `Logs`        | `List`                                                                                                                                                            |

`Activity.List` reads what happened in a workspace, newest first.

On X an inbox item can also be reposted and bookmarked (`Inbox.Repost`, `Unrepost`, `Bookmark`,
`Unbookmark`, where `CanRepost` and `CanBookmark` say so), an account can follow a user by
numeric id (`Accounts.FollowXUser`, `UnfollowXUser`), and a post can be read or searched with
the account's own token (`Inbox.GetXPost`, `Inbox.SearchXPosts`, paged by `Meta.NextToken`).
Reads need `inbox`; repost, bookmark and follow need `publish` as well. Every one of these is an
X API call metered to the workspace, and a `402` with code `platform_balance_spent` means X's own
pay-per-use balance is empty. An X account connected before these landed answers `409`
`reconnect_required` until it is reconnected, because the grant now asks for the bookmark and
follow permissions.
`ActivityKindSecurity` is the audit log: members joining, leaving or changing
role and access, and changes to two-step verification, passkeys, single sign-on
and signed-in devices. Those rows are append-only and never expire.

```go
page, err := client.Activity.List(ctx, &fopost.ListActivityParams{
    WorkspaceID: workspaceID,
    Kind:        fopost.ActivityKindSecurity,
})
for _, event := range page.Events {
    fmt.Printf("%s %s: %s\n", event.Time, event.Actor.Name, event.Summary)
}
// page.NextCursor is empty at the end of the list.
```

For an endpoint the SDK does not wrap yet, `Do` sends an authenticated request
and decodes the body as it came:

```go
var body map[string]any
err := client.Do(ctx, "GET", "/platforms", nil, nil, &body)
```

## Broadcasts and sequences

A broadcast is one message into every conversation you already have with a segment of your contacts; a sequence is a series of them on a delay. Neither opens a cold DM.

Nothing is sent into a closed messaging window: Messenger and Instagram take a business-initiated message only within 24 hours of the contact's last one, so recipients outside it come back skipped with `window_closed` rather than attempted. Telegram, Slack, Bluesky and Reddit have no window. The number sent is therefore often lower than the audience, and that is correct rather than a failure.

Reading needs the `inbox` scope; `Send`, `Cancel`, `Enroll` and `Unenroll` also need `publish`.

```go
broadcast, err := client.Broadcasts.Create(ctx, &fopost.CreateBroadcastRequest{
    WorkspaceID: workspaceID,
    AccountID:   accountID,
    Name:        "September check-in",
    Text:        "New colours just landed. Want a look?",
    Audience:    &fopost.AudienceFilter{Platforms: []string{"instagram"}},
})

// Recipients is how many contacts matched, not how many will be messaged.
sent, err := client.Broadcasts.Send(ctx, broadcast.ID)

// Who was skipped, and why.
page, err := client.Broadcasts.Recipients(ctx, broadcast.ID, &fopost.ListRecipientsParams{
    Status: fopost.RecipientSkipped,
})
for _, r := range page.Data {
    fmt.Printf("%s: %s\n", r.DisplayName, r.SkipReason)
}

sequence, err := client.Sequences.Create(ctx, &fopost.CreateSequenceRequest{
    WorkspaceID: workspaceID,
    AccountID:   accountID,
    Name:        "Welcome",
    Steps: []fopost.SequenceStep{
        {DelayHours: 0, Text: "Thanks for the follow — anything I can help with?"},
        {DelayHours: 48, Text: "Here is what people usually ask us first."},
    },
})

_, err = client.Sequences.Enroll(ctx, sequence.ID, &fopost.EnrollRequest{
    ContactIDs: []string{contactID},
})
// Nothing further fires for them.
_, err = client.Sequences.Unenroll(ctx, sequence.ID, []string{contactID})
```

## Blog

`Blog` plans keywords, writes articles and publishes them to a website. A post
goes to a social network; an article goes to a website. Reads, settings and
`SubmitArticle` need `posts`; `SuggestSeeds`, `PlanKeywords`, `RequestArticle` and
`RegenerateArticle` spend credits and need `ai` (an empty balance is a `402`, see
`UpgradeURL`); `PublishArticle` needs `publish` and answers `409` `no_target` until
the settings name a target.

`PlanKeywords` lands its results as `suggested` keywords; setting a keyword's
status to `planned` picks it and gives it the next free date. A written article is
a `draft` until `SubmitArticle` moves it to `in_review`; only then do
`PublishArticle` and `PublishAt` accept it.

```go
article, err := client.Blog.RequestArticle(ctx, workspaceID, &fopost.RequestBlogArticleRequest{
    Keyword:      "trail running shoes",
    Instructions: "Compare three models and close with a buying checklist",
})
// article.Status is "queued"; poll client.Blog.GetArticle until it is "draft".
_, err = client.Blog.SubmitArticle(ctx, workspaceID, article.ID)
// article.Status is now "in_review"; PublishArticle approves it.
```

With `TargetKind` set to `fopost.BlogTargetWebsite` and a `WebsiteURL`, every
published article is POSTed to that URL as a signed `article.published`
delivery. `UpdateSettings` returns the signing secret once, when the URL is
first set. `VerifyWebhook` checks the signature over the raw bytes, in constant
time, and refuses a timestamp more than five minutes off:

```go
func handle(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)
    envelope, err := fopost.VerifyWebhook(body, r.Header, os.Getenv("FOPOST_WEBHOOK_SECRET"))
    if err != nil {
        http.Error(w, "invalid signature", http.StatusUnauthorized)
        return
    }
    if envelope.Event == fopost.EventArticlePublished {
        var article fopost.BlogArticlePublished
        _ = envelope.Decode(&article)
        // article.BodyHTML, article.Images (URLs valid for six hours; store your own copy)
    }
    w.WriteHeader(http.StatusOK)
}
```

The same helper verifies every FoPost webhook, not only the blog ones.

## Scopes and limits

Requests send `X-API-Key`. A key carries only the scopes granted when it was
created: `posts` (which also covers publishing, deliveries, media, and `Validate`),
`workspaces`, `accounts` (which also covers `AccountGroups`), `labels`, `webhooks`, `analytics`, `automations`, `inbox`
(which also covers `Knowledge`), `ads`, `ai`. `Ads.Boost`, `Ads.Create`, `Ads.SetStatus`, `Ads.Delete`, `Ads.SetStatuses`
and the create, update, delete and duplicate calls on campaigns, ad sets and
network ads spend money and need `publish` as well as `ads`; a boost, ad or new
campaign object starts paused unless `Paused` is `fopost.Bool(false)`. A key may also be bound to a single workspace, in which case
naming any other one returns `403`.

Mutating endpoints require an active subscription and return `403` with code
`subscription_required` without one. Rate limits are per key, per minute, and
every response carries `X-RateLimit-Limit`, `X-RateLimit-Remaining` and
`X-RateLimit-Reset` — the SDK surfaces them on `*Error`.

## Example

[`examples/create-post`](examples/create-post/main.go) creates a post against a
running API:

```bash
export FOPOST_API_KEY=fp_...
go run ./examples/create-post "Hello from the Go SDK" --publish
```

## Chatbots and the inbox

The [chat adapter](https://fopost.com/docs/sdks/chat-adapter) turns the FoPost inbox into one send/receive channel for a chatbot
framework. It ships in the TypeScript and Python SDKs. There is no dedicated adapter here and no
API change behind it, so the same loop is three pieces with this client:

1. **Verify** the `inbox.message_received` webhook. The payload is ids only, on purpose, so
   nothing a customer wrote sits in your logs. The [signing scheme](https://fopost.com/docs/webhooks/verification)
   is HMAC-SHA256 over `{timestamp}.{body}`, refused past a five minute tolerance.
2. **Read** the item back with `client.Inbox.List(ctx, …)`, filtered to the payload's
   `accountId` and matched on its `itemId`.
3. **Answer** with `client.Inbox.Reply(ctx, item.ID, text)`, or open a thread with
   `client.Inbox.StartConversation(ctx, …)`.

Reading needs the `inbox` scope; answering needs `publish` as well.

## Contributing

Issues and pull requests are welcome at
[fopost/fopost-go](https://github.com/fopost/fopost-go).

```bash
go build ./...
go vet ./...
go test -race ./...
gofmt -l .
```

## License

MIT. See [LICENSE](LICENSE).

Questions or a problem: [fopost.com/contact](https://fopost.com/contact).

### Google Ads

Campaigns, ad groups, ads, audiences and insights are on `client.Ads` and dispatch by
connection. What only Google has is on `client.GoogleAds`:

```go
scope := fopost.GoogleScope{ConnectionID: "c4d5e6f7-…", CustomerID: "1234567890"}

keywords, err := client.GoogleAds.Keywords(ctx, scope, nil)

id, err := client.GoogleAds.CreateKeyword(ctx, &fopost.CreateGoogleKeywordRequest{
    GoogleScope: fopost.GoogleScope{
        WorkspaceID:  "7d2b8c11-…",
        ConnectionID: "c4d5e6f7-…",
        CustomerID:   "1234567890",
    },
    AdGroupID: "1234567890~adGroup~77",
    Text:      "running shoes",
    MatchType: fopost.GoogleMatchExact,
})
```

Also `KeywordIdeas`, `KeywordMetrics`, `SearchTerms`, `BidStrategies`, `AdSchedule` and
`SetAdSchedule`, the negative keyword lists, `Assets` and `AssetGroups`,
`LocalServicesLeads`, the conversion methods, and `Query` for a raw read-only GAQL
SELECT. Changes need the `publish` scope as well as `ads`; `CustomerID` has to name an
account the connection's grant reaches.
