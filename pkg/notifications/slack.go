package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func init() {
	initers = append(initers, func(cfg map[string]string) Notifier {
		if url, ok := cfg["slack_url"]; ok {
			notifier := &slackNotifier{
				URL: url,
			}
			return notifier
		}
		return nil
	})
}

// slackNotifier sends notifications to slack or mattermost
type slackNotifier struct {
	URL      string
	messages []string
}

func (s *slackNotifier) Notify(domain, provider, msg string, err error, preview bool) {
	var content string
	if preview {
		content = fmt.Sprintf(`**Preview: %s[%s] -** %s`, domain, provider, msg)
	} else if err != nil {
		content = fmt.Sprintf(`**ERROR running correction on %s[%s] -** (%s) Error: %s`, domain, provider, msg, err)
	} else {
		content = fmt.Sprintf(`Successfully ran correction for **%s[%s]** - %s`, domain, provider, msg)
	}
	s.messages = append(s.messages, content)
}

func (s *slackNotifier) Done() {
	var payload struct {
		Username string `json:"username"`
		Text     string `json:"text"`
	}
	payload.Username = "DNSControl"
	s.messages = append([]string{"New changes:"}, s.messages...)
	payload.Text = strings.Join(s.messages, "\n")

	json, _ := json.Marshal(payload)
	http.Post(s.URL, "text/json", bytes.NewReader(json))
}
