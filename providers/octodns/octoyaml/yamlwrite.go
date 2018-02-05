package octoyaml

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/miekg/dns/dnsutil"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

// WriteYaml outputs a yaml version of a list of RecordConfig.
func WriteYaml(w io.Writer, records models.Records, origin string) error {
	if len(records) == 0 {
		return nil
	}

	defaultTTL := mostCommonTTL(records)

	recsCopy := models.Records{}
	for _, r := range records {
		recsCopy = append(recsCopy, r)
	}
	for _, r := range recsCopy {
		if r.Name == "@" {
			r.Name = ""
		}
	}

	z := &genYamlData{
		Origin:     dnsutil.AddOrigin(origin, "."),
		DefaultTTL: defaultTTL,
		Records:    recsCopy,
	}
	sort.Sort(z)

	fmt.Fprintln(w, "---")
	yb, err := yaml.Marshal(z.labelRanges(w))
	if err != nil {
		return err
	}
	_, err = w.Write(yb)

	return err
}

func (z *genYamlData) labelRanges(w io.Writer) yaml.MapSlice {
	var yam yaml.MapSlice
	// Group the records by label
	order, groups := z.Records.GroupedByLabel()
	for _, label := range order {
		group := groups[label]
		// Within the group, sort the similar Types together:
		sort.SliceStable(group, func(i, j int) bool { return zoneRrtypeLess(group[i].Type, group[j].Type) })
		// Generate the YAML records:
		yam = append(yam, oneLabel(group))
	}
	return yam
}

type simple struct {
	TTL   uint32 `yaml:"ttl,omitempty"`
	Type  string `yaml:"type"`
	Value string `yaml:"value"`
}

type many struct {
	TTL    uint32   `yaml:"ttl,omitempty"`
	Type   string   `yaml:"type"`
	Values []string `yaml:"values"`
}

func sameType(records models.Records) bool {
	t := records[0].Type
	for _, r := range records {
		if r.Type != t {
			return false
		}
	}
	return true
}

func oneLabel(records models.Records) yaml.MapItem {

	item := yaml.MapItem{
		// a yaml.MapItem is a YAML map that retains the key order.
		Key: records[0].Name,
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
			e := fmt.Errorf("yamlwrite:oneLabel:len1 rtype not implemented: %s", rtype)
			panic(e)
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
			e := fmt.Errorf("oneLabel:many rtype not implemented: %s", rtype)
			panic(e)
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

type complexItems []interface{}
type complexVals struct {
	TTL    uint32   `yaml:"ttl,omitempty"`
	Type   string   `yaml:"type"`
	Value  string   `yaml:"value,omitempty"`
	Values []string `yaml:"values,omitempty"`
}
type complexFields struct {
	TTL    uint32   `yaml:"ttl,omitempty"`
	Type   string   `yaml:"type"`
	Fields []fields `yaml:"values,omitempty"`
}
type fields struct {
	Priority  uint16 `yaml:"priority,omitempty"`
	SrvWeight uint16 `yaml:"weight,omitempty"`
	SrvPort   uint16 `yaml:"port,omitempty"`
	Value     string `yaml:"value,omitempty"`
}

// FIXME(tlim): An MX record with .Priority=0 will not output the priority.

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
		e := errors.Errorf("yamlwrite:oneType rtype=%s not implemented", rtype)
		fmt.Println(e)
		panic(e)
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
