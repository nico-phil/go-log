package main

import (
	"context"
	"fmt"
	"os"
	"time"

	api "github.com/nico-phil/go-log/api/v1"
	"github.com/nico-phil/go-log/internal/agent"
	"github.com/nico-phil/go-log/internal/config"
	"github.com/nico-phil/go-log/internal/log"
	"github.com/travisjeffery/go-dynaport"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	err := os.Mkdir("segment-demo", 0755)

	seg0, err := log.NewSegment("segment-demo", 0, log.Config{})
	if err != nil {
		return
	}

	fmt.Println("seg0", seg0.BaseOffset())

	seg1, err := log.NewSegment("segment-demo", 1, log.Config{})
	if err != nil {
		return
	}

	fmt.Println("seg1", seg1.BaseOffset())

	seg2, err := log.NewSegment("segment-demo", 2, log.Config{})
	if err != nil {
		return
	}

	fmt.Println("seg2", seg2.BaseOffset())
}

func runAgent() {
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

		var startJoinAddrs []string
		if i != 0 {
			startJoinAddrs = append(startJoinAddrs, agents[0].Config.BindAddr)
		}

		config := agent.Config{
			NodeName:        fmt.Sprintf("%d", i),
			ServerTlsConfig: serverTlsconfig,
			PeerTlsConfig:   peerTlsConfig,
			Datadir:         "agent-demo",
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
