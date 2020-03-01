package normalize

import (
	"fmt"
	"net"
	"strings"

	"github.com/StackExchange/dnscontrol/v2/models"
	"github.com/StackExchange/dnscontrol/v2/pkg/transform"
	"github.com/StackExchange/dnscontrol/v2/providers"
	"github.com/miekg/dns"
	"github.com/miekg/dns/dnsutil"
)

// Returns false if target does not validate.
func checkIPv4(label string) error {
	if net.ParseIP(label).To4() == nil {
		return fmt.Errorf("WARNING: target (%v) is not an IPv4 address", label)
	}
	return nil
}

// Returns false if target does not validate.
func checkIPv6(label string) error {
	if net.ParseIP(label).To16() == nil {
		return fmt.Errorf("WARNING: target (%v) is not an IPv6 address", label)
	}
	return nil
}

// make sure target is valid reference for cnames, mx, etc.
func checkTarget(target string) error {
	if target == "@" {
		return nil
	}
	if len(target) < 1 {
		return fmt.Errorf("empty target")
	}
	if strings.ContainsAny(target, `'" +,|!£$%&/()=?^*ç°§;:<>[]()@`) {
		return fmt.Errorf("target (%v) includes invalid char", target)
	}
	// If it containts a ".", it must end in a ".".
	if strings.ContainsRune(target, '.') && target[len(target)-1] != '.' {
		return fmt.Errorf("target (%v) must end with a (.) [https://stackexchange.github.io/dnscontrol/why-the-dot]", target)
	}
	return nil
}

// validateRecordTypes list of valid rec.Type values. Returns true if this is a real DNS record type, false means it is a pseudo-type used internally.
func validateRecordTypes(rec *models.RecordConfig, domain string, pTypes []string) error {
	var validTypes = map[string]bool{
		"A":                true,
		"AAAA":             true,
		"CNAME":            true,
		"CAA":              true,
		"TLSA":             true,
		"IMPORT_TRANSFORM": false,
		"MX":               true,
		"SRV":              true,
		"SSHFP":            true,
		"TXT":              true,
		"NS":               true,
		"PTR":              true,
		"NAPTR":            true,
		"ALIAS":            false,
	}
	_, ok := validTypes[rec.Type]
	if !ok {
		cType := providers.GetCustomRecordType(rec.Type)
		if cType == nil {
			return fmt.Errorf("Unsupported record type (%v) domain=%v name=%v", rec.Type, domain, rec.GetLabel())
		}
		for _, providerType := range pTypes {
			if providerType != cType.Provider {
				return fmt.Errorf("Custom record type %s is not compatible with provider type %s", rec.Type, providerType)
			}
		}
		// it is ok. Lets replace the type with real type and add metadata to say we checked it
		rec.Metadata["orig_custom_type"] = rec.Type
		if cType.RealType != "" {
			rec.Type = cType.RealType
		}
	}
	return nil
}

// underscores in names are often used erroneously. They are valid for dns records, but invalid for urls.
// here we list common records expected to have underscores. Anything else containing an underscore will print a warning.
var labelUnderscores = []string{
	"_acme-challenge",
	"_amazonses",
	"_dmarc",
	"_domainkey",
	"_jabber",
	"_mta-sts",
	"_sip",
	"_xmpp",
}

// these record types may contain underscores
var rTypeUnderscores = []string{"SRV", "TLSA", "TXT"}

