package octoyaml

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/miekg/dns/dnsutil"
	yaml "gopkg.in/yaml.v2"
)

// WriteYaml outputs a yaml version of a list of RecordConfig.
func WriteYaml(w io.Writer, records models.Records, origin string) error {
	if len(records) == 0 {
		return nil
	}

	// Pick the most common TTL as the default so we can
	// write the fewest "ttl:" lines.
	defaultTTL := mostCommonTTL(records)

	// Make a copy of the records, since we want to sort and muck with them.
	recsCopy := models.Records{}
	recsCopy = append(recsCopy, records...)
	for _, r := range recsCopy {
		if r.GetLabel() == "@" {
			//r.Name = ""
			r.UnsafeSetLabelNull()
		}
	}

	z := &genYamlData{
		Origin:     dnsutil.AddOrigin(origin, "."),
		DefaultTTL: defaultTTL,
		Records:    recsCopy,
	}

	// Sort in the weird order that OctoDNS expects:
	sort.Sort(z)

	// Generate the YAML:
	fmt.Fprintln(w, "---")
	yb, err := yaml.Marshal(z.genInterfaceList(w))
	if err != nil {
		return err
	}
	_, err = w.Write(yb)

	return err
}

// genInterfaceList outputs YAML ordered slices for the entire zone.
// Each item in the list is an interface that will MarshallYAML to
// the desired output.
func (z *genYamlData) genInterfaceList(w io.Writer) yaml.MapSlice {
	var yam yaml.MapSlice
	// Group the records by label.
	order, groups := z.Records.GroupedByLabel()
	// For each group, generate the YAML.
	for _, label := range order {
		group := groups[label]
		// Within the group, sort the similar Types together:
		sort.SliceStable(group, func(i, j int) bool { return zoneRrtypeLess(group[i].Type, group[j].Type) })
		// Generate the YAML records:
		yam = append(yam, oneLabel(group))
	}
	return yam
}

// "simple" records are when a label has a single rtype.
// It may have a single (simple) or multiple (many) values.

// Used to generate:
//  label:
//    type: A
//    value: 1.2.3.4
type simple struct {
	TTL   uint32 `yaml:"ttl,omitempty"`
	Type  string `yaml:"type"`
	Value string `yaml:"value"`
}

// Used to generate:
//  label:
//    type: A
//    values:
//    - 1.2.3.4
//    - 2.3.4.5
type many struct {
	TTL    uint32   `yaml:"ttl,omitempty"`
	Type   string   `yaml:"type"`
	Values []string `yaml:"values"`
}

// complexItems are when a single label has multiple rtypes
// associated with it. For example, a label with both an A and MX record.
type complexItems []interface{}

// Used to generate a complex item with either a single value or multiple values:
// 'thing':                             >> complexVals
//   - type: CNAME
//     value: newplace.example.com.     << value
// 'www':
//   - type: A
//     values:
//       - 1.2.3.4                      << values
//       - 1.2.3.5                      << values
//   - type: MX
//     values:
//       - priority: 10                 << fields
//         value: mx1.example.com.      << fields
//       - priority: 10                 << fields
//         value: mx2.example.com.      << fields
type complexVals struct {
	TTL    uint32   `yaml:"ttl,omitempty"`
	Type   string   `yaml:"type"`
	Value  string   `yaml:"value,omitempty"`
	Values []string `yaml:"values,omitempty"`
}

// Used to generate rtypes like MX rand SRV ecords, which have multiple
// fields within the rtype.
type complexFields struct {
	TTL    uint32   `yaml:"ttl,omitempty"`
	Type   string   `yaml:"type"`
	Fields []fields `yaml:"values,omitempty"`
}

// Used to generate the fields themselves:
type fields struct {
	Priority  uint16 `yaml:"priority,omitempty"`
	SrvWeight uint16 `yaml:"weight,omitempty"`
	SrvPort   uint16 `yaml:"port,omitempty"`
	Value     string `yaml:"value,omitempty"`
}

// FIXME(tlim): An MX record with .Priority=0 will not output the priority.

// sameType returns true if all records have the same type.
func sameType(records models.Records) bool {
	t := records[0].Type
	for _, r := range records {
		if r.Type != t {
			return false
		}
	}
	return true
}

