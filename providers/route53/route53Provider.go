package route53

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff2"
	"github.com/StackExchange/dnscontrol/v3/pkg/printer"
	"github.com/StackExchange/dnscontrol/v3/pkg/txtutil"
	"github.com/StackExchange/dnscontrol/v3/providers"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	r53 "github.com/aws/aws-sdk-go-v2/service/route53"
	r53Types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	r53d "github.com/aws/aws-sdk-go-v2/service/route53domains"
	r53dTypes "github.com/aws/aws-sdk-go-v2/service/route53domains/types"
)

type route53Provider struct {
	client          *r53.Client
	registrar       *r53d.Client
	delegationSet   *string
	zonesByID       map[string]r53Types.HostedZone
	zonesByDomain   map[string]r53Types.HostedZone
	originalRecords []r53Types.ResourceRecordSet
}

func newRoute53Reg(conf map[string]string) (providers.Registrar, error) {
	return newRoute53(conf, nil)
}

func newRoute53Dsp(conf map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	return newRoute53(conf, metadata)
}

func newRoute53(m map[string]string, metadata json.RawMessage) (*route53Provider, error) {
	optFns := []func(*config.LoadOptions) error{
		// Route53 uses a global endpoint and route53domains
		// currently only has a single regional endpoint in us-east-1
		// http://docs.aws.amazon.com/general/latest/gr/rande.html#r53_region
		config.WithRegion("us-east-1"),
	}

	keyID, secretKey, tokenID := m["KeyId"], m["SecretKey"], m["Token"]
	// Token is optional and left empty unless required
	if keyID != "" || secretKey != "" {
		optFns = append(optFns, config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(keyID, secretKey, tokenID)))
	}

	config, err := config.LoadDefaultConfig(context.Background(), optFns...)
	if err != nil {
		return nil, err
	}

	var dls *string
	if val, ok := m["DelegationSet"]; ok {
		printer.Printf("ROUTE53 DelegationSet %s configured\n", val)
		dls = aws.String(val)
	}
	api := &route53Provider{client: r53.NewFromConfig(config), registrar: r53d.NewFromConfig(config), delegationSet: dls}
	err = api.getZones()
	if err != nil {
		return nil, err
	}
	return api, nil
}

var features = providers.DocumentationNotes{
	providers.CanGetZones:            providers.Can(),
	providers.CanUseAlias:            providers.Cannot("R53 does not provide a generic ALIAS functionality. Use R53_ALIAS instead."),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseRoute53Alias:     providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.DocCreateDomains:       providers.Can(),
	providers.DocDualHost:            providers.Can(),
	providers.DocOfficiallySupported: providers.Can(),
}

func init() {
	fns := providers.DspFuncs{
		Initializer:   newRoute53Dsp,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType("ROUTE53", fns, features)
	providers.RegisterRegistrarType("ROUTE53", newRoute53Reg)
	providers.RegisterCustomRecordType("R53_ALIAS", "ROUTE53", "")
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
			printer.Printf("============ Route53 rate limit exceeded. Waiting %s to retry.\n", sleepTime)
			time.Sleep(sleepTime)
		} else {
			return
		}
	}
}

// ListZones lists the zones on this account.
func (r *route53Provider) ListZones() ([]string, error) {
	if err := r.getZones(); err != nil {
		return nil, err
	}
	var zones []string
	for i := range r.zonesByDomain {
		zones = append(zones, i)
	}
	return zones, nil
}

func (r *route53Provider) getZones() error {
	if r.zonesByDomain != nil {
		return nil
	}

	var nextMarker *string
	r.zonesByDomain = make(map[string]r53Types.HostedZone)
	r.zonesByID = make(map[string]r53Types.HostedZone)
	for {
		var out *r53.ListHostedZonesOutput
		var err error
		withRetry(func() error {
			inp := &r53.ListHostedZonesInput{Marker: nextMarker}
			out, err = r.client.ListHostedZones(context.Background(), inp)
			return err
		})
		if err != nil && strings.Contains(err.Error(), "is not authorized") {
			return errors.New("check your credentials, you're not authorized to perform actions on Route 53 AWS Service")
		} else if err != nil {
			return err
		}
		for _, z := range out.HostedZones {
			domain := strings.TrimSuffix(aws.ToString(z.Name), ".")
			r.zonesByDomain[domain] = z
			r.zonesByID[parseZoneID(aws.ToString(z.Id))] = z
		}
		if out.NextMarker != nil {
			nextMarker = out.NextMarker
		} else {
			break
		}
	}
	return nil
}

