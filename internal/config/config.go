// Package config provides configuration for the application.
package config

// ConfigForOverride represents the configuration for the override service
type ConfigForOverride struct {
	Env      *string         `mapstructure:"env"`
	Override *OverrideConfig `mapstructure:"override"`
}

// ValidConfigForOverride represents the configuration for the override service
type ValidConfigForOverride struct {
	Env      string              `mapstructure:"env"`
	Override ValidOverrideConfig `mapstructure:"override"`
}

// Validate validates the configuration for the override service
func (c ConfigForOverride) Validate() (ValidConfigForOverride, error) {
	var valid ValidConfigForOverride
	if c.Env == nil {
		return ValidConfigForOverride{}, ErrEnvRequired
	}
	valid.Env = *c.Env
	if c.Override == nil {
		return ValidConfigForOverride{}, ErrOverrideRequired
	}
	validOverride, err := c.Override.Validate()
	if err != nil {
		return ValidConfigForOverride{}, err
	}
	valid.Override = validOverride

	return valid, nil
}

// Config represents the application configuration
type Config struct {
	Env      *string         `mapstructure:"env"`
	Loader   *LoaderConfig   `mapstructure:"loader"`
	Targets  *TargetConfig   `mapstructure:"targets"`
	Outputs  *OutputConfig   `mapstructure:"outputs"`
	Store    *StoreConfig    `mapstructure:"store"`
	Encrypts *EncryptConfig  `mapstructure:"encrypts"`
	Auth     *AuthConfig     `mapstructure:"auth"`
	Server   *ServerConfig   `mapstructure:"server"`
	Logging  *LoggingConfig  `mapstructure:"logging"`
	Clock    *ClockConfig    `mapstructure:"clock"`
	Language *LanguageConfig `mapstructure:"language"`
	Override *OverrideConfig `mapstructure:"override"`
}

// ValidConfig represents the application configuration
type ValidConfig struct {
	Env      string
	Loader   ValidLoaderConfig
	Targets  ValidTargetConfig
	Outputs  ValidOutputConfig
	Store    ValidStoreConfig
	Encrypts ValidEncryptConfig
	Auth     ValidAuthConfig
	Server   ValidServerConfig
	Logging  ValidLoggingConfig
	Clock    ValidClockConfig
	Language ValidLanguageConfig
	Override ValidOverrideConfig
}

// Validate validates the configuration
func (c Config) Validate() (ValidConfig, error) {
	var valid ValidConfig
	if c.Env == nil {
		return ValidConfig{}, ErrEnvRequired
	}
	valid.Env = *c.Env

	if c.Loader == nil {
		return ValidConfig{}, ErrLoaderRequired
	}
	validLoader, err := c.Loader.Validate()
	if err != nil {
		return ValidConfig{}, err
	}
	valid.Loader = validLoader

	if c.Targets == nil {
		return ValidConfig{}, ErrTargetsRequired
	}
	validTargets, err := c.Targets.Validate()
	if err != nil {
		return ValidConfig{}, err
	}
	valid.Targets = validTargets

	if c.Outputs == nil {
		return ValidConfig{}, ErrOutputsRequired
	}
	validOutputs, err := c.Outputs.Validate()
	if err != nil {
		return ValidConfig{}, err
	}
	valid.Outputs = validOutputs

	if c.Store == nil {
		return ValidConfig{}, ErrStoreRequired
	}
	validStore, err := c.Store.Validate()
	if err != nil {
		return ValidConfig{}, err
	}
	valid.Store = validStore

	if c.Encrypts == nil {
		return ValidConfig{}, ErrEncryptsRequired
	}
	validEncrypts, err := c.Encrypts.Validate()
	if err != nil {
		return ValidConfig{}, err
	}
	valid.Encrypts = validEncrypts

	if c.Auth == nil {
		return ValidConfig{}, ErrAuthRequired
	}
	validAuth, err := c.Auth.Validate()
	if err != nil {
		return ValidConfig{}, err
	}
	valid.Auth = validAuth

	if c.Server == nil {
		return ValidConfig{}, ErrServerRequired
	}
	validServer, err := c.Server.Validate()
	if err != nil {
		return ValidConfig{}, err
	}
	valid.Server = validServer

	if c.Logging == nil {
		return ValidConfig{}, ErrLoggingRequired
	}
	validLogging, err := c.Logging.Validate()
	if err != nil {
		return ValidConfig{}, err
	}
	valid.Logging = validLogging

	if c.Clock == nil {
		return ValidConfig{}, ErrClockRequired
	}
	validClock, err := c.Clock.Validate()
	if err != nil {
		return ValidConfig{}, err
	}
	valid.Clock = validClock

	if c.Language == nil {
		return ValidConfig{}, ErrLanguageRequired
	}
	validLanguage, err := c.Language.Validate()
	if err != nil {
		return ValidConfig{}, err
	}
	valid.Language = validLanguage

	if c.Override == nil {
		return ValidConfig{}, ErrOverrideRequired
	}
	validOverride, err := c.Override.Validate()
	if err != nil {
		return ValidConfig{}, err
	}
	valid.Override = validOverride

	return valid, nil
}
