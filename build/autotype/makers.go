package main

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// funcs lists Go functions imported into the template system.
var funcs = template.FuncMap{
	"join": strings.Join,
}

// valuesTemplate executes a template with the entire Values struct.
func valuesTemplate(theTemplate *template.Template, vals Values) []byte {
	b := &bytes.Buffer{}
	err := theTemplate.Execute(b, vals)
	if err != nil {
		panic(err)
	}
	return b.Bytes()
}

// rtypeTemplate executes a template with values for a single RTypeConfig.
func rtypeTemplate(theTemplate *template.Template, vals RTypeConfig) []byte {
	b := &bytes.Buffer{}
	err := theTemplate.Execute(b, vals)
	if err != nil {
		panic(err)
	}
	return b.Bytes()
}

// RecordType

// makeInterfaceConstraint generates the RecordType interface constraint.
// Makes: `dnscontrol/models/generated_types.go` type RecordType
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

// makeInit generates the init() function that registers all types.
// Makes: `models/generated_types.go` func init()
func makeInit(vals Values) []byte {
	return valuesTemplate(RegisterTypeTmpl, vals)
}

var RegisterTypeTmpl = template.Must(template.New("RegisterType").Parse(`
package models

import (
	"fmt"
	"strconv"

	"github.com/StackExchange/dnscontrol/v4/pkg/fieldtypes"
	"github.com/qdm12/reprint"
)

func init() {
	{{- range .TypeNamesAndFields }}
	MustRegisterType("{{ .Config.Token }}", RegisterOpts{PopulateFromRaw: PopulateFromRaw{{ .Config.Name -}} })
	{{- end }}
}
`))

// ImportFromLegacy

// makeImportFromLegacy generates the function that upstreams legacy data to the upgraded record types.
// Makes: `models/generated_types.go` func ImportFromLegacy
func makeImportFromLegacy(vals Values) []byte {
	return valuesTemplate(importFromLegacyTmpl, vals)
}

var importFromLegacyTmpl = template.Must(template.New("ImportFromLegacy").Parse(`
// ImportFromLegacy copies the legacy fields (MxPreference, SrvPort, etc.) to
// the .Fields structure.  It is the reverse of Seal*().
func (rc *RecordConfig) ImportFromLegacy(origin string) error {

	if IsTypeLegacy(rc.Type) {
		// Nothing to convert!
		return nil
	}

	switch rc.Type {
{{- range .TypeNamesAndFields -}}
{{- if eq .Config.ConstructFromLegacyFields "IP" }}
	case "A":
		ip, err := fieldtypes.ParseIPv4(rc.target)
		if err != nil {
			return err
		}
		return RecordUpdateFields(rc, A{A: ip}, nil)
{{- else if .Config.ConstructFromLegacyFields }}
	case "{{ .Name }}":
		return RecordUpdateFields(rc,
			{{ .Name }}{ {{- .Config.ConstructFromLegacyFields -}} },
			nil,
		)
{{- end }}
{{- end }}
	}
	panic("Should not happen")
}
`))

func mkConstructFromLegacyFields(fields []Field) string {
	var ac []string
	for i, field := range fields {
		if HasTagOption(field.Tags, "dnscontrol", "noraw") {
			continue
		}
		if field.LegacyName != "" {
			ac = append(ac, fmt.Sprintf("%s: rc.%s", field.Name, field.LegacyName))
		} else if HasTagOption(field.Tags, "dns", "a") {
			return "IP"
		} else if i == (len(fields) - 1) {
			// The last field defaults to .target
			ac = append(ac, fmt.Sprintf("%s: rc.%s", field.Name, "target"))
		}
	}
	return strings.Join(ac, ", ")
}

// Seal

// makeSeal generates the function that downstreams upgraded fields data to the legacy record fields.
// Makes: `models/generated_types.go` func ImportFromLegacy
func makeSeal(vals Values) []byte {
	return valuesTemplate(importSealTmpl, vals)
}

