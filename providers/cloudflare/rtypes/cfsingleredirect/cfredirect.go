package cfsingleredirect

import (
	"fmt"
	"sync"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/rtypecontrol"
)

func init() {
	rtypecontrol.Register(&CfRedirect{})
	rtypecontrol.Register(&CfTempRedirect{})
}

type CfRedirect struct{}

// Name returns the text (all caps) name of the rtype.
func (handle *CfRedirect) Name() string {
	return "CF_REDIRECT"
}

func (handle *CfRedirect) FromArgs(dc *models.DomainConfig, rec *models.RecordConfig, args []any) error {
	//fmt.Printf("DEBUG: CF_REDIRECT FromArgs called with args=%+v\n", args)
	return FromArgs_helper(dc, rec, args, 301)
}

type CfTempRedirect struct{}

// Name returns the text (all caps) name of the rtype.
func (handle *CfTempRedirect) Name() string {
	return "CF_TEMP_REDIRECT"
}

var services = map[string]int{}
var serviceMutex = sync.RWMutex{}

func inc(name string) {
	serviceMutex.Lock()
	defer serviceMutex.Unlock()

	services[name]++
}

func get(name string) int {
	serviceMutex.Lock()
	defer serviceMutex.Unlock()
	return services[name]
}

func (handle *CfTempRedirect) FromArgs(dc *models.DomainConfig, rec *models.RecordConfig, args []any) error {
	//fmt.Printf("DEBUG: CF_TEMP_REDIRECT FromArgs called with args=%+v\n", args)
	return FromArgs_helper(dc, rec, args, 302)
}

func FromArgs_helper(dc *models.DomainConfig, rec *models.RecordConfig, args []any, code int) error {

	// Pave the args to be the expected types.
	if err := rtypecontrol.PaveArgs(args, "ss"); err != nil {
		return err
	}

	// Convert old-style patterns to new-style rules:
	prWhen := args[0].(string)
	prThen := args[1].(string)
	srWhen, srThen, err := makeRuleFromPattern(prWhen, prThen)
	if err != nil {
		return err
	}

	inc(dc.UniqueName)
	name := fmt.Sprintf("%03d,%03d,%s,%s", get(dc.UniqueName), code, prWhen, prThen)

	sr := SingleRedirectConfig{}
	rec.Type = sr.Name() // This record is now a CLOUDFLAREAPI_SINGLE_REDIRECT
	return sr.FromArgs(dc, rec, []any{name, code, srWhen, srThen})
}
