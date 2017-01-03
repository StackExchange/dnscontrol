package namedotcom

import (
	"fmt"
	"log"
	"strings"

	"github.com/miekg/dns/dnsutil"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers/diff"
	"strconv"
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
	actual := make([]*models.RecordConfig, len(records))
	for i, r := range records {
		actual[i] = r.toRecord()
	}

	checkNSModifications(dc)

	differ := diff.New(dc)
	_, create, del, mod := differ.IncrementalDiff(actual)
	corrections := []*models.Correction{}

	for _, d := range del {
		rec := d.Existing.Original.(*nameComRecord)
		c := &models.Correction{Msg: d.String(), F: func() error { return n.deleteRecord(rec.RecordID, dc.Name) }}
		corrections = append(corrections, c)
	}
	for _, cre := range create {
		rec := cre.Desired.Original.(*models.RecordConfig)
		c := &models.Correction{Msg: cre.String(), F: func() error { return n.createRecord(rec, dc.Name) }}
		corrections = append(corrections, c)
	}
	for _, chng := range mod {
		old := chng.Existing.Original.(*nameComRecord)
		new := chng.Desired
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

func checkNSModifications(dc *models.DomainConfig) {
	newList := make([]*models.RecordConfig, 0, len(dc.Records))
	for _, rec := range dc.Records {
		if rec.Type == "NS" && rec.NameFQDN == dc.Name {
			// name.com does change base domain NS records. dnscontrol will print warnings if you try to set them to anything besides the name.com defaults.
			if !strings.HasSuffix(rec.Target, ".name.com.") {
				log.Printf("Warning: name.com does not allow NS records on base domain to be modified. %s will not be added.", rec.Target)
			}
			continue
		}
		newList = append(newList, rec)
	}
	dc.Records = newList
}

func (r *nameComRecord) toRecord() *models.RecordConfig {
	ttl, _ := strconv.ParseUint(r.TTL, 10, 32)
	prio, _ := strconv.ParseUint(r.Priority, 10, 16)
	return &models.RecordConfig{
		Name:     r.Name,
		Type:     r.Type,
		Target:   r.Content,
		TTL:      uint32(ttl),
		Priority: uint16(prio),
		Original: r,
	}
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
