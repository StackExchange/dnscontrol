package normalize

import (
	"fmt"
	"net"
	"strings"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/pkg/transform"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/miekg/dns"
	"github.com/miekg/dns/dnsutil"
	"github.com/pkg/errors"
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
		return errors.Errorf("target (%v) includes invalid char", target)
	}
	// If it containts a ".", it must end in a ".".
	if strings.ContainsRune(target, '.') && target[len(target)-1] != '.' {
		return fmt.Errorf("target (%v) must end with a (.) [Required if target is not single label]", target)
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
		"IMPORT_TRANSFORM": false,
		"MX":               true,
		"SRV":              true,
		"TXT":              true,
		"NS":               true,
		"PTR":              true,
		"ALIAS":            false,
	}
	_, ok := validTypes[rec.Type]
	if !ok {
		cType := providers.GetCustomRecordType(rec.Type)
		if cType == nil {
			return fmt.Errorf("Unsupported record type (%v) domain=%v name=%v", rec.Type, domain, rec.Name)
		}
		for _, providerType := range pTypes {
			if providerType != cType.Provider {
				return fmt.Errorf("Custom record type %s is not compatible with provider type %s", rec.Type, providerType)
			}
		}
		//it is ok. Lets replace the type with real type and add metadata to say we checked it
		rec.Metadata["orig_custom_type"] = rec.Type
		if cType.RealType != "" {
			rec.Type = cType.RealType
		}
	}
	return nil
}

// underscores in names are often used erroneously. They are valid for dns records, but invalid for urls.
// here we list common records expected to have underscores. Anything else containing an underscore will print a warning.
var expectedUnderscores = []string{"_domainkey", "_dmarc", "_amazonses"}

func checkLabel(label string, rType string, domain string) error {
	if label == "@" {
		return nil
	}
	if len(label) < 1 {
		return fmt.Errorf("empty %s label in %s", rType, domain)
	}
	if label[len(label)-1] == '.' {
		return fmt.Errorf("label %s.%s ends with a (.)", label, domain)
	}

	//underscores are warnings
	if rType != "SRV" && strings.ContainsRune(label, '_') {
		//unless it is in our exclusion list
		ok := false
		for _, ex := range expectedUnderscores {
			if strings.Contains(label, ex) {
				ok = true
				break
			}
		}
		if !ok {
			return Warning{fmt.Errorf("label %s.%s contains an underscore", label, domain)}
		}
	}
	return nil
}

