package route53

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/StackExchange/dnscontrol/providers/diff"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	r53 "github.com/aws/aws-sdk-go/service/route53"
	"github.com/pkg/errors"
)

type route53Provider struct {
	client *r53.Route53
	zones  map[string]*r53.HostedZone
}

func newRoute53(m map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	keyId, secretKey := m["KeyId"], m["SecretKey"]

	config := &aws.Config{
		Region: aws.String("us-west-2"),
	}

	if keyId != "" || secretKey != "" {
		config.Credentials = credentials.NewStaticCredentials(keyId, secretKey, "")
	}
	sess := session.New(config)

	api := &route53Provider{client: r53.New(sess)}
	err := api.getZones()
	if err != nil {
		return nil, err
	}
	return api, nil
}

func init() {
	providers.RegisterDomainServiceProviderType("ROUTE53", newRoute53, providers.CanUsePTR, providers.CanUseSRV)
}
func sPtr(s string) *string {
	return &s
}

func (r *route53Provider) getZones() error {
	var nextMarker *string
	r.zones = make(map[string]*r53.HostedZone)
	for {
		if nextMarker != nil {
			fmt.Println(*nextMarker)
		}
		inp := &r53.ListHostedZonesInput{Marker: nextMarker}
		out, err := r.client.ListHostedZones(inp)
		if err != nil && strings.Contains(err.Error(), "is not authorized") {
			return errors.New("Check your credentials, your not authorized to perform actions on Route 53 AWS Service")
		} else if err != nil {
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

func getKey(r *models.RecordConfig) key {
	return key{r.NameFQDN, r.Type}
}

type errNoExist struct {
	domain string
}

func (e errNoExist) Error() string {
	return fmt.Sprintf("Domain %s not found in your route 53 account", e.domain)
}

func (r *route53Provider) GetNameservers(domain string) ([]*models.Nameserver, error) {

	zone, ok := r.zones[domain]
	if !ok {
		return nil, errNoExist{domain}
	}
	z, err := r.client.GetHostedZone(&r53.GetHostedZoneInput{Id: zone.Id})
	if err != nil {
		return nil, err
	}
	ns := []*models.Nameserver{}
	if z.DelegationSet != nil {
		for _, nsPtr := range z.DelegationSet.NameServers {
			ns = append(ns, &models.Nameserver{Name: *nsPtr})
		}
	}
	return ns, nil
}

func (r *route53Provider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	dc.Punycode()

	var corrections = []*models.Correction{}
	zone, ok := r.zones[dc.Name]
	// add zone if it doesn't exist
	if !ok {
		return nil, errNoExist{dc.Name}
	}

	records, err := r.fetchRecordSets(zone.Id)
	if err != nil {
		return nil, err
	}

	var existingRecords = []*models.RecordConfig{}
	for _, set := range records {
		for _, rec := range set.ResourceRecords {
			if *set.Type == "SOA" {
				continue
			}
			r := &models.RecordConfig{
				NameFQDN:       unescape(set.Name),
				Type:           *set.Type,
				Target:         *rec.Value,
				TTL:            uint32(*set.TTL),
				CombinedTarget: true,
			}
			existingRecords = append(existingRecords, r)
		}
	}
	for _, want := range dc.Records {
		want.MergeToTarget()
	}

	//diff
	differ := diff.New(dc)
	_, create, delete, modify := differ.IncrementalDiff(existingRecords)

	namesToUpdate := map[key][]string{}
	for _, c := range create {
		namesToUpdate[getKey(c.Desired)] = append(namesToUpdate[getKey(c.Desired)], c.String())
	}
	for _, d := range delete {
		namesToUpdate[getKey(d.Existing)] = append(namesToUpdate[getKey(d.Existing)], d.String())
	}
	for _, m := range modify {
		namesToUpdate[getKey(m.Desired)] = append(namesToUpdate[getKey(m.Desired)], m.String())
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

	dels := []*r53.Change{}
	changes := []*r53.Change{}
	changeDesc := ""
	delDesc := ""
	for k, recs := range updates {
		chg := &r53.Change{}
		var rrset *r53.ResourceRecordSet
		if len(recs) == 0 {
			dels = append(dels, chg)
			chg.Action = sPtr("DELETE")
			delDesc += strings.Join(namesToUpdate[k], "\n") + "\n"
			// on delete just submit the original resource set we got from r53.
			for _, r := range records {
				if *r.Name == k.Name+"." && *r.Type == k.Type {
					rrset = r
					break
				}
			}
		} else {
			changes = append(changes, chg)
			changeDesc += strings.Join(namesToUpdate[k], "\n") + "\n"
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

	delReq := &r53.ChangeResourceRecordSetsInput{
		ChangeBatch: &r53.ChangeBatch{Changes: dels},
	}

	addCorrection := func(msg string, req *r53.ChangeResourceRecordSetsInput) {
		corrections = append(corrections,
			&models.Correction{
				Msg: msg,
				F: func() error {
					req.HostedZoneId = zone.Id
					_, err := r.client.ChangeResourceRecordSets(req)
					return err
				},
			})
	}

	if len(dels) > 0 {
		addCorrection(delDesc, delReq)
	}

	if len(changes) > 0 {
		addCorrection(changeDesc, changeReq)
	}

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

func (r *route53Provider) EnsureDomainExists(domain string) error {
	if _, ok := r.zones[domain]; ok {
		return nil
	}
	fmt.Printf("Adding zone for %s to route 53 account\n", domain)
	in := &r53.CreateHostedZoneInput{
		Name:            &domain,
		CallerReference: sPtr(fmt.Sprint(time.Now().UnixNano())),
	}
	_, err := r.client.CreateHostedZone(in)
	return err

}
