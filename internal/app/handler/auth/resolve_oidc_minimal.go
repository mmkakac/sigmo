//go:build minimal

package auth

import (
	"context"

	"github.com/damonto/sigmo/internal/app/auth"
	"github.com/damonto/sigmo/internal/pkg/settings"
	"github.com/damonto/sigmo/internal/pkg/storage"
)

// resolveOIDCHandoff is unavailable in minimal builds because OIDC login is not
// compiled there, so any OIDC handoff attempt is rejected.
func resolveOIDCHandoff(ctx context.Context, store *auth.Store, cfg settings.Auth, code, verifier string) (storage.AuthSession, error) {
	return storage.AuthSession{}, auth.ErrInvalidLogin
}
