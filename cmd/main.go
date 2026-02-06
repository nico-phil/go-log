package main

import (
	"context"
	"fmt"
	"os"
	"time"

	api "github.com/nico-phil/go-log/api/v1"
	"github.com/nico-phil/go-log/internal/agent"
	"github.com/nico-phil/go-log/internal/config"
	"github.com/travisjeffery/go-dynaport"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	serverTlsconfig, err := config.SetupTLSConfig(config.TLSConfig{
		CertFile:      config.ServerCertFile,
		KeyFile:       config.ServerKeyFile,
		CAFile:        config.CAFile,
		ServerAddress: "127.0.0.1",
		Server:        true,
	})
	if err != nil {
		return
	}

	peerTlsConfig, err := config.SetupTLSConfig(config.TLSConfig{
		CertFile:      config.RootClientCertFile,
		KeyFile:       config.RootClientKeyFile,
		CAFile:        config.CAFile,
		ServerAddress: "127.0.0.1",
		Server:        false,
	})
	if err != nil {
		return
	}

	agents := make([]*agent.Agent, 3)
	for i := 0; i < 3; i++ {

		ports := dynaport.Get(2)
		bindArr := fmt.Sprintf("127.0.0.1:%d", ports[0])

		err = os.Mkdir(fmt.Sprintf("%s%d", "agent-demo", i), 0755)
		if err != nil {
			return
		}

		var startJoinAddrs []string
		if i != 0 {
			startJoinAddrs = append(startJoinAddrs, agents[0].Config.BindAddr)
		}

		config := agent.Config{
			NodeName:        fmt.Sprintf("%d", i),
			ServerTlsConfig: serverTlsconfig,
			PeerTlsConfig:   peerTlsConfig,
			Datadir:         fmt.Sprintf("%s%d", "agent-demo", i),
			BindAddr:        bindArr,
			RPCPort:         ports[1],
			StartJoinAddr:   startJoinAddrs,
			ACLModelFile:    config.ACLModelFile,
			ACLPolicyFile:   config.ACLPolicyFile,
		}

		ag, _ := agent.New(config)
		agents[i] = ag
	}

	leaderClient := client(agents[0])

	want := []byte("hello")
	produceReponse, err := leaderClient.Produce(context.Background(), &api.ProduceRequest{
		Record: &api.Record{Value: want},
	})

	consumeReponse, err := leaderClient.Consume(context.Background(), &api.ConsumeRequest{
		Offset: produceReponse.Offset,
	})

	fmt.Println("consumeReponse0", consumeReponse.Record)

	time.Sleep(3 * time.Second)

	followerClient := client(agents[1])
	consumeReponse, err = followerClient.Consume(context.Background(), &api.ConsumeRequest{
		Offset: produceReponse.Offset,
	})

}

func client(agent *agent.Agent) api.LogClient {
	tlsCreds := credentials.NewTLS(agent.PeerTlsConfig)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(tlsCreds)}

	rpcAddr, _ := agent.Config.RPCAddr()

	conn, _ := grpc.NewClient(rpcAddr, opts...)

	logClient := api.NewLogClient(conn)

	return logClient
}
