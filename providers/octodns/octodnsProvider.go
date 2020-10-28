package octodns

/*

octodns -
  Generate zonefiles suitiable for OctoDNS.

	The zonefiles are read and written to the directory octoconfig

	If the old octoconfig files are readable, we read them to determine
	if an update is actually needed.

	The YAML input and output code is extremely complicated because
	the format does not fit well with a statically typed language.
	The YAML format changes drastically if the label has single
	or multiple rtypes associated with it, and if there is a single
	or multiple rtype data.

*/

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/providers"
	"github.com/StackExchange/dnscontrol/v3/providers/octodns/octoyaml"
)

var features = providers.DocumentationNotes{
	//providers.CanUseCAA: providers.Can(),
	providers.CanUsePTR: providers.Can(),
	providers.CanUseSRV: providers.Can(),
	//providers.CanUseTXTMulti:   providers.Can(),
	providers.DocCreateDomains: providers.Cannot("Driver just maintains list of OctoDNS config files. You must manually create the master config files that refer these."),
	providers.DocDualHost:      providers.Cannot("Research is needed."),
	providers.CanGetZones:      providers.Unimplemented(),
}

func initProvider(config map[string]string, providermeta json.RawMessage) (providers.DNSServiceProvider, error) {
	// config -- the key/values from creds.json
	// meta -- the json blob from NewReq('name', 'TYPE', meta)
	api := &octodnsProvider{
		directory: config["directory"],
	}
	if api.directory == "" {
		api.directory = "config"
	}
	if len(providermeta) != 0 {
		err := json.Unmarshal(providermeta, api)
		if err != nil {
			return nil, err
		}
	}
	//api.nameservers = models.StringsToNameservers(api.DefaultNS)
	return api, nil
}

func init() {
	providers.RegisterDomainServiceProviderType("OCTODNS", initProvider, features)
}

// octodnsProvider is the provider handle for the OctoDNS driver.
type octodnsProvider struct {
	//DefaultNS   []string `json:"default_ns"`
	//DefaultSoa  SoaInfo  `json:"default_soa"`
	//nameservers []*models.Nameserver
	directory string
}

// GetNameservers returns the nameservers for a domain.
func (c *octodnsProvider) GetNameservers(string) ([]*models.Nameserver, error) {
	return nil, nil
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (c *octodnsProvider) GetZoneRecords(domain string) (models.Records, error) {
	return nil, fmt.Errorf("not implemented")
	// This enables the get-zones subcommand.
	// Implement this by extracting the code from GetDomainCorrections into
	// a single function.  For most providers this should be relatively easy.
}

// GetDomainCorrections returns a list of corrections to update a domain.
func (c *octodnsProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	dc.Punycode()
	// Phase 1: Copy everything to []*models.RecordConfig:
	//    expectedRecords < dc.Records[i]
	//    foundRecords < zonefile
	//
	// Phase 2: Do any manipulations:
	// add NS
	// manipulate SOA
	//
	// Phase 3: Convert to []diff.Records and compare:
	// expectedDiffRecords < expectedRecords
	// foundDiffRecords < foundRecords
	// diff.Inc...(foundDiffRecords, expectedDiffRecords )

	// Read foundRecords:
	var foundRecords models.Records
	zoneFileFound := true
	zoneFileName := filepath.Join(c.directory, strings.Replace(strings.ToLower(dc.Name), "/", "_", -1)+".yaml")
	foundFH, err := os.Open(zoneFileName)
	if err != nil {
		if os.IsNotExist(err) {
			zoneFileFound = false
		} else {
			return nil, fmt.Errorf("can't open %s: %w", zoneFileName, err)
		}
	} else {
		foundRecords, err = octoyaml.ReadYaml(foundFH, dc.Name)
		if err != nil {
			return nil, fmt.Errorf("can not get corrections: %w", err)
		}
	}

	// Normalize
	models.PostProcessRecords(foundRecords)

	differ := diff.New(dc)
	_, create, del, mod, err := differ.IncrementalDiff(foundRecords)
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	// Print a list of changes. Generate an actual change that is the zone
	changes := false
	for _, i := range create {
		changes = true
		fmt.Fprintln(buf, i)
	}
	for _, i := range del {
		changes = true
		fmt.Fprintln(buf, i)
	}
	for _, i := range mod {
		changes = true
		fmt.Fprintln(buf, i)
	}
	msg := fmt.Sprintf("GENERATE_CONFIGFILE: %s", dc.Name)
	if zoneFileFound {
		msg += "\n"
		msg += buf.String()
	} else {
		msg += fmt.Sprintf(" (%d records)\n", len(create))
	}
	corrections := []*models.Correction{}
	if changes {
		corrections = append(corrections,
			&models.Correction{
				Msg: msg,
				F: func() error {
					fmt.Printf("CREATING CONFIGFILE: %v\n", zoneFileName)
					zf, err := os.Create(zoneFileName)
					if err != nil {
						log.Fatalf("Could not create zonefile: %v", err)
					}
					//err = WriteZoneFile(zf, dc.Records, dc.Name)
					err = octoyaml.WriteYaml(zf, dc.Records, dc.Name)
					if err != nil {
						log.Fatalf("WriteZoneFile error: %v\n", err)
					}
					err = zf.Close()
					if err != nil {
						log.Fatalf("Closing: %v", err)
					}
					return nil
				},
			})
	}

	return corrections, nil
}
