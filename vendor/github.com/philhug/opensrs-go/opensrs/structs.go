package opensrs

type NameserverList []struct {
	Name      string `json:"name"`
	IpAddress string `json:"ipaddress,omitempty"`
	Ipv6      string `json:"ipv6,omitempty"`
	SortOrder string `json:"sortorder,omitempty"`
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
	} `json:"attributes"`
}

type OpsRequest struct {
	Action     string               `json:"action"`
	Object     string               `json:"object"`
	Protocol   string               `json:"protocol"`
	Attributes OpsRequestAttributes `json:"attributes"`
}
