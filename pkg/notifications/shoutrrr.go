package notifications

import (
	"fmt"

	"github.com/containrrr/shoutrrr"
)

func init() {
	initers = append(initers, func(cfg map[string]string) Notifier {
		if url, ok := cfg["shoutrrr_url"]; ok {
			return shoutrrrNotifier(url)
		}
		return nil
	})
}

type shoutrrrNotifier string

func (b shoutrrrNotifier) Notify(domain, provider, msg string, err error, preview bool) {
	var payload string
	if preview {
		payload = fmt.Sprintf("DNSControl preview: %s[%s]:\n%s", domain, provider, msg)
	} else if err != nil {
		payload = fmt.Sprintf("DNSControl ERROR running correction on %s[%s]:\n%s\nError: %s", domain, provider, msg, err)
	} else {
		payload = fmt.Sprintf("DNSControl successfully ran correction for %s[%s]:\n%s", domain, provider, msg)
	}
	shoutrrr.Send(string(b), payload)
}

func (b shoutrrrNotifier) Done() {}
