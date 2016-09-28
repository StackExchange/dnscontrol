package normalize

import (
	"fmt"
	"net"
	"strings"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/transform"
	"github.com/miekg/dns"
	"github.com/miekg/dns/dnsutil"
)

// Returns false if label does not validate.
func assert_no_enddot(label string) error {
	if label == "@" {
		return nil
	}
	if len(label) < 1 {
		return fmt.Errorf("WARNING: null label.")
	}
	if label[len(label)-1] == '.' {
		return fmt.Errorf("WARNING: label (%v) ends with a (.)", label)
	}
	return nil
}

// Returns false if label does not validate.
func assert_no_underscores(label string) error {
	if strings.ContainsRune(label, '_') {
		return fmt.Errorf("WARNING: label (%v) contains an underscore", label)
	}
	return nil
}

// Returns false if target does not validate.
func assert_valid_ipv4(label string) error {
	if net.ParseIP(label).To4() == nil {
		return fmt.Errorf("WARNING: target (%v) is not an IPv4 address", label)
	}
	return nil
}

// Returns false if target does not validate.
func assert_valid_ipv6(label string) error {
	if net.ParseIP(label).To16() == nil {
		return fmt.Errorf("WARNING: target (%v) is not an IPv6 address", label)
	}
	return nil
}

// assert_valid_cname_target returns 1 if target is not valid for cnames.
func assert_valid_target(label string) error {
	if label == "@" {
		return nil
	}
	if len(label) < 1 {
		return fmt.Errorf("WARNING: null label.")
	}
	// If it containts a ".", it must end in a ".".
	if strings.ContainsRune(label, '.') && label[len(label)-1] != '.' {
		return fmt.Errorf("WARNING: label (%v) includes a (.), must end with a (.)", label)
	}
	return nil
}

// validateRecordTypes list of valid rec.Type values. Returns true if this is a real DNS record type, false means it is a pseudo-type used internally.
func validateRecordTypes(rec *models.RecordConfig, domain_name string) error {
	var valid_types = map[string]bool{
		"A":                true,
		"AAAA":             true,
		"CNAME":            true,
		"IMPORT_TRANSFORM": false,
		"MX":               true,
		"TXT":              true,
		"NS":               true,
	}

	if _, ok := valid_types[rec.Type]; !ok {
		return fmt.Errorf("Unsupported record type (%v) domain=%v name=%v", rec.Type, domain_name, rec.Name)
	}
	return nil
}

// validateTargets returns true if rec.Target is valid for the rec.Type.
func validateTargets(rec *models.RecordConfig, domain_name string) (errs []error) {
	label := rec.Name
	target := rec.Target
	check := func(e error) {
		if e != nil {
			errs = append(errs, e)
		}
	}
	switch rec.Type {
	case "A":
		check(assert_no_enddot(label))
		check(assert_no_underscores(label))
		check(assert_valid_ipv4(target))
	case "AAAA":
		check(assert_no_enddot(label))
		check(assert_no_underscores(label))
		check(assert_valid_ipv6(target))
	case "CNAME":
		check(assert_no_enddot(label))
		check(assert_no_underscores(label))
		check(assert_valid_target(target))
	case "MX":
		check(assert_no_enddot(label))
		check(assert_no_underscores(label))
		check(assert_valid_target(target))
	case "NS":
		check(assert_no_enddot(label))
		check(assert_no_underscores(label))
		check(assert_valid_target(target))
	case "TXT", "IMPORT_TRANSFORM":
	default:
		errs = append(errs, fmt.Errorf("Unimplemented record type (%v) domain=%v name=%v",
			rec.Type, domain_name, rec.Name))
	}
	return
}

func transform_cname(target, old_domain, new_domain string) string {
	// Canonicalize. If it isn't a FQDN, add the new_domain.
	result := dnsutil.AddOrigin(target, old_domain)
	if dns.IsFqdn(result) {
		result = result[:len(result)-1]
	}
	return dnsutil.AddOrigin(result, new_domain) + "."
}

// import_transform imports the records of one zone into another, modifying records along the way.
func import_transform(src_domain, dst_domain *models.DomainConfig, transforms []transform.IpConversion) error {
	// Read src_domain.Records, transform, and append to dst_domain.Records:
	// 1. Skip any that aren't A or CNAMEs.
	// 2. Append dest_domainname to the end of the label.
	// 3. For CNAMEs, append dest_domainname to the end of the target.
	// 4. For As, change the target as described the transforms.

	for _, rec := range src_domain.Records {
		newRec := func() *models.RecordConfig {
			rec2, _ := rec.Copy()
			rec2.Name = rec2.NameFQDN
			rec2.NameFQDN = dnsutil.AddOrigin(rec2.Name, dst_domain.Name)
			rec2.TTL = 60
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
				dst_domain.Records = append(dst_domain.Records, r)
			}
		case "CNAME":
			r := newRec()
			r.Target = transform_cname(r.Target, src_domain.Name, dst_domain.Name)
			dst_domain.Records = append(dst_domain.Records, r)
		case "MX", "NS", "TXT":
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

func NormalizeAndValidateConfig(config *models.DNSConfig) (errs []error) {
	// TODO(tlim): Before any changes are made, we should check the rules
	// such as MX/CNAME/NS .Target must be a single word, "@", or FQDN.
	// Validate and normalize
	for _, domain := range config.Domains {

		// Normalize Nameservers.
		for _, ns := range domain.Nameservers {
			ns.Name = dnsutil.AddOrigin(ns.Name, domain.Name)
			ns.Name = strings.TrimRight(ns.Name, ".")
		}

		// Normalize Records.
		for _, rec := range domain.Records {

			// Validate the unmodified inputs:
			if err := validateRecordTypes(rec, domain.Name); err != nil {
				errs = append(errs, err)
			}
			if errs2 := validateTargets(rec, domain.Name); errs2 != nil {
				errs = append(errs, errs2...)
			}

			// Canonicalize Targets.
			if rec.Type == "CNAME" || rec.Type == "MX" || rec.Type == "NS" {
				rec.Target = dnsutil.AddOrigin(rec.Target, domain.Name+".")
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
				err = import_transform(config.FindDomain(rec.Target), domain, table)
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

	// Verify all labels are FQDN ending with ".":
	for _, domain := range config.Domains {
		for _, rec := range domain.Records {
			// .Name must NOT end in "."
			if rec.Name[len(rec.Name)-1] == '.' {
				errs = append(errs, fmt.Errorf("Should not happen: Label ends with '.': %v %v %v %v %v",
					domain.Name, rec.Name, rec.NameFQDN, rec.Type, rec.Target))
			}
			// .NameFQDN must NOT end in "."
			if rec.NameFQDN[len(rec.NameFQDN)-1] == '.' {
				errs = append(errs, fmt.Errorf("Should not happen: FQDN ends with '.': %v %v %v %v %v",
					domain.Name, rec.Name, rec.NameFQDN, rec.Type, rec.Target))
			}
			// .Target MUST end in "."
			if rec.Type == "CNAME" || rec.Type == "NS" || rec.Type == "MX" {
				if rec.Target[len(rec.Target)-1] != '.' {
					errs = append(errs, fmt.Errorf("Should not happen: Target does NOT ends with '.': %v %v %v %v %v",
						domain.Name, rec.Name, rec.NameFQDN, rec.Type, rec.Target))
				}
			}
		}
	}
	return errs
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