func checkLabel(label string, rType string, target, domain string, meta map[string]string) error {
	if label == "@" {
		return nil
	}
	if len(label) < 1 {
		return fmt.Errorf("empty %s label in %s", rType, domain)
	}
	if label[len(label)-1] == '.' {
		return fmt.Errorf("label %s.%s ends with a (.)", label, domain)
	}
	if strings.HasSuffix(label, domain) {
		if m := meta["skip_fqdn_check"]; m != "true" {
			return fmt.Errorf(`label %s ends with domain name %s. Record names should not be fully qualified. Add {skip_fqdn_check:"true"} to this record if you really want to make %s.%s`, label, domain, label, domain)
		}
	}

	// Underscores are permitted in labels, but we print a warning unless they
	// are used in a way we consider typical.  Yes, we're opinionated here.

	// Don't warn for certain rtypes:
	for _, ex := range rTypeUnderscores {
		if rType == ex {
			return nil
		}
	}
	// Don't warn for CNAMEs if the target ends with acm-validations.aws
	// See https://github.com/StackExchange/dnscontrol/issues/519
	if strings.HasPrefix(label, "_") && rType == "CNAME" && strings.HasSuffix(target, ".acm-validations.aws.") {
		return nil
	}
	// Don't warn for certain label substrings
	for _, ex := range labelUnderscores {
		if strings.Contains(label, ex) {
			return nil
		}
	}
	// Otherwise, warn.
	if strings.ContainsRune(label, '_') {
		return Warning{fmt.Errorf("label %s.%s contains an underscore", label, domain)}
	}

	return nil
}

// checkTargets returns true if rec.Target is valid for the rec.Type.
func checkTargets(rec *models.RecordConfig, domain string) (errs []error) {
	label := rec.GetLabel()
	target := rec.GetTargetField()
	check := func(e error) {
		if e != nil {
			err := fmt.Errorf("In %s %s.%s: %s", rec.Type, rec.GetLabel(), domain, e.Error())
			if _, ok := e.(Warning); ok {
				err = Warning{err}
			}
			errs = append(errs, err)
		}
	}
	switch rec.Type { // #rtype_variations
	case "A":
		check(checkIPv4(target))
	case "AAAA":
		check(checkIPv6(target))
	case "CNAME":
		check(checkTarget(target))
		if label == "@" {
			check(fmt.Errorf("cannot create CNAME record for bare domain"))
		}
	case "MX":
		check(checkTarget(target))
	case "NS":
		check(checkTarget(target))
		if label == "@" {
			check(fmt.Errorf("cannot create NS record for bare domain. Use NAMESERVER instead"))
		}
	case "PTR":
		check(checkTarget(target))
	case "NAPTR":
		check(checkTarget(target))
	case "ALIAS":
		check(checkTarget(target))
	case "SOA":
		check(checkTarget(target))
	case "SRV":
		check(checkTarget(target))
	case "TXT", "IMPORT_TRANSFORM", "CAA", "SSHFP", "TLSA":
	default:
		if rec.Metadata["orig_custom_type"] != "" {
			// it is a valid custom type. We perform no validation on target
			return
		}
		errs = append(errs, fmt.Errorf("checkTargets: Unimplemented record type (%v) domain=%v name=%v",
			rec.Type, domain, rec.GetLabel()))
	}
	return
}

func transformCNAME(target, oldDomain, newDomain string) string {
	// Canonicalize. If it isn't a FQDN, add the newDomain.
	result := dnsutil.AddOrigin(target, oldDomain)
	if dns.IsFqdn(result) {
		result = result[:len(result)-1]
	}
	return dnsutil.AddOrigin(result, newDomain) + "."
}

// import_transform imports the records of one zone into another, modifying records along the way.
func importTransform(srcDomain, dstDomain *models.DomainConfig, transforms []transform.IpConversion, ttl uint32) error {
	// Read srcDomain.Records, transform, and append to dstDomain.Records:
	// 1. Skip any that aren't A or CNAMEs.
	// 2. Append destDomainname to the end of the label.
	// 3. For CNAMEs, append destDomainname to the end of the target.
	// 4. For As, change the target as described the transforms.

	for _, rec := range srcDomain.Records {
		if dstDomain.Records.HasRecordTypeName(rec.Type, rec.GetLabelFQDN()) {
			continue
		}
		newRec := func() *models.RecordConfig {
			rec2, _ := rec.Copy()
			newlabel := rec2.GetLabelFQDN()
			rec2.SetLabel(newlabel, dstDomain.Name)
			if ttl != 0 {
				rec2.TTL = ttl
			}
			return rec2
		}
		switch rec.Type { // #rtype_variations
		case "A":
			trs, err := transform.TransformIPToList(net.ParseIP(rec.GetTargetField()), transforms)
			if err != nil {
				return fmt.Errorf("import_transform: TransformIP(%v, %v) returned err=%s", rec.GetTargetField(), transforms, err)
			}
			for _, tr := range trs {
				r := newRec()
				r.SetTarget(tr.String())
				dstDomain.Records = append(dstDomain.Records, r)
			}
		case "CNAME":
			r := newRec()
			r.SetTarget(transformCNAME(r.GetTargetField(), srcDomain.Name, dstDomain.Name))
			dstDomain.Records = append(dstDomain.Records, r)
		case "MX", "NAPTR", "NS", "SOA", "SRV", "TXT", "CAA", "TLSA":
			// Not imported.
			continue
		default:
			return fmt.Errorf("import_transform: Unimplemented record type %v (%v)",
				rec.Type, rec.GetLabel())
		}
	}
	return nil
}

