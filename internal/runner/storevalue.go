package runner

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ablankz/bloader/internal/container"
)

// StoreValue represents the StoreValue runner
type StoreValue struct {
	Data []StoreValueData `yaml:"data"`
}

// ValidStoreValue represents the valid StoreValue runner
type ValidStoreValue struct {
	Data []ValidStoreValueData
}

// Validate validates the StoreValue
func (r StoreValue) Validate() (ValidStoreValue, error) {
	var validData []ValidStoreValueData
	for i, d := range r.Data {
		valid, err := d.Validate()
		if err != nil {
			return ValidStoreValue{}, fmt.Errorf("failed to validate data at index %d: %v", i, err)
		}
		validData = append(validData, valid)
	}
	return ValidStoreValue{
		Data: validData,
	}, nil
}

// StoreValueData represents the data for the StoreValue runner
type StoreValueData struct {
	BucketID *string                 `yaml:"bucket_id"`
	Key      *string                 `yaml:"key"`
	Value    *any                    `yaml:"value"`
	Encrypt  CredentialEncryptConfig `yaml:"encrypt"`
}

// ValidStoreValueData represents the valid data for the StoreValue runner
type ValidStoreValueData struct {
	BucketID string
	Key      string
	Value    any
	Encrypt  ValidCredentialEncryptConfig
}

// Validate validates the StoreValueData
func (d StoreValueData) Validate() (ValidStoreValueData, error) {
	if d.BucketID == nil {
		return ValidStoreValueData{}, fmt.Errorf("bucket_id is required")
	}
	if d.Key == nil {
		return ValidStoreValueData{}, fmt.Errorf("key is required")
	}
	if d.Value == nil {
		return ValidStoreValueData{}, fmt.Errorf("value is required")
	}
	validEncrypt, err := d.Encrypt.Validate()
	if err != nil {
		return ValidStoreValueData{}, fmt.Errorf("failed to validate encrypt: %v", err)
	}
	return ValidStoreValueData{
		BucketID: *d.BucketID,
		Key:      *d.Key,
		Value:    *d.Value,
		Encrypt:  validEncrypt,
	}, nil
}

// Run runs the StoreValue runner
func (r ValidStoreValue) Run(ctx context.Context, ctr *container.Container) error {
	for _, d := range r.Data {
		valBytes, err := json.Marshal(d.Value)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %v", err)
		}
		if d.Encrypt.Enabled {
			encryptor, ok := ctr.EncypterContainer[d.Encrypt.EncryptID]
			if !ok {
				return fmt.Errorf("encryptor not found: %s", d.Encrypt.EncryptID)
			}
			encryptedVal, err := encryptor.Encrypt(valBytes)
			if err != nil {
				return fmt.Errorf("failed to encrypt value: %v", err)
			}
			valBytes = []byte(encryptedVal)
		}
		ctr.Store.PutObject(d.BucketID, d.Key, valBytes)
	}
	return nil
}
