package msdns

import (
	"github.com/StackExchange/dnscontrol/v3/internal/dnscontrol"
)

// just to be safe that we have at least a global context
var ctx = dnscontrol.GetContext()

func SetContext(ctx *dnscontrol.Context) {
	ctx = ctx
}
