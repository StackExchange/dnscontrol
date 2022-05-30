package cscglobal

import (
	"fmt"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/pkg/txtutil"
)

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (client *providerClient) GetZoneRecords(domain string) (models.Records, error) {
	records, err := client.getZoneRecordsAll(domain)
	if err != nil {
		return nil, err
	}

	// Convert them to DNScontrol's native format:

	existingRecords := []*models.RecordConfig{}

	// Option 1: One long list.  If your provider returns one long list,
	// convert each one to RecordType like this:
	// for _, rr := range records {
	// 	existingRecords = append(existingRecords, nativeToRecord(rr, domain))
	//}

	// Option 2: Grouped records. Sometimes the provider returns one item per
	// label. Each item contains a list of all the records at that label.
	// You'll need to split them out into one RecordConfig for each record.  An
	// example of this is the ROUTE53 provider.
	// for _, rg := range records {
	// 	for _, rr := range rg {
	// 		existingRecords = append(existingRecords, nativeToRecords(rg, rr, domain)...)
	// 	}
	// }

	// Option 3: Something else.  In this case, we get a big massive structure
	// which needs to be broken up.  Still, we're returning a list of
	// RecordConfig structures.
	defaultTTL := records.Soa.TTL
	for _, rr := range records.A {
		existingRecords = append(existingRecords, nativeToRecordA(rr, domain, defaultTTL))
	}
	for _, rr := range records.Mx {
		existingRecords = append(existingRecords, nativeToRecordMX(rr, domain, defaultTTL))
	}

	return existingRecords, nil
}

func (client *providerClient) GetNameservers(string) ([]*models.Nameserver, error) {
	// TODO: If using AD for publicly hosted zones, probably pull these from config.
	return nil, nil
}

// GetDomainCorrections get the current and existing records,
// post-process them, and generate corrections.
// NB(tlim): This function should be exactly the same in all DNS providers.  Once
// all providers do this, we can eliminate it and use a Go interface instead.
func (client *providerClient) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	existing, err := client.GetZoneRecords(dc.Name)
	if err != nil {
		return nil, err
	}
	models.PostProcessRecords(existing)
	txtutil.SplitSingleLongTxt(dc.Records) // Autosplit long TXT records

	clean := PrepFoundRecords(existing)
	PrepDesiredRecords(dc)
	return client.GenerateDomainCorrections(dc, clean)
}

// PrepFoundRecords munges any records to make them compatible with
// this provider. Usually this is a no-op.
func PrepFoundRecords(recs models.Records) models.Records {
	// If there are records that need to be modified, removed, etc. we
	// do it here.  Usually this is a no-op.
	return recs
}

// PrepDesiredRecords munges any records to best suit this provider.
func PrepDesiredRecords(dc *models.DomainConfig) {
	// Sort through the dc.Records, eliminate any that can't be
	// supported; modify any that need adjustments to work with the
	// provider.  We try to do minimal changes otherwise it gets
	// confusing.

	dc.Punycode()
}

// GetDomainCorrections gets existing records, diffs them against existing, and returns corrections.
func (client *providerClient) GenerateDomainCorrections(dc *models.DomainConfig, existing models.Records) ([]*models.Correction, error) {

	// Read foundRecords:
	foundRecords, err := client.GetZoneRecords(dc.Name)
	if err != nil {
		return nil, fmt.Errorf("c.GetDNSZoneRecords(%v) failed: %v", dc.Name, err)
	}

	// Normalize
	models.PostProcessRecords(foundRecords)
	//txtutil.SplitSingleLongTxt(dc.Records) // Autosplit long TXT records

	differ := diff.New(dc)
	_, creates, dels, modifications, err := differ.IncrementalDiff(foundRecords)
	if err != nil {
		return nil, err
	}

	// For most providers you'll see something like this:

	// // Generate changes.
	//	corrections := []*models.Correction{}
	//	for _, del := range dels {
	//		corrections = append(corrections, client.deleteRec(client.dnsserver, dc.Name, del))
	//	}
	//	for _, cre := range creates {
	//		corrections = append(corrections, client.createRec(client.dnsserver, dc.Name, cre)...)
	//	}
	//	for _, m := range modifications {
	//		corrections = append(corrections, client.modifyRec(client.dnsserver, dc.Name, m))
	//	}
	//	return corrections, nil

	// However, CSCGlobal has a weird API and therefore we do this:
	var edits []ZoneResourceRecordEdit
	var descriptions []string
	for _, del := range dels {
		edits = append(edits, makePurge(dc.Name, del))
		descriptions = append(descriptions, del.String())
	}
	for _, cre := range creates {
		edits = append(edits, makeAdd(dc.Name, cre))
		descriptions = append(descriptions, cre.String())
	}
	for _, m := range modifications {
		edits = append(edits, makeEdit(dc.Name, m))
		descriptions = append(descriptions, m.String())
	}
	corrections := []*models.Correction{}
	if len(edits) > 0 {
		c := &models.Correction{
			Msg: "\t" + strings.Join(descriptions, "\n\t"),
			F: func() error {
				// CSCGlobal only permits one pending request at a time; pending
				// requests includes failed requests waiting for a human to
				// acknowledge the failure and delete the request.  Therefore, we
				// cancel any pending requests. What a stupid API decision.
				err := client.ClearRequests(dc.Name)
				if err != nil {
					return err
				}
				return client.SendZoneEditRequest(dc.Name, edits)
			},
		}
		corrections = append(corrections, c)
	}
	return corrections, nil
}

func makePurge(domainname string, cor diff.Correlation) ZoneResourceRecordEdit {
	zer := ZoneResourceRecordEdit{
		Action:       "PURGE",
		RecordType:   cor.Existing.Type,
		CurrentKey:   cor.Existing.Name,
		CurrentValue: cor.Existing.GetTargetField(),
	}
	return zer
}

func makeAdd(domainname string, cre diff.Correlation) ZoneResourceRecordEdit {
	rec := cre.Desired
	zer := ZoneResourceRecordEdit{
		Action:     "ADD",
		RecordType: rec.Type,
		NewKey:     rec.Name,
		NewValue:   rec.GetTargetField(),
		NewTTL:     rec.TTL,
	}
	return zer
}

func makeEdit(domainname string, m diff.Correlation) ZoneResourceRecordEdit {
	old, rec := m.Existing, m.Desired
	// TODO: Assert that old.Type == rec.Type
	// TODO: Assert that old.Name == rec.Name
	zer := ZoneResourceRecordEdit{
		Action:       "EDIT",
		RecordType:   old.Type,
		CurrentKey:   old.Name,
		CurrentValue: old.GetTargetField(),
	}
	if old.GetTargetField() != rec.GetTargetField() {
		zer.NewValue = rec.GetTargetField()
	}
	if old.TTL != rec.TTL {
		zer.NewTTL = rec.TTL
	}

	switch old.Type {
	case "A", "CNAME", "NS", "TXT":
		// Nothing to do.

	case "CAA":
		zer.CurrentTag = old.CaaTag
		if old.CaaTag != rec.CaaTag {
			zer.NewTag = rec.CaaTag
		}
		if old.CaaFlag != rec.CaaFlag {
			zer.NewFlag = rec.CaaFlag
		}

	case "MX":
		if old.MxPreference != rec.MxPreference {
			zer.NewPriority = rec.MxPreference
		}

	default:
		panic(fmt.Sprintf("CSC Not implemented: %s\n", old.Type))

	}
	// TODO(tlim): Depending on the rType, we will have to set different fields.
	return zer
}
