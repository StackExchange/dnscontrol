package activedir

// Convert the provider's native record description to models.RecordConfig.

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/StackExchange/dnscontrol/v3/models"
)

func extractProps(cip []ciProperty) (map[string]string, map[string]uint32, error) {

	// Sadly this structure is dynamic JSON. That is, depending on .Name,
	// the .Value could be an int, string, or a map.
	// We peek at the first byte to guess at the contents.

	// We store strings in sprops, ints in uprops. Maps are special: Currently
	// the only map we decode is a map with the same duration in many units. We
	// simply pick the units we want.

	sprops := map[string]string{}
	uprops := map[string]uint32{}
	for _, p := range cip {
		name := p.Name
		if len(p.Value) == 0 {
			sprops[name] = ""
			uprops[name] = 0
		} else if p.Value[0] == '"' {
			var svalue string
			err := json.Unmarshal(p.Value, &svalue)
			if err != nil {
				return nil, nil, fmt.Errorf("could not unmarshal string value=%q: %w", p.Value, err)
			}
			sprops[name] = svalue
		} else if p.Value[0] == '{' {
			var dvalue ciValueDuration
			err := json.Unmarshal(p.Value, &dvalue)
			if err != nil {
				return nil, nil, fmt.Errorf("could not unmarshal duration value=%q: %w", p.Value, err)
			}
			uprops[name] = uint32(dvalue.TotalSeconds)
		} else {
			var uvalue uint32
			err := json.Unmarshal(p.Value, &uvalue)
			if err != nil {
				return nil, nil, fmt.Errorf("could not unmarshal uint value=%q: %w", p.Value, err)
			}
			uprops[name] = uvalue

		}
		//fmt.Printf("NAME=%q value=%q\n", name, p.Value)
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
	//fmt.Printf("TTL = %v\n", rc.TTL)

	sprops, uprops, err := extractProps(nr.RecordData.CimInstanceProperties)
	if err != nil {
		return nil, err
	}

	switch rtype := nr.RecordType; rtype {
	case "A":
		contents := sprops["IPv4Address"]
		ip := net.ParseIP(contents)
		if ip == nil || ip.To4() == nil {
			return nil, fmt.Errorf("invalid IP in A record: %s", contents)
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
	case "SOA":
		rc.SetTargetSOA(
			sprops["PrimaryServer"],
			sprops["ResponsiblePerson"],
			uprops["SerialNumber"],
			uprops["RefreshInterval"],
			uprops["RetryDelay"],
			uprops["ExpireLimit"],
			uprops["MinimumTimeToLive"])
		return nil, nil
	case "TXT":
		rc.SetTargetTXTString(sprops["DescriptiveText"])
	default:
		return nil, fmt.Errorf(
			"activedir/convert.go:nativeToRecord rtype=%q unknown: props=%+v and %+v",
			rtype, sprops, uprops)
	}

	//fmt.Printf("RECORD=%+v\n", rc)

	return rc, nil
}

//// recordsToNative takes RecordConfig and returns provider's native format.
//func recordsToNative(rcs []*models.RecordConfig, origin string) []livedns.DomainRecord {
//	// Take a list of RecordConfig and return an equivalent list of ZoneRecords.
//
//	return zrs
//}
//
