package GoDNSMadeEasy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

// Domains returns the list of domains managed by DNS Made Easy
func (dme *GoDMEConfig) Domains() ([]Domain, error) {
	req, err := dme.newRequest("GET", "dns/managed/", nil)
	if err != nil {
		return nil, err
	}

	genericResponse := &GenericResponse{}
	err = dme.doDMERequest(req, &genericResponse)
	if err != nil {
		return nil, err
	}

	domainData := []Domain{}
	err = json.Unmarshal(genericResponse.Data, &domainData)
	return domainData, err
}

// Domain returns the summary data for a single domain. This is essentially the same as Domains(), but only returns one domain.
func (dme *GoDMEConfig) Domain(DomainID int) (*Domain, error) {
	reqStub := fmt.Sprintf("dns/managed/%v", DomainID)
	req, err := dme.newRequest("GET", reqStub, nil)
	if err != nil {
		return nil, err
	}

	domainResponse := &Domain{}
	err = dme.doDMERequest(req, domainResponse)
	if err != nil {
		return nil, err
	}

	return domainResponse, nil
}

// Records returns the records for a given domain. The domain is specified by its ID, which can be retrieved from Domains()
func (dme *GoDMEConfig) Records(DomainID int) ([]Record, error) {
	reqStub := fmt.Sprintf("dns/managed/%v/records", DomainID)
	req, err := dme.newRequest("GET", reqStub, nil)
	if err != nil {
		return nil, err
	}

	genericResponse := &GenericResponse{}
	err = dme.doDMERequest(req, &genericResponse)
	if err != nil {
		return nil, err
	}

	recordData := []Record{}
	err = json.Unmarshal(genericResponse.Data, &recordData)
	return recordData, err

}

// Record returns the record for a given record ID. This is essentially the same as Records(), but only returns one record
func (dme *GoDMEConfig) Record(DomainID, RecordID int) (*Record, error) {
	return nil, fmt.Errorf("Record() Not yet implemented")
}

// SOA returns custom Start of Authority records for an account.
func (dme *GoDMEConfig) SOA() ([]SOA, error) {
	req, err := dme.newRequest("GET", "dns/soa", nil)
	if err != nil {
		return nil, err
	}

	genericResponse := &GenericResponse{}
	err = dme.doDMERequest(req, &genericResponse)
	if err != nil {
		return nil, err
	}

	soaData := []SOA{}
	json.Unmarshal(genericResponse.Data, &soaData)

	return soaData, nil
}

// Vanity returns custom Vanity name servers for an account
func (dme *GoDMEConfig) Vanity() ([]Vanity, error) {
	req, err := dme.newRequest("GET", "dns/vanity", nil)
	if err != nil {
		return nil, err
	}

	vanityResponse := &GenericResponse{}
	err = dme.doDMERequest(req, &vanityResponse)
	if err != nil {
		return nil, err
	}

	vanityData := []Vanity{}
	json.Unmarshal(vanityResponse.Data, &vanityData)

	return vanityData, nil
}

// IPSets returns custom IPSets for an account, used for secondary DNS
func (dme *GoDMEConfig) IPSets() ([]IPSet, error) {
	req, err := dme.newRequest("GET", "dns/secondary/ipSet", nil)
	if err != nil {
		return nil, err
	}

	genericResponse := &GenericResponse{}
	err = dme.doDMERequest(req, &genericResponse)
	if err != nil {
		return nil, err
	}

	ipSetData := []IPSet{}
	json.Unmarshal(genericResponse.Data, &ipSetData)

	return ipSetData, nil
}

// SecondaryDomains returns the list of secondary domains belonging to an account
func (dme *GoDMEConfig) SecondaryDomains() ([]SecondaryDomain, error) {
	req, err := dme.newRequest("GET", "dns/secondary", nil)
	if err != nil {
		return nil, err
	}

	genericResponse := &GenericResponse{}
	err = dme.doDMERequest(req, &genericResponse)
	if err != nil {
		return nil, err
	}

	secondaryDomains := []SecondaryDomain{}
	json.Unmarshal(genericResponse.Data, &secondaryDomains)

	return secondaryDomains, nil
}

