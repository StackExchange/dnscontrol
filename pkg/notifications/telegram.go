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
		if bot_token, ok := cfg["telegram_bot_token"]; ok {
			if chat_id, ok := cfg["telegram_chat_id"]; ok {
				notifier := &telegramNotifier{
					BOT_TOKEN: bot_token,
					CHAT_ID: chat_id,
				}
				return notifier
			}
		}		
		return nil
	})
}

// telegramNotifier sends notifications to Telegram
type telegramNotifier struct {
	BOT_TOKEN string
	CHAT_ID string
}

func (s *telegramNotifier) Notify(domain, provider, msg string, err error, preview bool) {
	var payload struct {
		ChatID int64 `json:"chat_id"`
		Text string `json:"text"`
	}

	var url = fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.BOT_TOKEN)

	payload.ChatID, _ = strconv.ParseInt(s.CHAT_ID, 10, 64)

	if preview {
		payload.Text = fmt.Sprintf(`DNSControl preview: %s[%s] -** %s`, domain, provider, msg)
	} else if err != nil {
		payload.Text = fmt.Sprintf(`DNSControl ERROR running correction on %s[%s] -** (%s) Error: %s`, domain, provider, msg, err)
	} else {
		payload.Text = fmt.Sprintf(`DNSControl successfully ran correction for **%s[%s]** - %s`, domain, provider, msg)
	}

	marshaled_payload, _ := json.Marshal(payload)

	_, _ = http.Post(url, "application/json", bytes.NewBuffer(marshaled_payload))

}

func (s *telegramNotifier) Done() {}
