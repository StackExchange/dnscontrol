// Copyright (c) 2018 Kai Schwarz (1API GmbH). All rights reserved.
//
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE.md file.

// Package hashresponse covers all functionality to handle an API response in hash format and provides access to a response template manager
// to cover http error cases etc. with API response format.
package hashresponse

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// HashResponse class provides basic functionality to work with API responses.
type HashResponse struct {
	// represents the parsed API response data
	hash map[string]interface{}
	// represents the raw API response data
	raw string
	// represents the pattern to match columns used for pagination
	pagerRegexp regexp.Regexp
	// represents the column filter pattern
	columnFilterRegexp regexp.Regexp
	// represents an flag to turn column filter on/off
	columnFilterActive bool
}

// NewHashResponse represents the constructor for struct HashResponse.
// Provide the raw api response string as parameter.
func NewHashResponse(r string) *HashResponse {
	res := r
	if len(res) == 0 {
		res = NewTemplates().Get("empty")
	}
	hr := &HashResponse{
		raw:                res,
		columnFilterActive: false,
		pagerRegexp:        *regexp.MustCompile("^(TOTAL|FIRST|LAST|LIMIT|COUNT)$"),
	}
	hr.hash = hr.Parse(hr.raw)
	return hr
}

// GetRaw method to return the api raw (but filtered - in case of useColRegexp) response data
func (hr *HashResponse) GetRaw() string {
	return hr.GetRawByFilter(false)
}

// GetRawByFilter method to return the api raw response data.
// Use noColumnFilter parameter to explicitly suppress a current active column filter.
func (hr *HashResponse) GetRawByFilter(noColumnFilter bool) string {
	if noColumnFilter || !hr.columnFilterActive {
		return hr.raw
	}
	return hr.Serialize(hr.GetHash())
}

// GetHash method to return the parsed api response
func (hr *HashResponse) GetHash() map[string]interface{} {
	if hr.columnFilterActive {
		var h = make(map[string]interface{})
		for k, v := range hr.hash {
			h[k] = v
		}
		properties := hr.hash["PROPERTY"]
		if properties != nil {
			d := make(map[string][]string)
			for k, v := range properties.(map[string][]string) {
				if hr.columnFilterRegexp.MatchString(k) {
					d[k] = v
				}
			}
			h["PROPERTY"] = d
		}
		return h
	}
	return hr.hash
}

// DisableColumnFilter method to turn of column filter
func (hr *HashResponse) DisableColumnFilter() {
	hr.columnFilterActive = false
	// hr.columnFilterRegexp = nil
}

// EnableColumnFilter method to set a column filter
func (hr *HashResponse) EnableColumnFilter(pattern string) {
	hr.columnFilterActive = true
	hr.columnFilterRegexp = *regexp.MustCompile(pattern)
}

// Code method to access the api response code
func (hr *HashResponse) Code() int {
	var x int
	fmt.Sscanf(hr.hash["CODE"].(string), "%d", &x)
	return x
}

// Description method to access the api response description
func (hr *HashResponse) Description() string {
	return hr.hash["DESCRIPTION"].(string)
}

// Runtime method to access the api response runtime
func (hr *HashResponse) Runtime() float64 {
	s, _ := strconv.ParseFloat(hr.hash["RUNTIME"].(string), 64)
	return s
}

// Queuetime method to access the api response queuetime
func (hr *HashResponse) Queuetime() float64 {
	s, _ := strconv.ParseFloat(hr.hash["QUEUETIME"].(string), 64)
	return s
}

// First method to access the pagination data "first".
// Represents the row index of 1st row of the current response of the whole result set
func (hr *HashResponse) First() int {
	val, _ := hr.GetColumnIndex("FIRST", 0)
	if len(val) == 0 {
		return 0
	}
	var x int
	fmt.Sscanf(val, "%d", &x)
	return x
}

// Count method to access the pagination data "count"
// Represents the count of rows returned in the current response
func (hr *HashResponse) Count() int {
	val, _ := hr.GetColumnIndex("COUNT", 0)
	if len(val) != 0 {
		var x int
		fmt.Sscanf(val, "%d", &x)
		return x
	}
	c := 0
	max := 0
	cols := hr.GetColumnKeys()
	for _, el := range cols {
		col := hr.GetColumn(el)
		c = len(col)
		if c > max {
			max = c
		}
	}
	return c
}

// Last method to access the pagination data "last"
// Represents the row index of last row of the current response of the whole result set
func (hr *HashResponse) Last() int {
	val, _ := hr.GetColumnIndex("LAST", 0)
	if len(val) == 0 {
		return hr.Count() - 1
	}
	var x int
	fmt.Sscanf(val, "%d", &x)
	return x
}

