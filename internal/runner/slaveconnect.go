package runner

import (
	"fmt"

	"github.com/ablankz/bloader/internal/master"
)

// SlaveConnect represents the SlaveConnect runner
type SlaveConnect struct {
	Slaves []SlaveConnectData `yaml:"slaves"`
}

// Validate validates the SlaveConnect
func (r SlaveConnect) Validate() (master.SlaveConnect, error) {
	var validSlaves []master.SlaveConnectData
	for i, d := range r.Slaves {
		valid, err := d.Validate()
		if err != nil {
			return master.SlaveConnect{}, fmt.Errorf("failed to validate data at index %d: %v", i, err)
		}
		validSlaves = append(validSlaves, valid)
	}
	return master.SlaveConnect{
		Slaves: validSlaves,
	}, nil
}

// SlaveConnectData represents the data for the SlaveConnect
type SlaveConnectData struct {
	ID          *string                 `yaml:"id"`
	URI         *string                 `yaml:"uri"`
	Certificate SlaveConnectCertificate `yaml:"certificate"`
	Encrypt     CredentialEncryptConfig `yaml:"encrypt"`
}

// Validate validates the SlaveConnectData
func (d SlaveConnectData) Validate() (master.SlaveConnectData, error) {
	var valid master.SlaveConnectData
	if d.ID == nil {
		return master.SlaveConnectData{}, fmt.Errorf("id is required")
	}
	valid.ID = *d.ID
	if d.URI == nil {
		return master.SlaveConnectData{}, fmt.Errorf("uri is required")
	}
	valid.URI = *d.URI
	validCertificate, err := d.Certificate.Validate()
	if err != nil {
		return master.SlaveConnectData{}, fmt.Errorf("failed to validate certificate: %v", err)
	}
	valid.Certificate = validCertificate
	validEncrypt, err := d.Encrypt.Validate()
	if err != nil {
		return master.SlaveConnectData{}, fmt.Errorf("failed to validate encrypt: %v", err)
	}
	valid.Encrypt = master.CredentialEncryptConfig(validEncrypt)
	return valid, nil
}

// SlaveConnectCertificate represents the certificate for the Slave
type SlaveConnectCertificate struct {
	Enabled            bool    `yaml:"enabled"`
	CACert             *string `yaml:"ca_cert"`
	ServerNameOverride string  `yaml:"server_name_override"`
	InsecureSkipVerify bool    `yaml:"insecure_skip_verify"`
}

// Validate validates the SlaveConnectCertificate
func (c SlaveConnectCertificate) Validate() (master.SlaveConnectCertificate, error) {
	if !c.Enabled {
		return master.SlaveConnectCertificate{}, nil
	}
	if c.CACert == nil {
		return master.SlaveConnectCertificate{}, fmt.Errorf("ca_cert is required")
	}
	return master.SlaveConnectCertificate{
		Enabled:            c.Enabled,
		CACert:             *c.CACert,
		ServerNameOverride: c.ServerNameOverride,
		InsecureSkipVerify: c.InsecureSkipVerify,
	}, nil
}
