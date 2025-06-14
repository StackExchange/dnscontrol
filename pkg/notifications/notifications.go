package notifications

import "regexp"

// Notifier is a type that can send a notification
type Notifier interface {
	// Notify will be called after a correction is performed.
	// It will be given the correction's message, the result of executing it,
	// and a flag for whether this is a preview or if it actually ran.
	// If preview is true, err will always be nil.
	Notify(domain, provider string, message string, err error, preview bool) error
	// Done will be called exactly once after all notifications are done. This will allow "batched" notifiers to flush and send
	Done()
}

// new notification types should add themselves to this array
var initers = []func(map[string]string) Notifier{}

// matches ansi color codes
var ansiColorRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// Init will take the given config map (from creds.json notifications key) and create a single Notifier with
// all notifications it has full config for.
func Init(config map[string]string) Notifier {
	notifiers := multiNotifier{}
	for _, i := range initers {
		n := i(config)
		if n != nil {
			notifiers = append(notifiers, n)
		}
	}
	return notifiers
}

type multiNotifier []Notifier

// removes any ansi color codes from a given string
func stripAnsiColors(colored string) string {
	return ansiColorRegex.ReplaceAllString(colored, "")
}

func (m multiNotifier) Notify(domain, provider string, message string, err error, preview bool) error {
	// force-remove ansi colors that might come with the message from dnscontrol.
	// These usually don't render well in notifiers, outputting escape codes.
	// If a notifier wants to output colors, they should probably implement
	// them natively.
	nMsg := stripAnsiColors(message)
	errs := make([]error, 0)
	for _, n := range m {
		err := n.Notify(domain, provider, nMsg, err, preview)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

func (m multiNotifier) Done() {
	for _, n := range m {
		n.Done()
	}
}
