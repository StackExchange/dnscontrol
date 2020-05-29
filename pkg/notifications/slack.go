package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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
	URL string
}

func (s *slackNotifier) Notify(domain, provider, msg string, err error, preview bool) {
	var payload struct {
		Username string `json:"username"`
		Text     string `json:"text"`
	}
	payload.Username = "DNSControl"

	if preview {
		payload.Text = fmt.Sprintf(`**Preview: %s[%s] -** %s`, domain, provider, msg)
	} else if err != nil {
		payload.Text = fmt.Sprintf(`**ERROR running correction on %s[%s] -** (%s) Error: %s`, domain, provider, msg, err)
	} else {
		payload.Text = fmt.Sprintf(`Successfully ran correction for **%s[%s]** - %s`, domain, provider, msg)
	}

	json, _ := json.Marshal(payload)
	http.Post(s.URL, "text/json", bytes.NewReader(json))
}

func (s *slackNotifier) Done() {}
