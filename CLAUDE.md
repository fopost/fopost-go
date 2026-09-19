# CLAUDE.md

Guidance for Claude Code (claude.ai/code) when working in this repository.

## What This Is

Go module `github.com/fopost/fopost-go` — the official Go client for the FoPost REST API
(`fopost.com`). `fopost.Version` is `0.3.0`. It wraps the API's HTTP surface in services
hung off `*Client`: `Posts`, `Workspaces`, `Accounts`, `Communities`, `Labels`, `Webhooks`,
`Analytics`, `Automations`, `Media`, `Inbox`, `Ads`, `Validate`.

`go 1.22` minimum (the code uses the `min` builtin, so 1.21+ is required regardless).
**Standard library only** — `go.mod` has no `require` block.

This is the most complete FoPost SDK by resource coverage; treat it as the reference when
another client is missing an endpoint.

## Downstream Packages

These repos wrap this SDK and must be updated in lockstep:

- `fopost-cli` — command-line client for the FoPost API
- `fopost-terraform-provider` — Terraform provider for FoPost resources

**Whenever you change this SDK's public surface — a renamed method, a changed parameter,
a new or removed resource, a new error type, a bumped minimum language version — you must
open a matching PR in every repo listed above in the same session.** They are separate
git repos, checked out as siblings at `../fopost-<child>`. A parent release that silently
breaks a child is only discovered by the user who upgrades first.

Also bump the child's dependency constraint on this package and note the change in its
CHANGELOG when this package is released.

## Brand Rules

- The product is **FoPost** (`fopost.com`). Never write "OwlStack" — retired Aug 2026.
- Never write an email address. Support is https://fopost.com/contact and GitHub issues.
- Never name AI providers/models, infrastructure vendors, or any person.
- Never type a platform count in doc comments or the README.

## Architecture

One flat package `fopost` at the repository root, one file per API group:

| File | Contents |
| :--- | :--- |
| `client.go` | `Client`, `Option`s, the retry loop, decode, envelope unwrap, rate-limit and `Retry-After` parsing |
| `errors.go` | `*Error`, `RateLimit`, `APIError`/`StatusOf`/`CodeOf`, `Is*` predicates |
| `types.go` | `Time`, `PageMeta`, `ContentBlock`, `Text`/`Thread`, `queryBuilder`, `Bool`/`String`/`Int` |
| `multipart.go` | `buildMultipart` — media upload and CSV bulk import bodies |
| `posts.go` `accounts.go` `workspaces.go` `communities.go` `labels.go` `webhooks.go` `analytics.go` `automations.go` `media.go` `inbox.go` `ads.go` `validate.go` | one `*Service` each, with its request/response types |
| `internal/version/main.go` | prints `fopost.Version` so the release workflow can check it against the tag |

Request flow: a service method builds its query with `newQuery()` and its body as a struct
or map, then calls `s.client.json(ctx, method, path, body, query, out)` →
`Client.do()` (retry loop) → `Client.decode()` → `unwrapEnvelope()` → `json.Unmarshal`.

`Client.json` sets `unwrap: true`; the public `Client.Do` does **not** unwrap, so an
escape-hatch caller sees the raw envelope. Every method takes `ctx` first. `Client` is safe
for concurrent use.

## API Contract

- Base URL: `DefaultBaseURL = "https://api.fopost.com/v1"`, overridden by the
  `FOPOST_BASE_URL` environment variable in `New`, then by `WithBaseURL`. (The doc comment
  on `WithBaseURL` mentions the env var, but the fallback actually lives in `New`.)
- Auth: header `X-API-Key`. An empty key passed to `New` falls back to `FOPOST_API_KEY`;
  missing on both returns an error.