type errDomainNoExist struct {
	domain string
}

type errZoneNoExist struct {
	zoneID string
}

func (e errDomainNoExist) Error() string {
	return fmt.Sprintf("Domain %s not found in your route 53 account", e.domain)
}

func (e errZoneNoExist) Error() string {
	return fmt.Sprintf("Zone with id %s not found in your route 53 account", e.zoneID)
}

func (r *route53Provider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	if err := r.getZones(); err != nil {
		return nil, err
	}

	zone, ok := r.zonesByDomain[domain]
	if !ok {
		return nil, errDomainNoExist{domain}
	}
	var z *r53.GetHostedZoneOutput
	var err error
	withRetry(func() error {
		z, err = r.client.GetHostedZone(context.Background(), &r53.GetHostedZoneInput{Id: zone.Id})
		return err
	})
	if err != nil {
		return nil, err
	}

	var nss []string
	if z.DelegationSet != nil {
		nss = z.DelegationSet.NameServers
	}
	return models.ToNameservers(nss)
}

func (r *route53Provider) GetZoneRecords(domain string) (models.Records, error) {
	if err := r.getZones(); err != nil {
		return nil, err
	}

	if zone, ok := r.zonesByDomain[domain]; ok {
		return r.getZoneRecords(zone)
	}

	return nil, errDomainNoExist{domain}
}

func (r *route53Provider) getZone(dc *models.DomainConfig) (r53Types.HostedZone, error) {
	if err := r.getZones(); err != nil {
		return r53Types.HostedZone{}, err
	}

	if zoneID, ok := dc.Metadata["zone_id"]; ok {
		zone, ok := r.zonesByID[zoneID]
		if !ok {
			return r53Types.HostedZone{}, errZoneNoExist{zoneID}
		}
		return zone, nil
	}

	if zone, ok := r.zonesByDomain[dc.Name]; ok {
		return zone, nil
	}

	return r53Types.HostedZone{}, errDomainNoExist{dc.Name}
}

func (r *route53Provider) getZoneRecords(zone r53Types.HostedZone) (models.Records, error) {
	records, err := r.fetchRecordSets(zone.Id)
	if err != nil {
		return nil, err
	}
	r.originalRecords = records

	var existingRecords = []*models.RecordConfig{}
	for _, set := range records {
		rts, err := nativeToRecords(set, unescape(zone.Name))
		if err != nil {
			return nil, err
		}
		existingRecords = append(existingRecords, rts...)
	}
	return existingRecords, nil
}

