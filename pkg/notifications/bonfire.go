package notifications

import (
	"fmt"
	"net/http"
	"strings"
)

func init() {
	initers = append(initers, func(cfg map[string]string) Notifier {
		if url, ok := cfg["bonfire_url"]; ok {
			return bonfireNotifier(url)
		}
		return nil
	})
}

// bonfire notifier for stack exchange internal chat. String is just url with room and token in it
type bonfireNotifier string

func (b bonfireNotifier) Notify(domain, provider, msg string, err error, preview bool) {
	var payload string
	if preview {
		payload = fmt.Sprintf(`**Preview: %s[%s] -** %s`, domain, provider, msg)
	} else if err != nil {
		payload = fmt.Sprintf(`**ERROR running correction on %s[%s] -** (%s) Error: %s`, domain, provider, msg, err)
	} else {
		payload = fmt.Sprintf(`Successfully ran correction for **%s[%s]** - %s`, domain, provider, msg)
	}
	// chat doesn't markdownify multiline messages. Split in two so the first line can have markdown
	parts := strings.SplitN(payload, "\n", 2)
	for _, p := range parts {
		http.Post(string(b), "text/markdown", strings.NewReader(p))
	}
}

func (b bonfireNotifier) Done() {}
