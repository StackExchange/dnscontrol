package msdns

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/pkg/txtutil"
)

// GetDomainCorrections gets existing records, diffs them against existing, and returns corrections.
func (c *msdnsProvider) GenerateDomainCorrections(dc *models.DomainConfig, existing models.Records) ([]*models.Correction, error) {

	// Read foundRecords:
	foundRecords, err := c.GetZoneRecords(dc.Name)
	if err != nil {
		return nil, fmt.Errorf("c.GetDNSZoneRecords(%v) failed: %v", dc.Name, err)
	}

	// Normalize
	models.PostProcessRecords(foundRecords)
	txtutil.SplitSingleLongTxt(dc.Records) // Autosplit long TXT records

	differ := diff.New(dc)
	_, creates, dels, modifications, err := differ.IncrementalDiff(foundRecords)
	if err != nil {
		return nil, err
	}

	// Generate changes.
	corrections := []*models.Correction{}
	for _, del := range dels {
		corrections = append(corrections, c.deleteRec(c.dnsserver, dc.Name, del))
	}
	for _, cre := range creates {
		corrections = append(corrections, c.createRec(c.dnsserver, dc.Name, cre)...)
	}
	for _, m := range modifications {
		corrections = append(corrections, c.modifyRec(c.dnsserver, dc.Name, m))
	}
	return corrections, nil

}

func (c *msdnsProvider) deleteRec(dnsserver, domainname string, cor diff.Correlation) *models.Correction {
	rec := cor.Existing
	return &models.Correction{
		Msg: cor.String(),
		F: func() error {
			return c.shell.RecordDelete(dnsserver, domainname, rec)
		},
	}
}

func (c *msdnsProvider) createRec(dnsserver, domainname string, cre diff.Correlation) []*models.Correction {
	rec := cre.Desired
	arr := []*models.Correction{{
		Msg: cre.String(),
		F: func() error {
			return c.shell.RecordCreate(dnsserver, domainname, rec)
		},
	}}
	return arr
}

func (c *msdnsProvider) modifyRec(dnsserver, domainname string, m diff.Correlation) *models.Correction {
	old, rec := m.Existing, m.Desired
	return &models.Correction{
		Msg: m.String(),
		F: func() error {
			return c.shell.RecordModify(dnsserver, domainname, old, rec)
		},
	}
}
