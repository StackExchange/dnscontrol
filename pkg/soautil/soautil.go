package soautil

import "strings"

// RFC5322MailToBind converts a user@host email address to BIND format.
func RFC5322MailToBind(rfc5322Mail string) string {
	res := strings.SplitN(rfc5322Mail, "@", 2)
	user, domain := res[0], res[1]
	// RFC-1035 [Section-8]
	user = strings.ReplaceAll(user, ".", "\\.")
	return user + "." + domain
}
