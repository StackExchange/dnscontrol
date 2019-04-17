package notifications

// Notifier is a type that can send a notification
type Notifier interface {
	// Notify will be called after a correction is performed.
	// It will be given the correction's message, the result of executing it,
	// and a flag for whether this is a preview or if it actually ran.
	// If preview is true, err will always be nil.
	Notify(domain, provider string, message string, err error, preview bool)
	// Done will be called exactly once after all notifications are done. This will allow "batched" notifiers to flush and send
	Done()
}

// new notification types should add themselves to this array
var initers = []func(map[string]string) Notifier{}

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

func (m multiNotifier) Notify(domain, provider string, message string, err error, preview bool) {
	for _, n := range m {
		n.Notify(domain, provider, message, err, preview)
	}
}
func (m multiNotifier) Done() {
	for _, n := range m {
		n.Done()
	}
}
