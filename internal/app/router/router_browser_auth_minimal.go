//go:build minimal

package router

import (
	"github.com/labstack/echo/v5"

	hauth "github.com/damonto/sigmo/internal/app/handler/auth"
)

// registerBrowserAuth is a no-op in minimal builds: OIDC and Cloudflare Access
// login are not compiled, so the corresponding routes are never registered.
func registerBrowserAuth(group *echo.Group, authHandler *hauth.Handler) {}
