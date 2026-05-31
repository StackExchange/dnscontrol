package models

/*
rc := mk.MakeRC(origin, dc.LabelFromConfig(label), 0, TypeA, "1.2.3.4") -- dc.LabelFromConfig returns (string, error) and must be checked.
rc := mk.MakeRC(origin, dc.LabelFromShort(label), 0, TypeA, "1.2.3.4")
rc := mk.MakeRC(origin, dc.LabelFromFQDN(label), 0, TypeA, "1.2.3.4")
rc := mk.MakeRC(origin, dc.LabelFromFQDNDot(label), 0, TypeA, "1.2.3.4"
*/

// LabelFromConfig processes a label from dnsconfig.js. It returns a shortname (that is, the name without the domain).
// It performs the following transformations:
// Phase 1: Turn into a FQDN:
// - If it ends in ".", this phase is done.
// - If it is "", return an error suggesting to use "@" instead.
// - If it contains
// - If it is "@", it becomes dc.Name + "." (the domain name). This phase is done.
// - If it starts with "*.", it becomes "*." + dc.Name + "." (a wildcard). This phase is done.
// - If it is "*", it becomes "*." + dc.Name + "." (a wildcard). This phase is done.
// - Otherwise, it becomes label + "." + dc.Name + ".". This phase is done.
// Phase 2: Normalize the FQDN:
// - Convert unicode to ASCII (punycode).
// - Downcase.
// - Convert back to Unicode.
// Phase 3: store
// - Store:
//   - Name
//   - NameUnicode
//   - NameFQDN
//   - NameUnicodeUnicode
// - Convert unicode to ASCII (punycode).
// - Downcase.
// - If it i
// - If the result doesn't end in ".",
// - If the result ends in "."

func (dc *DomainConfig) LabelFromConfig(label string) (string, error) {
	// Normalize like we do from a human.
	// convert unicode to ASCII. Downcase.
	// Fail if result ends in "."
	//
	return label, nil
}
func (dc *DomainConfig) LabelFromShort(label string) (string, error) {
	return label, nil
}
func (dc *DomainConfig) LabelFromFQDNNoDot(label string) (string, error) {
	return label, nil
}
