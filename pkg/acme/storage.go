package acme

import "github.com/xenolf/lego/acme"

type Storage interface {
	// Get Existing certificate, or return nil if it does not exist
	GetCertificate(name string) (*acme.CertificateResource, error)
	StoreCertificate(name string, cert *acme.CertificateResource) error

	GetAccount(acmeHost string) (*Account, error)
	StoreAccount(acmeHost string, account *Account) error
}
