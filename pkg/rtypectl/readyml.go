package rtypectl

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Fields struct {
	Name       string `yaml:"name"`
	Type       string `yaml:"type"`
	LegacyName string `yaml:"legacy_name,omitempty"`
}

type RTypeConfig struct {
	Name      string   `yaml:"name"`
	RawFields []string `yaml:"fields"`
	Init      string   `yaml:"init,omitempty"`
	Fields    []Fields `yaml:"parsed_fields,omitempty"`
}

type RTypeInfo struct {
	Types []*RTypeConfig
}

func New(filename string) (*RTypeInfo, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config RTypeInfo
	err = yaml.Unmarshal(data, &(config.Types))
	if err != nil {
		return nil, err
	}

	// Parse RawFields into Fields.
	for _, rtype := range config.Types {
		rtype.Fields = make([]Fields, len(rtype.RawFields))
		for i, rawField := range rtype.RawFields {
			parts := strings.Split(rawField, ":")
			if len(parts) < 2 || len(parts) > 3 {
				return nil, fmt.Errorf("invalid field format: %s", rawField)
			}
			rtype.Fields[i] = Fields{
				Name: parts[0],
				Type: parts[1],
			}
			if len(parts) == 3 {
				rtype.Fields[i].LegacyName = parts[2]
			}
		}
	}

	return &config, nil
}
