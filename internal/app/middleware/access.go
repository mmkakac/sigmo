//go:build !minimal

package middleware

import (
	"github.com/labstack/echo/v5"

	"github.com/damonto/sigmo/internal/app/auth"
	"github.com/damonto/sigmo/internal/pkg/settings"
	"github.com/damonto/sigmo/internal/pkg/storage"
)

// verifyAccessSession authenticates a signed Cloudflare Access edge assertion.
// It is compiled only into non-minimal builds.
func verifyAccessSession(store *auth.Store, cfg settings.Auth, c *echo.Context) (storage.AuthSession, error) {
	return store.VerifyAccess(c.Request().Context(), auth.AccessRequest{Policy: cfg, Assertion: c.Request().Header.Get("Cf-Access-Jwt-Assertion")})
}
