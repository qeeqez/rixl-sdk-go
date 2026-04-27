package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	kiotahttp "github.com/microsoft/kiota-http-go"
	"github.com/qeeqez/rixl-sdk-go/examples/internal/exauth"
	"github.com/qeeqez/rixl-sdk-go/examples/internal/exenv"
	"github.com/qeeqez/rixl-sdk-go/sdk"
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	page, err := client.Images().Get(ctx, nil)
	if err != nil {
		log.Fatalf("list images: %v", err)
	}
	fmt.Printf("Listed %d images\n", len(page.GetData()))
	for _, img := range page.GetData() {
		if img.GetId() != nil {
			fmt.Printf("  - %s\n", *img.GetId())
		}
	}

	if id := os.Getenv("IMAGE_ID"); id != "" {
		image, err := client.Images().ByImageId(id).Get(ctx, nil)
		if err != nil {
			log.Fatalf("get image %s: %v", id, err)
		}
		fmt.Printf("Image %s: %dx%d\n",
			exenv.Deref(image.GetId()), exenv.Int32Or(image.GetWidth()), exenv.Int32Or(image.GetHeight()))
	}
}
