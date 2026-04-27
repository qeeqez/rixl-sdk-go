# RIXL Go SDK

The official Go client for the [RIXL](https://rixl.com) API.

[![Go Reference](https://pkg.go.dev/badge/github.com/qeeqez/rixl-sdk-go.svg)](https://pkg.go.dev/github.com/qeeqez/rixl-sdk-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/qeeqez/rixl-sdk-go)](https://goreportcard.com/report/github.com/qeeqez/rixl-sdk-go)

[Installation](#installation) • [Quick start](#quick-start) • [Authentication](#authentication) • [Resources](#resources) • [Pagination](#pagination) • [Errors](#errors)

## Features

- Typed fluent API generated from the RIXL OpenAPI spec
- `context.Context` on every request — deadlines and cancellation built in
- Pre-mapped error responses for 400, 401, 403, 404, and 500
- Pluggable `RequestAdapter` and authentication providers
- Support for JSON, form, multipart, and plain-text payloads

## Requirements

- Go 1.22+
- A RIXL API key

## Installation

```bash
go get github.com/qeeqez/rixl-sdk-go
go get github.com/microsoft/kiota-http-go
go get github.com/microsoft/kiota-abstractions-go
```

## Quick start

```go
package main

import (
    "context"
    "fmt"

    "github.com/microsoft/kiota-abstractions-go/authentication"
    kiotahttp "github.com/microsoft/kiota-http-go"
    "github.com/qeeqez/rixl-sdk-go/sdk"
)

func main() {
    auth, _ := authentication.NewApiKeyAuthenticationProvider(
        "YOUR_RIXL_API_KEY", "X-API-Key", authentication.HEADER_KEYLOCATION,
    )
    adapter, _ := kiotahttp.NewNetHttpRequestAdapter(auth)
    client := sdk.NewRixlClient(adapter)

    image, err := client.Images().ByImageId("PS5IMKoFLm").Get(context.Background(), nil)
    if err != nil {
        panic(err)
    }
    fmt.Println(*image.GetId(), *image.GetWidth(), *image.GetHeight())
}
```

Base URL defaults to `https://api.rixl.com`. Override with `adapter.SetBaseUrl("...")`.

## Authentication

```go
import "github.com/microsoft/kiota-abstractions-go/authentication"

// API key in a header
auth, _ := authentication.NewApiKeyAuthenticationProvider(
    "YOUR_RIXL_API_KEY", "X-API-Key", authentication.HEADER_KEYLOCATION,
)

// Bearer token
// Implement authentication.AccessTokenProvider, then:
// auth, _ := authentication.NewBaseBearerTokenAuthenticationProvider(tokenProvider)
```

## Resources

### Feeds

```go
posts, err := client.Feeds().ByFeedId("FD4y3QB38S").Get(ctx, nil)
for _, post := range posts.GetData() {
    fmt.Println(*post.GetId())
}
```

### Images

```go
// List
page, err := client.Images().Get(ctx, nil)

// Get
image, err := client.Images().ByImageId("PS5IMKoFLm").Get(ctx, nil)

// Delete
err = client.Images().ByImageId("PS5IMKoFLm").Delete(ctx, nil)

// Upload (init → PUT bytes to presigned URL → complete)
import (
    "bytes"
    "net/http"
    imghandler "github.com/qeeqez/rixl-sdk-go/sdk/models/internal_images_handler"
)

initReq := imghandler.NewUploadInitRequest()
name, format := "photo.jpg", "jpeg"
initReq.SetName(&name)
initReq.SetFormat(&format)

initRes, err := client.Images().Upload().Init().Post(ctx, initReq, nil)
// initRes.GetImageId(), initRes.GetPresignedUrl()

// Upload the raw bytes to the presigned URL with a standard HTTP client.
req, _ := http.NewRequestWithContext(ctx, http.MethodPut,
    *initRes.GetPresignedUrl(), bytes.NewReader(imageBytes))
req.Header.Set("Content-Type", "image/jpeg")
req.ContentLength = int64(len(imageBytes))
http.DefaultClient.Do(req)

// Finalize.
completeReq := imghandler.NewCompleteRequest()
completeReq.SetImageId(initRes.GetImageId())
attached := false
completeReq.SetAttachedToVideo(&attached)
image, err := client.Images().Upload().Complete().Post(ctx, completeReq, nil)
```

### Videos

```go
// List
videos, err := client.Videos().Get(ctx, nil)

// Get
video, err := client.Videos().ByVideoId("VI9VXQxWXQ").Get(ctx, nil)

// Subtitle tracks
tracks, err := client.Videos().ByVideoId("VI9VXQxWXQ").Subtitles().Get(ctx, nil)

// Upload (init returns presigned URLs for both the video and a poster image)
import (
    "github.com/qeeqez/rixl-sdk-go/sdk/models"
    vidupload "github.com/qeeqez/rixl-sdk-go/sdk/models/github_com_qeeqez_api_internal_videos_handler_upload"
)

initReq := models.NewVideoUploadInitRequest()
fileName, posterFormat := "clip.mp4", "jpeg"
initReq.SetFileName(&fileName)
initReq.SetImageFormat(&posterFormat)

initRes, err := client.Videos().Upload().Init().Post(ctx, initReq, nil)
// PUT video bytes to *initRes.GetVideoPresignedUrl()
// PUT poster bytes to *initRes.GetPosterPresignedUrl()

completeReq := vidupload.NewCompleteRequest()
completeReq.SetVideoId(initRes.GetVideoId())
video, err := client.Videos().Upload().Complete().Post(ctx, completeReq, nil)
```

## Pagination

List endpoints accept `limit`, `offset`, `sort`, and `order`:

```go
import (
    kabs "github.com/microsoft/kiota-abstractions-go"
    "github.com/qeeqez/rixl-sdk-go/sdk/images"
)

limit, offset := int32(50), int32(0)
for {
    config := &kabs.RequestConfiguration[images.ImagesRequestBuilderGetQueryParameters]{
        QueryParameters: &images.ImagesRequestBuilderGetQueryParameters{
            Limit: &limit, Offset: &offset,
        },
    }
    page, err := client.Images().Get(ctx, config)
    if err != nil {
        return err
    }
    for _, img := range page.GetData() {
        // ...
    }
    if offset+int32(len(page.GetData())) >= *page.GetPagination().GetTotal() {
        break
    }
    offset += limit
}
```

## Errors

API errors (400, 401, 403, 404, 500) are returned as `*ErrorResponse`:

```go
import (
    "errors"
    apierr "github.com/qeeqez/rixl-sdk-go/sdk/models/github_com_qeeqez_api_internal_errors"
)

image, err := client.Images().ByImageId("PS5IMKoFLm").Get(ctx, nil)
if err != nil {
    var apiErr *apierr.ErrorResponse
    if errors.As(err, &apiErr) {
        fmt.Printf("HTTP %d: %s\n", *apiErr.GetCode(), *apiErr.GetError())
    }
    return err
}
```

## Models

Generated types live under `github.com/qeeqez/rixl-sdk-go/sdk/models/`:

| Package | Contents |
|---------|----------|
| `models` | `Imageable`, `Videoable`, `Postable`, `Fileable` |
| `models/pagination` | `PaginatedResponseImageable`, `PaginatedResponseVideoable`, `PaginatedResponsePostable` |
| `models/internal_images_handler` | Upload request and response payloads for images |
| `models/github_com_qeeqez_api_internal_videos_handler_upload` | Upload request and response payloads for videos |
| `models/internal_videos_handler_subtitles` | Subtitle PUT payloads |
| `models/github_com_qeeqez_api_internal_errors` | `ErrorResponse` |

Getters return pointers; dereference before use.

## Context

Every request takes `context.Context` as its first argument:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

image, err := client.Images().ByImageId("PS5IMKoFLm").Get(ctx, nil)
```

## Examples

Runnable demos live in [`examples/`](./examples):

- `basic/` — list images and fetch one by ID (uses `X-API-Key`).
- `advanced/` — full image and video upload pipelines (uses `X-API-Key`).
- `bearer/` — mint a short-lived client JWT via `POST /clientauth/token`, then call with `Authorization: Bearer …`. Use this pattern when the consumer can't safely hold a long-lived API key (browser, mobile).

```bash
export RIXL_BASE_URL=http://localhost:8081  # optional; defaults to api.rixl.com
cd examples

# API key flows
export RIXL_API_KEY=<your key>
go run ./basic
go run ./advanced

# Client JWT flow
# Mint your client_id and client_secret in the RIXL dashboard
# (Organization → Client Auth → Create credential), then:
export RIXL_CLIENT_ID=<copied from the dashboard>
export RIXL_CLIENT_SECRET=<copied from the dashboard>
export RIXL_PROJECT_ID=<project ID>
export RIXL_SUBJECT=user-42
go run ./bearer
```

## Support

Open an issue at [github.com/qeeqez/rixl-sdk-go](https://github.com/qeeqez/rixl-sdk-go/issues).
