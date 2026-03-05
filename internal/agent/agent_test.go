package agent

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	api "github.com/nico-phil/go-log/api/v1"
	"github.com/nico-phil/go-log/internal/config"
	"github.com/stretchr/testify/require"
	"github.com/travisjeffery/go-dynaport"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// TestAgent
func TestAgent(t *testing.T) {
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

	agents := make([]*Agent, 3)
	for i := 0; i < 3; i++ {

		dir, err := os.MkdirTemp("", "agent-test-log")
		require.NoError(t, err)

		ports := dynaport.Get(2)
		bindArr := fmt.Sprintf("127.0.0.1:%d", ports[0])

		var startJoinAddrs []string
		if i != 0 {
			startJoinAddrs = append(startJoinAddrs, agents[0].Config.BindAddr)
		}

		config := Config{
			NodeName:        fmt.Sprintf("%d", i),
			ServerTlsConfig: serverTlsconfig,
			PeerTlsConfig:   peerTlsConfig,
			Datadir:         dir,
			BindAddr:        bindArr,
			RPCPort:         ports[1],
			StartJoinAddr:   startJoinAddrs,
			ACLModelFile:    config.ACLModelFile,
			ACLPolicyFile:   config.ACLPolicyFile,
			Bootstrap:       i == 0,
		}

		ag, err := New(config)
		require.NoError(t, err)

		agents[i] = ag
	}

	// shutdown agent
	defer func() {
		for _, agent := range agents {
			err := agent.Shutdown()
			require.NoError(t, err)
			require.NoError(t, os.RemoveAll(agent.Datadir))
		}
	}()

	leaderClient := client(t, agents[0])

	want := []byte("hello")

	produceReponse, err := leaderClient.Produce(context.Background(), &api.ProduceRequest{
		Record: &api.Record{Value: want},
	})
	require.NoError(t, err)

	consumeReponse, err := leaderClient.Consume(context.Background(), &api.ConsumeRequest{
		Offset: produceReponse.Offset,
	})
	require.NoError(t, err)
	require.Equal(t, consumeReponse.Record.Value, want)

	// we want to wait until replication finished
	time.Sleep(3 * time.Second)

	followerClient := client(t, agents[1])
	consumeReponse, err = followerClient.Consume(context.Background(), &api.ConsumeRequest{
		Offset: produceReponse.Offset,
	})

	require.NoError(t, err)

	require.Equal(t, consumeReponse.Record.Value, want)

}

// client creates new client for testing purpose
func client(t *testing.T, agent *Agent) api.LogClient {
	tlsCreds := credentials.NewTLS(agent.PeerTlsConfig)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(tlsCreds)}

	rpcAddr, err := agent.Config.RPCAddr()
	require.NoError(t, err)

	conn, err := grpc.NewClient(rpcAddr, opts...)
	require.NoError(t, err)

	logClient := api.NewLogClient(conn)

	return logClient
}
