package encrypt

import "github.com/ablankz/bloader/internal/store"

// StaticEncrypter is the static encrypter.
type StaticEncrypter struct {
	key    []byte
	method EncryptType
}

// NewStaticEncrypter creates a new static encrypter.
func NewStaticEncrypter(key []byte, method EncryptType) *StaticEncrypter {
	return &StaticEncrypter{
		key:    key,
		method: method,
	}
}

// Encrypt encrypts the plaintext using the static encrypter.
func (e *StaticEncrypter) Encrypt(str store.Store, plaintext []byte) (string, error) {
	ciphertext, err := Encrypt(plaintext, e.key, e.method)
	if err != nil {
		return "", err
	}
	return ciphertext, nil
}

// Decrypt decrypts the ciphertext using the static encrypter.
func (e *StaticEncrypter) Decrypt(str store.Store, ciphertextBase64 string) ([]byte, error) {
	plaintext, err := Decrypt(ciphertextBase64, e.key, e.method)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

var _ Encrypter = (*StaticEncrypter)(nil)
