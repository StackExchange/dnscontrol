package printer

import (
	"github.com/StackExchange/dnscontrol/models"
)

//ReplayCLI is a CLI that simply stores all output calls on it to be replayed later. Calling input functions like PromptToRun will panic
type ReplayCLI struct {
	fs []func(CLI)
}

func (r *ReplayCLI) StartDomain(domain string) {
	r.fs = append(r.fs, func(c CLI) { c.StartDomain(domain) })
}

func (r *ReplayCLI) PrintCorrection(i int, correction *models.Correction) {
	r.fs = append(r.fs, func(c CLI) { c.PrintCorrection(i, correction) })
}

func (r *ReplayCLI) PromptToRun() bool {
	panic("Cannot call PromptToRun on ReplayCLI")
}

func (r *ReplayCLI) EndCorrection(err error) {
	r.fs = append(r.fs, func(c CLI) { c.EndCorrection(err) })
}

func (r *ReplayCLI) StartDNSProvider(provider string, skip bool) {
	r.fs = append(r.fs, func(c CLI) { c.StartDNSProvider(provider, skip) })
}

func (r *ReplayCLI) StartRegistrar(provider string, skip bool) {
	r.fs = append(r.fs, func(c CLI) { c.StartRegistrar(provider, skip) })
}

func (r *ReplayCLI) EndProvider(numCorrections int, err error) {
	r.fs = append(r.fs, func(c CLI) { c.EndProvider(numCorrections, err) })
}

func (r *ReplayCLI) Debugf(format string, args ...interface{}) {
	r.fs = append(r.fs, func(c CLI) { c.Debugf(format, args) })
}

func (r *ReplayCLI) Warnf(format string, args ...interface{}) {
	r.fs = append(r.fs, func(c CLI) { c.Warnf(format, args) })
}

func (r *ReplayCLI) Replay(c CLI) {
	for _, f := range r.fs {
		f(c)
	}
}
