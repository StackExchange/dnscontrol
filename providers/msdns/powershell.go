package msdns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/TomOnTime/utfutil"
	ps "github.com/bhendo/go-powershell"
	"github.com/bhendo/go-powershell/backend"
	"github.com/bhendo/go-powershell/middleware"
)

type psHandle struct {
	shell ps.Shell
}

func newPowerShell(config map[string]string) (*psHandle, error) {

	back := &backend.Local{}
	sh, err := ps.New(back)
	if err != nil {
		return nil, err
	}
	shell := sh

	pssession := config["pssession"]
	if pssession != "" {
		fmt.Printf("INFO: PowerShell commands will run on %q\n", pssession)
		// create a remote shell by wrapping the existing one in the session middleware
		mconfig := middleware.NewSessionConfig()
		mconfig.ComputerName = pssession

		session, err := middleware.NewSession(sh, mconfig)
		if err != nil {
			panic(err)
		}
		shell = session
	}

	psh := &psHandle{
		shell: shell,
	}
	return psh, nil
}

func (psh *psHandle) Exit() {
	psh.shell.Exit()
}

type dnsZone map[string]interface{}

func (psh *psHandle) GetDNSServerZoneAll(dnsserver string) ([]string, error) {
	stdout, stderr, err := psh.shell.Execute(generatePSZoneAll(dnsserver))
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

// powerShellDump runs a PowerShell command to get a dump of all records in a DNS zone.
func generatePSZoneAll(dnsserver string) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, `Get-DnsServerZone`)
	if dnsserver != "" {
		fmt.Fprintf(&b, ` -ComputerName "%v"`, dnsserver)
	}
	fmt.Fprintf(&b, ` | `)
	fmt.Fprintf(&b, `ConvertTo-Json`)
	return b.String()
}

func (psh *psHandle) GetDNSZoneRecords(dnsserver, domain string) ([]nativeRecord, error) {
	tmpfile, err := ioutil.TempFile("", "zonerecords.*.json")
	if err != nil {
		log.Fatal(err)
	}
	tmpfile.Close()

	stdout, stderr, err := psh.shell.Execute(generatePSZoneDump(dnsserver, domain, tmpfile.Name()))
	if err != nil {
		return nil, err
	}
	if stdout != "" {
		fmt.Printf("STDOUT = %q\n", stderr)
	}
	if stderr != "" {
		fmt.Printf("STDERROR = %q\n", stderr)
		return nil, fmt.Errorf("unexpected stderr from PSZoneDump: %q", stderr)
	}

	contents, err := utfutil.ReadFile(tmpfile.Name(), utfutil.WINDOWS)
	if err != nil {
		return nil, err
	}
	os.Remove(tmpfile.Name())

	var records []nativeRecord
	json.Unmarshal([]byte(contents), &records)

	return records, nil
}

// powerShellDump runs a PowerShell command to get a dump of all records in a DNS zone.
func generatePSZoneDump(dnsserver, domainname string, filename string) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, `Get-DnsServerResourceRecord`)
	if dnsserver != "" {
		fmt.Fprintf(&b, ` -ComputerName "%v"`, dnsserver)
	}
	fmt.Fprintf(&b, ` -ZoneName "%v"`, domainname)
	fmt.Fprintf(&b, ` | `)
	fmt.Fprintf(&b, `ConvertTo-Json -depth 4`) // Tested with 3 (causes errors).  4 and larger work.
	fmt.Fprintf(&b, ` > %s`, filename)
	//fmt.Printf("DEBUG PSZoneDump CMD = (\n%s\n)\n", b.String())
	return b.String()
}

// Functions for record manipulation

func (psh *psHandle) RecordDelete(dnsserver, domain string, rec *models.RecordConfig) error {
	_, stderr, err := psh.shell.Execute(generatePSDelete(dnsserver, domain, rec))
	if err != nil {
		return err
	}
	if stderr != "" {
		fmt.Printf("STDERROR = %q\n", stderr)
		return fmt.Errorf("unexpected stderr from PSDelete: %q", stderr)
	}
	return nil
}

