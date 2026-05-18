package rtypecontrol

import (
	"fmt"

	dnsutilv2 "codeberg.org/miekg/dns/dnsutil"
	dnsrdatav2 "codeberg.org/miekg/dns/rdata"
	"github.com/DNSControl/dnscontrol/v4/models"
)

// FixLegacyDC populates .F to compenstate for providers that have not been
// updated to support RecordConfigV2 when creating RecordConfig.
// It is called anywhere dc.PostProcess() or models.PostProcessRecords() is
// called.  Those functions can't call it directly because that would cause an
// import cycle.
func FixLegacyDC(dc *models.DomainConfig) {
	FixLegacyRecords(&dc.Records)
}

// FixLegacyRecords populates .F to compenstate for providers that have not been
// updated to support RecordConfigV2 when creating RecordConfig.
// It is called anywhere provider.GetZoneRecords() is called. GetZoneRecords()
// can't call it directly because that would involve modifying every provider.
// Instead, providers should be fixed to generate records properly.
func FixLegacyRecords(recs *models.Records) {
	for _, rec := range *recs {
		FixLegacyRecord(rec)
	}
}

// FixLegacyRecord populates .F to compenstate for providers that have not been
// updated to support RecordConfigV2 when creating RecordConfig.
func FixLegacyRecord(rec *models.RecordConfig) {
	// Populate .F if needed: (legacy)
	// That is... If rec.F == nil and this is a "modern" type.
	if rec.F == nil {
		if fixer, ok := Func[rec.Type]; ok {
			fixer.CopyFromLegacyFields(rec)
		}
	}

	// Populate .RDATA if needed:
	if rec.RDATA == nil {

		// The .RDATA structure itself.
		switch rec.Type {
		case "A":
			rec.RDATA = dnsrdatav2.A{Addr: rec.GetTargetIP()}
		case "AAAA":
			rec.RDATA = dnsrdatav2.AAAA{Addr: rec.GetTargetIP()}

		case "CAA":
			rec.RDATA = dnsrdatav2.CAA{Flag: rec.CaaFlag, Tag: rec.CaaTag, Value: rec.GetTargetField()}
		case "CNAME":
			rec.RDATA = dnsrdatav2.CNAME{Target: rec.GetTargetField()}

		case "HTTPS":
			// no-op.  See pkg/rtype/t_svcb.go:SetTargetSVCB
			panic("HTTPS should already be converted to RDATA")

		case "MX":
			rec.RDATA = dnsrdatav2.MX{Preference: rec.MxPreference, Mx: rec.GetTargetField()}

		case "RP":
			// no-op.  See pkg/rtype/rp.go:FromStruct.
			panic("RP should already be converted to RDATA")

		case "SOA":
			rec.RDATA = dnsrdatav2.SOA{Ns: rec.GetTargetField(), Mbox: rec.SoaMbox, Serial: rec.SoaSerial, Refresh: rec.SoaRefresh, Retry: rec.SoaRetry, Expire: rec.SoaExpire, Minttl: rec.SoaMinttl}
		case "SRV":
			rec.RDATA = dnsrdatav2.SRV{Priority: rec.SrvPriority, Weight: rec.SrvWeight, Port: rec.SrvPort, Target: rec.GetTargetField()}

		case "SVCB":
			// no-op.  See pkg/rtype/t_svcb.go:SetTargetSVCB
			panic("SVCB should already be converted to RDATA")

		case "TXT":
			rec.RDATA = dnsrdatav2.TXT{Txt: []string{rec.GetTargetField()}}

		default:
			panic(fmt.Sprintf("RDATA CONVERSION NOT IMPLEMENTED TYPE=%q", rec.Type))
		}

		if rec.RDATA != nil {

			// TypeNum:
			tn, err := dnsutilv2.StringToType(rec.Type)
			if err != nil {
				panic("fix me")
			}
			rec.TypeNum = tn

			// Comparable:
			rec.Comparable = fmt.Sprintf("%s", rec.RDATA)

		}

	}
}
