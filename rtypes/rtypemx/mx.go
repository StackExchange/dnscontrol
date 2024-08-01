package rtypemx

import (
	"encoding/json"
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

// MarshalJSON is: struct to JSON string
func (rdata *MX) MarshalJSON() ([]byte, error) {
	return json.Marshal(rdata.ComputeComparableMini())
}

// UnmarshalJSON is: JSON string to struct
//func (rdata *MX) UnmarshalJSON(b []byte) error {
//	return json.Unmarshal(b, rdata)
//}

// JSON string to struct
// func (rdata *MX) UnmarshalJSON(b []byte) error {
// 	mx, err := dns.NewRR(fmt.Sprintf("@ MX %d %s", rdata.Preference, rdata.Mx))
// 	if err != nil {
// 		return err
// 	}
// 	rdata.Preference = mx.(*dns.MX).Preference
// 	rdata.Mx = mx.(*dns.MX).Mx
// 	return nil
// }

// //return json.Marshal(rdata.ComputeComparableMini())
// //return json.Marshal(*rdata)
// //return json.Marshal("YYYtestYYY")

// //return json.Marshal(*rdata)
// r, err := json.Marshal(*rdata)
// fmt.Printf("DEBUG: mx marshal = %q\n", r)
// //panic("STOP")
// //r = append(r, []byte("foo")...)
// return r, err
//}

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
