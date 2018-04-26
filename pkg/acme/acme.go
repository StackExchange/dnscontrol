// Package acme provides a means of performing Let's Encrypt DNS challenges via a DNSConfig
package acme

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/pkg/nameservers"
	"github.com/xenolf/lego/acmev2"
)

type CertConfig struct {
	CertName string   `json:"cert_name"`
	Names    []string `json:"names"`
}

type Client interface {
	IssueOrRenewCert(config *CertConfig, renewUnder int, verbose bool) (bool, error)
}

type certManager struct {
	directory      string
	email          string
	acmeDirectory  string
	acmeHost       string
	cfg            *models.DNSConfig
	checkedDomains map[string]bool

	account *account
	client  *acme.Client
}

const (
	LetsEncryptLive  = "https://acme-v02.api.letsencrypt.org/directory"
	LetsEncryptStage = "https://acme-staging-v02.api.letsencrypt.org/directory"
)

func New(cfg *models.DNSConfig, directory string, email string, server string) (Client, error) {
	u, err := url.Parse(server)
	if err != nil || u.Host == "" {
		return nil, fmt.Errorf("ACME directory '%s' is not a valid URL", server)
	}
	c := &certManager{
		directory:      directory,
		email:          email,
		acmeDirectory:  server,
		acmeHost:       u.Host,
		cfg:            cfg,
		checkedDomains: map[string]bool{},
	}

	if err := c.loadOrCreateAccount(); err != nil {
		return nil, err
	}
	c.client.ExcludeChallenges([]acme.Challenge{acme.HTTP01})
	c.client.SetChallengeProvider(acme.DNS01, c)
	return c, nil
}

// IssueOrRenewCert will obtain a certificate with the given name if it does not exist,
// or renew it if it is close enough to the expiration date.
// It will return true if it issued or updated the certificate.
func (c *certManager) IssueOrRenewCert(cfg *CertConfig, renewUnder int, verbose bool) (bool, error) {
	if !verbose {
		acme.Logger = log.New(ioutil.Discard, "", 0)
	}

	log.Printf("Checking certificate [%s]", cfg.CertName)
	if err := os.MkdirAll(filepath.Dir(c.certFile(cfg.CertName, "json")), perms); err != nil {
		return false, err
	}
	existing, err := c.readCertificate(cfg.CertName)
	if err != nil {
		return false, err
	}

	var action = func() (acme.CertificateResource, error) {
		return c.client.ObtainCertificate(cfg.Names, true, nil, true)
	}

	if existing == nil {
		log.Println("No existing cert found. Issuing new...")
	} else {
		names, daysLeft, err := getCertInfo(existing.Certificate)
		if err != nil {
			return false, err
		}
		log.Printf("Found existing cert. %0.2f days remaining.", daysLeft)
		namesOK := dnsNamesEqual(cfg.Names, names)
		if daysLeft >= float64(renewUnder) && namesOK {
			log.Println("Nothing to do")
			//nothing to do
			return false, nil
		}
		if !namesOK {
			log.Println("DNS Names don't match expected set. Reissuing.")
		} else {
			log.Println("Renewing cert")
			action = func() (acme.CertificateResource, error) {
				return c.client.RenewCertificate(*existing, true, true)
			}
		}
	}

	certResource, err := action()
	if err != nil {
		return false, err
	}
	fmt.Printf("Obtained certificate for %s\n", cfg.CertName)
	return true, c.writeCertificate(cfg.CertName, &certResource)
}

// filename for certifiacte / key / json file
func (c *certManager) certFile(name, ext string) string {
	return filepath.Join(c.directory, "certificates", name, name+"."+ext)
}

func (c *certManager) writeCertificate(name string, cr *acme.CertificateResource) error {
	jDAt, err := json.MarshalIndent(cr, "", "  ")
	if err != nil {
		return err
	}
	if err = ioutil.WriteFile(c.certFile(name, "json"), jDAt, perms); err != nil {
		return err
	}
	if err = ioutil.WriteFile(c.certFile(name, "crt"), cr.Certificate, perms); err != nil {
		return err
	}
	return ioutil.WriteFile(c.certFile(name, "key"), cr.PrivateKey, perms)
}

