package rtypemx

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/pkg/rtypecontrol"
	"github.com/miekg/dns"
)

const Token = "MX"

type TypeMX struct {
	dns.MX
}

func init() {
	rtypecontrol.Register(Token)
}

// // FromRawArgs update a models.RecordConfig using the args (from a
// // models.RawRecord.Args). In other words, use the data from dnsconfig.js's
// // rawrecordBuilder to create (actually... update) a models.RecordConfig.
// func FromRawArgs(zone string, rc *models.RecordConfig, items []any) error {

// 	if err := rtypecontrol.PaveArgs(items, "sis"); err != nil {
// 		return err
// 	}

// 	var name = items[0].(string)
// 	var preference = items[1].(uint16)
// 	var mx = items[2].(string)

// 	rc.SetLabel(name, zone)

// 	r := new(dns.MX)
// 	r.Hdr = dns.RR_Header{Name: name, Rrtype: dns.TypeMX, Class: dns.ClassINET, Ttl: 3600}
// 	r.Preference = preference
// 	r.Mx = mx

// 	rc.Rdata = r

// 	return nil
// }

/*

// from dnsconfig.js:

rc := &models.RecordConfig{}
rc.Rdata = rtype.MX{}
rc.Rdata.PopulateFromRawArgs(items)

rc.Rdata = rtype.MX{}.PopulateFromRawArgs(items)

// from API data:

rc := &models.RecordConfig{}
rc.Rdata = rtype.MX{}
rc.Rdata.SetTargetMX(pref, target)

rc.Rdata = rtype.MX{}.SetTargetMX(pref, target)

// API that returns a line of a zonefile:

rc := &models.RecordConfig{}
rc.PopulateFromString(rtype, target, origin)

rc := &models.RecordConfig{}
rc.PopulateFromArgs(origin, rtype, item, item, item)


	x := RecordConfig{}
	x.Rdata = &rtype.MX{}
	x.Rdata.(*rtype.MX).Mx = "foo"
	x.Rdata.(*rtype.MX).Preference = 99
	fmt.Println(x)

	y := RecordConfig{}
	y.Rdata = &rtype.MX{}
	y.Rdata.(*rtype.MX).SetUp(1, "bar")
	fmt.Println(y)

	m := &rtype.MX{}
	m.SetUp(1, "bar")
	z := RecordConfig{Rdata: m}
	fmt.Println(z)

*/

// SetTargetMX sets the MX fields.
func (rdat *TypeMX) SetTargetMX(pref uint16, target string) error {
	rdat.Preference = pref
	rdat.Mx = target
	return nil
}

// SetTargetMXStrings is like SetTargetMX but accepts strings.
func (rdat *TypeMX) SetTargetMXStrings(pref, target string) error {
	u64pref, err := strconv.ParseUint(pref, 10, 16)
	if err != nil {
		return fmt.Errorf("can't parse MX data: %w", err)
	}
	return rdat.SetTargetMX(uint16(u64pref), target)
}

// SetTargetMXString is like SetTargetMX but accepts one big string.
func (rdat *TypeMX) SetTargetMXString(s string) error {
	part := strings.Fields(s)
	if len(part) != 2 {
		return fmt.Errorf("MX value does not contain 2 fields: (%#v)", s)
	}
	return rdat.SetTargetMXStrings(part[0], part[1])
}
