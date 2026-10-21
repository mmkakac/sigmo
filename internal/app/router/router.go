package router

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	"github.com/damonto/sigmo/apps/web"
	"github.com/damonto/sigmo/internal/app/auth"
	"github.com/damonto/sigmo/internal/app/buildinfo"
	appconnectivity "github.com/damonto/sigmo/internal/app/connectivity"
	"github.com/damonto/sigmo/internal/app/forwarder"
	"github.com/damonto/sigmo/internal/app/handler/appinfo"
	hauth "github.com/damonto/sigmo/internal/app/handler/auth"
	"github.com/damonto/sigmo/internal/app/handler/capability"
	"github.com/damonto/sigmo/internal/app/handler/communication"
	"github.com/damonto/sigmo/internal/app/handler/esim"
	"github.com/damonto/sigmo/internal/app/handler/euicc"
	hinternet "github.com/damonto/sigmo/internal/app/handler/internet"
	"github.com/damonto/sigmo/internal/app/handler/message"
	hmodem "github.com/damonto/sigmo/internal/app/handler/modem"
	hnativepush "github.com/damonto/sigmo/internal/app/handler/nativepush"
	"github.com/damonto/sigmo/internal/app/handler/network"
	"github.com/damonto/sigmo/internal/app/handler/notification"
	hsettings "github.com/damonto/sigmo/internal/app/handler/settings"
	"github.com/damonto/sigmo/internal/app/handler/simapp"
	"github.com/damonto/sigmo/internal/app/handler/ussd"
	appmiddleware "github.com/damonto/sigmo/internal/app/middleware"
	"github.com/damonto/sigmo/internal/app/modemstatus"
	pinternet "github.com/damonto/sigmo/internal/pkg/internet"
	"github.com/damonto/sigmo/internal/pkg/lpa"
	pmessage "github.com/damonto/sigmo/internal/pkg/message"
	"github.com/damonto/sigmo/internal/pkg/modem"
	"github.com/damonto/sigmo/internal/pkg/reminder"
	"github.com/damonto/sigmo/internal/pkg/settings"
	"github.com/damonto/sigmo/internal/pkg/storage"
	pussd "github.com/damonto/sigmo/internal/pkg/ussd"
)

type Extension func(*echo.Group, RegisterConfig) error

type RegisterConfig struct {
	// ResourceRoutes is for short-lived capability URLs validated by their owner.
	// Actual requests retain the Access boundary; OPTIONS returns only CORS metadata.
	ResourceRoutes      *echo.Group
	Build               buildinfo.Info
	Store               *settings.Store
	Registry            *modem.Registry
	InternetConnector   *pinternet.Connector
	LPAClients          *lpa.Pool
	InternetConnections appconnectivity.InternetConnections
	Relay               *forwarder.Relay
	Network             *network.Handler
	Storage             *storage.Store
	WebPush             webPushClient
	Reminders           *reminder.Scheduler
	MessageRoute        pmessage.Route
	USSDRoute           pussd.Route
	ModemOverview       []modemstatus.Extension
	Features            []string
	Extensions          []Extension
	MCP                 mcpController
	MCPKeys             mcpKeyStore
	Updates             updateController
	LicenseRoutes       func(*echo.Group)
}