- Headers sent: `Accept: application/json`, `X-API-Key`, `User-Agent: fopost-go/<Version>`
  (`WithUserAgent` prepends the caller's own), plus `Content-Type` when there is a body.
- Timeout: `DefaultTimeout = 30 * time.Second` on the `http.Client`; `WithTimeout` changes
  it, `WithHTTPClient` replaces the client entirely and wins over `WithTimeout`.
- **Retries: `DefaultMaxRetries = 3` total attempts (two retries).** Retried on HTTP 429,
  HTTP >= 500, and transport errors. Backoff is `500ms * 2^(attempt-1)` capped at 60s;
  on a response carrying `Retry-After` that value wins, also capped at 60s. `Retry-After`
  is read as delta-seconds or an HTTP date. **A cancelled context never retries** — `do`
  checks `ctx.Err()` before treating a transport error as a blip, and `sleep` returns
  `ctx.Err()`. `WithMaxRetries(1)` disables retrying.
- Success envelope: `unwrapEnvelope` peels `{"data": ...}` when a `data` key is present,
  otherwise passes the body through. Paginated lists decode `meta` into `PageMeta`
  (`current_page`, `per_page`, `total`, `last_page`, `from`, `to`); the inbox lists carry
  the camelCase `InboxPageMeta` (`page`, `perPage`, `total`) instead.
- Scopes: one per service, named after it. `Validate` needs `posts`; `Inbox` needs `inbox`; `Ads` needs `ads`, and
  `Ads.Boost`/`Create`/`SetStatus`/`Delete` also need `publish` because they spend money.
  A boost or ad starts paused unless `Paused` is `Bool(false)`. `/inbox/chat/*` (browser-
  encrypted X Chat) and the attachment stream `/inbox/{id}/attachments/{index}` are not
  wrapped; the SDK has no binary-download pattern.
- Error envelope: `{"error": "<code>", "message": "<text>"}` becomes a single `*Error` with
  `Status`, `Code`, `Message`, the raw `Body`, `RetryAfter`, and `RateLimit`. **There is no
  error subclass hierarchy** — the brief's per-status types are expressed as predicates:
  `IsUnauthorized`, `IsPaymentRequired`, `IsForbidden`, `IsNotFound`, `IsConflict`,
  `IsRateLimited`, plus `APIError(err)`, `StatusOf(err)`, `CodeOf(err)`.
  `(*Error).UpgradeURL()` reads a 402's `upgrade_url`; `(*Error).Field(name, out)` decodes
  any other body field.
- Rate-limit headers: `X-RateLimit-Limit`, `-Remaining`, `-Reset` are parsed into
  `Error.RateLimit`. **They are only populated on error responses** — `errorFromResponse`
  is the sole caller of `rateLimitFrom`, so a successful call does not expose the budget.
  Note this before documenting otherwise.
- Escape hatch: `client.Do(ctx, method, path, body, query, out)`.

## Commands

```bash
go build ./...
go vet ./...
go test ./...
go test -race -coverprofile=coverage.txt ./...   # what CI runs
gofmt -l .                                       # must print nothing
go run ./examples/create-post "Hello" --publish  # needs FOPOST_API_KEY
go run ./internal/version                        # prints fopost.Version
```

`.github/workflows/ci.yml` runs the gofmt check, `go vet`, `go build`, and
`go test -race` on Go `1.22` and `stable`, on push to `main`, on PRs, and on dispatch.

## Conventions

- `gofmt` is enforced by CI — a listed file fails the build. There is no golangci-lint.
- Doc comments on every exported symbol, starting with the identifier. Short `why`
  comments elsewhere; no narrated docblocks.
- `ctx context.Context` is the first parameter of every service method.
- Configuration is functional options (`Option func(*Client)`), each defensive about zero
  values (`WithMaxRetries` clamps below 1 to 1, `WithBaseURL` ignores empty).
- Optional scalars use the `Bool`/`String`/`Int` pointer helpers in `types.go`; query
  building goes through `queryBuilder` (`str`, `num`, `boolPtr`, `time`), which drops zero
  values so they are never sent.
- `Time` wraps `time.Time` with a tolerant `UnmarshalJSON` (several API formats, plus null)
  and marshals RFC 3339. Use it for API timestamps, not bare `time.Time`.
- Content helpers: `Text(s)` for one block, `Thread(a, b, ...)` for several.
- Errors are wrapped with `fmt.Errorf("fopost: ...: %w", err)` so `errors.As` reaches
  `*Error`.

## Testing

Plain `testing` plus `net/http/httptest` — no test framework, no external dependency.
**Tests never hit the live API.** `client_test.go` defines the `testClient` helper: it
stands up an `httptest.Server`, points the client at it with `WithBaseURL(server.URL)`, and
sets `WithMaxRetries(1)` so nothing ever waits on a real backoff. Tests that exercise
retries pass a larger `WithMaxRetries` and a handler that counts attempts with
`sync/atomic`.

Coverage today: transport (auth header, user agent, query omission, envelope unwrap, path
escaping, escape hatch, 429/5xx retry, no retry on 4xx, context cancellation, non-JSON
bodies), errors, posts (pagination, `Each`, bulk, multipart import), the other resources
(including inbox filters, snake_case read body, reply, approvals, and ads camelCase body,
workspace query, audiences, lead cursor), and `types.go` time/helpers. Add to the matching
`*_test.go` rather than a new file.

## Releasing

Go modules have no registry upload: **tag `v<version>` and the module proxy serves it.**
`.github/workflows/release.yml` gates the tag on a green build — it runs `go vet` and
`go test -race`, verifies the tag equals `fopost.Version` via `go run ./internal/version`,
then curls `proxy.golang.org` so `pkg.go.dev` indexes the version straight away (a failure
there is only a warning).

**No repo secrets are required**, and none are referenced. Bumping a release means editing
the `Version` constant in `client.go` and tagging the same value.

## Git

Conventional Commits (`<type>(<scope>): <description>`), atomic — one logical change per
commit. Branch `feature/<description>` off a fresh `main`, merge via PR.
Never `gh pr create` — push the branch and hand over the compare link
(`https://github.com/fopost/fopost-go/compare/main...<branch>`).
