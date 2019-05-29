// Package acme provides a means of performing Let's Encrypt DNS challenges via a DNSConfig
package acme

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/pkg/nameservers"
	"github.com/StackExchange/dnscontrol/pkg/notifications"
	"github.com/xenolf/lego/acme"
	acmelog "github.com/xenolf/lego/log"
)

type CertConfig struct {
	CertName   string   `json:"cert_name"`
	Names      []string `json:"names"`
	UseECC     bool     `json:"use_ecc"`
	MustStaple bool     `json:"must_staple"`
}

type Client interface {
	IssueOrRenewCert(config *CertConfig, renewUnder int, verbose bool) (bool, error)
}

type certManager struct {
	email         string
	acmeDirectory string
	acmeHost      string

	storage         Storage
	cfg             *models.DNSConfig
	domains         map[string]*models.DomainConfig
	originalDomains []*models.DomainConfig

	notifier notifications.Notifier

	account    *Account
	waitedOnce bool
}

const (
	LetsEncryptLive  = "https://acme-v02.api.letsencrypt.org/directory"
	LetsEncryptStage = "https://acme-staging-v02.api.letsencrypt.org/directory"
)

func New(cfg *models.DNSConfig, directory string, email string, server string, notify notifications.Notifier) (Client, error) {
	return commonNew(cfg, directoryStorage(directory), email, server, notify)
}

func commonNew(cfg *models.DNSConfig, storage Storage, email string, server string, notify notifications.Notifier) (Client, error) {
	u, err := url.Parse(server)
	if err != nil || u.Host == "" {
		return nil, fmt.Errorf("ACME directory '%s' is not a valid URL", server)
	}
	c := &certManager{
		storage:       storage,
		email:         email,
		acmeDirectory: server,
		acmeHost:      u.Host,
		cfg:           cfg,
		domains:       map[string]*models.DomainConfig{},
		notifier:      notify,
	}

	acct, err := c.getOrCreateAccount()
	if err != nil {
		return nil, err
	}
	c.account = acct
	return c, nil
}

func NewVault(cfg *models.DNSConfig, vaultPath string, email string, server string, notify notifications.Notifier) (Client, error) {
	storage, err := makeVaultStorage(vaultPath)
	if err != nil {
		return nil, err
	}
	return commonNew(cfg, storage, email, server, notify)
}

// IssueOrRenewCert will obtain a certificate with the given name if it does not exist,
// or renew it if it is close enough to the expiration date.
// It will return true if it issued or updated the certificate.
func (c *certManager) IssueOrRenewCert(cfg *CertConfig, renewUnder int, verbose bool) (bool, error) {
	if !verbose {
		acmelog.Logger = log.New(ioutil.Discard, "", 0)
	}
	defer c.finalCleanUp()

	log.Printf("Checking certificate [%s]", cfg.CertName)
	existing, err := c.storage.GetCertificate(cfg.CertName)
	if err != nil {
		return false, err
	}

	var client *acme.Client

	var action = func() (*acme.CertificateResource, error) {
		return client.ObtainCertificate(cfg.Names, true, nil, cfg.MustStaple)
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
			action = func() (*acme.CertificateResource, error) {
				return client.RenewCertificate(*existing, true, cfg.MustStaple)
			}
		}
	}

	kt := acme.RSA2048
	if cfg.UseECC {
		kt = acme.EC256
	}
	client, err = acme.NewClient(c.acmeDirectory, c.account, kt)
	if err != nil {
		return false, err
	}
	client.ExcludeChallenges([]acme.Challenge{acme.HTTP01, acme.TLSALPN01})
	client.SetChallengeProvider(acme.DNS01, c)

	acme.PreCheckDNS = c.preCheckDNS
	defer func() { acme.PreCheckDNS = acmePreCheck }()

	certResource, err := action()
	if err != nil {
		return false, err
	}
	fmt.Printf("Obtained certificate for %s\n", cfg.CertName)
	if err = c.storage.StoreCertificate(cfg.CertName, certResource); err != nil {
		return true, err
	}

	return true, nil
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

func (c *certManager) Present(domain, token, keyAuth string) (e error) {
	d := c.cfg.DomainContainingFQDN(domain)
	name := d.Name
	if seen := c.domains[name]; seen != nil {
		// we've already pre-processed this domain, just need to add to it.
		d = seen
	} else {
		// one-time tasks to get this domain ready.
		// if multiple validations on a single domain, we don't need to rebuild all this.

		// fix NS records for this domain's DNS providers
		nsList, err := nameservers.DetermineNameservers(d)
		if err != nil {
			return err
		}
		d.Nameservers = nsList
		nameservers.AddNSRecords(d)

		// make sure we have the latest config before we change anything.
		// alternately, we could avoid a lot of this trouble if we really really trusted no-purge in all cases
		if err := c.ensureNoPendingCorrections(d); err != nil {
			return err
		}

		// copy domain and work from copy from now on. That way original config can be used to "restore" when we are all done.
		copy, err := d.Copy()
		if err != nil {
			return err
		}
		c.originalDomains = append(c.originalDomains, d)
		c.domains[name] = copy
		d = copy
	}

	fqdn, val, _ := acme.DNS01Record(domain, keyAuth)
	txt := &models.RecordConfig{Type: "TXT"}
	txt.SetTargetTXT(val)
	txt.SetLabelFromFQDN(fqdn, d.Name)
	d.Records = append(d.Records, txt)
	return c.getAndRunCorrections(d)
}

func (c *certManager) ensureNoPendingCorrections(d *models.DomainConfig) error {
	corrections, err := c.getCorrections(d)
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

func (c *certManager) getCorrections(d *models.DomainConfig) ([]*models.Correction, error) {
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

func (c *certManager) getAndRunCorrections(d *models.DomainConfig) error {
	cs, err := c.getCorrections(d)
	if err != nil {
		return err
	}
	fmt.Printf("%d corrections\n", len(cs))
	for _, corr := range cs {
		fmt.Printf("Running [%s]\n", corr.Msg)
		err = corr.F()
		c.notifier.Notify(d.Name, "certs", corr.Msg, err, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *certManager) CleanUp(domain, token, keyAuth string) error {
	// do nothing for now. We will do a final clean up step at the very end.
	return nil
}

func (c *certManager) finalCleanUp() error {
	log.Println("Cleaning up all records we made")
	var lastError error
	for _, d := range c.originalDomains {
		if err := c.getAndRunCorrections(d); err != nil {
			log.Printf("ERROR cleaning up: %s", err)
			lastError = err
		}
	}
	return lastError
}
