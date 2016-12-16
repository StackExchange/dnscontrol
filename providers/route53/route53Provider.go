package route53

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/StackExchange/dnscontrol/providers/diff"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	r53 "github.com/aws/aws-sdk-go/service/route53"
)

type route53Provider struct {
	client *r53.Route53
	zones  map[string]*r53.HostedZone
}

func newRoute53(m map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	keyId, secretKey := m["KeyId"], m["SecretKey"]
	if keyId == "" || secretKey == "" {
		return nil, fmt.Errorf("Route53 KeyId and SecretKey must be provided.")
	}
	sess := session.New(&aws.Config{
		Region:      aws.String("us-west-2"),
		Credentials: credentials.NewStaticCredentials(keyId, secretKey, ""),
	})

	api := &route53Provider{client: r53.New(sess)}
	return api, nil
}

func init() {
	providers.RegisterDomainServiceProviderType("ROUTE53", newRoute53)
}
func sPtr(s string) *string {
	return &s
}
func (r *route53Provider) getZones() error {
	if r.zones != nil {
		return nil
	}
	var nextMarker *string
	r.zones = make(map[string]*r53.HostedZone)
	for {
		if nextMarker != nil {
			fmt.Println(*nextMarker)
		}
		inp := &r53.ListHostedZonesInput{Marker: nextMarker}
		out, err := r.client.ListHostedZones(inp)
		if err != nil {
			return err
		}
		for _, z := range out.HostedZones {
			domain := strings.TrimSuffix(*z.Name, ".")
			r.zones[domain] = z
		}
		if out.NextMarker != nil {
			nextMarker = out.NextMarker
		} else {
			break
		}
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

func (r *route53Provider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	if err := r.getZones(); err != nil {
		return nil, err
	}
	zone, ok := r.zones[domain]
	if !ok {
		log.Printf("WARNING: Domain %s is not on your route 53 account. Dnscontrol will add it, but you will need to run a second time to configure nameservers properly.", domain)
		return nil, nil
	}
	z, err := r.client.GetHostedZone(&r53.GetHostedZoneInput{Id: zone.Id})
	if err != nil {
		return nil, err
	}
	ns := []*models.Nameserver{}
	for _, nsPtr := range z.DelegationSet.NameServers {
		ns = append(ns, &models.Nameserver{Name: *nsPtr})
	}
	return ns, nil
}

func (r *route53Provider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	if err := r.getZones(); err != nil {
		return nil, err
	}

	var corrections = []*models.Correction{}
	zone, ok := r.zones[dc.Name]
	// add zone if it doesn't exist
	if !ok {
		//add correction to add zone
		corrections = append(corrections,
			&models.Correction{
				Msg: "Add zone to aws",
				F: func() error {
					in := &r53.CreateHostedZoneInput{
						Name:            &dc.Name,
						CallerReference: sPtr(fmt.Sprint(time.Now().UnixNano())),
					}
					out, err := r.client.CreateHostedZone(in)
					zone = out.HostedZone
					return err
				},
			})
		//fake zone
		zone = &r53.HostedZone{
			Id: sPtr(""),
		}
	}

	records, err := r.fetchRecordSets(zone.Id)
	if err != nil {
		return nil, err
	}

	//convert to dnscontrol RecordConfig format
	var existingRecords = []diff.Record{}
	for _, set := range records {
		for _, rec := range set.ResourceRecords {
			if *set.Type == "SOA" {
				continue
			}
			r := &models.RecordConfig{
				NameFQDN: unescape(set.Name),
				Type:     *set.Type,
				Target:   *rec.Value,
				TTL:      uint32(*set.TTL),
			}
			existingRecords = append(existingRecords, r)
		}
	}
	w := []diff.Record{}
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
	changeDesc := ""
	_, create, delete, modify := diff.IncrementalDiff(existingRecords, w)

	namesToUpdate := map[key]bool{}
	for _, c := range create {
		namesToUpdate[getKey(c.Desired)] = true
		changeDesc += fmt.Sprintln(c)
	}
	for _, d := range delete {
		namesToUpdate[getKey(d.Existing)] = true
		changeDesc += fmt.Sprintln(d)
	}
	for _, m := range modify {
		namesToUpdate[getKey(m.Desired)] = true
		changeDesc += fmt.Sprintln(m)
	}

	if len(namesToUpdate) == 0 {
		return nil, nil
	}

	updates := map[key][]*models.RecordConfig{}
	//for each name we need to update, collect relevant records from dc
	for k := range namesToUpdate {
		updates[k] = nil
		for _, rc := range dc.Records {
			if getKey(rc) == k {
				updates[k] = append(updates[k], rc)
			}
		}
	}

	changes := []*r53.Change{}
	for k, recs := range updates {
		chg := &r53.Change{}
		changes = append(changes, chg)
		var rrset *r53.ResourceRecordSet
		if len(recs) == 0 {
			chg.Action = sPtr("DELETE")
			// on delete just submit the original resource set we got from r53.
			for _, r := range records {
				if *r.Name == k.Name+"." && *r.Type == k.Type {
					rrset = r
					break
				}
			}
		} else {
			//on change or create, just build a new record set from our desired state
			chg.Action = sPtr("UPSERT")
			rrset = &r53.ResourceRecordSet{
				Name:            sPtr(k.Name),
				Type:            sPtr(k.Type),
				ResourceRecords: []*r53.ResourceRecord{},
			}
			for _, r := range recs {
				val := r.Target
				rr := &r53.ResourceRecord{
					Value: &val,
				}
				rrset.ResourceRecords = append(rrset.ResourceRecords, rr)
				i := int64(r.TTL)
				rrset.TTL = &i //TODO: make sure that ttls are consistent within a set
			}
		}
		chg.ResourceRecordSet = rrset
	}

	changeReq := &r53.ChangeResourceRecordSetsInput{
		ChangeBatch: &r53.ChangeBatch{Changes: changes},
	}

	corrections = append(corrections,
		&models.Correction{
			Msg: changeDesc,
			F: func() error {
				changeReq.HostedZoneId = zone.Id
				_, err := r.client.ChangeResourceRecordSets(changeReq)
				return err
			},
		})
	return corrections, nil

}

func (r *route53Provider) fetchRecordSets(zoneID *string) ([]*r53.ResourceRecordSet, error) {
	if zoneID == nil || *zoneID == "" {
		return nil, nil
	}
	var next *string
	var nextType *string
	var records []*r53.ResourceRecordSet
	for {
		listInput := &r53.ListResourceRecordSetsInput{
			HostedZoneId:    zoneID,
			StartRecordName: next,
			StartRecordType: nextType,
			MaxItems:        sPtr("100"),
		}
		list, err := r.client.ListResourceRecordSets(listInput)
		if err != nil {
			return nil, err
		}
		records = append(records, list.ResourceRecordSets...)
		if list.NextRecordName != nil {
			next = list.NextRecordName
			nextType = list.NextRecordType
		} else {
			break
		}
	}
	return records, nil
}

//we have to process names from route53 to match what we expect and to remove their odd octal encoding
func unescape(s *string) string {
	if s == nil {
		return ""
	}
	name := strings.TrimSuffix(*s, ".")
	name = strings.Replace(name, `\052`, "*", -1) //TODO: escape all octal sequences
	return name
}
