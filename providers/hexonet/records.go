package hexonet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
)

// HXRecord covers an individual DNS resource record.
type HXRecord struct {
	// Raw api value of that RR
	Raw string
	// DomainName is the zone that the record belongs to.
	DomainName string
	// Host is the hostname relative to the zone: e.g. for a record for blog.example.org, domain would be "example.org" and host would be "blog".
	// An apex record would be specified by either an empty host "" or "@".
	// A SRV record would be specified by "_{service}._{protocol}.{host}": e.g. "_sip._tcp.phone" for _sip._tcp.phone.example.org.
	Host string
	// FQDN is the Fully Qualified Domain Name. It is the combination of the host and the domain name. It always ends in a ".". FQDN is ignored in CreateRecord, specify via the Host field instead.
	Fqdn string
	// Type is one of the following: A, AAAA, ANAME, CNAME, MX, NS, SRV, or TXT.
	Type string
	// Answer is either the IP address for A or AAAA records; the target for ANAME, CNAME, MX, or NS records; the text for TXT records.
	// For SRV records, answer has the following format: "{weight} {port} {target}" e.g. "1 5061 sip.example.org".
	Answer string
	// TTL is the time this record can be cached for in seconds.
	TTL uint32
	// Priority is only required for MX and SRV records, it is ignored for all others.
	Priority uint32
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (n *HXClient) GetZoneRecords(domain string) (models.Records, error) {
	records, err := n.getRecords(domain)
	if err != nil {
		return nil, err
	}
	actual := make([]*models.RecordConfig, len(records))
	for i, r := range records {
		actual[i] = toRecord(r, domain)
	}

	for _, rec := range actual {
		if rec.Type == "ALIAS" {
			return nil, fmt.Errorf("we support realtime ALIAS RR over our X-DNS service, please get in touch with us")
		}
	}

	return actual, nil

}

// GetDomainCorrections gathers correctios that would bring n to match dc.
func (n *HXClient) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	dc.Punycode()

	actual, err := n.GetZoneRecords(dc.Name)
	if err != nil {
		return nil, err
	}

	//checkNSModifications(dc)

	// Normalize
	models.PostProcessRecords(actual)

	differ := diff.New(dc)
	_, create, del, mod, err := differ.IncrementalDiff(actual)
	if err != nil {
		return nil, err
	}

	corrections := []*models.Correction{}

	buf := &bytes.Buffer{}
	// Print a list of changes. Generate an actual change that is the zone
	changes := false
	params := map[string]interface{}{}
	delrridx := 0
	addrridx := 0
	for _, cre := range create {
		changes = true
		fmt.Fprintln(buf, cre)
		rec := cre.Desired
		recordString, err := n.createRecordString(rec, dc.Name)
		if err != nil {
			return corrections, err
		}
		params[fmt.Sprintf("ADDRR%d", addrridx)] = recordString
		addrridx++
	}
	for _, d := range del {
		changes = true
		fmt.Fprintln(buf, d)
		rec := d.Existing.Original.(*HXRecord)
		params[fmt.Sprintf("DELRR%d", delrridx)] = n.deleteRecordString(rec, dc.Name)
		delrridx++
	}
	for _, chng := range mod {
		changes = true
		fmt.Fprintln(buf, chng)
		old := chng.Existing.Original.(*HXRecord)
		new := chng.Desired
		params[fmt.Sprintf("DELRR%d", delrridx)] = n.deleteRecordString(old, dc.Name)
		newRecordString, err := n.createRecordString(new, dc.Name)
		if err != nil {
			return corrections, err
		}
		params[fmt.Sprintf("ADDRR%d", addrridx)] = newRecordString
		addrridx++
		delrridx++
	}
	msg := fmt.Sprintf("GENERATE_ZONEFILE: %s\n", dc.Name) + buf.String()

	if changes {
		corrections = append(corrections, &models.Correction{
			Msg: msg,
			F: func() error {
				return n.updateZoneBy(params, dc.Name)
			},
		})
	}
	return corrections, nil
}

func toRecord(r *HXRecord, origin string) *models.RecordConfig {
	rc := &models.RecordConfig{
		Type:     r.Type,
		TTL:      r.TTL,
		Original: r,
	}
	fqdn := r.Fqdn[:len(r.Fqdn)-1]
	rc.SetLabelFromFQDN(fqdn, origin)

	switch rtype := r.Type; rtype {
	case "TXT":
		rc.SetTargetTXTs(decodeTxt(r.Answer))
	case "MX":
		if err := rc.SetTargetMX(uint16(r.Priority), r.Answer); err != nil {
			panic(fmt.Errorf("unparsable MX record received from hexonet api: %w", err))
		}
	case "SRV":
		if err := rc.SetTargetSRVPriorityString(uint16(r.Priority), r.Answer); err != nil {
			panic(fmt.Errorf("unparsable SRV record received from hexonet api: %w", err))
		}
	default: // "A", "AAAA", "ANAME", "CNAME", "NS"
		if err := rc.PopulateFromString(rtype, r.Answer, r.Fqdn); err != nil {
			panic(fmt.Errorf("unparsable record received from hexonet api: %w", err))
		}
	}
	return rc
}