func (r *route53Provider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	dc.Punycode()

	zone, err := r.getZone(dc)
	if err != nil {
		return nil, err
	}

	existingRecords, err := r.getZoneRecords(zone)
	if err != nil {
		return nil, err
	}

	for _, want := range dc.Records {
		// update zone_id to current zone.id if not specified by the user
		if want.Type == "R53_ALIAS" && want.R53Alias["zone_id"] == "" {
			want.R53Alias["zone_id"] = getZoneID(zone, want)
		}
	}

	// Normalize
	models.PostProcessRecords(existingRecords)
	txtutil.SplitSingleLongTxt(dc.Records) // Autosplit long TXT records

	var corrections []*models.Correction
	if !diff2.EnableDiff2 || true { // Remove "|| true" when diff2 version arrives

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
		dels := []r53Types.Change{}
		delDesc := []string{}
		changes := []r53Types.Change{}
		changeDesc := []string{}

		for _, currentKey := range updateOrder {
			recs := updates[currentKey]
			// If there are no records in our desired state for a key, this
			// indicates we should delete all records at that key.
			if len(recs) == 0 {
				// To delete, we submit the original resource set we got from r53.
				var (
					rrset r53Types.ResourceRecordSet
					found bool
				)
				// Find the original resource set:
				for _, orec := range r.originalRecords {
					if unescape(orec.Name) == currentKey.NameFQDN && (string(orec.Type) == currentKey.Type || currentKey.Type == "R53_ALIAS_"+string(orec.Type)) {
						rrset = orec
						found = true
						break
					}
				}
				if !found {
					// This should not happen.
					return nil, fmt.Errorf("no record set found to delete. Name: '%s'. Type: '%s'", currentKey.NameFQDN, currentKey.Type)
				}
				// Assemble the change and add it to the list:
				chg := r53Types.Change{
					Action:            r53Types.ChangeActionDelete,
					ResourceRecordSet: &rrset,
				}
				dels = append(dels, chg)
				delDesc = append(delDesc, strings.Join(namesToUpdate[currentKey], "\n"))
			} else {
				// If it isn't a delete, it must be either a change or create. In
				// either case, we build a new record set from the desired state and
				// UPSERT it.

				if strings.HasPrefix(currentKey.Type, "R53_ALIAS_") {
					// Each R53_ALIAS_* requires an individual change.
					if len(recs) != 1 {
						log.Fatal("Only one R53_ALIAS_ permitted on a label")
					}
					for _, rec := range recs {
						rrset := aliasToRRSet(zone, rec)
						rrset.Name = aws.String(currentKey.NameFQDN)
						// Assemble the change and add it to the list:
						chg := r53Types.Change{
							Action:            r53Types.ChangeActionUpsert,
							ResourceRecordSet: rrset,
						}
						changes = append(changes, chg)
						changeDesc = append(changeDesc, strings.Join(namesToUpdate[currentKey], "\n"))
					}
				} else {
					// All other keys combine their updates into one rrset:
					rrset := &r53Types.ResourceRecordSet{
						Name: aws.String(currentKey.NameFQDN),
						Type: r53Types.RRType(currentKey.Type),
					}
					for _, rec := range recs {
						val := rec.GetTargetCombined()
						rr := r53Types.ResourceRecord{
							Value: aws.String(val),
						}
						rrset.ResourceRecords = append(rrset.ResourceRecords, rr)
						i := int64(rec.TTL)
						rrset.TTL = &i // TODO: make sure that ttls are consistent within a set
					}
					// Assemble the change and add it to the list:
					chg := r53Types.Change{
						Action:            r53Types.ChangeActionUpsert,
						ResourceRecordSet: rrset,
					}
					changes = append(changes, chg)
					changeDesc = append(changeDesc, strings.Join(namesToUpdate[currentKey], "\n"))
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
							_, err = r.client.ChangeResourceRecordSets(context.Background(), req)
							return err
						})
						return err
					},
				})
		}

		batcher := newChangeBatcher(dels)
		for batcher.Next() {
			start, end := batcher.Batch()
			batch := dels[start:end]
			descBatchStr := "\n" + strings.Join(delDesc[start:end], "\n") + "\n"
			req := &r53.ChangeResourceRecordSetsInput{
				ChangeBatch: &r53Types.ChangeBatch{Changes: batch},
			}
			addCorrection(descBatchStr, req)
		}
		if err := batcher.Err(); err != nil {
			return nil, err
		}

		batcher = newChangeBatcher(changes)
		for batcher.Next() {
			start, end := batcher.Batch()
			batch := changes[start:end]
			descBatchStr := "\n" + strings.Join(changeDesc[start:end], "\n") + "\n"
			req := &r53.ChangeResourceRecordSetsInput{
				ChangeBatch: &r53Types.ChangeBatch{Changes: batch},
			}
			addCorrection(descBatchStr, req)
		}
		if err := batcher.Err(); err != nil {
			return nil, err
		}

		return corrections, nil

	}

	// Insert Future diff2 version here.

	return corrections, nil
}

