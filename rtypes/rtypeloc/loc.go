package rtypeloc

import (
	"encoding/json"
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/pkg/rtypecontrol"
)

// https://flyandwire.com/2020/08/10/back-to-basics-latitude-and-longitude-dms-dd-ddm/

// Name is the string name for this rType.
const Name = "LOC"

func init() {
	rtypecontrol.Register(rtypecontrol.RegisterTypeOpts{
		Name: Name,
	})
}

// LOC contains the data fiels for LOC.
type LOC struct {
	LocVersion   uint8  `json:"locversion,omitempty"`
	LocSize      uint8  `json:"locsize,omitempty"`
	LocHorizPre  uint8  `json:"lochorizpre,omitempty"`
	LocVertPre   uint8  `json:"locvertpre,omitempty"`
	LocLatitude  uint32 `json:"loclatitude,omitempty"`
	LocLongitude uint32 `json:"loclongitude,omitempty"`
	LocAltitude  uint32 `json:"localtitude,omitempty"`
}

func (rdata *LOC) Name() string {
	return Name
}

func (rdata *LOC) ComputeTarget() string {
	// The closest equivalent to a target "hostname" is the rule name.
	return rdata.SRName
}

func (rdata *LOC) ComputeComparableMini() string {
	// The differencing engine uses this.
	return rdata.SRDisplay
}

func (rdata *LOC) MarshalJSON() ([]byte, error) {
	return json.Marshal(*rdata)
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

	// Unpack the arguments:
	var code = items[0].(uint16)
	if code != 301 && code != 302 {
		return nil, fmt.Errorf("code (%03d) is not 301 or 302", code)
	}
	var when = items[1].(string)
	var then = items[2].(string)

	// Use the arguments to perfect the record:
	return makeLOC(code, name, when, then)
}
