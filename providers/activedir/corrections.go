package activedir

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
)

//
// import (
// 	"encoding/json"
// 	"fmt"
// 	"os"
// 	"strings"
// 	"time"
//
// 	"github.com/StackExchange/dnscontrol/v3/models"
// 	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
// 	"github.com/StackExchange/dnscontrol/v3/pkg/printer"
// 	"github.com/TomOnTime/utfutil"
// )
//
// const zoneDumpFilenamePrefix = "adzonedump"
//
// // RecordConfigJSON RecordConfig, reconfigured for JSON input/output.
// type RecordConfigJSON struct {
// 	Name string `json:"hostname"`
// 	Type string `json:"recordtype"`
// 	Data string `json:"recorddata"`
// 	TTL  uint32 `json:"timetolive"`
// }

// // list of types this provider supports.
// // until it is up to speed with all the built-in types.
// var supportedTypes = map[string]bool{
// 	"A":     true,
// 	"AAAA":  true,
// 	"CNAME": true,
// 	"NS":    true,
// }
//
// // GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
// func (c *activedirProvider) GetZoneRecords(domain string) (models.Records, error) {
// 	foundRecords, err := c.getExistingRecords(domain)
// 	if err != nil {
// 		return nil, fmt.Errorf("c.getExistingRecords(%q) failed: %v", domain, err)
// 	}
// 	return foundRecords, nil
// }

// GetDomainCorrections gets existing records, diffs them against existing, and returns corrections.
func (c *activedirProvider) GenerateDomainCorrections(dc *models.DomainConfig, existing models.Records) ([]*models.Correction, error) {

	// 	dc.Filter(func(r *models.RecordConfig) bool {
	// 		if r.Type == "NS" && r.Name == "@" {
	// 			return false
	// 		}
	// 		if !supportedTypes[r.Type] {
	// 			printer.Warnf("Active Directory only manages certain record types. Won't consider %s %s\n", r.Type, r.GetLabelFQDN())
	// 			return false
	// 		}
	// 		return true
	// 	})

	// Read foundRecords:
	foundRecords, err := c.GetZoneRecords(dc.Name)
	if err != nil {
		return nil, fmt.Errorf("c.GetDNSZoneRecords(%v) failed: %v", dc.Name, err)
	}

	// Normalize
	models.PostProcessRecords(foundRecords)

	differ := diff.New(dc)
	_, creates, dels, modifications, err := differ.IncrementalDiff(foundRecords)
	if err != nil {
		return nil, err
	}

	// Generate changes.
	corrections := []*models.Correction{}
	for _, del := range dels {
		corrections = append(corrections, c.deleteRec(dc.Name, del))
	}
	for _, cre := range creates {
		corrections = append(corrections, c.createRec(dc.Name, cre)...)
	}
	for _, m := range modifications {
		corrections = append(corrections, c.modifyRec(dc.Name, m))
	}
	return corrections, nil

}

func (c *activedirProvider) deleteRec(domainname string, cor diff.Correlation) *models.Correction {
	rec := cor.Existing
	return &models.Correction{
		Msg: cor.String(),
		F: func() error {
			return c.shell.RecordDelete(domainname, rec)
		},
	}
}

func (c *activedirProvider) createRec(domainname string, cre diff.Correlation) []*models.Correction {
	rec := cre.Desired
	arr := []*models.Correction{
		{
			Msg: cre.String(),
			F: func() error {
				return c.shell.RecordCreate(domainname, rec)
			}},
	}
	return arr
}

func (c *activedirProvider) modifyRec(domainname string, m diff.Correlation) *models.Correction {
	old, rec := m.Existing, m.Desired
	return &models.Correction{
		Msg: m.String(),
		F: func() error {
			return c.shell.RecordModify(domainname, old, rec)
		},
	}
}
