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

// RetryPolicy controls automatic re-sending of a forwarded SMS when no delivery
// confirmation arrives within a timeout. The relay owns the re-send loop; a nil
// policy passed to SendFunc disables re-sending (used for OTP codes, so only
// forwarded incoming SMS are retried).
type RetryPolicy struct {
	// TimeoutMinutes is how long to wait for a delivery confirmation before the
	// first re-send is triggered.
	TimeoutMinutes int
	// MaxRetries is how many times the SMS is re-sent when no confirmation
	// arrives.
	MaxRetries int
	// IntervalMinutes spaces consecutive re-sends.
	IntervalMinutes int
}

// Normalize fills safe defaults for a policy built from channel settings.
func (p RetryPolicy) Normalize() RetryPolicy {
	if p.TimeoutMinutes <= 0 {
		p.TimeoutMinutes = 30
	}
	if p.MaxRetries < 1 {
		p.MaxRetries = 1
	}
	if p.IntervalMinutes <= 0 {
		p.IntervalMinutes = 5
	}
	return p
}

// SendFunc transmits one SMS text to a destination. The non-nil retry policy is
// honoured only for forwarded incoming SMS; callers pass nil for events such as
// OTP codes that must never be re-sent automatically.
type SendFunc func(ctx context.Context, modemID, to, text string, retry *RetryPolicy) error

// Sender forwards incoming SMS notifications as SMS to configured numbers.
type Sender struct {
	numbers []string
	retry   *RetryPolicy
	send    SendFunc
}

// New creates an SMS-forward sender from a channel's configured recipients and
// delivery-retry settings.
func New(channel *settings.Channel) (*Sender, error) {
	if channel == nil {
		return nil, errors.New("nil sms channel")
	}
	rp := RetryPolicy{
		TimeoutMinutes:  channel.DeliveryTimeoutMinutes,
		MaxRetries:      channel.MaxDeliveryRetries,
		IntervalMinutes: channel.DeliveryRetryMinutes,
	}.Normalize()
	return &Sender{numbers: channel.Recipients.Strings(), retry: &rp}, nil
}

// Attach installs the modem transport used to deliver forwarded SMS. It is set
// by the relay, which owns the messenger and modem registry.
func (s *Sender) Attach(send SendFunc) {
	s.send = send
}

// Send forwards an incoming SMS event to every configured number. OTP events
// are also forwarded so the "sms" channel can deliver login verification codes,
// but OTP re-sends are never scheduled automatically. Events of other kinds are
// ignored so that a call or reminder is never sent as SMS.
func (s *Sender) Send(ctx context.Context, msg notifycontent.Message) error {
	if len(s.numbers) == 0 {
		// Fail loudly instead of reporting success while nothing is transmitted,
		// which previously hid an OTP code that was never delivered.
		return errors.New("sms channel has no forwarding number configured")
	}
	var text string
	var retry *RetryPolicy
	switch ev := msg.Event.(type) {
	case notifyevent.SMSEvent:
		if !ev.Incoming || strings.TrimSpace(ev.Text) == "" {
			return nil
		}
		text = msg.Text()
		retry = s.retry
	case notifyevent.OTPEvent:
		text = msg.Text()
		retry = nil
	default:
		return nil
	}
	if s.send == nil {
		return errors.New("sms channel has no modem transport attached")
	}
	var combined error
	for _, number := range s.numbers {
		if err := s.send(ctx, smsEventModemID(msg), number, text, retry); err != nil {
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