// deleteImportTransformRecords deletes any IMPORT_TRANSFORM records from a domain.
func deleteImportTransformRecords(domain *models.DomainConfig) {
	for i := len(domain.Records) - 1; i >= 0; i-- {
		rec := domain.Records[i]
		if rec.Type == "IMPORT_TRANSFORM" {
			domain.Records = append(domain.Records[:i], domain.Records[i+1:]...)
		}
	}
}

// Warning is a wrapper around error that can be used to indicate it should not
// stop execution, but is still likely a problem.
type Warning struct {
	error
}

// NormalizeAndValidateConfig performs and normalization and/or validation of the IR.
func NormalizeAndValidateConfig(config *models.DNSConfig) (errs []error) {
	for _, domain := range config.Domains {
		pTypes := []string{}
		txtMultiDissenters := []string{}
		for _, provider := range domain.DNSProviderInstances {
			pType := provider.ProviderType
			// If NO_PURGE is in use, make sure this *isn't* a provider that *doesn't* support NO_PURGE.
			if domain.KeepUnknown && providers.ProviderHasCapability(pType, providers.CantUseNOPURGE) {
				errs = append(errs, fmt.Errorf("%s uses NO_PURGE which is not supported by %s(%s)", domain.Name, provider.Name, pType))
			}

			// Record if any providers do not support TXTMulti:
			if !providers.ProviderHasCapability(pType, providers.CanUseTXTMulti) {
				txtMultiDissenters = append(txtMultiDissenters, provider.Name)
			}
		}

		// Normalize Nameservers.
		for _, ns := range domain.Nameservers {
			// NB(tlim): Like any target, NAMESERVER() is input by the user
			// as a shortname or a FQDN+dot.  It is stored as FQDN+dot.
			// Normalize it like we do any target to assure it is FQDN+dot
			ns.Name = dnsutil.AddOrigin(ns.Name, domain.Name+".")
			ns.Name = strings.TrimSuffix(ns.Name, ".")
			checkTarget(ns.Name)
		}

		// Normalize Records.
		models.PostProcessRecords(domain.Records)
		for _, rec := range domain.Records {
			if rec.TTL == 0 {
				rec.TTL = models.DefaultTTL
			}
			// Validate the unmodified inputs:
			if err := validateRecordTypes(rec, domain.Name, pTypes); err != nil {
				errs = append(errs, err)
			}
			if err := checkLabel(rec.GetLabel(), rec.Type, rec.GetTargetField(), domain.Name, rec.Metadata); err != nil {
				errs = append(errs, err)
			}
			if errs2 := checkTargets(rec, domain.Name); errs2 != nil {
				errs = append(errs, errs2...)
			}

			// Canonicalize Targets.
			if rec.Type == "CNAME" || rec.Type == "MX" || rec.Type == "NAPTR" || rec.Type == "NS" || rec.Type == "SRV" {
				// #rtype_variations
				// These record types have a target that is a hostname.
				// We normalize them to a FQDN so there is less variation to handle.  If a
				// provider API requires a shortname, the provider must do the shortening.
				rec.SetTarget(dnsutil.AddOrigin(rec.GetTargetField(), domain.Name+"."))
			} else if rec.Type == "A" || rec.Type == "AAAA" {
				rec.SetTarget(net.ParseIP(rec.GetTargetField()).String())
			} else if rec.Type == "PTR" {
				var err error
				var name string
				if name, err = transform.PtrNameMagic(rec.GetLabel(), domain.Name); err != nil {
					errs = append(errs, err)
				}
				rec.SetLabel(name, domain.Name)
			} else if rec.Type == "CAA" {
				if rec.CaaTag != "issue" && rec.CaaTag != "issuewild" && rec.CaaTag != "iodef" {
					errs = append(errs, fmt.Errorf("CAA tag %s is invalid", rec.CaaTag))
				}
			} else if rec.Type == "TLSA" {
				if rec.TlsaUsage < 0 || rec.TlsaUsage > 3 {
					errs = append(errs, fmt.Errorf("TLSA Usage %d is invalid in record %s (domain %s)",
						rec.TlsaUsage, rec.GetLabel(), domain.Name))
				}
				if rec.TlsaSelector < 0 || rec.TlsaSelector > 1 {
					errs = append(errs, fmt.Errorf("TLSA Selector %d is invalid in record %s (domain %s)",
						rec.TlsaSelector, rec.GetLabel(), domain.Name))
				}
				if rec.TlsaMatchingType < 0 || rec.TlsaMatchingType > 2 {
					errs = append(errs, fmt.Errorf("TLSA MatchingType %d is invalid in record %s (domain %s)",
						rec.TlsaMatchingType, rec.GetLabel(), domain.Name))
				}
			} else if rec.Type == "TXT" && len(txtMultiDissenters) != 0 && len(rec.TxtStrings) > 1 {
				// There are providers that  don't support TXTMulti yet there is
				// a TXT record with multiple strings:
				errs = append(errs,
					fmt.Errorf("TXT records with multiple strings (label %v domain: %v) not supported by %s",
						rec.GetLabel(), domain.Name, strings.Join(txtMultiDissenters, ",")))
			}

			// Populate FQDN:
			rec.SetLabel(rec.GetLabel(), domain.Name)
		}
	}

	// SPF flattening
	if ers := flattenSPFs(config); len(ers) > 0 {
		errs = append(errs, ers...)
	}

	// Process IMPORT_TRANSFORM
	for _, domain := range config.Domains {
		for _, rec := range domain.Records {
			if rec.Type == "IMPORT_TRANSFORM" {
				table, err := transform.DecodeTransformTable(rec.Metadata["transform_table"])
				if err != nil {
					errs = append(errs, err)
					continue
				}
				err = importTransform(config.FindDomain(rec.GetTargetField()), domain, table, rec.TTL)
				if err != nil {
					errs = append(errs, err)
				}
			}
		}
	}
	// Clean up:
	for _, domain := range config.Domains {
		deleteImportTransformRecords(domain)
	}
	// Run record transforms
	for _, domain := range config.Domains {
		if err := applyRecordTransforms(domain); err != nil {
			errs = append(errs, err)
		}
	}

	for _, d := range config.Domains {
		// Check that CNAMES don't have to co-exist with any other records
		errs = append(errs, checkCNAMEs(d)...)
		// Check that if any advanced record types are used in a domain, every provider for that domain supports them
		err := checkProviderCapabilities(d)
		if err != nil {
			errs = append(errs, err)
		}
		// Check for duplicates
		errs = append(errs, checkDuplicates(d.Records)...)
		// Validate FQDN consistency
		for _, r := range d.Records {
			if r.NameFQDN == "" || !strings.HasSuffix(r.NameFQDN, d.Name) {
				errs = append(errs, fmt.Errorf("Record named '%s' does not have correct FQDN in domain '%s'. FQDN: %s", r.Name, d.Name, r.NameFQDN))
			}
		}
	}

	return errs
}

