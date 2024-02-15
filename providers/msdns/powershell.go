package msdns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/StackExchange/dnscontrol/v4/models"
	ps "github.com/StackExchange/dnscontrol/v4/pkg/powershell"
	"github.com/StackExchange/dnscontrol/v4/pkg/powershell/backend"
	"github.com/StackExchange/dnscontrol/v4/pkg/powershell/middleware"
	"github.com/StackExchange/dnscontrol/v4/pkg/printer"
	"github.com/TomOnTime/utfutil"
)

type psHandle struct {
	shell ps.Shell
}

func eLog(s string) {
	f, _ := os.OpenFile("powershell.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	f.WriteString(s)
	f.WriteString("\n")
	f.Close()
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
		printer.Printf("INFO: PowerShell commands will run on %q\n", pssession)
		// create a remote shell by wrapping the existing one in the session middleware
		mconfig := middleware.NewSessionConfig()
		mconfig.ComputerName = pssession

		cred := &middleware.UserPasswordCredential{
			Username: config["psusername"],
			Password: config["pspassword"],
		}
		if cred.Password != "" && cred.Username != "" {
			mconfig.Credential = cred
		}

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
	stdout, stderr, err := psh.shell.Execute("\n\r" + generatePSZoneAll(dnsserver) + "\n\r")
	if err != nil {
		return nil, err
	}
	if stderr != "" {
		printer.Printf("STDERROR = %q\n", stderr)
		return nil, fmt.Errorf("unexpected stderr from Get-DnsServerZones: %q", stderr)
	}

	var zones []dnsZone
	err = json.Unmarshal([]byte(stdout), &zones)
	if err != nil {
		return nil, err
	}

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

	tmpfile, err := os.CreateTemp("", "zonerecords.*.json")
	if err != nil {
		log.Fatal(err)
	}
	filename := tmpfile.Name()
	tmpfile.Close()

	stdout, stderr, err := psh.shell.Execute(
		"\n\r" + generatePSZoneDump(dnsserver, domain, filename) + "\n\r")
	if err != nil {
		return nil, err
	}
	if stderr != "" {
		printer.Printf("STDERROR GetDNSZR = %q\n", stderr)
		return nil, fmt.Errorf("unexpected stderr from PSZoneDump: %q", stderr)
	}
	if stdout != "" {
		printer.Printf("STDOUT GetDNSZR = %q\n", stdout)
	}

	contents, err := utfutil.ReadFile(filename, utfutil.UTF8)
	if err != nil {
		return nil, err
	}
	os.Remove(filename) // TODO(tlim): There should be a debug flag that leaves the tmp file around.

	//printer.Printf("CONTENTS = %s\n", contents)
	//printer.Printf("CONTENTS STR = %q\n", contents[:10])
	//printer.Printf("CONTENTS HEX = %v\n", []byte(contents)[:10])
	//os.WriteFile("/temp/list.json", contents, 0777)
	var records []nativeRecord
	err = json.Unmarshal(contents, &records)
	if err != nil {
		// PowerShell generates bad JSON if there is only one record.  Therefore, if there
		// is an error we try decoding the bad format before completing erroring out.
		// The "bad JSON" is that they generate a single record instead of a list of length 1.
		records = append(records, nativeRecord{})
		err2 := json.Unmarshal(contents, &(records[0]))
		if err2 != nil {
			return nil, fmt.Errorf("PSZoneDump json error: %w", err)
		}
	}

	return records, nil
}

// powerShellDump runs a PowerShell command to get a dump of all records in a DNS zone.
func generatePSZoneDump(dnsserver, domainname, filename string) string {
	// @dnsserver: Hostname of the DNS server.
	// @domainname: Name of the domain.
	// @filename: Where to write the resulting JSON file.
	// NB(tlim): On Windows PowerShell, the JSON file will be UTF8 with
	// a BOM.  A UTF-8 file shouldn't have a BOM, but Microsoft messed up.
	// When we switch to PowerShell Core, the BOM will disappear.
	var b bytes.Buffer

	// Set the output to be UTF8.  Previously we didn't do that and the
	// output was twice as large, plus it required an extra conversion
	// step.  Windows PowerShell is native UTF16 but PowerShell Core is
	// native UTF8, thus this may not be needed if we move to Core.
	fmt.Fprintf(&b, `$OutputEncoding = [Text.UTF8Encoding]::UTF8 ; `)

	// Output everything we know about the zone.
	fmt.Fprintf(&b, `Get-DnsServerResourceRecord`)
	if dnsserver != "" {
		fmt.Fprintf(&b, ` -ComputerName "%v"`, dnsserver)
	}
	fmt.Fprintf(&b, ` -ZoneName "%v"`, domainname)

	// Strip out the `Cim*` properties at the root. This shrinks one
	// zone from 99M to 11M.  We don't need the Cim* properties (at
	// least the ones at the root) and decocding 99M of JSON was slow
	// (30+ minutes).
	// NB(tlim): Windows PowerShell requires the `-Property *` but
	// Windows PowerShell Core makes that optional.
	fmt.Fprintf(&b, ` | `)
	fmt.Fprintf(&b, `Select-Object -Property * -ExcludeProperty Cim*`)
	fmt.Fprintf(&b, ` | `)
	fmt.Fprintf(&b, `ConvertTo-Json -depth 4`) // Tested with 3 (causes errors).  4 and larger work.
	fmt.Fprintf(&b, ` | `)

	// Prevously we captured stdout. Now we write it to a file. This is
	// safer since there is no chance of junk accidentally being mixed
	// into stdout.
	fmt.Fprintf(&b, `Out-File "%s" -Encoding utf8`, filename)
	return b.String()
}

// Functions for record manipulation

func (psh *psHandle) RecordDelete(dnsserver, domain string, rec *models.RecordConfig) error {

	var c string
	if rec.Type == "NAPTR" {
		c = generatePSDeleteNaptr(dnsserver, domain, rec)
		//printer.Printf("DEBUG: deleteNAPTR: %s\n", c)
	} else {
		c = generatePSDelete(dnsserver, domain, rec)
	}

	eLog(c)
	_, stderr, err := psh.shell.Execute("\n\r" + c + "\n\r")
	if err != nil {
		printer.Printf("PowerShell code was:\nSTART\n%s\nEND\n", c)
		return err
	}
	if stderr != "" {
		printer.Printf("STDERROR = %q\n", stderr)
		printer.Printf("PowerShell code was:\nSTART\n%s\nEND\n", c)
		return fmt.Errorf("unexpected stderr from PSDelete: %q", stderr)
	}
	return nil
}

func generatePSDelete(dnsserver, domain string, rec *models.RecordConfig) string {

	var b bytes.Buffer
	fmt.Fprintf(&b, `echo DELETE "%s" "%s" %q`, rec.Type, rec.Name, rec.GetTargetCombined())
	fmt.Fprintf(&b, " ; ")

	if rec.Type == "NAPTR" {
		x := b.String() + generatePSDeleteNaptr(dnsserver, domain, rec)
		//printer.Printf("NAPTR DELETE: %s\n", x)
		return x
	}

	fmt.Fprintf(&b, `Remove-DnsServerResourceRecord`)
	if dnsserver != "" {
		fmt.Fprintf(&b, ` -ComputerName "%s"`, dnsserver)
	}
	fmt.Fprintf(&b, ` -Force`)
	fmt.Fprintf(&b, ` -ZoneName %q`, domain)
	fmt.Fprintf(&b, ` -Name %q`, rec.Name)
	fmt.Fprintf(&b, ` -RRType "%s"`, rec.Type)
	if rec.Type == "MX" {
		fmt.Fprintf(&b, ` -RecordData %d,%q`, rec.MxPreference, rec.GetTargetField())
	} else if rec.Type == "TXT" {
		fmt.Fprintf(&b, ` -RecordData %q`, rec.GetTargetTXTJoined())
	} else if rec.Type == "SRV" {
		// https://www.gitmemory.com/issue/MicrosoftDocs/windows-powershell-docs/1149/511916884
		fmt.Fprintf(&b, ` -RecordData %d,%d,%d,"%s"`, rec.SrvPriority, rec.SrvWeight, rec.SrvPort, rec.GetTargetField())
	} else {
		fmt.Fprintf(&b, ` -RecordData %q`, rec.GetTargetField())
	}
	//printer.Printf("DEBUG PSDelete CMD = (\n%s\n)\n", b.String())
	return b.String()
}

func (psh *psHandle) RecordCreate(dnsserver, domain string, rec *models.RecordConfig) error {

	var c string
	if rec.Type == "NAPTR" {
		c = generatePSCreateNaptr(dnsserver, domain, rec)
		//printer.Printf("DEBUG: createNAPTR: %s\n", c)
	} else {
		c = generatePSCreate(dnsserver, domain, rec)
		//printer.Printf("DEBUG: PScreate\n")
	}

	eLog(c)
	stdout, stderr, err := psh.shell.Execute("\n\r" + c + "\n\r")
	if err != nil {
		printer.Printf("PowerShell code was:\nSTART\n%s\nEND\n", c)
		return err
	}
	if stderr != "" {
		printer.Printf("STDOUT RecordCreate = %s\n", stdout)
		printer.Printf("STDERROR RecordCreate = %q\n", stderr)
		printer.Printf("PowerShell code was:\nSTART\n%s\nEND\n", c)
		return fmt.Errorf("unexpected stderr from PSCreate: %q", stderr)
	}
	return nil
}

func generatePSCreate(dnsserver, domain string, rec *models.RecordConfig) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, `echo CREATE "%s" "%s" %q`, rec.Type, rec.Name, rec.GetTargetCombined())
	fmt.Fprintf(&b, " ; ")

	if rec.Type == "NAPTR" {
		return b.String() + generatePSCreateNaptr(dnsserver, domain, rec)
	}

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
		//printer.Printf("DEBUG TXT len = %v\n", rec.GetTargetTXTSegmentCount())
		//printer.Printf("DEBUG TXT target = %q\n", rec.GetTargetField())
		fmt.Fprintf(&b, ` -Txt -DescriptiveText %q`, rec.GetTargetTXTJoined())
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
	//printer.Printf("DEBUG PSCreate CMD = (DEBUG-START\n%s\nDEBUG-END)\n", b.String())
	return b.String()
}

