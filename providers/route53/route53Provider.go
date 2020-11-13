package route53

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	r53 "github.com/aws/aws-sdk-go/service/route53"
	r53d "github.com/aws/aws-sdk-go/service/route53domains"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/providers"
)

type route53Provider struct {
	client          *r53.Route53
	registrar       *r53d.Route53Domains
	delegationSet   *string
	zones           map[string]*r53.HostedZone
	originalRecords []*r53.ResourceRecordSet
}

func newRoute53Reg(conf map[string]string) (providers.Registrar, error) {
	return newRoute53(conf, nil)
}

func newRoute53Dsp(conf map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	return newRoute53(conf, metadata)
}

func newRoute53(m map[string]string, metadata json.RawMessage) (*route53Provider, error) {
	keyID, secretKey, tokenID := m["KeyId"], m["SecretKey"], m["Token"]

	// Route53 uses a global endpoint and route53domains
	// currently only has a single regional endpoint in us-east-1
	// http://docs.aws.amazon.com/general/latest/gr/rande.html#r53_region
	config := &aws.Config{
		Region: aws.String("us-east-1"),
	}

	// Token is optional and left empty unless required
	if keyID != "" || secretKey != "" {
		config.Credentials = credentials.NewStaticCredentials(keyID, secretKey, tokenID)
	}
	sess := session.Must(session.NewSession(config))

	var dls *string
	if val, ok := m["DelegationSet"]; ok {
		fmt.Printf("ROUTE53 DelegationSet %s configured\n", val)
		dls = sPtr(val)
	}
	api := &route53Provider{client: r53.New(sess), registrar: r53d.New(sess), delegationSet: dls}
	err := api.getZones()
	if err != nil {
		return nil, err
	}
	return api, nil
}