// Folders returns the list of folders belonging to an account
func (dme *GoDMEConfig) Folders() ([]Folder, error) {
	req, err := dme.newRequest("GET", "security/folder", nil)
	if err != nil {
		return nil, err
	}

	folderList := []Folder{}
	err = dme.doDMERequest(req, &folderList)
	if err != nil {
		return nil, err
	}

	return folderList, nil
}

// AddRecord adds a DNS record to a given domain (identified by its ID)
func (dme *GoDMEConfig) AddRecord(DomainID int, RecordRecord *Record) (*Record, error) {
	reqStub := fmt.Sprintf("dns/managed/%v/records", DomainID)
	bodyData, err := json.Marshal(RecordRecord)
	if err != nil {
		return nil, err
	}
	bodyBuffer := bytes.NewReader(bodyData)
	req, err := dme.newRequest("POST", reqStub, bodyBuffer)
	if err != nil {
		return nil, err
	}

	returnedRecord := &Record{}
	err = dme.doDMERequest(req, returnedRecord)
	if err != nil {
		return nil, err
	}

	return returnedRecord, err
}

// AddDomain adds a domain to your DNS Made Easy account
func (dme *GoDMEConfig) AddDomain(DomainRecord *Domain) (*Domain, error) {
	reqStub := "dns/managed/"
	bodyData, err := json.Marshal(DomainRecord)
	if err != nil {
		return nil, err
	}
	bodyBuffer := bytes.NewReader(bodyData)
	req, err := dme.newRequest("POST", reqStub, bodyBuffer)
	if err != nil {

		return nil, err
	}

	returnedDomain := &Domain{}
	err = dme.doDMERequest(req, returnedDomain)
	if err != nil {
		return nil, err
	}

	return returnedDomain, err
}

// AddVanity creates a custom set of Vanity nameservers for an account. These can then be assigned to domains.
func (dme *GoDMEConfig) AddVanity(newVanity Vanity) (*Vanity, error) {
	bodyData, err := json.Marshal(newVanity)
	if err != nil {
		return nil, err
	}
	bodyBuffer := bytes.NewReader(bodyData)
	req, err := dme.newRequest("POST", "dns/vanity", bodyBuffer)

	returnedVanity := &Vanity{}
	err = dme.doDMERequest(req, returnedVanity)
	if err != nil {
		return nil, err
	}
	return returnedVanity, err
}

// AddSOA creates a custom SOA record for an account. These can then be assigned to domains.
func (dme *GoDMEConfig) AddSOA(newSOA SOA) (*SOA, error) {
	bodyData, err := json.Marshal(newSOA)
	if err != nil {
		return nil, err
	}
	bodyBuffer := bytes.NewReader(bodyData)
	req, err := dme.newRequest("POST", "dns/soa", bodyBuffer)

	returnedSOA := &SOA{}
	err = dme.doDMERequest(req, returnedSOA)
	if err != nil {
		return nil, err
	}
	return returnedSOA, err
}

// AddIPSet creates a custom IPSet record for an account. These can then be assigned to secondary domains.
func (dme *GoDMEConfig) AddIPSet(newIPSet IPSet) (*IPSet, error) {
	bodyData, err := json.Marshal(newIPSet)
	if err != nil {
		return nil, err
	}
	bodyBuffer := bytes.NewReader(bodyData)
	req, err := dme.newRequest("POST", "dns/secondary/ipSet", bodyBuffer)

	returnedIPSet := &IPSet{}
	err = dme.doDMERequest(req, returnedIPSet)
	if err != nil {
		return nil, err
	}
	return returnedIPSet, err
}

// AddSecondaryDomain adds a secondary domain to your account
func (dme *GoDMEConfig) AddSecondaryDomain(newSecondaryDomain SecondaryDomain) (*SecondaryDomain, error) {
	bodyData, err := json.Marshal(newSecondaryDomain)
	if err != nil {
		return nil, err
	}
	bodyBuffer := bytes.NewReader(bodyData)
	req, err := dme.newRequest("POST", "dns/secondary", bodyBuffer)

	returnedSecondaryDomain := &SecondaryDomain{}
	err = dme.doDMERequest(req, returnedSecondaryDomain)
	if err != nil {
		return nil, err
	}
	return returnedSecondaryDomain, err
}

