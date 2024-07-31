package rtypemx

import (
	"strings"

	"github.com/StackExchange/dnscontrol/v4/pkg/rtypecontrol"
	"github.com/miekg/dns"
)

// Name is the string version of this rtype.
const Name = "MX"

func init() {
	rtypecontrol.Register(rtypecontrol.RegisterTypeOpts{
		Name: Name,
		//FromRawArgsFn: FromRawArgs,
	})
}

type MX struct {
	dns.MX
}

func (rdata *MX) Name() string {
	return Name
}

func (rdata *MX) ComputeTarget() string {
	return rdata.MX.Mx
}

func (rdata *MX) ComputeComparableMini() string {

	header := rdata.Header().String()
	full := rdata.String()
	if !strings.HasPrefix(full, header) {
		panic("assertion failed. dns.Hdr.String() behavior has changed in an incompatible way")
	}
	return full[len(header):]

}

func FromRawArgs(items []any) (*MX, error) {

	if err := rtypecontrol.PaveArgs(items, "is"); err != nil {
		return nil, err
	}

	var preference = items[0].(uint16)
	var mx = items[1].(string)

	rdata := &MX{}
	rdata.Preference = preference
	rdata.Mx = mx

	return rdata, nil
}

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
