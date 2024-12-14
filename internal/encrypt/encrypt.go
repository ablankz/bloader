package encrypt

import (
	"github.com/ablankz/bloader/internal/config"
	"github.com/ablankz/bloader/internal/store"
)

// EncryptType is the type of the encrypt.
type EncryptType string

const (
	// EncryptTypeCBC is the type of the cbc.
	EncryptTypeCBC EncryptType = "CBC"
	// EncryptTypeCFB is the type of the cfb.
	EncryptTypeCFB EncryptType = "CFB"
	// EncryptTypeCTR is the type of the ctr.
	EncryptTypeCTR EncryptType = "CTR"
)

// Encrypter is the interface for the encrypter.
type Encrypter interface {
	Encrypt(plaintext []byte) (string, error)
	Decrypt(ciphertextBase64 string) ([]byte, error)
}

// EncrypterContainer is the container for the encrypter.
type EncrypterContainer map[string]Encrypter

// NewEncrypterContainer creates a new encrypter container.
func NewEncrypterContainerFromConfig(str store.Store, conf config.ValidEncryptConfig) (EncrypterContainer, error) {
	ec := make(EncrypterContainer)
	var err error
	for _, e := range conf {
		var encrypter Encrypter
		switch e.Type {
		case config.EncryptTypeStaticCBC:
			encrypter, err = NewStaticEncrypter([]byte(e.Key), EncryptTypeCBC)
			if err != nil {
				return nil, err
			}
		case config.EncryptTypeStaticCFB:
			encrypter, err = NewStaticEncrypter([]byte(e.Key), EncryptTypeCFB)
			if err != nil {
				return nil, err
			}
		case config.EncryptTypeStaticCTR:
			encrypter, err = NewStaticEncrypter([]byte(e.Key), EncryptTypeCTR)
			if err != nil {
				return nil, err
			}
		case config.EncryptTypeDynamicCBC:
			encrypter, err = NewDynamicEncrypter(str, e.Store.BucketID, e.Store.Key, EncryptTypeCBC)
			if err != nil {
				return nil, err
			}
		case config.EncryptTypeDynamicCFB:
			encrypter, err = NewDynamicEncrypter(str, e.Store.BucketID, e.Store.Key, EncryptTypeCFB)
			if err != nil {
				return nil, err
			}
		case config.EncryptTypeDynamicCTR:
			encrypter, err = NewDynamicEncrypter(str, e.Store.BucketID, e.Store.Key, EncryptTypeCTR)
			if err != nil {
				return nil, err
			}
		}
		ec[e.ID] = encrypter
	}
	return ec, nil
}