// UpdateRecord updates an existing DNS record (identified by its ID) in a given domain. DNS Made Easy only returns success/fail for this method.
func (dme *GoDMEConfig) UpdateRecord(DomainID int, Record *Record) error {
	reqStub := fmt.Sprintf("dns/managed/%v/records/%v", DomainID, Record.ID)
	bodyData, err := json.Marshal(Record)
	if err != nil {
		return err
	}
	return dme.genericUpdate(reqStub, bodyData)
}

// UpdateVanity updates an existing Vanity DNS Template (identified by its ID) for your account. DNS Made Easy only returns success/fail for this method.
func (dme *GoDMEConfig) UpdateVanity(Vanity *Vanity) error {
	reqStub := fmt.Sprintf("dns/vanity/%v", Vanity.ID)
	bodyData, err := json.Marshal(Vanity)
	if err != nil {
		return err
	}
	return dme.genericUpdate(reqStub, bodyData)
}

// UpdateDomain updates an existing Domain (identified by its ID) for your account. DNS Made Easy only returns success/fail for this method.
func (dme *GoDMEConfig) UpdateDomain(Domain *Domain) error {
	reqStub := fmt.Sprintf("dns/managed/%v", Domain.ID)
	bodyData, err := json.Marshal(Domain)
	if err != nil {
		return err
	}
	return dme.genericUpdate(reqStub, bodyData)
}

// UpdateSOA updates an existing Domain (identified by its ID) for your account. DNS Made Easy only returns success/fail for this method.
func (dme *GoDMEConfig) UpdateSOA(SOA *SOA) error {
	reqStub := fmt.Sprintf("dns/soa/%v", SOA.ID)
	bodyData, err := json.Marshal(SOA)
	if err != nil {
		return err
	}
	return dme.genericUpdate(reqStub, bodyData)
}

// UpdateIPSet updates an existing IPSet (identified by its ID) for your account. DNS Made Easy only returns success/fail for this method.
func (dme *GoDMEConfig) UpdateIPSet(IPSet *IPSet) error {
	reqStub := fmt.Sprintf("dns/secondary/ipSet/%v", IPSet.ID)
	bodyData, err := json.Marshal(IPSet)
	if err != nil {
		return err
	}
	return dme.genericUpdate(reqStub, bodyData)
}

// UpdateSecondaryDomain updates an existing secondary domain (identified by its ID) for your account. DNS Made Easy only returns success/fail for this method.
func (dme *GoDMEConfig) UpdateSecondaryDomain(SecondaryDomain *SecondaryDomain) error {
	reqStub := fmt.Sprintf("dns/secondary/%v", SecondaryDomain.ID)
	bodyData, err := json.Marshal(SecondaryDomain)
	if err != nil {
		return err
	}
	return dme.genericUpdate(reqStub, bodyData)
}

// All of the PUT updates are basically the same, so we can make a fairly generic wrapper
func (dme *GoDMEConfig) genericUpdate(Endpoint string, BodyData []byte) error {
	bodyBuffer := bytes.NewReader(BodyData)
	req, err := dme.newRequest("PUT", Endpoint, bodyBuffer)
	if err != nil {
		return err
	}
	return dme.doDMERequest(req, nil)
}

// DeleteRecord deletes an existing DNS record (identified by its ID) in a given domain
func (dme *GoDMEConfig) DeleteRecord(DomainID, RecordID int) error {
	reqStub := fmt.Sprintf("dns/managed/%v/records/%v", DomainID, RecordID)
	req, err := dme.newRequest("DELETE", reqStub, nil)
	if err != nil {
		return err
	}
	return dme.doDMERequest(req, nil)
}

// DeleteRecords deletes a DNS record (identified by their IDs) in a given domain
func (dme *GoDMEConfig) DeleteRecords(DomainID int, RecordIDs []int) error {
	var queryString string
	for _, record := range RecordIDs {
		queryString = fmt.Sprintf("%sids=%v&", queryString, record)
	}
	reqStub := fmt.Sprintf("dns/managed/%v/records?%s", DomainID, queryString)
	req, err := dme.newRequest("DELETE", reqStub, nil)
	if err != nil {
		return err
	}
	return dme.doDMERequest(req, nil)
}

