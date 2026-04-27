package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	kiotahttp "github.com/microsoft/kiota-http-go"
	"github.com/qeeqez/rixl-sdk-go/examples/internal/exauth"
	"github.com/qeeqez/rixl-sdk-go/examples/internal/exenv"
	"github.com/qeeqez/rixl-sdk-go/sdk"
	"github.com/qeeqez/rixl-sdk-go/sdk/models"
	apierr "github.com/qeeqez/rixl-sdk-go/sdk/models/github_com_qeeqez_api_internal_errors"
	vidupload "github.com/qeeqez/rixl-sdk-go/sdk/models/github_com_qeeqez_api_internal_videos_handler_upload"
	imghandler "github.com/qeeqez/rixl-sdk-go/sdk/models/internal_images_handler"
)

const (
	sampleImageURL = "https://picsum.photos/seed/rixl/800/600.jpg"
	sampleVideoURL = "https://download.samplelib.com/mp4/sample-5s.mp4"
)

func main() {
	apiKey := exenv.MustEnv("RIXL_API_KEY")
	baseURL := exenv.EnvOr("RIXL_BASE_URL", "http://localhost:8081")

	adapter, err := kiotahttp.NewNetHttpRequestAdapter(&exauth.APIKey{Key: apiKey})
	if err != nil {
		log.Fatalf("adapter: %v", err)
	}
	adapter.SetBaseUrl(baseURL)
	client := sdk.NewRixlClient(adapter)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	uploadImage(ctx, client)
	uploadVideo(ctx, client)
}

func uploadImage(ctx context.Context, c *sdk.RixlClient) {
	fmt.Println("== Image upload ==")
	body, err := download(ctx, sampleImageURL)
	if err != nil {
		log.Fatalf("download image: %v", err)
	}

	initReq := imghandler.NewUploadInitRequest()
	name, format := "sample.jpg", "jpeg"
	initReq.SetName(&name)
	initReq.SetFormat(&format)

	initRes, err := c.Images().Upload().Init().Post(ctx, initReq, nil)
	if err != nil {
		log.Fatalf("image init: %s", explain(err))
	}
	fmt.Printf("init: image_id=%s\n", exenv.Deref(initRes.GetImageId()))

	if err := putBytes(ctx, exenv.Deref(initRes.GetPresignedUrl()), body, "image/jpeg"); err != nil {
		log.Fatalf("PUT image: %v", err)
	}

	completeReq := imghandler.NewCompleteRequest()
	completeReq.SetImageId(initRes.GetImageId())
	notAttached := false
	completeReq.SetAttachedToVideo(&notAttached)

	image, err := c.Images().Upload().Complete().Post(ctx, completeReq, nil)
	if err != nil {
		log.Fatalf("image complete: %s", explain(err))
	}
	fmt.Printf("complete: id=%s %dx%d\n\n",
		exenv.Deref(image.GetId()), exenv.Int32Or(image.GetWidth()), exenv.Int32Or(image.GetHeight()))
}

func uploadVideo(ctx context.Context, c *sdk.RixlClient) {
	fmt.Println("== Video upload ==")
	video, err := download(ctx, sampleVideoURL)
	if err != nil {
		log.Fatalf("download video: %v", err)
	}
	poster, err := download(ctx, sampleImageURL)
	if err != nil {
		log.Fatalf("download poster: %v", err)
	}

	initReq := models.NewVideoUploadInitRequest()
	fileName, posterFormat := "sample.mp4", "jpeg"
	initReq.SetFileName(&fileName)
	initReq.SetImageFormat(&posterFormat)

	initRes, err := c.Videos().Upload().Init().Post(ctx, initReq, nil)
	if err != nil {
		log.Fatalf("video init: %s", explain(err))
	}
	fmt.Printf("init: video_id=%s poster_id=%s\n",
		exenv.Deref(initRes.GetVideoId()), exenv.Deref(initRes.GetPosterId()))

	if err := putBytes(ctx, exenv.Deref(initRes.GetVideoPresignedUrl()), video, "video/mp4"); err != nil {
		log.Fatalf("PUT video: %v", err)
	}
	if err := putBytes(ctx, exenv.Deref(initRes.GetPosterPresignedUrl()), poster, "image/jpeg"); err != nil {
		log.Fatalf("PUT poster: %v", err)
	}

	completeReq := vidupload.NewCompleteRequest()
	completeReq.SetVideoId(initRes.GetVideoId())

	finished, err := c.Videos().Upload().Complete().Post(ctx, completeReq, nil)
	if err != nil {
		log.Fatalf("video complete: %s", explain(err))
	}
	fmt.Printf("complete: id=%s\n", exenv.Deref(finished.GetId()))
}

func download(ctx context.Context, url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func putBytes(ctx context.Context, url string, body []byte, contentType string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.ContentLength = int64(len(body))
	req.Header.Set("Content-Type", contentType)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("PUT %s: %s: %s", url, resp.Status, string(b))
	}
	return nil
}

func explain(err error) string {
	var apiErr *apierr.ErrorResponse
	if errors.As(err, &apiErr) {
		code, msg := int32(0), ""
		if apiErr.GetCode() != nil {
			code = *apiErr.GetCode()
		}
		if apiErr.GetErrorEscaped() != nil {
			msg = *apiErr.GetErrorEscaped()
		}
		return fmt.Sprintf("HTTP %d: %s", code, msg)
	}
	return err.Error()
}
