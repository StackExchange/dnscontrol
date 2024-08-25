package rtypeloc

import (
	"encoding/json"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/pkg/rtypecontrol"
	"github.com/miekg/dns"
)

// https://flyandwire.com/2020/08/10/back-to-basics-latitude-and-longitude-dms-dd-ddm/

// Name is the string name for this rType.
const Name = "LOC"

func init() {
	rtypecontrol.Register(rtypecontrol.RegisterTypeOpts{
		Name: Name,
	})
}

type LOC struct {
	dns.LOC
}

func (rdata *LOC) Name() string {
	return Name
}

func (rdata *LOC) ComputeTarget() string {
	return "target" // FIXME(tlim): Convert to ZONE?
}

func (rdata *LOC) ComputeComparableMini() string {

	header := rdata.Header().String()
	full := rdata.String()
	if !strings.HasPrefix(full, header) {
		panic("assertion failed. dns.Hdr.String() behavior has changed in an incompatible way")
	}
	return full[len(header):]

}

// MarshalJSON is: struct to JSON string
func (rdata *LOC) MarshalJSON() ([]byte, error) {
	return json.Marshal(rdata.ComputeComparableMini())
}

// FromRawArgs creates a Rdata...
// update a RecordConfig using the args (from a
// RawRecord.Args). In other words, use the data from dnsconfig.js's
// rawrecordBuilder to create (actually... update) a models.RecordConfig.
func FromRawArgs(items []any) (*LOC, error) {

	// Pave the arguments.
	if err := rtypecontrol.PaveArgs(items, "iss"); err != nil {
		return nil, err
	}

	/*
		// Unpack the arguments:
		var code = items[0].(uint16)
		if code != 301 && code != 302 {
			return nil, fmt.Errorf("code (%03d) is not 301 or 302", code)
		}
		var when = items[1].(string)
		var then = items[2].(string)

		// Use the arguments to perfect the record:
		return makeLOC(code, name, when, then)
	*/

	return nil, nil
}
