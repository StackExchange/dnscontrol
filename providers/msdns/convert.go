package msdns

// Convert the provider's native record description to models.RecordConfig.

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/StackExchange/dnscontrol/v3/models"
)

// extractProps and collects Name/Value pairs into maps for easier access.
func extractProps(cip []ciProperty) (map[string]string, map[string]uint32, error) {

	// Sadly this structure is dynamic JSON i.e. .Value could be an int, string,
	// or a map. We peek at the first byte to guess at the contents.

	// We store strings in sprops, numbers in uprops. Maps are special: Currently
	// the only map we decode is a map with the same duration in many units. We
	// simply pick the Seconds unit and store it as a number.

	sprops := map[string]string{}
	uprops := map[string]uint32{}
	for _, p := range cip {
		name := p.Name
		if len(p.Value) == 0 {
			// Empty string? Skip it.
		} else if p.Value[0] == '"' {
			// First byte is a quote. Must be a string.
			var svalue string
			err := json.Unmarshal(p.Value, &svalue)
			if err != nil {
				return nil, nil, fmt.Errorf("could not unmarshal string value=%q: %w", p.Value, err)
			}
			sprops[name] = svalue
		} else if p.Value[0] == '{' {
			// First byte is {.  Must be a map.
			var dvalue ciValueDuration
			err := json.Unmarshal(p.Value, &dvalue)
			if err != nil {
				return nil, nil, fmt.Errorf("could not unmarshal duration value=%q: %w", p.Value, err)
			}
			uprops[name] = uint32(dvalue.TotalSeconds)
		} else {
			// Assume it is a number.
			var uvalue uint32
			err := json.Unmarshal(p.Value, &uvalue)
			if err != nil {
				return nil, nil, fmt.Errorf("could not unmarshal uint value=%q: %w", p.Value, err)
			}
			uprops[name] = uvalue
		}
	}
	return sprops, uprops, nil
}

// nativeToRecord takes a DNS record from DNS and returns a native RecordConfig struct.
func nativeToRecords(nr nativeRecord, origin string) (*models.RecordConfig, error) {
	rc := &models.RecordConfig{
		Type:     nr.RecordType,
		Original: nr,
	}
	rc.SetLabel(nr.HostName, origin)
	rc.TTL = uint32(nr.TimeToLive.TotalSeconds)

	sprops, uprops, err := extractProps(nr.RecordData.CimInstanceProperties)
	if err != nil {
		return nil, err
	}

	switch rtype := nr.RecordType; rtype {
	case "A":
		contents := sprops["IPv4Address"]
		ip := net.ParseIP(contents)
		if ip == nil || ip.To4() == nil {
			return nil, fmt.Errorf("invalid IP in A record: %q", contents)
		}
		rc.SetTargetIP(ip)
	case "AAAA":
		contents := sprops["IPv6Address"]
		ip := net.ParseIP(contents)
		if ip == nil || ip.To16() == nil {
			return nil, fmt.Errorf("invalid IPv6 in AAAA record: %q", contents)
		}
		rc.SetTargetIP(ip)
	case "CNAME":
		rc.SetTarget(sprops["HostNameAlias"])
	case "MX":
		rc.SetTargetMX(uint16(uprops["Preference"]), sprops["MailExchange"])
	case "NS":
		rc.SetTarget(sprops["NameServer"])
	case "PTR":
		rc.SetTarget(sprops["PtrDomainName"])
	case "SRV":
		rc.SetTargetSRV(
			uint16(uprops["Priority"]),
			uint16(uprops["Weight"]),
			uint16(uprops["Port"]),
			sprops["DomainName"],
		)
	case "SOA":
		// We discard SOA records for now. Windows DNS doesn't let us delete
		// them and they get in the way of integration tests. In the future,
		// we should support SOA records by (1) ignoring them in the
		// integration tests. (2) generatePSModify will have to special-case
		// updates.
		return nil, nil
		// If we weren't ignoring them, the code would look like this:
		//rc.SetTargetSOA(sprops["PrimaryServer"], sprops["ResponsiblePerson"],
		//	uprops["SerialNumber"], uprops["RefreshInterval"], uprops["RetryDelay"],
		//	uprops["ExpireLimit"], uprops["MinimumTimeToLive"])
	case "TXT":
		rc.SetTargetTXTString(sprops["DescriptiveText"])
	default:
		return nil, fmt.Errorf(
			"msdns/convert.go:nativeToRecord rtype=%q unknown: props=%+v and %+v",
			rtype, sprops, uprops)
	}

	return rc, nil
}
