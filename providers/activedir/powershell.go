package activedir

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/StackExchange/dnscontrol/v3/models"
	ps "github.com/bhendo/go-powershell"
	"github.com/bhendo/go-powershell/backend"
)

type psHandle struct {
	shell ps.Shell
}

func newPowerShell() (*psHandle, error) {

	back := &backend.Local{}

	sh, err := ps.New(back)
	if err != nil {
		return nil, err
	}
	//defer sh.Exit()

	psh := &psHandle{
		shell: sh,
	}
	return psh, nil

}

func (psh *psHandle) Exit() {
	psh.shell.Exit()
}

type dnsZone map[string]interface{}

func (psh *psHandle) GetDNSServerZoneAll() ([]string, error) {
	stdout, stderr, err := psh.shell.Execute(`Get-DnsServerZone | ConvertTo-Json`)
	if err != nil {
		return nil, err
	}
	if stderr != "" {
		fmt.Printf("STDERROR = %q\n", stderr)
		return nil, fmt.Errorf("unexpected stderr from Get-DnsServerZones: %q", stderr)
	}

	var zones []dnsZone
	json.Unmarshal([]byte(stdout), &zones)

	var result []string
	for _, z := range zones {
		zonename := z["ZoneName"].(string)
		result = append(result, zonename)
	}

	return result, nil
}

func (psh *psHandle) GetDNSZoneRecords(domain string) ([]nativeRecord, error) {
	stdout, stderr, err := psh.shell.Execute(generatePSZoneDump(domain))
	if err != nil {
		return nil, err
	}
	if stderr != "" {
		fmt.Printf("STDERROR = %q\n", stderr)
		return nil, fmt.Errorf("unexpected stderr from PSZoneDump: %q", stderr)
	}
	//fmt.Printf("OUT = \n%v\n", string(stdout))

	var records []nativeRecord
	json.Unmarshal([]byte(stdout), &records)
	//fmt.Printf("RECORDS = \n%v\n", records)

	return records, nil
}

// powerShellDump runs a PowerShell command to get a dump of all records in a DNS zone.
func generatePSZoneDump(domainname string) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, `Get-DnsServerResourceRecord -ZoneName "%v"`, domainname)
	//fmt.Fprintf(&b, ` | `)
	//fmt.Fprintf(&b, `select hostname,recordtype,@{n="timestamp";e={$_.timestamp.tostring()}},@{n="timetolive";e={$_.timetolive.totalseconds}},@{n="recorddata";e={($_.recorddata.ipv4address,$_.recorddata.ipv6address,$_.recorddata.HostNameAlias,$_.recorddata.NameServer,"unsupported_record_type" -ne $null)[0]-as [string]}}`)
	fmt.Fprintf(&b, ` | `)
	fmt.Fprintf(&b, `ConvertTo-Json -depth 10`)
	return b.String()
}

//

func generatePSDelete(domain string, rec *models.RecordConfig) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, `echo DELETE "%s" "%s" "%s"`, rec.Type, rec.Name, rec.GetTargetCombined())
	fmt.Fprintf(&b, " ; ")
	fmt.Fprintf(&b, `Remove-DnsServerResourceRecord -Force`)
	fmt.Fprintf(&b, ` -ZoneName "%s"`, domain)
	fmt.Fprintf(&b, ` -Name "%s"`, rec.Name)
	fmt.Fprintf(&b, ` -RRType "%s"`, rec.Type)
	if rec.Type == "MX" {
		fmt.Fprintf(&b, ` -RecordData %d,"%s"`, rec.MxPreference, rec.GetTargetField())
	} else {
		fmt.Fprintf(&b, ` -RecordData "%s"`, rec.GetTargetField())
	}
	//fmt.Printf("DEBUG DCMD: %s\n", b.String())
	return b.String()
}

func (psh *psHandle) RecordCreate(domain string, rec *models.RecordConfig) error {
	_, stderr, err := psh.shell.Execute(generatePSCreate(domain, rec))
	if err != nil {
		return err
	}
	if stderr != "" {
		fmt.Printf("STDERROR = %q\n", stderr)
		return fmt.Errorf("unexpected stderr from PSCreate: %q", stderr)
	}
	return nil
}

