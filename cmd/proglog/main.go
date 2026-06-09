package proglog

import (
	"log"
	"os"
	"path"

	"github.com/nico-phil/go-log/internal/agent"
	"github.com/nico-phil/go-log/internal/config"
	"github.com/spf13/cobra"
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
	_, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	cmd.Flags().String("config-file", "", "Path to config file.")

	dataDir := path.Join(os.TempDir(), "proglog")

	cmd.Flags().String("data-dir", dataDir, "Directory to store log and Raft Data")

	cmd.Flags().String("bind-addr", "127.0.0.1:5000", "Directory to store log and Raft Data")
	return nil
}
