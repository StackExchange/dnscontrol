package acme

import "github.com/xenolf/lego/acme"

// Storage is an abstracrion around how certificates, keys, and account info are stored on disk or elsewhere.
type Storage interface {
	// Get Existing certificate, or return nil if it does not exist
	GetCertificate(name string) (*acme.CertificateResource, error)
	StoreCertificate(name string, cert *acme.CertificateResource) error

	GetAccount(acmeHost string) (*Account, error)
	StoreAccount(acmeHost string, account *Account) error
}
