package powerdns

import (
	"context"
	"net/http"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/mittwald/go-powerdns/apis/zones"
	"github.com/mittwald/go-powerdns/pdnshttp"
)

// GetNameservers returns the nameservers for a domain.
func (dsp *powerdnsProvider) GetNameservers(string) ([]*models.Nameserver, error) {
	var r []string
	for _, j := range dsp.nameservers {
		r = append(r, j.Name)
	}
	return models.ToNameservers(r)
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (dsp *powerdnsProvider) GetZoneRecords(domain string) (models.Records, error) {
	zone, err := dsp.client.Zones().GetZone(context.Background(), dsp.ServerName, domain)
	if err != nil {
		return nil, err
	}

	curRecords := models.Records{}
	// loop over grouped records by type, called RRSet
	for _, rrset := range zone.ResourceRecordSets {
		if rrset.Type == "SOA" {
			continue
		}
		// loop over single records of this group and create records
		for _, pdnsRecord := range rrset.Records {
			r, err := toRecordConfig(domain, pdnsRecord, rrset.TTL, rrset.Name, rrset.Type)
			if err != nil {
				return nil, err
			}
			curRecords = append(curRecords, r)
		}
	}

	return curRecords, nil
}

// GetDomainCorrections returns a list of corrections to update a domain.
func (dsp *powerdnsProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	var corrections []*models.Correction

	// get current zone records
	curRecords, err := dsp.GetZoneRecords(dc.Name)
	if err != nil {
		return nil, err
	}

	// post-process records
	if err := dc.Punycode(); err != nil {
		return nil, err
	}
	models.PostProcessRecords(curRecords)

	// create record diff by group
	keysToUpdate, err := (diff.New(dc)).ChangedGroups(curRecords)
	if err != nil {
		return nil, err
	}
	desiredRecords := dc.Records.GroupedByKey()

	var cuCorrections []*models.Correction
	var dCorrections []*models.Correction

	// add create/update and delete corrections separately
	for label, msgs := range keysToUpdate {
		labelName := label.NameFQDN + "."
		labelType := label.Type
		msgJoined := strings.Join(msgs, "\n   ")

		if _, ok := desiredRecords[label]; !ok {
			// no record found so delete it
			dCorrections = append(dCorrections, &models.Correction{
				Msg: msgJoined,
				F: func() error {
					return dsp.client.Zones().RemoveRecordSetFromZone(context.Background(), dsp.ServerName, dc.Name, labelName, labelType)
				},
			})
		} else {
			// record found so create or update it
			ttl := desiredRecords[label][0].TTL
			var records []zones.Record
			for _, recordContent := range desiredRecords[label] {
				records = append(records, zones.Record{
					Content: recordContent.GetTargetCombined(),
				})
			}
			cuCorrections = append(cuCorrections, &models.Correction{
				Msg: msgJoined,
				F: func() error {
					return dsp.client.Zones().AddRecordSetToZone(context.Background(), dsp.ServerName, dc.Name, zones.ResourceRecordSet{
						Name:       labelName,
						Type:       labelType,
						TTL:        int(ttl),
						Records:    records,
						ChangeType: zones.ChangeTypeReplace,
					})
				},
			})
		}
	}

	// append corrections in the right order
	// delete corrections must be run first to avoid correlations with existing RR
	corrections = append(corrections, dCorrections...)
	corrections = append(corrections, cuCorrections...)

	// DNSSec corrections
	dnssecCorrections, err := dsp.getDNSSECCorrections(dc)
	if err != nil {
		return nil, err
	}
	corrections = append(corrections, dnssecCorrections...)

	return corrections, nil
}

// EnsureDomainExists adds a domain to the DNS service if it does not exist
func (dsp *powerdnsProvider) EnsureDomainExists(domain string) error {
	if _, err := dsp.client.Zones().GetZone(context.Background(), dsp.ServerName, domain+"."); err != nil {
		if e, ok := err.(pdnshttp.ErrUnexpectedStatus); ok {
			if e.StatusCode != http.StatusNotFound {
				return err
			}
		}
	} else { // domain seems to be there
		return nil
	}

	_, err := dsp.client.Zones().CreateZone(context.Background(), dsp.ServerName, zones.Zone{
		Name:        domain + ".",
		Type:        zones.ZoneTypeZone,
		DNSSec:      dsp.DNSSecOnCreate,
		Nameservers: dsp.DefaultNS,
	})
	return err
}
