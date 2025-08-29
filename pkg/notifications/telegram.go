package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"io"
)

func init() {
	initers = append(initers, func(cfg map[string]string) Notifier {
		if botToken, ok := cfg["telegram_bot_token"]; ok {
			if chatID, ok := cfg["telegram_chat_id"]; ok {
				notifier := &telegramNotifier{
					BotToken: botToken,
					ChatID:   chatID,
				}
				return notifier
			}
		}
		return nil
	})
}

// telegramNotifier sends notifications to Telegram
type telegramNotifier struct {
	BotToken string
	ChatID   string
}

func (s *telegramNotifier) Notify(domain, provider, msg string, err error, preview bool) error {
	var payload struct {
		ChatID    int64  `json:"chat_id"`
		Text      string `json:"text"`
		ParseMode string `json:"parse_mode"`
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.BotToken)

	payload.ChatID, _ = strconv.ParseInt(s.ChatID, 10, 64)
	payload.ParseMode = "MarkdownV2"

	if preview {
		payload.Text = fmt.Sprintf("DNSControl preview: %s[%s]:\n%s", domain, provider, msg)
	} else if err != nil {
		payload.Text = fmt.Sprintf("DNSControl ERROR running correction on %s[%s]:\n%s\nError: %s", domain, provider, msg, err)
	} else {
		payload.Text = fmt.Sprintf("DNSControl successfully ran correction for %s[%s]:\n%s", domain, provider, msg)
	}

// Debugging version
marshaledPayload, err := json.Marshal(payload)
if err != nil {
    return fmt.Errorf("failed to marshal telegram payload: %w", err)
}

resp, posterr := http.Post(url, "application/json", bytes.NewBuffer(marshaledPayload))
if posterr != nil {
    return posterr
}
defer resp.Body.Close()

// Check if Telegram returned an error (anything other than a 200 OK)
if resp.StatusCode != http.StatusOK {
    body, _ := io.ReadAll(resp.Body) // You'll need to import "io"
    return fmt.Errorf("telegram API error: %s", string(body))
}

return nil
}

func (s *telegramNotifier) Done() {}