func (psh *psHandle) RecordModify(dnsserver, domain string, old, rec *models.RecordConfig) error {
	c := generatePSModify(dnsserver, domain, old, rec)
	eLog(c)
	_, stderr, err := psh.shell.Execute("\n\r" + c + "\n\r")
	if err != nil {
		printer.Printf("PowerShell code was:\nSTART\n%s\nEND\n", c)
		return err
	}
	if stderr != "" {
		printer.Printf("STDERROR = %q\n", stderr)
		printer.Printf("PowerShell code was:\nSTART\n%s\nEND\n", c)
		return fmt.Errorf("unexpected stderr from PSModify: %q", stderr)
	}
	return nil
}
func generatePSModify(dnsserver, domain string, old, rec *models.RecordConfig) string {
	// The simple way is to just remove the old record and insert the new record.
	return "\n\r" + generatePSDelete(dnsserver, domain, old) + " ; " + generatePSCreate(dnsserver, domain, rec) + "\n\r"
	// NB: SOA records can't be deleted. When we implement them, we'll
	// need to special case them and generate an in-place modification
	// command.
}

func (psh *psHandle) RecordModifyTTL(dnsserver, domain string, old *models.RecordConfig, newTTL uint32) error {
	c := generatePSModifyTTL(dnsserver, domain, old, newTTL)
	eLog(c)
	_, stderr, err := psh.shell.Execute("\n\r" + c + "\n\r")
	if err != nil {
		printer.Printf("PowerShell code was:\nSTART\n%s\nEND\n", c)
		return err
	}
	if stderr != "" {
		printer.Printf("STDERROR = %q\n", stderr)
		printer.Printf("PowerShell code was:\nSTART\n%s\nEND\n", c)
		return fmt.Errorf("unexpected stderr from PSModify: %q", stderr)
	}
	return nil
}

