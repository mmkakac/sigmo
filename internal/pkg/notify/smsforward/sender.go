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

// Send forwards an incoming SMS event to every configured number. OTP events
// are also forwarded so the "sms" channel can deliver login verification codes.
// Events of other kinds are ignored so that a call or reminder is never sent
// as SMS.
func (s *Sender) Send(ctx context.Context, msg notifycontent.Message) error {
	if len(s.numbers) == 0 {
		// Fail loudly instead of reporting success while nothing is transmitted,
		// which previously hid an OTP code that was never delivered.
		return errors.New("sms channel has no forwarding number configured")
	}
	var text string
	switch ev := msg.Event.(type) {
	case notifyevent.SMSEvent:
		if !ev.Incoming || strings.TrimSpace(ev.Text) == "" {
			return nil
		}
		text = msg.Text()
	case notifyevent.OTPEvent:
		text = msg.Text()
	default:
		return nil
	}
	if s.send == nil {
		return errors.New("sms channel has no modem transport attached")
	}
	var combined error
	for _, number := range s.numbers {
		if err := s.send(ctx, smsEventModemID(msg), number, text); err != nil {
			combined = errors.Join(combined, fmt.Errorf("forward to %s: %w", number, err))
		}
	}
	return combined
}

// smsEventModemID returns the modem that carried the event, or "" for events
// (such as OTP) that have no originating modem so the transport can fall back
// to any available modem.
func smsEventModemID(msg notifycontent.Message) string {
	if ev, ok := msg.Event.(notifyevent.SMSEvent); ok {
		return ev.ModemID
	}
	return ""
}