// checkTargets returns true if rec.Target is valid for the rec.Type.
func checkTargets(rec *models.RecordConfig, domain string) (errs []error) {
	label := rec.Name
	target := rec.Target
	check := func(e error) {
		if e != nil {
			err := fmt.Errorf("In %s %s.%s: %s", rec.Type, rec.Name, domain, e.Error())
			if _, ok := e.(Warning); ok {
				err = Warning{err}
			}
			errs = append(errs, err)
		}
	}
	switch rec.Type {
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
	case "ALIAS":
		check(checkTarget(target))
	case "SRV":
		check(checkTarget(target))
	case "TXT", "IMPORT_TRANSFORM", "CAA":
	default:
		if rec.Metadata["orig_custom_type"] != "" {
			//it is a valid custom type. We perform no validation on target
			return
		}
		errs = append(errs, fmt.Errorf("checkTargets: Unimplemented record type (%v) domain=%v name=%v",
			rec.Type, domain, rec.Name))
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
		if dstDomain.HasRecordTypeName(rec.Type, rec.NameFQDN) {
			continue
		}
		newRec := func() *models.RecordConfig {
			rec2, _ := rec.Copy()
			rec2.Name = rec2.NameFQDN
			rec2.NameFQDN = dnsutil.AddOrigin(rec2.Name, dstDomain.Name)
			if ttl != 0 {
				rec2.TTL = ttl
			}
			return rec2
		}
		switch rec.Type {
		case "A":
			trs, err := transform.TransformIPToList(net.ParseIP(rec.Target), transforms)
			if err != nil {
				return fmt.Errorf("import_transform: TransformIP(%v, %v) returned err=%s", rec.Target, transforms, err)
			}
			for _, tr := range trs {
				r := newRec()
				r.Target = tr.String()
				dstDomain.Records = append(dstDomain.Records, r)
			}
		case "CNAME":
			r := newRec()
			r.Target = transformCNAME(r.Target, srcDomain.Name, dstDomain.Name)
			dstDomain.Records = append(dstDomain.Records, r)
		case "MX", "NS", "SRV", "TXT", "CAA":
			// Not imported.
			continue
		default:
			return fmt.Errorf("import_transform: Unimplemented record type %v (%v)",
				rec.Type, rec.Name)
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

func NormalizeAndValidateConfig(config *models.DNSConfig) (errs []error) {
	ptypeMap := map[string]string{}
	for _, p := range config.DNSProviders {
		ptypeMap[p.Name] = p.Type
	}

	for _, domain := range config.Domains {
		pTypes := []string{}
		for p := range domain.DNSProviders {
			pType, ok := ptypeMap[p]
			if !ok {
				errs = append(errs, fmt.Errorf("%s uses undefined DNS provider %s", domain.Name, p))
			} else {
				pTypes = append(pTypes, pType)
			}
		}

		// Normalize Nameservers.
		for _, ns := range domain.Nameservers {
			ns.Name = dnsutil.AddOrigin(ns.Name, domain.Name)
			ns.Name = strings.TrimRight(ns.Name, ".")
		}
		// Normalize Records.
		for _, rec := range domain.Records {
			if rec.TTL == 0 {
				rec.TTL = models.DefaultTTL
			}
			// Validate the unmodified inputs:
			if err := validateRecordTypes(rec, domain.Name, pTypes); err != nil {
				errs = append(errs, err)
			}
			if err := checkLabel(rec.Name, rec.Type, domain.Name); err != nil {
				errs = append(errs, err)
			}
			if errs2 := checkTargets(rec, domain.Name); errs2 != nil {
				errs = append(errs, errs2...)
			}

			// Canonicalize Targets.
			if rec.Type == "CNAME" || rec.Type == "MX" || rec.Type == "NS" {
				rec.Target = dnsutil.AddOrigin(rec.Target, domain.Name+".")
			} else if rec.Type == "A" || rec.Type == "AAAA" {
				rec.Target = net.ParseIP(rec.Target).String()
			} else if rec.Type == "PTR" {
				var err error
				if rec.Name, err = transform.PtrNameMagic(rec.Name, domain.Name); err != nil {
					errs = append(errs, err)
				}
			} else if rec.Type == "CAA" {
				if rec.CaaTag != "issue" && rec.CaaTag != "issuewild" && rec.CaaTag != "iodef" {
					errs = append(errs, fmt.Errorf("CAA tag %s is invalid", rec.CaaTag))
				}
			}
			// Populate FQDN:
			rec.NameFQDN = dnsutil.AddOrigin(rec.Name, domain.Name)
		}
	}

	// Process any pseudo-records:
	for _, domain := range config.Domains {
		for _, rec := range domain.Records {
			if rec.Type == "IMPORT_TRANSFORM" {
				table, err := transform.DecodeTransformTable(rec.Metadata["transform_table"])
				if err != nil {
					errs = append(errs, err)
					continue
				}
				err = importTransform(config.FindDomain(rec.Target), domain, table, rec.TTL)
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

	//Check that CNAMES don't have to co-exist with any other records
	for _, d := range config.Domains {
		errs = append(errs, checkCNAMEs(d)...)
	}

	//Check that if any aliases / ptr / etc.. are used in a domain, every provider for that domain supports them
	for _, d := range config.Domains {
		err := checkProviderCapabilities(d, config.DNSProviders)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func checkCNAMEs(dc *models.DomainConfig) (errs []error) {
	cnames := map[string]bool{}
	for _, r := range dc.Records {
		if r.Type == "CNAME" {
			if cnames[r.Name] {
				errs = append(errs, fmt.Errorf("Cannot have multiple CNAMEs with same name: %s", r.NameFQDN))
			}
			cnames[r.Name] = true
		}
	}
	for _, r := range dc.Records {
		if cnames[r.Name] && r.Type != "CNAME" {
			errs = append(errs, fmt.Errorf("Cannot have CNAME and %s record with same name: %s", r.Type, r.NameFQDN))
		}
	}
	return
}

func checkProviderCapabilities(dc *models.DomainConfig, pList []*models.DNSProviderConfig) error {
	types := []struct {
		rType string
		cap   providers.Capability
	}{
		{"ALIAS", providers.CanUseAlias},
		{"PTR", providers.CanUsePTR},
		{"SRV", providers.CanUseSRV},
		{"CAA", providers.CanUseCAA},
	}
	for _, ty := range types {
		hasAny := false
		for _, r := range dc.Records {
			if r.Type == ty.rType {
				hasAny = true
				break
			}
		}
		if !hasAny {
			continue
		}
		for pName := range dc.DNSProviders {
			for _, p := range pList {
				if p.Name == pName {
					if !providers.ProviderHasCabability(p.Type, ty.cap) {
						return fmt.Errorf("Domain %s uses %s records, but DNS provider type %s does not support them", dc.Name, ty.rType, p.Type)
					}
					break
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
		ip := net.ParseIP(rec.Target) //ip already validated above
		newIPs, err := transform.TransformIPToList(net.ParseIP(rec.Target), table)
		if err != nil {
			return err
		}
		for i, newIP := range newIPs {
			if i == 0 && !newIP.Equal(ip) {
				rec.Target = newIP.String() //replace target of first record if different
			} else if i > 0 {
				// any additional ips need identical records with the alternate ip added to the domain
				copy, err := rec.Copy()
				if err != nil {
					return err
				}
				copy.Target = newIP.String()
				domain.Records = append(domain.Records, copy)
			}
		}
	}
	return nil
}