// Limit method to access the pagination data "limit"
// represents the limited amount of rows requested to be returned
func (hr *HashResponse) Limit() int {
	val, _ := hr.GetColumnIndex("LIMIT", 0)
	if len(val) == 0 {
		return hr.Count()
	}
	var x int
	fmt.Sscanf(val, "%d", &x)
	return x
}

// Total method to access the pagination data "total"
// represents the total amount of rows available in the whole result set
func (hr *HashResponse) Total() int {
	val, _ := hr.GetColumnIndex("TOTAL", 0)
	if len(val) == 0 {
		return hr.Count()
	}
	var x int
	fmt.Sscanf(val, "%d", &x)
	return x
}

// Pages method to return the amount of pages of the current result set
func (hr *HashResponse) Pages() int {
	t := hr.Total()
	if t > 0 {
		return int(math.Ceil(float64(t) / float64(hr.Limit())))
	}
	return 1
}

// Page method to return the number of the current page
func (hr *HashResponse) Page() int {
	if hr.Count() > 0 {
		// limit cannot be 0 as this.count() will cover this, no worries
		d := float64(hr.First()) / float64(hr.Limit())
		return int(math.Floor(d)) + 1
	}
	return 1
}

// Prevpage method to get the previous page number
func (hr *HashResponse) Prevpage() int {
	p := hr.Page() - 1
	if p > 0 {
		return p
	}
	return 1
}

// Nextpage method to get the next page number
func (hr *HashResponse) Nextpage() int {
	p := hr.Page() + 1
	pages := hr.Pages()
	if p <= pages {
		return p
	}
	return pages
}

// GetPagination method to return all pagination data at once
func (hr *HashResponse) GetPagination() map[string]int {
	pagination := make(map[string]int)
	pagination["FIRST"] = hr.First()
	pagination["LAST"] = hr.Last()
	pagination["COUNT"] = hr.Count()
	pagination["TOTAL"] = hr.Total()
	pagination["LIMIT"] = hr.Limit()
	pagination["PAGES"] = hr.Pages()
	pagination["PAGE"] = hr.Page()
	pagination["PAGENEXT"] = hr.Nextpage()
	pagination["PAGEPREV"] = hr.Prevpage()
	return pagination
}

// IsSuccess method to check if the api response represents a success case
func (hr *HashResponse) IsSuccess() bool {
	code := hr.Code()
	return (code >= 200 && code < 300)
}

// IsTmpError method to check if the api response represents a temporary error case
func (hr *HashResponse) IsTmpError() bool {
	code := hr.Code()
	return (code >= 400 && code < 500)
}

// IsError method to check if the api response represents an error case
func (hr *HashResponse) IsError() bool {
	code := hr.Code()
	return (code >= 500 && code <= 600)
}

// GetColumnKeys method to get a full list available columns in api response
func (hr *HashResponse) GetColumnKeys() []string {
	var columns []string
	if hr.hash == nil {
		return columns
	}
	property := hr.hash["PROPERTY"]
	if property == nil {
		return columns
	}
	for k := range property.(map[string][]string) {
		if !hr.pagerRegexp.MatchString(k) {
			columns = append(columns, k)
		}
	}
	return columns
}

// GetColumn method to get the full column data for the given column id
func (hr *HashResponse) GetColumn(columnid string) []string {
	if hr.hash == nil || hr.hash["PROPERTY"] == nil {
		return nil
	}
	return hr.hash["PROPERTY"].(map[string][]string)[columnid]
}

// GetColumnIndex method to get a response data field by column id and index
func (hr *HashResponse) GetColumnIndex(columnid string, index int) (string, error) {
	if hr.hash == nil || hr.hash["PROPERTY"] == nil {
		return "", errors.New("column not found")
	}
	column := hr.hash["PROPERTY"].(map[string][]string)[columnid]
	if column == nil || len(column) <= index {
		return "", errors.New("index not found")
	}
	return column[index], nil
}

// Serialize method to stringify a parsed api response
func (hr *HashResponse) Serialize(hash map[string]interface{}) string {
	var plain strings.Builder
	plain.WriteString("[RESPONSE]")
	for k := range hash {
		if strings.Compare(k, "PROPERTY") == 0 {
			for k2, v2 := range hash[k].(map[string][]string) {
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

// Parse method to parse the given raw api response
func (hr *HashResponse) Parse(r string) map[string]interface{} {
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
	return hash
}