func generatePSModifyTTL(dnsserver, domain string, rec *models.RecordConfig, newTTL uint32) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, `echo MODIFY-TTL "%s" "%s" %q ttl=%d->%d`, rec.Name, rec.Type, rec.GetTargetCombined(), rec.TTL, newTTL)
	fmt.Fprintf(&b, " ; ")

	fmt.Fprint(&b, `Get-DnsServerResourceRecord`)
	if dnsserver != "" {
		fmt.Fprintf(&b, ` -ComputerName "%s"`, dnsserver)
	}
	fmt.Fprintf(&b, ` -ZoneName "%s"`, domain)
	fmt.Fprintf(&b, ` -Name "%s"`, rec.GetLabel())
	fmt.Fprintf(&b, ` -RRType %s`, rec.Type)
	fmt.Fprint(&b, ` | ForEach-Object { $NewRecord = $_.Clone() ;`)
	fmt.Fprintf(&b, `$NewRecord.TimeToLive = New-TimeSpan -Seconds %d`, newTTL)
	fmt.Fprintf(&b, " ; ")
	fmt.Fprintf(&b, `Set-DnsServerResourceRecord`)
	if dnsserver != "" {
		fmt.Fprintf(&b, ` -ComputerName "%s"`, dnsserver)
	}
	fmt.Fprint(&b, ` -NewInputObject $NewRecord -OldInputObject $_`)
	fmt.Fprintf(&b, ` -ZoneName "%s"`, domain)

	return b.String()
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
// 	//printer.Printf("DEBUG CCMD: %s\n", b.String())
//
// 	fmt.Fprintf(&b, "Set-DnsServerResourceRecord")
// 	fmt.Fprintf(&b, ` -ZoneName "%s"`, domain)
// 	fmt.Fprintf(&b, ` -NewInputObject $NewObj -OldInputObject $OldObj`)
//
//  printer.Printf("DEBUG MCMD: %s", b.String())
// 	return b.String()
