package cscglobal

import (
	"fmt"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/providers"
)

/*

CSC Global Registrar:

Info required in `creds.json`:
   - api-key             Api Key
   - user-token          User Token
   - notification_emails (optional) Comma separated list of email addresses to send notifications to
*/

func init() {
	providers.RegisterRegistrarType("CSCGLOBAL", newCscGlobal)
}

func newCscGlobal(m map[string]string) (providers.Registrar, error) {
	api := &cscglobalProvider{}

	api.key, api.token = m["api-key"], m["user-token"]
	if api.key == "" || api.token == "" {
		return nil, fmt.Errorf("missing CSC Global api-key and/or user-token")
	}

	if m["notification_emails"] != "" {
		api.notifyEmails = strings.Split(m["notification_emails"], ",")
	}

	return api, nil
}
