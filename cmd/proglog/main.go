package proglog

import (
	"log"
	"os"
	"path"

	"github.com/nico-phil/go-log/internal/agent"
	"github.com/nico-phil/go-log/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type cli struct {
	cfg
}

type cfg struct {
	agent.Config
	serverTlsConfig config.TLSConfig
	PeerConfig      config.TLSConfig
}

func main() {
	// cli := cli{}

	cmd := &cobra.Command{
		Use:    "proglog",
		PreRun: nil,
		RunE:   nil,
	}

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func run(c *cli) error {
	return nil
}

func setupFlags(cmd *cobra.Command) error {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	cmd.Flags().String("config-file", "", "Path to config file.")

	dataDir := path.Join(os.TempDir(), "proglog")
	cmd.Flags().String("data-dir",
		dataDir,
		"Directory to store log and Raft data.")

	cmd.Flags().String("node-name", hostname, "Unique server ID.")

	cmd.Flags().String("bind-addr",
		"127.0.0.1:8401",
		"Address to bind Serf on.")

	cmd.Flags().Int("rpc-port",
		8400,
		"Port for RPC clients (and Raft) connections.")

	cmd.Flags().StringSlice("start-join-addrs",
		nil,
		"Serf addresses to join.")

	cmd.Flags().Bool("bootstrap", false, "Bootstrap the cluster.")

	cmd.Flags().String("acl-model-file", "", "Path to ACL model.")

	cmd.Flags().String("acl-policy-file", "", "Path to ACL policy.")
	cmd.Flags().String("server-tls-cert-file", "", "Path to server tls cert.")
	cmd.Flags().String("server-tls-key-file", "", "Path to server tls key.")
	cmd.Flags().String("server-tls-ca-file",
		"",
		"Path to server certificate authority.")

	cmd.Flags().String("peer-tls-cert-file", "", "Path to peer tls cert.")
	cmd.Flags().String("peer-tls-key-file", "", "Path to peer tls key.")
	cmd.Flags().String("peer-tls-ca-file",
		"",
		"Path to peer certificate authority.")

	return viper.BindPFlags(cmd.Flags())
}
