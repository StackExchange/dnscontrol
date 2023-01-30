package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
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

var delimiterRegex = regexp.MustCompile(`(?m)^---\n`)

func parseFrontMatter(content string) (map[string]interface{}, string, error) {
	delimiterIndices := delimiterRegex.FindAllStringIndex(content, 2)
	startIndex := delimiterIndices[0][0]
	endIndex := delimiterIndices[1][0]
	yamlString := content[startIndex+4 : endIndex]
	var frontMatter map[string]interface{}
	err := yaml.Unmarshal([]byte(yamlString), &frontMatter)
	if err != nil {
		return nil, "", err
	}
	return frontMatter, content[endIndex+4:], nil
}

var returnTypes = map[string]string{
	"domain": "DomainModifier",
	"global": "void",
	"record": "RecordModifier",
}

func generateFunctionTypes() (string, error) {
	funcs := []Function{}

	srcRoot := join("documentation", "functions")
	types, err := os.ReadDir(srcRoot)
	if err != nil {
		return "", err
	}
	for _, t := range types {
		if !t.IsDir() {
			return "", errors.New("not a directory: " + join(srcRoot, t.Name()))
		}
		tPath := join(srcRoot, t.Name())
		funcNames, err := os.ReadDir(tPath)
		if err != nil {
			return "", err
		}

		for _, f := range funcNames {
			fPath := join(tPath, f.Name())
			if f.IsDir() {
				return "", errors.New("not a file: " + fPath)
			}
			// println("Processing", fPath)
			content, err := os.ReadFile(fPath)
			if err != nil {
				return "", err
			}
			frontMatter, body, err := parseFrontMatter(string(content))
			if err != nil {
				println("Error parsing front matter in", fPath)
				return "", err
			}
			if frontMatter["ts_ignore"] == true {
				continue
			}

			body = body + "\n"
			body = strings.ReplaceAll(body, "{% hint style=\"danger\" %}", "")
			body = strings.ReplaceAll(body, "{% hint style=\"info\" %}", "")
			body = strings.ReplaceAll(body, "{% hint style=\"success\" %}", "")
			body = strings.ReplaceAll(body, "{% hint style=\"warning\" %}", "")
			body = strings.ReplaceAll(body, "{% endhint %}", "")
			body = strings.ReplaceAll(body, "**NOTE**", "NOTE")
			body = strings.ReplaceAll(body, "**WARNING**", "WARNING")
			body = fixRuns(body)

			paramNames := []string{}
			if frontMatter["parameters"] != nil {
				for _, p := range frontMatter["parameters"].([]interface{}) {
					paramNames = append(paramNames, p.(string))
				}
			}

			suppliedParamTypes := map[string]string{}
			if frontMatter["parameter_types"] != nil {
				rawTypes := frontMatter["parameter_types"].(map[string]interface{})
				for k, v := range rawTypes {
					suppliedParamTypes[k] = v.(string)
				}
			}

			params := []Param{}
			for _, p := range paramNames {
				// start with supplied type, fall back to defaultParamType
				paramType := suppliedParamTypes[p]
				if paramType == "" {
					println("WARNING:", fPath+":", "no type for parameter ", "'"+p+"'")
					paramType = "unknown"
				}
				params = append(params, Param{Name: p, Type: paramType})
			}

			returnType := returnTypes[t.Name()]
			if frontMatter["ts_return"] != nil {
				returnType = frontMatter["ts_return"].(string)
			} else if frontMatter["return"] != nil {
				returnType = frontMatter["return"].(string)
			}

			if len(params) == 0 {
				if frontMatter["ts_is_function"] != true {
					params = nil
				}
			}

			funcs = append(funcs, Function{
				Name:        frontMatter["name"].(string),
				Params:      params,
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
	return content, nil
}

// Function is a struct the stores information about functions.
type Function struct {
	Name        string
	Params      []Param
	ObjectParam bool
	Deprecated  bool
	ReturnType  string
	Description string
}

// Param is a struct that stores a parameter.
type Param struct {
	Name string
	Type string
}

var caser = cases.Title(language.AmericanEnglish)

func (f Function) formatParams() string {
	var params []string
	for _, p := range f.Params {
		name := p.Name
		if strings.HasSuffix(name, "...") {
			name = "..." + name[:len(name)-3]
		}
		if strings.Contains(name, " ") {
			name = strings.ReplaceAll(caser.String(name), " ", "")
			name = strings.ToLower(name[:1]) + name[1:]
		}

		typeName := p.Type
		if strings.HasSuffix(typeName, "?") {
			typeName = typeName[:len(typeName)-1]
			name += "?"
		}

		params = append(params, fmt.Sprintf("%s: %s", name, typeName))
	}
	if f.ObjectParam {
		return "opts: { " + strings.Join(params, "; ") + " }"
	}
	return strings.Join(params, ", ")
}

func (f Function) docs() string {
	content := f.Description
	if f.Deprecated {
		content += "\n\n@deprecated"
	}
	content += "\n\n@see https://dnscontrol.org/js#" + f.Name
	return "/**\n * " + strings.ReplaceAll(content, "\n", "\n * ") + "\n */"
}

func (f Function) formatMain() string {
	if f.Params == nil {
		return fmt.Sprintf("declare const %s: %s", f.Name, f.ReturnType)
	}
	return fmt.Sprintf("declare function %s(%s): %s", f.Name, f.formatParams(), f.ReturnType)
}

func (f Function) String() string {
	return fmt.Sprintf("%s\n%s;\n\n", f.docs(), f.formatMain())
}
