package encrypt

import (
	"fmt"

	"github.com/ablankz/bloader/internal/store"
	"github.com/ablankz/bloader/internal/utils"
)

// DynamicEncrypter is the dynamic encrypter.
type DynamicEncrypter struct {
	key    []byte
	method EncryptType
}

// NewDynamicEncrypter creates a new dynamic encrypter.
func NewDynamicEncrypter(str store.Store, storeBucketID, storeKey string, method EncryptType) (*DynamicEncrypter, error) {
	key, err := str.GetObject(storeBucketID, storeKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get key from store: %v", err)
	}
	if len(key) == 0 {
		key, err = utils.GenerateRandomBytes(32) // 256-bit key
		if err != nil {
			return nil, fmt.Errorf("failed to generate key: %v", err)
		}
		if err := str.PutObject(storeBucketID, storeKey, key); err != nil {
			return nil, fmt.Errorf("failed to store key: %v", err)
		}
	}

	return &DynamicEncrypter{
		key:    key,
		method: method,
	}, nil
}

// Encrypt encrypts the plaintext using the dynamic encrypter.
func (e *DynamicEncrypter) Encrypt(plaintext []byte) (string, error) {
	ciphertext, err := Encrypt(plaintext, e.key, e.method)
	if err != nil {
		return "", err
	}
	return ciphertext, nil
}

// Decrypt decrypts the ciphertext using the dynamic encrypter.
func (e *DynamicEncrypter) Decrypt(ciphertextBase64 string) ([]byte, error) {
	plaintext, err := Decrypt(ciphertextBase64, e.key, e.method)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

var _ Encrypter = (*DynamicEncrypter)(nil)