func checkCNAMEs(dc *models.DomainConfig) (errs []error) {
	cnames := map[string]bool{}
	for _, r := range dc.Records {
		if r.Type == "CNAME" {
			if cnames[r.GetLabel()] {
				errs = append(errs, fmt.Errorf("Cannot have multiple CNAMEs with same name: %s", r.GetLabelFQDN()))
			}
			cnames[r.GetLabel()] = true
		}
	}
	for _, r := range dc.Records {
		if cnames[r.GetLabel()] && r.Type != "CNAME" {
			errs = append(errs, fmt.Errorf("Cannot have CNAME and %s record with same name: %s", r.Type, r.GetLabelFQDN()))
		}
	}
	return
}

func checkDuplicates(records []*models.RecordConfig) (errs []error) {
	seen := map[string]*models.RecordConfig{}
	for _, r := range records {
		diffable := fmt.Sprintf("%s %s %s", r.GetLabelFQDN(), r.Type, r.ToDiffable())
		if seen[diffable] != nil {
			errs = append(errs, fmt.Errorf("Exact duplicate record found: %s", diffable))
		}
		seen[diffable] = r
	}
	return errs
}

// We pull this out of checkProviderCapabilities() so that it's visible within
// the package elsewhere, so that our test suite can look at the list of
// capabilities we're checking and make sure that it's up-to-date.
var providerCapabilityChecks []pairTypeCapability

