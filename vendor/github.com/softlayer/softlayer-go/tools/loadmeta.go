/**
 * Copyright 2016 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"
	"text/template"

	"golang.org/x/tools/imports"
)

const SoftLayerMetadataAPIURL = "https://api.softlayer.com/metadata/v3.1"

type Type struct {
	Name       string              `json:"name"`
	Base       string              `json:"base"`
	TypeDoc    string              `json:"typeDoc"`
	Properties map[string]Property `json:"properties"`
	ServiceDoc string              `json:"serviceDoc"`
	Methods    map[string]Method   `json:"methods"`
	NoService  bool                `json:"noservice"`
}

type Property struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	TypeArray bool   `json:"typeArray"`
	Form      string `json:"form"`
	Doc       string `json:"doc"`
}

type Method struct {
	Name       string      `json:"name"`
	Type       string      `json:"type"`
	TypeArray  bool        `json:"typeArray"`
	Doc        string      `json:"doc"`
	Static     bool        `json:"static"`
	NoAuth     bool        `json:"noauth"`
	Limitable  bool        `json:"limitable"`
	Filterable bool        `json:"filterable"`
	Maskable   bool        `json:"maskable"`
	Parameters []Parameter `json:"parameters"`
}

type Parameter struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	TypeArray    bool        `json:"typeArray"`
	Doc          string      `json:"doc"`
	DefaultValue interface{} `json:"defaultValue"`
}

// Define custom template functions
var fMap = template.FuncMap{
	"convertType":     ConvertType,         // Converts SoftLayer types to Go types
	"removePrefix":    RemovePrefix,        // Remove 'SoftLayer_' prefix. if it exists
	"removeReserved":  RemoveReservedWords, // Substitute language-reserved identifiers
	"titleCase":       strings.Title,       // TitleCase the argument
	"desnake":         Desnake,             // Remove '_' from Snake_Case
	"goDoc":           GoDoc,               // Format a go doc string
	"tags":            Tags,                // Remove omitempty tags if required
	"phraseMethodArg": phraseMethodArg,     // Get proper phrase for method argument
}

var datatype = fmt.Sprintf(`%s

%s

package datatypes

{{range .}}{{.TypeDoc|goDoc}}
type {{.Name|removePrefix}} struct {
	{{.Base|removePrefix}}

	{{$base := .Name}}{{range .Properties}}{{.Doc|goDoc}}
	{{.Name|titleCase}} {{if .TypeArray}}[]{{else}}*{{end}}{{convertType .Type "datatypes" $base .Name}}`+
	"`json:\"{{.Name|tags}}\" xmlrpc:\"{{.Name|tags}}\"`"+`

	{{end}}
}

{{end}}
`, license, codegenWarning)

var services = fmt.Sprintf(`%s

%s

package services

import (
	"fmt"
	"strings"
)

{{range .}}{{$base := .Name|removePrefix}}{{.TypeDoc|goDoc}}
	type {{$base}} struct {
		Session *session.Session
		Options sl.Options
	}

	// Get{{$base | desnake}}Service returns an instance of the {{$base}} SoftLayer service
	func Get{{$base | desnake}}Service(sess *session.Session) {{$base}} {
		return {{$base}}{Session: sess}
	}

	func (r {{$base}}) Id(id int) {{$base}} {
		r.Options.Id = &id
		return r
	}

	func (r {{$base}}) Mask(mask string) {{$base}} {
		if !strings.HasPrefix(mask, "mask[") && (strings.Contains(mask, "[") || strings.Contains(mask, ",")) {
			mask = fmt.Sprintf("mask[%%s]", mask)
		}

		r.Options.Mask = mask
		return r
	}

	func (r {{$base}}) Filter(filter string) {{$base}} {
		r.Options.Filter = filter
		return r
	}

	func (r {{$base}}) Limit(limit int) {{$base}} {
		r.Options.Limit = &limit
		return r
	}

	func (r {{$base}}) Offset(offset int) {{$base}} {
		r.Options.Offset = &offset
		return r
	}

	{{$rawBase := .Name}}{{range .Methods}}{{$methodName := .Name}}{{.Doc|goDoc}}
	func (r {{$base}}) {{.Name|titleCase}}({{range .Parameters}}{{phraseMethodArg $methodName .Name .TypeArray .Type}}{{end}}) ({{if .Type|ne "void"}}resp {{if .TypeArray}}[]{{end}}{{convertType .Type "services"}}, {{end}}err error) {
		{{if .Type|eq "void"}}var resp datatypes.Void
		{{end}}{{if or (eq .Name "placeOrder") (eq .Name "verifyOrder")}}err = datatypes.SetComplexType(orderData)
		if err != nil {
			return
		}
		{{end}}{{if len .Parameters | lt 0}}params := []interface{}{
			{{range .Parameters}}{{.Name|removeReserved}},
			{{end}}
		}
		{{end}}err = r.Session.DoRequest("{{$rawBase}}", "{{.Name}}", {{if len .Parameters | lt 0}}params{{else}}nil{{end}}, &r.Options, &resp)
	return
	}
	{{end}}

{{end}}
`, license, codegenWarning)

func generateAPI() {
	var meta map[string]Type

	flagset := flag.NewFlagSet(os.Args[1], flag.ExitOnError)
	outputPath := flag.String("o", ".", "the root of the go project to be refreshed")
	flagset.Parse(os.Args[2:])

	jsonResp, code, err := makeHttpRequest(SoftLayerMetadataAPIURL, "GET", new(bytes.Buffer))

	if err != nil {
		fmt.Printf("Error retrieving metadata API: %s", err)
		os.Exit(1)
	}

	if code != 200 {
		fmt.Printf("Unexpected HTTP status code received while retrieving metadata API: %d", code)
		os.Exit(1)
	}

	err = json.Unmarshal(jsonResp, &meta)
	if err != nil {
		fmt.Printf("Error unmarshaling json response: %s", err)
		os.Exit(1)
	}

	// Build an array of Types, sorted by name
	// This will ensure consistency in the order that code is later emitted
	keys := getSortedKeys(meta)

	sortedTypes := make([]Type, 0, len(keys))
	sortedServices := make([]Type, 0, len(keys))

	for _, name := range keys {
		t := meta[name]
		sortedTypes = append(sortedTypes, t)
		addComplexType(&t)
		fixDatatype(&t, meta)

		// Not every datatype is also a service
		if !t.NoService {
			createGetters(&t)
			sortedServices = append(sortedServices, t)
		}
	}

	// Services can be subclasses of other services. Copy methods from each service's 'Base' entity to
	// the child service, only if a same-named method does not already exist (i.e., overridden by the
	// child service)
	for i, service := range sortedServices {
		sortedServices[i].Methods = getBaseMethods(service, meta)
		fixReturnType(&sortedServices[i])
	}

	err = writePackage(*outputPath, "datatypes", sortedTypes, datatype)
	if err != nil {
		fmt.Printf("Error writing to file: %s", err)
	}

	err = writePackage(*outputPath, "services", sortedServices, services)
	if err != nil {
		fmt.Printf("Error writing to file: %s", err)
	}
}

// Exported template functions

func RemovePrefix(args ...interface{}) string {
	s := args[0].(string)

	if strings.HasPrefix(s, "SoftLayer_") {
		return s[10:]
	}

	return s
}

// ConvertType takes the name of the type to convert, and the package context.
func ConvertType(args ...interface{}) string {
	t := args[0].(string)
	p := args[1].(string)

	// Convert softlayer types to golang types
	switch t {
	case "unsignedLong", "unsignedInt":
		return "uint"
	case "boolean":
		return "bool"
	case "dateTime":
		if p != "datatypes" {
			return "datatypes.Time"
		} else {
			return "Time"
		}
	case "decimal", "float":
		if p != "datatypes" {
			return "datatypes.Float64"
		} else {
			return "Float64"
		}
	case "base64Binary":
		return "[]byte"
	case "json", "enum":
		return "string"
	}

	if strings.HasPrefix(t, "SoftLayer_") {
		t = RemovePrefix(t)
		if p != "datatypes" {
			return "datatypes." + t
		}
		return t
	}

	if strings.HasPrefix(t, "McAfee_") {
		if p != "datatypes" {
			return "datatypes." + t
		}
		return t
	}

	return t
}

func RemoveReservedWords(args ...interface{}) string {
	n := args[0].(string)

	// Replace language reserved identifiers with alternatives
	switch n {
	case "type":
		return "typ"
	}

	return n
}

// Remove '_' from Snake_Case values
func Desnake(args ...interface{}) string {
	s := args[0].(string)
	return strings.Replace(s, "_", "", -1)
}

// Formats a string into a comment.  For now, just each comment line with "//"
func GoDoc(args ...interface{}) string {
	s := args[0].(string)
	if s == "" {
		s = "no documentation yet"
	}

	return "// " + strings.Replace(s, "\n", "\n// ", -1)
}

// Remove omitempty tags if required
func Tags(args ...interface{}) string {
	n := args[0].(string)

	switch n {
	case "resourceRecords":
		return n
	default:
		return n + ",omitempty"
	}
}

// private

func createGetters(service *Type) {
	for _, p := range service.Properties {
		if p.Form == "relational" {
			m := Method{
				Name:       "get" + strings.Title(p.Name),
				Type:       p.Type,
				TypeArray:  p.TypeArray,
				Doc:        "Retrieve " + p.Doc, // TODO lowercase the first letter
				Parameters: []Parameter{},
			}

			service.Methods[m.Name] = m
		}
	}
}

// Special case for ensuring we can set a complexType on product orders.
func addComplexType(dataType *Type) {
	// Only adding this to the base product order type. All others embed this one.
	if dataType.Name == "SoftLayer_Container_Product_Order" {
		dataType.Properties["complexType"] = Property{
			Name: "complexType",
			Type: "string",
			Form: "local",
			Doc:  "Added by softlayer-go. This hints to the API what kind of product order this is.",
		}
	} else if dataType.Name == "SoftLayer_Container_User_Customer_External_Binding" {
		dataType.Properties["complexType"] = Property{
			Name: "complexType",
			Type: "string",
			Form: "local",
			Doc:  "Added by softlayer-go. This hints to the API what kind of binding this is.",
		}
	}
}

// Special case for fixing some datatype properties in the metadata
func fixDatatype(t *Type, meta map[string]Type) {
	if strings.HasPrefix(t.Name, "SoftLayer_Dns_Domain_ResourceRecord_") {
		baseRecordType, _ := meta["SoftLayer_Dns_Domain_ResourceRecord"]
		for propName, prop := range t.Properties {
			baseRecordType.Properties[propName] = prop
		}
		meta["SoftLayer_Dns_Domain_ResourceRecord"] = baseRecordType
	} else if t.Name == "SoftLayer_Container_User_Customer_External_Binding_Verisign" || t.Name == "SoftLayer_Container_User_Customer_External_Binding_Verisign_Totp" {
		baseType, _ := meta["SoftLayer_Container_User_Customer_External_Binding"]
		for propName, prop := range t.Properties {
			baseType.Properties[propName] = prop
		}
		meta["SoftLayer_Container_User_Customer_External_Binding"] = baseType
	}
}

// Special case for fixing some broken return types in the metadata
func fixReturnType(service *Type) {
	brokenServices := map[string]string{
		"SoftLayer_Network_Application_Delivery_Controller_LoadBalancer_Service":       "deleteObject",
		"SoftLayer_Network_Application_Delivery_Controller_LoadBalancer_VirtualServer": "deleteObject",
		"SoftLayer_Network_Application_Delivery_Controller":                            "deleteLiveLoadBalancerService",
	}

	if methodName, ok := brokenServices[service.Name]; ok {
		method := service.Methods[methodName]
		method.Type = "void"
		service.Methods[methodName] = method
	}
}

// Return formatted method argument phrase used by the method generation.
func phraseMethodArg(methodName string, argName string, isArray bool, argType string) string {
	argName = RemoveReservedWords(argName)

	// Handle special case - placeOrder/verifyOrder should take any kind of order type.
	if (methodName == "placeOrder" || methodName == "verifyOrder") &&
		strings.HasPrefix(argType, "SoftLayer_Container_Product_Order") {
		return fmt.Sprintf("%s interface{}, ", argName)
	}

	refPrefix := "*"
	if isArray {
		refPrefix = "[]"
	}

	argType = ConvertType(argType, "services")

	return fmt.Sprintf("%s %s%s, ", argName, refPrefix, argType)
}

func combineMethods(baseMethods map[string]Method, subclassMethods map[string]Method) map[string]Method {
	r := map[string]Method{}

	// Copy all subclass methods into the result set
	for k, v := range subclassMethods {
		r[k] = v
	}

	// Copy each method from the base class into the result set, but only if a like-named method
	// does not already exist (a method in the child should override a same-named method in the parent)
	for k, v := range baseMethods {
		if _, ok := r[k]; !ok {
			r[k] = v
		}
	}

	return r
}

func getBaseMethods(s Type, typeMap map[string]Type) map[string]Method {
	var methods, baseMethods map[string]Method

	methods = s.Methods

	if s.Base != "SoftLayer_Entity" {
		baseMethods = getBaseMethods(typeMap[s.Base], typeMap)

		// Add base methods to current service methods
		methods = combineMethods(baseMethods, methods)
	}

	// return my methods
	return methods
}

func getSortedKeys(m map[string]Type) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	return keys
}

func writePackage(base string, pkg string, meta []Type, ts string) error {
	var currPrefix string
	var start int

	for i, t := range meta {
		components := strings.Split(RemovePrefix(t.Name), "_")

		if i == 0 {
			currPrefix = components[0]
			continue
		}

		if components[0] != currPrefix {
			err := writeGoFile(base, pkg, currPrefix, meta[start:i], ts)
			if err != nil {
				return err
			}

			currPrefix = components[0]
			start = i
		}
	}

	writeGoFile(base, pkg, currPrefix, meta[start:], ts)

	return nil
}

// Executes a template against the metadata structure, and generates a go source file with the result
func writeGoFile(base string, pkg string, name string, meta []Type, ts string) error {
	filename := base + "/" + pkg + "/" + strings.ToLower(name) + ".go"

	// Generate the source
	var buf bytes.Buffer
	t := template.New(pkg).Funcs(fMap)
	template.Must(t.Parse(ts)).Execute(&buf, meta)

	/*if pkg == "services" && name == "Account"{
		fmt.Println(string(buf.String()))
		os.Exit(0)
	}*/

	// Add the imports
	src, err := imports.Process(filename, buf.Bytes(), &imports.Options{Comments: true})
	if err != nil {
		fmt.Printf("Error processing imports: %s", err)
	}

	// Format
	pretty, err := format.Source(src)
	if err != nil {
		return fmt.Errorf("Error while formatting source: %s", err)
	}

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("Error creating file: %s", err)
	}
	defer f.Close()
	fmt.Fprintf(f, "%s", pretty)

	return nil
}

func makeHttpRequest(url string, requestType string, requestBody *bytes.Buffer) ([]byte, int, error) {
	client := http.DefaultClient

	req, err := http.NewRequest(requestType, url, requestBody)
	if err != nil {
		return nil, 0, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, 520, err
	}

	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return responseBody, resp.StatusCode, nil
}
