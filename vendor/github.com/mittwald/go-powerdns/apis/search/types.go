package search

import "fmt"

// ObjectType represents the object type for which a search should be performed
type ObjectType int

// Possible object types; according to the PowerDNS documentation, this list is
// exhaustive.
const (
	_                        = iota
	ObjectTypeAll ObjectType = iota
	ObjectTypeZone
	ObjectTypeRecord
	ObjectTypeComment
)

// String makes this type implement fmt.Stringer
func (t ObjectType) String() string {
	switch t {
	case ObjectTypeAll:
		return "all"
	case ObjectTypeZone:
		return "zone"
	case ObjectTypeRecord:
		return "record"
	case ObjectTypeComment:
		return "comment"
	}

	return ""
}

// UnmarshalJSON makes this type implement json.Unmarshaler
func (t *ObjectType) UnmarshalJSON(b []byte) error {
	switch string(b) {
	case `"all"`:
		*t = ObjectTypeAll
	case `"zone"`:
		*t = ObjectTypeZone
	case `"record"`:
		*t = ObjectTypeRecord
	case `"comment"`:
		*t = ObjectTypeComment
	default:
		return fmt.Errorf(`unknown search type: %s'`, string(b))
	}

	return nil
}

// Result represents a single search result. See the documentation for more
// information: https://doc.powerdns.com/authoritative/http-api/search.html#searchresult
type Result struct {
	Content    string     `json:"content"`
	Disabled   bool       `json:"disabled"`
	Name       string     `json:"name"`
	ObjectType ObjectType `json:"object_type"`
	ZoneID     string     `json:"zone_id"`
	Zone       string     `json:"zone"`
	Type       string     `json:"type"`
	TTL        int        `json:"ttl"`
}