func generatePSCreate(domain string, rec *models.RecordConfig) string {
	content := rec.GetTargetField()

	var b bytes.Buffer
	fmt.Fprintf(&b, `echo CREATE "%s" "%s" "%s"`, rec.Type, rec.Name, rec.GetTargetCombined())
	fmt.Fprintf(&b, " ; ")

	fmt.Fprintf(&b, `Add-DnsServerResourceRecord -ZoneName "%s" -Name "%s"`, domain, rec.GetLabel())
	fmt.Fprintf(&b, ` -TimeToLive $(New-TimeSpan -Seconds %d)`, rec.TTL)
	switch rec.Type {
	case "A":
		fmt.Fprintf(&b, ` -A -IPv4Address "%s"`, rec.GetTargetIP())
	case "AAAA":
		fmt.Fprintf(&b, ` -AAAA -IPv6Address "%s"`, rec.GetTargetIP())
	//case "ATMA":
	//	fmt.Fprintf(&b, ` -Atma -Address <String> -AddressType {E164 | AESA}`, rec.GetTargetField())
	//case "AFSDB":
	//	fmt.Fprintf(&b, ` -Afsdb -ServerName <String> -SubType <UInt16>`, rec.GetTargetField())
	case "SRV":
		fmt.Fprintf(&b, ` -Srv -DomainName "%s" -Port %d -Priority %d -Weight %d`, rec.GetTargetField(), rec.SrvPort, rec.SrvPriority, rec.SrvWeight)
	case "CNAME":
		fmt.Fprintf(&b, ` -CName -HostNameAlias "%s"`, rec.GetTargetField())
	//case "X25":
	//	fmt.Fprintf(&b, ` -X25 -PsdnAddress <String>`, rec.GetTargetField())
	//case "WKS":
	//	fmt.Fprintf(&b, ` -Wks -InternetAddress <IPAddress> -InternetProtocol {UDP | TCP} -Service <String[]>`, rec.GetTargetField())
	case "TXT":
		fmt.Fprintf(&b, ` -Txt -DescriptiveText "%s"`, rec.GetTargetField())
	//case "RT":
	//	fmt.Fprintf(&b, ` -RT -IntermediateHost <String> -Preference <UInt16>`, rec.GetTargetField())
	//case "RP":
	//	fmt.Fprintf(&b, ` -RP -Description <String> -ResponsiblePerson <String>`, rec.GetTargetField())
	case "PTR":
		fmt.Fprintf(&b, ` -Ptr -PtrDomainName "%s"`, rec.GetTargetField())
	case "NS":
		fmt.Fprintf(&b, ` -NS -NameServer "%s"`, rec.GetTargetField())
	case "MX":
		fmt.Fprintf(&b, ` -MX -MailExchange "%s" -Preference %d`, rec.GetTargetField(), rec.MxPreference)
	//case "ISDN":
	//	fmt.Fprintf(&b, ` -Isdn -IsdnNumber <String> -IsdnSubAddress <String>`, rec.GetTargetField())
	//case "HINFO":
	//	fmt.Fprintf(&b, ` -HInfo -Cpu <String> -OperatingSystem <String>`, rec.GetTargetField())
	//case "DNAME":
	//	fmt.Fprintf(&b, ` -DName -DomainNameAlias <String>`, rec.GetTargetField())
	//case "DHCID":
	//	fmt.Fprintf(&b, ` -DhcId -DhcpIdentifier <String>`, rec.GetTargetField())
	//case "TLSA":
	//	fmt.Fprintf(&b, ` -TLSA -CertificateAssociationData <System.String> -CertificateUsage {CAConstraint | ServiceCertificateConstraint | TrustAnchorAssertion | DomainIssuedCertificate} -MatchingType {ExactMatch | Sha256Hash | Sha512Hash} -Selector {FullCertificate | SubjectPublicKeyInfo}`, rec.GetTargetField())
	default:
		panic(fmt.Errorf("generatePSCreate() has not implemented recType=%s recName=%#v content=%#v)",
			rec.Type, rec.GetLabel(), content))
		// We panic so that we quickly find any switch statements
		// that have not been updated for a new RR type.
	}
	//fmt.Printf("DEBUG CCMD: %s\n", b.String())
	return b.String()
}

func (psh *psHandle) RecordDelete(domain string, rec *models.RecordConfig) error {
	_, stderr, err := psh.shell.Execute(generatePSDelete(domain, rec))
	if err != nil {
		return err
	}
	if stderr != "" {
		fmt.Printf("STDERROR = %q\n", stderr)
		return fmt.Errorf("unexpected stderr from PSDelete: %q", stderr)
	}
	return nil
}

