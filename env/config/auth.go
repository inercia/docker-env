package config

import (
	"bytes"
	"path/filepath"

	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/auth"
	"gopkg.in/yaml.v2"
)

type authConfig auth.Options

func NewAuthConfig(_ libmachine.API) *authConfig {
	certDir := mcndirs.GetMachineCertDir()
	return &authConfig{
		CertDir:          certDir,
		CaCertPath:       filepath.Join(certDir, "ca.pem"),
		CaPrivateKeyPath: filepath.Join(certDir, "ca-key.pem"),
		ClientCertPath:   filepath.Join(certDir, "cert.pem"),
		ClientKeyPath:    filepath.Join(certDir, "key.pem"),
		ServerCertPath:   filepath.Join(certDir, "server.pem"),
		ServerKeyPath:    filepath.Join(certDir, "server-key.pem"),
	}
}

func (auth authConfig) Copy() *authConfig {
	return &auth
}

func (auth *authConfig) Populate(api libmachine.API, root *Config, machine *machineConfig) error {
	if len(auth.CertDir) == 0 {
		auth.CertDir = mcndirs.GetMachineCertDir()
	}
	if len(auth.CaCertPath) == 0 {
		auth.CaCertPath = filepath.Join(auth.CertDir, "ca.pem")
	}
	if len(auth.CaPrivateKeyPath) == 0 {
		auth.CaPrivateKeyPath = filepath.Join(auth.CertDir, "ca-key.pem")
	}
	if len(auth.ClientCertPath) == 0 {
		auth.ClientCertPath = filepath.Join(auth.CertDir, "cert.pem")
	}
	if len(auth.ClientKeyPath) == 0 {
		auth.ClientKeyPath = filepath.Join(auth.CertDir, "key.pem")
	}
	if len(auth.ServerCertPath) == 0 {
		auth.ServerCertPath = filepath.Join(auth.CertDir, "server.pem")
	}
	if len(auth.ServerKeyPath) == 0 {
		auth.ServerKeyPath = filepath.Join(auth.CertDir, "server-key.pem")
	}

	return nil
}

func (auth *authConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	c := make(optionsMap)
	if err := unmarshal(c); err != nil {
		return err
	}

	// note: we must wait until Populate for setting default values, as some
	//       of them depend on "cert-dir" and that could be changed on a later
	//       call to UnmarshalYAML()

	if v, found := c["cert-dir"]; found {
		auth.CertDir = v
	}
	if v, found := c["tls-ca-cert"]; found {
		auth.CaCertPath = v
	}
	if v, found := c["tls-ca-key"]; found {
		auth.CaPrivateKeyPath = v
	}
	if v, found := c["tls-ca-cert-remote"]; found {
		auth.CaCertRemotePath = v
	}
	if v, found := c["tls-client-cert"]; found {
		auth.ClientCertPath = v
	}
	if v, found := c["tls-client-key"]; found {
		auth.ClientKeyPath = v
	}

	if v, found := c["server-cert-path"]; found {
		auth.ServerCertPath = v
	}
	if v, found := c["server-key-path"]; found {
		auth.ServerKeyPath = v
	}
	if v, found := c["server-cert-remote-path"]; found {
		auth.ServerCertRemotePath = v
	}
	if v, found := c["server-key-remote-path"]; found {
		auth.ServerKeyRemotePath = v
	}
	if v, found := c["server-cert-SANs"]; found {
		lst := []string{}
		err := yaml.Unmarshal(bytes.NewBufferString(v).Bytes(), lst)
		if err != nil {
			return err
		}
		auth.ServerCertSANs = lst
	}

	return nil
}