var importSealTmpl = template.Must(template.New("Seal").Parse(`
func (rc *RecordConfig) Seal() error {
	rc.Type = GetTypeName(rc.Fields)

	// Copy the fields to the legacy fields:
	// Pre-compute useful things
	switch rc.Type {
{{- range .TypeNamesAndFields -}}
	{{- if eq .Config.ConstructFromLegacyFields "IP" }}
	case "{{ .Name }}":
		f := rc.Fields.(*{{ .Name }})
		rc.target = f.A.String()
		rc.Comparable = fmt.Sprintf("%d.%d.%d.%d", f.A[0], f.A[1], f.A[2], f.A[3])
	{{- else if .Config.ConstructFromLegacyFields }}
	case "{{ .Name }}":
		f := rc.Fields.(*{{ .Name }})
		{{- range .Config.Fields }}
			{{- if .LegacyName }}
		rc.{{ .LegacyName }} = f.{{ .Name }}
			{{- else }}
		rc.target = f.{{ .Name }}
			{{- end }}
		{{- end }}

		rc.Comparable = {{ .Config.ComparableExpr }}
	{{- else }}
	case "{{ .Config.Token }}":
		f := rc.Fields.(*{{ .Name }})
		{{- range .Config.Fields }}
			{{- if .LegacyName }}
		rc.{{ .LegacyName }} = f.{{ .Name }}
			{{- end }}
		{{- end }}

		rc.Comparable = {{ .Config.ComparableExpr }}
	{{- end }}
{{- end }}
	default:
		return fmt.Errorf("unknown (Seal) rtype %q", rc.Type)
	}
	rc.Display = rc.Comparable

	return nil
}
`))

func mkComparableExpr(fields []Field) string {
	fmt.Printf("DEBUG: mkComparableExpr(%+v)\n", fields)

	// Single field that is a string, return it.
	if len(fields) == 1 && fields[0].Type == "string" {
		return `f.` + fields[0].Name
	}

	// Otherwise, use fmt.Sprintf to generate it.
	var fl []string
	var al []string
	for _, field := range fields {
		fmt.Printf("DEBUG: field = %+v\n", field)
		if HasTagOption(field.Tags, "dnscontrol", "noinput") {
			continue
		}
		switch {
		case HasTagOption(field.Tags, "dns", "a"):
			fl = append(fl, `"%s"`)
			al = append(al, "f."+field.Name)
		case HasTagOption(field.Tags, "dnscontrol", "anyascii"):
			fl = append(fl, "%q")
			al = append(al, "f."+field.Name)
		case field.Type == "int" || field.Type == "uint16" || field.Type == "uint8":
			fl = append(fl, "%d")
			al = append(al, "f."+field.Name)
		case field.Type == "string":
			fl = append(fl, "%s")
			al = append(al, "f."+field.Name)
		default:
			fl = append(fl, "%v")
			//fl = append(fl, fmt.Sprintf("%%v(%s)", field.Type))
			al = append(al, "f."+field.Name)
		}
	}
	x := `fmt.Sprintf("` + strings.Join(fl, " ") + `", ` + strings.Join(al, ", ") + `)`
	fmt.Printf("DEBUG: return = %v\n", x)
	return x
}

// Copy

// makeCopy generates the copy function.
// Makes: `models/generated_types.go` func Copy
func makeCopy(vals Values) []byte {
	return valuesTemplate(copyTmpl, vals)
}

var copyTmpl = template.Must(template.New("Copy").Parse(`
// Copy returns a deep copy of a RecordConfig.
func (rc *RecordConfig) Copy() (*RecordConfig, error) {
	newR := &RecordConfig{}
	// Copy the exported fields.
	err := reprint.FromTo(rc, newR) // Deep copy
	// Copy each unexported field.
	newR.target = rc.target

	// Copy the fields to new memory so there is no aliasing.
	switch rc.Type {
	{{- range .TypeNamesAndFields }}
	case "{{ .Name }}":
		newR.Fields = &{{ .Name }}{}
		newR.Fields = rc.Fields.(*{{ .Name }})
	{{- end }}
	}
	//fmt.Printf("DEBUG: COPYING rc=%v new=%v\n", rc.Fields, newR.Fields)
	return newR, err
}
`))

// GetTargetField

