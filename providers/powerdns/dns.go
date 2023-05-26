package powerdns

import (
	"context"
	"net/http"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/diff2"
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
func (dsp *powerdnsProvider) GetZoneRecords(domain string, meta map[string]string) (models.Records, error) {
	zone, err := dsp.client.Zones().GetZone(context.Background(), dsp.ServerName, domain)
	if err != nil {
		return nil, err
	}

	curRecords := models.Records{}
	// loop over grouped records by type, called RRSet
	for _, rrset := range zone.ResourceRecordSets {
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

// GetZoneRecordsCorrections returns a list of corrections that will turn existing records into dc.Records.
func (dsp *powerdnsProvider) GetZoneRecordsCorrections(dc *models.DomainConfig, existing models.Records) ([]*models.Correction, error) {
	// create record diff by group
	var err error

	var corrections []*models.Correction
	if !diff2.EnableDiff2 {
		corrections, err = dsp.getDiff1DomainCorrections(dc, existing)
	} else {
		corrections, err = dsp.getDiff2DomainCorrections(dc, existing)
	}
	if err != nil {
		return nil, err
	}

	// DNSSec corrections
	dnssecCorrections, err := dsp.getDNSSECCorrections(dc)
	if err != nil {
		return nil, err
	}

	return append(corrections, dnssecCorrections...), nil
}

// EnsureZoneExists creates a zone if it does not exist
func (dsp *powerdnsProvider) EnsureZoneExists(domain string) error {
	if _, err := dsp.client.Zones().GetZone(context.Background(), dsp.ServerName, canonical(domain)); err != nil {
		if e, ok := err.(pdnshttp.ErrUnexpectedStatus); ok {
			if e.StatusCode != http.StatusNotFound {
				return err
			}
		}
	} else { // zone seems to exist
		return nil
	}

	_, err := dsp.client.Zones().CreateZone(context.Background(), dsp.ServerName, zones.Zone{
		Name:        canonical(domain),
		Type:        zones.ZoneTypeZone,
		DNSSec:      dsp.DNSSecOnCreate,
		Nameservers: dsp.DefaultNS,
		Kind:        dsp.ZoneKind,
	})
	return err
}
