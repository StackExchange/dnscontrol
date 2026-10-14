package misaka

import (
	"testing"

	"github.com/DNSControl/dnscontrol/v5/models"
)

func TestAuditRecordsTTL(t *testing.T) {
	dc := models.MustNewDomainConfig(testOrigin)
	for _, tt := range []struct {
		ttl  uint32
		want int
	}{
		{ttl: 0, want: 0},
		{ttl: maxTTL, want: 0},
		{ttl: maxTTL + 1, want: 1},
		{ttl: 4294967295, want: 1},
	} {
		rc, err := dc.NewRecordConfigParse("big", tt.ttl, "A", "192.0.2.1")
		if err != nil {
			t.Fatal(err)
		}
		if errs := (&misakaProvider{}).AuditRecords(models.Records{rc}); len(errs) != tt.want {
			t.Errorf("AuditRecords(ttl=%d) = %v, want %d errors", tt.ttl, errs, tt.want)
		}
	}
}
