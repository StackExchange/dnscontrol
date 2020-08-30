package acme

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strings"

	"github.com/go-acme/lego/certificate"

	"github.com/hashicorp/vault/api"
)

type vaultStorage struct {
	path   string
	client *api.Logical
}

func makeVaultStorage(vaultPath string) (Storage, error) {
	if !strings.HasSuffix(vaultPath, "/") {
		vaultPath += "/"
	}
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return nil, err
	}
	storage := &vaultStorage{
		path:   vaultPath,
		client: client.Logical(),
	}
	return storage, nil
}

func (v *vaultStorage) GetCertificate(name string) (*certificate.Resource, error) {
	var err error

	path := v.certPath(name)
	secret, err := v.client.Read(path)
	if err != nil {
		return nil, err
	}
	if secret == nil {
		return nil, nil
	}
	cert := &certificate.Resource{}
	if dat, err := v.getString("meta", secret.Data, path); err != nil {
		return nil, err
	} else if err = json.Unmarshal(dat, cert); err != nil {
		return nil, err
	}

	var dat []byte
	if dat, err = v.getString("tls.cert", secret.Data, path); err != nil {
		return nil, err
	}
	cert.Certificate = dat

	if dat, err = v.getString("tls.key", secret.Data, path); err != nil {
		return nil, err
	}
	cert.PrivateKey = dat

	return cert, nil
}

func (v *vaultStorage) getString(key string, data map[string]interface{}, path string) ([]byte, error) {
	dat, ok := data[key]
	if !ok {
		return nil, fmt.Errorf("secret at %s does not have key %s", path, key)
	}
	str, ok := dat.(string)
	if !ok {
		return nil, fmt.Errorf("secret at %s is not string", path)
	}
	return []byte(str), nil
}

func (v *vaultStorage) StoreCertificate(name string, cert *certificate.Resource) error {
	jDat, err := json.MarshalIndent(cert, "", "  ")
	if err != nil {
		return err
	}
	pub := string(cert.Certificate)
	key := string(cert.PrivateKey)
	data := map[string]interface{}{
		"tls.cert":     pub,
		"tls.key":      key,
		"tls.combined": pub + "\n" + key,
		"meta":         string(jDat),
	}
	_, err = v.client.Write(v.certPath(name), data)
	return err
}

func (v *vaultStorage) registrationPath(acmeHost string) string {
	return v.path + ".letsencrypt/" + acmeHost
}

func (v *vaultStorage) certPath(name string) string {
	return v.path + name
}

func (v *vaultStorage) GetAccount(acmeHost string) (*Account, error) {
	path := v.registrationPath(acmeHost)
	secret, err := v.client.Read(path)
	if err != nil {
		return nil, err
	}
	if secret == nil {
		return nil, nil
	}
	acct := &Account{}
	if dat, err := v.getString("registration", secret.Data, path); err != nil {
		return nil, err
	} else if err = json.Unmarshal(dat, acct); err != nil {
		return nil, err
	}

	var key *ecdsa.PrivateKey
	var dat []byte
	var block *pem.Block
	if dat, err = v.getString("tls.key", secret.Data, path); err != nil {
		return nil, err
	} else if block, _ = pem.Decode(dat); block == nil {
		return nil, fmt.Errorf("error decoding account private key")
	} else if key, err = x509.ParseECPrivateKey(block.Bytes); err != nil {
		return nil, err
	}
	acct.key = key
	return acct, nil
}

func (v *vaultStorage) StoreAccount(acmeHost string, account *Account) error {
	acctBytes, err := json.MarshalIndent(account, "", "  ")
	if err != nil {
		return err
	}
	keyBytes, err := x509.MarshalECPrivateKey(account.key)
	if err != nil {
		return err
	}
	pemKey := &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes}
	pemBytes := pem.EncodeToMemory(pemKey)

	_, err = v.client.Write(v.registrationPath(acmeHost), map[string]interface{}{
		"registration": string(acctBytes),
		"tls.key":      string(pemBytes),
	})
	return err
}
