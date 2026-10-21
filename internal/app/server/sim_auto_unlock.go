package server

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	mmodem "github.com/damonto/sigmo/internal/pkg/modem"
	"github.com/damonto/sigmo/internal/pkg/settings"
)

// autoUnlockSIMs watches the modem registry and unlocks any SIM whose saved
// PIN can be reused, whenever the modem reports a locked SIM. It only acts on
// modems whose auto-unlock preference is enabled, so a wrong saved PIN can
// never cause an automatic lockout on a modem the user did not opt in.
func autoUnlockSIMs(ctx context.Context, store *settings.Store, registry *mmodem.Registry) error {
	unsubscribe, err := registry.Subscribe(ctx, func(event mmodem.ModemEvent) error {
		switch event.Type {
		case mmodem.ModemEventAdded, mmodem.ModemEventChanged, mmodem.ModemEventSIMChanged:
		default:
			return nil
		}
		modem := event.Modem
		if modem == nil {
			return nil
		}
		snapshot := modem.Snapshot()
		if !snapshot.Locked() {
			return nil
		}
		modemID := modem.EquipmentIdentifier
		prefs := store.FindModem(modemID)
		if !prefs.AutoUnlockSIM || strings.TrimSpace(prefs.SIMPin) == "" {
			return nil
		}
		if err := modem.UnlockSIM(ctx, mmodem.SIMUnlock{PIN: prefs.SIMPin}); err != nil {
			if errors.Is(err, mmodem.ErrSIMUnlockNotRequired) {
				return nil
			}
			slog.Warn("auto-unlock SIM failed", "imei", modemID, "error", err)
			return nil
		}
		slog.Info("auto-unlocked SIM", "imei", modemID)
		return nil
	})
	if err != nil {
		return err
	}
	defer func() { unsubscribe() }()
	<-ctx.Done()
	return nil
}
