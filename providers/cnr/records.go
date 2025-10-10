package cnr

import (
	"bytes"
	"errors"
	"fmt"
	"maps"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/diff"
	"github.com/StackExchange/dnscontrol/v4/pkg/txtutil"
)

// Record covers an individual DNS resource record.
type Record struct {
	// DomainName is the zone that the record belongs to.
	DomainName string
	// Host is the hostname relative to the zone: e.g. for a record for blog.example.org, domain would be "example.org" and host would be "blog".
	// An apex record would be specified by either an empty host "" or "@".
	// A SRV record would be specified by "_{service}._{protocol}.{host}": e.g. "_sip._tcp.phone" for _sip._tcp.phone.example.org.
	Host string
	// FQDN is the Fully Qualified Domain Name. It is the combination of the host and the domain name. It always ends in a ".". FQDN is ignored in CreateRecord, specify via the Host field instead.
	Fqdn string
	// Type is one of the following: A, AAAA, ANAME, ALIAS, CNAME, MX, NS, SRV, or TXT.
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
func (n *Client) GetZoneRecords(domain string, meta map[string]string) (models.Records, error) {
	records, err := n.getRecords(domain)
	if err != nil {
		return nil, err
	}
	actual := make([]*models.RecordConfig, len(records))
	for i, r := range records {
		actual[i] = toRecord(r, domain)
	}

	return actual, nil
}

// GetZoneRecordsCorrections returns a list of corrections that will turn existing records into dc.Records.
func (n *Client) GetZoneRecordsCorrections(dc *models.DomainConfig, actual models.Records) ([]*models.Correction, int, error) {
	toReport, create, del, mod, actualChangeCount, err := diff.NewCompat(dc).IncrementalDiff(actual)
	if err != nil {
		return nil, 0, err
	}
	// Start corrections with the reports
	corrections := diff.GenerateMessageCorrections(toReport)

	buf := &bytes.Buffer{}
	// Print a list of changes. Generate an actual change that is the zone
	changes := false
	var builder strings.Builder
	params := map[string]any{}
	delrridx := 0
	addrridx := 0

	for _, cre := range create {
		changes = true
		fmt.Fprintln(buf, cre)
		newRecordString, err := n.createRecordString(cre.Desired, dc.Name)
		if err != nil {
			return corrections, 0, err
		}
		key := fmt.Sprintf("ADDRR%d", addrridx)
		params[key] = newRecordString
		fmt.Fprintf(&builder, "\033[32m+ %s = %s\033[0m\n", key, newRecordString)
		addrridx++
	}
	for _, d := range del {
		changes = true
		fmt.Fprintln(buf, d)
		key := fmt.Sprintf("DELRR%d", delrridx)
		oldRecordString := n.deleteRecordString(d.Existing.Original.(*Record))
		params[key] = oldRecordString
		fmt.Fprintf(&builder, "\033[31m- %s = %s\033[0m\n", key, oldRecordString)
		delrridx++
	}
	for _, chng := range mod {
		changes = true
		fmt.Fprintln(buf, chng)
		// old record deletion
		key := fmt.Sprintf("DELRR%d", delrridx)
		oldRecordString := n.deleteRecordString(chng.Existing.Original.(*Record))
		params[key] = oldRecordString
		fmt.Fprintf(&builder, "\033[31m- %s = %s\033[0m\n", key, oldRecordString)
		delrridx++
		// new record creation
		newRecordString, err := n.createRecordString(chng.Desired, dc.Name)
		if err != nil {
			return corrections, 0, err
		}
		key = fmt.Sprintf("ADDRR%d", addrridx)
		params[key] = newRecordString
		fmt.Fprintf(&builder, "\033[32m+ %s = %s\033[0m\n", key, newRecordString)
		addrridx++
	}

	if changes {
		msg := fmt.Sprintf("GENERATE_ZONE: %s\n%s", dc.Name, buf.String())
		if n.isDebugOn() {
			msg = fmt.Sprintf("GENERATE_ZONE: %s\n%sPROVIDER CNR, API COMMAND PARAMETERS:\n%s", dc.Name, buf.String(), builder.String())
		}
		corrections = append(corrections, &models.Correction{
			Msg: msg,
			F: func() error {
				return n.updateZoneBy(params, dc.Name)
			},
		})
	}

	return corrections, actualChangeCount, nil
}

func toRecord(r *Record, origin string) *models.RecordConfig {
	rc := &models.RecordConfig{
		Type:     r.Type,
		TTL:      r.TTL,
		Original: r,
	}
	fqdn := r.Fqdn[:len(r.Fqdn)-1]
	rc.SetLabelFromFQDN(fqdn, origin)

	switch r.Type {
	case "MX", "SRV":
		if r.Priority > 65535 {
			panic(fmt.Errorf("priority value out of range for %s record: %d", r.Type, r.Priority))
		}
		if r.Type == "MX" {
			if err := rc.SetTargetMX(uint16(r.Priority), r.Answer); err != nil {
				panic(fmt.Errorf("unparsable MX record received from centralnic reseller API: %w", err))
			}
		} else {
			// _service._proto.name. TTL Type Priority Weight Port Target.
			// e.g. _sip._tcp.phone.example.org. 86400 IN SRV 5 6 7 sip.example.org.
			// r.Anser covers the format "Priority Weight Port Target" and we've to remove the priority from the string
			r.Answer = strings.TrimPrefix(r.Answer, fmt.Sprintf("%d ", r.Priority))
			if err := rc.SetTargetSRVPriorityString(uint16(r.Priority), r.Answer); err != nil {
				panic(fmt.Errorf("unparsable SRV record received from centralnic reseller API: %w", err))
			}
		}
	default: // "A", "AAAA", "ANAME", "ALIAS", "CNAME", "NS", "TXT", "CAA", "TLSA", "PTR"
		if err := rc.PopulateFromStringFunc(r.Type, r.Answer, r.Fqdn, txtutil.ParseQuoted); err != nil {
			panic(fmt.Errorf("unparsable record received from centralnic reseller API: %w", err))
		}
	}
	return rc
}

// updateZoneBy updates the zone with the provided changes.
func (n *Client) updateZoneBy(params map[string]any, domain string) error {
	zone := domain
	cmd := map[string]any{
		"COMMAND": "ModifyDNSZone",
		"DNSZONE": zone,
	}
	maps.Copy(cmd, params)
	r := n.client.Request(cmd)
	if !r.IsSuccess() {
		return n.GetAPIError("Error while updating zone", zone, r)
	}
	return nil
}

// deleteRecordString constructs the record string based on the provided Record.
func (n *Client) getRecords(domain string) ([]*Record, error) {
	var records []*Record

	// Command to find out the total numbers of resource records for the zone
	// so that the follow-up query can be done with the correct limit
	cmd := map[string]any{
		"COMMAND": "QueryDNSZoneRRList",
		"DNSZONE": domain,
		"ORDERBY": "type",
		"FIRST":   "0",
		"LIMIT":   "10000",
		"WIDE":    "1",
	}
	r := n.client.Request(cmd)

	// Check if the request was successful
	if !r.IsSuccess() {
		if r.GetCode() == 545 {
			// If dns zone does not exist create a new one automatically
			if !isNoPopulate() {
				err := n.EnsureZoneExists(domain)
				if err != nil {
					return nil, err
				}
			} else {
				// Return specific error if the zone does not exist
				return nil, n.GetAPIError("Use `dnscontrol create-domains` to create not-existing zone", domain, r)
			}
		}
		// Return general error for any other issues
		return nil, n.GetAPIError("Failed loading resource records for zone", domain, r)
	}
	totalRecords := r.GetRecordsTotalCount()
	if totalRecords <= 0 {
		return nil, nil
	}

	// loop over the records array
	rrs := r.GetRecords()
	for i := range len(rrs) {
		data := rrs[i].GetData()
		// fmt.Printf("Data: %+v\n", data)
		if _, exists := data["NAME"]; !exists {
			continue
		}

		if data["TYPE"] == "MX" {
			tmp := strings.Split(data["CONTENT"], " ")
			data["PRIO"] = tmp[0]
			data["CONTENT"] = tmp[1]
		}

		// Parse the TTL string to an unsigned integer
		ttl, err := strconv.ParseUint(data["TTL"], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid TTL value for domain %s: %s", domain, data["TTL"])
		}

		// Parse the TTL string to an unsigned integer
		priority, _ := strconv.ParseUint(data["PRIO"], 10, 32)

		// Add dot to Answer if supported by the record type
		pattern := `^ALIAS|CNAME|MX|NS|SRV|PTR$`
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("error compiling regex in getRecords: %w", err)
		}
		if re.MatchString(data["TYPE"]) && !strings.HasSuffix(data["CONTENT"], ".") {
			data["CONTENT"] = data["CONTENT"] + "."
		}

		// Only append domain if it's not already a fully qualified domain name
		fqdn := domain + "."
		if data["NAME"] != "@" && !strings.HasSuffix(data["NAME"], domain+".") {
			fqdn = fmt.Sprintf("%s.%s.", data["NAME"], domain)
		}

		// Initialize a new Record
		record := &Record{
			DomainName: domain,
			Host:       data["NAME"],
			Fqdn:       fqdn,
			Type:       data["TYPE"],
			Answer:     data["CONTENT"],
			TTL:        uint32(ttl),
			Priority:   uint32(priority),
		}
		// fmt.Printf("Record: %+v\n", record)

		// Append the record to the records slice
		records = append(records, record)
	}

	// Return the slice of records
	return records, nil
}

