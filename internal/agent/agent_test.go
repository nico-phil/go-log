package agent

import (
	"fmt"
	"os"
	"testing"

	"github.com/nico-phil/go-log/internal/config"
	"github.com/stretchr/testify/require"
	"github.com/travisjeffery/go-dynaport"
)

// TestAgent
func TestAgent(t *testing.T) {
	// lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", ports[0]))
	// require.NoError(t, err)

	serverTlsconfig, err := config.SetupTLSConfig(config.TLSConfig{
		CertFile:      config.ServerCertFile,
		KeyFile:       config.ServerKeyFile,
		CAFile:        config.CAFile,
		ServerAddress: "127.0.0.1",
		Server:        true,
	})
	require.NoError(t, err)

	peerTlsConfig, err := config.SetupTLSConfig(config.TLSConfig{
		CertFile:      config.RootClientCertFile,
		KeyFile:       config.RootClientKeyFile,
		CAFile:        config.CAFile,
		ServerAddress: "127.0.0.1",
		Server:        false,
	})
	require.NoError(t, err)

	for i := 0; i < 3; i++ {

		dir, err := os.MkdirTemp("", "agent-test-log")
		require.NoError(t, err)

		ports := dynaport.Get(2)
		bindArr := fmt.Sprintf("127.0.0.1:%d", ports[0])

		var startJoinAddrs []string
		if i != 0 {
			startJoinAddrs = append(startJoinAddrs, fmt.Sprintf("%d", i))
		}

		config := Config{
			NodeName:        fmt.Sprintf("%d", i),
			ServerTlsConfig: serverTlsconfig,
			PeerTlsConfig:   peerTlsConfig,
			Datadir:         dir,
			BindAddr:        bindArr,
			RPCPort:         ports[1],
			StartJoinAddr:   make([]string, 0),
			ACLModelFile:    config.ACLModelFile,
			ACLPolicyFile:   config.ACLPolicyFile,
		}

		_, err = New(config)
		// shutdown agen
		require.NoError(t, err)
	}

}
