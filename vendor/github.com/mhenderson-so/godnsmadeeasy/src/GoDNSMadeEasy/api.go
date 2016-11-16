package GoDNSMadeEasy

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Domains returns the list of domains managed by DNS Made Easy
func (dme *GoDNSMadeEasy) Domains() ([]Domain, error) {
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
func (dme *GoDNSMadeEasy) Domain(DomainID int) (*Domain, error) {
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
func (dme *GoDNSMadeEasy) Records(DomainID int) ([]Record, error) {
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
func (dme *GoDNSMadeEasy) Record(DomainID, RecordID int) (*Record, error) {
	return nil, fmt.Errorf("Record() Not yet implemented")
}

// SOA returns custom Start of Authority records for an account.
func (dme *GoDNSMadeEasy) SOA() ([]SOA, error) {
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
func (dme *GoDNSMadeEasy) Vanity() ([]Vanity, error) {
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

// ExportAllDomains returns a map with every domain that DNS Made Easy manages, along with its properties
func (dme *GoDNSMadeEasy) ExportAllDomains() (*AllDomainExport, error) {
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

// AddRecord adds a DNS record to a given domain (identified by its ID)
func (dme *GoDNSMadeEasy) AddRecord(DomainID int, RecordRecord *Record) (*Record, error) {
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

// UpdateRecord updates an existing DNS record (identified by its ID) in a given domain
func (dme *GoDNSMadeEasy) UpdateRecord(DomainID int, Record *Record) (*Record, error) {
	reqStub := fmt.Sprintf("dns/managed/%v/records/%v", DomainID, Record.ID)
	bodyData, err := json.Marshal(Record)
	if err != nil {
		return nil, err
	}
	bodyBuffer := bytes.NewReader(bodyData)
	req, err := dme.newRequest("PUT", reqStub, bodyBuffer)
	if err != nil {
		return nil, err
	}

	returnedRecord := &Record
	err = dme.doDMERequest(req, returnedRecord)
	if err != nil {
		return nil, err
	}

	return Record, err
}

// DeleteRecord deletes an existing DNS record (identified by its ID) in a given domain
func (dme *GoDNSMadeEasy) DeleteRecord(DomainID, RecordID int) error {
	reqStub := fmt.Sprintf("dns/managed/%v/records/%v", DomainID, RecordID)
	req, err := dme.newRequest("DELETE", reqStub, nil)
	if err != nil {
		return err
	}

	err = dme.doDMERequest(req, nil)
	if err != nil {
		return err
	}

	return nil
}

// AddDomain adds a domain to your DNS Made Easy account
func (dme *GoDNSMadeEasy) AddDomain(DomainRecord *Domain) (*Domain, error) {
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
