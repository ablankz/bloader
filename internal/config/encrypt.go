package config

import "fmt"

// EncryptRespectiveConfig is the configuration for the encrypt.
type EncryptRespectiveConfig struct {
	ID       *string `mapstructure:"id"`
	Type     *string `mapstructure:"type"`
	Key      *[]byte `mapstructure:"key"`
	StoreKey *string `mapstructure:"store_key"`
}

// ValidEncryptRespectiveConfig represents the valid encrypt configuration
type ValidEncryptRespectiveConfig struct {
	ID       string
	Type     EncryptType
	Key      []byte
	StoreKey string
}

// EncryptType is the type of the encrypt.
type EncryptType string

const (
	// EncryptTypeStaticCBC is the type of the static cbc.
	EncryptTypeStaticCBC EncryptType = "staticCBC"
	// EncryptTypeStaticCFB is the type of the static cfb.
	EncryptTypeStaticCFB EncryptType = "staticCFB"
	// EncryptTypeStaticCTR is the type of the static ctr.
	EncryptTypeStaticCTR EncryptType = "staticCTR"
	// EncryptTypeDynamicCBC is the type of the dynamic cbc.
	EncryptTypeDynamicCBC EncryptType = "dynamicCBC"
	// EncryptTypeDynamicCFB is the type of the dynamic cfb.
	EncryptTypeDynamicCFB EncryptType = "dynamicCFB"
	// EncryptTypeDynamicCTR is the type of the dynamic ctr.
	EncryptTypeDynamicCTR EncryptType = "dynamicCTR"
)

// EncryptConfig is the configuration for the encrypt.
type EncryptConfig []EncryptRespectiveConfig

// ValidEncryptConfig represents the valid encrypt configuration
type ValidEncryptConfig []ValidEncryptRespectiveConfig

// Validate validates the encrypt configuration.
func (c EncryptConfig) Validate() (ValidEncryptConfig, error) {
	var valid ValidEncryptConfig
	var idSet = make(map[string]struct{})
	for i, ec := range c {
		var validRespective ValidEncryptRespectiveConfig
		if ec.ID == nil {
			return ValidEncryptConfig{}, fmt.Errorf("encrypt[%d].id: %w", i, ErrEncryptIDRequired)
		}
		if _, ok := idSet[*ec.ID]; ok {
			return ValidEncryptConfig{}, fmt.Errorf("encrypt[%d].id: %w", i, ErrEncryptIDDuplicate)
		}
		idSet[*ec.ID] = struct{}{}
		validRespective.ID = *ec.ID
		if ec.Type == nil {
			return ValidEncryptConfig{}, fmt.Errorf("encrypt[%d].type: %w", i, ErrEncryptTypeRequired)
		}
		switch EncryptType(*ec.Type) {
		case EncryptTypeStaticCBC, EncryptTypeStaticCFB, EncryptTypeStaticCTR:
			validRespective.Type = EncryptType(*ec.Type)
			if ec.Key == nil {
				return ValidEncryptConfig{}, fmt.Errorf("encrypt[%d].key: %w", i, ErrEncryptKeyRequired)
			}
			validRespective.Key = *ec.Key
			if len(*ec.Key) != 16 && len(*ec.Key) != 24 && len(*ec.Key) != 32 {
				return ValidEncryptConfig{}, fmt.Errorf("encrypt[%d].key: %w", i, ErrEncryptRSAKeySizeInvalid)
			}
		case EncryptTypeDynamicCBC, EncryptTypeDynamicCFB, EncryptTypeDynamicCTR:
			validRespective.Type = EncryptType(*ec.Type)
			if ec.StoreKey == nil {
				return ValidEncryptConfig{}, fmt.Errorf("encrypt[%d].storeKey: %w", i, ErrEncryptStoreKeyRequired)
			}
			validRespective.StoreKey = *ec.StoreKey
		default:
			return ValidEncryptConfig{}, fmt.Errorf("encrypt[%d].type: %w", i, ErrEncryptTypeInvalid)
		}
		valid = append(valid, validRespective)
	}
	return valid, nil
}