var features = providers.DocumentationNotes{
	providers.CanUseAlias:            providers.Cannot("R53 does not provide a generic ALIAS functionality. Use R53_ALIAS instead."),
	providers.DocCreateDomains:       providers.Can(),
	providers.DocDualHost:            providers.Can(),
	providers.DocOfficiallySupported: providers.Can(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseTXTMulti:         providers.Can(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseRoute53Alias:     providers.Can(),
	providers.CanGetZones:            providers.Can(),
}

func init() {
	providers.RegisterDomainServiceProviderType("ROUTE53", newRoute53Dsp, features)
	providers.RegisterRegistrarType("ROUTE53", newRoute53Reg)
	providers.RegisterCustomRecordType("R53_ALIAS", "ROUTE53", "")
}

func sPtr(s string) *string {
	return &s
}

func withRetry(f func() error) {
	const maxRetries = 23
	// TODO: exponential backoff
	const sleepTime = 5 * time.Second
	var currentRetry int
	for {
		err := f()
		if err == nil {
			return
		}
		if strings.Contains(err.Error(), "Rate exceeded") {
			currentRetry++
			if currentRetry >= maxRetries {
				return
			}
			fmt.Printf("============ Route53 rate limit exceeded. Waiting %s to retry.\n", sleepTime)
			time.Sleep(sleepTime)
		} else {
			return
		}
	}
}

// ListZones lists the zones on this account.
func (r *route53Provider) ListZones() ([]string, error) {
	var zones []string
	// Assumes r.zones was filled already by newRoute53().
	for i := range r.zones {
		zones = append(zones, i)
	}
	return zones, nil
}

func (r *route53Provider) getZones() error {
	var nextMarker *string
	r.zones = make(map[string]*r53.HostedZone)
	for {
		var out *r53.ListHostedZonesOutput
		var err error
		withRetry(func() error {
			inp := &r53.ListHostedZonesInput{Marker: nextMarker}
			out, err = r.client.ListHostedZones(inp)
			return err
		})
		if err != nil && strings.Contains(err.Error(), "is not authorized") {
			return errors.New("check your credentials, you're not authorized to perform actions on Route 53 AWS Service")
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
	var z *r53.GetHostedZoneOutput
	var err error
	withRetry(func() error {
		z, err = r.client.GetHostedZone(&r53.GetHostedZoneInput{Id: zone.Id})
		return err
	})
	if err != nil {
		return nil, err
	}

	var nss []string
	if z.DelegationSet != nil {
		for _, nsPtr := range z.DelegationSet.NameServers {
			nss = append(nss, *nsPtr)
		}
	}
	return models.ToNameservers(nss)
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (r *route53Provider) GetZoneRecords(domain string) (models.Records, error) {

	zone, ok := r.zones[domain]
	if !ok {
		return nil, errNoExist{domain}
	}

	records, err := r.fetchRecordSets(zone.Id)
	if err != nil {
		return nil, err
	}
	r.originalRecords = records

	var existingRecords = []*models.RecordConfig{}
	for _, set := range records {
		rts, err := nativeToRecords(set, domain)
		if err != nil {
			return nil, err
		}
		existingRecords = append(existingRecords, rts...)
	}
	return existingRecords, nil
}

func (r *route53Provider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	dc.Punycode()

	var corrections = []*models.Correction{}

	existingRecords, err := r.GetZoneRecords(dc.Name)
	if err != nil {
		return nil, err
	}

	zone, ok := r.zones[dc.Name]
	if !ok {
		return nil, errNoExist{dc.Name}
	}

	for _, want := range dc.Records {
		// update zone_id to current zone.id if not specified by the user
		if want.Type == "R53_ALIAS" && want.R53Alias["zone_id"] == "" {
			want.R53Alias["zone_id"] = getZoneID(zone, want)
		}
	}

	// Normalize
	models.PostProcessRecords(existingRecords)

	// diff
	differ := diff.New(dc, getAliasMap)
	namesToUpdate, err := differ.ChangedGroups(existingRecords)
	if err != nil {
		return nil, err
	}

	if len(namesToUpdate) == 0 {
		return nil, nil
	}

	updates := map[models.RecordKey][]*models.RecordConfig{}

	// for each name we need to update, collect relevant records from our desired domain state
	for k := range namesToUpdate {
		updates[k] = nil
		for _, rc := range dc.Records {
			if rc.Key() == k {
				updates[k] = append(updates[k], rc)
			}
		}
	}

	// updateOrder is the order that the updates will happen.
	// The order should be sorted by NameFQDN, then Type, with R53_ALIAS_*
	// types sorted after all other types. R53_ALIAS_* needs to be last
	// because they are order dependent (aliases must refer to labels
	// that already exist).
	var updateOrder []models.RecordKey
	// Collect the keys
	for k := range updates {
		updateOrder = append(updateOrder, k)
	}
	// Sort themm
	sort.Slice(updateOrder, func(i, j int) bool {
		if updateOrder[i].Type == updateOrder[j].Type {
			return updateOrder[i].NameFQDN < updateOrder[j].NameFQDN
		}

		if strings.HasPrefix(updateOrder[i].Type, "R53_ALIAS_") {
			return false
		}
		if strings.HasPrefix(updateOrder[j].Type, "R53_ALIAS_") {
			return true
		}

		if updateOrder[i].NameFQDN == updateOrder[j].NameFQDN {
			return updateOrder[i].Type < updateOrder[j].Type
		}
		return updateOrder[i].NameFQDN < updateOrder[j].NameFQDN
	})

	// we collect all changes into one of two categories now:
	// pure deletions where we delete an entire record set,
	// or changes where we upsert an entire record set.
	dels := []*r53.Change{}
	delDesc := []string{}
	changes := []*r53.Change{}
	changeDesc := []string{}

	for _, k := range updateOrder {
		recs := updates[k]
		// If there are no records in our desired state for a key, this
		// indicates we should delete all records at that key.
		if len(recs) == 0 {
			// To delete, we submit the original resource set we got from r53.
			var rrset *r53.ResourceRecordSet
			// Find the original resource set:
			for _, r := range r.originalRecords {
				if unescape(r.Name) == k.NameFQDN && (*r.Type == k.Type || k.Type == "R53_ALIAS_"+*r.Type) {
					rrset = r
					break
				}
			}
			if rrset == nil {
				// This should not happen.
				return nil, fmt.Errorf("no record set found to delete. Name: '%s'. Type: '%s'", k.NameFQDN, k.Type)
			}
			// Assemble the change and add it to the list:
			chg := &r53.Change{
				Action:            sPtr("DELETE"),
				ResourceRecordSet: rrset,
			}
			dels = append(dels, chg)
			delDesc = append(delDesc, strings.Join(namesToUpdate[k], "\n"))
		} else {
			// If it isn't a delete, it must be either a change or create. In
			// either case, we build a new record set from the desired state and
			// UPSERT it.

			if strings.HasPrefix(k.Type, "R53_ALIAS_") {
				// Each R53_ALIAS_* requires an individual change.
				if len(recs) != 1 {
					log.Fatal("Only one R53_ALIAS_ permitted on a label")
				}
				for _, r := range recs {
					rrset := aliasToRRSet(zone, r)
					rrset.Name = sPtr(k.NameFQDN)
					// Assemble the change and add it to the list:
					chg := &r53.Change{
						Action:            sPtr("UPSERT"),
						ResourceRecordSet: rrset,
					}
					changes = append(changes, chg)
					changeDesc = append(changeDesc, strings.Join(namesToUpdate[k], "\n"))
				}
			} else {
				// All other keys combine their updates into one rrset:
				rrset := &r53.ResourceRecordSet{
					Name: sPtr(k.NameFQDN),
					Type: sPtr(k.Type),
				}
				for _, r := range recs {
					val := r.GetTargetCombined()
					rr := &r53.ResourceRecord{
						Value: &val,
					}
					rrset.ResourceRecords = append(rrset.ResourceRecords, rr)
					i := int64(r.TTL)
					rrset.TTL = &i // TODO: make sure that ttls are consistent within a set
				}
				// Assemble the change and add it to the list:
				chg := &r53.Change{
					Action:            sPtr("UPSERT"),
					ResourceRecordSet: rrset,
				}
				changes = append(changes, chg)
				changeDesc = append(changeDesc, strings.Join(namesToUpdate[k], "\n"))
			}

		}
	}

	addCorrection := func(msg string, req *r53.ChangeResourceRecordSetsInput) {
		corrections = append(corrections,
			&models.Correction{
				Msg: msg,
				F: func() error {
					var err error
					req.HostedZoneId = zone.Id
					withRetry(func() error {
						_, err = r.client.ChangeResourceRecordSets(req)
						return err
					})
					return err
				},
			})
	}

	getBatchSize := func(size, max int) int {
		if size > max {
			return max
		}
		return size
	}

	for len(dels) > 0 {
		batchSize := getBatchSize(len(dels), 1000)
		batch := dels[:batchSize]
		dels = dels[batchSize:]
		delDescBatch := delDesc[:batchSize]
		delDesc = delDesc[batchSize:]

		delDescBatchStr := "\n" + strings.Join(delDescBatch, "\n") + "\n"

		delReq := &r53.ChangeResourceRecordSetsInput{
			ChangeBatch: &r53.ChangeBatch{Changes: batch},
		}
		addCorrection(delDescBatchStr, delReq)
	}

	for len(changes) > 0 {
		batchSize := getBatchSize(len(changes), 500)
		batch := changes[:batchSize]
		changes = changes[batchSize:]
		changeDescBatch := changeDesc[:batchSize]
		changeDesc = changeDesc[batchSize:]
		changeDescBatchStr := "\n" + strings.Join(changeDescBatch, "\n") + "\n"

		changeReq := &r53.ChangeResourceRecordSetsInput{
			ChangeBatch: &r53.ChangeBatch{Changes: batch},
		}
		addCorrection(changeDescBatchStr, changeReq)
	}

	return corrections, nil

}

func nativeToRecords(set *r53.ResourceRecordSet, origin string) ([]*models.RecordConfig, error) {
	results := []*models.RecordConfig{}
	if set.AliasTarget != nil {
		rc := &models.RecordConfig{
			Type: "R53_ALIAS",
			TTL:  300,
			R53Alias: map[string]string{
				"type":    *set.Type,
				"zone_id": *set.AliasTarget.HostedZoneId,
			},
		}
		rc.SetLabelFromFQDN(unescape(set.Name), origin)
		rc.SetTarget(aws.StringValue(set.AliasTarget.DNSName))
		results = append(results, rc)
	} else if set.TrafficPolicyInstanceId != nil {
		// skip traffic policy records
	} else {
		for _, rec := range set.ResourceRecords {
			switch rtype := *set.Type; rtype {
			case "SOA":
				continue
			case "SPF":
				// route53 uses a custom record type for SPF
				rtype = "TXT"
				fallthrough
			default:
				rc := &models.RecordConfig{TTL: uint32(*set.TTL)}
				rc.SetLabelFromFQDN(unescape(set.Name), origin)
				if err := rc.PopulateFromString(rtype, *rec.Value, origin); err != nil {
					return nil, fmt.Errorf("unparsable record received from R53: %w", err)
				}
				results = append(results, rc)
			}
		}
	}
	return results, nil
}

func getAliasMap(r *models.RecordConfig) map[string]string {
	if r.Type != "R53_ALIAS" {
		return nil
	}
	return r.R53Alias
}

func aliasToRRSet(zone *r53.HostedZone, r *models.RecordConfig) *r53.ResourceRecordSet {
	target := r.GetTargetField()
	zoneID := getZoneID(zone, r)
	targetHealth := false
	rrset := &r53.ResourceRecordSet{
		Type: sPtr(r.R53Alias["type"]),
		AliasTarget: &r53.AliasTarget{
			DNSName:              &target,
			HostedZoneId:         aws.String(zoneID),
			EvaluateTargetHealth: &targetHealth,
		},
	}
	return rrset
}

func getZoneID(zone *r53.HostedZone, r *models.RecordConfig) string {
	zoneID := r.R53Alias["zone_id"]
	if zoneID == "" {
		zoneID = aws.StringValue(zone.Id)
	}
	if strings.HasPrefix(zoneID, "/hostedzone/") {
		zoneID = strings.TrimPrefix(zoneID, "/hostedzone/")
	}
	return zoneID
}

func (r *route53Provider) GetRegistrarCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	corrections := []*models.Correction{}
	actualSet, err := r.getRegistrarNameservers(&dc.Name)
	if err != nil {
		return nil, err
	}
	sort.Strings(actualSet)
	actual := strings.Join(actualSet, ",")

	expectedSet := []string{}
	for _, ns := range dc.Nameservers {
		expectedSet = append(expectedSet, ns.Name)
	}
	sort.Strings(expectedSet)
	expected := strings.Join(expectedSet, ",")

	if actual != expected {
		return []*models.Correction{
			{
				Msg: fmt.Sprintf("Update nameservers %s -> %s", actual, expected),
				F: func() error {
					_, err := r.updateRegistrarNameservers(dc.Name, expectedSet)
					return err
				},
			},
		}, nil
	}

	return corrections, nil
}

func (r *route53Provider) getRegistrarNameservers(domainName *string) ([]string, error) {
	var domainDetail *r53d.GetDomainDetailOutput
	var err error
	withRetry(func() error {
		domainDetail, err = r.registrar.GetDomainDetail(&r53d.GetDomainDetailInput{DomainName: domainName})
		return err
	})
	if err != nil {
		return nil, err
	}

	nameservers := []string{}
	for _, ns := range domainDetail.Nameservers {
		nameservers = append(nameservers, *ns.Name)
	}

	return nameservers, nil
}

func (r *route53Provider) updateRegistrarNameservers(domainName string, nameservers []string) (*string, error) {
	servers := []*r53d.Nameserver{}
	for i := range nameservers {
		servers = append(servers, &r53d.Nameserver{Name: &nameservers[i]})
	}
	var domainUpdate *r53d.UpdateDomainNameserversOutput
	var err error
	withRetry(func() error {
		domainUpdate, err = r.registrar.UpdateDomainNameservers(&r53d.UpdateDomainNameserversInput{
			DomainName: &domainName, Nameservers: servers})
		return err
	})
	if err != nil {
		return nil, err
	}

	return domainUpdate.OperationId, nil
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
		var list *r53.ListResourceRecordSetsOutput
		var err error
		withRetry(func() error {
			list, err = r.client.ListResourceRecordSets(listInput)
			return err
		})
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

// we have to process names from route53 to match what we expect and to remove their odd octal encoding
func unescape(s *string) string {
	if s == nil {
		return ""
	}
	name := strings.TrimSuffix(*s, ".")
	name = strings.Replace(name, `\052`, "*", -1) // TODO: escape all octal sequences
	return name
}

func (r *route53Provider) EnsureDomainExists(domain string) error {
	if _, ok := r.zones[domain]; ok {
		return nil
	}
	if r.delegationSet != nil {
		fmt.Printf("Adding zone for %s to route 53 account with delegationSet %s\n", domain, *r.delegationSet)
	} else {
		fmt.Printf("Adding zone for %s to route 53 account\n", domain)
	}
	in := &r53.CreateHostedZoneInput{
		Name:            &domain,
		DelegationSetId: r.delegationSet,
		CallerReference: sPtr(fmt.Sprint(time.Now().UnixNano())),
	}
	var err error
	withRetry(func() error {
		_, err := r.client.CreateHostedZone(in)
		return err
	})
	return err
}
