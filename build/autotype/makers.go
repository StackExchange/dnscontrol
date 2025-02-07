package main

import (
	"bytes"
	"strings"
	"text/template"
)

var funcs = template.FuncMap{
	"join": strings.Join,
}

func simpleTemplate(theTemplate *template.Template, vals Values) []byte {
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

func makeRecordType(vals Values) []byte {
	return simpleTemplate(RecordTypeTmpl, vals)
}

// RegisterType

var RegisterTypeTmpl = template.Must(template.New("RegisterType").Funcs(funcs).Parse(`
package models

import "github.com/StackExchange/dnscontrol/v4/pkg/fieldtypes"

func init() {
  {{ range .TypeNames -}}
  MustRegisterType("{{ . }}", RegisterOpts{PopulateFromRaw: PopulateFromRaw{{ . }} })
  {{ end }}
}
`))

func makeRegisterType(vals Values) []byte {
	return simpleTemplate(RegisterTypeTmpl, vals)
}

// TypeTYPE

var TypeTYPETmpl = template.Must(template.New("TypeTYPE").Funcs(funcs).Parse(`
{{ range .TypeNamesAndFields }}
// {{ .Name }} is the fields needed to store a DNS record of type {{ .Name }}.
type {{ .Name }} struct {
{{- range .Fields }}
    {{ .Name }} {{ .Type }} {{ if .Tags }} ` + "`{{ .Tags }}`" + ` {{- end }}
{{- end }}
}
{{ end }}
`))

func makeTypeTYPE(vals Values) []byte {
	return simpleTemplate(TypeTYPETmpl, vals)
}