// oneLabel handles all the DNS records associated with a single label.
// It dispatches the right code whether the label is simple, many, or complex.
func oneLabel(records models.Records) yaml.MapItem {
	item := yaml.MapItem{
		// a yaml.MapItem is a YAML map that retains the key order.
		Key: records[0].GetLabel(),
	}
	//  Special case labels with a single record:
	if len(records) == 1 {
		switch rtype := records[0].Type; rtype {
		case "A", "CNAME", "NS", "PTR", "TXT":
			v := simple{
				Type:  rtype,
				Value: records[0].GetTargetField(),
				TTL:   records[0].TTL,
			}
			if v.Type == "TXT" {
				v.Value = strings.Replace(models.StripQuotes(v.Value), `;`, `\;`, -1)
			}
			//fmt.Printf("yamlwrite:oneLabel: simple ttl=%d\n", v.TTL)
			item.Value = v
			//fmt.Printf("yamlwrite:oneLabel: SIMPLE=%v\n", item)
			return item
		case "MX", "SRV":
			// Always processed as a complex{}
		default:
			panic(fmt.Errorf("yamlwrite:oneLabel:len1 rtype not implemented: %s", rtype))
		}
	}

	//  Special case labels with many records, all the same rType:
	if sameType(records) {
		switch rtype := records[0].Type; rtype {
		case "A", "CNAME", "NS":
			v := many{
				Type: rtype,
				TTL:  records[0].TTL,
			}
			for _, rec := range records {
				v.Values = append(v.Values, rec.GetTargetField())
			}
			item.Value = v
			//fmt.Printf("SIMPLE=%v\n", item)
			return item
		case "MX", "SRV":
			// Always processed as a complex{}
		default:
			panic(fmt.Errorf("oneLabel:many rtype not implemented: %s", rtype))
		}
	}

	// All other labels are complexItems

	var low int // First index of a run.
	var lst complexItems
	var last = records[0].Type
	for i := range records {
		if records[i].Type != last {
			//fmt.Printf("yamlwrite:oneLabel: Calling oneType( [%d:%d] ) last=%s type=%s\n", low, i, last, records[0].Type)
			lst = append(lst, oneType(records[low:i]))
			low = i // Current is the first of a run.
			last = records[i].Type
		}
		if i == (len(records) - 1) {
			// we are on the last element.
			//fmt.Printf("yamlwrite:oneLabel: Calling oneType( [%d:%d] ) last=%s type=%s\n", low, i+1, last, records[0].Type)
			lst = append(lst, oneType(records[low:i+1]))
		}
	}
	item.Value = lst

	return item
}

// oneType returns interfaces that will MarshalYAML properly for a label with
// one or more records, all the same rtype.
func oneType(records models.Records) interface{} {
	//fmt.Printf("yamlwrite:oneType len=%d type=%s\n", len(records), records[0].Type)
	rtype := records[0].Type
	switch rtype {
	case "A", "AAAA", "NS":
		vv := complexVals{
			Type: rtype,
			TTL:  records[0].TTL,
		}
		if len(records) == 1 {
			vv.Value = records[0].GetTargetField()
		} else {
			for _, rc := range records {
				vv.Values = append(vv.Values, rc.GetTargetCombined())
			}
		}
		return vv
	case "MX":
		vv := complexFields{
			Type: rtype,
			TTL:  records[0].TTL,
		}
		for _, rc := range records {
			vv.Fields = append(vv.Fields, fields{
				Value:    rc.GetTargetField(),
				Priority: rc.MxPreference,
			})
		}
		return vv
	case "SRV":
		vv := complexFields{
			Type: rtype,
			TTL:  records[0].TTL,
		}
		for _, rc := range records {
			vv.Fields = append(vv.Fields, fields{
				Value:     rc.GetTargetField(),
				Priority:  rc.SrvPriority,
				SrvWeight: rc.SrvWeight,
				SrvPort:   rc.SrvPort,
			})
		}
		return vv
	case "TXT":
		vv := complexVals{
			Type: rtype,
			TTL:  records[0].TTL,
		}
		if len(records) == 1 {
			vv.Value = strings.Replace(models.StripQuotes(records[0].GetTargetField()), `;`, `\;`, -1)
		} else {
			for _, rc := range records {
				vv.Values = append(vv.Values, models.StripQuotes(rc.GetTargetCombined()))
			}
		}
		return vv

	default:
		panic(fmt.Errorf("yamlwrite:oneType rtype=%s not implemented", rtype))
	}
}

// mostCommonTTL returns the most common TTL in a set of records. If there is
// a tie, the highest TTL is selected. This makes the results consistent.
// NS records are not included in the analysis because Tom said so.
func mostCommonTTL(records models.Records) uint32 {
	// Index the TTLs in use:
	d := make(map[uint32]int)
	for _, r := range records {
		if r.Type != "NS" {
			d[r.TTL]++
		}
	}
	// Find the largest count:
	var mc int
	for _, value := range d {
		if value > mc {
			mc = value
		}
	}
	// Find the largest key with that count:
	var mk uint32
	for key, value := range d {
		if value == mc {
			if key > mk {
				mk = key
			}
		}
	}
	return mk
}
