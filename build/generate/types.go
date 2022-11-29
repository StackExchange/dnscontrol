package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/gernest/front"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func join(parts ...string) string {
	return strings.Join(parts, string(os.PathSeparator))
}

func fixRuns(s string) string {
	lines := strings.Split(s, "\n")
	var out []string
	for _, line := range lines {
		if len(line) == 0 {
			if len(out) > 0 && len(out[len(out)-1]) == 0 {
				continue
			}
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

var returnTypes = map[string]string{
	"domain": "DomainModifier",
	"global": "void",
	"record": "RecordModifier",
}
var paramTypeDefaults = map[string]string{
	"name": "string",
	"target": "string",
	"value": "string",
	"destination": "string",
	"address": "string | number",
	"priority": "number",
	"registrar": "string",
	"source": "string",
	"ttl": "Duration",
	"modifiers...": "RecordModifier[]",
}

func generateTypes() error {
	funcs := []Function{}

	m := front.NewMatter()
	m.Handle("---", front.YAMLHandler)

	types, err := os.ReadDir(join("docs", "_functions"))
	if err != nil {
		return err
	}
	for _, t := range types {
		if !t.IsDir() {
			return errors.New("not a directory: " + join("docs", "_functions", t.Name()))
		}
		funcNames, err := os.ReadDir(join("docs", "_functions", t.Name()))
		if err != nil {
			return err
		}

		for _, f := range funcNames {
			if f.IsDir() {
				return errors.New("not a file: " + join("docs", "_functions", t.Name(), f.Name()))
			}
			println("Processing ", join("docs", "_functions", t.Name(), f.Name()))
			content, err := os.ReadFile(join("docs", "_functions", t.Name(), f.Name()))
			if err != nil {
				return err
			}
			frontMatter, body, err := m.Parse(bytes.NewReader(content))
			if err != nil {
				return err
			}
			if frontMatter["ts_ignore"] == true {
				continue
			}

			params := []string{}
			if frontMatter["parameters"] != nil {
				for _, p := range frontMatter["parameters"].([]interface{}) {
					params = append(params, p.(string))
				}
			}

			body = body + "\n"
			body = strings.ReplaceAll(body, "{{site.github.url}}", "https://dnscontrol.org/")
			body = strings.ReplaceAll(body, "{% capture example %}\n", "")
			body = strings.ReplaceAll(body, "{% capture example2 %}\n", "")
			body = strings.ReplaceAll(body, "{% endcapture %}\n", "")
			body = strings.ReplaceAll(body, "{% include example.html content=example %}\n", "")
			body = strings.ReplaceAll(body, "{% include example.html content=example2 %}\n", "")
			body = strings.ReplaceAll(body, "](#", "](https://dnscontrol.org/js#")
			body = fixRuns(body)

			suppliedParamTypes := map[string]string{}
			if frontMatter["parameter_types"] != nil {
				rawTypes := frontMatter["parameter_types"].(map[interface {}]interface {})
				for k, v := range rawTypes {
					suppliedParamTypes[k.(string)] = v.(string)
				}
			}

			paramTypes := []string{}
			for _, p := range params {
				// start with supplied type, fall back to defaultParamType
				paramType := suppliedParamTypes[p]
				if paramType == "" {
					paramType = paramTypeDefaults[p]
				}
				if paramType == "" {
					println("WARNING:", f.Name() + ":", "no type for parameter ", "'" + p + "'")
					paramType = "unknown"
				}
				paramTypes = append(paramTypes, paramType)
			}

			returnType := returnTypes[t.Name()]
			if frontMatter["ts_return"] != nil {
				returnType = frontMatter["ts_return"].(string)
			} else if frontMatter["return"] != nil {
				returnType = frontMatter["return"].(string)
			}

			funcs = append(funcs, Function{
				Name:        frontMatter["name"].(string),
				Params:      params,
				ParamTypes:  paramTypes,
				ObjectParam: frontMatter["parameters_object"] == true,
				Deprecated:  frontMatter["deprecated"] == true,
				ReturnType:  returnType,
				Description: strings.TrimSpace(body),
			})
		}
	}

	content := ""
	for _, f := range funcs {
		content += f.String()
	}
	return os.WriteFile("docs/_includes/functions.d.ts", []byte(content), 0644)
}

type Function struct {
	Name        string
	Params      []string
	ParamTypes  []string
	ObjectParam bool
	Deprecated  bool
	ReturnType  string
	Description string
}

var caser = cases.Title(language.AmericanEnglish)

func (f Function) formatParams() string {
	var params []string
	for i, p := range f.Params {
		typeName := f.ParamTypes[i]
		name := p
		if strings.HasSuffix(name, "...") {
			name = "..." + name[:len(name)-3]
		}
		if strings.HasSuffix(typeName, "?") {
			typeName = typeName[:len(typeName)-1]
			name += "?"
		}
		if strings.Contains(name, " ") {
			name = strings.ReplaceAll(caser.String(p), " ", "")
			name = strings.ToLower(name[:1]) + name[1:]
		}
		params = append(params, fmt.Sprintf("%s: %s", name, typeName))
	}
	if f.ObjectParam {
		return "opts: { " + strings.Join(params, "; ") + " }"
	} else {
		return strings.Join(params, ", ")
	}
}

func (f Function) Docs() string {
	content := f.Description
	if f.Deprecated {
		content += "\n\n@deprecated"
	}
	return "/**\n * " + strings.ReplaceAll(content, "\n", "\n * ") + "\n */"
}

func (f Function) String() string {
	return fmt.Sprintf("%s\ndeclare function %s(%s): %s;\n\n", f.Docs(), f.Name, f.formatParams(), f.ReturnType)
}