// makeGetTargetField generates the function that returns the last field of a record type.
// Makes: `models/generated_types.go` func GetTargetField
func makeGetTargetField(vals Values) []byte {
	return valuesTemplate(getTargetFieldTmpl, vals)
}

var getTargetFieldTmpl = template.Must(template.New("GetTargetField").Parse(`
// GetTargetField returns the target. There may be other fields, but they are
// not included. For example, the .MxPreference field of an MX record isn't included.
func (rc *RecordConfig) GetTargetField() string {
	switch rc.Type { // #rtype_variations
{{- range .TypeNamesAndFields -}}
	{{- if eq .Config.ConstructFromLegacyFields "IP" }}
	case "{{ .Name }}":
		return rc.As{{ .Name }}().{{ .Name }}.String()
	{{- else }}
	case "{{ .Name }}":
		return rc.As{{ .Name }}().{{ .Config.TargetField }}
	{{- end }}
{{- end }}
	}
	return rc.target
}
`))

// mkTargetField returns the field that an average person would consider the
// "target" field of this rtype.  This is used in the GetTargetField function,
// which is kind of ambiguous and should be removed.
func mkTargetField(fields []Field) (string, string) {
	// If there is only one field, use it.
	if len(fields) == 1 {
		return fields[0].Name, fields[0].LegacyName
	}

	// Use the field named "Target", or tagged as the target, or if the legacy
	// name is "target".
	for _, field := range fields {
		if field.Name == "Target" || HasTagOption(field.Tags, "dnscontrol", "target") || field.LegacyName == "target" {
			return field.Name, field.LegacyName
		}
	}

	// Really? No target field?  Use the last field.
	return fields[len(fields)-1].Name, fields[len(fields)-1].LegacyName
}

// // SetTarget

// // makeSetTarget generates the function that sets the last field of a record type.
// // Makes: `models/generated_types.go` func SetTarget
// func makeSetTarget(vals Values) []byte {
// 	return valuesTemplate(importSetTargetTmpl, vals)
// }

// var importSetTargetTmpl = template.Must(template.New("SetTarget").Parse(`
// // SetTarget sets the target, assuming that the rtype is appropriate.
// func (rc *RecordConfig) SetTarget(s string) error {
// 	// Legacy
// 	rc.target = s

// 	switch rc.Type { // #rtype_variations
// {{- range .TypeNamesAndFields -}}
// 	{{- if eq .Config.ConstructFromLegacyFields "IP" }}
// 	case "A":
// 		return rc.SetTargetA(s)
// 	{{- else }}
// 	case "{{ .Name }}":
// 		// COUNT={{- .Config.NumRawFields }}
// 		{{- if ne .Config.NumRawFields 1 }}
// 		if rc.Fields == nil {
// 			return rc.SetTarget{{ .Name }}(rc.{{ .Config.LegacyTargetField }}, s)
// 		}
// 		{{- end }}
// 		f := rc.As{{ .Name }}()
// 		return rc.SetTarget{{ .Name }}(f.{{ .Config.TargetField }}, s)
// 	{{- end }}
// {{- end }}
// 	}
// 	return nil
// `))

// mkTargetField() defined above

// PopulateFromFields

// makePopulateFromFields generates the function that populates the Fields from individual strings.
// Makes: `models/generated_types.go` func PopulateFromFields
func makePopulateFromFields(vals Values) []byte {
	return valuesTemplate(PopulateFromFieldsTmpl, vals)
}

var PopulateFromFieldsTmpl = template.Must(template.New("PopulateFromFields").Parse(`
func PopulateFromFields(rc *RecordConfig, rtype string, fields []string, origin string) error {
	switch rtype {
{{- range .TypeNamesAndFields }}
	case "{{ .Name }}":
		if rdata, err := Parse{{ .Name }}(fields, origin); err == nil {
		return RecordUpdateFields(rc, rdata, nil)
	}
{{- end }}
	}
	return fmt.Errorf("rtype %q not found (%v)", rtype, fields)
 }
`))

// TypeTYPE

// makeTypeTYPE generates the Type{TYPE} for a record type.
// Makes: `models/generated_types.go` type Type{TYPE}
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

