//go:build !minimal

package router

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/damonto/sigmo/internal/app/auth"
	"github.com/damonto/sigmo/internal/app/handler/mcpadmin"
	hupdate "github.com/damonto/sigmo/internal/app/handler/update"
	hwebpush "github.com/damonto/sigmo/internal/app/handler/webpush"
	"github.com/damonto/sigmo/internal/app/mcpauth"
	"github.com/damonto/sigmo/internal/app/mcpserver"
	appmiddleware "github.com/damonto/sigmo/internal/app/middleware"
	appupdate "github.com/damonto/sigmo/internal/app/update"
	"github.com/damonto/sigmo/internal/pkg/webpush"
)

// Aliases referenced by the common RegisterConfig. They exist only in feature
// builds; the minimal build defines the same names as inert types.
type (
	webPushClient    = *webpush.Client
	mcpController    = *mcpserver.Controller
	mcpKeyStore      = *mcpauth.Store
	updateController = *appupdate.Controller
)

const mcpEndpointPath = mcpserver.EndpointPath

// registerFeatureRoutes installs the optional MCP, update, and web push routes.
func registerFeatureRoutes(e *echo.Echo, protected *echo.Group, authStore *auth.Store, deps RegisterConfig) error {
	if deps.MCP != nil {
		e.Match(
			[]string{http.MethodGet, http.MethodPost, http.MethodDelete},
			mcpEndpointPath,
			echo.WrapHandler(deps.MCP.Handler()),
			appmiddleware.AccessBoundary(authStore, deps.Store),
		)
	}
	if deps.Updates != nil {
		h := hupdate.New(deps.Updates)
		protected.GET("/settings/updates", h.GetSettings)
		protected.PUT("/settings/updates", h.UpdateSettings)
		protected.POST("/update-checks", h.CreateCheck)
		protected.POST("/update-installations", h.CreateInstallation)
		protected.GET("/update-installations/current", h.CurrentInstallation)
	}
	if deps.MCP != nil && deps.MCPKeys != nil {
		h := mcpadmin.New(mcpadmin.Config{Settings: deps.Store, Keys: deps.MCPKeys, Storage: deps.Storage, Controller: deps.MCP})
		protected.GET("/settings/mcp", h.GetSettings)
		protected.PUT("/settings/mcp", h.UpdateSettings)
		protected.GET("/settings/mcp/api-keys", h.ListAPIKeys)
		protected.POST("/settings/mcp/api-keys", h.CreateAPIKey)
		protected.DELETE("/settings/mcp/api-keys/:id", h.RevokeAPIKey)
		protected.GET("/settings/mcp/audit-events", h.ListAuditEvents)
		protected.GET("/settings/mcp/skills/sigmo-control", h.DownloadSkill)
	}
	if deps.WebPush != nil {
		h := hwebpush.New(deps.WebPush)
		protected.GET("/web-push", h.Get)
		protected.PUT("/web-push", h.Update)
		protected.GET("/web-push/subscriptions", h.ListSubscriptions)
		protected.POST("/web-push/subscriptions", h.Register)
		protected.PATCH("/web-push/subscriptions/:subscriptionId", h.Rename)
		protected.DELETE("/web-push/subscriptions/:subscriptionId", h.Delete)
	}
	return nil
}
