// Copyright (c) 2018 Kai Schwarz (HEXONET GmbH). All rights reserved.
//
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE.md file.

// Package responseparser provides functionality to cover API response
// data parsing and serializing.
package responseparser

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Parse method to return plain API response parsed into hash format
func Parse(r string) map[string]interface{} {
	hash := make(map[string]interface{})
	tmp := strings.Split(strings.Replace(r, "\r", "", -1), "\n")
	p1 := regexp.MustCompile("^([^\\=]*[^\\t\\= ])[\\t ]*=[\\t ]*(.*)$")
	p2 := regexp.MustCompile("(?i)^property\\[([^\\]]*)\\]\\[([0-9]+)\\]")
	properties := make(map[string][]string)
	for _, row := range tmp {
		m := p1.MatchString(row)
		if m {
			groups := p1.FindStringSubmatch(row)
			property := strings.ToUpper(groups[1])
			mm := p2.MatchString(property)
			if mm {
				groups2 := p2.FindStringSubmatch(property)
				key := strings.Replace(strings.ToUpper(groups2[1]), "\\s", "", -1)
				// idx2 := strconv.Atoi(groups2[2])
				list := make([]string, len(properties[key]))
				copy(list, properties[key])
				pat := regexp.MustCompile("[\\t ]*$")
				rep1 := "${1}$2"
				list = append(list, pat.ReplaceAllString(groups[2], rep1))
				properties[key] = list
			} else {
				val := groups[2]
				if len(val) > 0 {
					pat := regexp.MustCompile("[\\t ]*$")
					hash[property] = pat.ReplaceAllString(val, "")
				}
			}
		}
	}
	if len(properties) > 0 {
		hash["PROPERTY"] = properties
	}
	if _, ok := hash["DESCRIPTION"]; !ok {
		hash["DESCRIPTION"] = ""
	}
	return hash
}

// Serialize method to serialize API response hash format back to string
func Serialize(hash map[string]interface{}) string {
	var plain strings.Builder
	plain.WriteString("[RESPONSE]")
	keys := []string{}
	for k := range hash {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if strings.Compare(k, "PROPERTY") == 0 {
			p := hash[k].(map[string][]string)
			keys2 := []string{}
			for k2 := range p {
				keys2 = append(keys2, k2)
			}
			sort.Strings(keys2)
			for _, k2 := range keys2 {
				v2 := p[k2]
				for i, v3 := range v2 {
					plain.WriteString("\r\nPROPERTY[")
					plain.WriteString(k2)
					plain.WriteString("][")
					plain.WriteString(fmt.Sprintf("%d", i))
					plain.WriteString("]=")
					plain.WriteString(v3)
				}
			}
		} else {
			tmp := hash[k].(string)
			if len(tmp) > 0 {
				plain.WriteString("\r\n")
				plain.WriteString(k)
				plain.WriteString("=")
				plain.WriteString(tmp)
			}
		}
	}
	plain.WriteString("\r\nEOF\r\n")
	return plain.String()
}