func (n *HXClient) showCommand(cmd map[string]string) {
	b, err := json.MarshalIndent(cmd, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Print(string(b))
}

func (n *HXClient) updateZoneBy(params map[string]interface{}, domain string) error {
	zone := domain + "."
	cmd := map[string]interface{}{
		"COMMAND":   "UpdateDNSZone",
		"DNSZONE":   zone,
		"INCSERIAL": "1",
	}
	for key, val := range params {
		cmd[key] = val
	}
	// n.showCommand(cmd)
	r := n.client.Request(cmd)
	if !r.IsSuccess() {
		return n.GetHXApiError("Error while updating zone", zone, r)
	}
	return nil
}

func (n *HXClient) getRecords(domain string) ([]*HXRecord, error) {
	var records []*HXRecord
	zone := domain + "."
	cmd := map[string]interface{}{
		"COMMAND":  "QueryDNSZoneRRList",
		"DNSZONE":  zone,
		"SHORT":    "1",
		"EXTENDED": "0",
	}
	r := n.client.Request(cmd)
	if !r.IsSuccess() {
		if r.GetCode() == 545 {
			return nil, n.GetHXApiError("Use `dnscontrol create-domains` to create not-existing zone", domain, r)
		}
		return nil, n.GetHXApiError("Failed loading resource records for zone", domain, r)
	}
	rrColumn := r.GetColumn("RR")
	if rrColumn == nil {
		return nil, fmt.Errorf("failed getting RR column for domain: %s", domain)
	}
	rrs := rrColumn.GetData()
	for _, rr := range rrs {
		spl := strings.Split(rr, " ")
		if spl[3] != "SOA" {
			record := &HXRecord{
				Raw:        rr,
				DomainName: domain,
				Host:       spl[0],
				Fqdn:       domain + ".",
				Type:       spl[3],
			}
			ttl, _ := strconv.ParseUint(spl[1], 10, 32)
			record.TTL = uint32(ttl)
			if record.Host != "@" {
				record.Fqdn = spl[0] + "." + record.Fqdn
			}
			if record.Type == "MX" || record.Type == "SRV" {
				prio, _ := strconv.ParseUint(spl[4], 10, 32)
				record.Priority = uint32(prio)
				record.Answer = strings.Join(spl[5:], " ")
			} else {
				record.Answer = strings.Join(spl[4:], " ")
			}
			records = append(records, record)
		}
	}
	return records, nil
}

func (n *HXClient) createRecordString(rc *models.RecordConfig, domain string) (string, error) {
	record := &HXRecord{
		DomainName: domain,
		Host:       rc.GetLabel(),
		Type:       rc.Type,
		Answer:     rc.GetTargetField(),
		TTL:        rc.TTL,
		Priority:   uint32(rc.MxPreference),
	}
	switch rc.Type { // #rtype_variations
	case "A", "AAAA", "ANAME", "CNAME", "MX", "NS", "PTR":
		// nothing
	case "TLSA":
		record.Answer = fmt.Sprintf(`%v %v %v %s`, rc.TlsaUsage, rc.TlsaSelector, rc.TlsaMatchingType, rc.Target)
	case "CAA":
		record.Answer = fmt.Sprintf(`%v %s "%s"`, rc.CaaFlag, rc.CaaTag, record.Answer)
	case "TXT":
		record.Answer = encodeTxt(rc.TxtStrings)
	case "SRV":
		if rc.GetTargetField() == "." {
			return "", fmt.Errorf("SRV records with empty targets are not supported (as of 2020-02-27, the API returns 'Invalid attribute value syntax')")
		}
		record.Answer = fmt.Sprintf("%d %d %v", rc.SrvWeight, rc.SrvPort, record.Answer)
		record.Priority = uint32(rc.SrvPriority)
	default:
		panic(fmt.Sprintf("createRecordString rtype %v unimplemented", rc.Type))
		// We panic so that we quickly find any switch statements
		// that have not been updated for a new RR type.
	}

	str := record.Host + " " + fmt.Sprint(record.TTL) + " IN " + record.Type + " "
	if record.Type == "MX" || record.Type == "SRV" {
		str += fmt.Sprint(record.Priority) + " "
	}
	str += record.Answer
	return str, nil
}

func (n *HXClient) deleteRecordString(record *HXRecord, domain string) string {
	return record.Raw
}

// encodeTxt encodes TxtStrings for sending in the CREATE/MODIFY API:
func encodeTxt(txts []string) string {
	ans := txts[0]

	if len(txts) > 1 {
		ans = ""
		for _, t := range txts {
			ans += `"` + strings.Replace(t, `"`, `\"`, -1) + `"`
		}
	}
	return ans
}

// finds a string surrounded by quotes that might contain an escaped quote character.
var quotedStringRegexp = regexp.MustCompile(`"((?:[^"\\]|\\.)*)"`)

// decodeTxt decodes the TXT record as received from hexonet api and
// returns the list of strings.
func decodeTxt(s string) []string {

	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		txtStrings := []string{}
		for _, t := range quotedStringRegexp.FindAllStringSubmatch(s, -1) {
			txtString := strings.Replace(t[1], `\"`, `"`, -1)
			txtStrings = append(txtStrings, txtString)
		}
		return txtStrings
	}
	return []string{s}
}
