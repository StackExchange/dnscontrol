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

var redirCount = map[string]int{}
var redirCountMutex = sync.RWMutex{}

func incRedirCount(name string) {
	redirCountMutex.Lock()
	defer redirCountMutex.Unlock()

	redirCount[name]++
}

func getRedirCount(name string) int {
	redirCountMutex.Lock()
	defer redirCountMutex.Unlock()
	return redirCount[name]
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

	incRedirCount(dc.UniqueName)
	name := fmt.Sprintf("%03d,%03d,%s,%s", getRedirCount(dc.UniqueName), code, prWhen, prThen)
	target := fmt.Sprintf("%s,%s", prWhen, prThen)

	sr := SingleRedirectConfig{}
	rec.Type = sr.Name() // This record is now a CLOUDFLAREAPI_SINGLE_REDIRECT
	err = sr.FromArgs(dc, rec, []any{name, code, srWhen, srThen})
	if err != nil {
		return err
	}
	_ = rec.SetTarget(target) // Overwrite target to old-style for compatibility
	return nil
}