// DeleteDomain deletes a domain from your DNS Made Easy account. The DeleteTimeout argument indicates how long we should keep trying to
// delete the domain if DNS Made Easy says it can't delete the domain due to a pending operation. In these cases, usually deleting a domain
// name will succeed after a certain period of time. You may not want to wait for this time though, so specify 0 here to never retry.
func (dme *GoDMEConfig) DeleteDomain(DomainID int, DeleteTimeout time.Duration) error {
	return dme.genericDelete(fmt.Sprintf("dns/managed/%v", DomainID), DeleteTimeout)
}

// DeleteSOA deletes an existing SOA record (identified by its ID). The SOA must not be in use before deleting.
func (dme *GoDMEConfig) DeleteSOA(SoaID int) error {
	return dme.genericDelete(fmt.Sprintf("dns/soa/%v", SoaID), 0)
}

// DeleteVanity deletes an existing Vanity record (identified by its ID). The Vanity configuration must not be in use before deleting.
func (dme *GoDMEConfig) DeleteVanity(VanityID int) error {
	return dme.genericDelete(fmt.Sprintf("dns/vanity/%v", VanityID), 0)
}

// DeleteIPSet deletes an existing IPSet (identified by its ID). The IPSet must not be in use before deleting.
func (dme *GoDMEConfig) DeleteIPSet(IPsetID int) error {
	return dme.genericDelete(fmt.Sprintf("dns/secondary/ipSet/%v", IPsetID), 0)
}

// DeleteSecondaryDomain deletes a secondary domain from your DNS Made Easy account. The DeleteTimeout argument indicates how long we should keep trying to
// delete the domain if DNS Made Easy says it can't delete the domain due to a pending operation. In these cases, usually deleting a domain
// name will succeed after a certain period of time. You may not want to wait for this time though, so specify 0 here to never retry.
func (dme *GoDMEConfig) DeleteSecondaryDomain(SecondaryDomainID int, DeleteTimeout time.Duration) error {
	return dme.genericDelete(fmt.Sprintf("dns/secondary/%v", SecondaryDomainID), DeleteTimeout)
}

//All deletes are the same, but a different API endpoint, and some need a timeout.
func (dme *GoDMEConfig) genericDelete(Endpoint string, DeleteTimeout time.Duration) error {
	timeOutAt := time.Now().Add(DeleteTimeout)

	req, err := dme.newRequest("DELETE", Endpoint, nil)
	if err != nil {
		return err
	}
	//Try to delete once
	deleteError := dme.doDMERequest(req, nil)
	if deleteError == nil || DeleteTimeout == 0 {
		return deleteError
	}

	//If we were unsuccessful in deleting the first time, try try again until the timeout
	for time.Now().Before(timeOutAt) {
		req, _ := dme.newRequest("DELETE", Endpoint, nil)
		deleteError := dme.doDMERequest(req, nil)
		//No error? Then we're all done.
		if deleteError == nil {
			return deleteError
		}
		//We got a different error this time that is not a pending delete error
		if deleteError.Error() != pendingDeleteError {
			return deleteError
		}
		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("Could not delete after %s (%s)", DeleteTimeout.String(), err)

}

// ExportAllDomains returns a map with every domain that DNS Made Easy manages, along with its properties
func (dme *GoDMEConfig) ExportAllDomains() (*AllDomainExport, error) {
	allDomains, err := dme.Domains()
	if err != nil {
		return nil, err
	}
	allSOA, err := dme.SOA()
	if err != nil {
		return nil, err
	}
	allVanity, err := dme.Vanity()
	if err != nil {
		return nil, err
	}

	thisExport := make(AllDomainExport)

	for _, domain := range allDomains {
		var thisSOA *SOA
		var thisVanity *Vanity

		//Find the correct SOA record
		for _, s := range allSOA {
			if s.ID == domain.SoaID {
				thisSOA = &s
			}
		}

		//Find the correct NS records
		for _, v := range allVanity {
			if v.ID == domain.VanityID {
				thisVanity = &v
			}
		}

		//Get DNS records
		thisRecords, err := dme.Records(domain.ID)
		if err != nil {
			return nil, err
		}

		thisExport[domain.Name] = DomainExport{
			Info:      &domain,
			SOA:       thisSOA,
			DefaultNS: thisVanity,
			Records:   &thisRecords,
		}
	}

	return &thisExport, nil
}
