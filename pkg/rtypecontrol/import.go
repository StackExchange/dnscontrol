package rtypecontrol

import (
	"fmt"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/domaintags"
	"github.com/miekg/dns"
)

// ImportRawRecords imports the RawRecordConfigs into RecordConfigs.
func ImportRawRecords(domains []*models.DomainConfig) error {
	for _, dc := range domains {
		for _, rawRec := range dc.RawRecords {

			rec, err := NewRecordConfigFromRaw(rawRec.Type, rawRec.TTL, rawRec.Args, dc.DomainNameVarieties(), models.FixPosition(rawRec.FilePos))
			if err != nil {
				return err
			}
			if rec.Metadata["skip_fqdn_check"] != "true" && stutters(rec.Name, dc.Name) {
				var shortname string
				if rec.Name == dc.Name {
					shortname = "@"
				} else {
					shortname = strings.TrimSuffix(rec.Name, "."+dc.Name)
				}
				return fmt.Errorf(
					"The name %q is an error (repeats the domain). Maybe instead of %q you intended %q? If not add DISABLE_REPEATED_DOMAIN_CHECK to this record to disable this check",
					rec.NameFQDNRaw,
					rec.NameRaw,
					shortname,
				)
			}

			// Free memeory:
			clear(rawRec.Args)
			rawRec.Args = nil

			dc.Records = append(dc.Records, rec)
		}
		dc.RawRecords = nil
	}

	return nil
}

func stutters(name, domain string) bool {
	if name == "@" {
		return false
	}
	if name == domain || strings.HasSuffix(name, "."+domain) {
		return true
	}
	return false
}

// NewRecordConfigFromRaw creates a new RecordConfig from the raw ([]any) args,
// usually from the parsed dnsconfig.js file, but also useful when a provider
// returns the fields of a record as individual values.
func NewRecordConfigFromRaw(t string, ttl uint32, args []any, dcn *domaintags.DomainNameVarieties, FilePos string) (*models.RecordConfig, error) {
	if _, ok := Func[t]; !ok {
		return nil, fmt.Errorf("record type %q is not supported", t)
	}
	if t == "" {
		panic("rtypecontrol: NewRecordConfigFromRaw: empty record type")
	}

	// Create as much of the RecordConfig as we can now. Allow New() to fill in the reset.
	rec := &models.RecordConfig{
		Type:     t,
		TTL:      ttl,
		Metadata: map[string]string{},
		FilePos:  FilePos,
	}
	if err := setRecordNames(rec, dcn, args[0].(string)); err != nil {
		return rec, err
	}
	if strings.HasSuffix(rec.Name, ".") {
		return nil, fmt.Errorf("label %q is not in zone %s", args[0].(string), dcn.DisplayName)
	}

	// Fill in the .F/.Fields* fields.
	err := Func[t].FromArgs(dcn, rec, args)
	if err != nil {
		return nil, err
	}

	return rec, nil
}

// NewRecordConfigFromString creates a new RecordConfig from a string in the
// format usually used in a zonefile but typically also used by providers
// returning the fields of a record as a string.
func NewRecordConfigFromString(name string, ttl uint32, t string, s string, dcn *domaintags.DomainNameVarieties) (*models.RecordConfig, error) {
	if _, ok := Func[t]; !ok {
		return nil, fmt.Errorf("record type %q is not supported", t)
	}
	if t == "" {
		panic("rtypecontrol: NewRecordConfigFromStruct: empty record type")
	}

	rec, err := dns.NewRR(fmt.Sprintf("$ORIGIN .\n. %d IN %s %s", ttl, t, s))
	if err != nil {
		return nil, err
	}
	return NewRecordConfigFromStruct(name, ttl, t, rec, dcn)

}

// NewRecordConfigFromStruct creates a new RecordConfig from a struct, typically
// a miekg/dns struct. It must be the exact struct type used by the FromStruct()
// method of the rtype package.
func NewRecordConfigFromStruct(name string, ttl uint32, t string, fields any, dcn *domaintags.DomainNameVarieties) (*models.RecordConfig, error) {
	if _, ok := Func[t]; !ok {
		return nil, fmt.Errorf("record type %q is not supported", t)
	}
	if t == "" {
		panic("rtypecontrol: NewRecordConfigFromStruct: empty record type")
	}

	// Create as much of the RecordConfig as we can now. Allow New() to fill in the reset.
	rec := &models.RecordConfig{
		Type:     t,
		TTL:      ttl,
		Metadata: map[string]string{},
	}
	if err := setRecordNames(rec, dcn, name); err != nil {
		return rec, err
	}
	if strings.HasSuffix(rec.Name, ".") {
		return nil, fmt.Errorf("label %q is not in zone %q", name, dcn.NameASCII+".")
	}

	err := Func[t].FromStruct(dcn, rec, name, fields)
	if err != nil {
		return nil, err
	}

	return rec, nil
}