// Function to create record string from given RecordConfig for the ADDRR# API parameter
func (n *Client) createRecordString(rc *models.RecordConfig, domain string) (string, error) {
	host := rc.GetLabel()
	answer := ""

	switch rc.Type { // #rtype_variations
	case "A", "AAAA", "ANAME", "ALIAS", "CNAME", "MX", "NS", "PTR":
		answer = rc.GetTargetField()
		if domain == host {
			host = host + "."
		}
	case "SSHFP":
		answer = fmt.Sprintf(`%v %v %s`, rc.SshfpAlgorithm, rc.SshfpFingerprint, rc.GetTargetField())
		if domain == host {
			host = host + "."
		}
	case "NAPTR":
		answer = fmt.Sprintf(`%v %v "%v" "%v" "%v" %v`, rc.NaptrOrder, rc.NaptrPreference, rc.NaptrFlags, rc.NaptrService, rc.NaptrRegexp, rc.GetTargetField())
		if domain == host {
			host = host + "."
		}
	case "TLSA":
		answer = fmt.Sprintf(`%v %v %v %s`, rc.TlsaUsage, rc.TlsaSelector, rc.TlsaMatchingType, rc.GetTargetField())
	case "CAA":
		answer = fmt.Sprintf(`%v %s "%s"`, rc.CaaFlag, rc.CaaTag, rc.GetTargetField())
	case "TXT":
		answer = txtutil.EncodeQuoted(rc.GetTargetTXTJoined())
	case "SRV":
		if rc.GetTargetField() == "." {
			return "", errors.New("SRV records with empty targets are not supported")
		}
		// _service._proto.name. TTL Type Priority Weight Port Target.
		// e.g. _sip._tcp.phone.example.org. 86400 IN SRV 5 6 7 sip.example.org.
		answer = fmt.Sprintf("%d %d %d %v", uint32(rc.SrvPriority), rc.SrvWeight, rc.SrvPort, rc.GetTargetField())
	default:
		panic(fmt.Sprintf("createRecordString rtype %v unimplemented", rc.Type))
		// We panic so that we quickly find any switch statements
		// that have not been updated for a new RR type.
	}

	str := host + " " + strconv.FormatUint(uint64(rc.TTL), 10) + " "

	if rc.Type != "NS" { // TODO
		str += "IN "
	}
	str += rc.Type + " "
	// Handle MX records which have priority
	if rc.Type == "MX" {
		str += strconv.FormatUint(uint64(uint32(rc.MxPreference)), 10) + " "
	}
	str += answer
	return str, nil
}

// deleteRecordString constructs the record string based on the provided Record.
func (n *Client) deleteRecordString(record *Record) string {
	// Initialize values slice
	values := []string{
		record.Host,
		strconv.FormatUint(uint64(record.TTL), 10),
		"IN",
		record.Type,
	}
	if record.Type == "SRV" {
		values = append(values, strconv.FormatUint(uint64(record.Priority), 10))
	}
	values = append(values, record.Answer)

	// fmt.Printf("Values: %+v\n", values)

	// Remove IN if the record type is "NS" TODO
	if record.Type == "NS" {
		values = append(values[:2], values[3:]...) // Skip over the "IN"
	}

	// Return the final string by joining the elements with spaces
	return strings.Join(values, " ")
}

// Function to check the no-populate argument
func isNoPopulate() bool {
	return slices.Contains(os.Args, "--no-populate")
}

// Function to check if debug mode is enabled
func (n *Client) isDebugOn() bool {
	debugMode, exists := n.conf["debugmode"]
	return exists && (debugMode == "1" || debugMode == "2")
}
