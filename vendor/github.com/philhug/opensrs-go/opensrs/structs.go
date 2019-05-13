package opensrs

type NameserverList []struct {
	Name      string `json:"name"`
	IpAddress string `json:"ipaddress,omitempty"`
	Ipv6      string `json:"ipv6,omitempty"`
	SortOrder string `json:"sortorder,omitempty"`
}

type ARecord struct {
	IpAddress string `json:"ipaddress,omitempty"`
	SubDomain string `json:"subdomain,omitempty"`
}

type AAAARecord struct {
	Ipv6Address string `json:"ipv6_address,omitempty"`
	SubDomain   string `json:"subdomain,omitempty"`
}

type CNAMERecord struct {
	HostName  string `json:"hostname,omitempty"`
	SubDomain string `json:"subdomain,omitempty"`
}

type MXRecord struct {
	Priority  string `json:"priority,omitempty"`
	SubDomain string `json:"subdomain,omitempty"`
	HostName  string `json:"hostname,omitempty"`
}

type SRVRecord struct {
	Priority  string `json:"priority,omitempty"`
	Weight    string `json:"weight,omitempty"`
	SubDomain string `json:"subdomain,omitempty"`
	HostName  string `json:"hostname,omitempty"`
	Port      string `json:"port,omitempty"`
}

type TXTRecord struct {
	SubDomain string `json:"subdomain,omitempty"`
	Text      string `json:"text,omitempty"`
}

type DnsRecords struct {
	A     []ARecord     `json:"A,omitempty"`
	AAAA  []AAAARecord  `json:"AAAA,omitempty"`
	CNAME []CNAMERecord `json:"CNAME,omitempty"`
	MX    []MXRecord    `json:"MX,omitempty"`
	SRV   []SRVRecord   `json:"SRV,omitempty"`
	TXT   []TXTRecord   `json:"TXT,omitempty"`
}

func (n NameserverList) ToString() []string {
	domains := make([]string, len(n))
	for i, ns := range n {
		domains[i] = ns.Name
	}
	return domains
}

type OpsRequestAttributes struct {
	Domain         string         `json:"domain"`
	Limit          string         `json:"limit,omitempty"`
	Type           string         `json:"type,omitempty"`
	Data           string         `json:"data,omitempty"`
	AffectDomains  string         `json:"affect_domains,omitempty"`
	NameserverList NameserverList `json:"nameserver_list,omitempty"`
	OpType         string         `json:"op_type,omitempty"`
	AssignNs       []string       `json:"assign_ns,omitempty"`
}

type OpsResponse struct {
	Action       string `json:"action"`
	Object       string `json:"object"`
	Protocol     string `json:"protocol"`
	IsSuccess    string `json:"is_success"`
	ResponseCode string `json:"response_code"`
	ResponseText string `json:"response_text"`
	Attributes   struct {
		NameserverList NameserverList `json:"nameserver_list,omitempty"`
		Type           string         `json:"type,omitempty"`
		LockState      string         `json:"lock_state,omitempty"`
		Records        DnsRecords     `json:"records,omitempty"`
	} `json:"attributes"`
}

type OpsRequest struct {
	Action     string               `json:"action"`
	Object     string               `json:"object"`
	Protocol   string               `json:"protocol"`
	Attributes OpsRequestAttributes `json:"attributes"`
}
