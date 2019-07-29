package acme

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"strings"

	"github.com/go-acme/lego/certificate"

	"github.com/hashicorp/vault/api"
)

type vaultStorage struct {
	path    string
	client  *api.Client
	v2Mount string
	isV2    bool
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
		client: client,
	}
	storage.inferKVVersion(vaultPath)
	return storage, nil
}

func (v *vaultStorage) inferKVVersion(vaultPath string) {
	mounts, err := v.client.Sys().ListMounts()
	if err != nil {
		log.Printf("Error listing vault mounts: '%s'. Assuming KV v1 mount.", err)
		return
	}
	longestMatch := 0
	var match *api.MountOutput
	var matchName string
	for name, m := range mounts {
		if m.Type != "kv" || !strings.HasPrefix(vaultPath, "/"+name) {
			continue
		}
		if len(name) > longestMatch {
			longestMatch = len(name)
			match = m
			matchName = "/" + name
		}
	}
	if match == nil {
		log.Printf("Unable to locate kv secret backend matching '%s'. Assuming KV v1 mount.", vaultPath)
		return
	}
	ver := match.Options["version"]
	if ver == "2" {
		log.Println("Found kv v2 mount")
		v.v2Mount = matchName
		v.isV2 = true
	} else if ver == "1" {
		log.Println("Found kv v1 mount")
	} else {
		log.Printf("Unknown kv version '%s' Assuming KV v1 mount.", ver)
	}
}

func (v *vaultStorage) GetCertificate(name string) (*certificate.Resource, error) {
	path := v.certPath(name)
	secret, err := v.client.Logical().Read(path)
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

	if dat, err := v.getString("tls.cert", secret.Data, path); err != nil {
		return nil, err
	} else {
		cert.Certificate = dat
	}

	if dat, err := v.getString("tls.key", secret.Data, path); err != nil {
		return nil, err
	} else {
		cert.PrivateKey = dat
	}

	return cert, nil
}

func (v *vaultStorage) getString(key string, data map[string]interface{}, path string) ([]byte, error) {
	if v.isV2 {
		data = data["data"].(map[string]interface{})
	}
	dat, ok := data[key]
	if !ok {
		return nil, fmt.Errorf("Secret at %s does not have key %s", path, key)
	}
	str, ok := dat.(string)
	if !ok {
		return nil, fmt.Errorf("Secret at %s is not string", path)
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
	if v.isV2 {
		data = map[string]interface{}{
			"data": data,
		}
	}
	_, err = v.client.Logical().Write(v.certPath(name), data)
	return err
}

func (v *vaultStorage) registrationPath(acmeHost string) string {
	return v.basePath() + ".letsencrypt/" + acmeHost
}

func (v *vaultStorage) certPath(name string) string {
	return v.basePath() + name
}

func (v *vaultStorage) basePath() string {
	if v.isV2 {
		return fmt.Sprintf("%sdata/%s", v.v2Mount, strings.TrimPrefix(v.path, v.v2Mount))
	}
	return v.path
}

func (v *vaultStorage) GetAccount(acmeHost string) (*Account, error) {
	path := v.registrationPath(acmeHost)
	secret, err := v.client.Logical().Read(path)
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

	if dat, err := v.getString("tls.key", secret.Data, path); err != nil {
		return nil, err
	} else if block, _ := pem.Decode(dat); block == nil {
		return nil, fmt.Errorf("Error decoding account private key")
	} else if key, err := x509.ParseECPrivateKey(block.Bytes); err != nil {
		return nil, err
	} else {
		acct.key = key
	}

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

	log.Println(v.registrationPath(acmeHost))
	data := map[string]interface{}{
		"registration": string(acctBytes),
		"tls.key":      string(pemBytes),
	}
	if v.isV2 {
		data = map[string]interface{}{
			"data": data,
		}
	}
	_, err = v.client.Logical().Write(v.registrationPath(acmeHost), data)
	return err
}
