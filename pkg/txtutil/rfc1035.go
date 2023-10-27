package txtutil

import (
	"strings"
	//"github.com/facebook/dns/dnsrocks/dnsdata/quote"
)

func RFC1035Quoted(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `"` + s + `"`

	//sb := []byte(s)        // The string, as []byte
	//qb := quote.Bquote(sb) // Quote it.
	//q := string(qb[:])     // Convert to string
	//return `"` + q + `"`
}

func RFC1035ChunkedAndQuoted(s string) string {

	parts := ToChunks(s)
	var quotedParts []string

	for _, part := range parts {
		quotedParts = append(quotedParts, RFC1035Quoted(part))
	}

	return strings.Join(quotedParts, " ")
}