func generatePSDelete(dnsserver, domain string, rec *models.RecordConfig) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, `echo DELETE "%s" "%s" "%s"`, rec.Type, rec.Name, rec.GetTargetCombined())
	fmt.Fprintf(&b, " ; ")
	fmt.Fprintf(&b, `Remove-DnsServerResourceRecord`)
	if dnsserver != "" {
		fmt.Fprintf(&b, ` -ComputerName "%s"`, dnsserver)
	}
	fmt.Fprintf(&b, ` -Force`)
	fmt.Fprintf(&b, ` -ZoneName "%s"`, domain)
	fmt.Fprintf(&b, ` -Name "%s"`, rec.Name)
	fmt.Fprintf(&b, ` -RRType "%s"`, rec.Type)
	if rec.Type == "MX" {
		fmt.Fprintf(&b, ` -RecordData %d,"%s"`, rec.MxPreference, rec.GetTargetField())
	} else if rec.Type == "TXT" {
		fmt.Fprintf(&b, ` -RecordData %s`, rec.GetTargetField())
	} else if rec.Type == "SRV" {
		// https://www.gitmemory.com/issue/MicrosoftDocs/windows-powershell-docs/1149/511916884
		fmt.Fprintf(&b, ` -RecordData %d,%d,%d,"%s"`, rec.SrvPriority, rec.SrvWeight, rec.SrvPort, rec.GetTargetField())
	} else {
		fmt.Fprintf(&b, ` -RecordData "%s"`, rec.GetTargetField())
	}
	//fmt.Printf("DEBUG PSDelete CMD = (\n%s\n)\n", b.String())
	return b.String()
}

func (psh *psHandle) RecordCreate(dnsserver, domain string, rec *models.RecordConfig) error {
	_, stderr, err := psh.shell.Execute(generatePSCreate(dnsserver, domain, rec))
	if err != nil {
		return err
	}
	if stderr != "" {
		fmt.Printf("STDERROR = %q\n", stderr)
		return fmt.Errorf("unexpected stderr from PSCreate: %q", stderr)
	}
	return nil
}

func generatePSCreate(dnsserver, domain string, rec *models.RecordConfig) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, `echo CREATE "%s" "%s" "%s"`, rec.Type, rec.Name, rec.GetTargetCombined())
	fmt.Fprintf(&b, " ; ")

	fmt.Fprint(&b, `Add-DnsServerResourceRecord`)
	if dnsserver != "" {
		fmt.Fprintf(&b, ` -ComputerName "%s"`, dnsserver)
	}
	fmt.Fprintf(&b, ` -ZoneName "%s"`, domain)
	fmt.Fprintf(&b, ` -Name "%s"`, rec.GetLabel())
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
		fmt.Printf("DEBUG TXT len = %v\n", rec.TxtStrings)
		fmt.Printf("DEBUG TXT target = %q\n", rec.GetTargetField())
		fmt.Fprintf(&b, ` -Txt -DescriptiveText %s`, rec.GetTargetField())
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
			rec.Type, rec.GetLabel(), rec.GetTargetField()))
		// We panic so that we quickly find any switch statements
		// that have not been updated for a new RR type.
	}
	//fmt.Printf("DEBUG PSCreate CMD = (\n%s\n)\n", b.String())
	return b.String()
}

func (psh *psHandle) RecordModify(dnsserver, domain string, old, rec *models.RecordConfig) error {
	_, stderr, err := psh.shell.Execute(generatePSModify(dnsserver, domain, old, rec))
	if err != nil {
		return err
	}
	if stderr != "" {
		fmt.Printf("STDERROR = %q\n", stderr)
		return fmt.Errorf("unexpected stderr from PSModify: %q", stderr)
	}
	return nil
}

