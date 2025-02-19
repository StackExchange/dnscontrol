package main

import (
	"bytes"
	"fmt"
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

func makeInterfaceConstraint(vals Values) []byte {
	return valuesTemplate(RecordTypeTmpl, vals)
}

var RecordTypeTmpl = template.Must(template.New("RecordType").Funcs(funcs).Parse(`
// RecordType is a constraint for DNS records.
type RecordType interface {
   {{ join .TypeNames " | " }}
}
`))

// RegisterType

func makeInit(vals Values) []byte {
	return valuesTemplate(RegisterTypeTmpl, vals)
}

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

// TypeTYPE

func makeTypeTYPE(rtconfig RTypeConfig) []byte {
	return rtypeTemplate(TypeTYPETmpl, rtconfig)
}

var TypeTYPETmpl = template.Must(template.New("TypeTYPE").Parse(`
//// {{ .Name }}

// {{ .Name }} is the fields needed to store a DNS record of type {{ .Name }}.
type {{ .Name }} struct {
{{- range .Fields }}
    {{ .Name }} {{ .Type }} {{ if .TagsString }} ` + "`{{ .Tags }}`" + ` {{- end }}
{{- end }}
}

`))

// ParseTYPE

func makeParseTYPE(rtconfig RTypeConfig) []byte {
	return rtypeTemplate(ParseTYPETmpl, rtconfig)
}

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

func mkConstructAll(fields []Field) string {
	var ac []string
	for _, field := range fields {
		if HasTagOption(field.Tags, "dnscontrol", "srdisplay") {
			ac = append(ac, fmt.Sprintf(`%s: cfSingleRedirecttargetFromRaw(srname, code, srwhen, srthen)`, field.Name))
		} else if HasTagOption(field.Tags, "dnscontrol", "parsereturnunknowable") {
			ac = append(ac, fmt.Sprintf(`%s: "UNKNOWABLE"`, field.Name))
		} else if HasTagOption(field.Tags, "dnscontrol", "noparsereturn") {
			// Skip this field.
		} else {
			ac = append(ac, fmt.Sprintf("%s: %s", field.Name, field.NameLower))
		}
	}
	return strings.Join(ac, ", ")
}

// PopulateFromRawTYPE

func makePopulateFromRawTYPE(rtconfig RTypeConfig) []byte {
	return rtypeTemplate(PopulateFromRawTYPETmpl, rtconfig)
}

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

// AsTYPE

func makeAsTYPE(rtconfig RTypeConfig) []byte {
	return rtypeTemplate(AsTYPETmpl, rtconfig)
}

var AsTYPETmpl = template.Must(template.New("AsTYPE").Parse(`
// As{{ .Name }} returns rc.Fields as an {{ .Name }} struct.
func (rc *RecordConfig) As{{ .Name }}() *{{ .Name }} {
  return rc.Fields.(*{{ .Name }})
}

`))

// GetFieldsTYPE

func makeGetFieldsTYPE(rtconfig RTypeConfig) []byte {
	return rtypeTemplate(GetFieldsTYPETmpl, rtconfig)
}

var GetFieldsTYPETmpl = template.Must(template.New("GetFieldsTYPE").Parse(`
// GetFields{{ .Name }} returns rc.Fields as individual typed values.
func (rc *RecordConfig) GetFields{{ .Name }}() ({{ .FieldTypesCommaSep }}) {
  n := rc.As{{ .Name }}()
  return {{ .ReturnIndividualFieldsList }}
}

`))

func mkFieldTypesCommaSep(fields []Field) string {
	var ac []string
	for _, field := range fields {
		if !HasTagOption(field.Tags, "dnscontrol", "noraw") {
			ac = append(ac, field.Type)
		}
	}
	return strings.Join(ac, ", ")
}

func mkReturnIndividualFieldsList(fields []Field) string {
	var ac []string
	for _, field := range fields {
		if HasTagOption(field.Tags, "dnscontrol", "noraw") {
			continue
		}
		if HasTagOption(field.Tags, "dns", "a") {
			ac = append(ac, fmt.Sprintf("n.%s", field.Name))
		} else if field.Type == "fieldtypes.IPv4" {
			ac = append(ac, fmt.Sprintf("n.%s.String()", field.Name))
		} else {
			ac = append(ac, fmt.Sprintf("n.%s", field.Name))
		}
	}
	return strings.Join(ac, ", ")
}

// GetFieldsAsStringsTYPE

func makeGetFieldsAsStringsTYPE(rtconfig RTypeConfig) []byte {
	return rtypeTemplate(GetFieldsAsStringsTYPETmpl, rtconfig)
}

var GetFieldsAsStringsTYPETmpl = template.Must(template.New("GetFieldsAsStringsTYPE").Parse(`
// GetFieldsAsStrings{{ .Name }} returns rc.Fields as individual strings.
func (rc *RecordConfig) GetFieldsAsStrings{{ .Name }}() [{{ .NumRawFields }}]string {
  n := rc.As{{ .Name }}()
  return {{ .ReturnAsStringsList }}
}

`))

func mkReturnAsStringsList(fields []Field) string {
	var ac []string
	for _, field := range fields {
		if HasTagOption(field.Tags, "dnscontrol", "noraw") {
			continue
		}
		if HasTagOption(field.Tags, "dns", "a") {
			ac = append(ac, fmt.Sprintf("n.%s.String()", field.Name))
		} else if field.Type == "fieldtypes.IPv4" {
			ac = append(ac, fmt.Sprintf("n.%s.String()", field.Name))
		} else if field.Type == "uint16" {
			ac = append(ac, fmt.Sprintf("strconv.Itoa(int(n.%s))", field.Name))
		} else {
			ac = append(ac, fmt.Sprintf("n.%s", field.Name))
		}
	}
	return fmt.Sprintf("[%d]string{", len(ac)) + strings.Join(ac, ", ") + "}"
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

func makeIntTestConstructor(rtconfig RTypeConfig) []byte {
	return rtypeTemplate(IntTestConstructorTmpl, rtconfig)
}

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
  return models.MustCreateRecord(name, rdata, nil, 300, "**current-domain**")
}

`))

func mkFieldsAsSignature(fields []Field) string {
	var ac []string
	for _, field := range fields {
		if field.Type == "fieldtypes.IPv4" {
			// accept input as string.
			ac = append(ac, fmt.Sprintf("%s string", field.NameLower))
		} else {
			ac = append(ac, fmt.Sprintf("%s %s", field.NameLower, field.Type))
		}
	}
	return strings.Join(ac, ", ")
}

func mkConvertToString(f Field) string {
	if HasTagOption(f.Tags, "dns", "a") {
		return ""
	}

	switch f.Type {

	case "string":
		return ""

	case "uint16":
		return fmt.Sprintf("s%s := strconv.Itoa(int(%s))", f.NameLower, f.NameLower)

	case "int":
		return fmt.Sprintf("s%s := strconv.Itoa(%s)", f.NameLower, f.NameLower)
	}

	return fmt.Sprintf("s%s := UNKNOWN(int(%s))", f.NameLower, f.NameLower)

}

func mkFieldsAsSVars(fields []Field) string {
	var ac []string
	for _, field := range fields {
		if HasTagOption(field.Tags, "dns", "a") {
			ac = append(ac, field.NameLower)
		} else if field.Type == "string" {
			ac = append(ac, field.NameLower)
		} else {
			ac = append(ac, "s"+field.NameLower)
		}
	}
	return strings.Join(ac, ", ")
}

// helpersRawRecordBuilder

func makehelpersRawRecordBuilder(vals Values) []byte {
	return valuesTemplate(helpersRawRecordBuilderTmpl, vals)
}

var helpersRawRecordBuilderTmpl = template.Must(template.New("helpersRawRecordBuilder").Parse(`
{{- range .TypeNamesAndFields -}}
var {{ .Config.Token }} = rawrecordBuilder('{{ .Config.Token }}');
{{ end -}}
`))
