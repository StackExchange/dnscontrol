package acme

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/xenolf/lego/acme"
)

// directoryStorage implements storage in a local file directory
type directoryStorage string

// filename for certifiacte / key / json file
func (d directoryStorage) certFile(name, ext string) string {
	return filepath.Join(d.certDir(name), name+"."+ext)
}
func (d directoryStorage) certDir(name string) string {
	return filepath.Join(string(d), "certificates", name)
}

func (d directoryStorage) accountDirectory(acmeHost string) string {
	return filepath.Join(string(d), ".letsencrypt", acmeHost)
}

func (d directoryStorage) accountFile(acmeHost string) string {
	return filepath.Join(d.accountDirectory(acmeHost), "account.json")
}
func (d directoryStorage) accountKeyFile(acmeHost string) string {
	return filepath.Join(d.accountDirectory(acmeHost), "account.key")
}

// TODO: probably lock these down more
const perms os.FileMode = 0644
const dirPerms os.FileMode = 0700

func (d directoryStorage) GetCertificate(name string) (*acme.CertificateResource, error) {
	f, err := os.Open(d.certFile(name, "json"))
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
	crtBytes, err := ioutil.ReadFile(d.certFile(name, "crt"))
	if err != nil {
		return nil, err
	}
	cr.Certificate = crtBytes
	return cr, nil
}

func (d directoryStorage) StoreCertificate(name string, cert *acme.CertificateResource) error {
	// make sure actual cert data never gets into metadata json
	if err := os.MkdirAll(d.certDir(name), dirPerms); err != nil {
		return err
	}
	pub := cert.Certificate
	cert.Certificate = nil
	priv := cert.PrivateKey
	cert.PrivateKey = nil
	jDAt, err := json.MarshalIndent(cert, "", "  ")
	if err != nil {
		return err
	}
	if err = ioutil.WriteFile(d.certFile(name, "json"), jDAt, perms); err != nil {
		return err
	}
	if err = ioutil.WriteFile(d.certFile(name, "crt"), pub, perms); err != nil {
		return err
	}
	return ioutil.WriteFile(d.certFile(name, "key"), priv, perms)
}

func (d directoryStorage) GetAccount(acmeHost string) (*Account, error) {
	f, err := os.Open(d.accountFile(acmeHost))
	if err != nil && os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	acct := &Account{}
	if err = dec.Decode(acct); err != nil {
		return nil, err
	}
	keyBytes, err := ioutil.ReadFile(d.accountKeyFile(acmeHost))
	if err != nil {
		return nil, err
	}
	keyBlock, _ := pem.Decode(keyBytes)
	if keyBlock == nil {
		return nil, fmt.Errorf("Error decoding account private key.")
	}
	acct.key, err = x509.ParseECPrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, err
	}
	return acct, nil
}
func (d directoryStorage) StoreAccount(acmeHost string, account *Account) error {
	return nil
}
