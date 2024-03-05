package akamaiedgedns

/*
For information about Akamai's "Edge DNS Zone Management API", see:
https://developer.akamai.com/api/cloud_security/edge_dns_zone_management/v2.html

For information about "AkamaiOPEN-edgegrid-golang" library, see:
https://github.com/akamai/AkamaiOPEN-edgegrid-golang
*/

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/printer"
	dnsv2 "github.com/akamai/AkamaiOPEN-edgegrid-golang/configdns-v2"
	"github.com/akamai/AkamaiOPEN-edgegrid-golang/edgegrid"
)

// initialize initializes the "Akamai OPEN EdgeGrid" library
func initialize(clientSecret string, host string, accessToken string, clientToken string) {

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

// zoneDoesExist returns true if the zone exists, false otherwise.
func zoneDoesExist(zonename string) bool {
	_, err := dnsv2.GetZone(zonename)
	return err == nil
}

// createZone create a new zone and creates SOA and NS records for the zone.
// Akamai assigns a unique set of authoritative nameservers for each contract. These authorities should be
// used as the NS records on all zones belonging to this contract.
func createZone(zonename string, contractID string, groupID string) error {
	zone := &dnsv2.ZoneCreate{
		Zone:                  zonename,
		Type:                  "PRIMARY",
		Comment:               "This zone created by DNSControl (http://dnscontrol.org)",
		SignAndServe:          false,
		SignAndServeAlgorithm: "RSA_SHA512",
		ContractId:            contractID,
	}

	queryArgs := &dnsv2.ZoneQueryString{
		Contract: contractID,
		Group:    groupID,
	}

	err := dnsv2.ValidateZone(zone)
	if err != nil {
		return fmt.Errorf("invalid value provided for zone. error: %s", err.Error())
	}

	err = zone.Save(*queryArgs)
	if err != nil {
		return fmt.Errorf("zone create failed. error: %s", err.Error())
	}

	// Indirectly create NS and SOA records
	err = zone.SaveChangelist()
	if err != nil {
		return fmt.Errorf("zone initialization failed. SOA and NS records need to be created")
	}
	err = zone.SubmitChangelist()
	if err != nil {
		return fmt.Errorf("zone create failed. error: %s", err.Error())
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

// listZones lists all zones associated with this contract.
func listZones(contractID string) ([]string, error) {
	queryArgs := dnsv2.ZoneListQueryArgs{
		ContractIds: contractID,
		ShowAll:     true,
	}

	zoneListResp, err := dnsv2.ListZones(queryArgs)
	if err != nil {
		return nil, fmt.Errorf("zone list retrieval failed. error: %s", err.Error())
	}

	edgeDNSZones := zoneListResp.Zones // what we have
	var zones []string                 // what we return

	for _, edgeDNSZone := range edgeDNSZones {
		zones = append(zones, edgeDNSZone.Zone)
	}

	return zones, nil
}

// isAutoDNSSecEnabled returns true if AutoDNSSEC (SignAndServe) is enabled, false otherwise.
func isAutoDNSSecEnabled(zonename string) (bool, error) {
	zone, err := dnsv2.GetZone(zonename)
	if err != nil {
		if dnsv2.IsConfigDNSError(err) && err.(dnsv2.ConfigDNSError).NotFound() {
			return false, fmt.Errorf("zone %s does not exist. error: %s",
				zonename, err.Error())
		}
		return false, fmt.Errorf("error retrieving information for zone %s. error: %s",
			zonename, err.Error())
	}
	return zone.SignAndServe, nil
}

// autoDNSSecEnable enables or disables AutoDNSSEC (SignAndServe) for the zone.
func autoDNSSecEnable(enable bool, zonename string) error {
	zone, err := dnsv2.GetZone(zonename)
	if err != nil {
		if dnsv2.IsConfigDNSError(err) && err.(dnsv2.ConfigDNSError).NotFound() {
			return fmt.Errorf("zone %s does not exist. error: %s",
				zonename, err.Error())
		}
		return fmt.Errorf("error retrieving information for zone %s. error: %s",
			zonename, err.Error())
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
		SignAndServe:          enable, // AutoDNSSEC
		SignAndServeAlgorithm: algorithm,
		TsigKey:               zone.TsigKey,
		EndCustomerId:         zone.EndCustomerId,
		ContractId:            zone.ContractId,
	}

	queryArgs := dnsv2.ZoneQueryString{}

	err = modifiedzone.Update(queryArgs)
	if err != nil {
		return fmt.Errorf("error updating zone %s. error: %s",
			zonename, err.Error())
	}

	return nil
}

// getAuthorities returns the list of authoritative nameservers for the contract.
// Akamai assigns a unique set of authoritative nameservers for each contract. These authorities should be
// used as the NS records on all zones belonging to this contract.
func getAuthorities(contractID string) ([]string, error) {
	authorityResponse, err := dnsv2.GetAuthorities(contractID)
	if err != nil {
		return nil, fmt.Errorf("getAuthorities - contractid %s: authorities retrieval failed. Error: %s",
			contractID, err.Error())
	}
	contracts := authorityResponse.Contracts
	if len(contracts) != 1 {
		return nil, fmt.Errorf("getAuthorities - contractid %s: Expected 1 element in array but got %d",
			contractID, len(contracts))
	}
	cid := contracts[0].ContractID
	if cid != contractID {
		return nil, fmt.Errorf("getAuthorities - contractID %s: got authorities for wrong contractID (%s)",
			contractID, cid)
	}
	authorities := contracts[0].Authorities
	return authorities, nil
}

// rcToRs converts DNSControl RecordConfig records to an AkamaiEdgeDNS recordset.
func rcToRs(records []*models.RecordConfig) (*dnsv2.RecordBody, error) {
	if len(records) == 0 {
		return nil, fmt.Errorf("no records to replace")
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

// createRecordset creates a new AkamaiEdgeDNS recordset in the zone.
func createRecordset(records []*models.RecordConfig, zonename string) error {
	akaRecord, err := rcToRs(records)
	if err != nil {
		return err
	}

	err = akaRecord.Save(zonename, true)
	if err != nil {
		return fmt.Errorf("recordset creation failed. error: %s", err.Error())
	}
	return nil
}

// replaceRecordset replaces an existing AkamaiEdgeDNS recordset in the zone.
func replaceRecordset(records []*models.RecordConfig, zonename string) error {
	akaRecord, err := rcToRs(records)
	if err != nil {
		return err
	}

	err = akaRecord.Update(zonename, true)
	if err != nil {
		return fmt.Errorf("recordset update failed. error: %s", err.Error())
	}
	return nil
}

// deleteRecordset deletes an existing AkamaiEdgeDNS recordset in the zone.
func deleteRecordset(records []*models.RecordConfig, zonename string) error {
	akaRecord, err := rcToRs(records)
	if err != nil {
		return err
	}

	err = akaRecord.Delete(zonename, true)
	if err != nil {
		if dnsv2.IsConfigDNSError(err) && err.(dnsv2.ConfigDNSError).NotFound() {
			return fmt.Errorf("recordset not found")
		}
		return fmt.Errorf("failed to delete recordset. error: %s", err.Error())
	}
	return nil
}

/*
  Example AkamaiEdgeDNS Recordset (as JSON):
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

// getRecords returns all RecordConfig records in the zone.
func getRecords(zonename string) ([]*models.RecordConfig, error) {
	queryArgs := dnsv2.RecordsetQueryArgs{ShowAll: true}

	rsetResp, err := dnsv2.GetRecordsets(zonename, queryArgs)
	if err != nil {
		return nil, fmt.Errorf("recordset list retrieval failed. error: %s", err.Error())
	}

	akaRecordsets := rsetResp.Recordsets     // what we have
	var recordConfigs []*models.RecordConfig // what we return

	// For each AkamaiEdgeDNS recordset...
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
