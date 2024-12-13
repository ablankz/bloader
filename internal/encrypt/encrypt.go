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
	Encrypt(str store.Store, plaintext []byte) (string, error)
	Decrypt(str store.Store, ciphertextBase64 string) ([]byte, error)
}

// EncrypterContainer is the container for the encrypter.
type EncrypterContainer map[string]Encrypter

// NewEncrypterContainer creates a new encrypter container.
func NewEncrypterContainerFromConfig(conf config.ValidEncryptConfig) (EncrypterContainer, error) {
	ec := make(EncrypterContainer)
	for _, e := range conf {
		var encrypter Encrypter
		switch e.Type {
		case config.EncryptTypeStaticCBC:
			encrypter = NewStaticEncrypter([]byte(e.Key), EncryptTypeCBC)
		case config.EncryptTypeStaticCFB:
			encrypter = NewStaticEncrypter([]byte(e.Key), EncryptTypeCFB)
		case config.EncryptTypeStaticCTR:
			encrypter = NewStaticEncrypter([]byte(e.Key), EncryptTypeCTR)
		case config.EncryptTypeDynamicCBC:
			encrypter = NewDynamicEncrypter(e.Store.BucketID, e.Store.Key, EncryptTypeCBC)
		case config.EncryptTypeDynamicCFB:
			encrypter = NewDynamicEncrypter(e.Store.BucketID, e.Store.Key, EncryptTypeCFB)
		case config.EncryptTypeDynamicCTR:
			encrypter = NewDynamicEncrypter(e.Store.BucketID, e.Store.Key, EncryptTypeCTR)
		}
		ec[e.ID] = encrypter
	}
	return ec, nil
}