func getCertInfo(pemBytes []byte) (names []string, remaining float64, err error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, 0, fmt.Errorf("Invalid certificate pem data")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, 0, err
	}
	var daysLeft = float64(cert.NotAfter.Sub(time.Now())) / float64(time.Hour*24)
	return cert.DNSNames, daysLeft, nil
}

// checks two lists of sans to make sure they have all the same names in them.
func dnsNamesEqual(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	sort.Strings(a)
	sort.Strings(b)
	for i, s := range a {
		if b[i] != s {
			return false
		}
	}
	return true
}

func (c *certManager) readCertificate(name string) (*acme.CertificateResource, error) {
	f, err := os.Open(c.certFile(name, "json"))
	if err != nil && os.IsNotExist(err) {
		// if json does not exist, nothing does
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	cr := &acme.CertificateResource{}
	if err = dec.Decode(cr); err != nil {
		return nil, err
	}
	// load cert
	crtBytes, err := ioutil.ReadFile(c.certFile(name, "crt"))
	if err != nil {
		return nil, err
	}
	cr.Certificate = crtBytes
	return cr, nil
}

func (c *certManager) Present(domain, token, keyAuth string) (e error) {
	d := c.cfg.DomainContainingFQDN(domain)
	// fix NS records for this domain's DNS providers
	// only need to do this once per domain
	const metaKey = "x-fixed-nameservers"
	if d.Metadata[metaKey] == "" {
		nsList, err := nameservers.DetermineNameservers(d)
		if err != nil {
			return err
		}
		d.Nameservers = nsList
		nameservers.AddNSRecords(d)
		d.Metadata[metaKey] = "true"
	}
	// copy now so we can add txt record safely, and just run unmodified version later to cleanup
	d, err := d.Copy()
	if err != nil {
		return err
	}
	if err := c.ensureNoPendingCorrections(d); err != nil {
		return err
	}
	fqdn, val, _ := acme.DNS01Record(domain, keyAuth)
	fmt.Println(fqdn, val)
	txt := &models.RecordConfig{Type: "TXT"}
	txt.SetTargetTXT(val)
	txt.SetLabelFromFQDN(fqdn, d.Name)
	d.Records = append(d.Records, txt)

	return getAndRunCorrections(d)
}

func (c *certManager) ensureNoPendingCorrections(d *models.DomainConfig) error {
	// only need to check a domain once per app run
	if c.checkedDomains[d.Name] {
		return nil
	}
	corrections, err := getCorrections(d)
	if err != nil {
		return err
	}
	if len(corrections) != 0 {
		// TODO: maybe allow forcing through this check.
		for _, c := range corrections {
			fmt.Println(c.Msg)
		}
		return fmt.Errorf("Found %d pending corrections for %s. Not going to proceed issuing certificates", len(corrections), d.Name)
	}
	return nil
}

// IgnoredProviders is a lit of provider names that should not be used to fill challenges.
var IgnoredProviders = map[string]bool{}

func getCorrections(d *models.DomainConfig) ([]*models.Correction, error) {
	cs := []*models.Correction{}
	for _, p := range d.DNSProviderInstances {
		if IgnoredProviders[p.Name] {
			continue
		}
		dc, err := d.Copy()
		if err != nil {
			return nil, err
		}
		corrections, err := p.Driver.GetDomainCorrections(dc)
		if err != nil {
			return nil, err
		}
		for _, c := range corrections {
			c.Msg = fmt.Sprintf("[%s] %s", p.Name, strings.TrimSpace(c.Msg))
		}
		cs = append(cs, corrections...)
	}
	return cs, nil
}

func getAndRunCorrections(d *models.DomainConfig) error {
	cs, err := getCorrections(d)
	if err != nil {
		return err
	}
	fmt.Printf("%d corrections\n", len(cs))
	for _, c := range cs {
		fmt.Printf("Running [%s]\n", c.Msg)
		err = c.F()
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *certManager) CleanUp(domain, token, keyAuth string) error {
	d := c.cfg.DomainContainingFQDN(domain)
	return getAndRunCorrections(d)
}
