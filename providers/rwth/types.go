package rwth

import (
	"github.com/miekg/dns"
	"time"
)

type RecordReply struct {
	ID        int       `json:"id"`
	ZoneID    int       `json:"zone_id"`
	Type      string    `json:"type"`
	Content   string    `json:"content"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
	Editable  bool      `json:"editable"`
	rec       dns.RR    // Store miekg/dns
}

type zone struct {
	ID         int       `json:"id"`
	ZoneName   string    `json:"zone_name"`
	Status     string    `json:"status"`
	UpdatedAt  time.Time `json:"updated_at"`
	LastDeploy time.Time `json:"last_deploy"`
	Dnssec     struct {
		ZoneSigningKey struct {
			CreatedAt time.Time `json:"created_at"`
		} `json:"zone_signing_key"`
		KeySigningKey struct {
			CreatedAt time.Time `json:"created_at"`
		} `json:"key_signing_key"`
	} `json:"dnssec"`
}
