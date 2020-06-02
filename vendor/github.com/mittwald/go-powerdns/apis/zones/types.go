package zones

type ResourceRecordSet struct {
	Name       string              `json:"name"`
	Type       string              `json:"type"`
	TTL        int                 `json:"ttl"`
	ChangeType RecordSetChangeType `json:"changetype,omitempty"`
	Records    []Record            `json:"records"`
	Comments   []Comment           `json:"comments"`
}

type Record struct {
	Content  string `json:"content"`
	Disabled bool   `json:"disabled"`
	SetPTR   bool   `json:"set-ptr,omitempty"`
}

type Comment struct {
	Content    string `json:"content"`
	Account    string `json:"account"`
	ModifiedAt int    `json:"modified_at"`
}
