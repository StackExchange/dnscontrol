package soautil

import "strings"

func RFC5322MailToBind(rfc5322Mail string) string {
	res := strings.SplitN(rfc5322Mail, "@", 2)
	user_part, domain_part := res[0], res[1]
	// RFC-1035 [Section-8]
	user_part = strings.ReplaceAll(user_part, ".", "\\.")
	return user_part + "." + domain_part
}
