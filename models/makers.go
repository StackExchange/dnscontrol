package models

import (
	dnsv2 "codeberg.org/miekg/dns"
	dnsrdatav2 "codeberg.org/miekg/dns/rdata"

	"github.com/DNSControl/dnscontrol/v4/pkg/mustbe"
	_ "github.com/DNSControl/dnscontrol/v4/pkg/privatetypes"
)

func MakeCNAME(origin, target string) (dnsv2.RDATA, error) {
	return dnsrdatav2.CNAME{Target: mustbe.TargetHost(origin, target)}, nil
}

/*
			//targ := dnsutilv1.AddOrigin(rc.GetTargetField(), origin+".")
			//rc.RDATA = dnsrdatav2.CNAME{Target: targ}
			rc.RDATA, err = MakeCNAME(origin, rc.GetTargetField())

		case "CF_WORKER_ROUTE":
			part := strings.SplitN(rc.GetTargetField(), ",", 2)
			rc.RDATA = privatetypesrdata.CFWORKERROUTE{When: part[0], Then: part[1]}

		case "DHCID":
			rc.RDATA = dnsrdatav2.DHCID{Digest: rc.GetTargetField()}
		case "DNAME":
			targ := dnsutilv1.AddOrigin(rc.GetTargetField(), origin+".")
			rc.RDATA = dnsrdatav2.DNAME{Target: targ}
		case "DNSKEY":
			rc.RDATA = dnsrdatav2.DNSKEY{Flags: rc.DnskeyFlags, Protocol: rc.DnskeyProtocol, Algorithm: rc.DnskeyAlgorithm, PublicKey: rc.DnskeyPublicKey}
		case "DS":
			rc.RDATA = dnsrdatav2.DS{KeyTag: rc.DsKeyTag, Algorithm: rc.DsAlgorithm, DigestType: rc.DsDigestType, Digest: rc.GetTargetField()}

		case "FRAME":
			rc.RDATA = privatetypesrdata.FRAME{Target: rc.GetTargetField()}

		case "HTTPS":
			if rc.SvcPriority == 0 {
				rc.RDATA = dnsrdatav2.SVCB{Priority: rc.SvcPriority, Target: rc.GetTargetField()}
			} else {
				p := rc.SvcParams
				p = strings.ReplaceAll(p, `ech=IGNORE`, ``)
				p = strings.ReplaceAll(p, ` `, ` `) // Collapse 2 spaces into 1
				p = strings.TrimSpace(p)
				rd, err := dnsv2.NewData(dnsv2.TypeHTTPS, fmt.Sprintf("%d %s %s", rc.SvcPriority, rc.GetTargetField(), p), origin)
				if err != nil {
					panic(fmt.Sprintf("BUG: Failed to create RDATA for HTTPS record: %v", err))
				}
				rc.RDATA = rd
			}

		case "LOC":
			rc.RDATA = dnsrdatav2.LOC{Version: rc.LocVersion, Size: rc.LocSize, HorizPre: rc.LocHorizPre, VertPre: rc.LocVertPre, Latitude: rc.LocLatitude, Longitude: rc.LocLongitude, Altitude: rc.LocAltitude}

		case "MIKROTIK_FWD":
			rc.RDATA = privatetypesrdata.MIKROTIKFWD{ForwardTo: rc.GetTargetField()}
		case "MIKROTIK_NXDOMAIN":
			rc.RDATA = privatetypesrdata.MIKROTIKNXDOMAIN{}
		case "MX":
			rc.RDATA = dnsrdatav2.MX{Preference: rc.MxPreference, Mx: rc.GetTargetField()}

		case "NS":
			rc.RDATA = dnsrdatav2.NS{Ns: rc.GetTargetField()}
		case "NAPTR":
			rc.RDATA = dnsrdatav2.NAPTR{Order: rc.NaptrOrder, Preference: rc.NaptrPreference, Flags: rc.NaptrFlags, Service: rc.NaptrService, Regexp: rc.NaptrRegexp, Replacement: rc.GetTargetField()}

		case "OPENPGPKEY":
			rc.RDATA = dnsrdatav2.OPENPGPKEY{PublicKey: rc.GetTargetField()}

		case "PORKBUN_URLFWD":
			rc.RDATA = privatetypesrdata.PORKBUNURLFWD{}
		case "PTR":
			rc.RDATA = dnsrdatav2.PTR{Ptr: rc.GetTargetField()}

		case "RP":
			rc.RDATA = dnsrdatav2.RP{Mbox: rc.F.(dnsv1.RP).Mbox, Txt: rc.F.(dnsv1.RP).Txt}
		case "R53_ALIAS":
			rc.RDATA = privatetypesrdata.R53ALIAS{
				AliasType:        rc.R53Alias["type"],
				Target:           rc.GetTargetField(),
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
			if rc.SvcPriority == 0 {
				rc.RDATA = dnsrdatav2.SVCB{Priority: rc.SvcPriority, Target: rc.GetTargetField()}
			} else {
				rd, err := dnsv2.NewData(dnsv2.TypeSVCB, fmt.Sprintf("%d %s %s", rc.SvcPriority, rc.GetTargetField(), rc.SvcParams), origin)
				if err != nil {
					panic(fmt.Sprintf("BUG: Failed to create RDATA for HTTPS record: %v", err))
				}
				rc.RDATA = rd
			}

		case "TLSA":
			rc.RDATA = dnsrdatav2.TLSA{Usage: rc.TlsaUsage, Selector: rc.TlsaSelector, MatchingType: rc.TlsaMatchingType, Certificate: rc.GetTargetField()}

		case "TXT":
			rc.RDATA = dnsrdatav2.TXT{Txt: []string{rc.GetTargetField()}}

		case "URL":
			rc.RDATA = privatetypesrdata.URL{Location: rc.GetTargetField()}
		case "URL301":
			rc.RDATA = privatetypesrdata.URL{Location: rc.GetTargetField()}

		default:
			panic(fmt.Sprintf("RDATA FIXUP NOT IMPLEMENTED TYPE=%q", rc.Type))
		}
	}

	// .ComparableV3:
	if rc.ComparableV3 == "" {
		switch rc.Type {
		case "SOA":
			// The comparable string for SOA intentionally excludes the serial
			// number, because the serial number changes on every update and
			// would prevent correct diffing. List it as "X" so-as it stands out
			// in debug output that the serial is intentionally excluded.
			rc.ComparableV3 = fmt.Sprintf("%s %s X %d %d %d %d", rc.GetTargetField(), rc.SoaMbox, rc.SoaRefresh, rc.SoaRetry, rc.SoaExpire, rc.SoaMinttl)
		default:
			rc.ComparableV3 = strings.TrimSpace(rc.RDATA.String())
		}

		// Note to self: RDATA.String() sometimes leaves a trailing space.  File a bug.
		// if strings.HasSuffix(rc.ComparableV3, " ") {
		// 	rc.ComparableV3 = rc.ComparableV3 + "W"
		// }
	}
}
*/
