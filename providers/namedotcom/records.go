package namedotcom

import (
	"fmt"
	"log"
	"strings"

	"github.com/miekg/dns/dnsutil"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers/diff"
)

var defaultNameservers = []*models.Nameserver{
	{Name: "ns1.name.com"},
	{Name: "ns2.name.com"},
	{Name: "ns3.name.com"},
	{Name: "ns4.name.com"},
}

func (n *nameDotCom) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	records, err := n.getRecords(dc.Name)
	if err != nil {
		return nil, err
	}
	actual := make([]diff.Record, len(records))
	for i := range records {
		actual[i] = records[i]
	}

	desired := make([]diff.Record, 0, len(dc.Records))
	for _, rec := range dc.Records {
		if rec.TTL == 0 {
			rec.TTL = 300
		}
		if rec.Type == "NS" {
			// name.com does not really let you manage NS records via api. They explicitly state that you cannot change the base domain NS records,
			// but the api also will not return you NS records for subdomains either. Maybe a bug.
			// dnscontrol will print warnings if you try to manage NS records
			if rec.NameFQDN == dc.Name {
				if !strings.HasSuffix(rec.Target, ".name.com.") {
					log.Printf("Warning: name.com does not allow NS records on base domain to be modified. %s will not be added.", rec.Target)
				}
			} else {
				log.Printf("Warning: name.com does not allow NS records to be modified via api. NS for %s will not be managed.", rec.Name)
			}
			continue
		}
		desired = append(desired, rec)
	}

	_, create, del, mod := diff.IncrementalDiff(actual, desired)
	corrections := []*models.Correction{}

	for _, d := range del {
		rec := d.Existing.(*nameComRecord)
		c := &models.Correction{Msg: d.String(), F: func() error { return n.deleteRecord(rec.RecordID, dc.Name) }}
		corrections = append(corrections, c)
	}
	for _, cre := range create {
		rec := cre.Desired.(*models.RecordConfig)
		c := &models.Correction{Msg: cre.String(), F: func() error { return n.createRecord(rec, dc.Name) }}
		corrections = append(corrections, c)
	}
	for _, chng := range mod {
		old := chng.Existing.(*nameComRecord)
		new := chng.Desired.(*models.RecordConfig)
		c := &models.Correction{Msg: chng.String(), F: func() error {
			err := n.deleteRecord(old.RecordID, dc.Name)
			if err != nil {
				return err
			}
			return n.createRecord(new, dc.Name)
		}}
		corrections = append(corrections, c)
	}
	return corrections, nil
}

func apiGetRecords(domain string) string {
	return fmt.Sprintf("%s/dns/list/%s", apiBase, domain)
}
func apiCreateRecord(domain string) string {
	return fmt.Sprintf("%s/dns/create/%s", apiBase, domain)
}
func apiDeleteRecord(domain string) string {
	return fmt.Sprintf("%s/dns/delete/%s", apiBase, domain)
}

type nameComRecord struct {
	RecordID string `json:"record_id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Content  string `json:"content"`
	TTL      string `json:"ttl"`
	Priority string `json:"priority"`
}

func (r *nameComRecord) GetName() string {
	return r.Name
}
func (r *nameComRecord) GetType() string {
	return r.Type
}
func (r *nameComRecord) GetContent() string {
	return r.Content
}
func (r *nameComRecord) GetComparisionData() string {
	mxPrio := ""
	if r.Type == "MX" {
		mxPrio = fmt.Sprintf(" %s ", r.Priority)
	}
	return fmt.Sprintf("%s%s", r.TTL, mxPrio)
}

type listRecordsResponse struct {
	*apiResult
	Records []*nameComRecord `json:"records"`
}

func (n *nameDotCom) getRecords(domain string) ([]*nameComRecord, error) {
	result := &listRecordsResponse{}
	err := n.get(apiGetRecords(domain), result)
	if err != nil {
		return nil, err
	}
	if err = result.getErr(); err != nil {
		return nil, err
	}

	for _, rc := range result.Records {
		if rc.Type == "CNAME" || rc.Type == "MX" || rc.Type == "NS" {
			rc.Content = rc.Content + "."
		}
	}
	return result.Records, nil
}

func (n *nameDotCom) createRecord(rc *models.RecordConfig, domain string) error {
	target := rc.Target
	if rc.Type == "CNAME" || rc.Type == "MX" || rc.Type == "NS" {
		if target[len(target)-1] == '.' {
			target = target[:len(target)-1]
		} else {
			return fmt.Errorf("Unexpected. CNAME/MX/NS target did not end with dot.\n")
		}
	}
	dat := struct {
		Hostname string `json:"hostname"`
		Type     string `json:"type"`
		Content  string `json:"content"`
		TTL      uint32 `json:"ttl,omitempty"`
		Priority uint16 `json:"priority,omitempty"`
	}{
		Hostname: dnsutil.TrimDomainName(rc.NameFQDN, domain),
		Type:     rc.Type,
		Content:  target,
		TTL:      rc.TTL,
		Priority: rc.Priority,
	}
	if dat.Hostname == "@" {
		dat.Hostname = ""
	}
	resp, err := n.post(apiCreateRecord(domain), dat)
	if err != nil {
		return err
	}
	return resp.getErr()
}

func (n *nameDotCom) deleteRecord(id, domain string) error {
	dat := struct {
		ID string `json:"record_id"`
	}{id}
	resp, err := n.post(apiDeleteRecord(domain), dat)
	if err != nil {
		return err
	}
	return resp.getErr()
}
