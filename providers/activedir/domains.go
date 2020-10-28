package activedir

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/pkg/printer"
	"github.com/TomOnTime/utfutil"
)

const zoneDumpFilenamePrefix = "adzonedump"

// RecordConfigJSON RecordConfig, reconfigured for JSON input/output.
type RecordConfigJSON struct {
	Name string `json:"hostname"`
	Type string `json:"recordtype"`
	Data string `json:"recorddata"`
	TTL  uint32 `json:"timetolive"`
}

func (c *activedirProvider) GetNameservers(string) ([]*models.Nameserver, error) {
	// TODO: If using AD for publicly hosted zones, probably pull these from config.
	return nil, nil
}

// list of types this provider supports.
// until it is up to speed with all the built-in types.
var supportedTypes = map[string]bool{
	"A":     true,
	"AAAA":  true,
	"CNAME": true,
	"NS":    true,
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (c *activedirProvider) GetZoneRecords(domain string) (models.Records, error) {
	foundRecords, err := c.getExistingRecords(domain)
	if err != nil {
		return nil, fmt.Errorf("c.getExistingRecords(%q) failed: %v", domain, err)
	}
	return foundRecords, nil
}

// GetDomainCorrections gets existing records, diffs them against existing, and returns corrections.
func (c *activedirProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {

	dc.Filter(func(r *models.RecordConfig) bool {
		if r.Type == "NS" && r.Name == "@" {
			return false
		}
		if !supportedTypes[r.Type] {
			printer.Warnf("Active Directory only manages certain record types. Won't consider %s %s\n", r.Type, r.GetLabelFQDN())
			return false
		}
		return true
	})

	// Read foundRecords:
	foundRecords, err := c.getExistingRecords(dc.Name)
	if err != nil {
		return nil, fmt.Errorf("c.getExistingRecords(%v) failed: %v", dc.Name, err)
	}

	// Normalize
	models.PostProcessRecords(foundRecords)

	differ := diff.New(dc)
	_, creates, dels, modifications, err := differ.IncrementalDiff(foundRecords)
	if err != nil {
		return nil, err
	}
	// NOTE(tlim): This provider does not delete records.  If
	// you need to delete a record, either delete it manually
	// or see providers/activedir/doc.md for implementation tips.

	// Generate changes.
	corrections := []*models.Correction{}
	for _, del := range dels {
		corrections = append(corrections, c.deleteRec(dc.Name, del))
	}
	for _, cre := range creates {
		corrections = append(corrections, c.createRec(dc.Name, cre)...)
	}
	for _, m := range modifications {
		corrections = append(corrections, c.modifyRec(dc.Name, m))
	}
	return corrections, nil

}

// zoneDumpFilename returns the filename to use to write or read
// an activedirectory zone dump for a particular domain.
func zoneDumpFilename(domainname string) string {
	return zoneDumpFilenamePrefix + "." + domainname + ".json"
}

// readZoneDump reads a pre-existing zone dump from adzonedump.*.json.
func (c *activedirProvider) readZoneDump(domainname string) ([]byte, error) {
	// File not found is considered an error.
	dat, err := utfutil.ReadFile(zoneDumpFilename(domainname), utfutil.WINDOWS)
	if err != nil {
		printer.Printf("Powershell to generate zone dump:\n")
		printer.Printf("%v\n", c.generatePowerShellZoneDump(domainname))
	}
	return dat, err
}

// powerShellLogCommand logs to flagPsLog that a PowerShell command is going to be run.
func (c *activedirProvider) logCommand(command string) error {
	return c.logHelper(fmt.Sprintf("# %s\r\n%s\r\n", time.Now().UTC(), strings.TrimSpace(command)))
}

// powerShellLogOutput logs to flagPsLog that a PowerShell command is going to be run.
func (c *activedirProvider) logOutput(s string) error {
	return c.logHelper(fmt.Sprintf("OUTPUT: START\r\n%s\r\nOUTPUT: END\r\n", s))
}

// powerShellLogErr logs that a PowerShell command had an error.
func (c *activedirProvider) logErr(e error) error {
	err := c.logHelper(fmt.Sprintf("ERROR: %v\r\r", e)) // Log error to powershell.log
	if err != nil {
		return err // Bubble up error created in logHelper
	}
	return e // Bubble up original error
}

func (c *activedirProvider) logHelper(s string) error {
	logfile, err := os.OpenFile(c.psLog, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0660)
	if err != nil {
		return fmt.Errorf("error: Can not create/append to %#v: %v", c.psLog, err)
	}
	_, err = fmt.Fprintln(logfile, s)
	if err != nil {
		return fmt.Errorf("append to %#v failed: %v", c.psLog, err)
	}
	if logfile.Close() != nil {
		return fmt.Errorf("closing %#v failed: %v", c.psLog, err)
	}
	return nil
}

// powerShellRecord records that a PowerShell command should be executed later.
func (c *activedirProvider) powerShellRecord(command string) error {
	recordfile, err := os.OpenFile(c.psOut, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0660)
	if err != nil {
		return fmt.Errorf("can not create/append to %#v: %v", c.psOut, err)
	}
	_, err = recordfile.WriteString(command)
	if err != nil {
		return fmt.Errorf("append to %#v failed: %v", c.psOut, err)
	}
	return recordfile.Close()
}

func (c *activedirProvider) getExistingRecords(domainname string) ([]*models.RecordConfig, error) {
	// Get the JSON either from adzonedump or by running a PowerShell script.
	data, err := c.getRecords(domainname)
	if err != nil {
		return nil, fmt.Errorf("getRecords failed on %#v: %v", domainname, err)
	}

	var recs []*RecordConfigJSON
	jdata := string(data)
	// when there is only a single record, AD powershell does not
	// wrap it in an array as our types expect. This makes sure it is always an array.
	if strings.HasPrefix(strings.TrimSpace(jdata), "{") {
		jdata = "[" + jdata + "]"
		data = []byte(jdata)
	}
	err = json.Unmarshal(data, &recs)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal failed on %#v: %v", domainname, err)
	}

	result := make([]*models.RecordConfig, 0, len(recs))
	unsupportedCounts := map[string]int{}
	for _, rec := range recs {
		t, supportedType := rec.unpackRecord(domainname)
		if !supportedType {
			unsupportedCounts[rec.Type]++
		}
		if t != nil {
			result = append(result, t)
		}
	}
	for t, count := range unsupportedCounts {
		printer.Warnf("%d records of type %s found in AD zone. These will be ignored.\n", count, t)
	}

	return result, nil
}

func (r *RecordConfigJSON) unpackRecord(origin string) (rc *models.RecordConfig, supported bool) {
	rc = &models.RecordConfig{
		Type: r.Type,
		TTL:  r.TTL,
	}
	rc.SetLabel(r.Name, origin)
	switch rtype := rc.Type; rtype { // #rtype_variations
	case "A", "AAAA":
		rc.SetTarget(r.Data)
	case "CNAME":
		rc.SetTarget(strings.ToLower(r.Data))
	case "NS":
		// skip root NS
		if rc.Name == "@" {
			return nil, true
		}
		rc.SetTarget(strings.ToLower(r.Data))
	case "SOA":
		return nil, true
	default:
		return nil, false
	}
	return rc, true
}

// powerShellDump runs a PowerShell command to get a dump of all records in a DNS zone.
func (c *activedirProvider) generatePowerShellZoneDump(domainname string) string {
	cmdTxt := `@("REPLACE_WITH_ZONE") | %{
Get-DnsServerResourceRecord -ComputerName REPLACE_WITH_COMPUTER_NAME -ZoneName $_ | select hostname,recordtype,@{n="timestamp";e={$_.timestamp.tostring()}},@{n="timetolive";e={$_.timetolive.totalseconds}},@{n="recorddata";e={($_.recorddata.ipv4address,$_.recorddata.ipv6address,$_.recorddata.HostNameAlias,$_.recorddata.NameServer,"unsupported_record_type" -ne $null)[0]-as [string]}} | ConvertTo-Json > REPLACE_WITH_FILENAMEPREFIX.REPLACE_WITH_ZONE.json
}`
	cmdTxt = strings.Replace(cmdTxt, "REPLACE_WITH_ZONE", domainname, -1)
	cmdTxt = strings.Replace(cmdTxt, "REPLACE_WITH_COMPUTER_NAME", c.adServer, -1)
	cmdTxt = strings.Replace(cmdTxt, "REPLACE_WITH_FILENAMEPREFIX", zoneDumpFilenamePrefix, -1)
	return cmdTxt
}

// generatePowerShellCreate generates PowerShell commands to ADD a record.
func (c *activedirProvider) generatePowerShellCreate(domainname string, rec *models.RecordConfig) string {
	content := rec.GetTargetField()
	text := "\r\n" // Skip a line.
	funcSuffix := rec.Type
	if rec.Type == "NS" {
		funcSuffix = ""
	}
	text += fmt.Sprintf("Add-DnsServerResourceRecord%s", funcSuffix)
	text += fmt.Sprintf(` -ComputerName "%s"`, c.adServer)
	text += fmt.Sprintf(` -ZoneName "%s"`, domainname)
	text += fmt.Sprintf(` -Name "%s"`, rec.GetLabel())
	text += fmt.Sprintf(` -TimeToLive $(New-TimeSpan -Seconds %d)`, rec.TTL)
	switch rec.Type { // #rtype_variations
	case "CNAME":
		text += fmt.Sprintf(` -HostNameAlias "%s"`, content)
	case "A":
		text += fmt.Sprintf(` -IPv4Address "%s"`, content)
	case "NS":
		text += fmt.Sprintf(` -NS -NameServer "%s"`, content)
	default:
		panic(fmt.Errorf("generatePowerShellCreate() does not yet handle recType=%s recName=%#v content=%#v)",
			rec.Type, rec.GetLabel(), content))
		// We panic so that we quickly find any switch statements
		// that have not been updated for a new RR type.
	}
	text += "\r\n"

	return text
}

// generatePowerShellModify generates PowerShell commands to MODIFY a record.
func (c *activedirProvider) generatePowerShellModify(domainname, recName, recType, oldContent, newContent string, oldTTL, newTTL uint32) string {

	var queryField, queryContent string
	queryContent = `"` + oldContent + `"`

	switch recType { // #rtype_variations
	case "A":
		queryField = "IPv4address"
	case "CNAME":
		queryField = "HostNameAlias"
	case "NS":
		queryField = "NameServer"
	default:
		panic(fmt.Errorf("generatePowerShellModify() does not yet handle recType=%s recName=%#v content=(%#v, %#v)", recType, recName, oldContent, newContent))
		// We panic so that we quickly find any switch statements
		// that have not been updated for a new RR type.
	}

	text := "\r\n" // Skip a line.
	text += fmt.Sprintf(`echo "MODIFY %s %s %s old=%s new=%s"`, recName, domainname, recType, oldContent, newContent)
	text += "\r\n"

	text += "$OldObj = Get-DnsServerResourceRecord"
	text += fmt.Sprintf(` -ComputerName "%s"`, c.adServer)
	text += fmt.Sprintf(` -ZoneName "%s"`, domainname)
	text += fmt.Sprintf(` -Name "%s"`, recName)
	text += fmt.Sprintf(` -RRType "%s"`, recType)
	text += fmt.Sprintf(" |  Where-Object {$_.RecordData.%s -eq %s -and $_.HostName -eq \"%s\"}", queryField, queryContent, recName)
	text += "\r\n"
	text += `if($OldObj.Length -ne $null){ throw "Error, multiple results for Get-DnsServerResourceRecord" }`
	text += "\r\n"

	text += "$NewObj = $OldObj.Clone()"
	text += "\r\n"

	if oldContent != newContent {
		text += fmt.Sprintf(`$NewObj.RecordData.%s = "%s"`, queryField, newContent)
		text += "\r\n"
	}

	if oldTTL != newTTL {
		text += fmt.Sprintf(`$NewObj.TimeToLive = New-TimeSpan -Seconds %d`, newTTL)
		text += "\r\n"
	}

	text += "Set-DnsServerResourceRecord"
	text += fmt.Sprintf(` -ComputerName "%s"`, c.adServer)
	text += fmt.Sprintf(` -ZoneName "%s"`, domainname)
	text += ` -NewInputObject $NewObj -OldInputObject $OldObj`
	text += "\r\n"

	return text
}

func (c *activedirProvider) generatePowerShellDelete(domainname, recName, recType, content string) string {
	text := fmt.Sprintf(`echo "DELETE %s %s %s"`, recType, recName, content)
	text += "\r\n"
	text += `Remove-DnsServerResourceRecord -Force -ComputerName "%s" -ZoneName "%s" -Name "%s" -RRType "%s" -RecordData "%s"`
	text += "\r\n"
	return fmt.Sprintf(text, c.adServer, domainname, recName, recType, content)
}

func (c *activedirProvider) createRec(domainname string, cre diff.Correlation) []*models.Correction {
	rec := cre.Desired
	arr := []*models.Correction{
		{
			Msg: cre.String(),
			F: func() error {
				return c.powerShellDoCommand(c.generatePowerShellCreate(domainname, rec), true)
			}},
	}
	return arr
}

func (c *activedirProvider) modifyRec(domainname string, m diff.Correlation) *models.Correction {
	old, rec := m.Existing, m.Desired
	return &models.Correction{
		Msg: m.String(),
		F: func() error {
			return c.powerShellDoCommand(c.generatePowerShellModify(domainname, rec.GetLabel(), rec.Type, old.GetTargetField(), rec.GetTargetField(), old.TTL, rec.TTL), true)
		},
	}
}

func (c *activedirProvider) deleteRec(domainname string, cor diff.Correlation) *models.Correction {
	rec := cor.Existing
	return &models.Correction{
		Msg: cor.String(),
		F: func() error {
			return c.powerShellDoCommand(c.generatePowerShellDelete(domainname, rec.GetLabel(), rec.Type, rec.GetTargetField()), true)
		},
	}
}
