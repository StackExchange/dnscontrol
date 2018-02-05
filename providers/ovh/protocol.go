package ovh

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/miekg/dns/dnsutil"
	"github.com/pkg/errors"
)

// Void an empty structure.
type Void struct {
}

// fetchDomainList gets list of zones for account
func (c *ovhProvider) fetchZones() error {
	if c.zones != nil {
		return nil
	}
	c.zones = map[string]bool{}

	var response []string

	err := c.client.Call("GET", "/domain/zone", nil, &response)

	if err != nil {
		return err
	}

	for _, d := range response {
		c.zones[d] = true
	}
	return nil
}

// Zone describes the attributes of a DNS zone.
type Zone struct {
	LastUpdate      string   `json:"lastUpdate,omitempty"`
	HasDNSAnycast   bool     `json:"hasDNSAnycast,omitempty"`
	NameServers     []string `json:"nameServers"`
	DNSSecSupported bool     `json:"dnssecSupported"`
}

// get info about a zone.
func (c *ovhProvider) fetchZone(fqdn string) (*Zone, error) {
	var response Zone

	err := c.client.Call("GET", "/domain/zone/"+fqdn, nil, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// Record describes a DNS record.
type Record struct {
	Target    string `json:"target,omitempty"`
	Zone      string `json:"zone,omitempty"`
	TTL       uint32 `json:"ttl,omitempty"`
	FieldType string `json:"fieldType,omitempty"`
	ID        int64  `json:"id,omitempty"`
	SubDomain string `json:"subDomain,omitempty"`
}

type records struct {
	recordsID []int
}

func (c *ovhProvider) fetchRecords(fqdn string) ([]*Record, error) {
	var recordIds []int

	err := c.client.Call("GET", "/domain/zone/"+fqdn+"/record", nil, &recordIds)
	if err != nil {
		return nil, err
	}

	records := make([]*Record, len(recordIds))
	for i, id := range recordIds {
		r, err := c.fecthRecord(fqdn, id)
		if err != nil {
			return nil, err
		}
		records[i] = r
	}

	return records, nil
}

func (c *ovhProvider) fecthRecord(fqdn string, id int) (*Record, error) {
	var response Record

	err := c.client.Call("GET", fmt.Sprintf("/domain/zone/%s/record/%d", fqdn, id), nil, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Returns a function that can be invoked to delete a record in a zone.
func (c *ovhProvider) deleteRecordFunc(id int64, fqdn string) func() error {
	return func() error {
		err := c.client.Call("DELETE", fmt.Sprintf("/domain/zone/%s/record/%d", fqdn, id), nil, nil)
		if err != nil {
			return err
		}
		return nil
	}
}

// Returns a function that can be invoked to create a record in a zone.
func (c *ovhProvider) createRecordFunc(rc *models.RecordConfig, fqdn string) func() error {
	return func() error {
		record := Record{
			SubDomain: dnsutil.TrimDomainName(rc.NameFQDN, fqdn),
			FieldType: rc.Type,
			Target:    rc.GetTargetCombined(),
			TTL:       rc.TTL,
		}
		if record.SubDomain == "@" {
			record.SubDomain = ""
		}
		var response Record
		err := c.client.Call("POST", fmt.Sprintf("/domain/zone/%s/record", fqdn), &record, &response)
		return err
	}
}

// Returns a function that can be invoked to update a record in a zone.
func (c *ovhProvider) updateRecordFunc(old *Record, rc *models.RecordConfig, fqdn string) func() error {
	return func() error {
		record := Record{
			SubDomain: dnsutil.TrimDomainName(rc.NameFQDN, fqdn),
			FieldType: rc.Type,
			Target:    rc.GetTargetCombined(),
			TTL:       rc.TTL,
			Zone:      fqdn,
			ID:        old.ID,
		}
		if record.SubDomain == "@" {
			record.SubDomain = ""
		}

		return c.client.Call("PUT", fmt.Sprintf("/domain/zone/%s/record/%d", fqdn, old.ID), &record, &Void{})
	}
}

func (c *ovhProvider) refreshZone(fqdn string) error {
	return c.client.Call("POST", fmt.Sprintf("/domain/zone/%s/refresh", fqdn), nil, &Void{})
}

// fetch the NS OVH attributed to this zone (which is distinct from fetchRealNS which
// get the exact NS stored at the registrar
func (c *ovhProvider) fetchNS(fqdn string) ([]string, error) {
	zone, err := c.fetchZone(fqdn)
	if err != nil {
		return nil, err
	}

	return zone.NameServers, nil
}

// CurrentNameServer stores information about nameservers.
type CurrentNameServer struct {
	ToDelete bool   `json:"toDelete,omitempty"`
	IP       string `json:"ip,omitempty"`
	IsUsed   bool   `json:"isUsed,omitempty"`
	ID       int    `json:"id,omitempty"`
	Host     string `json:"host,omitempty"`
}

// Retrieve the NS currently being deployed to the registrar
func (c *ovhProvider) fetchRegistrarNS(fqdn string) ([]string, error) {
	var nameServersID []int
	err := c.client.Call("GET", "/domain/"+fqdn+"/nameServer", nil, &nameServersID)
	if err != nil {
		return nil, err
	}

	var nameServers []string
	for _, id := range nameServersID {
		var ns CurrentNameServer
		err = c.client.Call("GET", fmt.Sprintf("/domain/%s/nameServer/%d", fqdn, id), nil, &ns)
		if err != nil {
			return nil, err
		}

		// skip NS that we asked for deletion
		if ns.ToDelete {
			continue
		}
		nameServers = append(nameServers, ns.Host)
	}

	return nameServers, nil
}

// DomainNS describes a domain's NS in ovh's protocol.
type DomainNS struct {
	Host string `json:"host,omitempty"`
	IP   string `json:"ip,omitempty"`
}

// UpdateNS describes a list of nameservers in ovh's protocol.
type UpdateNS struct {
	NameServers []DomainNS `json:"nameServers"`
}

// Task describes a task in ovh's protocol.
type Task struct {
	Function      string `json:"function,omitempty"`
	Status        string `json:"status,omitempty"`
	CanAccelerate bool   `json:"canAccelerate,omitempty"`
	LastUpdate    string `json:"lastUpdate,omitempty"`
	CreationDate  string `json:"creationDate,omitempty"`
	Comment       string `json:"comment,omitempty"`
	TodoDate      string `json:"todoDate,omitempty"`
	ID            int64  `json:"id,omitempty"`
	CanCancel     bool   `json:"canCancel,omitempty"`
	DoneDate      string `json:"doneDate,omitempty"`
	CanRelaunch   bool   `json:"canRelaunch,omitempty"`
}

// Domain describes a domain in ovh's protocol.
type Domain struct {
	NameServerType     string `json:"nameServerType,omitempty"`
	TransferLockStatus string `json:"transferLockStatus,omitempty"`
}

func (c *ovhProvider) updateNS(fqdn string, ns []string) error {
	// we first need to make sure we can edit the NS
	// by default zones are in "hosted" mode meaning they default
	// to OVH default NS. In this mode, the NS can't be updated.
	domain := Domain{NameServerType: "external"}
	err := c.client.Call("PUT", fmt.Sprintf("/domain/%s", fqdn), &domain, &Void{})
	if err != nil {
		return err
	}

	var newNs []DomainNS
	for _, n := range ns {
		newNs = append(newNs, DomainNS{
			Host: n,
		})
	}

	update := UpdateNS{
		NameServers: newNs,
	}
	var task Task
	err = c.client.Call("POST", fmt.Sprintf("/domain/%s/nameServers/update", fqdn), &update, &task)
	if err != nil {
		return err
	}

	if task.Status == "error" {
		return errors.Errorf("API error while updating ns for %s: %s", fqdn, task.Comment)
	}

	// we don't wait for the task execution. One of the reason is that
	// NS modification can take time in the registrar, the other is that every task
	// in OVH is usually executed a few minutes after they have been registered.
	// We count on the fact that `GetNameservers` uses the registrar API to get
	// a coherent view (including pending modifications) of the registered NS.

	return nil
}
