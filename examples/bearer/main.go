package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	kiotahttp "github.com/microsoft/kiota-http-go"
	"github.com/qeeqez/rixl-sdk-go/examples/internal/exauth"
	"github.com/qeeqez/rixl-sdk-go/examples/internal/exenv"
	"github.com/qeeqez/rixl-sdk-go/sdk"
)

type tokenRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Subject      string `json:"subject"`
	ProjectID    string `json:"project_id,omitempty"`
	TTLMinutes   *int   `json:"ttl_minutes,omitempty"`
}

type tokenResponse struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresIn   int64     `json:"expires_in"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func mintToken(ctx context.Context, baseURL string, body tokenRequest) (*tokenResponse, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/clientauth/token", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("mint token: %s: %s", resp.Status, string(raw))
	}
	var out tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func main() {
	clientID := exenv.MustEnv("RIXL_CLIENT_ID")
	clientSecret := exenv.MustEnv("RIXL_CLIENT_SECRET")
	projectID := exenv.MustEnv("RIXL_PROJECT_ID")
	subject := exenv.MustEnv("RIXL_SUBJECT")
	baseURL := exenv.EnvOr("RIXL_BASE_URL", "http://localhost:8081")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tok, err := mintToken(ctx, baseURL, tokenRequest{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Subject:      subject,
		ProjectID:    projectID,
	})
	if err != nil {
		log.Fatalf("mint: %v", err)
	}
	fmt.Printf("minted token (expires_in=%ds, type=%s)\n", tok.ExpiresIn, tok.TokenType)

	adapter, err := kiotahttp.NewNetHttpRequestAdapter(&exauth.Bearer{Token: tok.AccessToken})
	if err != nil {
		log.Fatalf("adapter: %v", err)
	}
	adapter.SetBaseUrl(baseURL)
	client := sdk.NewRixlClient(adapter)

	page, err := client.Images().Get(ctx, nil)
	if err != nil {
		log.Fatalf("list images: %v", err)
	}
	fmt.Printf("Listed %d images\n", len(page.GetData()))
}
