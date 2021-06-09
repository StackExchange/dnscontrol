/*
  For information about Akamai's "Edge DNS Zone Management API", see:
  https://developer.akamai.com/api/cloud_security/edge_dns_zone_management/v2.html

  For information about "AkamaiOPEN-edgegrid-golang" library, see:
  https://github.com/akamai/AkamaiOPEN-edgegrid-golang
*/
package akamai

import (
	"fmt"
	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/printer"
	dnsv2 "github.com/akamai/AkamaiOPEN-edgegrid-golang/configdns-v2"
	"github.com/akamai/AkamaiOPEN-edgegrid-golang/edgegrid"
)

func AkaInitialize(clientSecret string, host string, accessToken string, clientToken string) {

	eg := edgegrid.Config{
		ClientSecret: clientSecret,
		Host:         host,
		AccessToken:  accessToken,
		ClientToken:  clientToken,
		MaxBody:      131072,
		Debug:        false,
	}
	dnsv2.Init(eg)
}

func AkaZoneDoesExist(zonename string) bool {
	_, err := dnsv2.GetZone(zonename)
	if err == nil {
		return true
	}
	return false
}

func AkaCreateZone(zonename string, contractId string, groupId string) error {
	zone := &dnsv2.ZoneCreate{
		Zone:                  zonename,
		Type:                  "PRIMARY",
		Comment:               "This zone created by DNSControl (https://stackexchange.github.io/dnscontrol/)",
		SignAndServe:          false,
		SignAndServeAlgorithm: "RSA_SHA512",
		ContractId:            contractId,
	}

	queryArgs := &dnsv2.ZoneQueryString{
		Contract: contractId,
		Group:    groupId,
	}

	err := dnsv2.ValidateZone(zone)
	if err != nil {
		return fmt.Errorf("Invalid value provided for zone. Error: %s", err.Error())
	}

	err = zone.Save(*queryArgs)
	if err != nil {
		return fmt.Errorf("Zone create failed. Error: %s", err.Error())
	}

	// Indirectly create NS and SOA records
	err = zone.SaveChangelist()
	if err != nil {
		return fmt.Errorf("Zone initialization failed. SOA and NS records need to be created.")
	}
	err = zone.SubmitChangelist()
	if err != nil {
		return fmt.Errorf("Zone create failed. Error: %s", err.Error())
	}

	printer.Printf("Created zone: %s\n", zone.Zone)
	printer.Printf("  Type: %s\n", zone.Type)
	printer.Printf("  Comment: %s\n", zone.Comment)
	printer.Printf("  SignAndServe: %v\n", zone.SignAndServe)
	printer.Printf("  SignAndServeAlgorithm: %s\n", zone.SignAndServeAlgorithm)
	printer.Printf("  ContractId: %s\n", zone.ContractId)
	printer.Printf("  GroupId: %s\n", queryArgs.Group)

	return nil
}

func AkaListZones(contractId string) ([]string, error) {
	queryArgs := dnsv2.ZoneListQueryArgs{
		ContractIds: contractId,
		ShowAll:     true,
	}

	zoneListResp, err := dnsv2.ListZones(queryArgs)
	if err != nil {
		return nil, fmt.Errorf("Zone List retrieval failed. Error: %s", err.Error())
	}

	akaZones := zoneListResp.Zones // what we have
	var zones []string             // what we return

	for _, akaZone := range akaZones {
		zones = append(zones, akaZone.Zone)
	}

	return zones, nil
}

func AkaIsAutoDnsSecEnabled(zonename string) (bool, error) {
	zone, err := dnsv2.GetZone(zonename)
	if err != nil {
		if dnsv2.IsConfigDNSError(err) && err.(dnsv2.ConfigDNSError).NotFound() {
			return false, fmt.Errorf("Zone %s does not exist. Error: %s",
				zonename, err.Error())
		} else {
			return false, fmt.Errorf("Error retrieving information for zone %s. Error: %s",
				zonename, err.Error())
		}
	}
	return zone.SignAndServe, nil
}

func AkaAutoDnsSecEnable(enable bool, zonename string) error {
	zone, err := dnsv2.GetZone(zonename)
	if err != nil {
		if dnsv2.IsConfigDNSError(err) && err.(dnsv2.ConfigDNSError).NotFound() {
			return fmt.Errorf("Zone %s does not exist. Error: %s",
				zonename, err.Error())
		} else {
			return fmt.Errorf("Error retrieving information for zone %s. Error: %s",
				zonename, err.Error())
		}
	}

	algorithm := "RSA_SHA512"
	if zone.SignAndServeAlgorithm != "" {
		algorithm = zone.SignAndServeAlgorithm
	}

	modifiedzone := &dnsv2.ZoneCreate{
		Zone:                  zone.Zone,
		Type:                  zone.Type,
		Masters:               zone.Masters,
		Comment:               zone.Comment,
		SignAndServe:          enable, // AutoDnsSec enable
		SignAndServeAlgorithm: algorithm,
		TsigKey:               zone.TsigKey,
		EndCustomerId:         zone.EndCustomerId,
		ContractId:            zone.ContractId,
	}

	queryArgs := dnsv2.ZoneQueryString{}

	err = modifiedzone.Update(queryArgs)
	if err != nil {
		return fmt.Errorf("Error updating zone %s. Error: %s",
			zonename, err.Error())
	}

	return nil
}

