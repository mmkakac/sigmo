package modem

import (
	"context"
	"fmt"
	"strings"

	mmodem "github.com/damonto/sigmo/internal/pkg/modem"
	appsettings "github.com/damonto/sigmo/internal/pkg/settings"
)

type modemSettings struct {
	store *appsettings.Store
}

func newSettings(store *appsettings.Store) *modemSettings {
	return &modemSettings{store: store}
}

func (s *modemSettings) Update(ctx context.Context, modem *mmodem.Modem, req UpdateModemSettingsRequest) error {
	modemID := modem.EquipmentIdentifier
	current := s.store.FindModem(modemID)
	current.Alias = strings.TrimSpace(req.Alias)
	current.MSS = req.MSS
	current.AutoUnlockSIM = req.AutoUnlockSIM
	if req.AutoUnlockSIM && strings.TrimSpace(req.SIMPin) != "" {
		current.SIMPin = strings.TrimSpace(req.SIMPin)
	}
	if !req.AutoUnlockSIM {
		current.SIMPin = ""
	}
	if err := s.store.UpdateModem(ctx, modemID, current); err != nil {
		return fmt.Errorf("save modem settings: %w", err)
	}
	return nil
}

func (s *modemSettings) Get(modem *mmodem.Modem) *ModemSettingsResponse {
	current := s.store.FindModem(modem.EquipmentIdentifier)
	return &ModemSettingsResponse{
		Alias:         current.Alias,
		MSS:           current.MSS,
		AutoUnlockSIM: current.AutoUnlockSIM,
		SIMPinSaved:   strings.TrimSpace(current.SIMPin) != "",
	}
}

// SaveUnlockedPin stores a PIN after a successful manual unlock while the
// auto-unlock preference is enabled, so the modem can be unlocked
// automatically the next time it reports a locked SIM.
func (s *modemSettings) SaveUnlockedPin(ctx context.Context, modemID, pin string) error {
	pin = strings.TrimSpace(pin)
	if pin == "" {
		return nil
	}
	current := s.store.FindModem(modemID)
	if !current.AutoUnlockSIM {
		return nil
	}
	current.SIMPin = pin
	if err := s.store.UpdateModem(ctx, modemID, current); err != nil {
		return fmt.Errorf("save modem pin: %w", err)
	}
	return nil
}
