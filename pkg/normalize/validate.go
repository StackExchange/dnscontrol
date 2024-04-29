package normalize

import (
	"fmt"
	"net"
	"sort"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/transform"
	"github.com/StackExchange/dnscontrol/v4/providers"
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
	if target == "" {
		return fmt.Errorf("empty target")
	}
	if strings.ContainsAny(target, `'" +,|!£$%&()=?^*ç°§;:<>[]()@`) {
		return fmt.Errorf("target (%v) includes invalid char", target)
	}
	if !strings.HasSuffix(target, ".in-addr.arpa.") && strings.Contains(target, "/") {
		return fmt.Errorf("target (%v) includes invalid char", target)
	}
	// If it contains a ".", it must end in a ".".
	if strings.ContainsRune(target, '.') && target[len(target)-1] != '.' {
		return fmt.Errorf("target (%v) must end with a (.) [https://docs.dnscontrol.org/language-reference/why-the-dot]", target)
	}
	return nil
}

// validateRecordTypes list of valid rec.Type values. Returns true if this is a real DNS record type, false means it is a pseudo-type used internally.
func validateRecordTypes(rec *models.RecordConfig, domain string, pTypes []string) error {
	// #rtype_variations
	var validTypes = map[string]bool{
		"A":                true,
		"AAAA":             true,
		"ALIAS":            false,
		"CAA":              true,
		"CNAME":            true,
		"DHCID":            true,
		"DNAME":            true,
		"DS":               true,
		"DNSKEY":           true,
		"HTTPS":            true,
		"IMPORT_TRANSFORM": false,
		"LOC":              true,
		"MX":               true,
		"NAPTR":            true,
		"NS":               true,
		"PTR":              true,
		"SOA":              true,
		"SRV":              true,
		"SSHFP":            true,
		"SVCB":             true,
		"TLSA":             true,
		"TXT":              true,
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

func errorRepeat(label, domain string) string {
	shortname := strings.TrimSuffix(label, "."+domain)
	return fmt.Sprintf(
		`The name "%s.%s." is an error (repeats the domain). Maybe instead of "%s" you intended "%s"? If not add DISABLE_REPEATED_DOMAIN_CHECK to this record to permit this as-is.`,
		label, domain,
		label,
		shortname,
	)
}

func checkLabel(label string, rType string, domain string, meta map[string]string) error {
	if label == "@" {
		return nil
	}
	if label == "" {
		return fmt.Errorf("empty %s label in %s", rType, domain)
	}
	if label[len(label)-1] == '.' {
		return fmt.Errorf("label %s.%s ends with a (.)", label, domain)
	}
	if label == domain || strings.HasSuffix(label, "."+domain) {
		if m := meta["skip_fqdn_check"]; m != "true" {
			return fmt.Errorf(errorRepeat(label, domain))
		}
	}

	// Underscores are permitted in labels, but we print a warning unless they
	// are used in a way we consider typical.  Yes, we're opinionated here.

	// Don't warn for certain rtypes:
	for _, ex := range []string{"SRV", "TLSA", "TXT"} {
		if rType == ex {
			return nil
		}
	}
	// Don't warn for records that start with _
	// See https://github.com/StackExchange/dnscontrol/issues/829
	if strings.HasPrefix(label, "_") || strings.Contains(label, "._") || strings.HasPrefix(label, "sql-") {
		return nil
	}

	// Otherwise, warn.
	if strings.ContainsRune(label, '_') {
		return Warning{fmt.Errorf("label %s.%s contains \"_\" (can't be used in a URL)", label, domain)}
	}

	return nil
}

func checkSoa(expire uint32, minttl uint32, refresh uint32, retry uint32, mbox string) error {
	if expire <= 0 {
		return fmt.Errorf("SOA Expire must be > 0")
	}
	if minttl <= 0 {
		return fmt.Errorf("SOA Minimum TTL must be > 0")
	}
	if refresh <= 0 {
		return fmt.Errorf("SOA Refresh must be > 0")
	}
	if retry <= 0 {
		return fmt.Errorf("SOA Retry must be > 0")
	}
	if mbox == "" {
		return fmt.Errorf("SOA MBox must be specified")
	}
	if strings.ContainsRune(mbox, '@') {
		return fmt.Errorf("SOA MBox must have '.' instead of '@'")
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
	case "ALIAS":
		check(checkTarget(target))
	case "CNAME":
		check(checkTarget(target))
		if label == "@" {
			check(fmt.Errorf("cannot create CNAME record for bare domain"))
		}
		labelFQDN := dnsutil.AddOrigin(label, domain)
		targetFQDN := dnsutil.AddOrigin(target, domain)
		if labelFQDN == targetFQDN {
			check(fmt.Errorf("CNAME loop (target points at itself)"))
		}
	case "DNAME":
		check(checkTarget(target))
	case "LOC":
	case "MX":
		check(checkTarget(target))
	case "NAPTR":
		if target != "" {
			check(checkTarget(target))
		}
	case "NS":
		check(checkTarget(target))
		if label == "@" {
			check(fmt.Errorf("cannot create NS record for bare domain. Use NAMESERVER instead"))
		}
	case "NS1_URLFWD":
		if len(strings.Fields(target)) != 5 {
			check(fmt.Errorf("record should follow format: \"from to redirectType pathForwardingMode queryForwarding\""))
		}
	case "PTR":
		check(checkTarget(target))
	case "SOA":
		check(checkSoa(rec.SoaExpire, rec.SoaMinttl, rec.SoaRefresh, rec.SoaRetry, rec.SoaMbox))
		check(checkTarget(target))
		if label != "@" {
			check(fmt.Errorf("SOA record is only valid for bare domain"))
		}
	case "SRV":
		check(checkTarget(target))
	case "CAA", "DHCID", "DNSKEY", "DS", "HTTPS", "IMPORT_TRANSFORM", "SSHFP", "SVCB", "TLSA", "TXT":
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
		switch rec.Type {
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
		default:
			// Anything else is ignored.
			continue
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
	err := processSplitHorizonDomains(config)
	if err != nil {
		return []error{err}
	}

	for _, domain := range config.Domains {
		pTypes := []string{}
		for _, provider := range domain.DNSProviderInstances {
			pType := provider.ProviderType
			if pType == "-" {
				// "-" indicates that we don't yet know who the provider type
				// is.  This is probably due to the fact that `dnscontrol
				// check` doesn't read creds.json, which is where the TYPE is
				// set.  We will skip this test in this instance.  Later if
				// `dnscontrol preview` or `push` is used, the full check will
				// be performed.
				continue
			}
			//			// If NO_PURGE is in use, make sure this *isn't* a provider that *doesn't* support NO_PURGE.
			//			if domain.KeepUnknown && providers.ProviderHasCapability(pType, providers.CantUseNOPURGE) {
			//				errs = append(errs, fmt.Errorf("%s uses NO_PURGE which is not supported by %s(%s)", domain.Name, provider.Name, pType))
			//			}
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

			// Canonicalize Label:
			if rec.GetLabel() == (domain.Name + ".") {
				// If label == ${domain}DOT, change to "@"
				rec.SetLabel("@", domain.Name)
			} else if lab, suf := rec.GetLabel(), "."+domain.Name+"."; strings.HasSuffix(lab, suf) {
				// If label ends with DOT${domain}DOT, strip it to a short name.
				rec.SetLabel(lab[:len(lab)-len(suf)], domain.Name)
			}
			// If label ends with dot, add to the list of errors.
			if strings.HasSuffix(rec.GetLabel(), ".") {
				errs = append(errs, fmt.Errorf("label %q does not match D(%q)", rec.GetLabel(), domain.Name))
				return errs // Exit early.
			}

			// in-addr.arpa magic
			if strings.HasSuffix(domain.Name, ".in-addr.arpa") || strings.HasSuffix(domain.Name, ".ip6.arpa") {
				label := rec.GetLabel()
				if strings.HasSuffix(label, "."+domain.Name) {
					rec.SetLabel(label[0:(len(label)-len("."+domain.Name))], domain.Name)
				}
			}

			// Validate the unmodified inputs:
			if err := validateRecordTypes(rec, domain.Name, pTypes); err != nil {
				errs = append(errs, err)
			}
			if err := checkLabel(rec.GetLabel(), rec.Type, domain.Name, rec.Metadata); err != nil {
				errs = append(errs, err)
			}

			if errs2 := checkTargets(rec, domain.Name); errs2 != nil {
				errs = append(errs, errs2...)
			}

			// Canonicalize Targets.
			if rec.Type == "ALIAS" || rec.Type == "CNAME" || rec.Type == "MX" || rec.Type == "NS" || rec.Type == "SRV" {
				// #rtype_variations
				// These record types have a target that is a hostname.
				// We normalize them to a FQDN so there is less variation to handle.  If a
				// provider API requires a shortname, the provider must do the shortening.
				origin := domain.Name + "."
				if rec.SubDomain != "" {
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
			}

			// Populate FQDN:
			rec.SetLabel(rec.GetLabel(), domain.Name)

			if _, ok := rec.Metadata["ignore_name_disable_safety_check"]; ok {
				errs = append(errs, fmt.Errorf("IGNORE_NAME_DISABLE_SAFETY_CHECK no longer supported. Please use DISABLE_IGNORE_SAFETY_CHECK for the entire domain"))
			}

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
				c := config.FindDomain(rec.GetTargetField())
				if c == nil {
					err = fmt.Errorf("IMPORT_TRANSFORM mentions non-existant domain %q", rec.GetTargetField())
					errs = append(errs, err)
				}
				err = importTransform(c, domain, table, rec.TTL)
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
		// Check for different TTLs under the same label
		errs = append(errs, checkRecordSetHasMultipleTTLs(d.Records)...)
		// Validate FQDN consistency
		for _, r := range d.Records {
			if r.NameFQDN == "" || !strings.HasSuffix(r.NameFQDN, d.Name) {
				errs = append(errs, fmt.Errorf("record named '%s' does not have correct FQDN for domain '%s'. FQDN: %s", r.Name, d.Name, r.NameFQDN))
			}
		}
		// Verify AutoDNSSEC is valid.
		errs = append(errs, checkAutoDNSSEC(d)...)
	}

	// At this point we've munged anything that needs to be munged, and
	// validated anything that can be globally validated.
	// Let's ask the provider if there are any records they can't handle.
	for _, domain := range config.Domains { // For each domain..
		for _, provider := range domain.DNSProviderInstances { // For each provider...
			if provider.ProviderBase.ProviderType == "-" {
				// "-" indicates that we don't yet know who the provider type
				// is.  This is probably due to the fact that `dnscontrol
				// check` doesn't read creds.json, which is where the TYPE is
				// set.  We will skip this test in this instance.  Later if
				// `dnscontrol preview` or `push` is used, the full check will
				// be performed.
				continue
			}
			if es := providers.AuditRecords(provider.ProviderBase.ProviderType, domain.Records); len(es) != 0 {
				for _, e := range es {
					errs = append(errs, fmt.Errorf("%s rejects domain %s: %w", provider.ProviderBase.ProviderType, domain.Name, e))
				}
			}
		}
	}

	return errs
}

// processSplitHorizonDomains finds "domain.tld!tag" domains and pre-processes them.
func processSplitHorizonDomains(config *models.DNSConfig) error {
	// Parse out names and tags.
	for _, d := range config.Domains {
		d.UpdateSplitHorizonNames()
	}

	// Verify uniquenames are unique
	seen := map[string]bool{}
	for _, d := range config.Domains {
		uniquename := d.GetUniqueName()
		if seen[uniquename] {
			return fmt.Errorf("duplicate domain name: %q", uniquename)
		}
		seen[uniquename] = true
	}

	return nil
}

func checkAutoDNSSEC(dc *models.DomainConfig) (errs []error) {
	if strings.ToLower(dc.RegistrarName) == "none" {
		return
	}
	if dc.AutoDNSSEC == "on" {
		for providerName := range dc.DNSProviderNames {
			if dc.RegistrarName != providerName {
				errs = append(errs, Warning{fmt.Errorf("AutoDNSSEC is enabled, but DNS provider %s does not match registrar %s", providerName, dc.RegistrarName)})
			}
		}
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
		diffable := fmt.Sprintf("%s %s %s", r.GetLabelFQDN(), r.Type, r.ToComparableNoTTL())
		if seen[diffable] != nil {
			errs = append(errs, fmt.Errorf("exact duplicate record found: %s", diffable))
		}
		seen[diffable] = r
	}
	return errs
}

func checkRecordSetHasMultipleTTLs(records []*models.RecordConfig) (errs []error) {
	// The RFCs say that all records at a particular recordset should have
	// the same TTL.  Most providers don't care, and if they do the
	// dnscontrol provider code usually picks the lowest TTL for all of them.

	// General algorithm:
	// gather all records at a particular label.
	//     has[label] -> ttl -> type(s)
	// for each label, if there is more than one ttl, output ttl:A/TXT ttl:TXT/NS

	// Find the inconsistencies:
	m := make(map[string]map[uint32]map[string]bool)
	for _, r := range records {
		label := r.GetLabelFQDN()
		ttl := r.TTL
		rtype := r.Type

		if _, ok := m[label]; !ok {
			m[label] = make(map[uint32]map[string]bool)
		}
		if _, ok := m[label][ttl]; !ok {
			m[label][ttl] = make(map[string]bool)
		}
		m[label][ttl][rtype] = true
	}

	labels := make([]string, len(m))
	i := 0
	for k := range m {
		labels[i] = k
		i++
	}
	sort.Strings(labels)
	// NB(tlim): No need to de-dup labels. They come from map keys.

	for _, label := range labels {
		if len(m[label]) > 1 {
			// Invert for a more clear error message:
			r := make(map[string]map[uint32]bool)
			for ttl, rtypes := range m[label] {
				for rtype := range rtypes {
					if _, ok := r[rtype]; !ok {
						r[rtype] = make(map[uint32]bool)
					}
					r[rtype][ttl] = true
				}
			}

			// Report any cases where a RecordSet has > 1 different TTLs
			for rtype := range r {
				if len(r[rtype]) > 1 {
					result := formatInconsistency(r)
					errs = append(errs, Warning{fmt.Errorf("inconsistent TTLs at %q: %s", label, result)})
				}
			}
		}
	}

	return errs
}

func formatInconsistency(r map[string]map[uint32]bool) string {
	var rtypeResult []string
	for rtype, ttlsMap := range r {

		ttlList := make([]int, len(ttlsMap))
		i := 0
		for k := range ttlsMap {
			ttlList[i] = int(k)
			i++
		}

		sort.Ints(ttlList)

		rtypeResult = append(rtypeResult, fmt.Sprintf("%s:%v", rtype, commaSepInts(ttlList)))
	}
	sort.Strings(rtypeResult)
	return strings.Join(rtypeResult, " ")
}

func commaSepInts(list []int) string {
	slist := make([]string, len(list))
	for i, v := range list {
		slist[i] = fmt.Sprintf("%d", v)
	}
	return strings.Join(slist, ",")
}

// We pull this out of checkProviderCapabilities() so that it's visible within
// the package elsewhere, so that our test suite can look at the list of
// capabilities we're checking and make sure that it's up-to-date.
var providerCapabilityChecks = []pairTypeCapability{
	// #rtype_variations
	// If a zone uses rType X, the provider must support capability Y.
	//{"X", providers.Y},
	capabilityCheck("AKAMAICDN", providers.CanUseAKAMAICDN),
	capabilityCheck("ALIAS", providers.CanUseAlias),
	capabilityCheck("AUTODNSSEC", providers.CanAutoDNSSEC),
	capabilityCheck("AZURE_ALIAS", providers.CanUseAzureAlias),
	capabilityCheck("CAA", providers.CanUseCAA),
	capabilityCheck("DHCID", providers.CanUseDHCID),
	capabilityCheck("DNAME", providers.CanUseDNAME),
	capabilityCheck("DNSKEY", providers.CanUseDNSKEY),
	capabilityCheck("HTTPS", providers.CanUseHTTPS),
	capabilityCheck("LOC", providers.CanUseLOC),
	capabilityCheck("NAPTR", providers.CanUseNAPTR),
	capabilityCheck("PTR", providers.CanUsePTR),
	capabilityCheck("R53_ALIAS", providers.CanUseRoute53Alias),
	capabilityCheck("SOA", providers.CanUseSOA),
	capabilityCheck("SRV", providers.CanUseSRV),
	capabilityCheck("SSHFP", providers.CanUseSSHFP),
	capabilityCheck("SVCB", providers.CanUseSVCB),
	capabilityCheck("TLSA", providers.CanUseTLSA),

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
			if provider.ProviderType == "-" {
				// "-" indicates that we don't yet know who the provider type
				// is.  This is probably due to the fact that `dnscontrol
				// check` doesn't read creds.json, which is where the TYPE is
				// set.  We will skip this test in this instance.  Later if
				// `dnscontrol preview` or `push` is used, the full check will
				// be performed.
				continue
			}
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