// Akamai assigns a unique set of authoritative nameservers for each contract. These authorities should be
// used as the NS records on all zones belonging to this contract.
func AkaGetAuthorities(contractId string) ([]string, error) {
	authorityResponse, err := dnsv2.GetAuthorities(contractId)
	if err != nil {
		return nil, fmt.Errorf("AkaGetAuthorities - ContractID %s: Authorities retrieval failed. Error: %s",
			contractId, err.Error())
	}
	contracts := authorityResponse.Contracts
	if len(contracts) != 1 {
		return nil, fmt.Errorf("AkaGetAuthorities - ContractID %s: Expected 1 element in array but got %d",
			contractId, len(contracts))
	}
	cid := contracts[0].ContractID
	if cid != contractId {
		return nil, fmt.Errorf("AkaGetAuthorities - ContractID %s: Got authorities for wrong contractId (%s)",
			contractId, cid)
	}
	authorities := contracts[0].Authorities
	return authorities, nil
}

func createAkaRecord(records []*models.RecordConfig, zonename string) (*dnsv2.RecordBody, error) {
	if len(records) == 0 {
		return nil, fmt.Errorf("Recordset replace failed. No records to replace.")
	}

	akaRecord := &dnsv2.RecordBody{
		Name:       records[0].NameFQDN,
		RecordType: records[0].Type,
		TTL:        int(records[0].TTL),
	}

	for _, r := range records {
		akaRecord.Target = append(akaRecord.Target, r.GetTargetCombined())
	}

	return akaRecord, nil
}

func AkaCreateRecordset(records []*models.RecordConfig, zonename string) error {
	akaRecord, err := createAkaRecord(records, zonename)
	if err != nil {
		return err
	}

	err = akaRecord.Save(zonename, true)
	if err != nil {
		return fmt.Errorf("Recordset creation failed. Error: %s", err.Error())
	}
	return nil
}

func AkaReplaceRecordset(records []*models.RecordConfig, zonename string) error {
	akaRecord, err := createAkaRecord(records, zonename)
	if err != nil {
		return err
	}

	err = akaRecord.Update(zonename, true)
	if err != nil {
		return fmt.Errorf("Recordset update failed. Error: %s", err.Error())
	}
	return nil
}

func AkaDeleteRecordset(records []*models.RecordConfig, zonename string) error {
	akaRecord, err := createAkaRecord(records, zonename)
	if err != nil {
		return err
	}

	err = akaRecord.Delete(zonename, true)
	if err != nil {
		if dnsv2.IsConfigDNSError(err) && err.(dnsv2.ConfigDNSError).NotFound() {
			return fmt.Errorf("Recordset not found")
		} else {
			return fmt.Errorf("Failed to delete recordset. Error: %s", err.Error())
		}
	}
	return nil
}

/*
  Example Recordset (as JSON):
        {
            "name": "test.com",
            "rdata": [
                "a7.akafp.net.",
                "a4.akafp.net.",
                "a0.akafp.net."
            ],
            "ttl": 10000,
            "type": "NS"
        }
*/
func AkaGetRecords(zonename string) ([]*models.RecordConfig, error) {
	queryArgs := dnsv2.RecordsetQueryArgs{ShowAll: true}

	rsetResp, err := dnsv2.GetRecordsets(zonename, queryArgs)
	if err != nil {
		return nil, fmt.Errorf("Recordset list retrieval failed. Error: %s", err.Error())
	}

	akaRecordsets := rsetResp.Recordsets     // what we have
	var recordConfigs []*models.RecordConfig // what we return

	// For each Akamai recordset...
	for _, akarecset := range akaRecordsets {
		akaname := akarecset.Name
		akatype := akarecset.Type
		akattl := akarecset.TTL

		// Don't report the existence of an SOA record (because DnsControl will try to delete the SOA record).
		if akatype == "SOA" {
			continue
		}

		// ... convert the recordset into 1 or more RecordConfig structs
		for _, r := range akarecset.Rdata {
			rc := &models.RecordConfig{
				Type: akatype,
				TTL:  uint32(akattl),
			}
			rc.SetLabelFromFQDN(akaname, zonename)
			err = rc.PopulateFromString(akatype, r, zonename)
			if err != nil {
				return nil, err
			}

			recordConfigs = append(recordConfigs, rc)
		}
	}

	return recordConfigs, nil
}
