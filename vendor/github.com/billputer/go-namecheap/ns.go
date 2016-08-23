package namecheap

import "net/url"

const (
	nsCreate  = "namecheap.domains.ns.create"
	nsDelete  = "namecheap.domains.ns.delete"
	nsGetInfo = "namecheap.domains.ns.getInfo"
	nsUpdate  = "namecheap.domains.ns.update"
)

type DomainNSInfoResult struct {
	Domain     string   `xml:"Domain,attr"`
	Nameserver string   `xml:"Nameserver,attr"`
	IP         string   `xml:"IP,attr"`
	Statuses   []string `xml:"NameserverStatuses>Status"`
}

func (client *Client) NSGetInfo(sld, tld, nameserver string) (*DomainNSInfoResult, error) {
	requestInfo := &ApiRequest{
		command: nsGetInfo,
		method:  "GET",
		params:  url.Values{},
	}
	requestInfo.params.Set("SLD", sld)
	requestInfo.params.Set("TLD", tld)
	requestInfo.params.Set("Nameserver", nameserver)

	resp, err := client.do(requestInfo)
	if err != nil {
		return nil, err
	}

	return resp.DomainNSInfo, nil
}