type pairTypeCapability struct {
	rType string
	cap   providers.Capability
}

func init() {
	providerCapabilityChecks = []pairTypeCapability{
		// If a zone uses rType X, the provider must support capability Y.
		//{"X", providers.Y},
		{"ALIAS", providers.CanUseAlias},
		{"AUTODNSSEC", providers.CanAutoDNSSEC},
		{"CAA", providers.CanUseCAA},
		{"NAPTR", providers.CanUseNAPTR},
		{"PTR", providers.CanUsePTR},
		{"R53_ALIAS", providers.CanUseRoute53Alias},
		{"SSHFP", providers.CanUseSSHFP},
		{"SRV", providers.CanUseSRV},
		{"TLSA", providers.CanUseTLSA},
	}
}

func checkProviderCapabilities(dc *models.DomainConfig) error {
	// Check if the zone uses a capability that the provider doesn't
	// support.
	for _, ty := range providerCapabilityChecks {
		hasAny := false
		switch ty.rType {
		case "AUTODNSSEC":
			if dc.AutoDNSSEC {
				hasAny = true
			}
		default:
			for _, r := range dc.Records {
				if r.Type == ty.rType {
					hasAny = true
					break
				}
			}

		}
		if !hasAny {
			continue
		}
		for _, provider := range dc.DNSProviderInstances {
			// fmt.Printf("  (checking if %q can %q for domain %q)\n", provider.ProviderType, ty.rType, dc.Name)
			if !providers.ProviderHasCapability(provider.ProviderType, ty.cap) {
				return fmt.Errorf("Domain %s uses %s records, but DNS provider type %s does not support them", dc.Name, ty.rType, provider.ProviderType)
			}
		}
	}
	return nil
}

func applyRecordTransforms(domain *models.DomainConfig) error {
	for _, rec := range domain.Records {
		if rec.Type != "A" {
			continue
		}
		tt, ok := rec.Metadata["transform"]
		if !ok {
			continue
		}
		table, err := transform.DecodeTransformTable(tt)
		if err != nil {
			return err
		}
		ip := net.ParseIP(rec.GetTargetField()) // ip already validated above
		newIPs, err := transform.TransformIPToList(net.ParseIP(rec.GetTargetField()), table)
		if err != nil {
			return err
		}
		for i, newIP := range newIPs {
			if i == 0 && !newIP.Equal(ip) {
				rec.SetTarget(newIP.String()) // replace target of first record if different
			} else if i > 0 {
				// any additional ips need identical records with the alternate ip added to the domain
				copy, err := rec.Copy()
				if err != nil {
					return err
				}
				copy.SetTarget(newIP.String())
				domain.Records = append(domain.Records, copy)
			}
		}
	}
	return nil
}
