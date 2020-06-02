package zones

import "encoding/json"

// ZoneNameservers is a special list type to represent the nameservers of a zone.
// When nil, this type will still serialize to an empty JSON list.
// See https://github.com/mittwald/go-powerdns/issues/4 for more information
type ZoneNameservers []string

// MarshalJSON implements the `json.Marshaler` interface
func (z ZoneNameservers) MarshalJSON() ([]byte, error) {
	if z == nil {
		return []byte("[]"), nil
	}

	return json.Marshal([]string(z))
}

type Zone struct {
	ID                 string              `json:"id,omitempty"`
	Name               string              `json:"name"`
	Type               ZoneType            `json:"type"`
	URL                string              `json:"url,omitempty"`
	Kind               ZoneKind            `json:"kind,omitempty"`
	ResourceRecordSets []ResourceRecordSet `json:"rrsets,omitempty"`
	Serial             int                 `json:"serial,omitempty"`
	NotifiedSerial     int                 `json:"notified_serial,omitempty"`
	Masters            []string            `json:"masters,omitempty"`
	DNSSec             bool                `json:"dnssec,omitempty"`
	NSec3Param         string              `json:"nsec3param,omitempty"`
	NSec3Narrow        bool                `json:"nsec3narrow,omitempty"`
	Presigned          bool                `json:"presigned,omitempty"`
	SOAEdit            string              `json:"soa_edit,omitempty"`
	SOAEditAPI         string              `json:"soa_edit_api,omitempty"`
	APIRectify         bool                `json:"api_rectify,omitempty"`
	Zone               string              `json:"zone,omitempty"`
	Account            string              `json:"account,omitempty"`
	Nameservers        ZoneNameservers     `json:"nameservers"`
	TSIGMasterKeyIDs   []string            `json:"tsig_master_key_ids,omitempty"`
	TSIGSlaveKeyIDs    []string            `json:"tsig_slave_key_ids,omitempty"`
}

func (z *Zone) GetRecordSet(name, recordType string) *ResourceRecordSet {
	for i := range z.ResourceRecordSets {
		if z.ResourceRecordSets[i].Name == name && z.ResourceRecordSets[i].Type == recordType {
			return &z.ResourceRecordSets[i]
		}
	}

	return nil
}
