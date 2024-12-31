package master

// SlaveConnect represents the valid SlaveConnect runner
type SlaveConnect struct {
	Slaves []SlaveConnectData
}

// SlaveConnectData represents the valid data for the SlaveConnect
type SlaveConnectData struct {
	ID          string
	URI         string
	Certificate SlaveConnectCertificate
	Encrypt     CredentialEncryptConfig
}

// SlaveConnectCertificate represents the valid certificate for the Slave
type SlaveConnectCertificate struct {
	Enabled            bool
	CACert             string
	ServerNameOverride string
	InsecureSkipVerify bool
}

// CredentialEncryptConfig represents the valid auth credential encrypt configuration
type CredentialEncryptConfig struct {
	Enabled   bool
	EncryptID string
}
