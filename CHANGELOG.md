# Changelog

## Unreleased

- `Media.Limits` reads `GET /media/limits`: the account's upload ceiling, the multipart per-file ceiling, the next tier and the allowed types.
- `Posts.Create` and `Posts.Update` take `PostFile`s. Files under the multipart ceiling travel with the post in one request; larger ones go through `UploadDirect` and are attached to the first content block. `OpenPostFile` opens a path.
- `PresignInput` gains `FolderID`, `Origin`, `BrandRole`, `BrandVoiceID` and `PersonConsent`; `PresignedUpload` gains `MaxFileBytes`. The 50 MB presign ceiling is gone; sizes go up to the account's `MaxFileBytes`.
- `IsPayloadTooLarge` and `(*Error).MaxFileBytes()` read a `413 file_too_large`.
- **Breaking:** `Media.UploadDirect(ctx, *PresignInput, io.Reader)` streams the file instead of taking a `[]byte`.
