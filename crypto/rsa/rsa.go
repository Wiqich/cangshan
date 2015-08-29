package rsacrypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"

	"github.com/yangchenxing/cangshan/application"
)

func init() {
	application.RegisterModulePrototype("RSACrypto", new(RSACrypto))
}

type RSACrypto struct {
	PEMFiles []string
	keys     map[string]*rsa.PrivateKey
}

func (crypto *RSACrypto) Initialize() error {
	crypto.keys = make(map[string]*rsa.PrivateKey)
	for _, path := range crypto.PEMFiles {
		raw, err := ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("Read PEM file \"%s\" fail: %s", path, err.Error())
		}
		for block, data := pem.Decode(raw); block != nil; block, data = pem.Decode(data) {
			id := block.Headers["id"]
			if crypto.keys[id], err = x509.ParsePKCS1PrivateKey(block.Bytes); err != nil {
				return fmt.Errorf("Parse PKCS1 fail: %s", err.Error())
			}
		}
	}
	return nil
}

func (crypto *RSACrypto) Decrypt(keyID string, input []byte) ([]byte, error) {
	if key := crypto.keys[keyID]; key == nil {
		return nil, fmt.Errorf("Unknown RSA private key: %s", keyID)
	} else {
		return rsa.DecryptPKCS1v15(nil, key, input)
	}
}

func (crypto *RSACrypto) Encrypt(keyID string, input []byte) ([]byte, error) {
	if key := crypto.keys[keyID]; key == nil {
		return nil, fmt.Errorf("Unknown RSA private key: %s", keyID)
	} else {
		return rsa.EncryptPKCS1v15(rand.Reader, &key.PublicKey, input)
	}
}
