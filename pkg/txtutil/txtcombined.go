//go:generate stringer -type=State

package txtutil

// func ParseCombined(s string) (string, error) {
// 	return txtDecodeCombined(s)
// }

// // // txtDecode decodes TXT strings received from ROUTE53 and GCLOUD.
// func txtDecodeCombined(s string) (string, error) {

// 	// The dns package doesn't expose the quote parser. Therefore we create a TXT record and extract the strings.
// 	rr, err := dns.NewRR("example.com. IN TXT " + s)
// 	if err != nil {
// 		return "", fmt.Errorf("could not parse %q TXT: %w", s, err)
// 	}

// 	return strings.Join(rr.(*dns.TXT).Txt, ""), nil
// }

// func EncodeCombined(t string) string {
// 	return txtEncodeCombined(ToChunks(t))
// }

// // txtEncode encodes TXT strings as the old GetTargetCombined() function did.
// func txtEncodeCombined(ts []string) string {
// 	//printer.Printf("DEBUG: route53 txt outboundv=%v\n", ts)

// 	// Don't call this on fake types.
// 	rdtype := dns.StringToType["TXT"]

// 	// Magically create an RR of the correct type.
// 	rr := dns.TypeToRR[rdtype]()

// 	// Fill in the header.
// 	rr.Header().Name = "example.com."
// 	rr.Header().Rrtype = rdtype
// 	rr.Header().Class = dns.ClassINET
// 	rr.Header().Ttl = 300

// 	// Fill in the TXT data.
// 	rr.(*dns.TXT).Txt = ts

// 	// Generate the quoted string:
// 	header := rr.Header().String()
// 	full := rr.String()
// 	if !strings.HasPrefix(full, header) {
// 		panic("assertion failed. dns.Hdr.String() behavior has changed in an incompatible way")
// 	}

// 	//printer.Printf("DEBUG: route53 txt  encodedv=%v\n", t)
// 	return full[len(header):]
// }
