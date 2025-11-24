package cfsingleredirect

import (
	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/rtypecontrol"
)

// "github.com/StackExchange/dnscontrol/v4/models"
// "github.com/StackExchange/dnscontrol/v4/pkg/rtypecontrol"

func init() {
	rtypecontrol.Register(SingleRedirect{})
}

type SingleRedirect struct{}

func (handle *SingleRedirect) Name() string {
	return "CLOUDFLAREAPI_SINGLE_REDIRECT"
}

// func MakeSingleRedirect() SingleRedirect                               { return SingleRedirect{} }
func (handle *SingleRedirect) FromArgs([]any) (*models.RecordConfig, error) {
	rec := &models.RecordConfig{
		Type: handle.Name(),
		TTL: ttl,

		//FilePos = FixFilePos(handle.FilePos)
	}
	return rec, nil
}
// 	// Validate types.
// 	if err := rtypecontrol.PaveArgs(items, "siss"); err != nil {
// 		return err
// 	}

// 	// Unpack the args:
// 	var name, when, then string
// 	var code uint16

// 	name = items[0].(string)
// 	code = items[1].(uint16)
// 	if code != 301 && code != 302 && code != 303 && code != 307 && code != 308 {
// 		return fmt.Errorf("%s: code (%03d) is not 301,302,303,307,308", rc.FilePos, code)
// 	}
// 	when = items[2].(string)
// 	then = items[3].(string)

// 	return makeSingleRedirectFromRawRec(rc, code, name, when, then)
	return &models.RecordConfig{}, nil
}

//func (handle *SingleRedirect) IDNFields(argsRaw) (argsIDN, argsUnicode, error) {}
//func (handle *SingleRedirect) AsRFC1038String(*models.RecordConfig) string     {}
//func (handle *SingleRedirect) CopyToLegacyFields(*models.RecordConfig)         {}
//func (handle *SingleRedirect) CopyFromLegacyFields(*models.RecordConfig)       {}

// // FromRaw convert RecordConfig using data from a RawRecordConfig's parameters.
// func FromRaw(rc *models.RecordConfig, items []any) error {
// 	// Validate types.
// 	if err := rtypecontrol.PaveArgs(items, "siss"); err != nil {
// 		return err
// 	}

// 	// Unpack the args:
// 	var name, when, then string
// 	var code uint16

// 	name = items[0].(string)
// 	code = items[1].(uint16)
// 	if code != 301 && code != 302 && code != 303 && code != 307 && code != 308 {
// 		return fmt.Errorf("%s: code (%03d) is not 301,302,303,307,308", rc.FilePos, code)
// 	}
// 	when = items[2].(string)
// 	then = items[3].(string)

// 	return makeSingleRedirectFromRawRec(rc, code, name, when, then)
// }
