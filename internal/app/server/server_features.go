//go:build !minimal

package server

import (
	"context"
	"fmt"

	hnetwork "github.com/damonto/sigmo/internal/app/handler/network"
	"github.com/damonto/sigmo/internal/app/forwarder"
	"github.com/damonto/sigmo/internal/app/mcpauth"
	"github.com/damonto/sigmo/internal/app/mcpserver"
	appupdate "github.com/damonto/sigmo/internal/app/update"
	"github.com/damonto/sigmo/internal/pkg/internet"
	"github.com/damonto/sigmo/internal/pkg/lpa"
	"github.com/damonto/sigmo/internal/pkg/modem"
	"github.com/damonto/sigmo/internal/pkg/nativepush"
	"github.com/damonto/sigmo/internal/pkg/reminder"
	"github.com/damonto/sigmo/internal/pkg/settings"
	"github.com/damonto/sigmo/internal/pkg/storage"
	"github.com/damonto/sigmo/internal/pkg/webpush"
)

// The aliases below let the common server/runtime/router code reference
// optional feature components without importing their packages, so a minimal
// build can drop those packages entirely.
type (
	webPushClient     = *webpush.Client
	nativePushClient  = *nativepush.Dispatcher
	updateSource      = appupdate.Source
	updateController  = *appupdate.Controller
	licenseProvider   = appupdate.LicenseProvider
	mcpToolsExtension = mcpserver.Extension
	mcpController     = *mcpserver.Controller
	mcpKeyStore       = *mcpauth.Store
)

func newWebPushClient(ctx context.Context, db *storage.Store) (webPushClient, error) {
	return webpush.New(ctx, db)
}

func newNativePushClient(db *storage.Store, store *settings.Store, authorized func() bool) (nativePushClient, error) {
	return nativepush.New(nativepush.Config{
		Storage:    db,
		Settings:   store,
		Authorized: authorized,
	})
}

func newUpdateController(cfg Config, runtime *Runtime, store *settings.Store, restart func()) (updateController, error) {
	if runtime.UpdateSource == nil {
		runtime.UpdateSource = appupdate.NewGitHubSource(nil, "", cfg.Build.ReleasePublicKey)
	}
	return appupdate.NewController(appupdate.ControllerConfig{
		Build:    cfg.Build,
		Settings: store,
		Source:   runtime.UpdateSource,
		License:  runtime.License,
		Restart:  restart,
	})
}

func newMCPController(
	cfg Config,
	runtime *Runtime,
	store *settings.Store,
	registry *modem.Registry,
	internetConnector *internet.Connector,
	lpaClients *lpa.Pool,
	relay *forwarder.Relay,
	networkHandler *hnetwork.Handler,
	db *storage.Store,
	reminderScheduler *reminder.Scheduler,
) (mcpController, mcpKeyStore, error) {
	mcpKeys, err := mcpauth.NewStore(db)
	if err != nil {
		return nil, nil, fmt.Errorf("configure MCP API keys: %w", err)
	}
	mcpCatalog := mcpserver.NewCatalog()
	if err := mcpserver.RegisterCoreTools(mcpCatalog, mcpserver.CoreToolsConfig{
		Store:               store,
		Registry:            registry,
		InternetConnector:   internetConnector,
		LPAClients:          lpaClients,
		InternetConnections: runtime.internetConnections,
		Relay:               relay,
		Network:             networkHandler,
		Storage:             db,
		Reminders:           reminderScheduler,
		MessageRoute:        runtime.messageRoute,
		USSDRoute:           runtime.ussdRoute,
		ModemOverview:       runtime.modemOverview,
	}); err != nil {
		return nil, nil, fmt.Errorf("configure MCP core tools: %w", err)
	}
	for _, extension := range runtime.mcpTools {
		if err := extension(mcpCatalog); err != nil {
			return nil, nil, fmt.Errorf("configure MCP extension: %w", err)
		}
	}
	controller, err := mcpserver.New(mcpserver.Config{
		BuildVersion: cfg.Build.Version,
		Settings:     store,
		Keys:         mcpKeys,
		Storage:      db,
		Catalog:      mcpCatalog,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("configure MCP server: %w", err)
	}
	return controller, mcpKeys, nil
}

func closeMCPController(controller mcpController) {
	if controller != nil {
		controller.Close()
	}
}

func runUpdateController(ctx context.Context, controller updateController) error {
	return controller.Run(ctx)
}

func runNativePush(ctx context.Context, client nativePushClient) error {
	return client.Run(ctx)
}

func confirmUpdateHealth(ctx context.Context, controller updateController) {
	if controller != nil {
		confirmUpdateHealthy(ctx, controller.MarkHealthy)
	}
}

func markUpdatedBinaryHealthy(ctx context.Context, executable string) {
	confirmUpdateHealthy(ctx, func() error { return appupdate.MarkHealthy(executable) })
}

// featureFlags advertises the optional capabilities present in this build so
// the shared frontend can hide settings pages whose routes do not exist.
func featureFlags() []string {
	return []string{"webPush", "mcp", "updates"}
}
