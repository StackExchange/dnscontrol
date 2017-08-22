package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/StackExchange/dnscontrol/pkg/spflib"
	"github.com/gopherjs/jquery"
)

type gResolver struct{}

type gResp struct {
	Status int
	Answer []struct {
		Data string `json:"data"`
	}
}

func (g gResolver) GetTxt(fqdn string) ([]string, error) {
	resp, err := http.Get("https://dns.google.com/resolve?type=txt&name=" + fqdn)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	dat := &gResp{}
	if err = dec.Decode(dat); err != nil {
		return nil, err
	}
	list := []string{}
	for _, a := range dat.Answer {
		list = append(list, strings.Trim(a.Data, "\""))
	}
	return list, nil
}

var jq = jquery.NewJQuery
var parsed *spflib.SPFRecord
var domain string

func main() {
	jq(func() {
		print("Your current jQuery version is: " + jq().Jquery)
		jq("#lookup_btn").On(jquery.CLICK, func(e jquery.Event) {
			go func() {
				domain = jq("#domain").Val()
				rec, err := spflib.Lookup(domain, gResolver{})
				if err != nil {
					panic(err)
				}
				parsed, err = spflib.Parse(rec, gResolver{})
				if err != nil {
					// todo: show a better error
					panic(err)
				}
				jq("#results").SetHtml(buildHTML(parsed, domain))
				renderResults()
			}()
		})
	})
}

func renderResults() {
	flat := parsed.Flatten("*")
	content := fmt.Sprintf(`
<h3> Fully flattened (length %d)</h3><code>%s</code>	
`, len(flat.TXT()), flat.TXT())
	split := flat.TXTSplit("_netblocks%d." + domain)
	if len(split) > 0 {
		content += fmt.Sprintf("<h3>Fully flattened split (%d lookups)</h3>", len(split)-1)
		for k, v := range split {
			content += fmt.Sprintf("<h4>%s</h4><code>%s</code>", k, v)
		}
	}
	jq("#flattened").SetHtml(content)
}

func buildHTML(rec *spflib.SPFRecord, domain string) string {
	h := "<h1>" + domain + "</h1>"
	h += fmt.Sprintf("<h2>%d lookups</h2>", rec.Lookups())
	return h + genRoot(rec)
}

// html based on https://codepen.io/khoama/pen/hpljA
func genRoot(rec *spflib.SPFRecord) string {
	h := fmt.Sprintf(` 
<ul>
	<li class='root'>%s</li>
	`, rec.TXT())
	for _, p := range rec.Parts {
		h += genPart(p)
	}
	h += "</ul>"
	return h
}

func genPart(rec *spflib.SPFPart) string {
	if !rec.IsLookup {
		return fmt.Sprintf(`<li>%s</li>`, rec.Text)
	}
	h := fmt.Sprintf(`<li>
	<input type="checkbox" id="%s" name="%s" />
	<label for="%s">%s(%d lookups)</label>`, rec.IncludeDomain, rec.IncludeDomain, rec.IncludeDomain, rec.Text, rec.IncludeRecord.Lookups()+1)
	h += fmt.Sprintf("<ul>")
	for _, p := range rec.IncludeRecord.Parts {
		h += genPart(p)
	}
	h += "</ul>"
	h += "</li>"
	return h
}
