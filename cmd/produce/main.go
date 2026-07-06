package main

import (
	"context"
	"flag"
	"fmt"

	api "github.com/nico-phil/go-log/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	addr := flag.String("addr", ":8400", "service addr")

	// peerTlsConfig, err := config.SetupTLSConfig(config.TLSConfig{
	// 	CertFile:      config.RootClientCertFile,
	// 	KeyFile:       config.RootClientKeyFile,
	// 	CAFile:        config.CAFile,
	// 	ServerAddress: "127.0.0.1",
	// 	Server:        false,
	// })

	// tlsCreds := credentials.NewTLS(peerTlsConfig)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	conn, err := grpc.NewClient(*addr, opts...)
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		return
	}

	client := api.NewLogClient(conn)

	record := &api.Record{
		Value: []byte("hello, world!"),
	}

	resp, err := client.Produce(context.Background(), &api.ProduceRequest{Record: record})
	if err != nil {
		fmt.Printf("Error producing record: %v\n", err)
		return
	}

	fmt.Printf("Response: %v\n", resp)

}
