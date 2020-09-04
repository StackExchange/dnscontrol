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
		url, ok := cfg["teams_url"]
		if !ok {
			return nil
		}

		notifier := &teamsNotifier{
			URL: url,
		}
		return notifier
	})
}

// teamsNotifier sends notifications to teams or mattermost
type teamsNotifier struct {
	URL string
}

func (s *teamsNotifier) Notify(domain, provider, msg string, err error, preview bool) {
	var payload struct {
		Username string `json:"username"`
		Text     string `json:"text"`
	}
	payload.Username = "DnsControl"

	// Format changes as 'preformated' text
	msg = strings.ReplaceAll(msg, "\n", "\n    ")

	if preview {
		payload.Text = fmt.Sprintf("**DnsControl Preview %s**\n%s", domain, msg)
	} else if err != nil {
		payload.Text = fmt.Sprintf("**DnsControl Error Making Changes %s**\n%s\nError: %s", domain, msg, err)
	} else {
		payload.Text = fmt.Sprintf("**DnsControl Successfully Changed %s**\n%s", domain, msg)
	}

	json, _ := json.Marshal(payload)
	http.Post(s.URL, "text/json", bytes.NewReader(json))
}

func (s *teamsNotifier) Done() {}
