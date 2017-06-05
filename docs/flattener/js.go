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

func main() {
	jq(func() {
		print("Your current jQuery version is: " + jq().Jquery)
		jq("#lookup_btn").On(jquery.CLICK, func(e jquery.Event) {
			go func() {
				dom := jq("#domain").Val()
				rec, err := spflib.Lookup(dom, gResolver{})
				if err != nil {
					panic(err)
				}
				parsed, err := spflib.Parse(rec, gResolver{})
				if err != nil {
					panic(err)
				}
				jq("#results").SetHtml(buildHTML(parsed, dom))

				jq("#star").SetText(parsed.Flatten("mailgun.org,spf-basic.fogcreek.com").TXT())
			}()
		})
	})
}

func buildHTML(rec *spflib.SPFRecord, domain string) string {
	h := "<h1>" + domain + "</h1>"
	h += fmt.Sprintf("<h2>%d lookups</h2>", rec.Lookups())
	return h + recHTML(rec)
}

func recHTML(rec *spflib.SPFRecord) string {
	//open panel
	h := fmt.Sprintf(`<div class="panel panel-default">
  		<div class="panel-heading">%s (%d lookups)</div><div class="panel-body"><ul class="list-group">`, rec.TXT(), rec.Lookups())
	for _, p := range rec.Parts {
		class := "list-group-item-success"
		if p.IsLookup {
			class = "list-group-item-danger"
		}
		// open list item
		h += fmt.Sprintf("<li class='list-group-item %s'>%s", class, p.Text)
		if p.IncludeRecord != nil {
			h += recHTML(p.IncludeRecord)
		}
		// close list item
		h += "</li>"
	}
	//close panel
	h += "</ul></div></div>"
	return h
}
