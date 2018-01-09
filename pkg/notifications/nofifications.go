package notifications

// Notifier is a function that will be run after a correction is performed.
// It will be given the correction's message, the result of executing it,
// and a flag for whether this is a preview or if it actually ran.
// If preview is true, err will always be nil.
type Notifier func(domain, provider string, message string, err error, preview bool)

// new notification types should add themselves to this array
var initers = []func(map[string]string) Notifier{}

// Init will take the given config map (from creds.json notifications key) and create a single Notifier with
// all notifications it has full config for.
func Init(config map[string]string) Notifier {
	notifiers := []Notifier{}
	for _, i := range initers {
		n := i(config)
		if n != nil {
			notifiers = append(notifiers, n)
		}
	}
	return func(domain, provider string, message string, err error, preview bool) {
		for _, n := range notifiers {
			n(domain, provider, message, err, preview)
		}
	}
}
