package main

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Specification represents application environment variables configuration.
type Specification struct {
	// KubeConfig is an absolute path to the kubeconfig file.
	KubeConfig string `envconfig:"KUBECONFIG" required:"false"`

	// VaultAddr defines the address of Vault that dweller is communicating with.
	VaultAddr string `envconfig:"VAULT_ADDR" required:"true"`

	// ValutToken defines the Vault token that dweller is authenticating with.
	ValutToken string `envconfig:"VAULT_TOKEN" required:"true"`

	// LogLevel defines log level for the logger. By default level is "info".
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`
}

// SpecificationFromEnvironment returns specification loaded from environment
// variables.
func SpecificationFromEnvironment() (Specification, error) {
	if err := godotenv.Overload(); err != nil {
		// we do not care if there is no .env file.
	}

	var s Specification
	err := envconfig.Process("", &s)
	if err != nil {
		return s, err
	}

	s.KubeConfig = os.ExpandEnv(s.KubeConfig)
	return s, nil
}
