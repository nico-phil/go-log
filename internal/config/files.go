package config

import (
	"os"
	"path/filepath"
)

var (
	// Certificate Autority file
	CAFile = configFile("ca.pem")

	// certificate and private key file for server
	ServerCertFile = configFile("server.pem")
	ServerKeyFile  = configFile("server-key.pem")

	// certificate and private key file for client
	ClientCertFile = configFile("client.pem")
	ClientKeyFile  = configFile("client-key.pem")

	// certificate and private key file for client
	RootClientCertFile = configFile("root-client.pem")
	RootClientKeyFile  = configFile("root-client-key.pem")

	// / certificate and private key file for unthorize client(for testing purpose)
	NoBodyClientCertFile = configFile("nobody-client.pem")
	NoBodyClientKeyFile  = configFile("nobody-client-key.pem")

	// Casbin model and policy file
	ACLModelFile  = configFile("model.conf")
	ACLPolicyFile = configFile("policy.csv")
)

// configFile returns the full path of certs and other file
func configFile(filename string) string {
	if dir := os.Getenv("CONFIG_DIR"); dir != "" {
		return filepath.Join(dir, filename)
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return filepath.Join(homeDir, ".proglog", filename)
}
