package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/acme"
	"github.com/StackExchange/dnscontrol/v3/pkg/normalize"
	"github.com/StackExchange/dnscontrol/v3/pkg/printer"
	"github.com/urfave/cli/v2"
)

var _ = cmd(catUtils, func() *cli.Command {
	var args GetCertsArgs
	return &cli.Command{
		Name:  "get-certs",
		Usage: "Issue certificates via Let's Encrypt",
		Action: func(c *cli.Context) error {
			return exit(GetCerts(args))
		},
		Flags: args.flags(),
	}
}())

// GetCertsArgs stores the flags and arguments common to cert commands
type GetCertsArgs struct {
	GetDNSConfigArgs
	GetCredentialsArgs

	ACMEServer     string
	CertsFile      string
	RenewUnderDays int
	CertDirectory  string
	Email          string
	AgreeTOS       bool
	Verbose        bool
	Vault          bool
	VaultPath      string
	Only           string

	Notify bool

	IgnoredProviders string
}

func (args *GetCertsArgs) flags() []cli.Flag {
	flags := args.GetDNSConfigArgs.flags()
	flags = append(flags, args.GetCredentialsArgs.flags()...)

	flags = append(flags, &cli.StringFlag{
		Name:        "acme",
		Destination: &args.ACMEServer,
		Value:       "live",
		Usage:       `ACME server to issue against. Give full directory endpoint. Can also use 'staging' or 'live' for standard Let's Encrypt endpoints.`,
	})
	flags = append(flags, &cli.IntFlag{
		Name:        "renew",
		Destination: &args.RenewUnderDays,
		Value:       15,
		Usage:       `Renew certs with less than this many days remaining`,
	})
	flags = append(flags, &cli.StringFlag{
		Name:        "dir",
		Destination: &args.CertDirectory,
		Value:       ".",
		Usage:       `Directory to store certificates and other data`,
	})
	flags = append(flags, &cli.StringFlag{
		Name:        "certConfig",
		Destination: &args.CertsFile,
		Value:       "certs.json",
		Usage:       `Json file containing list of certificates to issue`,
	})
	flags = append(flags, &cli.StringFlag{
		Name:        "email",
		Destination: &args.Email,
		Value:       "",
		Usage:       `Email to register with let's encrypt`,
	})
	flags = append(flags, &cli.BoolFlag{
		Name:        "agreeTOS",
		Destination: &args.AgreeTOS,
		Usage:       `Must provide this to agree to Let's Encrypt terms of service`,
	})
	flags = append(flags, &cli.BoolFlag{
		Name:        "vault",
		Destination: &args.Vault,
		Usage:       `Store certificates as secrets in hashicorp vault instead of on disk.`,
	})
	flags = append(flags, &cli.StringFlag{
		Name:        "vaultPath",
		Destination: &args.VaultPath,
		Value:       "/secret/certs",
		Usage:       `Path in vault to store certificates`,
	})
	flags = append(flags, &cli.StringFlag{
		Name:        "skip",
		Destination: &args.IgnoredProviders,
		Value:       "",
		Usage:       `Provider names to not use for challenges (comma separated)`,
	})
	flags = append(flags, &cli.BoolFlag{
		Name:        "verbose",
		Destination: &args.Verbose,
		Usage:       "Enable detailed logging (deprecated: use the global -v flag)",
	})
	flags = append(flags, &cli.BoolFlag{
		Name:        "notify",
		Destination: &args.Notify,
		Usage:       `set to true to send notifications to configured destinations`,
	})
	flags = append(flags, &cli.StringFlag{
		Name:        "only",
		Destination: &args.Only,
		Usage:       `Only check a single cert. Provide cert name.`,
	})
	return flags
}

// GetCerts implements the get-certs command.
func GetCerts(args GetCertsArgs) error {
	fmt.Println(args.JSFile)
	// check agree flag
	if !args.AgreeTOS {
		return fmt.Errorf("you must agree to the Let's Encrypt Terms of Service by using -agreeTOS")
	}
	if args.Email == "" {
		return fmt.Errorf("must provide email to use for Let's Encrypt registration")
	}

	// load dns config
	cfg, err := GetDNSConfig(args.GetDNSConfigArgs)
	if err != nil {
		return err
	}
	errs := normalize.ValidateAndNormalizeConfig(cfg)
	if PrintValidationErrors(errs) {
		return fmt.Errorf("exiting due to validation errors")
	}
	notifier, err := InitializeProviders(args.CredsFile, cfg, args.Notify)
	if err != nil {
		return err
	}

	for _, skip := range strings.Split(args.IgnoredProviders, ",") {
		acme.IgnoredProviders[skip] = true
	}

	// load cert list
	certList := []*acme.CertConfig{}
	f, err := os.Open(args.CertsFile)
	if err != nil {
		return err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	err = dec.Decode(&certList)
	if err != nil {
		return err
	}
	if len(certList) == 0 {
		return fmt.Errorf("must provide at least one certificate to issue in cert configuration")
	}
	if err = validateCertificateList(certList, cfg); err != nil {
		return err
	}

	acmeServer := args.ACMEServer
	if acmeServer == "live" {
		acmeServer = acme.LetsEncryptLive
	} else if acmeServer == "staging" {
		acmeServer = acme.LetsEncryptStage
	}

	var client acme.Client

	if args.Vault {
		client, err = acme.NewVault(cfg, args.VaultPath, args.Email, acmeServer, notifier)
	} else {
		client, err = acme.New(cfg, args.CertDirectory, args.Email, acmeServer, notifier)
	}
	if err != nil {
		return err
	}
	var manyerr error
	for _, cert := range certList {
		if args.Only != "" && cert.CertName != args.Only {
			continue
		}
		v := args.Verbose || printer.DefaultPrinter.Verbose
		issued, err := client.IssueOrRenewCert(cert, args.RenewUnderDays, v)
		if issued || err != nil {
			notifier.Notify(cert.CertName, "certificate", "Issued new certificate", err, false)
		}
		if err != nil {
			if manyerr == nil {
				manyerr = err
			} else {
				manyerr = fmt.Errorf("%w; %v", manyerr, err)
			}
		}
	}
	notifier.Done()
	return manyerr
}

var validCertNamesRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_\-]*$`)

func validateCertificateList(certs []*acme.CertConfig, cfg *models.DNSConfig) error {
	for _, cert := range certs {
		name := cert.CertName
		if !validCertNamesRegex.MatchString(name) {
			return fmt.Errorf("'%s' is not a valid certificate name. Only alphanumerics, - and _ allowed", name)
		}
		sans := cert.Names
		if len(sans) > 100 {
			return fmt.Errorf("certificate '%s' has too many SANs. Max of 100", name)
		}
		if len(sans) == 0 {
			return fmt.Errorf("certificate '%s' needs at least one SAN", name)
		}
		for _, san := range sans {
			d := cfg.DomainContainingFQDN(san)
			if d == nil {
				return fmt.Errorf("DNS config has no domain that matches SAN '%s'", san)
			}
		}
	}
	return nil
}
