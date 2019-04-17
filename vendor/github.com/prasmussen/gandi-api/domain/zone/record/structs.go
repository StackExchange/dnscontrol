package record

type RecordInfo struct {
	Id    string
	Name  string
	Ttl   int64
	Type  string
	Value string
}

type RecordAdd struct {
	Zone    int64  `goptions:"-z, --zone, obligatory, description='Zone id'"`
	Version int64  `goptions:"-v, --version, obligatory, description='Zone version'"`
	Name    string `goptions:"-n, --name, obligatory, description='Record name. Relative name, may contain leading wildcard. @ for empty name'"`
	Type    string `goptions:"-t, --type, obligatory, description='Record type'"`
	Value   string `goptions:"-V, --value, obligatory, description='Value for record. Semantics depends on the record type.'"`
	Ttl     int64  `goptions:"-T, --ttl, description='Time to live, in seconds, between 5 minutes and 30 days'"`
}

type RecordUpdate struct {
	Zone    int64  `goptions:"-z, --zone, obligatory, description='Zone id'"`
	Version int64  `goptions:"-v, --version, obligatory, description='Zone version'"`
	Name    string `goptions:"-n, --name, obligatory, description='Record name. Relative name, may contain leading wildcard. @ for empty name'"`
	Type    string `goptions:"-t, --type, obligatory, description='Record type'"`
	Value   string `goptions:"-V, --value, obligatory, description='Value for record. Semantics depends on the record type.'"`
	Ttl     int64  `goptions:"-T, --ttl, description='Time to live, in seconds, between 5 minutes and 30 days'"`
	Id      string `goptions:"-r, --record, obligatory, description='Record id'"`
}

type RecordSet map[string]interface{}
