// Package container provides the dependencies for the application.
package container

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ablankz/bloader/internal/auth"
	"github.com/ablankz/bloader/internal/clock"
	"github.com/ablankz/bloader/internal/clock/fakeclock"
	"github.com/ablankz/bloader/internal/config"
	"github.com/ablankz/bloader/internal/i18n"
	"github.com/ablankz/bloader/internal/logger"
	"gopkg.in/yaml.v3"
)

// Container holds the dependencies for the application
type Container struct {
	Ctx        context.Context
	Clocker    clock.Clock
	Translator i18n.Translation
	Config     config.Config
	Logger     logger.Logger
	AuthToken  *auth.AuthToken
}

// NewContainer creates a new Container
func NewContainer() *Container {
	return &Container{}
}

// Init initializes the Container
func (c *Container) Init(cfg config.Config) error {
	c.Ctx = context.Background()
	var err error

	// ----------------------------------------
	// Set Config
	// ----------------------------------------
	c.Config = cfg

	// ----------------------------------------
	// Set Default Language
	// ----------------------------------------
	switch c.Config.Lang {
	case "en":
		i18n.Default = i18n.English
	case "ja":
		i18n.Default = i18n.Japanese
	case "":
		fmt.Println("No language specified. Defaulting to English.")
		c.Config.Lang = "en"
		i18n.Default = i18n.English
	default:
		fmt.Println("Invalid language specified. Defaulting to English.")
		c.Config.Lang = "en"
		i18n.Default = i18n.English
	}

	// ----------------------------------------
	// Set Clock
	// ----------------------------------------
	if _, err = time.Parse(c.Config.Clock.Format, c.Config.Clock.Format); err != nil {
		fmt.Println("Invalid clock format. Defaulting to 2006-01-02 15:04:05.\n Error:", err)
		c.Config.Clock.Format = "2006-01-02 15:04:05"
	}

	clk := clock.New()
	if cfg.Clock.Fake.Enabled {
		fakeClk := fakeclock.New(cfg.Clock.Fake.Time)
		clk = fakeClk
	}
	c.Clocker = clk

	//----------------------------------------
	// Set Translator
	//----------------------------------------
	c.Translator, err = i18n.NewTranslator()
	if err != nil {
		return fmt.Errorf("failed to create translator: %w", err)
	}

	// ----------------------------------------
	// Set Logger
	// ----------------------------------------
	c.Logger = logger.NewSlogLogger()
	if err := c.Logger.SetupLogger(&cfg.Logging); err != nil {
		return fmt.Errorf("failed to setup logger: %w", err)
	}

	// ----------------------------------------
	// Set AuthToken
	// ----------------------------------------
	defaultAuth := auth.AuthToken{
		AccessToken:  "",
		RefreshToken: "",
		TokenType:    "",
		Expiry:       time.Time{},
	}
	if _, err := os.Stat(cfg.Credential.Path); os.IsNotExist(err) {
		fmt.Println("credential file does not exist. creating a new one.")
		if err := createFileWithDefaultConfig(cfg.Credential.Path, &defaultAuth); err != nil {
			return fmt.Errorf("failed to create credential file: %w", err)
		}
	}
	tokenInfo, err := readAuthTokenConfig(cfg.Credential.Path)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}
	c.AuthToken = &tokenInfo

	return nil
}

type savable interface {
	Save(encoder *yaml.Encoder) error
}

func createFileWithDefaultConfig(filename string, defaultConf savable) error {
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	defer encoder.Close()

	if err := defaultConf.Save(encoder); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	return nil
}

func readAuthTokenConfig(filename string) (auth.AuthToken, error) {
	file, err := os.Open(filename)
	if err != nil {
		return auth.AuthToken{}, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	if fileStat, err := file.Stat(); err != nil {
		return auth.AuthToken{}, fmt.Errorf("failed to get file stat: %w", err)
	} else if fileStat.Size() == 0 {
		return auth.AuthToken{}, nil
	}

	var authToken auth.AuthToken
	decoder := yaml.NewDecoder(file)
	if err := authToken.Load(decoder); err != nil {
		return auth.AuthToken{}, fmt.Errorf("failed to load config: %w", err)
	}

	return authToken, nil
}

// Close closes the Container
func (c *Container) Close() error {
	c.Logger.Close()
	return nil
}
