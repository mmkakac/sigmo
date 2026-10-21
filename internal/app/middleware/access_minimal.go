//go:build minimal

package middleware

import (
	"github.com/labstack/echo/v5"

	"github.com/damonto/sigmo/internal/app/auth"
	"github.com/damonto/sigmo/internal/pkg/settings"
	"github.com/damonto/sigmo/internal/pkg/storage"
)

// verifyAccessSession is unavailable in minimal builds: Cloudflare Access
// cannot be enabled there, so any Access-bound request is rejected.
func verifyAccessSession(store *auth.Store, cfg settings.Auth, c *echo.Context) (storage.AuthSession, error) {
	return storage.AuthSession{}, storage.ErrInvalidSession
}
