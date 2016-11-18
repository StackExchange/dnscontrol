package dnsmadeeasy

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/StackExchange/dnscontrol/providers/diff"
	"github.com/mhenderson-so/godnsmadeeasy/src/GoDNSMadeEasy"
)

type dmeProvider struct {
	client *GoDNSMadeEasy.GoDMEConfig
	zones  map[string]*GoDNSMadeEasy.Domain
}

func newDNSMadeEasy(m map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	keyID, secretKey, apiURL := m["KeyId"], m["SecretKey"], m["APIUrl"]
	if keyID == "" || secretKey == "" {
		return nil, fmt.Errorf("DNS Made Easy KeyId and SecretKey must be provided")
	}
	if apiURL == "" {
		return nil, fmt.Errorf("DNS Made Easy API URL must be provided")
	}

	dmeClient, err := GoDNSMadeEasy.NewGoDNSMadeEasy(&GoDNSMadeEasy.GoDMEConfig{
		APIKey:    keyID,
		SecretKey: secretKey,
		APIUrl:    apiURL,
	})
	if err != nil {
		return nil, err
	}

	api := &dmeProvider{client: dmeClient}
	return api, nil
}

func init() {
	providers.RegisterDomainServiceProviderType("DNSMADEEASY", newDNSMadeEasy)
}
func sPtr(s string) *string {
	return &s
}

func (r *dmeProvider) getZones() error {
	r.zones = make(map[string]*GoDNSMadeEasy.Domain)
	dmeZones, err := r.client.Domains()
	if err != nil {
		return err
	}
	for _, domain := range dmeZones {
		r.zones[domain.Name] = &domain
	}
	return nil
}

//map key for grouping records
type key struct {
	Name, Type string
}

func getKey(r diff.Record) key {
	return key{r.GetName(), r.GetType()}
}

func (r *dmeProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	if r.zones == nil {
		if err := r.getZones(); err != nil {
			return nil, err
		}
	}
	var corrections = []*models.Correction{}
	zone, ok := r.zones[dc.Name]
	records := []GoDNSMadeEasy.Record{}
	// add zone if it doesn't exist
	if !ok {
		//add correction to add zone
		corrections = append(corrections,
			&models.Correction{
				Msg: "Add zone to DNS Made Easy",
				F: func() error {
					//DME Client doesn't support making a zone yet
					newDomain, err := r.client.AddDomain(&GoDNSMadeEasy.Domain{
						Name: dc.Name,
					})
					zone = newDomain
					return err
				},
			})
		//fake zone
		zone = &GoDNSMadeEasy.Domain{
			ID: 0,
		}
	}

	if zone.ID != 0 {
		theseRecords, err := r.client.Records(zone.ID)
		if err != nil {
			return nil, err
		}
		records = theseRecords
	}

	//convert to dnscontrol RecordConfig format
	dc.Nameservers = nil
	var existingRecords = []*models.RecordConfig{}
	for _, rec := range records {
		recordName := rec.Name
		if recordName == "" {
			recordName = "@"
		}

		if rec.Type == "SOA" {
			continue
		}

		if rec.Type == "NS" && recordName == "@" {
			dc.Nameservers = append(dc.Nameservers, &models.Nameserver{Name: strings.TrimSuffix(rec.Value, ".")})
			//	continue
		}

		r := &models.RecordConfig{
			Name:     recordName,
			NameFQDN: buildFQDN(rec.Name, dc.Name),
			Type:     rec.Type,
			Target:   rec.Value,
			TTL:      uint32(rec.TTL),
		}
		existingRecords = append(existingRecords, r)
	}

	e, w := []diff.Record{}, []diff.Record{}
	for _, ex := range existingRecords {
		e = append(e, ex)
	}
	for _, want := range dc.Records {
		if want.TTL == 0 {
			want.TTL = 300
		}
		if want.Type == "MX" {
			want.Target = fmt.Sprintf("%d %s", want.Priority, want.Target)
			want.Priority = 0
		} else if want.Type == "TXT" {
			want.Target = fmt.Sprintf(`"%s"`, want.Target) //FIXME: better escaping/quoting
		}
		w = append(w, want)
	}

	//diff
	_, create, delete, modify := diff.IncrementalDiff(e, w)

	//Create new records by referencing our desired record state
	for _, c := range create {
		newRecord := findRecord(c.Desired, dc.Records)
		if newRecord == nil {
			continue
		}
		corrections = append(corrections,
			&models.Correction{
				Msg: fmt.Sprintf("CREATE %s %s %s %v", newRecord.Type, newRecord.Name, newRecord.Target, newRecord.TTL),
				F: func() error {
					_, err := r.client.AddRecord(zone.ID, &GoDNSMadeEasy.Record{
						Type:        newRecord.Type,
						Name:        fixRecordName(newRecord.Name),
						Value:       newRecord.Target,
						TTL:         int(newRecord.TTL),
						GtdLocation: "DEFAULT",
					})
					return err
				},
			})
	}

	//Delete records by referencing their original state
	for _, d := range delete {
		thisType := d.Existing.GetType()
		thisName := d.Existing.GetName()
		thisVal := d.Existing.GetContent()
		for _, rec := range records {
			recType := rec.Type
			recName := buildFQDN(rec.Name, zone.Name)
			recVal := rec.Value
			recID := rec.ID
			if recType == thisType && recName == thisName && recVal == thisVal {
				corrections = append(corrections,
					&models.Correction{
						Msg: fmt.Sprintf("DELETE %s %s %s %v", rec.Type, rec.Name, rec.Value, rec.TTL),
						F: func() error {
							return r.client.DeleteRecord(zone.ID, recID)
						},
					})
				continue
			}
		}
	}

	for _, m := range modify {
		newRecord := findRecord(m.Desired, dc.Records)
		oldRecord := findRecord(m.Existing, existingRecords)
		if newRecord == nil {
			continue
		}
		for _, rec := range records {
			recType := rec.Type
			recName := buildFQDN(rec.Name, zone.Name)
			recVal := rec.Value
			recID := rec.ID
			if recType == oldRecord.Type && recName == oldRecord.Name && recVal == oldRecord.Target {
				corrections = append(corrections,
					&models.Correction{
						Msg: fmt.Sprintf("UPDATE %s %s %s %v", rec.Type, rec.Name, rec.Value, rec.TTL),
						F: func() error {
							err := r.client.UpdateRecord(zone.ID, &GoDNSMadeEasy.Record{
								ID:          recID,
								Type:        newRecord.Type,
								Name:        fixRecordName(newRecord.Name),
								Value:       newRecord.Target,
								TTL:         int(newRecord.TTL),
								GtdLocation: "DEFAULT",
							})
							return err
						},
					})
			}
		}

	}
	return corrections, nil
}

func buildFQDN(name, domain string) string {
	return strings.TrimPrefix(strings.TrimSuffix(fmt.Sprintf("%s.%s", name, domain), "."), ".")
}

func findRecord(Needle diff.Record, Haystack []*models.RecordConfig) *models.RecordConfig {
	needleType := Needle.GetType()
	needleName := Needle.GetName()
	needleContent := Needle.GetContent()

	for _, straw := range Haystack {
		if straw.NameFQDN == needleName && straw.Type == needleType && straw.Target == needleContent {
			return straw
		}
	}

	return nil
}

func fixRecordName(Name string) string {
	if Name == "@" {
		return ""
	}
	return Name
}