func generatePSModify(dnsserver, domain string, old, rec *models.RecordConfig) string {
	// The simple way is to just remove the old record and insert the new record.
	return generatePSDelete(dnsserver, domain, old) + ` ; ` + generatePSCreate(dnsserver, domain, rec)
	// NB: SOA records can't be deleted. When we implement them, we'll
	// need to special case them and generate an in-place modification
	// command.
}

// Note about the old generatePSModify:
//
// The old method is to generate PowerShell code that extracts the resource
// record, clones it, makes modifications to the clone, and replaces the old
// object with the modified clone. In theory this is cleaner.
//
// Sadly that technique is considerably slower (PowerShell seems to take a
// long time doing it) and it is more brittle (each new rType seems to be a
// new adventure).
//
// The other benefit of the Delete/Create method is that it more heavily
// exercises existing code that is known to work.
//
// Sadly I can't bring myself to erase the code yet. I still hope this can
// be fixed.  Deep down I know we should just accept that Del/Create is better.

// 	if old.GetLabel() != rec.GetLabel() {
// 		panic(fmt.Sprintf("generatePSModify assertion failed: %q != %q", old.GetLabel(), rec.GetLabel()))
// 	}
//
// 	var b bytes.Buffer
// 	fmt.Fprintf(&b, `echo "MODIFY %s %s %s old=(%s) new=(%s):"`, rec.GetLabel(), domain, rec.Type, old.GetTargetCombined(), rec.GetTargetCombined())
// 	fmt.Fprintf(&b, " ; ")
// 	fmt.Fprintf(&b, "$OldObj = Get-DnsServerResourceRecord")
// 	fmt.Fprintf(&b, ` -ZoneName "%s"`, domain)
// 	fmt.Fprintf(&b, ` -Name "%s"`, old.GetLabel())
// 	fmt.Fprintf(&b, ` -RRType "%s"`, old.Type)
// 	fmt.Fprintf(&b, ` | Where-Object {$_.HostName eq "%s" -and -RRType -eq "%s" -and `, old.GetLabel(), rec.Type)
// 	switch old.Type {
// 	case "A":
// 		fmt.Fprintf(&b, `$_.RecordData.IPv4Address -eq "%s"`, old.GetTargetIP())
// 	case "AAAA":
// 		fmt.Fprintf(&b, `$_.RecordData.IPv6Address -eq "%s"`, old.GetTargetIP())
// 	//case "ATMA":
// 	//	fmt.Fprintf(&b, ` -Atma -Address <String> -AddressType {E164 | AESA}`, old.GetTargetField())
// 	//case "AFSDB":
// 	//	fmt.Fprintf(&b, ` -Afsdb -ServerName <String> -SubType <UInt16>`, old.GetTargetField())
// 	case "SRV":
// 		fmt.Fprintf(&b, `$_.RecordData.DomainName -eq "%s" -and $_.RecordData.Port -eq %d -and $_.RecordData.Priority -eq %d -and $_.RecordData.Weight -eq %d`, old.GetTargetField(), old.SrvPort, old.SrvPriority, old.SrvWeight)
// 	case "CNAME":
// 		fmt.Fprintf(&b, `$_.RecordData.HostNameAlias -eq "%s"`, old.GetTargetField())
// 	//case "X25":
// 	//	fmt.Fprintf(&b, ` -X25 -PsdnAddress <String>`, old.GetTargetField())
// 	//case "WKS":
// 	//	fmt.Fprintf(&b, ` -Wks -InternetAddress <IPAddress> -InternetProtocol {UDP | TCP} -Service <String[]>`, old.GetTargetField())
// 	case "TXT":
// 		fmt.Fprintf(&b, `$_.RecordData.DescriptiveText -eq "%s"`, old.GetTargetField())
// 	//case "RT":
// 	//	fmt.Fprintf(&b, ` -RT -IntermediateHost <String> -Preference <UInt16>`, old.GetTargetField())
// 	//case "RP":
// 	//	fmt.Fprintf(&b, ` -RP -Description <String> -ResponsiblePerson <String>`, old.GetTargetField())
// 	case "PTR":
// 		fmt.Fprintf(&b, `$_.RecordData.PtrDomainName -eq "%s"`, old.GetTargetField())
// 	case "NS":
// 		fmt.Fprintf(&b, `$_.RecordData.NameServer -eq "%s"`, old.GetTargetField())
// 	case "MX":
// 		fmt.Fprintf(&b, `$_.RecordData.MailExchange -eq "%s" -and $_.RecordData.Preference -eq %d`, old.GetTargetField(), old.MxPreference)
// 	//case "ISDN":
// 	//	fmt.Fprintf(&b, ` -Isdn -IsdnNumber <String> -IsdnSubAddress <String>`, old.GetTargetField())
// 	//case "HINFO":
// 	//	fmt.Fprintf(&b, ` -HInfo -Cpu <String> -OperatingSystem <String>`, old.GetTargetField())
// 	//case "DNAME":
// 	//	fmt.Fprintf(&b, ` -DName -DomainNameAlias <String>`, old.GetTargetField())
// 	//case "DHCID":
// 	//	fmt.Fprintf(&b, ` -DhcId -DhcpIdentifier <String>`, old.GetTargetField())
// 	//case "TLSA":
// 	//	fmt.Fprintf(&b, ` -TLSA -CertificateAssociationData <System.String> -CertificateUsage {CAConstraint | ServiceCertificateConstraint | TrustAnchorAssertion | DomainIssuedCertificate} -MatchingType {ExactMatch | Sha256Hash | Sha512Hash} -Selector {FullCertificate | SubjectPublicKeyInfo}`, rec.GetTargetField())
// 	default:
// 		panic(fmt.Errorf("generatePSModify() has not implemented recType=%q recName=%q content=(%s))",
// 			rec.Type, rec.GetLabel(), rec.GetTargetCombined()))
// 		// We panic so that we quickly find any switch statements
// 		// that have not been updated for a new RR type.
// 	}
// 	fmt.Fprintf(&b, "}")
// 	fmt.Fprintf(&b, " ; ")
// 	fmt.Fprintf(&b, `if($OldObj.Length -ne 1){ throw "Error, multiple results for Get-DnsServerResourceRecord" }`)
// 	fmt.Fprintf(&b, " ; ")
// 	fmt.Fprintf(&b, "$NewObj = $OldObj.Clone()")
// 	fmt.Fprintf(&b, " ; ")
//
// 	if old.TTL != rec.TTL {
// 		fmt.Fprintf(&b, `$NewObj.TimeToLive = New-TimeSpan -Seconds %d`, rec.TTL)
// 		fmt.Fprintf(&b, " ; ")
// 	}
// 	switch rec.Type {
// 	case "A":
// 		fmt.Fprintf(&b, `$NewObj.RecordData.IPv4Address = "%s"`, rec.GetTargetIP())
// 	case "AAAA":
// 		fmt.Fprintf(&b, `$NewObj.RecordData.IPv6Address = "%s"`, rec.GetTargetIP())
// 	//case "ATMA":
// 	//	fmt.Fprintf(&b, ` -Atma -Address <String> -AddressType {E164 | AESA}`, rec.GetTargetField())
// 	//case "AFSDB":
// 	//	fmt.Fprintf(&b, ` -Afsdb -ServerName <String> -SubType <UInt16>`, rec.GetTargetField())
// 	case "SRV":
// 		fmt.Fprintf(&b, ` -Srv -DomainName "%s" -Port %d -Priority %d -Weight %d`, rec.GetTargetField(), rec.SrvPort, rec.SrvPriority, rec.SrvWeight)
// 		fmt.Fprintf(&b, `$NewObj.RecordData.DomainName = "%s"`, rec.GetTargetField())
// 		fmt.Fprintf(&b, " ; ")
// 		fmt.Fprintf(&b, `$NewObj.RecordData.Port = %d`, rec.SrvPort)
// 		fmt.Fprintf(&b, " ; ")
// 		fmt.Fprintf(&b, `$NewObj.RecordData.Priority = %d`, rec.SrvPriority)
// 		fmt.Fprintf(&b, " ; ")
// 		fmt.Fprintf(&b, `$NewObj.RecordData.Weight = "%d"`, rec.SrvWeight)
// 	case "CNAME":
// 		fmt.Fprintf(&b, `$NewObj.RecordData.HostNameAlias = "%s"`, rec.GetTargetField())
// 	//case "X25":
// 	//	fmt.Fprintf(&b, ` -X25 -PsdnAddress <String>`, rec.GetTargetField())
// 	//case "WKS":
// 	//	fmt.Fprintf(&b, ` -Wks -InternetAddress <IPAddress> -InternetProtocol {UDP | TCP} -Service <String[]>`, rec.GetTargetField())
// 	case "TXT":
// 		fmt.Fprintf(&b, `$NewObj.RecordData.DescriptiveText = "%s"`, rec.GetTargetField())
// 	//case "RT":
// 	//	fmt.Fprintf(&b, ` -RT -IntermediateHost <String> -Preference <UInt16>`, rec.GetTargetField())
// 	//case "RP":
// 	//	fmt.Fprintf(&b, ` -RP -Description <String> -ResponsiblePerson <String>`, rec.GetTargetField())
// 	case "PTR":
// 		fmt.Fprintf(&b, `$NewObj.RecordData.PtrDomainName = "%s"`, rec.GetTargetField())
// 	case "NS":
// 		fmt.Fprintf(&b, `$NewObj.RecordData.NameServer = "%s"`, rec.GetTargetField())
// 	case "MX":
// 		fmt.Fprintf(&b, `$NewObj.RecordData.MailExchange = "%s"`, rec.GetTargetField())
// 		fmt.Fprintf(&b, " ; ")
// 		fmt.Fprintf(&b, `$NewObj.RecordData.Preference = "%d"`, rec.MxPreference)
// 	//case "ISDN":
// 	//	fmt.Fprintf(&b, ` -Isdn -IsdnNumber <String> -IsdnSubAddress <String>`, rec.GetTargetField())
// 	//case "HINFO":
// 	//	fmt.Fprintf(&b, ` -HInfo -Cpu <String> -OperatingSystem <String>`, rec.GetTargetField())
// 	//case "DNAME":
// 	//	fmt.Fprintf(&b, ` -DName -DomainNameAlias <String>`, rec.GetTargetField())
// 	//case "DHCID":
// 	//	fmt.Fprintf(&b, ` -DhcId -DhcpIdentifier <String>`, rec.GetTargetField())
// 	//case "TLSA":
// 	//	fmt.Fprintf(&b, ` -TLSA -CertificateAssociationData <System.String> -CertificateUsage {CAConstraint | ServiceCertificateConstraint | TrustAnchorAssertion | DomainIssuedCertificate} -MatchingType {ExactMatch | Sha256Hash | Sha512Hash} -Selector {FullCertificate | SubjectPublicKeyInfo}`, rec.GetTargetField())
// 	default:
// 		panic(fmt.Errorf("generatePSModify() update has not implemented recType=%q recName=%q content=(%s))",
// 			rec.Type, rec.GetLabel(), rec.GetTargetCombined()))
// 		// We panic so that we quickly find any switch statements
// 		// that have not been updated for a new RR type.
// 	}
// 	fmt.Fprintf(&b, " ; ")
// 	//fmt.Printf("DEBUG CCMD: %s\n", b.String())
//
// 	fmt.Fprintf(&b, "Set-DnsServerResourceRecord")
// 	fmt.Fprintf(&b, ` -ZoneName "%s"`, domain)
// 	fmt.Fprintf(&b, ` -NewInputObject $NewObj -OldInputObject $OldObj`)
//
//  fmt.Printf("DEBUG MCMD: %s", b.String())
// 	return b.String()