// makeParseTYPE generates the func Parse{TYPE} for a record type.
// Makes: `models/generated_types.go` func Parse{TYPE}
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

func mkParser(i int, f Field) string {
	switch ty := f.Type; ty {
	case "int":
		if HasTagOption(f.Tags, "dnscontrol", "redirectcode") {
			return fmt.Sprintf(`fieldtypes.ParseRedirectCode(rawfields[%d], "", origin)`, i)
		}
		return fmt.Sprintf(`fieldtypes.ParseStringTrimmed(rawfields[%d])`, i)
	case "string":
		//fmt.Printf("DEBUG: parserFor(%d, %+v) ... %v\n", i, f, HasTagOption(f.Tags, "dns", "cdomain-name"))
		if HasTagOption(f.Tags, "dns", "cdomain-name") || HasTagOption(f.Tags, "dns", "domain-name") {
			return fmt.Sprintf(`fieldtypes.ParseHostnameDot(rawfields[%d], "", origin)`, i)
		}
		if HasTagOption(f.Tags, "dnscontrol", "allcaps") {
			return fmt.Sprintf(`fieldtypes.ParseStringTrimmedAllCaps(rawfields[%d])`, i)
		}
		return fmt.Sprintf(`fieldtypes.ParseStringTrimmed(rawfields[%d])`, i)
	case "fieldtypes.IPv4":
		return fmt.Sprintf(`fieldtypes.ParseIPv4(rawfields[%d])`, i)
	}

	return fmt.Sprintf(`fieldtypes.Parse%s(rawfields[%d])`, capFirst(f.Type), i)
}

func mkConstructAll(fields []Field) string {
	var ac []string
	for _, field := range fields {
		if HasTagOption(field.Tags, "dnscontrol", "srdisplay") {
			ac = append(ac, fmt.Sprintf(`%s: cfSingleRedirecttargetFromRaw(srname, code, srwhen, srthen)`, field.Name))
		} else if HasTagOption(field.Tags, "dnscontrol", "parsereturnunknowable") {
			ac = append(ac, fmt.Sprintf(`%s: "UNKNOWABLE"`, field.Name))
		} else if HasTagOption(field.Tags, "dnscontrol", "noinput") {
			// Skip this field.
		} else {
			ac = append(ac, fmt.Sprintf("%s: %s", field.Name, field.NameLower))
		}
	}
	return strings.Join(ac, ", ")
}

// PopulateFromRawTYPE

// makePopulateFromRawTYPE generates the func PopulateFromRaw{TYPE} for a given record type.
// Makes: `models/generated_types.go` func PopulateFromRaw{TYPE}
func makePopulateFromRawTYPE(rtconfig RTypeConfig) []byte {
	return rtypeTemplate(PopulateFromRawTYPETmpl, rtconfig)
}