func generatePSModify(domain string, old, rec *models.RecordConfig) string {

	if true {

		dcmd := generatePSDelete(domain, old)
		ccmd := generatePSCreate(domain, rec)
		return dcmd + ` ; ` + ccmd

	}
	recName := rec.Name
	recType := rec.Type
	oldContent := old.GetTargetField()
	newContent := rec.GetTargetField()
	oldTTL := old.TTL
	newTTL := rec.TTL

	var queryField, queryContent string
	queryContent = `"` + oldContent + `"`

	switch recType { // #rtype_variations
	case "A":
		queryField = "IPv4address"
	case "CNAME":
		queryField = "HostNameAlias"
	case "MX":
		queryField = ""
	case "NS":
		queryField = "NameServer"
	default:
		panic(fmt.Errorf("generatePowerShellModify() does not yet handle recType=%s recName=%#v content=(%#v, %#v)", recType, recName, oldContent, newContent))
		// We panic so that we quickly find any switch statements
		// that have not been updated for a new RR type.
	}

	var b bytes.Buffer

	fmt.Fprintf(&b, `echo "MODIFY %s %s %s old=%s new=%s"`, recName, domain, recType, oldContent, newContent)
	fmt.Fprintf(&b, " ; ")
	fmt.Fprintf(&b, "$OldObj = Get-DnsServerResourceRecord")
	//fmt.Fprintf(&b, ` -ComputerName "%s"`, c.adServer)
	fmt.Fprintf(&b, ` -ZoneName "%s"`, domain)
	fmt.Fprintf(&b, ` -Name "%s"`, recName)
	fmt.Fprintf(&b, ` -RRType "%s"`, recType)
	fmt.Fprintf(&b, " | ")
	if recType == "MX" {
		fmt.Fprintf(&b, `Where-Object {$_.RecordData.Preference -eq %d -and  $_.RecordData.MailExchange -eq "%s" -and $_.HostName -eq "%s"}`, old.MxPreference, old.GetTargetField(), recName)
	} else {
		fmt.Fprintf(&b, `Where-Object {$_.RecordData.%s -eq %s -and $_.HostName -eq "%s"}`, queryField, queryContent, recName)
	}
	fmt.Fprintf(&b, " ; ")
	fmt.Fprintf(&b, `if($OldObj.Length -ne 1){ throw "Error, multiple results for Get-DnsServerResourceRecord" }`)
	fmt.Fprintf(&b, " ; ")
	fmt.Fprintf(&b, "$NewObj = $OldObj.Clone()")

	if old.Type == "MX" {
		if old.MxPreference != rec.MxPreference {
			fmt.Fprintf(&b, " ; ")
			fmt.Fprintf(&b, `$NewObj.RecordData.Preference = %d`, rec.MxPreference)
		}
		if old.GetTargetField() != rec.GetTargetField() {
			fmt.Fprintf(&b, " ; ")
			fmt.Fprintf(&b, `$NewObj.RecordData.MailExchange = "%s"`, rec.GetTargetField())
		}
	} else if oldContent != newContent {
		fmt.Fprintf(&b, " ; ")
		fmt.Fprintf(&b, `$NewObj.RecordData.%s = "%s"`, queryField, newContent)
	}

	if rec.Type == "MX" && old.MxPreference != rec.MxPreference {
		fmt.Fprintf(&b, " ; ")
		fmt.Fprintf(&b, `$NewObj.RecordData.Preference = %d`, rec.MxPreference)
		fmt.Fprintf(&b, " ; ")
		fmt.Fprintf(&b, `$NewObj.RecordData.MailExchange = "%s"`, rec.GetTargetField())
	}

	if oldTTL != newTTL {
		fmt.Fprintf(&b, " ; ")
		fmt.Fprintf(&b, `$NewObj.TimeToLive = New-TimeSpan -Seconds %d`, newTTL)
	}

	fmt.Fprintf(&b, " ; ")
	fmt.Fprintf(&b, "Set-DnsServerResourceRecord")
	fmt.Fprintf(&b, ` -ZoneName "%s"`, domain)
	fmt.Fprintf(&b, ` -NewInputObject $NewObj -OldInputObject $OldObj`)

	fmt.Printf("MODIFY: %s", b.String())
	if old.Type == "MX" {
		os.Exit(1)
	}
	return b.String()
}

func (psh *psHandle) RecordModify(domain string, old, rec *models.RecordConfig) error {
	_, stderr, err := psh.shell.Execute(generatePSModify(domain, old, rec))
	if err != nil {
		return err
	}
	if stderr != "" {
		fmt.Printf("STDERROR = %q\n", stderr)
		return fmt.Errorf("unexpected stderr from PSModify: %q", stderr)
	}
	return nil
}
