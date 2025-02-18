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

import (
	"fmt"
	"strconv"

	"github.com/StackExchange/dnscontrol/v4/pkg/fieldtypes"
)

func init() {
  {{ range .TypeNamesAndFields -}}
  MustRegisterType("{{ .Config.Token }}", RegisterOpts{PopulateFromRaw: PopulateFromRaw{{ .Config.Name }} })
  {{ end -}}
}
`))

func makeInit(vals Values) []byte {
	return valuesTemplate(RegisterTypeTmpl, vals)
}

// TypeTYPE

var TypeTYPETmpl = template.Must(template.New("TypeTYPE").Parse(`
//// {{ .Name }}

// {{ .Name }} is the fields needed to store a DNS record of type {{ .Name }}.
type {{ .Name }} struct {
{{- range .Fields }}
    {{ .Name }} {{ .Type }} {{ if .TagsString }} ` + "`{{ .Tags }}`" + ` {{- end }}
{{- end }}
}

`))

func makeTypeTYPE(rtconfig RTypeConfig) []byte {
	return rtypeTemplate(TypeTYPETmpl, rtconfig)
}

// ParseTYPE

var ParseTYPETmpl = template.Must(template.New("ParseTYPE").Parse(`
func Parse{{ .Name }}(rawfields []string, origin string) ({{ .Name }}, error) {

        // Error checking
        if errorCheckFieldCount(rawfields, {{ .NumRawFields }}) {
                return {{ .Name }}{}, fmt.Errorf("rtype {{ .Name }} wants %d field(s), found %d: %+v", {{ .NumRawFields }}, len(rawfields)-1, rawfields[1:])
        }

{{- range .Fields }}
   {{- if not .NoRaw }}
    var {{ .NameLower }} {{ .Type }}
   {{- end }}
{{- end }}
        var err error
{{- range .Fields }}
   {{- if not .NoRaw }}
        if {{ .NameLower }}, err = {{ .Parser }}; err != nil {
                return {{ $.Name }}{}, err
        }
   {{- end }}
{{- end }}

    return {{ .Name }}{ {{- .ConstructAll -}} }, nil
}

`))

func makeParseTYPE(rtconfig RTypeConfig) []byte {
	return rtypeTemplate(ParseTYPETmpl, rtconfig)
}

// PopulateFromRawTYPE

var PopulateFromRawTYPETmpl = template.Must(template.New("PopulateFromRawTYPE").Parse(`
// PopulateFromRaw{{ .Name }} updates rc to be an {{ .Name }} record with contents from rawfields, meta and origin.
func PopulateFromRaw{{ .Name }}(rc *RecordConfig, rawfields []string, meta map[string]string, origin string) error {
  rc.Type = "{{ .Token }}"
  {{- if .TTL1 }}
  rc.TTL = 1
  {{- end }}

  // First rawfield is the label.
  if err := rc.SetLabel3(rawfields[0], rc.SubDomain, origin); err != nil {
    return err
  }

  // Parse the remaining fields.
  {{- if .NoLabel }}
  rdata, err := Parse{{ .Name }}(rawfields, origin)
  {{- else }}
  rdata, err := Parse{{ .Name }}(rawfields[1:], origin)
  {{- end }}
  if err != nil {
    return err
  }

  return RecordUpdateFields(rc, rdata, meta)
}

`))

func makePopulateFromRawTYPE(rtconfig RTypeConfig) []byte {
	return rtypeTemplate(PopulateFromRawTYPETmpl, rtconfig)
}

// AsTYPE

var AsTYPETmpl = template.Must(template.New("AsTYPE").Parse(`
// As{{ .Name }} returns rc.Fields as an {{ .Name }} struct.
func (rc *RecordConfig) As{{ .Name }}() *{{ .Name }} {
  return rc.Fields.(*{{ .Name }})
}

`))

func makeAsTYPE(rtconfig RTypeConfig) []byte {
	return rtypeTemplate(AsTYPETmpl, rtconfig)
}

// GetFieldsTYPE

var GetFieldsTYPETmpl = template.Must(template.New("GetFieldsTYPE").Parse(`
// GetFields{{ .Name }} returns rc.Fields as individual typed values.
func (rc *RecordConfig) GetFields{{ .Name }}() ({{ .FieldTypesCommaSep }}) {
  n := rc.As{{ .Name }}()
  return {{ .ReturnIndividualFieldsList }}
}

`))

func makeGetFieldsTYPE(rtconfig RTypeConfig) []byte {
	return rtypeTemplate(GetFieldsTYPETmpl, rtconfig)
}

// GetFieldsAsStringsTYPE

var GetFieldsAsStringsTYPETmpl = template.Must(template.New("GetFieldsAsStringsTYPE").Parse(`
// GetFieldsAsStrings{{ .Name }} returns rc.Fields as individual strings.
func (rc *RecordConfig) GetFieldsAsStrings{{ .Name }}() [{{ .NumRawFields }}]string {
  n := rc.As{{ .Name }}()
  return {{ .ReturnAsStringsList }}
}

`))

func makeGetFieldsAsStringsTYPE(rtconfig RTypeConfig) []byte {
	return rtypeTemplate(GetFieldsAsStringsTYPETmpl, rtconfig)
}

// makeIntTestHeader
func makeIntTestHeader() []byte {
	return []byte(`package main

import (
  "strconv"

  "github.com/StackExchange/dnscontrol/v4/models"
)

`)
}

// GetFieldsAsStringsTYPE

var IntTestConstructorTmpl = template.Must(template.New("IntTestConstructor").Parse(`
func {{ .NameLower }}(name string, {{ .FieldsAsSignature }}) *models.RecordConfig {
{{- range .Fields }}
{{- if .ConvertToString }}
  {{ .ConvertToString }}
{{- end }}
{{- end }}

  rdata, err := models.Parse{{ .Name }}([]string{ {{- .FieldsAsSVars -}} }, "**current-domain**")
  if err != nil {
    panic(err)
  }
  return models.MustCreateRecord(name, rdata, nil, 300, "", "**current-domain**")
}

`))

func makeIntTestConstructor(rtconfig RTypeConfig) []byte {
	return rtypeTemplate(IntTestConstructorTmpl, rtconfig)
}

// helpersRawRecordBuilder

var helpersRawRecordBuilderTmpl = template.Must(template.New("helpersRawRecordBuilder").Parse(`
{{- range .TypeNamesAndFields -}}
var {{ .Config.Token }} = rawrecordBuilder('{{ .Config.Token }}');
{{ end -}}
`))

func makehelpersRawRecordBuilder(vals Values) []byte {
	return valuesTemplate(helpersRawRecordBuilderTmpl, vals)
}
