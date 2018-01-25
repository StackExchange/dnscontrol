package octoyaml

import (
	"fmt"
	"io"
	"sort"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/miekg/dns/dnsutil"
	yaml "gopkg.in/yaml.v2"
)

// WriteYaml outputs a yaml version of a list of RecordConfig.
func WriteYaml(w io.Writer, records models.Records, origin string) error {
	defaultTTL := mostCommonTTL(records)
	//fmt.Printf("DEBUG: defaultTTL=%d\n", defaultTTL)

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
	Type  string `yaml:"type"`
	Value string `yaml:"value"`
	TTL   uint32 `yaml:"ttl,omitempty"`
}

type many struct {
	Type   string   `yaml:"type"`
	Values []string `yaml:"values"`
	TTL    uint32   `yaml:"ttl,omitempty"`
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
	var item yaml.MapItem

	// Single-item labels:

	item.Key = records[0].Name

	if len(records) == 1 {
		switch rtype := records[0].Type; rtype {
		case "A", "CNAME", "NS":
			v := simple{
				Type:  rtype,
				Value: records[0].Content(),
				TTL:   records[0].TTL,
			}
			item.Value = v
		default:
			e := fmt.Errorf("oneLabel:len1 rtype not implemented: %s", rtype)
			panic(e)
		}
		return item
	}

	if sameType(records) {
		switch rtype := records[0].Type; rtype {
		case "A", "CNAME", "NS":
			v := many{
				Type: rtype,
				TTL:  records[0].TTL,
			}
			for _, rec := range records {
				v.Values = append(v.Values, rec.Content())
			}
			item.Value = v
		default:
			e := fmt.Errorf("oneLabel:many rtype not implemented: %s", rtype)
			panic(e)
		}
		return item
	}

	// This is a single label with multiple rTypes within it.

	var low int // First index of a run.
	var lst complexItems
	var last = records[0].Type
	for i := range records {
		if records[i].Type != last {
			//fmt.Printf("DEBUG: Calling oneType( [%d:%d] ) last=%s type=%s\n", low, i, last, records[0].Type)
			lst = append(lst, oneType(records[low:i]))
			low = i // Current is the first of a run.
			last = records[i].Type
		}
		if i == (len(records) - 1) {
			// we are on the last element.
			//fmt.Printf("DEBUG: Calling oneType( [%d:%d] ) last=%s type=%s\n", low, i+1, last, records[0].Type)
			lst = append(lst, oneType(records[low:i+1]))
		}
	}
	item.Value = lst

	return item
}

type complexItems []interface{}
type complexVals struct {
	Type   string   `yaml:"type"`
	TTL    uint32   `yaml:"ttl,omitempty"`
	Value  string   `yaml:"value,omitempty"`
	Values []string `yaml:"values,omitempty"`
}
type complexFields struct {
	Type   string   `yaml:"type"`
	TTL    uint32   `yaml:"ttl,omitempty"`
	Fields []fields `yaml:"values,omitempty"`
}
type fields struct {
	Priority uint16 `yaml:"priority"`
	Value    string `yaml:"value"`
}

func oneType(records models.Records) interface{} {
	//fmt.Printf("DEBUG: oneType len=%d type=%s\n", len(records), records[0].Type)
	rtype := records[0].Type
	var v interface{}
	switch rtype {
	case "A":
		vv := complexVals{
			Type: rtype,
			TTL:  records[0].TTL,
		}
		if len(records) == 1 {
			vv.Value = records[0].Target
		} else {
			for _, rc := range records {
				vv.Values = append(vv.Values, rc.Content())
			}
		}
		v = vv
	case "MX":
		vv := complexFields{
			Type: rtype,
			TTL:  records[0].TTL,
		}
		for _, rc := range records {
			vv.Fields = append(vv.Fields, fields{
				Priority: rc.MxPreference,
				Value:    rc.Target,
			})
		}
		v = vv
	default:
		panic("oneType not implemented")
	}

	return v
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