func Register(e *echo.Echo, deps RegisterConfig) error {
	if deps.Network == nil {
		return errors.New("configure network routes: network handler is required")
	}
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Filesystem: web.Root(),
		Index:      "index.html",
		HTML5:      true,
		Skipper: func(c *echo.Context) bool {
			path := c.Request().URL.Path
			// In minimal builds mcpEndpointPath is empty; an empty prefix would
			// match every path and skip the static file server, so only apply
			// the MCP skip when the endpoint is actually registered.
			return strings.HasPrefix(path, "/api/") ||
				(mcpEndpointPath != "" && (path == mcpEndpointPath || strings.HasPrefix(path, mcpEndpointPath+"/")))
		},
	}))

	authStore, err := auth.NewStore(deps.Storage)
	if err != nil {
		return fmt.Errorf("configure auth store: %w", err)
	}

	v1 := e.Group("/api/v1", appmiddleware.AccessBoundary(authStore, deps.Store))
	deps.ResourceRoutes = e.Group("/api/v1", appmiddleware.ResourceAccessBoundary(authStore, deps.Store))
	if deps.LicenseRoutes != nil {
		deps.LicenseRoutes(v1)
	}

	appInfoHandler := appinfo.New(deps.Build)
	v1.GET("/app", appInfoHandler.Get)

	capabilityHandler := capability.New(append(slices.Clone(deps.Features), "contacts", "messageThreads", "historyAssignments"))
	v1.GET("/capabilities", capabilityHandler.List)

	authHandler := hauth.New(deps.Store, authStore)
	if deps.Relay != nil {
		authHandler.SetOTPSMSTransport(func(ctx context.Context, modemID, to, text string) error {
			return deps.Relay.SendSMS(ctx, modemID, to, text)
		})
	}
	v1.GET("/auth/methods", authHandler.Methods)
	v1.POST("/auth/otp/challenges", authHandler.CreateOTPChallenge)
	v1.POST("/auth/sessions", authHandler.CreateSession)
	v1.POST("/auth/session-refreshes", authHandler.RefreshSession)
	v1.POST("/auth/browser-sessions", authHandler.CreateBrowserSession)
	v1.GET("/auth/browser-sessions/:id", authHandler.StartBrowserSession)
	v1.GET("/auth/oidc/callback", authHandler.OIDCCallback)
	protected := v1.Group("")
	protected.Use(appmiddleware.Auth(authStore, deps.Store), appmiddleware.Locale(deps.Store))
	protected.GET("/auth/sessions/current", authHandler.CurrentSession)
	protected.DELETE("/auth/sessions/current", authHandler.DeleteSession)
	{
		{
			h := hnativepush.New(deps.Storage)
			protected.GET("/push-subscriptions", h.List)
			protected.POST("/push-subscriptions", h.Register)
			protected.PATCH("/push-subscriptions/:subscriptionId", h.Update)
			protected.DELETE("/push-subscriptions/:subscriptionId", h.Delete)
		}
		{
			h := hsettings.New(deps.Store, deps.InternetConnector, deps.Relay)
			protected.GET("/settings", h.Get)
			protected.POST("/settings/auth-tests", h.TestAuth)
			protected.PUT("/settings/auth", h.UpdateAuth)
			protected.PUT("/settings/proxy", h.UpdateProxy)
			protected.PUT("/settings/channels/:channel", h.UpdateNotificationChannel)
			protected.PUT("/settings/notifications", h.UpdateNotifications)
		}

		h := hmodem.New(hmodem.Config{
			Store: deps.Store, Storage: deps.Storage, Registry: deps.Registry, Internet: deps.InternetConnector,
			Reminders: deps.Reminders, LPA: deps.LPAClients, OverviewExtensions: deps.ModemOverview,
		})
		protected.GET("/modems", h.List)
		protected.GET("/modems/:id", h.Get)
		protected.POST("/modems/:id/sim-unlocks", h.UnlockSIM)
		protected.GET("/modems/:id/sim-pin", h.SIMPINStatus)
		protected.PUT("/modems/:id/sim-pin/protection", h.SetSIMPINProtection)
		protected.POST("/modems/:id/sim-pin-changes", h.ChangeSIMPIN)
		protected.PUT("/modems/:id/sim-slots/:slot", h.SwitchSIMSlot)
		protected.PUT("/modems/:id/msisdn", h.UpdateMSISDN)
		protected.PUT("/modems/:id/sims/:iccid/reminder", h.UpdateReminder)
		protected.DELETE("/modems/:id/sims/:iccid/reminder", h.DeleteReminder)
		protected.GET("/modems/:id/settings", h.Settings)
		protected.PUT("/modems/:id/settings", h.UpdateSettings)

		{
			h := simapp.New(deps.Registry)
			protected.GET("/modems/:id/sim-application/sessions", h.Session)
		}

		{
			h := message.New(deps.Registry, deps.Storage, deps.MessageRoute)
			protected.GET("/conversations", h.Conversations)
			protected.GET("/conversations/:conversationId/messages", h.ConversationMessages)
			protected.PATCH("/conversations/:conversationId", h.UpdateConversation)
			protected.DELETE("/conversations/:conversationId", h.DeleteConversation)
			protected.POST("/conversations/read-all", h.MarkAllRead)
			protected.GET("/modems/:id/messages", h.List)
			protected.GET("/modems/:id/messages/:participant", h.ListByParticipant)
			protected.POST("/modems/:id/messages", h.Send)
			protected.DELETE("/modems/:id/messages/:participant", h.DeleteByParticipant)
		}

		{
			h := ussd.New(deps.Registry, deps.USSDRoute)
			protected.POST("/modems/:id/ussd", h.Execute)
		}

		{
			h := deps.Network
			protected.GET("/modems/:id/networks", h.List)
			protected.POST("/modems/:id/network-scans", h.StartNetworkScan)
			protected.GET("/modems/:id/network-scans/:scanID", h.GetNetworkScan)
			protected.GET("/modems/:id/networks/airplane-mode", h.AirplaneMode)
			protected.PUT("/modems/:id/networks/airplane-mode", h.SetAirplaneMode)
			protected.GET("/modems/:id/networks/registration", h.Registration)
			protected.PUT("/modems/:id/networks/registration", h.SetRegistration)
			protected.GET("/modems/:id/networks/modes", h.Modes)
			protected.PUT("/modems/:id/networks/current-modes", h.SetCurrentModes)
			protected.GET("/modems/:id/networks/bands", h.Bands)
			protected.PUT("/modems/:id/networks/current-bands", h.SetCurrentBands)
		}

		{
			h := hinternet.New(deps.Registry, deps.InternetConnections)
			protected.GET("/modems/:id/internet-connections/current", h.Current)
			protected.GET("/modems/:id/internet-connections/public", h.Public)
			protected.POST("/modems/:id/internet-connections", h.Connect)
			protected.PUT("/modems/:id/internet-connections/current/preferences", h.UpdatePreferences)
			protected.DELETE("/modems/:id/internet-connections/current", h.Disconnect)
		}

		{
			h := euicc.New(deps.Registry, deps.LPAClients)
			protected.GET("/modems/:id/ses", h.Get)
		}

		{
			h := esim.New(esim.Config{
				Storage:   deps.Storage,
				Store:     deps.Store,
				Registry:  deps.Registry,
				LPA:       deps.LPAClients,
				Internet:  deps.InternetConnector,
				Reminders: deps.Reminders,
			})
			communication.New(communication.Config{Storage: deps.Storage, Registry: deps.Registry, ESIM: h, CallsEnabled: slices.Contains(deps.Features, "nativeCalls")}).Register(protected)
			protected.GET("/modems/:id/esims", h.List)
			protected.POST("/modems/:id/ses/:seId/esim-discoveries", h.Discovery)
			protected.GET("/modems/:id/esim-downloads/sessions", h.Download)
			protected.PUT("/modems/:id/ses/:seId/esims/:iccid/activation", h.Enable)
			protected.PUT("/modems/:id/ses/:seId/esims/:iccid/nickname", h.UpdateNickname)
			protected.DELETE("/modems/:id/ses/:seId/esims/:iccid", h.Delete)
			protected.PUT("/modems/:id/ses/:seId/esims/:iccid/reminder", h.UpdateReminder)
			protected.DELETE("/modems/:id/ses/:seId/esims/:iccid/reminder", h.DeleteReminder)
		}

		{
			h := notification.New(deps.Registry, deps.LPAClients)
			protected.GET("/modems/:id/notifications", h.List)
			protected.POST("/modems/:id/ses/:seId/notifications/:sequence/deliveries", h.Resend)
			protected.DELETE("/modems/:id/ses/:seId/notifications/:sequence", h.Delete)
		}

		if err := registerFeatureRoutes(e, protected, authStore, deps); err != nil {
			return err
		}

		for _, extension := range deps.Extensions {
			if err := extension(protected, deps); err != nil {
				return err
			}
		}
	}
	return nil
}
