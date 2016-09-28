package activedir

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/TomOnTime/utfutil"
	"github.com/miekg/dns/dnsutil"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers/diff"
)

const zoneDumpFilenamePrefix = "adzonedump"

type RecordConfigJson struct {
	Name string `json:"hostname"`
	Type string `json:"recordtype"`
	Data string `json:"recorddata"`
	TTL  uint32 `json:"timetolive"`
}

// GetDomainCorrections gets existing records, diffs them against existing, and returns corrections.
func (c *adProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {

	// Read foundRecords:
	foundRecords, err := c.getExistingRecords(dc.Name)
	if err != nil {
		return nil, fmt.Errorf("c.getExistingRecords(%v) failed: %v", dc.Name, err)
	}

	// Read expectedRecords:
	//expectedRecords := make([]*models.RecordConfig, len(dc.Records))
	expectedRecords := make([]diff.Record, len(dc.Records))
	for i, r := range dc.Records {
		if r.TTL == 0 {
			r.TTL = models.DefaultTTL
		}
		expectedRecords[i] = r
	}

	// Convert to []diff.Records and compare:
	foundDiffRecords := make([]diff.Record, 0, len(foundRecords))
	for _, rec := range foundRecords {
		foundDiffRecords = append(foundDiffRecords, rec)
	}

	_, creates, dels, modifications := diff.IncrementalDiff(foundDiffRecords, expectedRecords)
	// NOTE(tlim): This provider does not delete records.  If
	// you need to delete a record, either delete it manually
	// or see providers/activedir/doc.md for implementation tips.

	// Generate changes.
	corrections := []*models.Correction{}
	for _, cre := range creates {
		corrections = append(corrections, c.createRec(dc.Name, cre.Desired.(*models.RecordConfig))...)
	}
	for _, m := range modifications {
		corrections = append(corrections, c.modifyRec(dc.Name, m))
	}
	for _, del := range dels {
		corrections = append(corrections, c.deleteRec(dc.Name, del.Existing.(*models.RecordConfig))
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
func powerShellLogCommand(command string) error {
	return logHelper(fmt.Sprintf("# %s\r\n%s\r\n", time.Now().UTC(), strings.TrimSpace(command)))
}

// powerShellLogOutput logs to flagPsLog that a PowerShell command is going to be run.
func powerShellLogOutput(s string) error {
	return logHelper(fmt.Sprintf("OUTPUT: START\r\n%s\r\nOUTPUT: END\r\n", s))
}

// powerShellLogErr logs that a PowerShell command had an error.
func powerShellLogErr(e error) error {
	err := logHelper(fmt.Sprintf("ERROR: %v\r\r", e)) //Log error to powershell.log
	if err != nil {
		return err //Bubble up error created in logHelper
	}
	return e //Bubble up original error
}

func logHelper(s string) error {
	logfile, err := os.OpenFile(*flagPsLog, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0660)
	if err != nil {
		return fmt.Errorf("ERROR: Can not create/append to %#v: %v\n", *flagPsLog, err)
	}
	_, err = fmt.Fprintln(logfile, s)
	if err != nil {
		return fmt.Errorf("ERROR: Append to %#v failed: %v\n", *flagPsLog, err)
	}
	if logfile.Close() != nil {
		return fmt.Errorf("ERROR: Closing %#v failed: %v\n", *flagPsLog, err)
	}
	return nil
}

// powerShellRecord records that a PowerShell command should be executed later.
func powerShellRecord(command string) error {
	recordfile, err := os.OpenFile(*flagPsFuture, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0660)
	if err != nil {
		return fmt.Errorf("ERROR: Can not create/append to %#v: %v\n", *flagPsFuture, err)
	}
	_, err = recordfile.WriteString(command)
	if err != nil {
		return fmt.Errorf("ERROR: Append to %#v failed: %v\n", *flagPsFuture, err)
	}
	return recordfile.Close()
}

func (c *adProvider) getExistingRecords(domainname string) ([]*models.RecordConfig, error) {
	//log.Printf("getExistingRecords(%s)\n", domainname)

	// Get the JSON either from adzonedump or by running a PowerShell script.
	data, err := c.getRecords(domainname)
	if err != nil {
		return nil, fmt.Errorf("getRecords failed on %#v: %v\n", domainname, err)
	}

	var recs []*RecordConfigJson
	err = json.Unmarshal(data, &recs)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal failed on %#v: %v\n", domainname, err)
	}

	result := make([]*models.RecordConfig, 0, len(recs))
	for i := range recs {
		t, err := recs[i].unpackRecord(domainname)
		if err == nil {
			result = append(result, t)
		}
	}

	return result, nil
}

func (r *RecordConfigJson) unpackRecord(origin string) (*models.RecordConfig, error) {
	rc := models.RecordConfig{}

	rc.Name = strings.ToLower(r.Name)
	rc.NameFQDN = dnsutil.AddOrigin(rc.Name, origin)
	rc.Type = r.Type
	rc.TTL = r.TTL

	switch rc.Type {
	case "A":
		rc.Target = r.Data
	case "CNAME":
		rc.Target = strings.ToLower(r.Data)
	case "AAAA", "MX", "NAPTR", "NS", "SOA", "SRV":
		return nil, fmt.Errorf("Unimplemented: %v", r.Type)
	default:
		log.Fatalf("Unhandled models.RecordConfigJson type: %v (%v)\n", rc.Type, r)
	}

	return &rc, nil
}

// powerShellDump runs a PowerShell command to get a dump of all records in a DNS zone.
func (c *adProvider) generatePowerShellZoneDump(domainname string) string {
	cmd_txt := `@("REPLACE_WITH_ZONE") | %{
Get-DnsServerResourceRecord -ComputerName REPLACE_WITH_COMPUTER_NAME -ZoneName $_ | select hostname,recordtype,@{n="timestamp";e={$_.timestamp.tostring()}},@{n="timetolive";e={$_.timetolive.totalseconds}},@{n="recorddata";e={($_.recorddata.ipv4address,$_.recorddata.ipv6address,$_.recorddata.HostNameAlias,"other_record" -ne $null)[0]-as [string]}} | ConvertTo-Json > REPLACE_WITH_FILENAMEPREFIX.REPLACE_WITH_ZONE.json
}`
	cmd_txt = strings.Replace(cmd_txt, "REPLACE_WITH_ZONE", domainname, -1)
	cmd_txt = strings.Replace(cmd_txt, "REPLACE_WITH_COMPUTER_NAME", c.adServer, -1)
	cmd_txt = strings.Replace(cmd_txt, "REPLACE_WITH_FILENAMEPREFIX", zoneDumpFilenamePrefix, -1)

	return cmd_txt
}

// generatePowerShellCreate generates PowerShell commands to ADD a record.
func (c *adProvider) generatePowerShellCreate(domainname string, rec *models.RecordConfig) string {

	content := rec.Target

	text := "\r\n" // Skip a line.
	text += fmt.Sprintf("Add-DnsServerResourceRecord%s", rec.Type)
	text += fmt.Sprintf(` -ComputerName "%s"`, c.adServer)
	text += fmt.Sprintf(` -ZoneName "%s"`, domainname)
	text += fmt.Sprintf(` -Name "%s"`, rec.Name)
	switch rec.Type {
	case "CNAME":
		text += fmt.Sprintf(` -HostNameAlias "%s"`, content)
	case "A":
		text += fmt.Sprintf(` -IPv4Address "%s"`, content)
	case "NS":
		text = fmt.Sprintf("\r\n"+`echo "Skipping NS update (%v %v)"`+"\r\n", rec.Name, rec.Target)
	default:
		panic(fmt.Errorf("ERROR: generatePowerShellCreate() does not yet handle recType=%s recName=%#v content=%#v)\n", rec.Type, rec.Name, content))
	}
	text += "\r\n"

	return text
}

// generatePowerShellModify generates PowerShell commands to MODIFY a record.
func (c *adProvider) generatePowerShellModify(domainname, recName, recType, oldContent, newContent string, oldTTL, newTTL uint32) string {

	var queryField, queryContent string

	switch recType {
	case "A":
		queryField = "IPv4address"
		queryContent = `"` + oldContent + `"`
	case "CNAME":
		queryField = "HostNameAlias"
		queryContent = `"` + oldContent + `"`
	default:
		panic(fmt.Errorf("ERROR: generatePowerShellModify() does not yet handle recType=%s recName=%#v content=(%#v, %#v)\n", recType, recName, oldContent, newContent))
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

func (c *adProvider) generatePowerShellDelete(domainname, recName, recType) string {
	text := `# Remove-DnsServerResourceRecord -ComputerName "%s" -ZoneName "%s" -Name "%s" -RRType "%s"` //comment for now
	return fmt.Sprintf(text, c.adServer, domainname, recName, recType)
}

func (c *adProvider) createRec(domainname string, rec *models.RecordConfig) []*models.Correction {
	arr := []*models.Correction{
		{
			Msg: fmt.Sprintf("CREATE record: %s %s ttl(%d) %s", rec.Name, rec.Type, rec.TTL, rec.Target),
			F: func() error {
				return powerShellDoCommand(c.generatePowerShellCreate(domainname, rec))
			}},
	}
	return arr
}

func (c *adProvider) modifyRec(domainname string, m diff.Correlation) *models.Correction {

	old, rec := m.Existing.(*models.RecordConfig), m.Desired.(*models.RecordConfig)
	oldContent := old.GetContent()
	newContent := rec.GetContent()

	return &models.Correction{
		Msg: m.String(),
		F: func() error {
			return powerShellDoCommand(c.generatePowerShellModify(domainname, rec.Name, rec.Type, oldContent, newContent, old.TTL, rec.TTL))
		},
	}
}

func (c *adProvider) deleteRec(domainname string, rec *models.RecordConfig) *models.Correction {
	return &models.Correction{
		Msg: fmt.Sprintf("DELETE record: %s %s ttl(%d) %s", rec.Name, rec.Type, rec.TTL, rec.Target),
		F: func() error {
			return powerShellDoCommand(c.generatePowerShellDelete(domainname, rec.Name, rec.Type))
		},
	}
}