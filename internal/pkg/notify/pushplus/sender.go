package pushplus

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	notifycontent "github.com/damonto/sigmo/internal/pkg/notify/content"
	"github.com/damonto/sigmo/internal/pkg/settings"
)

const pushplusEndpoint = "https://www.pushplus.plus/send"

// Sender delivers notifications through PushPlus using the instance token.
type Sender struct {
	client *http.Client
	token  string
}

type payload struct {
	Token    string `json:"token"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Template string `json:"template"`
}

func New(channel *settings.Channel) (*Sender, error) {
	token := strings.TrimSpace(channel.BotToken)
	if token == "" {
		return nil, errors.New("pushplus token is required")
	}
	return &Sender{
		client: &http.Client{Timeout: 10 * time.Second},
		token:  token,
	}, nil
}

func (s *Sender) Send(ctx context.Context, msg notifycontent.Message) error {
	body, err := json.Marshal(payload{Token: s.token, Title: msg.Title(), Content: msg.Text(), Template: "html"})
	if err != nil {
		return fmt.Errorf("encoding pushplus message: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, pushplusEndpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("building pushplus request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("sending pushplus message: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("pushplus response status %s: %s", resp.Status, strings.TrimSpace(string(responseBody)))
	}
	return nil
}
