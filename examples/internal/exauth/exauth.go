// Package exauth provides AuthenticationProvider implementations that
// skip Kiota's stock HTTPS-only check, so the examples can run against
// a local plaintext server.
package exauth

import (
	"context"

	abs "github.com/microsoft/kiota-abstractions-go"
)

type APIKey struct{ Key string }

func (a *APIKey) AuthenticateRequest(_ context.Context, req *abs.RequestInformation, _ map[string]any) error {
	req.Headers.Add("X-API-Key", a.Key)
	return nil
}

type Bearer struct{ Token string }

func (b *Bearer) AuthenticateRequest(_ context.Context, req *abs.RequestInformation, _ map[string]any) error {
	req.Headers.Add("Authorization", "Bearer "+b.Token)
	return nil
}
