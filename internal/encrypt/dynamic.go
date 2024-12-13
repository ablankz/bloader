package encrypt

import (
	"fmt"

	"github.com/ablankz/bloader/internal/store"
	"github.com/ablankz/bloader/internal/utils"
)

// DynamicEncrypter is the dynamic encrypter.
type DynamicEncrypter struct {
	storeBucketID string
	storeKey      string
	method        EncryptType
}

// NewDynamicEncrypter creates a new dynamic encrypter.
func NewDynamicEncrypter(storeBucketID, storeKey string, method EncryptType) *DynamicEncrypter {
	return &DynamicEncrypter{
		storeBucketID: storeBucketID,
		storeKey:      storeKey,
		method:        method,
	}
}

// Encrypt encrypts the plaintext using the dynamic encrypter.
func (e *DynamicEncrypter) Encrypt(str store.Store, plaintext []byte) (string, error) {
	key, err := utils.GenerateRandomBytes(32) // 256-bit key
	if err != nil {
		return "", fmt.Errorf("failed to generate key: %v", err)
	}
	ciphertext, err := Encrypt(plaintext, key, e.method)
	if err != nil {
		return "", err
	}
	if err := str.PutObject(e.storeBucketID, e.storeKey, key); err != nil {
		return "", fmt.Errorf("failed to store key: %v", err)
	}
	return ciphertext, nil
}

// Decrypt decrypts the ciphertext using the dynamic encrypter.
func (e *DynamicEncrypter) Decrypt(str store.Store, ciphertextBase64 string) ([]byte, error) {
	key, err := str.GetObject(e.storeBucketID, e.storeKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get key: %v", err)
	}
	plaintext, err := Decrypt(ciphertextBase64, key, e.method)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

var _ Encrypter = (*DynamicEncrypter)(nil)
