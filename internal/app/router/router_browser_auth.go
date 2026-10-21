//go:build !minimal

package router

import (
	"github.com/labstack/echo/v5"

	hauth "github.com/damonto/sigmo/internal/app/handler/auth"
)

// registerBrowserAuth mounts the OIDC browser login routes. It is compiled only
// into non-minimal builds.
func registerBrowserAuth(group *echo.Group, authHandler *hauth.Handler) {
	group.POST("/auth/browser-sessions", authHandler.CreateBrowserSession)
	group.GET("/auth/browser-sessions/:id", authHandler.StartBrowserSession)
	group.GET("/auth/oidc/callback", authHandler.OIDCCallback)
}
