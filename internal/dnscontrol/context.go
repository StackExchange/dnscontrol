package dnscontrol

import (
	"context"
	"github.com/StackExchange/dnscontrol/v3/internal/printer"
)

func GetDomainContext(ctx Context, domain string) Context {
	return Context{
		parent:  &ctx,
		Context: ctx,
		Log:     printer.GetBufferedPrinter(domain),
		domain:  domain,
	}
}

func GetContext() Context {
	return Context{
		Context: context.Background(),
		Log:     printer.DefaultPrinter,
	}
}

type Context struct {
	context.Context

	parent *Context

	Log    printer.Printer
	domain string
}

// StartDomain is called at the start of each domain.
func (c *Context) StartDomain() {
	if c.domain != "" {
		c.Log.Printf("******************** Domain: %s\n", c.domain)
	}
}

// EndDomain is called at the end of each domain and flushes the buffer to the supplied printer.
func (c *Context) EndDomain() {
	if c.domain != "" {
		if writer := printer.GetDomainWriter(c.domain); writer != nil {
			c.parent.Log.Printf(writer.String())
		}
	}
}

// StartDNSProvider is called at the start of each new provider.
func (c *Context) StartDNSProvider(provider string, skip bool) {
	lbl := ""
	if skip {
		lbl = " (skipping)\n"
	}
	c.Log.Debugf("----- DNS Provider: %s...%s\n", provider, lbl)
}

// EndProvider is called at the end of each provider.
func (c *Context) EndProvider(numCorrections int, err error) {
	if err != nil {
		c.Log.Println("ERROR")
		c.Log.Printf("Error getting corrections: %s\n", err)
	} else {
		plural := "s"
		if numCorrections == 1 {
			plural = ""
		}
		if numCorrections == 0 {
			c.Log.Debugf("%d correction%s\n", numCorrections, plural)
		} else {
			c.Log.Printf("%d correction%s\n", numCorrections, plural)
		}
	}
}

// PrintCorrectionMsg is called to print/format each correction.
func (c *Context) PrintCorrectionMsg(i int, msg string) {
	c.Log.Printf("#%d: %s\n", i+1, msg)
}

// EndCorrection is called at the end of each correction.
func (c *Context) EndCorrection(err error) {
	if err != nil {
		c.Log.Printf("FAILURE! %w\n", err)
	} else {
		c.Log.Println("SUCCESS!")
	}
}

// StartRegistrar is called at the start of each new registrar.
func (c *Context) StartRegistrar(provider string, skip bool) {
	lbl := ""
	if skip {
		lbl = " (skipping)\n"
	}
	c.Log.Debugf("----- Registrar: %s...%s\n", provider, lbl)
}
