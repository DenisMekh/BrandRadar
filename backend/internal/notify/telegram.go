package notify

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Notifier отправляет текстовые уведомления во внешний канал.
type Notifier interface {
	Notify(ctx context.Context, message string) error
}

// TelegramNotifier отправляет сообщения в Telegram-канал через Bot API.
type TelegramNotifier struct {
	botToken string
	chatID   string
	client   *http.Client
}

// NewTelegramNotifier создаёт TelegramNotifier.
// Если botToken или chatID пусты — возвращает NopNotifier.
func NewTelegramNotifier(botToken, chatID string) Notifier {
	if botToken == "" || chatID == "" {
		return NopNotifier{}
	}
	return &TelegramNotifier{
		botToken: botToken,
		chatID:   chatID,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (t *TelegramNotifier) Notify(ctx context.Context, message string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.botToken)

	params := url.Values{}
	params.Set("chat_id", t.chatID)
	params.Set("text", message)
	params.Set("parse_mode", "HTML")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL+"?"+params.Encode(), nil)
	if err != nil {
		return fmt.Errorf("telegram notify: build request: %w", err)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("telegram notify: send: %w", err)
	}
	defer func(Body io.ReadCloser) {
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram notify: unexpected status %d", resp.StatusCode)
	}

	return nil
}

// NopNotifier — пустая реализация, когда Telegram не сконфигурирован.
type NopNotifier struct{}

func (NopNotifier) Notify(_ context.Context, _ string) error { return nil }
