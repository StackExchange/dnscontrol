package models

import (
	"strings"

	"github.com/pkg/errors"
)

// SetTargetCAA sets the CAA fields.
func (rc *RecordConfig) SetTargetCAA(flag uint8, tag string, target string) {
	rc.CaaTag = tag
	rc.CaaFlag = flag
	rc.Target = target
	if rc.Type == "" {
		rc.Type = "CAA"
	}
	if rc.Type != "CAA" {
		panic("SetTargetCAA called when .Type is not CAA")
	}
	// TODO(tlim): Validate that tag is one of issue, issuewild, iodef.
}

// SetTargetCAAStrings is like SetTargetCAA but accepts strings.
func (rc *RecordConfig) SetTargetCAAStrings(flag, tag, target string) {
	rc.SetTargetCAA(atou8(flag), tag, target)
}

// SetTargetCAAString is like SetTargetCAA but accepts one big string.
// Ex: `0 issue "letsencrypt.org"`
func (rc *RecordConfig) SetTargetCAAString(s string) {
	// fmt.Printf("DEBUG: caa=(%s)\n", s)
	part := strings.Fields(s)
	if len(part) != 3 {
		panic(errors.Errorf("CAA value %#v contains too many fields", s))
	}
	rc.SetTargetCAAStrings(part[0], part[1], StripQuotes(part[2]))
}