func nativeToRecords(set r53Types.ResourceRecordSet, origin string) ([]*models.RecordConfig, error) {
	results := []*models.RecordConfig{}
	if set.AliasTarget != nil {
		rc := &models.RecordConfig{
			Type: "R53_ALIAS",
			TTL:  300,
			R53Alias: map[string]string{
				"type":    string(set.Type),
				"zone_id": aws.ToString(set.AliasTarget.HostedZoneId),
			},
		}
		rc.SetLabelFromFQDN(unescape(set.Name), origin)
		rc.SetTarget(aws.ToString(set.AliasTarget.DNSName))
		results = append(results, rc)
	} else if set.TrafficPolicyInstanceId != nil {
		// skip traffic policy records
	} else {
		for _, rec := range set.ResourceRecords {
			switch rtype := set.Type; rtype {
			case r53Types.RRTypeSoa:
				continue
			case r53Types.RRTypeSpf:
				// route53 uses a custom record type for SPF
				rtype = "TXT"
				fallthrough
			default:
				ty := string(rtype)
				val := *rec.Value

				// AWS Route53 has a bug.  Sometimes it returns a target
				// without a trailing dot. In this case we add the dot.  It is
				// not risky to "just add the dot" because this field never
				// includes shortnames.  That said, we only do it for certain
				// record types where we can show the problem exists.
				// 2022-02-23: NS records do NOT have this bug.
				//
				// NOTE: The dot is missing when the record is added via the
				// AWS web console manually.
				//
				// The next "dnscontrol push" will update the record, even
				// though it doesn't seem to be broken. This only happens once
				// per record.  Sadly the updates only fix the first record.
				// So, if n records are affected by this bug, the next n
				// pushes will be required to clean up all the records.
				// Someone converting a new zone will see this issue for the
				// first n pushes. It will seem odd but this is AWS's bug.
				// The UPSERT command only fixes the first record, even if
				// the UPSET received a list of corrections.
				if ty == "CNAME" || ty == "MX" {
					if !strings.HasSuffix(val, ".") {
						val = val + "."
					}
				}

				rc := &models.RecordConfig{TTL: uint32(aws.ToInt64(set.TTL))}
				rc.SetLabelFromFQDN(unescape(set.Name), origin)
				if err := rc.PopulateFromString(string(rtype), val, origin); err != nil {
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

func aliasToRRSet(zone r53Types.HostedZone, r *models.RecordConfig) *r53Types.ResourceRecordSet {
	target := r.GetTargetField()
	zoneID := getZoneID(zone, r)
	rrset := &r53Types.ResourceRecordSet{
		Type: r53Types.RRType(r.R53Alias["type"]),
		AliasTarget: &r53Types.AliasTarget{
			DNSName:              &target,
			HostedZoneId:         aws.String(zoneID),
			EvaluateTargetHealth: false,
		},
	}
	return rrset
}

func getZoneID(zone r53Types.HostedZone, r *models.RecordConfig) string {
	zoneID := r.R53Alias["zone_id"]
	if zoneID == "" {
		zoneID = aws.ToString(zone.Id)
	}
	return parseZoneID(zoneID)
}

/** Removes "/hostedzone/"" prefix from AWS ZoneId */
func parseZoneID(zoneID string) string {
	return strings.TrimPrefix(zoneID, "/hostedzone/")
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
		domainDetail, err = r.registrar.GetDomainDetail(context.Background(), &r53d.GetDomainDetailInput{DomainName: domainName})
		return err
	})
	if err != nil {
		return nil, err
	}

	nameservers := []string{}
	for _, ns := range domainDetail.Nameservers {
		nameservers = append(nameservers, aws.ToString(ns.Name))
	}

	return nameservers, nil
}

func (r *route53Provider) updateRegistrarNameservers(domainName string, nameservers []string) (*string, error) {
	servers := make([]r53dTypes.Nameserver, len(nameservers))
	for i := range nameservers {
		servers[i] = r53dTypes.Nameserver{Name: aws.String(nameservers[i])}
	}
	var domainUpdate *r53d.UpdateDomainNameserversOutput
	var err error
	withRetry(func() error {
		domainUpdate, err = r.registrar.UpdateDomainNameservers(context.Background(), &r53d.UpdateDomainNameserversInput{
			DomainName:  aws.String(domainName),
			Nameservers: servers,
		})
		return err
	})
	if err != nil {
		return nil, err
	}

	return domainUpdate.OperationId, nil
}

func (r *route53Provider) fetchRecordSets(zoneID *string) ([]r53Types.ResourceRecordSet, error) {
	if zoneID == nil || *zoneID == "" {
		return nil, nil
	}
	var next *string
	var nextType r53Types.RRType
	var records []r53Types.ResourceRecordSet
	for {
		listInput := &r53.ListResourceRecordSetsInput{
			HostedZoneId:    zoneID,
			StartRecordName: next,
			StartRecordType: nextType,
			MaxItems:        aws.Int32(100),
		}
		var list *r53.ListResourceRecordSetsOutput
		var err error
		withRetry(func() error {
			list, err = r.client.ListResourceRecordSets(context.Background(), listInput)
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
	if err := r.getZones(); err != nil {
		return err
	}

	if _, ok := r.zonesByDomain[domain]; ok {
		return nil
	}
	if r.delegationSet != nil {
		printer.Printf("Adding zone for %s to route 53 account with delegationSet %s\n", domain, *r.delegationSet)
	} else {
		printer.Printf("Adding zone for %s to route 53 account\n", domain)
	}
	in := &r53.CreateHostedZoneInput{
		Name:            &domain,
		DelegationSetId: r.delegationSet,
		CallerReference: aws.String(fmt.Sprint(time.Now().UnixNano())),
	}

	// reset zone cache
	r.zonesByDomain = nil
	r.zonesByID = nil

	var err error
	withRetry(func() error {
		_, err := r.client.CreateHostedZone(context.Background(), in)
		return err
	})
	return err
}

// changeBatcher takes a set of r53Types.Changes and turns them into a series of
// batches that meet the limits of the ChangeResourceRecordSets API.
//
// See also: https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/DNSLimitations.html#limits-api-requests-changeresourcerecordsets
type changeBatcher struct {
	changes []r53Types.Change

	maxSize  int // Max records per request.
	maxChars int // Max record value characters per request.

	start, end int   // Cursors into changes.
	err        error // Populated by Next.
}

// newChangeBatcher returns a new changeBatcher.
func newChangeBatcher(changes []r53Types.Change) *changeBatcher {
	return &changeBatcher{
		changes:  changes,
		maxSize:  1000,  // "A request cannot contain more than 1,000 ResourceRecord elements."
		maxChars: 32000, // "The sum of the number of characters (including spaces) in all Value elements in a request cannot exceed 32,000 characters."
	}
}

// Next returns true if there is another batch of Changes.
// It returns false if there are no more batches or an error occurred.
func (b *changeBatcher) Next() bool {
	if b.end >= len(b.changes) || b.err != nil {
		return false
	}

	start, end := b.end, b.end
	var (
		reqSize  int
		reqChars int
	)
	for end < len(b.changes) {
		c := &b.changes[end]

		// Check that we won't exceed 1000 ResourceRecords in the request.
		rrsetSize := len(c.ResourceRecordSet.ResourceRecords)
		if c.Action == r53Types.ChangeActionUpsert {
			// "When the value of the Action element is UPSERT, each ResourceRecord element is counted twice."
			rrsetSize *= 2
		}
		if newReqSize := reqSize + rrsetSize; newReqSize > b.maxSize {
			break
		} else {
			reqSize = newReqSize
		}

		// Check that we won't exceed 32000 Value characters in the request.
		var rrsetChars int
		for _, rr := range c.ResourceRecordSet.ResourceRecords {
			rrsetChars += utf8.RuneCountInString(aws.ToString(rr.Value))
		}
		if c.Action == r53Types.ChangeActionUpsert {
			// "When the value of the Action element is UPSERT, each character in a Value element is counted twice."
			rrsetChars *= 2
		}
		if newReqChars := reqChars + rrsetChars; newReqChars > b.maxChars {
			break
		} else {
			reqChars = newReqChars
		}

		end++
	}

	if start == end {
		b.err = errors.New("could not create ChangeResourceRecordSets request within AWS API limits")
		return false
	}

	b.start = start
	b.end = end

	return true
}

// Batch returns the current batch. It should only be called
// after Next returns true.
func (b *changeBatcher) Batch() (start, end int) {
	return b.start, b.end
}

// Err returns the error encountered during the previous call to Next.
func (b *changeBatcher) Err() error {
	return b.err
}
