package main

import (
	"bytes"
	"strings"
	"text/template"
)

var funcs = template.FuncMap{
	"join": strings.Join,
}

func valuesTemplate(theTemplate *template.Template, vals Values) []byte {
	b := &bytes.Buffer{}
	err := theTemplate.Execute(b, vals)
	if err != nil {
		panic(err)
	}
	return b.Bytes()
}

func rtypeTemplate(theTemplate *template.Template, vals RTypeConfig) []byte {
	b := &bytes.Buffer{}
	err := theTemplate.Execute(b, vals)
	if err != nil {
		panic(err)
	}
	return b.Bytes()
}

// RecordType

var RecordTypeTmpl = template.Must(template.New("RecordType").Funcs(funcs).Parse(`
// RecordType is a constraint for DNS records.
type RecordType interface {
   {{ join .TypeNames " | " }}
}
`))

func makeInterfaceConstraint(vals Values) []byte {
	return valuesTemplate(RecordTypeTmpl, vals)
}

// RegisterType

var RegisterTypeTmpl = template.Must(template.New("RegisterType").Parse(`
package models

import "github.com/StackExchange/dnscontrol/v4/pkg/fieldtypes"

func init() {
  {{ range .TypeNames -}}
  MustRegisterType("{{ . }}", RegisterOpts{PopulateFromRaw: PopulateFromRaw{{ . }} })
  {{ end }}
}
`))

func makeInit(vals Values) []byte {
	return valuesTemplate(RegisterTypeTmpl, vals)
}

// TypeTYPE

var TypeTYPETmpl = template.Must(template.New("TypeTYPE").Parse(`
// {{ .Name }} is the fields needed to store a DNS record of type {{ .Name }}.
type {{ .Name }} struct {
{{- range .Fields }}
    {{ .Name }} {{ .Type }} {{ if .Tags }} ` + "`{{ .Tags }}`" + ` {{- end }}
{{- end }}
}

`))

func makeTypeTYPE(rtconfig RTypeConfig) []byte {
	return rtypeTemplate(TypeTYPETmpl, rtconfig)
}
