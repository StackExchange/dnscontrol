package acme

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"

	"github.com/go-acme/lego/certcrypto"
	"github.com/go-acme/lego/lego"
	"github.com/go-acme/lego/registration"
)

func (c *certManager) getOrCreateAccount() (*Account, error) {
	account, err := c.storage.GetAccount(c.acmeHost)
	if err != nil {
		return nil, err
	}
	if account != nil {
		return account, nil
	}
	// register new
	account, err = c.createAccount(c.email)
	if err != nil {
		return nil, err
	}
	err = c.storage.StoreAccount(c.acmeHost, account)
	return account, err
}

func (c *certManager) createAccount(email string) (*Account, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return nil, err
	}
	acct := &Account{
		key:   privateKey,
		Email: c.email,
	}
	config := lego.NewConfig(acct)
	config.CADirURL = c.acmeDirectory
	config.Certificate.KeyType = certcrypto.EC384
	client, err := lego.NewClient(config)
	if err != nil {
		return nil, err
	}
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return nil, err
	}
	acct.Registration = reg
	return acct, nil
}

// Account stores the data related to an ACME account.
type Account struct {
	Email        string                 `json:"email"`
	Registration *registration.Resource `json:"registration"`
	key          *ecdsa.PrivateKey
}

// GetEmail is a getter for the Email field.
func (a *Account) GetEmail() string {
	return a.Email
}

// GetPrivateKey is a getter for the PrivateKey field.
func (a *Account) GetPrivateKey() crypto.PrivateKey {
	return a.key
}

// GetRegistration is a getter for the registration field.
func (a *Account) GetRegistration() *registration.Resource {
	return a.Registration
}
