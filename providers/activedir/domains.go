package activedir

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers/diff"
	"github.com/TomOnTime/utfutil"
	"github.com/miekg/dns/dnsutil"
	"github.com/pkg/errors"
)

const zoneDumpFilenamePrefix = "adzonedump"

// RecordConfigJson RecordConfig, reconfigured for JSON input/output.
type RecordConfigJson struct {
	Name string `json:"hostname"`
	Type string `json:"recordtype"`
	Data string `json:"recorddata"`
	TTL  uint32 `json:"timetolive"`
}

func (c *adProvider) GetNameservers(string) ([]*models.Nameserver, error) {
	// TODO: If using AD for publicly hosted zones, probably pull these from config.
	return nil, nil
}

// GetDomainCorrections gets existing records, diffs them against existing, and returns corrections.
func (c *adProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {

	dc.Filter(func(r *models.RecordConfig) bool {
		if r.Type != "A" && r.Type != "CNAME" {
			log.Printf("WARNING: Active Directory only manages A and CNAME records. Won't consider %s %s", r.Type, r.NameFQDN)
			return false
		}
		return true
	})

	// Read foundRecords:
	foundRecords, err := c.getExistingRecords(dc.Name)
	if err != nil {
		return nil, errors.Errorf("c.getExistingRecords(%v) failed: %v", dc.Name, err)
	}

	// Normalize
	models.PostProcessRecords(foundRecords)

	differ := diff.New(dc)
	_, creates, dels, modifications := differ.IncrementalDiff(foundRecords)
	// NOTE(tlim): This provider does not delete records.  If
	// you need to delete a record, either delete it manually
	// or see providers/activedir/doc.md for implementation tips.

	// Generate changes.
	corrections := []*models.Correction{}
	for _, del := range dels {
		corrections = append(corrections, c.deleteRec(dc.Name, del.Existing))
	}
	for _, cre := range creates {
		corrections = append(corrections, c.createRec(dc.Name, cre.Desired)...)
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
func (c *adProvider) readZoneDump(domainname string) ([]byte, error) {
	// File not found is considered an error.
	dat, err := utfutil.ReadFile(zoneDumpFilename(domainname), utfutil.WINDOWS)
	if err != nil {
		fmt.Println("Powershell to generate zone dump:")
		fmt.Println(c.generatePowerShellZoneDump(domainname))
	}
	return dat, err
}

// powerShellLogCommand logs to flagPsLog that a PowerShell command is going to be run.
func (c *adProvider) logCommand(command string) error {
	return c.logHelper(fmt.Sprintf("# %s\r\n%s\r\n", time.Now().UTC(), strings.TrimSpace(command)))
}

// powerShellLogOutput logs to flagPsLog that a PowerShell command is going to be run.
func (c *adProvider) logOutput(s string) error {
	return c.logHelper(fmt.Sprintf("OUTPUT: START\r\n%s\r\nOUTPUT: END\r\n", s))
}

// powerShellLogErr logs that a PowerShell command had an error.
func (c *adProvider) logErr(e error) error {
	err := c.logHelper(fmt.Sprintf("ERROR: %v\r\r", e)) // Log error to powershell.log
	if err != nil {
		return err // Bubble up error created in logHelper
	}
	return e // Bubble up original error
}

func (c *adProvider) logHelper(s string) error {
	logfile, err := os.OpenFile(c.psLog, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0660)
	if err != nil {
		return errors.Errorf("error: Can not create/append to %#v: %v", c.psLog, err)
	}
	_, err = fmt.Fprintln(logfile, s)
	if err != nil {
		return errors.Errorf("Append to %#v failed: %v", c.psLog, err)
	}
	if logfile.Close() != nil {
		return errors.Errorf("Closing %#v failed: %v", c.psLog, err)
	}
	return nil
}

// powerShellRecord records that a PowerShell command should be executed later.
func (c *adProvider) powerShellRecord(command string) error {
	recordfile, err := os.OpenFile(c.psOut, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0660)
	if err != nil {
		return errors.Errorf("can not create/append to %#v: %v", c.psOut, err)
	}
	_, err = recordfile.WriteString(command)
	if err != nil {
		return errors.Errorf("append to %#v failed: %v", c.psOut, err)
	}
	return recordfile.Close()
}

func (c *adProvider) getExistingRecords(domainname string) ([]*models.RecordConfig, error) {
	// log.Printf("getExistingRecords(%s)\n", domainname)

	// Get the JSON either from adzonedump or by running a PowerShell script.
	data, err := c.getRecords(domainname)
	if err != nil {
		return nil, errors.Errorf("getRecords failed on %#v: %v", domainname, err)
	}

	var recs []*RecordConfigJson
	err = json.Unmarshal(data, &recs)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed on %#v: %v", domainname, err)
	}

	result := make([]*models.RecordConfig, 0, len(recs))
	for i := range recs {
		t := recs[i].unpackRecord(domainname)
		if t != nil {
			result = append(result, t)
		}
	}

	return result, nil
}

func (r *RecordConfigJson) unpackRecord(origin string) *models.RecordConfig {
	rc := models.RecordConfig{}

	rc.Name = strings.ToLower(r.Name)
	rc.NameFQDN = dnsutil.AddOrigin(rc.Name, origin)
	rc.Type = r.Type
	rc.TTL = r.TTL

	switch rc.Type { // #rtype_variations
	case "A":
		rc.Target = r.Data
	case "CNAME":
		rc.Target = strings.ToLower(r.Data)
	case "NS", "SOA":
		return nil
	default:
		log.Printf("Warning: Record of type %s found in AD zone. Will be ignored.", rc.Type)
		return nil
	}
	return &rc
}

// powerShellDump runs a PowerShell command to get a dump of all records in a DNS zone.
func (c *adProvider) generatePowerShellZoneDump(domainname string) string {
	cmdTxt := `@("REPLACE_WITH_ZONE") | %{
Get-DnsServerResourceRecord -ComputerName REPLACE_WITH_COMPUTER_NAME -ZoneName $_ | select hostname,recordtype,@{n="timestamp";e={$_.timestamp.tostring()}},@{n="timetolive";e={$_.timetolive.totalseconds}},@{n="recorddata";e={($_.recorddata.ipv4address,$_.recorddata.ipv6address,$_.recorddata.HostNameAlias,"other_record" -ne $null)[0]-as [string]}} | ConvertTo-Json > REPLACE_WITH_FILENAMEPREFIX.REPLACE_WITH_ZONE.json
}`
	cmdTxt = strings.Replace(cmdTxt, "REPLACE_WITH_ZONE", domainname, -1)
	cmdTxt = strings.Replace(cmdTxt, "REPLACE_WITH_COMPUTER_NAME", c.adServer, -1)
	cmdTxt = strings.Replace(cmdTxt, "REPLACE_WITH_FILENAMEPREFIX", zoneDumpFilenamePrefix, -1)

	return cmdTxt
}

// generatePowerShellCreate generates PowerShell commands to ADD a record.
func (c *adProvider) generatePowerShellCreate(domainname string, rec *models.RecordConfig) string {
	content := rec.Target
	text := "\r\n" // Skip a line.
	text += fmt.Sprintf("Add-DnsServerResourceRecord%s", rec.Type)
	text += fmt.Sprintf(` -ComputerName "%s"`, c.adServer)
	text += fmt.Sprintf(` -ZoneName "%s"`, domainname)
	text += fmt.Sprintf(` -Name "%s"`, rec.Name)
	text += fmt.Sprintf(` -TimeToLive $(New-TimeSpan -Seconds %d)`, rec.TTL)
	switch rec.Type { // #rtype_variations
	case "CNAME":
		text += fmt.Sprintf(` -HostNameAlias "%s"`, content)
	case "A":
		text += fmt.Sprintf(` -IPv4Address "%s"`, content)
	case "NS":
		text = fmt.Sprintf("\r\n"+`echo "Skipping NS update (%v %v)"`+"\r\n", rec.Name, rec.Target)
	default:
		panic(errors.Errorf("generatePowerShellCreate() does not yet handle recType=%s recName=%#v content=%#v)", rec.Type, rec.Name, content))
		// We panic so that we quickly find any switch statements
		// that have not been updated for a new RR type.
	}
	text += "\r\n"

	return text
}

// generatePowerShellModify generates PowerShell commands to MODIFY a record.
func (c *adProvider) generatePowerShellModify(domainname, recName, recType, oldContent, newContent string, oldTTL, newTTL uint32) string {

	var queryField, queryContent string

	switch recType { // #rtype_variations
	case "A":
		queryField = "IPv4address"
		queryContent = `"` + oldContent + `"`
	case "CNAME":
		queryField = "HostNameAlias"
		queryContent = `"` + oldContent + `"`
	default:
		panic(errors.Errorf("generatePowerShellModify() does not yet handle recType=%s recName=%#v content=(%#v, %#v)", recType, recName, oldContent, newContent))
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
	text += fmt.Sprintf(` -NewInputObject $NewObj -OldInputObject $OldObj`)
	text += "\r\n"

	return text
}

func (c *adProvider) generatePowerShellDelete(domainname, recName, recType, content string) string {
	text := fmt.Sprintf(`echo "DELETE %s %s %s"`, recType, recName, content)
	text += "\r\n"
	text += `Remove-DnsServerResourceRecord -Force -ComputerName "%s" -ZoneName "%s" -Name "%s" -RRType "%s" -RecordData "%s"`
	text += "\r\n"
	return fmt.Sprintf(text, c.adServer, domainname, recName, recType, content)
}

func (c *adProvider) createRec(domainname string, rec *models.RecordConfig) []*models.Correction {
	arr := []*models.Correction{
		{
			Msg: fmt.Sprintf("CREATE record: %s %s ttl(%d) %s", rec.Name, rec.Type, rec.TTL, rec.Target),
			F: func() error {
				return c.powerShellDoCommand(c.generatePowerShellCreate(domainname, rec), true)
			}},
	}
	return arr
}

func (c *adProvider) modifyRec(domainname string, m diff.Correlation) *models.Correction {
	old, rec := m.Existing, m.Desired
	return &models.Correction{
		Msg: m.String(),
		F: func() error {
			return c.powerShellDoCommand(c.generatePowerShellModify(domainname, rec.Name, rec.Type, old.Target, rec.Target, old.TTL, rec.TTL), true)
		},
	}
}

func (c *adProvider) deleteRec(domainname string, rec *models.RecordConfig) *models.Correction {
	return &models.Correction{
		Msg: fmt.Sprintf("DELETE record: %s %s ttl(%d) %s", rec.Name, rec.Type, rec.TTL, rec.Target),
		F: func() error {
			return c.powerShellDoCommand(c.generatePowerShellDelete(domainname, rec.Name, rec.Type, rec.Target), true)
		},
	}
}
