package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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
	payload.ParseMode = "HTML"

	if preview {
		payload.Text = fmt.Sprintf("DNSControl preview: %s[%s]:\n%s", domain, provider, msg)
	} else if err != nil {
		payload.Text = fmt.Sprintf("DNSControl ERROR running correction on %s[%s]:\n%s\nError: %s", domain, provider, msg, err)
	} else {
		payload.Text = fmt.Sprintf("DNSControl successfully ran correction for %s[%s]:\n%s", domain, provider, msg)
	}

	marshaledPayload, _ := json.Marshal(payload)

	_, posterr := http.Post(url, "application/json", bytes.NewBuffer(marshaledPayload))
	return posterr
}

func (s *telegramNotifier) Done() {}