var PopulateFromRawTYPETmpl = template.Must(template.New("PopulateFromRawTYPE").Parse(`
// PopulateFromRaw{{ .Name }} updates rc to be an {{ .Name }} record with contents from rawfields, meta and origin.
func PopulateFromRaw{{ .Name }}(rc *RecordConfig, rawfields []string, meta map[string]string, origin string) error {
	{{ if .IsBuilder -}}
	rawfields, meta, err := Builder{{ .Name }}(rawfields, meta, origin)
	if err != nil {
		return err
	}

	{{ end -}}

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

// makeAsTYPE generates the func As{TYPE} which returns the type struct.
// Makes: `models/generated_types.go` func As{TYPE}
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

// makeGetFieldsTYPE generates the method func GetFields{TYPE} for a given record type.
// Makes: `models/generated_types.go` func GetFields{TYPE}
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

// makeGetFieldsAsStringsTYPE generates the GetFieldsAsStrings{TYPE} function.
// Makes: `models/generated_types.go` func GetFieldsAsStrings{TYPE}
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
		} else if field.Type == "uint16" || field.Type == "uint8" {
			ac = append(ac, fmt.Sprintf("strconv.Itoa(int(n.%s))", field.Name))
		} else {
			ac = append(ac, fmt.Sprintf("n.%s", field.Name))
		}
	}
	return fmt.Sprintf("[%d]string{", len(ac)) + strings.Join(ac, ", ") + "}"
}

// SetTargetTYPE

// makeSetTargetTYPE generates the SetTarget{TYPE} function.
// Makes: `models/generated_types.go` func SetTarget{TYPE}
func makeSetTargetTYPE(rtconfig RTypeConfig) []byte {
	return rtypeTemplate(SetTargetTYPETmpl, rtconfig)
}

var SetTargetTYPETmpl = template.Must(template.New("SetTargetTYPE").Parse(`
// SetTarget{{ .Name }} sets the {{ .Name }} fields.
func (rc *RecordConfig) SetTarget{{ .Name }}({{ .InputFieldsAsSignature }}) error {
	rc.Type = "{{ .Name }}"
{{- if eq .ConstructFromLegacyFields "IP" }}
	rdata, err := ParseA([]string{a}, "")
	if err != nil {
		return err
	}
	return RecordUpdateFields(rc, rdata, nil)
{{- else }}
	return RecordUpdateFields(rc, {{ .Name }}{ {{- .ConstructAll -}} }, nil)
{{- end }}
}
`))

// mkConstructAll() defined above

// IntTestHeader

// makeIntTestHeader generates the header for the integration tests data.
// Makes: `integrationTest/generated_helpers.go` package main
func makeIntTestHeader() []byte {
	return []byte(`package main

import (
	"strconv"

	"github.com/StackExchange/dnscontrol/v4/models"
)

`)
}

// IntTestConstructor

// makeIntTestConstructor makes the Integration Test helper function that constructs a {type}.
// Makes: integrationTest/generated_helpers.go func {type}
func makeIntTestConstructor(rtconfig RTypeConfig) []byte {
	return rtypeTemplate(IntTestConstructorTmpl, rtconfig)
}

var IntTestConstructorTmpl = template.Must(template.New("IntTestConstructor").Parse(`
{{- if .NoLabel }}
func {{ .NameLower }}({{ .InputFieldsAsSignature }}) *models.RecordConfig {
{{- else }}
func {{ .NameLower }}(name string, {{ .InputFieldsAsSignature }}) *models.RecordConfig {
{{- end }}
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

func mkInputFieldsAsSignature(fields []Field) string {
	var ac []string
	for _, field := range fields {
		if HasTagOption(field.Tags, "dnscontrol", "noinput") {
			continue
		}
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

	if HasTagOption(f.Tags, "dnscontrol", "label") {
		// When NoLabel is in use, this marks the actual label.
		return fmt.Sprintf("name := %s", f.NameLower)
	}

	// Skip fields that are not input.
	if HasTagOption(f.Tags, "dns", "a") || HasTagOption(f.Tags, "dnscontrol", "noinput") {
		return ""
	}

	switch f.Type {

	case "string":
		return ""

	case "uint16", "uint8":
		return fmt.Sprintf("s%s := strconv.Itoa(int(%s))", f.NameLower, f.NameLower)

	case "int":
		return fmt.Sprintf("s%s := strconv.Itoa(%s)", f.NameLower, f.NameLower)
	}

	// There is no "UNKNOWN() function, but this will cause a compile error
	// which indicates this function needs to add a conversion.
	return fmt.Sprintf("s%s := UNKNOWN(int(%s))", f.NameLower, f.NameLower)

}

func mkFieldsAsSVars(fields []Field) string {
	var ac []string
	for _, field := range fields {
		if HasTagOption(field.Tags, "dnscontrol", "noinput") {
			continue
		}
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

// makehelpersRawRecordBuilder generates the helpers-types.js entry for {TYPE}.
// Makes `pkg/js/helpers-types.js` var {TYPE
func makehelpersRawRecordBuilder(vals Values) []byte {
	return valuesTemplate(helpersRawRecordBuilderTmpl, vals)
}

var helpersRawRecordBuilderTmpl = template.Must(template.New("helpersRawRecordBuilder").Parse(`
{{- range .TypeNamesAndFields -}}
var {{ .Config.Token }} = rawrecordBuilder('{{ .Config.Token }}');
{{ end -}}
`))
