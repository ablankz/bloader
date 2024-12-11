package encrypt

import "github.com/ablankz/bloader/internal/config"

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
func NewEncrypterContainerFromConfig(conf config.ValidEncryptConfig) (EncrypterContainer, error) {
	ec := make(EncrypterContainer)
	for _, e := range conf {
		var encrypter Encrypter
		switch e.Type {
		case EncryptTypeCBC:
			encrypter = NewStaticEncrypter([]byte(e.Key), []byte(e.IV), e.Type)
		case EncryptTypeCFB:
			encrypter = NewStaticEncrypter([]byte(e.Key), []byte(e.IV), e.Type)
		case EncryptTypeCTR:
			encrypter = NewStaticEncrypter([]byte(e.Key), []byte(e.IV), e.Type)
		}
		ec[e.ID] = encrypter
	}
	return ec, nil
}
