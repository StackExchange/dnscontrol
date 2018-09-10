package acme

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"

	"github.com/xenolf/lego/acme"
)

func (c *certManager) createAcmeClient() (*acme.Client, error) {
	account, err := c.storage.GetAccount(c.acmeHost)
	if err != nil {
		return nil, err
	}
	if account == nil {
		// register new
		account, err = c.createAccount(c.email)
		if err != nil {
			return nil, err
		}
		if err := c.storage.StoreAccount(c.acmeHost, account); err != nil {
			return nil, err
		}
	}
	client, err := acme.NewClient(c.acmeDirectory, account, acme.RSA2048) // TODO: possibly make configurable on a cert-by cert basis
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (c *certManager) createAccount(email string) (*Account, error) {
	// if err := os.MkdirAll(c.accountDirectory(), dirPerms); err != nil {
	// 	return err
	// }
	privateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return nil, err
	}
	acct := &Account{
		key:   privateKey,
		Email: c.email,
	}
	c.client, err = acme.NewClient(c.acmeDirectory, acct, acme.EC384)
	if err != nil {
		return nil, err
	}
	reg, err := c.client.Register(true)
	if err != nil {
		return nil, err
	}
	acct.Registration = reg
	return acct, nil
	// acctBytes, err := json.MarshalIndent(c.account, "", "  ")
	// if err != nil {
	// 	return err
	// }
	// if err = ioutil.WriteFile(c.accountFile(), acctBytes, perms); err != nil {
	// 	return err
	// }
	// keyBytes, err := x509.MarshalECPrivateKey(privateKey)
	// if err != nil {
	// 	return err
	// }
	// pemKey := &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes}
	// pemBytes := pem.EncodeToMemory(pemKey)
	// if err = ioutil.WriteFile(c.accountKeyFile(), pemBytes, perms); err != nil {
	// 	return err
	// }
}

type Account struct {
	Email        string                     `json:"email"`
	key          crypto.PrivateKey          `json:"-"`
	Registration *acme.RegistrationResource `json:"registration"`
}

func (a *Account) GetEmail() string {
	return a.Email
}
func (a *Account) GetPrivateKey() crypto.PrivateKey {
	return a.key
}
func (a *Account) GetRegistration() *acme.RegistrationResource {
	return a.Registration
}
