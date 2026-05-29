package models

import (
	"fmt"
	"strings"

	dnsutilv2 "codeberg.org/miekg/dns/dnsutil"
	dnsrdatav2 "codeberg.org/miekg/dns/rdata"
	privatetypesrdata "github.com/DNSControl/dnscontrol/v4/pkg/privatetypes/rdata"
	dnsv1 "github.com/miekg/dns"
	dnsutilv1 "github.com/miekg/dns/dnsutil"

	_ "github.com/DNSControl/dnscontrol/v4/pkg/privatetypes"
)

// FixUp populates the "V3 Fields": .TypeNum, .RDATA and .ComparableV3.
func (rc *RecordConfig) FixUp(origin string) {

	// TypeNum:
	if rc.TypeNum == 0 && rc.Type != "ALIAS" {
		tn, err := dnsutilv2.StringToType(rc.Type)
		if err != nil {
			panic(fmt.Sprintf("BUG: Unknown type %s", rc.Type))
		}
		rc.TypeNum = tn
	}

	// Populate .RDATA if needed:
	if rc.RDATA == nil {

		switch rc.Type {

		// Incomplete
		case "MIKROTIK_FWD":
			rc.RDATA = privatetypesrdata.MIKROTIK_FWD{}
		case "MIKROTIK_NXDOMAIN":
			rc.RDATA = privatetypesrdata.MIKROTIK_NXDOMAIN{}
		case "PORKBUN_URLFWD":
			rc.RDATA = privatetypesrdata.PORKBUN_URLFWD{}
		case "URL":
			rc.RDATA = privatetypesrdata.URL{}
		case "URL301":
			rc.RDATA = privatetypesrdata.URL301{}
		case "FRAME":
			rc.RDATA = privatetypesrdata.FRAME{}
		case "BUNNY_DNS_PZ":
			rc.RDATA = privatetypesrdata.BUNNY_DNS_PZ{}
		case "LUA":
			rc.RDATA = privatetypesrdata.LUA{}
		case "CLOUDNS_WR":
			rc.RDATA = privatetypesrdata.CLOUDNS_WR{}
		case "NETLIFY":
			rc.RDATA = privatetypesrdata.NETLIFY{}
		case "NETLIFYV6":
			rc.RDATA = privatetypesrdata.NETLIFYV6{}
		case "AKAMAICDN":
			rc.RDATA = privatetypesrdata.AKAMAICDN{}
		case "AKAMAITLC":
			rc.RDATA = privatetypesrdata.AKAMAITLC{}
		case "BUNNY_DNS_RDR":
			rc.RDATA = privatetypesrdata.BUNNY_DNS_RDR{}

		case "A":
			rc.RDATA = dnsrdatav2.A{Addr: rc.GetTargetIP()}
		case "ALIAS":
			rc.RDATA = privatetypesrdata.ALIAS{Target: rc.GetTargetField()}
		case "AAAA":
			rc.RDATA = dnsrdatav2.AAAA{Addr: rc.GetTargetIP()}
		case "ADGUARDHOME_A_PASSTHROUGH":
			rc.RDATA = privatetypesrdata.ADGUARDHOME_A_PASSTHROUGH{}
		case "ADGUARDHOME_AAAA_PASSTHROUGH":
			rc.RDATA = privatetypesrdata.ADGUARDHOME_AAAA_PASSTHROUGH{}
		case "AZURE_ALIAS":
			rc.RDATA = privatetypesrdata.AZURE_ALIAS{Target: rc.GetTargetField(), AliasType: rc.AzureAlias["type"]}

		case "CAA":
			rc.RDATA = dnsrdatav2.CAA{Flag: rc.CaaFlag, Tag: rc.CaaTag, Value: rc.GetTargetField()}
		case "CNAME":
			targ := dnsutilv1.AddOrigin(rc.GetTargetField(), origin+".")
			rc.RDATA = dnsrdatav2.CNAME{Target: targ}

		case "CF_WORKER_ROUTE":
			part := strings.SplitN(rc.GetTargetField(), ",", 2)
			rc.RDATA = privatetypesrdata.CFWORKERROUTE{When: part[0], Then: part[1]}

		case "DHCID":
			rc.RDATA = dnsrdatav2.DHCID{Digest: rc.GetTargetField()}
		case "DNAME":
			rc.RDATA = dnsrdatav2.DNAME{Target: rc.GetTargetField()}
		case "DNSKEY":
			rc.RDATA = dnsrdatav2.DNSKEY{Flags: rc.DnskeyFlags, Protocol: rc.DnskeyProtocol, Algorithm: rc.DnskeyAlgorithm, PublicKey: rc.GetTargetField()}
		case "DS":
			rc.RDATA = dnsrdatav2.DS{KeyTag: rc.DsKeyTag, Algorithm: rc.DsAlgorithm, DigestType: rc.DsDigestType, Digest: rc.GetTargetField()}

		case "HTTPS":
			valuev2, err := convertSVCBv1v2(rc.GetSVCBValue())
			if err != nil {
				panic("BUG: Failed to convert SVCB value to v2: " + err.Error())
			}
			rc.RDATA = dnsrdatav2.SVCB{Priority: rc.SvcPriority, Target: rc.GetTargetField(), Value: valuev2}
			x1, x2, x3 := rc.RDATA.String(), rc.String(), rc.GetTargetCombined()
			if x1 != x2 {
				panic(fmt.Sprintf("BUG: SVCB String() is not stable: RDATA.String()=%s RecordConfig.String()=%s", x1, x2))
			}
			if x1 != x3 {
				panic(fmt.Sprintf("BUG: SVCB String() is not stable: RDATA.String()=%s GetTargetCombined()=%s", x1, x3))
			}

		case "LOC":
			rc.RDATA = dnsrdatav2.LOC{Version: rc.LocVersion, Size: rc.LocSize, HorizPre: rc.LocHorizPre, VertPre: rc.LocVertPre, Latitude: rc.LocLatitude, Longitude: rc.LocLongitude, Altitude: rc.LocAltitude}

		case "MX":
			rc.RDATA = dnsrdatav2.MX{Preference: rc.MxPreference, Mx: rc.GetTargetField()}

		case "NS":
			rc.RDATA = dnsrdatav2.NS{Ns: rc.GetTargetField()}
		case "NAPTR":
			rc.RDATA = dnsrdatav2.NAPTR{Order: rc.NaptrOrder, Preference: rc.NaptrPreference, Flags: rc.NaptrFlags, Service: rc.NaptrService, Regexp: rc.NaptrRegexp, Replacement: rc.GetTargetField()}

		case "OPENPGPKEY":
			rc.RDATA = dnsrdatav2.OPENPGPKEY{PublicKey: rc.GetTargetField()}

		case "PTR":
			rc.RDATA = dnsrdatav2.PTR{Ptr: rc.GetTargetField()}

		case "RP":
			rc.RDATA = dnsrdatav2.RP{Mbox: rc.F.(dnsv1.RP).Mbox, Txt: rc.F.(dnsv1.RP).Txt}
		case "R53_ALIAS":
			rc.RDATA = privatetypesrdata.R53_ALIAS{
				Target:           rc.GetTargetField(),
				AliasType:        rc.R53Alias["type"],
				ZoneID:           rc.R53Alias["zone_id"],
				EvalTargetHealth: rc.R53Alias["evaluate_target_health"],
			}

		case "SMIMEA":
			rc.RDATA = dnsrdatav2.SMIMEA{Usage: rc.SmimeaUsage, Selector: rc.SmimeaSelector, MatchingType: rc.SmimeaMatchingType, Certificate: rc.GetTargetField()}
		case "SOA":
			rc.RDATA = dnsrdatav2.SOA{Ns: rc.GetTargetField(), Mbox: rc.SoaMbox, Serial: rc.SoaSerial, Refresh: rc.SoaRefresh, Retry: rc.SoaRetry, Expire: rc.SoaExpire, Minttl: rc.SoaMinttl}
		case "SRV":
			rc.RDATA = dnsrdatav2.SRV{Priority: rc.SrvPriority, Weight: rc.SrvWeight, Port: rc.SrvPort, Target: rc.GetTargetField()}
		case "SSHFP":
			rc.RDATA = dnsrdatav2.SSHFP{Algorithm: rc.SshfpAlgorithm, Type: rc.SshfpFingerprint, FingerPrint: rc.GetTargetField()}
		case "SVCB":
			valuev2, err := convertSVCBv1v2(rc.GetSVCBValue())
			if err != nil {
				panic("BUG: Failed to convert SVCB value to v2: " + err.Error())
			}
			rc.RDATA = dnsrdatav2.SVCB{Priority: rc.SvcPriority, Target: rc.GetTargetField(), Value: valuev2}

		case "TLSA":
			rc.RDATA = dnsrdatav2.TLSA{Usage: rc.TlsaUsage, Selector: rc.TlsaSelector, MatchingType: rc.TlsaMatchingType, Certificate: rc.GetTargetField()}

		case "TXT":
			rc.RDATA = dnsrdatav2.TXT{Txt: []string{rc.GetTargetField()}}

		default:
			panic(fmt.Sprintf("RDATA FIXUP NOT IMPLEMENTED TYPE=%q", rc.Type))
		}
	}

	// .ComparableV3:
	if rc.ComparableV3 == "" {
		rc.ComparableV3 = rc.RDATA.String()
		if strings.HasSuffix(rc.ComparableV3, " ") {
			rc.ComparableV3 += "W"
		}
	}
}
