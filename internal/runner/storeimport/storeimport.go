package storeimport

import (
	"context"
	"fmt"
	"sync"

	"github.com/ablankz/bloader/internal/container"
)

// StoreImport represents the StoreImport runner
type StoreImport struct {
	Data []StoreImportData `yaml:"data"`
}

// ValidStoreImport represents the valid StoreImport runner
type ValidStoreImport struct {
	Data []ValidStoreImportData
}

// Validate validates the StoreImport
func (r StoreImport) Validate() (ValidStoreImport, error) {
	var validData []ValidStoreImportData
	for i, d := range r.Data {
		valid, err := d.Validate()
		if err != nil {
			return ValidStoreImport{}, fmt.Errorf("failed to validate data at index %d: %v", i, err)
		}
		validData = append(validData, valid)
	}
	return ValidStoreImport{
		Data: validData,
	}, nil
}

// StoreImportData represents the data for the StoreImport runner
type StoreImportData struct {
	BucketID *string                 `yaml:"bucket_id"`
	Key      *string                 `yaml:"key"`
	StoreKey *string                 `yaml:"store_key"`
	Encrypt  CredentialEncryptConfig `yaml:"encrypt"`
}

// ValidStoreImportData represents the valid data for the StoreImport runner
type ValidStoreImportData struct {
	BucketID string
	Key      string
	StoreKey string
	Encrypt  ValidCredentialEncryptConfig
}

// Validate validates the StoreImportData
func (d StoreImportData) Validate() (ValidStoreImportData, error) {
	if d.BucketID == nil {
		return ValidStoreImportData{}, fmt.Errorf("bucket_id is required")
	}
	if d.Key == nil {
		return ValidStoreImportData{}, fmt.Errorf("key is required")
	}
	if d.StoreKey == nil {
		return ValidStoreImportData{}, fmt.Errorf("store_key is required")
	}
	validEncrypt, err := d.Encrypt.Validate()
	if err != nil {
		return ValidStoreImportData{}, fmt.Errorf("failed to validate encrypt: %v", err)
	}
	return ValidStoreImportData{
		BucketID: *d.BucketID,
		Key:      *d.Key,
		StoreKey: *d.StoreKey,
		Encrypt:  validEncrypt,
	}, nil
}

// CredentialEncryptConfig is the configuration for the credential encrypt.
type CredentialEncryptConfig struct {
	Enabled   bool    `yaml:"enabled"`
	EncryptID *string `yaml:"encrypt_id"`
}

// ValidCredentialEncryptConfig represents the valid auth credential encrypt configuration
type ValidCredentialEncryptConfig struct {
	Enabled   bool
	EncryptID string
}

// Validate validates the credential encrypt configuration
func (c CredentialEncryptConfig) Validate() (ValidCredentialEncryptConfig, error) {
	if !c.Enabled {
		return ValidCredentialEncryptConfig{}, nil
	}
	if c.EncryptID == nil {
		return ValidCredentialEncryptConfig{}, fmt.Errorf("encrypt_id is required")
	}
	return ValidCredentialEncryptConfig{
		Enabled:   c.Enabled,
		EncryptID: *c.EncryptID,
	}, nil
}

// Run runs the StoreImport runner
func (r ValidStoreImport) Run(ctx context.Context, ctr *container.Container, store *sync.Map) error {
	for _, d := range r.Data {
		valBytes, err := ctr.Store.GetObject(d.BucketID, d.StoreKey)
		if err != nil {
			return fmt.Errorf("failed to get object: %v", err)
		}
		if d.Encrypt.Enabled {
			encryptor, ok := ctr.EncypterContainer[d.Encrypt.EncryptID]
			if !ok {
				return fmt.Errorf("encryptor not found: %s", d.Encrypt.EncryptID)
			}
			decryptedVal, err := encryptor.Decrypt(string(valBytes))
			if err != nil {
				return fmt.Errorf("failed to decrypt value: %v", err)
			}
			valBytes = []byte(decryptedVal)
		}
		store.Store(d.Key, valBytes)
	}
	return nil
}
