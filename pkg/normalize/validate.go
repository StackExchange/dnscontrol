package normalize

import (
	"fmt"
	"net"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/transform"
	"github.com/StackExchange/dnscontrol/v3/providers"
	"github.com/miekg/dns"
	"github.com/miekg/dns/dnsutil"
	"golang.org/x/text/encoding/charmap"
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
	if target == "" {
		return fmt.Errorf("empty target")
	}
	if strings.ContainsAny(target, `'" +,|!£$%&/()=?^*ç°§;:<>[]()@`) {
		return fmt.Errorf("target (%v) includes invalid char", target)
	}
	// If it contains a ".", it must end in a ".".
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
		"DS":               true,
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
			return fmt.Errorf("unsupported record type (%v) domain=%v name=%v", rec.Type, domain, rec.GetLabel())
		}
		for _, providerType := range pTypes {
			if providerType != cType.Provider {
				return fmt.Errorf("custom record type %s is not compatible with provider type %s", rec.Type, providerType)
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

// these record types may contain underscores
var rTypeUnderscores = []string{"SRV", "TLSA", "TXT"}

func checkLabel(label string, rType string, target, domain string, meta map[string]string) error {
	if label == "@" {
		return nil
	}
	if label == "" {
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
	// Don't warn for records that start with _
	// See https://github.com/StackExchange/dnscontrol/issues/829
	if strings.HasPrefix(label, "_") || strings.Contains(label, "._") {
		return nil
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
			err := fmt.Errorf("in %s %s.%s: %s", rec.Type, rec.GetLabel(), domain, e.Error())
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
	case "TXT", "IMPORT_TRANSFORM", "CAA", "SSHFP", "TLSA", "DS":
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
func importTransform(srcDomain, dstDomain *models.DomainConfig, transforms []transform.IPConversion, ttl uint32) error {
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
			trs, err := transform.IPToList(net.ParseIP(rec.GetTargetField()), transforms)
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

// ValidateAndNormalizeConfig performs and normalization and/or validation of the IR.
func ValidateAndNormalizeConfig(config *models.DNSConfig) (errs []error) {
	for _, domain := range config.Domains {
		pTypes := []string{}
		for _, provider := range domain.DNSProviderInstances {
			pType := provider.ProviderType
			// If NO_PURGE is in use, make sure this *isn't* a provider that *doesn't* support NO_PURGE.
			if domain.KeepUnknown && providers.ProviderHasCapability(pType, providers.CantUseNOPURGE) {
				errs = append(errs, fmt.Errorf("%s uses NO_PURGE which is not supported by %s(%s)", domain.Name, provider.Name, pType))
			}
		}

		// Normalize Nameservers.
		for _, ns := range domain.Nameservers {
			// NB(tlim): Like any target, NAMESERVER() is input by the user
			// as a shortname or a FQDN+dot.
			if err := checkTarget(ns.Name); err != nil {
				errs = append(errs, err)
			}
			// Unlike any other FQDN in this system, it is stored as a FQDN without the trailing dot.
			n := dnsutil.AddOrigin(ns.Name, domain.Name+".")
			ns.Name = strings.TrimSuffix(n, ".")
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
				origin := domain.Name + "."
				if len(rec.SubDomain) > 0 {
					origin = rec.SubDomain + "." + origin
				}
				rec.SetTarget(dnsutil.AddOrigin(rec.GetTargetField(), origin))
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
				if rec.TlsaUsage > 3 {
					errs = append(errs, fmt.Errorf("TLSA Usage %d is invalid in record %s (domain %s)",
						rec.TlsaUsage, rec.GetLabel(), domain.Name))
				}
				if rec.TlsaSelector > 1 {
					errs = append(errs, fmt.Errorf("TLSA Selector %d is invalid in record %s (domain %s)",
						rec.TlsaSelector, rec.GetLabel(), domain.Name))
				}
				if rec.TlsaMatchingType > 2 {
					errs = append(errs, fmt.Errorf("TLSA MatchingType %d is invalid in record %s (domain %s)",
						rec.TlsaMatchingType, rec.GetLabel(), domain.Name))
				}
			} else if rec.Type == "TXT" {
				for i := range rec.TxtStrings {
					encoded, err := charmap.ISO8859_1.NewEncoder().String(rec.TxtStrings[i])
					if err != nil {
						errs = append(errs, fmt.Errorf("TXT record %s contains characters > 0xFF (domain %s)",
							rec.GetLabel(), domain.Name), err)
					} else {
						rec.TxtStrings[i] = encoded
					}
				}
			}

			// Populate FQDN:
			rec.SetLabel(rec.GetLabel(), domain.Name)
		}
	}

	// SPF flattening
	if ers := flattenSPFs(config); len(ers) > 0 {
		errs = append(errs, ers...)
	}

	// Split TXT targets that are >255 bytes (if permitted)
	for _, domain := range config.Domains {
		for _, rec := range domain.Records {
			if rec.Type == "TXT" {
				if txtAlgo, ok := rec.Metadata["txtSplitAlgorithm"]; ok {
					rec.TxtNormalize(txtAlgo)
				}
			}
		}
	}

	// Validate TXT records.
	for _, domain := range config.Domains {
		// Collect the names of providers that don't support TXTMulti:
		txtMultiDissenters := []string{}
		for _, provider := range domain.DNSProviderInstances {
			pType := provider.ProviderType
			if !providers.ProviderHasCapability(pType, providers.CanUseTXTMulti) {
				txtMultiDissenters = append(txtMultiDissenters, provider.Name)
			}
		}
		// Validate TXT records.
		for _, rec := range domain.Records {
			if rec.Type == "TXT" {
				// If TXTMulti is required, all providers must support that feature.
				if len(rec.TxtStrings) > 1 && len(txtMultiDissenters) > 0 {
					errs = append(errs,
						fmt.Errorf("TXT records with multiple strings not supported by %s (label=%q domain=%v)",
							strings.Join(txtMultiDissenters, ","), rec.GetLabel(), domain.Name))
				}
				// Validate the record:
				if err := models.ValidateTXT(rec); err != nil {
					errs = append(errs, err)
				}
			}
		}
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
				errs = append(errs, fmt.Errorf("record named '%s' does not have correct FQDN for domain '%s'. FQDN: %s", r.Name, d.Name, r.NameFQDN))
			}
		}
		// Verify AutoDNSSEC is valid.
		errs = append(errs, checkAutoDNSSEC(d)...)
	}

	return errs
}

func checkAutoDNSSEC(dc *models.DomainConfig) (errs []error) {
	if dc.AutoDNSSEC != "" && dc.AutoDNSSEC != "on" && dc.AutoDNSSEC != "off" {
		errs = append(errs, fmt.Errorf("Domain %q AutoDNSSEC=%q is invalid (expecting \"\", \"off\", or \"on\")", dc.Name, dc.AutoDNSSEC))
	}
	return
}

func checkCNAMEs(dc *models.DomainConfig) (errs []error) {
	cnames := map[string]bool{}
	for _, r := range dc.Records {
		if r.Type == "CNAME" {
			if cnames[r.GetLabel()] {
				errs = append(errs, fmt.Errorf("cannot have multiple CNAMEs with same name: %s", r.GetLabelFQDN()))
			}
			cnames[r.GetLabel()] = true
		}
	}
	for _, r := range dc.Records {
		if cnames[r.GetLabel()] && r.Type != "CNAME" {
			errs = append(errs, fmt.Errorf("cannot have CNAME and %s record with same name: %s", r.Type, r.GetLabelFQDN()))
		}
	}
	return
}

func checkDuplicates(records []*models.RecordConfig) (errs []error) {
	seen := map[string]*models.RecordConfig{}
	for _, r := range records {
		diffable := fmt.Sprintf("%s %s %s", r.GetLabelFQDN(), r.Type, r.ToDiffable())
		if seen[diffable] != nil {
			errs = append(errs, fmt.Errorf("exact duplicate record found: %s", diffable))
		}
		seen[diffable] = r
	}
	return errs
}

// We pull this out of checkProviderCapabilities() so that it's visible within
// the package elsewhere, so that our test suite can look at the list of
// capabilities we're checking and make sure that it's up-to-date.
var providerCapabilityChecks = []pairTypeCapability{
	// If a zone uses rType X, the provider must support capability Y.
	//{"X", providers.Y},
	capabilityCheck("ALIAS", providers.CanUseAlias),
	capabilityCheck("AUTODNSSEC", providers.CanAutoDNSSEC),
	capabilityCheck("CAA", providers.CanUseCAA),
	capabilityCheck("NAPTR", providers.CanUseNAPTR),
	capabilityCheck("PTR", providers.CanUsePTR),
	capabilityCheck("R53_ALIAS", providers.CanUseRoute53Alias),
	capabilityCheck("SSHFP", providers.CanUseSSHFP),
	capabilityCheck("SRV", providers.CanUseSRV),
	capabilityCheck("TLSA", providers.CanUseTLSA),
	capabilityCheck("AZURE_ALIAS", providers.CanUseAzureAlias),

	// DS needs special record-level checks
	{
		rType:     "DS",
		caps:      []providers.Capability{providers.CanUseDS, providers.CanUseDSForChildren},
		checkFunc: checkProviderDS,
	},
}

type pairTypeCapability struct {
	rType string
	// Capabilities the provider must implement if any records of type rType are found
	// in the zonefile. This is a disjunction - implementing at least one of the listed
	// capabilities is sufficient.
	caps []providers.Capability
	// checkFunc provides additional checks of each provider. This function should be
	// called if records of type rType are found in the zonefile.
	checkFunc func(pType string, _ models.Records) error
}

func capabilityCheck(rType string, caps ...providers.Capability) pairTypeCapability {
	return pairTypeCapability{
		rType: rType,
		caps:  caps,
	}
}

func providerHasAtLeastOneCapability(pType string, caps ...providers.Capability) bool {
	for _, cap := range caps {
		if providers.ProviderHasCapability(pType, cap) {
			return true
		}
	}

	return false
}

func checkProviderDS(pType string, records models.Records) error {
	switch {
	case providers.ProviderHasCapability(pType, providers.CanUseDS):
		// The provider can use DS records anywhere, including at the root
		return nil
	case !providers.ProviderHasCapability(pType, providers.CanUseDSForChildren):
		// Provider has no support for DS records
		return fmt.Errorf("provider %s uses DS records but does not support them", pType)
	default:
		// Provider supports DS records but not at the root
		for _, record := range records {
			if record.Type == "DS" && record.Name == "@" {
				return fmt.Errorf(
					"provider %s only supports child DS records, but zone had a record at the root (@)",
					pType,
				)
			}
		}
	}

	return nil
}

func checkProviderCapabilities(dc *models.DomainConfig) error {
	// Check if the zone uses a capability that the provider doesn't
	// support.
	for _, ty := range providerCapabilityChecks {
		hasAny := false
		switch ty.rType {
		case "AUTODNSSEC":
			if dc.AutoDNSSEC != "" {
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
			if !providerHasAtLeastOneCapability(provider.ProviderType, ty.caps...) {
				return fmt.Errorf("domain %s uses %s records, but DNS provider type %s does not support them", dc.Name, ty.rType, provider.ProviderType)
			}

			if ty.checkFunc != nil {
				checkErr := ty.checkFunc(provider.ProviderType, dc.Records)
				if checkErr != nil {
					return fmt.Errorf("while checking %s records in domain %s: %w", ty.rType, dc.Name, checkErr)
				}
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
		newIPs, err := transform.IPToList(net.ParseIP(rec.GetTargetField()), table)
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
