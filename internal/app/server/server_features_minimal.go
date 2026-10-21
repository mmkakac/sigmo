//go:build minimal

package server

import (
	"context"

	notifyevent "github.com/damonto/sigmo/internal/pkg/notify/event"
	"github.com/damonto/sigmo/internal/pkg/settings"
	"github.com/damonto/sigmo/internal/pkg/storage"
)

// Minimal aliases keep the common server/runtime/router code compiling while
// the MCP, update, web push, and native push packages are not compiled. Every
// feature is simply absent: constructors return nils and runners are no-ops.
type (
	webPushClient interface {
		Send(ctx context.Context, event notifyevent.Event) error
	}
	nativePushClient interface {
		Run(context.Context) error
	}
	updateSource      interface{}
	updateController  interface{}
	licenseProvider   interface{ Licensee() any }
	mcpToolsExtension interface{}
	mcpController     interface{}
	mcpKeyStore       interface{}
)

func newWebPushClient(ctx context.Context, db *storage.Store) (webPushClient, error) {
	return nil, nil
}

func newNativePushClient(db *storage.Store, store *settings.Store, authorized func() bool) (nativePushClient, error) {
	return nil, nil
}

func newUpdateController(cfg Config, runtime *Runtime, store *settings.Store, restart func()) (updateController, error) {
	return nil, nil
}

func newMCPController(
	cfg Config,
	runtime *Runtime,
	store *settings.Store,
	registry any,
	internetConnector any,
	lpaClients any,
	relay any,
	networkHandler any,
	db *storage.Store,
	reminderScheduler any,
) (mcpController, mcpKeyStore, error) {
	return nil, nil, nil
}

func closeMCPController(controller mcpController) {}

func runUpdateController(ctx context.Context, controller updateController) error {
	return nil
}

func runNativePush(ctx context.Context, client nativePushClient) error {
	return nil
}

func confirmUpdateHealth(ctx context.Context, controller updateController) {}

func markUpdatedBinaryHealthy(ctx context.Context, executable string) {}

// featureFlags is empty in minimal builds: MCP, update, and web push routes
// are not compiled, so the frontend hides those settings pages.
func featureFlags() []string {
	return nil
}
