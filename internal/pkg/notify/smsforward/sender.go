// Package smsforward forwards notification content as an SMS message to one or
// more configured phone numbers. Unlike the other channels it needs a modem to
// transmit through, so the relay attaches a transport before sending; without
// an attached transport Send reports a clear error instead of silently dropping.
package smsforward

import (
	"context"
	"errors"
	"fmt"
	"strings"

	notifycontent "github.com/damonto/sigmo/internal/pkg/notify/content"
	notifyevent "github.com/damonto/sigmo/internal/pkg/notify/event"
	"github.com/damonto/sigmo/internal/pkg/settings"
)

// SendFunc transmits one SMS text to a destination through the modem that
// carried the original event.
type SendFunc func(ctx context.Context, modemID, to, text string) error

// Sender forwards incoming SMS notifications as SMS to configured numbers.
type Sender struct {
	numbers []string
	send    SendFunc
}

// New creates an SMS-forward sender from a channel's configured recipients.
func New(channel *settings.Channel) (*Sender, error) {
	if channel == nil {
		return nil, errors.New("nil sms channel")
	}
	return &Sender{numbers: channel.Recipients.Strings()}, nil
}

// Attach installs the modem transport used to deliver forwarded SMS. It is set
// by the relay, which owns the messenger and modem registry.
func (s *Sender) Attach(send SendFunc) {
	s.send = send
}

// Send forwards an incoming SMS event to every configured number. Events of
// other kinds are ignored so that a call or reminder is never sent as SMS.
func (s *Sender) Send(ctx context.Context, msg notifycontent.Message) error {
	if len(s.numbers) == 0 {
		return nil
	}
	event, ok := msg.Event.(notifyevent.SMSEvent)
	if !ok {
		return nil
	}
	if !event.Incoming || strings.TrimSpace(event.Text) == "" {
		return nil
	}
	if s.send == nil {
		return errors.New("sms channel has no modem transport attached")
	}
	text := msg.Text()
	var combined error
	for _, number := range s.numbers {
		if err := s.send(ctx, event.ModemID, number, text); err != nil {
			combined = errors.Join(combined, fmt.Errorf("forward to %s: %w", number, err))
		}
	}
	return combined
}
