// Copyright (c) 2018 Kai Schwarz (1API GmbH). All rights reserved.
//
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE.md file.

package hashresponse

import (
	"strings"
)

// Templates class manages default api response templates to be used for different reasons.
// It also provides functionality to compare a response against a template.
//
// Basically used to provide custom response templates that are used in error cases to have a useful way to responds to the client.
type Templates struct {
	// represents the template container
	templates map[string]string
}

// NewTemplates represents the constructor for struct Templates.
func NewTemplates() *Templates {
	tpls := make(map[string]string)
	tpls["empty"] = "[RESPONSE]\r\ncode=423\r\ndescription=Empty API response\r\nEOF\r\n"
	tpls["error"] = "[RESPONSE]\r\ncode=421\r\ndescription=Command failed due to server error. Client should try again\r\nEOF\r\n"
	tpls["expired"] = "[RESPONSE]\r\ncode=530\r\ndescription=SESSION NOT FOUND\r\nEOF\r\n"
	tpls["commonerror"] = "[RESPONSE]\r\nDESCRIPTION=Command failed;####ERRMSG####;\r\nCODE=500\r\nQUEUETIME=0\r\nRUNTIME=0\r\nEOF"
	return &Templates{
		templates: tpls,
	}
}

// GetAll method to get all available response templates
func (dr *Templates) GetAll() map[string]string {
	return dr.templates
}

// GetParsed method to get a parsed response template by given template id.
func (dr *Templates) GetParsed(templateid string) map[string]interface{} {
	hr := NewHashResponse(dr.Get(templateid))
	return hr.GetHash()
}

// Get method to get a raw response template by given template id.
func (dr *Templates) Get(templateid string) string {
	return dr.templates[templateid]
}

// Set method to set a response template by given template id and content
func (dr *Templates) Set(templateid string, templatecontent string) {
	dr.templates[templateid] = templatecontent
}

// SetParsed method to set a response template by given template id and parsed content
func (dr *Templates) SetParsed(templateid string, templatecontent map[string]interface{}) {
	hr := NewHashResponse("")
	dr.templates[templateid] = hr.Serialize(templatecontent)
}

// Match method to compare a given raw api response with a response template identfied by id.
// It compares CODE and DESCRIPTION.
func (dr *Templates) Match(r string, templateid string) bool {
	tpl := NewHashResponse(dr.Get(templateid))
	rr := NewHashResponse(r)
	return (tpl.Code() == rr.Code() && strings.Compare(tpl.Description(), rr.Description()) == 0)
}

// MatchParsed method to compare a given parsed api response with a response template identified by id.
// It compares CODE and DESCRIPTION.
func (dr *Templates) MatchParsed(r map[string]interface{}, templateid string) bool {
	tpl := dr.GetParsed(templateid)
	return (strings.Compare(tpl["CODE"].(string), r["CODE"].(string)) == 0 &&
		strings.Compare(tpl["DESCRIPTION"].(string), r["DESCRIPTION"].(string)) == 0)
}
