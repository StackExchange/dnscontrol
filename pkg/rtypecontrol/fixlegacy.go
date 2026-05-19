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

		case "DHCID":
			rec.RDATA = dnsrdatav2.DHCID{Digest: rec.GetTargetField()}
		case "DNAME":
			rec.RDATA = dnsrdatav2.DNAME{Target: rec.GetTargetField()}
		case "DNSKEY":
			rec.RDATA = dnsrdatav2.DNSKEY{Flags: rec.DnskeyFlags, Protocol: rec.DnskeyProtocol, Algorithm: rec.DnskeyAlgorithm, PublicKey: rec.GetTargetField()}
		case "DS":
			// no-op.  See pkg/rtype/ds.go:FromStruct.
			panic("DS should already be converted to RDATA")

		case "HTTPS":
			// no-op.  See pkg/rtype/t_svcb.go:SetTargetSVCB
			panic("HTTPS should already be converted to RDATA")

		case "LOC":
			rec.RDATA = dnsrdatav2.LOC{Version: rec.LocVersion, Size: rec.LocSize, HorizPre: rec.LocHorizPre, VertPre: rec.LocVertPre, Latitude: rec.LocLatitude, Longitude: rec.LocLongitude, Altitude: rec.LocAltitude}

		case "MX":
			rec.RDATA = dnsrdatav2.MX{Preference: rec.MxPreference, Mx: rec.GetTargetField()}

		case "NS":
			rec.RDATA = dnsrdatav2.NS{Ns: rec.GetTargetField()}
		case "NAPTR":
			rec.RDATA = dnsrdatav2.NAPTR{Order: rec.NaptrOrder, Preference: rec.NaptrPreference, Flags: rec.NaptrFlags, Service: rec.NaptrService, Regexp: rec.NaptrRegexp, Replacement: rec.GetTargetField()}

		case "OPENPGPKEY":
			rec.RDATA = dnsrdatav2.OPENPGPKEY{PublicKey: rec.GetTargetField()}

		case "PTR":
			rec.RDATA = dnsrdatav2.PTR{Ptr: rec.GetTargetField()}

		case "RP":
			// no-op.  See pkg/rtype/rp.go:FromStruct.
			panic("RP should already be converted to RDATA")

		case "SMIMEA":
			rec.RDATA = dnsrdatav2.SMIMEA{Usage: rec.SmimeaUsage, Selector: rec.SmimeaSelector, MatchingType: rec.SmimeaMatchingType, Certificate: rec.GetTargetField()}
		case "SOA":
			rec.RDATA = dnsrdatav2.SOA{Ns: rec.GetTargetField(), Mbox: rec.SoaMbox, Serial: rec.SoaSerial, Refresh: rec.SoaRefresh, Retry: rec.SoaRetry, Expire: rec.SoaExpire, Minttl: rec.SoaMinttl}
		case "SRV":
			rec.RDATA = dnsrdatav2.SRV{Priority: rec.SrvPriority, Weight: rec.SrvWeight, Port: rec.SrvPort, Target: rec.GetTargetField()}
		case "SSHFP":
			rec.RDATA = dnsrdatav2.SSHFP{Algorithm: rec.SshfpAlgorithm, Type: rec.SshfpFingerprint, FingerPrint: rec.GetTargetField()}

		case "TLSA":
			rec.RDATA = dnsrdatav2.TLSA{Usage: rec.TlsaUsage, Selector: rec.TlsaSelector, MatchingType: rec.TlsaMatchingType, Certificate: rec.GetTargetField()}

		case "SVCB":
			// no-op.  See pkg/rtype/t_svcb.go:SetTargetSVCB
			panic("SVCB should already be converted to RDATA")

		case "TXT":
			rec.RDATA = dnsrdatav2.TXT{Txt: []string{rec.GetTargetField()}}

		default:
			panic(fmt.Sprintf("RDATA CONVERSION NOT IMPLEMENTED TYPE=%q", rec.Type))
		}

		// TypeNum:
		tn, err := dnsutilv2.StringToType(rec.Type)
		if err != nil {
			panic("fix me")
		}
		rec.TypeNum = tn

		// Comparable:
		rec.ComparableV3 = rec.RDATA.String()
		fmt.Printf("DEBUG: COMPARE for %s --- %s\n", rec.Type, rec.ComparableV3)

	}
}
