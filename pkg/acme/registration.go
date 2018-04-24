package acme

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/xenolf/lego/acmev2"
)

func (c *certManager) loadOrCreateAccount() error {
	f, err := os.Open(c.accountFile())
	if err != nil && os.IsNotExist(err) {
		return c.createAccount()
	}
	if err != nil {
		return err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	acct := &account{}
	if err = dec.Decode(acct); err != nil {
		return err
	}
	c.account = acct
	keyBytes, err := ioutil.ReadFile(c.accountKeyFile())
	if err != nil {
		return err
	}
	keyBlock, _ := pem.Decode(keyBytes)
	if keyBlock == nil {
		log.Fatal("WTF", keyBytes)
	}
	c.account.key, err = x509.ParseECPrivateKey(keyBlock.Bytes)
	if err != nil {
		return err
	}
	c.client, err = acme.NewClient(c.acmeDirectory, c.account, acme.RSA2048) // TODO: possibly make configurable on a cert-by cert basis
	if err != nil {
		return err
	}
	return nil
}

func (c *certManager) accountDirectory() string {
	dir := strings.TrimPrefix(c.acmeDirectory, "https://")
	dir = strings.TrimPrefix(dir, "http://")
	return filepath.Join(c.directory, ".letsencrypt", dir)
}
func (c *certManager) accountFile() string {
	return filepath.Join(c.accountDirectory(), "account.json")
}
func (c *certManager) accountKeyFile() string {
	return filepath.Join(c.accountDirectory(), "account.key")
}

const perms os.FileMode = 0644 // TODO: probably lock this down more

func (c *certManager) createAccount() error {
	if err := os.MkdirAll(c.accountDirectory(), perms); err != nil {
		return err
	}
	privateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return err
	}
	acct := &account{
		key:   privateKey,
		Email: c.email,
	}
	c.account = acct
	c.client, err = acme.NewClient(c.acmeDirectory, c.account, acme.EC384)
	if err != nil {
		return err
	}
	reg, err := c.client.Register(true)
	if err != nil {
		return err
	}
	c.account.Registration = reg
	acctBytes, err := json.MarshalIndent(c.account, "", "  ")
	if err != nil {
		return err
	}
	if err = ioutil.WriteFile(c.accountFile(), acctBytes, perms); err != nil {
		return err
	}
	keyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return err
	}
	pemKey := &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes}
	pemBytes := pem.EncodeToMemory(pemKey)
	if err = ioutil.WriteFile(c.accountKeyFile(), pemBytes, perms); err != nil {
		return err
	}
	return nil
}

type account struct {
	Email        string `json:"email"`
	key          crypto.PrivateKey
	Registration *acme.RegistrationResource `json:"registration"`
}

func (a *account) GetEmail() string {
	return a.Email
}
func (a *account) GetPrivateKey() crypto.PrivateKey {
	return a.key
}
func (a *account) GetRegistration() *acme.RegistrationResource {
	return a.Registration
}
