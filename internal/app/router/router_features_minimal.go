//go:build minimal

package router

import (
	"github.com/labstack/echo/v5"

	"github.com/damonto/sigmo/internal/app/auth"
)

// Minimal aliases keep the common RegisterConfig compiling without the MCP,
// update, and web push packages. The MCP endpoint path is empty so the static
// file skipper never special-cases it.
type (
	webPushClient    interface{}
	mcpController    interface{}
	mcpKeyStore      interface{}
	updateController interface{}
)

const mcpEndpointPath = ""

// registerFeatureRoutes installs no optional routes in minimal builds.
func registerFeatureRoutes(e *echo.Echo, protected *echo.Group, authStore *auth.Store, deps RegisterConfig) error {
	return nil
}
